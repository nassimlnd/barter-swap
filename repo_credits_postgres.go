package main

import (
	"context"
	"database/sql"
)

type PgCreditRepo struct{ db *sql.DB }

func NewPgCreditRepo(db *sql.DB) *PgCreditRepo { return &PgCreditRepo{db: db} }

func (r *PgCreditRepo) Insert(ctx context.Context, tx *CreditTransaction) (int, error) {
	conn := dbConn(ctx, r.db)
	var id int
	err := conn.QueryRowContext(ctx, `
        INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, tx.UserID, tx.ExchangeID, tx.Montant, tx.Type).Scan(&id)
	if err != nil {
		return 0, mapPgError(err)
	}
	return id, nil
}

func (r *PgCreditRepo) ListByUser(ctx context.Context, userID int, limit, offset int) ([]CreditTransaction, error) {
	conn := dbConn(ctx, r.db)
	rows, err := conn.QueryContext(ctx, `
        SELECT id, user_id, exchange_id, montant, type, created_at
        FROM credit_transactions
        WHERE user_id = $1
        ORDER BY created_at DESC, id DESC
        LIMIT $2 OFFSET $3
    `, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CreditTransaction
	for rows.Next() {
		var tx CreditTransaction
		var exchangeID sql.NullInt64
		if err := rows.Scan(&tx.ID, &tx.UserID, &exchangeID, &tx.Montant, &tx.Type, &tx.CreatedAt); err != nil {
			return nil, err
		}
		if exchangeID.Valid {
			v := int(exchangeID.Int64)
			tx.ExchangeID = &v
		}
		out = append(out, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PgCreditRepo) SumByUser(ctx context.Context, userID int) (int, error) {
	conn := dbConn(ctx, r.db)
	var sum int
	if err := conn.QueryRowContext(ctx, `
        SELECT COALESCE(SUM(montant), 0)
        FROM credit_transactions
        WHERE user_id = $1
    `, userID).Scan(&sum); err != nil {
		return 0, err
	}
	return sum, nil
}

var _ CreditRepository = (*PgCreditRepo)(nil)
