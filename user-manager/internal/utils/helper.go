package utils

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"strconv"
	"user-management-api/pkg/logger"

	"github.com/rs/zerolog"
)

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}
func GetIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return valueInt
}

func NewLoggerWithPath(fileName string, level string) *zerolog.Logger {
	dir, err := os.Getwd() // Lấy đường dẫn làm việc hiện tại
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("❌Unable to get working dir")
	}
	path := filepath.Join(dir, "internal/logs", fileName) // Tạo đường dẫn đến file .env
	config := logger.LoggerConfig{
		Level:      level,
		Filename:   path,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     5,    //
		Compress:   true, // disabled by default
		IsDev:      GetEnv("APP_ENV", "development"),
	}
	return logger.NewLogger(config)
}

func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	} // Sinh nonce ngẫu nhiên
	return base64.URLEncoding.EncodeToString(bytes), nil // Trả về chuỗi base64 của dữ liệu đã mã hóa
}

func MustGetWorkingDir() string {
	dir, err := os.Getwd() // Lấy đường dẫn làm việc hiện tại
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Unable to get working dir")
	}
	return dir
}
