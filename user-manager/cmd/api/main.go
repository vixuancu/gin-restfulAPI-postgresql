package main

import (
	"path/filepath"
	"user-management-api/internal/app"
	"user-management-api/internal/config"
	"user-management-api/internal/utils"
	"user-management-api/pkg/logger"

	"github.com/joho/godotenv"
)

func main() {
	rootDir := utils.MustGetWorkingDir()
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

	if err := godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		logger.Log.Warn().Msg("⚠️ No env file found")

	} else {
		logger.Log.Info().Msg("✅ Loaded successfully env in api file: ")
	}
	// Initialize configuration
	config := config.NewConfig()
	// init application
	application,err := app.NewApplication(config)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Unable to initialize application")
	}
	// start server
	if err := application.Run(); err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Unable to start server")
	}
}

