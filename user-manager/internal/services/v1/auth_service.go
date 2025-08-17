package v1services

import (
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *authService {
	return &authService{
		userRepo: repo,
	}
}
// authentication là xác thực người dùng
// authorization là phân quyền người dùng
func (as *authService) Login(c *gin.Context,email,password string) error {
	context := c.Request.Context() // Lấy context của go từ gin.Context
	email = utils.NormalizeString(email)  

	user,err := as.userRepo.GetByEmail(context, email)
	if err != nil {
		return utils.NewError("Invalid email or password",utils.ErrorCodeUnauthorized)
	}
	// Kiểm tra mật khẩu
	if err :=bcrypt.CompareHashAndPassword([]byte(user.UserPassword),[]byte(password)); err != nil {
		return utils.NewError("Invalid email or password",utils.ErrorCodeUnauthorized)
	}
	

	return nil
}

func (as *authService) Logout(c *gin.Context) error {
	// Implement login logic here

	return nil
}
