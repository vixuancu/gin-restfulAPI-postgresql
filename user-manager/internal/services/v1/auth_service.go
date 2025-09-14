package v1services

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/repository"
	"user-management-api/internal/utils"
	"user-management-api/pkg/auth"
	"user-management-api/pkg/cache"
	"user-management-api/pkg/email"
	emailpkg "user-management-api/pkg/email"
	"user-management-api/pkg/logger"
	"user-management-api/pkg/rabbitmq"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

type authService struct {
	userRepo     repository.UserRepository
	TokenService auth.TokenService
	cacheService cache.RedisCacheService
	mailService  emailpkg.EmailProviderService
	rabbitmq rabbitmq.RabbitMQService
}
type LoginAttempt struct {
	Limiter  *rate.Limiter
	Lastseen time.Time
}
type EmailPayload struct {
    TraceID  string       `json:"trace_id"`
    Email    *email.Email `json:"email"`
}

var (
	mu               sync.Mutex
	clients          = make(map[string]*LoginAttempt) // Lưu trữ client theo IP
	LoginAttemptTTL  = 5 * time.Minute                // 5 phút tương ứng với 5 token
	MaxLoginAttempts = 5                              // Số lần đăng nhập tối đa trong khoảng thời gian TTL
)

func NewAuthService(repo repository.UserRepository, TokenService auth.TokenService, cacheService cache.RedisCacheService, mailService email.EmailProviderService,rabbitmqService rabbitmq.RabbitMQService) AuthService {
	return &authService{
		userRepo:     repo,
		TokenService: TokenService,
		cacheService: cacheService,
		mailService:  mailService,
		rabbitmq: rabbitmqService,
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
		expUnix, _ := claims["exp"].(float64)    // Lấy thời gian hết hạn từ claims
		exp := time.Unix(int64(expUnix), 0)      // Chuyển đổi sang thời gian
		key := "blacklist:" + jti                // Tạo khóa blacklist
		ttl := time.Until(exp)                   // Tính thời gian hết hạn bằng thời gian còn lại của access token
		as.cacheService.Set(key, "revoked", ttl) // Lưu vào cache với thời gian hết hạn
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
	// Lấy trace_id từ request
    traceID := logger.GetTraceID(context)
	rateLimitKey := fmt.Sprintf("reset:ratelimit:%s", email)

	if exists, err := as.cacheService.Exists(rateLimitKey); err == nil && exists {
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

	err = as.cacheService.Set("reset:"+token, user.UserUuid, 1*time.Hour) // Lưu token vào cache với thời gian hết hạn 5 phút
	if err != nil {
		return utils.WrapError(err, "Failed to store forgot password", utils.ErrorCodeInternalServer)
	}
	err = as.cacheService.Set(rateLimitKey, "1", 5*time.Minute) // Lưu token vào cache với thời gian hết hạn 5 phút
	if err != nil {
		return utils.WrapError(err, "Failed to set rate limit key", utils.ErrorCodeInternalServer)
	}
	//view-to-reset-password là đường dẫn của frontend gửi đến để đặt lại mật khẩu
	resetLink := fmt.Sprintf("http://abc.com/view-to-reset-password?token=%s", token) // Tạo link reset password

	mailContent := &emailpkg.Email{
		To: []emailpkg.Address{
			{Email: email},
		},
		Subject:  "Password Reset Request",
		Text:     fmt.Sprintf("Hi %s,\n\nTo reset your password, please click the following link: \n%s\n\n The Link will expire in 1 hour.\n\nIf you did not request a password reset, please ignore this email.\n\nBest regards,\nCode Team", user.UserEmail, resetLink),
		Category: "password_reset",
	}

	// Cần update thành publish vào rabbitmq
	payloadEmail := EmailPayload{
		TraceID: traceID,
		Email:   mailContent,
	}
	if err := as.rabbitmq.Puclish(context,"auth_email_queue",payloadEmail); err != nil {
		return utils.NewError("Failed to send password reset email", utils.ErrorCodeInternalServer)
	} 

	return nil
}

func (as *authService) ResetPassword(c *gin.Context, token, newPassword string) error {
	context := c.Request.Context() // Lấy context của go từ gin.Context

	var userUUIDStr string
	err := as.cacheService.Get("reset:"+token, &userUUIDStr) // Lấy userUUID từ cache bằng token
	if err == redis.Nil || userUUIDStr == "" {
		return utils.NewError("Invalid or expired reset token", utils.ErrorCodeNotFound)
	}
	if err != nil {
		return utils.WrapError(err, "Failed to get user UUID from cache", utils.ErrorCodeInternalServer)
	}
	userUUID, err := uuid.Parse(userUUIDStr) // Chuyển đổi chuỗi UUID
	if err != nil {
		return utils.NewError("Invalid user UUID format", utils.ErrorCodeInternalServer)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost) // Mã hóa mật khẩu mới
	if err != nil {
		return utils.WrapError(err, "Failed to hash new password", utils.ErrorCodeInternalServer)
	}
	// Cập nhật mật khẩu mới cho người dùng
	var input = sqlc.UpdatePasswordParams{
		UserUuid:     userUUID,
		UserPassword: string(hashedPassword), // Chuyển đổi []byte sang string
	}
	_, err = as.userRepo.UpdatePassword(context, input)
	if err != nil {
		return utils.WrapError(err, "Failed to update password", utils.ErrorCodeInternalServer)
	}
	// Xóa token khỏi cache sau khi đặt lại mật khẩu thành công
	err = as.cacheService.Clear("reset:" + token)
	if err != nil {
		return utils.WrapError(err, "Failed to clear reset token from cache", utils.ErrorCodeInternalServer)
	}
	return nil
}
