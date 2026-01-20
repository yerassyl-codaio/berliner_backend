//go:build wireinject
// +build wireinject

package main

import (
	"github.com/I1Asyl/berliner_backend/pkg/handler"
	"github.com/I1Asyl/berliner_backend/pkg/repository"
	"github.com/I1Asyl/berliner_backend/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// Config holds the application configuration
type Config struct {
	DSN string
}

// ProvideRepository creates a new repository instance
func ProvideRepository(config Config) *repository.Repository {
	return repository.NewRepository(config.DSN)
}

// ProvideServices creates a new services instance
func ProvideServices(repo *repository.Repository) *services.Services {
	return services.NewService(repo)
}

// ProvideHandler creates a new handler instance
func ProvideHandler(services *services.Services) *handler.Handler {
	return handler.NewHandler(services)
}

// ProvideRouter creates a new Gin router
func ProvideRouter(handler *handler.Handler) *gin.Engine {
	return handler.InitRouter()
}

// InitializeApp wires up all dependencies and returns the router
func InitializeApp(config Config) (*gin.Engine, error) {
	wire.Build(
		ProvideRepository,
		ProvideServices,
		ProvideHandler,
		ProvideRouter,
	)
	return nil, nil
}
