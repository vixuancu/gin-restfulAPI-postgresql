package app

import (
	v1handler "user-management-api/internal/handler/v1"
	"user-management-api/internal/repository"
	"user-management-api/internal/routes"
	V1routes "user-management-api/internal/routes/v1"
	v1services "user-management-api/internal/services/v1"
)

type UserModule struct {
	routes routes.Routes
}

func NewUserModule() *UserModule {
	// Initialize repository
	userRepo := repository.NewSqlUserRepository()

	// Initialize service
	userService := v1services.NewUserService(userRepo)

	// Initialize handler
	userHandler := v1handler.NewUserHandler(userService)

	// Initialize routes
	userRoutes := 	V1routes.NewUserRoutes(userHandler)

	return &UserModule{
		routes: userRoutes,
	}
}
func (um *UserModule) Routes() routes.Routes {
	return um.routes
}
