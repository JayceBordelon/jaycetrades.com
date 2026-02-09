package routes

import (
	"jaycetrades.com/internal/handlers"
	"jaycetrades.com/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine) {
	// Global middleware
	r.Use(middleware.Logger())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	userHandler := handlers.NewUserHandler()

	// Health routes
	r.GET("/ping", healthHandler.Ping)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.POST("", userHandler.Create)
		}
	}
}
