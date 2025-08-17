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
	}
}


