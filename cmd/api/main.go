package main

import (
	"log"

	"jaycetrades.com/internal/config"
	"jaycetrades.com/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	routes.Setup(r)

	log.Printf("Starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
