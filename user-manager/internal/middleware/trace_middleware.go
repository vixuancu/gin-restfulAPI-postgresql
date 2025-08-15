package middleware

import (
	"context"
	"user-management-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)



func TraceMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID := ctx.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String() // Tạo trace ID mới nếu không có trong header
		}
		contextValue := context.WithValue(ctx.Request.Context(), logger.TraceIDKey, traceID) // Lưu trace ID vào context
		ctx.Request = ctx.Request.WithContext(contextValue) // Cập nhật context của request (go)
		ctx.Writer.Header().Set("X-Trace-ID", traceID) // Đặt trace ID vào header của response
		ctx.Set(string(logger.TraceIDKey), traceID) // Lưu trace ID vào context của Gin để có thể truy cập sau này
		ctx.Next()
	}
}