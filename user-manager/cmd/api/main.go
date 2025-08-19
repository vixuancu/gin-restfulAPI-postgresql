package main

import (
	"log"
	"os"
	"path/filepath"
	"user-management-api/internal/app"
	"user-management-api/internal/config"

	"github.com/joho/godotenv"
)

func main() {
	loadEnv()
	// Initialize configuration
	config := config.NewConfig()
	// init application
	application := app.NewApplication(config)

	// start server
	if err := application.Run(); err != nil {
		panic(err)
	}
}

func loadEnv() {
	dir, err := os.Getwd() // Lấy đường dẫn làm việc hiện tại
	if err != nil {
		log.Fatal("❌Unable to get working dir:", err)
	}
	envPath := filepath.Join(dir, ".env") // Tạo đường dẫn đến file .env
	err = godotenv.Load(envPath)          // Load environment variables from .env file
	if err != nil {
		log.Println("⚠️Error loading .env file")
	} else {
		log.Println("✅Loaded environment variables from .env file")
	}
}
