package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/utils"
	"user-management-api/pkg/cache"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	cache * cache.RedisCacheService
}

type EncryptedPayload struct {
	UserUUID string `json:"user_uuid"`
	Email    string `json:"email"`
	Role     int32  `json:"role"`
}
type RefreshToken struct {
	Token     string    `json:"token"`
	UserUUID  string    `json:"user_uuid"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

var (
	jwtSecret     = []byte(utils.GetEnv("JWT_SECRET", "secret-keysecret-keysecret-keyab")) // Lấy secret key từ biến môi trường
	jwtEncryptKey = []byte(utils.GetEnv("JWT_ENCRYPT_KEY", "12345678901234567890123456789012"))
)

const (
	AcessTokenTTL   = 24 * time.Hour      // Thời gian sống của access token
	RefreshTokenTTL = 30 * 24 * time.Hour // Thời gian sống của refresh token
)

func NewJWTService(cache * cache.RedisCacheService) TokenService {
	return &JWTService{
		cache: cache,
	}
}

/*88888888888888888888888888888888 Access Token 88888888888888888888888888888888*/
func (js *JWTService) GenerateAccessToken(user sqlc.User) (string, error) {
	payload := &EncryptedPayload{
		UserUUID: user.UserUuid.String(),
		Email:    user.UserEmail,
		Role:     user.UserLevel,
	}
	rawData, err := json.Marshal(payload) // Chuyển đổi payload thành JSON
	if err != nil {
		return "", err
	}
	encryptedData, err := utils.EncryptAES(rawData, jwtEncryptKey) // Mã hóa dữ liệu bằng AES
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"jti":  uuid.NewString(),                                  // ID duy nhất của token
		"data": encryptedData,                                     // Dữ liệu đã mã hóa
		"exp":  jwt.NewNumericDate(time.Now().Add(AcessTokenTTL)), // Thời gian hết hạn
		"iat":  jwt.NewNumericDate(time.Now()),                    // Thời gian phát hành
		"iss":  "vixuancu",                                        // Người phát hành token
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // Tạo token với phương thức mã hóa HS256
	return token.SignedString(jwtSecret)                       // Trả về chuỗi token đã ký
}

// hàm kiểm tra token có hợp lệ hay không
func (js *JWTService) ParseToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return jwtSecret, nil // Trả về secret key để xác thực token
	})
	if err != nil || !token.Valid {
		return nil, nil, utils.NewError("Invalid token", utils.ErrorCodeUnauthorized)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, utils.NewError("Invalid token claims", utils.ErrorCodeUnauthorized)
	}
	return token, claims, nil
}

func (js *JWTService) DecryptAccessTokenPayload(tokenString string) (*EncryptedPayload, error) {
	_, claims, err := js.ParseToken(tokenString)
	if err != nil {
		return nil, utils.WrapError(err, "cannot parse token", utils.ErrorCodeUnauthorized)
	}
	encryptedData, ok := claims["data"].(string)
	if !ok {
		return nil, utils.NewError("Invalid token data", utils.ErrorCodeUnauthorized)
	}
	decryptedByte, err := utils.DecryptAES(encryptedData, jwtEncryptKey)
	if err != nil {
		return nil, utils.NewError("Failed to decrypt token data", utils.ErrorCodeUnauthorized)
	}
	// Chuyển đổi dữ liệu đã giải mã thành EncryptedPayload (từ JSON sang struct)
	var payload EncryptedPayload
	if err := json.Unmarshal(decryptedByte, &payload); err != nil {
		return nil, utils.WrapError(err, "Failed to unmarshal token data", utils.ErrorCodeInternalServer)
	}
	return &payload, nil
}

/*88888888888888888888888888888888 Refresh Token 88888888888888888888888888888888*/

func (js *JWTService) GenerateRefreshToken(user sqlc.User) (RefreshToken, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return RefreshToken{}, err
	} // Sinh nonce ngẫu nhiên
	token := base64.URLEncoding.EncodeToString(tokenBytes) // Trả về chuỗi base64 của dữ liệu đã mã hóa

	return RefreshToken{
		Token: token,
		UserUUID:  user.UserUuid.String(),
		ExpiresAt: time.Now().Add(RefreshTokenTTL), // Thời gian hết
		Revoked:   false,                            // Chưa bị thu hồi
	}, nil
}

func (js *JWTService) StoreRefreshToken(token RefreshToken) error{
	cacheKey := "refresh_token:" + token.Token
	return js.cache.Set(cacheKey,token, RefreshTokenTTL)
}
