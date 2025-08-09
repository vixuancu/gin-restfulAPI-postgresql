package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
	"user-management-api/internal/utils"

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
		requestSecStr := utils.GetEnv("RATE_LIMITER_REQUEST_SEC", "5")
		requestSec,err :=strconv.Atoi(requestSecStr)
		if err != nil {
			panic("Invalid RATE_LIMITER_REQUEST_SEC value in .env file:" + err.Error())
		}
		requestBurstStr := utils.GetEnv("RATE_LIMITER_REQUEST_BURST", "10")
		requestBurst,err := strconv.Atoi(requestBurstStr)
		if err != nil {
			panic("Invalid RATE_LIMITER_REQUEST_BURST value in .env file:" + err.Error())
		}

		newclient := &Client{
			Limiter:  rate.NewLimiter(rate.Limit(requestSec), requestBurst),
			Lastseen: time.Now(),
		}
		clients[ip] = newclient
		// log.Printf("a client[%s]-{limiter: %v, lastseen: %v} is created", ip, newclient.Limiter, newclient.Lastseen)
		return newclient.Limiter
	} 
		// Cập nhật thời gian cuối cùng thấy client
		// log.Printf("a client[%s]-{limiter: %v, lastseen: %v} is created", ip, client.Limiter, client.Lastseen)
		client.Lastseen = time.Now()

		return client.Limiter
	

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

// test:ab -n 20 -c 1 -H "X-API-KEY:87f2f6bd-8095-44d4-9295-547136178207" http://localhost:8080/api/v1/users
func RateLimitMiddleware(rateLimiterLogger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// dùng ip xác đinh người dùng
		ip := getClientIP(c)
		//log.Println("ip:", ip) //::1 <=> 127.0.0.1
		limiter := getRateLimitter(ip)
		if !limiter.Allow() {
			rateLimiterLogger.Warn().
				Str("method", c.Request.Method).             // Ghi phương thức HTTP(GET, POST, PUT, DELETE, v.v.)
				Str("path", c.Request.URL.Path).             // Ghi đường dẫn của request(ví dụ: /api/v1/users)
				Str("query", c.Request.URL.RawQuery).        // Ghi query string nếu có (ví dụ: ?page=1&limit=10)
				Str("client_ip", c.ClientIP()).              // Ghi địa chỉ IP của client
				Str("user_agent", c.Request.UserAgent()).    // Ghi user agent của client (trình duyệt, ứng dụng, v.v.)
				Str("referer", c.Request.Referer()).         // Ghi referer nếu có (trang trước đó mà client đã truy cập)
				Str("protocol", c.Request.Proto).            // Ghi giao thức HTTP (HTTP/1.1, HTTP/2, v.v.)
				Str("host", c.Request.Host).                 // Ghi host của request (ví dụ: example.com)
				Str("remote_address", c.Request.RemoteAddr). // nếu địa chỉ IP của client không được cung cấp bởi c.ClientIP()
				Str("request_uri", c.Request.RequestURI).    // Ghi toàn bộ URI của request (bao gồm query string)
				Interface("headers", c.Request.Header).      // Ghi tất cả các header của request
				Msg("Rate limit exceeded for client")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests, please try again later",
				"message": "Bạn đã gửi quá nhiều yêu cầu, vui lòng thử lại sau",
			})
			return
		}
		c.Next() // Call the next handler in the chain
	}
}
