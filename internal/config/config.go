package config

import (
	"os"
	"strings"
)

type Config struct {
	CronSchedule    string   // Cron expression (e.g., "0 9 * * 1-5" for 9am weekdays)
	ResendAPIKey    string
	ClaudeAPIKey    string
	EmailRecipients []string // Comma-separated list of recipients
	EmailFrom       string
}

func Load() *Config {
	cronSchedule := os.Getenv("CRON_SCHEDULE")
	if cronSchedule == "" {
		cronSchedule = "25 9 * * 1-5" // Default: 9:25am EST (5 min before market open)
	}

	emailFrom := os.Getenv("EMAIL_FROM")
	if emailFrom == "" {
		emailFrom = "trades@jaycetrades.com"
	}

	var recipients []string
	if r := os.Getenv("EMAIL_RECIPIENTS"); r != "" {
		for _, email := range strings.Split(r, ",") {
			if trimmed := strings.TrimSpace(email); trimmed != "" {
				recipients = append(recipients, trimmed)
			}
		}
	}

	return &Config{
		CronSchedule:    cronSchedule,
		ResendAPIKey:    os.Getenv("RESEND_API_KEY"),
		ClaudeAPIKey:    os.Getenv("ANTHROPIC_API_KEY"),
		EmailRecipients: recipients,
		EmailFrom:       emailFrom,
	}
}
