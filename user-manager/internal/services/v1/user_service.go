package v1services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"
	"user-management-api/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo repository.UserRepository
	cache    *cache.RedisCacheService
}

func NewUserService(repo repository.UserRepository, redisClient *redis.Client) UserService {
	return &userService{
		userRepo: repo,
		cache:    cache.NewRedisCacheService(redisClient),
	}
}
func (us *userService) GetAllUsers(ctx *gin.Context, search, orderBy, sort string, limit, page int32, deleted bool) ([]sqlc.User, int32, error) {
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

	if limit < 1 || limit > 500 {
		envLimit := utils.GetEnv("LIMIT_ITEM_ON_PER_PAGE", "10") // Lấy giá trị từ biến môi trường, nếu không có thì mặc định là 10
		limitInt, err := strconv.Atoi(envLimit)                  // Chuyển đổi chuỗi sang số nguyên
		if err != nil && limitInt < 1 {
			limitInt = 10 // Nếu không thể chuyển đổi hoặc giá trị nhỏ hơn 1 thì mặc định là 10
		}
		limit = int32(limitInt) // Cập nhật giá trị limit
	}
	/* Get Cache Redis*/
	cacheKey := us.generateCacheKey(search, orderBy, sort, limit, page, deleted) // Tạo khóa cache dựa trên các tham số
	var cacheData struct {
		Users []sqlc.User `json:"users"`
		Total int32       `json:"total"`
	}
	if err := us.cache.Get(cacheKey, &cacheData); err == nil && cacheData.Users != nil {
		log.Println("🍻🍺Cache Redis GetAllUsers")
		return cacheData.Users, cacheData.Total, nil // Trả về dữ liệu từ cache nếu có
	}
	log.Println("🍻🍺Fetching from DB")
	total, err := us.userRepo.CountUsers(context, search, deleted) // Đếm tổng số người dùng
	if err != nil {
		return []sqlc.User{}, 0, utils.WrapError(err, "failed to count users", utils.ErrorCodeInternalServer)
	}
	offset := (page - 1) * limit // Tính toán offset dựa trên trang và giới hạn

	users, err := us.userRepo.GetAllV2(context, search, orderBy, sort, limit, offset, deleted)
	if err != nil {
		return []sqlc.User{}, 0, utils.WrapError(err, "failed to get all users", utils.ErrorCodeInternalServer)
	}

	/*Create Cache Data*/
	cacheData = struct {
		Users []sqlc.User `json:"users"`
		Total int32       `json:"total"`
	}{
		Users: users,
		Total: int32(total),
	}
	us.cache.Set(cacheKey, cacheData, 5*time.Minute) // Lưu dữ liệu vào cache với thời gian hết hạn là 5 phút
	return users, int32(total), nil
}
func (us *userService) CreateUser(ctx *gin.Context, input sqlc.CreateUserParams) (sqlc.User, error) {
	context := ctx.Request.Context() // Lấy context của go từ gin.Context

	input.UserEmail = utils.NormalizeString(input.UserEmail)                                           // Chuyển đổi email thành chữ thường và loại bỏ khoảng trắng
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.UserPassword), bcrypt.DefaultCost) // Mã hóa mật khẩu
	if err != nil {
		return sqlc.User{}, utils.WrapError(err, "failed to hash Password", utils.ErrorCodeInternalServer)
	}
	input.UserPassword = string(hashedPassword) // Cập nhật mật khẩu đã mã hóa vào user
	user, err := us.userRepo.Create(context, input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.User{}, utils.NewError("email already exists", utils.ErrorCodeConflict)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to create user", utils.ErrorCodeInternalServer)
	}
	// Xóa cache liên quan đến người dùng để đảm bảo dữ liệu mới được cập nhật
	 if err := us.cache.Clear("users:*"); err != nil {
		log.Printf("Failed to clear cache for user creation: %v", err)
	 } // Xóa cache liên quan đến người dùng
	return user, nil
}
func (us *userService) GetUserByUUID(c *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
	context := c.Request.Context() // Lấy context của go từ gin.Context

	user, err := us.userRepo.GetByUuid(context, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to get user by UUID", utils.ErrorCodeInternalServer)
	}
	return user, nil

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
	updatedUser, err := us.userRepo.Update(context, input) // Cập nhật thông tin người dùng
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to update user", utils.ErrorCodeInternalServer)
	}
	return updatedUser, nil
}
func (us *userService) DeleteUser(ctx *gin.Context, uuid uuid.UUID) error {
	context := ctx.Request.Context() // Lấy context của go từ gin.Context
	_, err := us.userRepo.Delete(context, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return utils.WrapError(err, "failed to Delete user", utils.ErrorCodeInternalServer)
	}
	return nil
}
func (us *userService) SoftDeleteUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context() // Lấy context của go từ gin.Context
	softDeletedUser, err := us.userRepo.SoftDelete(context, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to softDelete user", utils.ErrorCodeInternalServer)
	}
	return softDeletedUser, nil

}
func (us *userService) RestoreUser(ctx *gin.Context, uuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context() // Lấy context của go từ gin.Context
	restoreUser, err := us.userRepo.Restore(context, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.NewError("user not found", utils.ErrorCodeNotFound)
		}
		return sqlc.User{}, utils.WrapError(err, "failed to Restore user", utils.ErrorCodeInternalServer)
	}
	return restoreUser, nil
}

func (us *userService) generateCacheKey( search, orderBy, sort string, limit, page int32, deleted bool) string{
	search = strings.TrimSpace(search) // Loại bỏ khoảng trắng ở đầu và cuối
	if search == "" {
		search = "none"
	}

	orderBy = strings.TrimSpace(orderBy) // Loại bỏ khoảng trắng ở đầu và cuối
	if orderBy == "" {
		orderBy = "user_created_at"
	}
	sort = strings.TrimSpace(sort) // Loại bỏ khoảng trắng ở đầu và cuối
	if sort == "" {
		sort = "asc" // Mặc định là sắp xếp tăng dần
	}
	return fmt.Sprintf("users:%s:%s:%s:%d:%d:%t", search, orderBy, sort, limit, page, deleted)
}
