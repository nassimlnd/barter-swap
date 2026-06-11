package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type PgExchangeRepo struct{ db *sql.DB }

func NewPgExchangeRepo(db *sql.DB) *PgExchangeRepo { return &PgExchangeRepo{db: db} }

func (r *PgExchangeRepo) Insert(ctx context.Context, e *Exchange) (int, error) {
	conn := dbConn(ctx, r.db)
	var id int
	err := conn.QueryRowContext(ctx, `
        INSERT INTO exchanges (service_id, requester_id, owner_id, credits, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `, e.ServiceID, e.RequesterID, e.OwnerID, e.Credits, e.Status).Scan(&id)
	if err != nil {
		return 0, mapPgError(err)
	}
	return id, nil
}

func (r *PgExchangeRepo) GetByID(ctx context.Context, id int) (*Exchange, error) {
	conn := dbConn(ctx, r.db)
	e := &Exchange{}
	err := conn.QueryRowContext(ctx, `
        SELECT id, service_id, requester_id, owner_id, credits, status, created_at, updated_at
        FROM exchanges
        WHERE id = $1
    `, id).Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Credits, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return e, nil
}

func (r *PgExchangeRepo) UpdateStatus(ctx context.Context, id int, from, to ExchangeStatus) error {
	conn := dbConn(ctx, r.db)
	res, err := conn.ExecContext(ctx, `
        UPDATE exchanges
        SET status = $1, updated_at = now()
        WHERE id = $2 AND status = $3
    `, to, id, from)
	if err != nil {
		return mapPgError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrConflict
	}
	return nil
}

func (r *PgExchangeRepo) List(ctx context.Context, f ExchangeFilter) ([]Exchange, int, error) {
	conn := dbConn(ctx, r.db)
	where, args := buildExchangeWhere(f)

	var total int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM exchanges `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	query := `
        SELECT id, service_id, requester_id, owner_id, credits, status, created_at, updated_at
        FROM exchanges ` + where + `
        ORDER BY updated_at DESC, id DESC
        LIMIT $` + fmt.Sprint(len(args)-1) + ` OFFSET $` + fmt.Sprint(len(args))
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Exchange
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Credits, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func buildExchangeWhere(f ExchangeFilter) (string, []any) {
	var clauses []string
	var args []any
	if f.UserID > 0 {
		args = append(args, f.UserID)
		clauses = append(clauses, fmt.Sprintf("(requester_id = $%d OR owner_id = $%d)", len(args), len(args)))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

var _ ExchangeRepository = (*PgExchangeRepo)(nil)
