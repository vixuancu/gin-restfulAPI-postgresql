package v1services

import (
	"strings"
	"time"
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"
	"user-management-api/pkg/auth"
	"user-management-api/pkg/cache"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo     repository.UserRepository
	TokenService auth.TokenService
	cache        cache.RedisCacheService
}

func NewAuthService(repo repository.UserRepository, TokenService auth.TokenService, cache cache.RedisCacheService) *authService {
	return &authService{
		userRepo:     repo,
		TokenService: TokenService,
		cache:        cache,
	}
}

// authentication là xác thực người dùng
// authorization là phân quyền người dùng
func (as *authService) Login(c *gin.Context, email, password string) (string, string, int, error) {
	context := c.Request.Context() // Lấy context của go từ gin.Context
	email = utils.NormalizeString(email)

	user, err := as.userRepo.GetByEmail(context, email)
	if err != nil {
		return "", "", 0, utils.NewError("Invalid email or password", utils.ErrorCodeUnauthorized)
	}
	// Kiểm tra mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPassword), []byte(password)); err != nil {
		return "", "", 0, utils.NewError("Invalid email or password", utils.ErrorCodeUnauthorized)
	}
	acesstoken, err := as.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate access token", utils.ErrorCodeInternalServer)
	}
	refreshtoken, err := as.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate refresh token", utils.ErrorCodeInternalServer)
	}

	if err := as.TokenService.StoreRefreshToken(refreshtoken); err != nil {
		return "", "", 0, utils.NewError("Cannot Save refresh Token in redis", utils.ErrorCodeInternalServer)
	}

	return acesstoken, refreshtoken.Token, int(auth.AcessTokenTTL), nil
}

func (as *authService) Logout(c *gin.Context, refreshTokenString string) error {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.NewError("Authorization header missing ", utils.ErrorCodeUnauthorized) // Trả về lỗi nếu không có header
	}
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")  // Lấy token từ header
	_, claims, err := as.TokenService.ParseToken(accessToken) // Phân tích token
	if err != nil {
		return utils.NewError("Invalid accress token ", utils.ErrorCodeUnauthorized) // Trả về lỗi nếu không có header
	}
	// lấy jti từ claims
	if jti, ok := claims["jti"].(string); ok {
		expUnix, _ := claims["exp"].(float64) // Lấy thời gian hết hạn từ claims
		exp := time.Unix(int64(expUnix), 0)   // Chuyển đổi sang thời gian
		key := "blacklist:" + jti             // Tạo khóa blacklist
		ttl := time.Until(exp)                // Tính thời gian hết hạn bằng thời gian còn lại của access token
		as.cache.Set(key, "revoked", ttl)     // Lưu vào cache với thời gian hết hạn
	}

	// Vô hiệu hóa refresh token
	if _, err := as.TokenService.ValidateRefreshToken(refreshTokenString); err != nil {
		return utils.WrapError(err, "Invalid refresh token or revoked", utils.ErrorCodeUnauthorized)
	}
	if err := as.TokenService.RevokedRefreshToken(refreshTokenString); err != nil {
		return utils.NewError("Failed to revoke old refresh token", utils.ErrorCodeInternalServer)
	}
	return nil
}
func (as *authService) RefreshToken(c *gin.Context, refreshTokenString string) (string, string, int, error) {
	context := c.Request.Context() // Lấy context của go từ gin.Context
	token, err := as.TokenService.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", 0, utils.WrapError(err, "Invalid refresh token or revoked", utils.ErrorCodeUnauthorized)
	}
	// Kiểm tra refresh token. Trả thông tin user_uuid
	userUUID, _ := uuid.Parse(token.UserUUID)

	// Lấy thông tin user
	user, err := as.userRepo.GetByUuid(context, userUUID)
	if err != nil {
		return "", "", 0, utils.WrapError(err, "User not found", utils.ErrorCodeNotFound)
	}
	// Tạo access token mới
	acesstoken, err := as.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate access token", utils.ErrorCodeInternalServer)
	}
	// Tạo refresh token mới
	refreshtoken, err := as.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate refresh token", utils.ErrorCodeInternalServer)
	}
	// Vô hiệu hóa refresh token cũ
	if err := as.TokenService.RevokedRefreshToken(refreshTokenString); err != nil {
		return "", "", 0, utils.NewError("Failed to revoke old refresh token", utils.ErrorCodeInternalServer)
	}
	// Lưu refresh token mới vào Redis
	if err := as.TokenService.StoreRefreshToken(refreshtoken); err != nil {
		return "", "", 0, utils.NewError("Cannot Save refresh Token in redis", utils.ErrorCodeInternalServer)
	}
	return acesstoken, refreshtoken.Token, int(auth.AcessTokenTTL), nil
}
