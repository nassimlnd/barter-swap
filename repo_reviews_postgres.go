package main

import (
	"context"
	"database/sql"
)

type PgReviewRepo struct{ db *sql.DB }

func NewPgReviewRepo(db *sql.DB) *PgReviewRepo { return &PgReviewRepo{db: db} }

func (r *PgReviewRepo) Insert(ctx context.Context, rev *Review) (int, error) {
	conn := dbConn(ctx, r.db)
	var id int
	err := conn.QueryRowContext(ctx, `
        INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `, rev.ExchangeID, rev.AuthorID, rev.TargetID, rev.Note, rev.Commentaire).Scan(&id)
	if err != nil {
		return 0, mapPgError(err)
	}
	return id, nil
}

func (r *PgReviewRepo) ListByUser(ctx context.Context, targetUserID int, limit, offset int) ([]Review, error) {
	return r.list(ctx, `
        SELECT id, exchange_id, author_id, target_id, note, commentaire, created_at
        FROM reviews
        WHERE target_id = $1
        ORDER BY created_at DESC, id DESC
        LIMIT $2 OFFSET $3
    `, targetUserID, limit, offset)
}

func (r *PgReviewRepo) ListByService(ctx context.Context, serviceID int, limit, offset int) ([]Review, error) {
	return r.list(ctx, `
        SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, r.commentaire, r.created_at
        FROM reviews r
        JOIN exchanges e ON e.id = r.exchange_id
        WHERE e.service_id = $1
        ORDER BY r.created_at DESC, r.id DESC
        LIMIT $2 OFFSET $3
    `, serviceID, limit, offset)
}

func (r *PgReviewRepo) ExistsByAuthor(ctx context.Context, exchangeID, authorID int) (bool, error) {
	conn := dbConn(ctx, r.db)
	var exists bool
	if err := conn.QueryRowContext(ctx, `
        SELECT EXISTS(SELECT 1 FROM reviews WHERE exchange_id = $1 AND author_id = $2)
    `, exchangeID, authorID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PgReviewRepo) list(ctx context.Context, query string, args ...any) ([]Review, error) {
	conn := dbConn(ctx, r.db)
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Review
	for rows.Next() {
		var rev Review
		if err := rows.Scan(&rev.ID, &rev.ExchangeID, &rev.AuthorID, &rev.TargetID, &rev.Note, &rev.Commentaire, &rev.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

var _ ReviewRepository = (*PgReviewRepo)(nil)
