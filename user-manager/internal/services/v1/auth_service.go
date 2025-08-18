package v1services

import (
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"
	"user-management-api/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo     repository.UserRepository
	TokenService auth.TokenService
}

func NewAuthService(repo repository.UserRepository, TokenService auth.TokenService) *authService {
	return &authService{
		userRepo:     repo,
		TokenService: TokenService,
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

func (as *authService) Logout(c *gin.Context) error {
	// Implement login logic here

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
