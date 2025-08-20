package v1services

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"
	"user-management-api/pkg/auth"
	"user-management-api/pkg/cache"
	"user-management-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

type authService struct {
	userRepo     repository.UserRepository
	TokenService auth.TokenService
	cacheService        cache.RedisCacheService
}
type LoginAttempt struct {
	Limiter  *rate.Limiter
	Lastseen time.Time
}

var (
	mu               sync.Mutex
	clients          = make(map[string]*LoginAttempt) // Lưu trữ client theo IP
	LoginAttemptTTL  = 5 * time.Minute                // 5 phút tương ứng với 5 token
	MaxLoginAttempts = 5                              // Số lần đăng nhập tối đa trong khoảng thời gian TTL
)

func NewAuthService(repo repository.UserRepository, TokenService auth.TokenService, cacheService cache.RedisCacheService) *authService {
	return &authService{
		userRepo:     repo,
		TokenService: TokenService,
		cacheService:        cacheService,
	}
}

func (as *authService) getClientIP(c *gin.Context) string {
	// Lấy IP từ header X-Forwarded-For nếu có, nếu không thì lấy IP thực
	ip := c.ClientIP()
	if ip == "" {
		ip = c.Request.RemoteAddr
	}
	return ip
}

// Lấy ra rate limiter cho client theo IP
func (as *authService) getLoginAttempt(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	client, exists := clients[ip]
	if !exists {

		// rate.Limit đơn vị là số request/giây nên phải .Seconds() để chuyển đổi sang giây
		newclient := &LoginAttempt{
			Limiter:  rate.NewLimiter(rate.Limit(float32(MaxLoginAttempts)/float32(LoginAttemptTTL.Seconds())), MaxLoginAttempts),
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
func (as *authService) CleanupClients(ip string) {
	mu.Lock()
	defer mu.Unlock()
	delete(clients, ip) // Xóa client khỏi map
}

func (as *authService) CheckLoginAttempt(c *gin.Context) error {
	ip := as.getClientIP(c)           // Lấy IP của client
	limiter := as.getLoginAttempt(ip) // Lấy rate limiter cho client theo IP

	if !limiter.Allow() {
		return utils.NewError("Too many login attempts, please try again later", utils.ErrorCodeTooManyRequests)
	}
	return nil
}

// authentication là xác thực người dùng
// authorization là phân quyền người dùng
func (as *authService) Login(c *gin.Context, email, password string) (string, string, int, error) {
	context := c.Request.Context() // Lấy context của go từ gin.Context
	ip := as.getClientIP(c)        // Lấy IP của client

	if err := as.CheckLoginAttempt(c); err != nil {
		return "", "", 0, err
	}

	email = utils.NormalizeString(email)

	user, err := as.userRepo.GetByEmail(context, email)
	if err != nil {
		as.getLoginAttempt(ip) // Lấy rate limiter cho client theo IP
		return "", "", 0, utils.NewError("Invalid email or password", utils.ErrorCodeUnauthorized)
	}
	// Kiểm tra mật khẩu
	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPassword), []byte(password)); err != nil {
		as.getLoginAttempt(ip) // Lấy rate limiter cho client theo IP
		return "", "", 0, utils.NewError("Invalid email or password", utils.ErrorCodeUnauthorized)
	}
	acesstoken, err := as.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate access token", utils.ErrorCodeInternalServer)
	}
	refreshtoken, err := as.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate refresh token", utils.ErrorCodeInternalServer)
	}

	if err := as.TokenService.StoreRefreshToken(refreshtoken); err != nil {
		return "", "", 0, utils.NewError("Cannot Save refresh Token in redis", utils.ErrorCodeInternalServer)
	}

	as.CleanupClients(ip) // Xóa client khỏi map sau khi đăng nhập thành công
	return acesstoken, refreshtoken.Token, int(auth.AcessTokenTTL.Seconds()), nil
}

