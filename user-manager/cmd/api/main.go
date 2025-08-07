package main

import (
	"user-management-api/internal/app"
	"user-management-api/internal/config"
)

func main() {
	// Initialize configuration
	config := config.NewConfig()
	// init application
	application := app.NewApplication(config)

	// start server
	if err := application.Run(); err != nil {
		panic(err)
	}
}
