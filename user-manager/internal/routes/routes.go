package routes

import (
	"user-management-api/internal/middleware"
	"user-management-api/internal/utils"

	"github.com/gin-contrib/gzip"
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
		middleware.RecoveryMiddleware(recoveryLogger),
		middleware.CORSMiddleware(),
		middleware.LoggerMiddleware(httpLogger),
		middleware.APIKeyMiddleware(),
		middleware.RateLimitMiddleware(rateLimiterLogger),
		middleware.TraceMiddleware(),
		middleware.AuthMiddleware(),
	)
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	v1api := router.Group("/api/v1")
	for _, r := range routes {
		r.Register(v1api)
	}
	// Đăng ký các route không tìm thấy
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"path":  c.Request.URL.Path,
		})
	})
}
