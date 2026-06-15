// repo_tx_postgres.go — adapter SQL : implémentation TxRunner pour Postgres.
//
// Couche : adapters (Clean Architecture, out). Imports : context, database/sql.
// Aucune dépendance vers usecase_*.go ni http_*.go.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// PgTxRunner implémente TxRunner via *sql.DB.
type PgTxRunner struct {
	db *sql.DB
}

// NewPgTxRunner construit un TxRunner attaché à db.
func NewPgTxRunner(db *sql.DB) *PgTxRunner {
	return &PgTxRunner{db: db}
}

// txCtxKey est le type local utilisé comme clé de contexte pour transporter
// la *sql.Tx en cours. Type non exporté → pas de collision possible.
type txCtxKey struct{}

// InTx exécute fn dans une transaction. Si une Tx est déjà présente dans le
// contexte (nesting), elle est réutilisée — fn s'exécute alors sans rollback
// propre (la Tx racine commit/rollback). Sinon une nouvelle Tx est ouverte,
// commit en cas de succès, rollback sinon.
//
// L'erreur retournée par fn est propagée telle quelle (utile pour errors.Is).
func (r *PgTxRunner) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := txFromContext(ctx); ok {
		// Nesting : on partage la Tx racine.
		return fn(ctx)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Garantit le rollback en cas de panic ou de retour d'erreur. Le commit
	// explicite plus bas met la Tx dans un état où Rollback devient un no-op.
	defer func() {
		_ = tx.Rollback()
	}()

	ctx = context.WithValue(ctx, txCtxKey{}, tx)
	if err := fn(ctx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// txFromContext extrait la *sql.Tx présente en contexte, ou (nil, false).
func txFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(*sql.Tx)
	return tx, ok
}

// dbExecutor — interface minimale satisfaite à la fois par *sql.DB et *sql.Tx.
// Permet aux repos d'utiliser la même API que la Tx soit présente ou non.
type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// dbConn renvoie la Tx si présente en contexte, sinon le fallback (typiquement
// la *sql.DB du repo). À utiliser au début de chaque méthode SQL :
//
//	conn := dbConn(ctx, r.db)
//	rows, err := conn.QueryContext(ctx, "...", args...)
func dbConn(ctx context.Context, fallback dbExecutor) dbExecutor {
	if tx, ok := txFromContext(ctx); ok {
		return tx
	}
	return fallback
}

// errNoRows simplifie le mapping sql.ErrNoRows -> ErrNotFound côté repo.
func mapNoRows(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
