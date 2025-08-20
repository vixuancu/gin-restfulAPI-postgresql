package main

import (
	"os"
	"path/filepath"
	"user-management-api/internal/app"
	"user-management-api/internal/config"
	"user-management-api/internal/utils"
	"user-management-api/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	rootDir := mustGetWorkingDir()
	logFile := filepath.Join(rootDir, "internal/logs/app.log")
	logger.InitLogger(logger.LoggerConfig{
		Level:      "info",
		Filename:   logFile,
		MaxSize:    1, // megabytes
		MaxBackups: 5,
		MaxAge:     5,    //
		Compress:   true, // disabled by default
		IsDev:      utils.GetEnv("APP_ENV", "development"),
	})
	loadEnv(filepath.Join(rootDir, ".env"))
	// Initialize configuration
	config := config.NewConfig()
	// init application
	application := app.NewApplication(config)

	// start server
	if err := application.Run(); err != nil {
		panic(err)
	}
}
func mustGetWorkingDir() string {
	dir, err := os.Getwd() // Lấy đường dẫn làm việc hiện tại
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Unable to get working dir")
	}
	return dir
}
func loadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		logger.Log.Warn().Msg("⚠️ No env file found")

	} else {
		logger.Log.Info().Msg("✅ Loaded successfully env file: ")
	}
}
