package main

import (
	"context"
	"database/sql"
	"fmt"
)

type PgUserRepo struct{ db *sql.DB }

func NewPgUserRepo(db *sql.DB) *PgUserRepo { return &PgUserRepo{db: db} }

func (r *PgUserRepo) Insert(ctx context.Context, u *User) (int, error) {
	conn := dbConn(ctx, r.db)
	var id int
	err := conn.QueryRowContext(ctx, `
        INSERT INTO users (pseudo, bio, ville, credit_balance)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, u.Pseudo, u.Bio, u.Ville, u.CreditBalance).Scan(&id)
	if err != nil {
		return 0, mapPgError(err)
	}
	return id, nil
}

func (r *PgUserRepo) GetByID(ctx context.Context, id int) (*User, error) {
	conn := dbConn(ctx, r.db)
	u := &User{}
	err := conn.QueryRowContext(ctx, `
        SELECT id, pseudo, bio, ville, credit_balance, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &u.CreditBalance, &u.CreatedAt)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return u, nil
}

func (r *PgUserRepo) UpdateProfile(ctx context.Context, id int, pseudo, bio, ville string) error {
	conn := dbConn(ctx, r.db)
	res, err := conn.ExecContext(ctx, `
        UPDATE users
        SET pseudo = $1, bio = $2, ville = $3
        WHERE id = $4
    `, pseudo, bio, ville, id)
	if err != nil {
		return mapPgError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PgUserRepo) AdjustBalance(ctx context.Context, userID, delta int) error {
	conn := dbConn(ctx, r.db)
	res, err := conn.ExecContext(ctx, `
        UPDATE users
        SET credit_balance = credit_balance + $1
        WHERE id = $2
    `, delta, userID)
	if err != nil {
		return mapPgError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

var _ UserRepository = (*PgUserRepo)(nil)