func (as *authService) Logout(c *gin.Context, refreshTokenString string) error {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.NewError("Authorization header missing ", utils.ErrorCodeUnauthorized) // Trả về lỗi nếu không có header
	}
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")  // Lấy token từ header
	_, claims, err := as.TokenService.ParseToken(accessToken) // Phân tích token
	if err != nil {
		return utils.NewError("Invalid accress token ", utils.ErrorCodeUnauthorized) // Trả về lỗi nếu không có header
	}
	// lấy jti từ claims
	if jti, ok := claims["jti"].(string); ok {
		expUnix, _ := claims["exp"].(float64) // Lấy thời gian hết hạn từ claims
		exp := time.Unix(int64(expUnix), 0)   // Chuyển đổi sang thời gian
		key := "blacklist:" + jti             // Tạo khóa blacklist
		ttl := time.Until(exp)                // Tính thời gian hết hạn bằng thời gian còn lại của access token
		as.cacheService.Set(key, "revoked", ttl)     // Lưu vào cache với thời gian hết hạn
	}

	// Vô hiệu hóa refresh token
	if _, err := as.TokenService.ValidateRefreshToken(refreshTokenString); err != nil {
		return utils.WrapError(err, "Invalid refresh token or revoked", utils.ErrorCodeUnauthorized)
	}
	if err := as.TokenService.RevokedRefreshToken(refreshTokenString); err != nil {
		return utils.NewError("Failed to revoke old refresh token", utils.ErrorCodeInternalServer)
	}
	return nil
}
func (as *authService) RefreshToken(c *gin.Context, refreshTokenString string) (string, string, int, error) {
	context := c.Request.Context() // Lấy context của go từ gin.Context
	token, err := as.TokenService.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", 0, utils.WrapError(err, "Invalid refresh token or revoked", utils.ErrorCodeUnauthorized)
	}
	// Kiểm tra refresh token. Trả thông tin user_uuid
	userUUID, _ := uuid.Parse(token.UserUUID)

	// Lấy thông tin user
	user, err := as.userRepo.GetByUuid(context, userUUID)
	if err != nil {
		return "", "", 0, utils.WrapError(err, "User not found", utils.ErrorCodeNotFound)
	}
	// Tạo access token mới
	acesstoken, err := as.TokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate access token", utils.ErrorCodeInternalServer)
	}
	// Tạo refresh token mới
	refreshtoken, err := as.TokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.NewError("Failed to generate refresh token", utils.ErrorCodeInternalServer)
	}
	// Vô hiệu hóa refresh token cũ
	if err := as.TokenService.RevokedRefreshToken(refreshTokenString); err != nil {
		return "", "", 0, utils.NewError("Failed to revoke old refresh token", utils.ErrorCodeInternalServer)
	}
	// Lưu refresh token mới vào Redis
	if err := as.TokenService.StoreRefreshToken(refreshtoken); err != nil {
		return "", "", 0, utils.NewError("Cannot Save refresh Token in redis", utils.ErrorCodeInternalServer)
	}
	return acesstoken, refreshtoken.Token, int(auth.AcessTokenTTL.Seconds()), nil
}

func (as *authService) ForgotPassword(c *gin.Context, email string) error {
	context := c.Request.Context() // Lấy context của go từ gin.Context

	rateLimitKey := fmt.Sprintf("reset:ratelimit:%s", email)

	if exists,err := as.cacheService.Exists(rateLimitKey); err == nil && exists {
		return utils.NewError("You have already requested a password reset. Please try again later.", utils.ErrorCodeTooManyRequests)
	}

	user, err := as.userRepo.GetByEmail(context, email)
	if err != nil {
		return utils.NewError("Invalid email or password", utils.ErrorCodeUnauthorized)
	}
	token, err := utils.GenerateRandomString(16) // Tạo chuỗi ngẫu nhiên để làm token reset password
	if err != nil {
		return utils.WrapError(err, "Failed to generate reset password token", utils.ErrorCodeInternalServer)
	}

	err =as.cacheService.Set("reset:"+token,user.UserUuid, 1*time.Hour) // Lưu token vào cache với thời gian hết hạn 5 phút
	if err != nil {
		return utils.WrapError(err, "Failed to store forgot password", utils.ErrorCodeInternalServer)
	}
	err =as.cacheService.Set(rateLimitKey,"1", 5*time.Minute) // Lưu token vào cache với thời gian hết hạn 5 phút
	if err != nil {
		return utils.WrapError(err, "Failed to set rate limit key", utils.ErrorCodeInternalServer)
	}
	//view-to-reset-password là đường dẫn của frontend gửi đến để đặt lại mật khẩu
	resetLink := fmt.Sprintf("http://abc.com/view-to-reset-password?token=%s", token) // Tạo link reset password
	// link này sẽ được gửi đến email của người dùng
	logger.Log.Info().Msg(resetLink) // Log link reset password


	return nil
}
