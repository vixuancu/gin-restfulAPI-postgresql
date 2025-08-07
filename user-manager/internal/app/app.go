package app

import (
	"log"
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
	if err := app.router.Run(app.config.ServerAddress); err != nil {
		return err
	}
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
