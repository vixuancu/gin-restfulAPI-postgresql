package v1services

import (
	"database/sql"
	"errors"
	"strconv"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		userRepo: repo,
	}
}
func (us *userService) GetAllUsers(ctx *gin.Context, search, orderBy, sort string, limit, page  int32) ([]sqlc.User,int32, error) {
	context := ctx.Request.Context() // Lấy context của go từ gin.Context

	if sort == "" {
		sort = "asc" // Mặc định là sắp xếp tăng dần
	}
	if orderBy == "" {
		orderBy = "user_created_at" // Mặc định là sắp xếp theo user
	}
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit >500 {
		envLimit := utils.GetEnv("LIMIT_ITEM_ON_PER_PAGE", "10") // Lấy giá trị từ biến môi trường, nếu không có thì mặc định là 10
		limitInt,err := strconv.Atoi(envLimit) // Chuyển đổi chuỗi sang số nguyên
		if err != nil && limitInt < 1 {
			limitInt = 10 // Nếu không thể chuyển đổi hoặc giá trị nhỏ hơn 1 thì mặc định là 10
		}
		limit = int32(limitInt) // Cập nhật giá trị limit
	}
	total, err := us.userRepo.CountUsers(context, search) // Đếm tổng số người dùng
	if err != nil {
		return []sqlc.User{},0, utils.WrapError(err, "failed to count users", utils.ErrorCodeInternalServer)
	}
	offset := (page -1) * limit // Tính toán offset dựa trên trang và giới hạn
	users,err :=us.userRepo.GetAll(context, search, orderBy, sort, limit, offset) 
	if err != nil {
		return []sqlc.User{},0, utils.WrapError(err, "failed to get all users", utils.ErrorCodeInternalServer)
	}
	return users,int32(total), nil
}
func (us *userService) CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error) {
	context := ctx.Request.Context() // Lấy context của go từ gin.Context

	input.UserEmail = utils.NormalizeString(input.UserEmail)                                            // Chuyển đổi email thành chữ thường và loại bỏ khoảng trắng
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.UserPassword), bcrypt.DefaultCost) // Mã hóa mật khẩu
	if err != nil {
		return sqlc.User{}, utils.WrapError(err, "failed to hash Password", utils.ErrorCodeInternalServer)
	}
	input.UserPassword = string(hashedPassword) // Cập nhật mật khẩu đã mã hóa vào user
	user, err := us.userRepo.Create(context, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.User{}, utils.NewError( "email already exists", utils.ErrorCodeConflict)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to create user", utils.ErrorCodeInternalServer)
	}

	return user, nil
}
func (us *userService) GetUserByUUID(uuid string) {

}
func (us *userService) UpdateUser(ctx *gin.Context, input sqlc.UpdateUserParams) (sqlc.User, error) {
	context := ctx.Request.Context() // Lấy context của go từ gin.Context

	if input.UserPassword != nil && *input.UserPassword != "" {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.UserPassword), bcrypt.DefaultCost) // Mã hóa mật khẩu
	if err != nil {
		return sqlc.User{}, utils.WrapError(err, "failed to hash Password", utils.ErrorCodeInternalServer)
	}
	hashed := string(hashedPassword)
	input.UserPassword = &hashed // Cập nhật mật khẩu đã mã hóa vào user 
	}
	updatedUser,err:= us.userRepo.Update(context, input) // Cập nhật thông tin người dùng
	if err != nil {
		if errors.Is(err,sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to update user", utils.ErrorCodeInternalServer)
	}
	return updatedUser, nil
}
func (us *userService) DeleteUser(ctx *gin.Context, uuid uuid.UUID)  error {
context := ctx.Request.Context() // Lấy context của go từ gin.Context
_,err:= us.userRepo.Delete(context, uuid) 
	if err != nil {
		if errors.Is(err,sql.ErrNoRows) {
			return  utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return  utils.WrapError(err, "failed to Delete user", utils.ErrorCodeInternalServer)
	}
	return  nil
}
func (us *userService) SoftDeleteUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
context := ctx.Request.Context() // Lấy context của go từ gin.Context
softDeletedUser,err:= us.userRepo.SoftDelete(context, uuid) 
	if err != nil {
		if errors.Is(err,sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to softDelete user", utils.ErrorCodeInternalServer)
	}
	return softDeletedUser, nil

}
func (us *userService) RestoreUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
context := ctx.Request.Context() // Lấy context của go từ gin.Context
restoreUser,err:= us.userRepo.Restore(context, uuid) 
	if err != nil {
		if errors.Is(err,sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to Restore user", utils.ErrorCodeInternalServer)
	}
	return restoreUser, nil
}
