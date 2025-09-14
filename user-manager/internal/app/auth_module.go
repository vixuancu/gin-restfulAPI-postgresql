package app

import (
	v1handler "user-management-api/internal/handler/v1"
	"user-management-api/internal/repository"
	"user-management-api/internal/routes"
	V1routes "user-management-api/internal/routes/v1"
	v1services "user-management-api/internal/services/v1"
	"user-management-api/pkg/auth"
	"user-management-api/pkg/cache"
	"user-management-api/pkg/email"
	"user-management-api/pkg/rabbitmq"
)

type AuthModule struct {
	routes routes.Routes
}

func NewAuthModule(ctx *ModuleContext, tokenService auth.TokenService, cache cache.RedisCacheService, mailService email.EmailProviderService, rabbitmqService rabbitmq.RabbitMQService ) *AuthModule {
	// Initialize repository
	userRepo := repository.NewSqlUserRepository(ctx.DB)

	// Initialize service
	authService := v1services.NewAuthService(userRepo, tokenService, cache, mailService,rabbitmqService)

	// Initialize handler
	authHandler := v1handler.NewAuthHandler(authService)

	// Initialize routes
	authRoutes := V1routes.NewAuthRoutes(authHandler)

	return &AuthModule{
		routes: authRoutes,
	}
}
func (um *AuthModule) Routes() routes.Routes {
	return um.routes
}
