package v1services

import (
	"database/sql"
	"errors"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"

	"github.com/gin-gonic/gin"
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
func (us *userService) GetAllUsers(search string, page int, limit int) {

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
func (us *userService) DeleteUser(uuid string) {

}
