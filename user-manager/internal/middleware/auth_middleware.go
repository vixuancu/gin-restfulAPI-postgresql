package middleware

import (
	"net/http"
	"strings"
	"user-management-api/pkg/auth"

	"github.com/gin-gonic/gin"
)

var (
	jwtService auth.TokenService // Khai báo biến jwtService để sử dụng trong middleware
)

func InitAuthMiddleware(jwtSvc auth.TokenService) {
	jwtService = jwtSvc // Khởi tạo jwtService với TokenService đã được inject
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid1"}) // abort để dừng xử lý tiếp
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ") // Lấy token từ header
		_, _, err := jwtService.ParseToken(tokenString)          // Phân tích token
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid2"})
			return
		}

		payload, err := jwtService.DecryptAccessTokenPayload(tokenString) // Giải mã payload nếu cần thiết
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid3"})
			return
		}
		c.Set("user_uuid", payload.UserUUID) // Lưu thông tin người dùng vào context
		c.Set("user_email", payload.Email) // Lưu email người dùng vào context
		c.Set("user_role", payload.Role) // Lưu vai trò người dùng vào context
		c.Next()
	}
}
