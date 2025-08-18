package auth

import (
	"encoding/json"
	"time"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
}

type EncryptedPayload struct {
	UserUUID string `json:"user_uuid"`
	Email    string `json:"email"`
	Role     int32  `json:"role"`
}

var (
	jwtSecret     = []byte(utils.GetEnv("JWT_SECRET", "secret-keysecret-keysecret-keyab")) // Lấy secret key từ biến môi trường
	jwtEncryptKey = []byte(utils.GetEnv("JWT_ENCRYPT_KEY", "12345678901234567890123456789012"))
)

const (
	AcessTokenTTL = 24 * time.Hour // Thời gian sống của access token
)

func NewJWTService() *JWTService {
	return &JWTService{}
}

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

func (js *JWTService) GenerateRefreshToken() {

}
