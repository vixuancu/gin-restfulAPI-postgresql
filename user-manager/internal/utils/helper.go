package utils

import (
	"log"
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
		log.Fatal("❌Unable to get working dir:", err)
	}
	path := filepath.Join(dir,"internal/logs",fileName) // Tạo đường dẫn đến file .env
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
