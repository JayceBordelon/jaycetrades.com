package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jaycetrades.com/internal/config"
	"jaycetrades.com/internal/email"
	"jaycetrades.com/internal/sentiment"
	"jaycetrades.com/internal/templates"
	"jaycetrades.com/internal/trades"

	"github.com/robfig/cron/v3"
)

func main() {
	cfg := config.Load()

	if cfg.ResendAPIKey == "" {
		log.Fatal("RESEND_API_KEY is required")
	}
	if cfg.ClaudeAPIKey == "" {
		log.Fatal("ANTHROPIC_API_KEY is required")
	}
	if len(cfg.EmailRecipients) == 0 {
		log.Fatal("EMAIL_RECIPIENTS is required")
	}

	scraper := sentiment.NewScraper()
	analyzer := trades.NewAnalyzer(cfg.ClaudeAPIKey)
	emailClient := email.NewClient(cfg.ResendAPIKey)

	job := func() {
		runTradeAnalysis(cfg, scraper, analyzer, emailClient)
	}

	c := cron.New(cron.WithLocation(time.FixedZone("EST", -5*60*60)))
	_, err := c.AddFunc(cfg.CronSchedule, job)
	if err != nil {
		log.Fatalf("Failed to add cron job: %v", err)
	}

	c.Start()

	log.Printf("Options trade scanner started")
	log.Printf("Cron schedule: %s (EST)", cfg.CronSchedule)
	log.Printf("Emails will be sent to: %v", cfg.EmailRecipients)

	// Run immediately on startup if RUN_ON_START is set
	if os.Getenv("RUN_ON_START") == "true" {
		log.Println("Running initial analysis...")
		job()
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	c.Stop()
}

func runTradeAnalysis(cfg *config.Config, scraper *sentiment.Scraper, analyzer *trades.Analyzer, emailClient *email.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Println("Starting trade analysis...")

	// Get sentiment data from WSB
	log.Println("Scraping WSB sentiment...")
	sentimentData, err := scraper.GetTrendingTickers(ctx, 20)
	if err != nil {
		log.Printf("Error getting sentiment data: %v", err)
		return
	}
	log.Printf("Found %d trending tickers", len(sentimentData))

	// Get top 3 trades from Claude
	log.Println("Analyzing trades with Claude...")
	topTrades, err := analyzer.GetTopTrades(ctx, sentimentData)
	if err != nil {
		log.Printf("Error analyzing trades: %v", err)
		return
	}
	log.Printf("Generated %d trade recommendations", len(topTrades))

	// Convert to template trades
	templateTrades := make([]templates.Trade, len(topTrades))
	for i, t := range topTrades {
		templateTrades[i] = templates.Trade{
			Symbol:         t.Symbol,
			ContractType:   t.ContractType,
			StrikePrice:    t.StrikePrice,
			Expiration:     t.Expiration,
			DTE:            t.DTE,
			EstimatedPrice: t.EstimatedPrice,
			Thesis:         t.Thesis,
			SentimentScore: t.SentimentScore,
		}
	}

	// Render email template
	htmlContent, err := templates.RenderEmail(templateTrades)
	if err != nil {
		log.Printf("Error rendering email: %v", err)
		return
	}
	subject := fmt.Sprintf("Options Trades for %s", time.Now().Format("Monday, Jan 2"))

	log.Println("Sending email...")
	if err := emailClient.SendTradeEmail(cfg.EmailFrom, cfg.EmailRecipients, subject, htmlContent); err != nil {
		log.Printf("Error sending email: %v", err)
		return
	}

	log.Println("Trade analysis complete and email sent!")
}
