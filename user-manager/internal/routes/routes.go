package routes

import (
	"user-management-api/internal/middleware"
	"user-management-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type Routes interface {
	Register(r *gin.RouterGroup)
}

// lấy ra interface Routes để định nghĩa các route
func RegisterRoutes(router *gin.Engine, routes ...Routes) {

	httpLogger := utils.NewLoggerWithPath("../../internal/logs/http.log", "info")
	recoveryLogger := utils.NewLoggerWithPath("../../internal/logs/recovery.log", "warning")
	rateLimiterLogger := utils.NewLoggerWithPath("../../internal/logs/rate_limiter.log", "warning")

	router.Use(
		middleware.APIKeyMiddleware(),
		middleware.RateLimitMiddleware(rateLimiterLogger),
		middleware.LoggerMiddleware(httpLogger),
		middleware.RecoveryMiddleware(recoveryLogger),
		middleware.AuthMiddleware(),
	)
	v1api := router.Group("/api/v1")
	for _, r := range routes {
		r.Register(v1api)
	}
}
