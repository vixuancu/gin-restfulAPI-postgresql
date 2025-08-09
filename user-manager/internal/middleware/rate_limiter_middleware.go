package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Client struct {
	Limiter  *rate.Limiter
	Lastseen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client) // Lưu trữ client theo IP
)

func getClientIP(c *gin.Context) string {
	// Lấy IP từ header X-Forwarded-For nếu có, nếu không thì lấy IP thực
	ip := c.ClientIP()
	if ip == "" {
		ip = c.Request.RemoteAddr
	}
	return ip
}
func getRateLimitter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	client, exists := clients[ip]
	if !exists {
		// Tạo mới client với giới hạn 1 request mỗi giây và burst số lượng  5 requests xử lý đồng thời
		newclient := &Client{
			Limiter:  rate.NewLimiter(5, 10),
			Lastseen: time.Now(),
		}
		clients[ip] = newclient
		// log.Printf("a client[%s]-{limiter: %v, lastseen: %v} is created", ip, newclient.Limiter, newclient.Lastseen)
		return newclient.Limiter
	} else {
		// Cập nhật thời gian cuối cùng thấy client
		// log.Printf("a client[%s]-{limiter: %v, lastseen: %v} is created", ip, client.Limiter, client.Lastseen)
		client.Lastseen = time.Now()

		return client.Limiter
	}

}
func CleanupClients() {
	for {
		time.Sleep(time.Minute) // Chạy mỗi 1 phút
		mu.Lock()
		// Xoá các client không hoạt động trong 5 phút
		for ip, client := range clients {
			if time.Since(client.Lastseen) > 5*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

// test:ab -n 20 -c 1 -H "X-API-KEY:87f2f6bd-8095-44d4-9295-547136178207" http://localhost:8080/api/v1/category/golang
func RateLimitMiddleware(rateLimiterLogger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// dùng ip xác đinh người dùng
		ip := getClientIP(c)
		//log.Println("ip:", ip) //::1 <=> 127.0.0.1
		limiter := getRateLimitter(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests, please try again later",
				"message": "Bạn đã gửi quá nhiều yêu cầu, vui lòng thử lại sau",
			})
			return
		}
		c.Next() // Call the next handler in the chain
	}
}
