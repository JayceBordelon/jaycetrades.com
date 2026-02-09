package main

import (
	"log"

	"jaycetrades.com/internal/config"
	"jaycetrades.com/internal/routes"
	"jaycetrades.com/internal/schwab"
	"jaycetrades.com/internal/sentiment"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Validate required config
	if cfg.SchwabClientID == "" || cfg.SchwabSecret == "" {
		log.Println("WARNING: SCHWAB_CLIENT_ID and SCHWAB_CLIENT_SECRET not set")
		log.Println("Set these environment variables to enable Schwab API integration")
	}

	// Initialize Schwab client
	schwabClient := schwab.NewClient(schwab.Credentials{
		ClientID:     cfg.SchwabClientID,
		ClientSecret: cfg.SchwabSecret,
		RedirectURI:  cfg.SchwabCallback,
	})

	// Initialize sentiment scraper
	scraper := sentiment.NewScraper()

	// Setup Gin
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	routes.Setup(r, schwabClient, scraper)

	log.Printf("Starting options discovery agent on :%s", cfg.Port)
	log.Printf("Schwab callback URL: %s", cfg.SchwabCallback)
	log.Println("Endpoints:")
	log.Println("  GET  /auth/login    - Start Schwab OAuth")
	log.Println("  GET  /auth/callback - OAuth callback")
	log.Println("  GET  /auth/status   - Check auth status")
	log.Println("  GET  /api/v1/trending - Get trending tickers")
	log.Println("  POST /api/v1/scan   - Run options scan")

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
