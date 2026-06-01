package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// openDB ouvre une connexion sql.DB en pgx, applique des limites raisonnables,
// puis attend que la base soit joignable (jusqu'à ~30s, utile au démarrage Docker).
func openDB(dsn string, logger *slog.Logger) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := waitForDB(db, logger); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func waitForDB(db *sql.DB, logger *slog.Logger) error {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error
	for attempt := 1; time.Now().Before(deadline); attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := db.PingContext(ctx)
		cancel()
		if err == nil {
			logger.Info("database ready", "attempt", attempt)
			return nil
		}
		lastErr = err
		logger.Warn("database not ready, retrying", "attempt", attempt, "err", err)
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("database unreachable after retries: %w", lastErr)
}

// runMigrations applique tous les fichiers .sql du dossier indiqué dans l'ordre
// lexicographique, sans système de versioning sophistiqué : chaque fichier doit
// être idempotent (CREATE TABLE IF NOT EXISTS, etc.) ou bien le schéma initial
// est joué une seule fois. Pour ce projet, on garde une seule migration "001_schema.sql"
// idempotente jouée à chaque démarrage.
func runMigrations(ctx context.Context, db *sql.DB, dir string, logger *slog.Logger) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("migrations directory missing, skipping", "dir", dir)
			return nil
		}
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if _, err := db.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("apply %s: %w", path, err)
		}
		logger.Info("migration applied", "file", name)
	}
	return nil
}
