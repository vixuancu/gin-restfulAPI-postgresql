package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RecoveryMiddleware(recoveryLogger *zerolog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Ghi log lỗi panic
				stack := debug.Stack()
				stack_at := ExtractFirstAppStackLine(stack) // Lấy dòng đầu tiên của stack trace
				recoveryLogger.Error().
					Str("path", ctx.Request.URL.Path).
					Str("method", ctx.Request.Method).
					Str("client_ip", ctx.ClientIP()).
					Str("panic", fmt.Sprintf("%v", err)).
					Str("stack_at", stack_at).
					Str("stack", string(stack)).
					Msg("Panic recovered in request")
				// Xử lý lỗi panic tại đây, ví dụ: ghi log, trả về lỗi cho client, v.v.
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    "INTERNAL_SERVER_ERROR",
					"message": "Please try again later",
				})
			}
		}()
		ctx.Next()
	}
}

func ExtractFirstAppStackLine(stack []byte) string {
	lines := bytes.Split(stack, []byte("\n")) // Tách stack trace thành các dòng
	for _, line := range lines {
		if bytes.Contains((line), []byte(".go")) &&
			!bytes.Contains((line), []byte("/runtime/")) && // Loại trừ các dòng liên quan đến runtime
			!bytes.Contains((line), []byte("/debug/")) &&
			!bytes.Contains((line), []byte("revovery_middleware.go")) { // Loại trừ dòng của middleware này
			cleanLine := strings.TrimSpace(string(line)) // Loại bỏ khoảng trắng thừa
			return cleanLine                             // Trả về dòng đầu tiên chứa thông tin về file và dòng code
		}
	}
	return ""
}
