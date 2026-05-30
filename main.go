// BarterSwap — API d'échange de compétences entre particuliers via crédit-temps.
//
// Composition root : parse la configuration, ouvre la connexion DB, applique les
// migrations, instancie les adapters (repos + handlers) et démarre le serveur HTTP
// avec graceful shutdown. Toute la logique métier vit dans usecase_*.go ; les
// adapters HTTP/SQL sont isolés dans http_*.go et repo_*.go.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := loadConfig()
	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	logger.Info("starting barterswap",
		"port", cfg.Port,
		"log_level", cfg.LogLevel,
	)

	db, err := openDB(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("database open failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(context.Background(), db, "migrations", logger); err != nil {
		logger.Error("migrations failed", "err", err)
		os.Exit(1)
	}

	app := &App{
		DB:     db,
		Logger: logger,
		Now:    time.Now,
	}

	handler := app.buildHTTPHandler()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := runServer(srv, logger); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
	logger.Info("bye")
}

// runServer démarre srv et attend SIGINT/SIGTERM pour déclencher un shutdown
// gracieux (5s de deadline).
func runServer(srv *http.Server, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl}))
}
