package v1services

import (
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"
	"user-management-api/pkg/auth"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo repository.UserRepository
	TokenService auth.TokenService
}

func NewAuthService(repo repository.UserRepository, TokenService auth.TokenService) *authService {
	return &authService{
		userRepo: repo,
		TokenService: TokenService,
	}
}
// authentication là xác thực người dùng
// authorization là phân quyền người dùng
func (as *authService) Login(c *gin.Context,email,password string) (string,int,error)  {
	context := c.Request.Context() // Lấy context của go từ gin.Context
	email = utils.NormalizeString(email)  

	user,err := as.userRepo.GetByEmail(context, email)
	if err != nil {
		return "",0,utils.NewError("Invalid email or password",utils.ErrorCodeUnauthorized)
	}
	// Kiểm tra mật khẩu
	if err :=bcrypt.CompareHashAndPassword([]byte(user.UserPassword),[]byte(password)); err != nil {
		return "",0,utils.NewError("Invalid email or password",utils.ErrorCodeUnauthorized)
	}
	acesstoken,err := as.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "",0,utils.NewError("Failed to generate access token",utils.ErrorCodeInternalServer)
	}

	return acesstoken,int(auth.AcessTokenTTL),nil
}

func (as *authService) Logout(c *gin.Context) error {
	// Implement login logic here

	return nil
}
