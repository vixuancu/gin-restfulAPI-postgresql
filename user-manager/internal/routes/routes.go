package routes

import (
	"user-management-api/internal/middleware"
	"user-management-api/internal/utils"
	"user-management-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Routes interface {
	Register(r *gin.RouterGroup)
}

// lấy ra interface Routes để định nghĩa các route
func RegisterRoutes(router *gin.Engine, routes ...Routes) {

	httpLogger := NewLoggerWithPath("../../internal/logs/http.log", "info")
	recoveryLogger := NewLoggerWithPath("../../internal/logs/recovery.log", "warning")
	rateLimiterLogger := NewLoggerWithPath("../../internal/logs/recovery.log", "warning")

	router.Use(middleware.RateLimitMiddleware(rateLimiterLogger), middleware.LoggerMiddleware(httpLogger), middleware.RecoveryMiddleware(recoveryLogger), middleware.APIKeyMiddleware(), middleware.AuthMiddleware())
	v1api := router.Group("/api/v1")
	for _, r := range routes {
		r.Register(v1api)
	}
}

func NewLoggerWithPath(path string, level string) *zerolog.Logger {
	config := logger.LoggerConfig{
		Level:      level,
		Filename:   path,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     5,    //
		Compress:   true, // disabled by default
		IsDev:      utils.GetEnv("APP_ENV", "development"),
	}
	return logger.NewLogger(config)
}
