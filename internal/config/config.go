package config

import "os"

type Config struct {
	Port           string
	Env            string
	SchwabClientID string
	SchwabSecret   string
	SchwabCallback string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	callback := os.Getenv("SCHWAB_CALLBACK_URL")
	if callback == "" {
		callback = "https://localhost:8080/auth/callback"
	}

	return &Config{
		Port:           port,
		Env:            env,
		SchwabClientID: os.Getenv("SCHWAB_CLIENT_ID"),
		SchwabSecret:   os.Getenv("SCHWAB_CLIENT_SECRET"),
		SchwabCallback: callback,
	}
}
