package main

import "os"

// Config concentre la configuration injectée par l'environnement.
type Config struct {
	DatabaseURL string
	Port        string
	LogLevel    string
}

func loadConfig() Config {
	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://barter:barter@localhost:5432/barterswap?sslmode=disable"),
		Port:        getenv("PORT", "8080"),
		LogLevel:    getenv("LOG_LEVEL", "info"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
