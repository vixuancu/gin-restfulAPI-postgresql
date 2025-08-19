package middleware

import (
	"net/http"
	"strings"
	"user-management-api/pkg/auth"
	"user-management-api/pkg/cache"

	"github.com/gin-gonic/gin"
)

var (
	jwtService auth.TokenService // Khai báo biến jwtService để sử dụng trong middleware
	cacheService cache.RedisCacheService // Khai báo biến cacheService để sử dụng trong middleware
)

func InitAuthMiddleware(jwtSvc auth.TokenService,cache cache.RedisCacheService) {
	jwtService = jwtSvc // Khởi tạo jwtService với TokenService đã được inject
	cacheService = cache // Khởi tạo cacheService với RedisCacheService đã được inject
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid1"}) // abort để dừng xử lý tiếp
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ") // Lấy token từ header
		_, claims, err := jwtService.ParseToken(tokenString)     // Phân tích token
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid2"})
			return
		}
		if jti, ok := claims["jti"].(string); ok {
			key := "blacklist:" + jti // Tạo khóa blacklist
			exists,err:=cacheService.Exists(key) // Kiểm tra xem token có trong blacklist không
			if err != nil || exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is blacklisted"})
				return
			}
		}
		payload, err := jwtService.DecryptAccessTokenPayload(tokenString) // Giải mã payload nếu cần thiết
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid3"})
			return
		}
		c.Set("user_uuid", payload.UserUUID) // Lưu thông tin người dùng vào context
		c.Set("user_email", payload.Email)   // Lưu email người dùng vào context
		c.Set("user_role", payload.Role)     // Lưu vai trò người dùng vào context
		c.Next()
	}
}
