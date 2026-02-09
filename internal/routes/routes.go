package routes

import (
	"jaycetrades.com/internal/handlers"
	"jaycetrades.com/internal/middleware"
	"jaycetrades.com/internal/schwab"
	"jaycetrades.com/internal/sentiment"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, schwabClient *schwab.Client, scraper *sentiment.Scraper) {
	// Global middleware
	r.Use(middleware.Logger())

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(schwabClient)
	scanHandler := handlers.NewScanHandler(schwabClient, scraper)

	// Health routes
	r.GET("/ping", healthHandler.Ping)

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.GET("/login", authHandler.Login)
		auth.GET("/callback", authHandler.Callback)
		auth.GET("/status", authHandler.Status)
	}

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Scanning endpoints
		v1.POST("/scan", scanHandler.Scan)
		v1.GET("/scan", scanHandler.Scan)
		v1.GET("/trending", scanHandler.GetTrending)
	}
}
