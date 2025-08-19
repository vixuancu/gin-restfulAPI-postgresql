package routes

import (
	"user-management-api/internal/middleware"
	V1routes "user-management-api/internal/routes/v1"
	"user-management-api/internal/utils"
	"user-management-api/pkg/auth"
	"user-management-api/pkg/cache"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type Routes interface {
	Register(r *gin.RouterGroup)
}

// lấy ra interface Routes để định nghĩa các route
func RegisterRoutes(router *gin.Engine, authService auth.TokenService, cacheService cache.RedisCacheService ,routes ...Routes) {

	httpLogger := utils.NewLoggerWithPath("../../internal/logs/http.log", "info")
	recoveryLogger := utils.NewLoggerWithPath("../../internal/logs/recovery.log", "warning")
	rateLimiterLogger := utils.NewLoggerWithPath("../../internal/logs/rate_limiter.log", "warning")

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(
		middleware.RecoveryMiddleware(recoveryLogger),
		middleware.CORSMiddleware(),
		middleware.LoggerMiddleware(httpLogger),
		middleware.APIKeyMiddleware(),
		middleware.RateLimitMiddleware(rateLimiterLogger),
		middleware.TraceMiddleware(),
	)

	v1api := router.Group("/api/v1")
	middleware.InitAuthMiddleware(authService,cacheService)
	protected := v1api.Group("")
	protected.Use(middleware.AuthMiddleware())

	for _, r := range routes {

		switch r.(type) {
		case *V1routes.AuthRoutes:
			r.Register(v1api) // Routes KHÔNG cần authentication
		default:
			r.Register(protected) // Routes CẦN authentication
		}

	}
	// Đăng ký các route không tìm thấy
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Not Found",
			"path":  c.Request.URL.Path,
		})
	})
}
