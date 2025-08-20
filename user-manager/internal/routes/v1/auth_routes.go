package V1routes

import (
	v1handler "user-management-api/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

type AuthRoutes struct {
	authHandler *v1handler.AuthHandler
}

func NewAuthRoutes(handler *v1handler.AuthHandler) *AuthRoutes {
	return &AuthRoutes{
		authHandler: handler,
	}
}
func (ar *AuthRoutes) Register(r *gin.RouterGroup) {
	auth := r.Group("/auth") // Đường dẫn gốc cho auth
	{
		auth.POST("/login", ar.authHandler.Login) // Đăng nhập
		auth.POST("logout", ar.authHandler.Logout) // Đăng xuat
		auth.POST("/refresh", ar.authHandler.RefreshToken) // Làm mới token
		auth.POST("/forgot-password", ar.authHandler.ForgotPassword) // Quên mật khẩu
		auth.POST("/reset-password", ar.authHandler.ResetPassword) // Đặt lại mật khẩu
	}
}


