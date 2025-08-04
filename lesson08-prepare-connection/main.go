package main

import (
	"fmt"
	"hoc-gin/internal/db"
	"hoc-gin/internal/handlers"
	"hoc-gin/internal/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	db.InitDB()
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println(db.DB)
	r := gin.Default()
	userRepo := repository.NewSqlUserRepository()
	userHandler := handlers.NewUserHandler(userRepo)
	r.GET("/api/v1/users/:id", userHandler.GetUserById)
	r.POST("/api/v1/users", userHandler.CreateUser)
	r.Run(":8080") // listen and serve on
}
