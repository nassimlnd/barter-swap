package main

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// newTestDB ouvre la base de test pointée par TEST_DATABASE_URL, applique le
// schéma et tronque toutes les tables. Si la variable n'est pas définie, le
// test est skip — c'est le compromis pragmatique du sujet : les tests qui
// touchent la DB sont opt-in, ceux qui ne la touchent pas tournent partout.
//
// Convention : à appeler en début de TOUT test qui parle SQL.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL non défini (test SQL skip)")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping test DB: %v", err)
	}

	silent := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := runMigrations(ctx, db, "migrations", silent); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	truncateAll(t, db)
	return db
}

// truncateAll vide les 6 tables et réinitialise les séquences. CASCADE pour
// gérer les FK. Idempotent : sûr à appeler en début de chaque test.
func truncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, `
        TRUNCATE
            users,
            skills,
            services,
            exchanges,
            credit_transactions,
            reviews
        RESTART IDENTITY CASCADE
    `)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

// newTestApp construit un *App câblé sur la base de test, avec un logger silent
// et une horloge fixe (déterministe pour les assertions sur created_at).
func newTestApp(t *testing.T) *App {
	t.Helper()
	db := newTestDB(t)
	usersRepo := NewPgUserRepo(db)
	skillsRepo := NewPgSkillRepo(db)
	creditsRepo := NewPgCreditRepo(db)
	servicesRepo := NewPgServiceRepo(db)
	exchangesRepo := NewPgExchangeRepo(db)
	reviewsRepo := NewPgReviewRepo(db)
	statsRepo := NewPgStatsRepo(db)
	return &App{
		DB:     db,
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Now:    func() time.Time { return time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC) },
		Users: &UserUsecase{
			Users:   usersRepo,
			Skills:  skillsRepo,
			Credits: creditsRepo,
			Tx:      NewPgTxRunner(db),
		},
		Services: &ServiceUsecase{
			Users:    usersRepo,
			Skills:   skillsRepo,
			Services: servicesRepo,
		},
		Exchanges: &ExchangeUsecase{
			Users:     usersRepo,
			Services:  servicesRepo,
			Exchanges: exchangesRepo,
			Credits:   creditsRepo,
			Tx:        NewPgTxRunner(db),
		},
		Reviews: &ReviewUsecase{
			Exchanges: exchangesRepo,
			Reviews:   reviewsRepo,
		},
		Stats: &StatsUsecase{
			Users: usersRepo,
			Stats: statsRepo,
		},
	}
}
