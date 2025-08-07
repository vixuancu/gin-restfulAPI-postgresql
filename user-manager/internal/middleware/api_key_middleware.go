package middleware

import (
	"github.com/gin-gonic/gin"
	
	"net/http"
	"os"
)

func APIKeyMiddleware() gin.HandlerFunc {
	expectedKey := os.Getenv("API_KEY")
	if expectedKey == "" {
		expectedKey = "default_api_key" //
	}
	return func(c *gin.Context) {
		apikey := c.GetHeader("X-API-KEY")
		// log.Println("Apikey:", apikey)
		if apikey == "" {
			//AbortWithStatusJSON Abort: Dừng middleware hoặc handler hiện tại, không chạy tiếp các handler phía sau nữa.
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "API key is required",
			})
			return
		}
		if apikey != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}
		c.Next()

	}
}
