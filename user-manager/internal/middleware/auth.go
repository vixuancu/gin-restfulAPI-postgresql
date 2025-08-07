package middleware

import "github.com/gin-gonic/gin"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Implement your authentication logic here
		// If authentication fails, you can abort the request
		c.Next()
	}
}