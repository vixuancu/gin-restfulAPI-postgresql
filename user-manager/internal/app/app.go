package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-management-api/internal/config"
	"user-management-api/internal/db"
	"user-management-api/internal/db/sqlc"
	"user-management-api/internal/routes"
	"user-management-api/internal/validation"
	"user-management-api/pkg/auth"
	"user-management-api/pkg/cache"
	"user-management-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Module interface {
	Routes() routes.Routes
}
type Application struct {
	config  *config.Config
	router  *gin.Engine
	modules []Module
}

type ModuleContext struct {
	DB    sqlc.Querier
	Redis *redis.Client
}

func NewApplication(cfg *config.Config) *Application {

	r := gin.Default()

	if err := validation.InitValidator(); err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Failed to initialize validator:")
	}
	if err := db.InitDB(); err != nil {
		logger.Log.Fatal().Err(err).Msg("❌ Failed to connect to database:")
	}
	redisClient := config.NewRedisClient()
	cacheService := cache.NewRedisCacheService(redisClient)
	tokenService := auth.NewJWTService(cacheService)
	ctx := &ModuleContext{
		DB:    db.DB,
		Redis: redisClient,
	}

	modules := []Module{
		NewUserModule(ctx),
		NewAuthModule(ctx, tokenService, cacheService),
	}
	routes.RegisterRoutes(r, tokenService, cacheService, GetModuleRoutes(modules)...)
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
		logger.Log.Info().Msgf("❤️ Starting server on %s", app.config.ServerAddress)
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("❌ ListenAndServe failed:")
		}
	}()

	<-quit // Chờ tín hiệu dừng
	logger.Log.Info().Msg("🍺 Shutdow signal receiver")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := svr.Shutdown(ctx); err != nil {
		logger.Log.Error().Err(err).Msg("⚠️ Server forced to shutdown:")
	}
	logger.Log.Info().Msg("🍺 Server exiting gracefully")
	return nil
}

func GetModuleRoutes(modules []Module) []routes.Routes {
	routesList := make([]routes.Routes, len(modules))
	for i, module := range modules {
		routesList[i] = module.Routes()
	}
	return routesList
}
