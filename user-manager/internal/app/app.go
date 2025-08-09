package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-management-api/internal/config"
	"user-management-api/internal/routes"
	"user-management-api/internal/validation"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Module interface {
	Routes() routes.Routes
}
type Application struct {
	config  *config.Config
	router  *gin.Engine
	modules []Module
}

func NewApplication(cfg *config.Config) *Application {
	
	r := gin.Default()
	loadEnv()
	if err := validation.InitValidator(); err != nil {
		log.Fatal("Failed to initialize validator:", err)
	}
	modules := []Module{
		NewUserModule(),
	}
	routes.RegisterRoutes(r, GetModuleRoutes(modules)...)
	return &Application{
		config:  cfg,
		router:  r,
		modules: modules,
	}
}
func (app *Application) Run() error {
	// if err := app.router.Run(app.config.ServerAddress); err != nil {
	// 	return err
	// }
	// comment bằng Tiếng Việt
	svr := &http.Server{
		Addr:    app.config.ServerAddress,
		Handler: app.router,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP) // khi nhấn Ctrl+C hoặc dừng server hoắc reload

	// Chạy server trong một goroutine vì để tránh blocking
	go func() {
		log.Printf("❤️ Starting server on %s", app.config.ServerAddress)
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe failed: %v", err)
		}
	}()

	<- quit // Chờ tín hiệu dừng
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		log.Fatalf("✅Server forced to shutdown: %v", err)
	}
	log.Println("🍺Server exiting gracefully")
	return nil
}

func GetModuleRoutes(modules []Module) []routes.Routes {
	routesList := make([]routes.Routes, len(modules))
	for i, module := range modules {
		routesList[i] = module.Routes()
	}
	return routesList
}
func loadEnv() {
	err := godotenv.Load("../../.env") // Load environment variables from .env file
	if err != nil {
		log.Println("Error loading .env file")
	}
}
