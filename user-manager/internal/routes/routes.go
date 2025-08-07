package routes

import (
	"user-management-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Routes interface {
	Register(r *gin.RouterGroup)
	
}
// lấy ra interface Routes để định nghĩa các route
func RegisterRoutes(router *gin.Engine, routes ...Routes) {
	router.Use(middleware.LoggerMiddleware(), middleware.APIKeyMiddleware(), middleware.AuthMiddleware(), middleware.RateLimitMiddleware())
	v1api := router.Group("/api/v1")
	for _, r := range routes {
		r.Register(v1api)
	}
}