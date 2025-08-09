package V1routes

import (
	v1handler "user-management-api/internal/handler/v1"

	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	userHandler *v1handler.UserHandler
}

func NewUserRoutes(handler *v1handler.UserHandler) *UserRoutes {
	return &UserRoutes{
		userHandler: handler,
	}
}
func (ur *UserRoutes) Register(r *gin.RouterGroup) {
	userGroup := r.Group("/users")
	{
		userGroup.GET("/", ur.userHandler.GetAllUsers)
		userGroup.POST("/", ur.userHandler.CreateUsers)
		userGroup.GET("/:uuid", ur.userHandler.GetUserByUUID)
		userGroup.PUT("/:uuid", ur.userHandler.UpdateUser)
		userGroup.DELETE("/:uuid", ur.userHandler.DeleteUser)
		userGroup.GET("/panic", ur.userHandler.PanicUser)
	}
}


