package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

// helper : insère un user minimal directement en SQL (sans passer par le repo
// pour découpler du jalon 5).
func insertRawUser(t *testing.T, db dbExecutor, pseudo string) int {
	t.Helper()
	var id int
	err := db.QueryRowContext(context.Background(),
		`INSERT INTO users (pseudo) VALUES ($1) RETURNING id`, pseudo).Scan(&id)
	if err != nil {
		t.Fatalf("insertRawUser: %v", err)
	}
	return id
}

func countUsers(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		t.Fatalf("countUsers: %v", err)
	}
	return n
}

func TestPgTxRunner_Commit(t *testing.T) {
	db := newTestDB(t)
	runner := NewPgTxRunner(db)

	err := runner.InTx(context.Background(), func(ctx context.Context) error {
		// Vérifie que la Tx est bien présente en contexte
		if _, ok := txFromContext(ctx); !ok {
			t.Fatal("Tx absente du contexte dans fn")
		}
		// Vérifie que dbConn renvoie la Tx, pas la DB
		conn := dbConn(ctx, db)
		if _, isTx := conn.(*sql.Tx); !isTx {
			t.Fatalf("dbConn doit retourner *sql.Tx pendant InTx, got %T", conn)
		}
		_, err := conn.ExecContext(ctx, `INSERT INTO users (pseudo) VALUES ($1)`, "alice")
		return err
	})
	if err != nil {
		t.Fatalf("InTx commit: %v", err)
	}

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("attendu 1 user après commit, got %d", got)
	}
}

func TestPgTxRunner_RollbackOnError(t *testing.T) {
	db := newTestDB(t)
	runner := NewPgTxRunner(db)

	sentinel := errors.New("simulated failure")
	err := runner.InTx(context.Background(), func(ctx context.Context) error {
		conn := dbConn(ctx, db)
		if _, err := conn.ExecContext(ctx, `INSERT INTO users (pseudo) VALUES ($1)`, "bob"); err != nil {
			t.Fatalf("insert pendant tx: %v", err)
		}
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("attendu sentinel propagée, got %v", err)
	}
	if got := countUsers(t, db); got != 0 {
		t.Fatalf("attendu 0 user après rollback, got %d", got)
	}
}

func TestPgTxRunner_RollbackOnPanic(t *testing.T) {
	db := newTestDB(t)
	runner := NewPgTxRunner(db)

	defer func() {
		// On laisse remonter le panic au test runner ; on vérifie seulement
		// que la table est restée vide (rollback bien appelé via defer).
		_ = recover()
		if got := countUsers(t, db); got != 0 {
			t.Fatalf("attendu 0 user après rollback panic, got %d", got)
		}
	}()

	_ = runner.InTx(context.Background(), func(ctx context.Context) error {
		conn := dbConn(ctx, db)
		_, _ = conn.ExecContext(ctx, `INSERT INTO users (pseudo) VALUES ($1)`, "charlie")
		panic("boom")
	})
}

func TestPgTxRunner_Nesting(t *testing.T) {
	db := newTestDB(t)
	runner := NewPgTxRunner(db)

	err := runner.InTx(context.Background(), func(outerCtx context.Context) error {
		outerTx, _ := txFromContext(outerCtx)
		_, err := dbConn(outerCtx, db).ExecContext(outerCtx,
			`INSERT INTO users (pseudo) VALUES ($1)`, "outer")
		if err != nil {
			return err
		}
		// Imbrication : doit réutiliser la Tx racine
		return runner.InTx(outerCtx, func(innerCtx context.Context) error {
			innerTx, _ := txFromContext(innerCtx)
			if innerTx != outerTx {
				t.Fatal("inner Tx doit être identique à outer (nesting partagé)")
			}
			_, err := dbConn(innerCtx, db).ExecContext(innerCtx,
				`INSERT INTO users (pseudo) VALUES ($1)`, "inner")
			return err
		})
	})
	if err != nil {
		t.Fatalf("nested commit: %v", err)
	}
	if got := countUsers(t, db); got != 2 {
		t.Fatalf("attendu 2 users après nested commit, got %d", got)
	}
}

func TestDbConn_FallbackOutsideTx(t *testing.T) {
	db := newTestDB(t)
	conn := dbConn(context.Background(), db)
	if _, isTx := conn.(*sql.Tx); isTx {
		t.Fatal("dbConn doit retourner le fallback (DB) hors transaction")
	}
}

func TestMapNoRows(t *testing.T) {
	t.Parallel()
	if !errors.Is(mapNoRows(sql.ErrNoRows), ErrNotFound) {
		t.Fatal("sql.ErrNoRows doit être mappé vers ErrNotFound")
	}
	other := errors.New("autre")
	if !errors.Is(mapNoRows(other), other) {
		t.Fatal("autre erreur doit être propagée telle quelle")
	}
	if mapNoRows(nil) != nil {
		t.Fatal("nil doit rester nil")
	}
}
