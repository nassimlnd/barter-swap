package main

import (
	"log/slog"
	"os"
	"testing"
)

func TestGetenv(t *testing.T) {
	t.Setenv("BARTERSWAP_TEST_ENV", "value")
	if got := getenv("BARTERSWAP_TEST_ENV", "fallback"); got != "value" {
		t.Fatalf("getenv existing=%q", got)
	}
	os.Unsetenv("BARTERSWAP_TEST_MISSING")
	if got := getenv("BARTERSWAP_TEST_MISSING", "fallback"); got != "fallback" {
		t.Fatalf("getenv missing=%q", got)
	}
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("PORT", "9999")
	t.Setenv("LOG_LEVEL", "debug")
	cfg := loadConfig()
	if cfg.DatabaseURL != "postgres://x" || cfg.Port != "9999" || cfg.LogLevel != "debug" {
		t.Fatalf("bad config: %+v", cfg)
	}
}

func TestNewLogger(t *testing.T) {
	cases := []string{"debug", "info", "warn", "error", "weird"}
	for _, level := range cases {
		t.Run(level, func(t *testing.T) {
			logger := newLogger(level)
			if logger == nil {
				t.Fatal("logger nil")
			}
			logger.LogAttrs(t.Context(), slog.LevelInfo, "test")
		})
	}
}
