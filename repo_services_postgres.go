package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type PgServiceRepo struct{ db *sql.DB }

func NewPgServiceRepo(db *sql.DB) *PgServiceRepo { return &PgServiceRepo{db: db} }

func (r *PgServiceRepo) Insert(ctx context.Context, s *Service) (int, error) {
	conn := dbConn(ctx, r.db)
	var id int
	err := conn.QueryRowContext(ctx, `
        INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `, s.ProviderID, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif).Scan(&id)
	if err != nil {
		return 0, mapPgError(err)
	}
	return id, nil
}

func (r *PgServiceRepo) GetByID(ctx context.Context, id int) (*Service, error) {
	conn := dbConn(ctx, r.db)
	s := &Service{}
	err := conn.QueryRowContext(ctx, `
        SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at
        FROM services
        WHERE id = $1
    `, id).Scan(&s.ID, &s.ProviderID, &s.Titre, &s.Description, &s.Categorie, &s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt)
	if err != nil {
		return nil, mapNoRows(err)
	}
	return s, nil
}

func (r *PgServiceRepo) Update(ctx context.Context, id int, s *Service) error {
	conn := dbConn(ctx, r.db)
	res, err := conn.ExecContext(ctx, `
        UPDATE services
        SET titre = $1, description = $2, categorie = $3, duree_minutes = $4, credits = $5, ville = $6, actif = $7
        WHERE id = $8
    `, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif, id)
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

func (r *PgServiceRepo) Delete(ctx context.Context, id int) error {
	conn := dbConn(ctx, r.db)
	res, err := conn.ExecContext(ctx, `DELETE FROM services WHERE id = $1`, id)
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

func (r *PgServiceRepo) List(ctx context.Context, f ServiceFilter) ([]Service, int, error) {
	conn := dbConn(ctx, r.db)
	where, args := buildServiceWhere(f)

	var total int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM services `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, f.Limit, f.Offset)
	query := `
        SELECT id, provider_id, titre, description, categorie, duree_minutes, credits, ville, actif, created_at
        FROM services ` + where + `
        ORDER BY created_at DESC, id DESC
        LIMIT $` + fmt.Sprint(len(args)-1) + ` OFFSET $` + fmt.Sprint(len(args))
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Service
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.ID, &s.ProviderID, &s.Titre, &s.Description, &s.Categorie, &s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func buildServiceWhere(f ServiceFilter) (string, []any) {
	var clauses []string
	var args []any
	add := func(clause string, values ...any) {
		for _, v := range values {
			args = append(args, v)
			clause = strings.Replace(clause, "?", fmt.Sprintf("$%d", len(args)), 1)
		}
		clauses = append(clauses, clause)
	}
	if f.OnlyActif {
		clauses = append(clauses, "actif = TRUE")
	}
	if f.Categorie != "" {
		add("categorie = ?", f.Categorie)
	}
	if f.Ville != "" {
		add("ville ILIKE ?", f.Ville)
	}
	if f.Search != "" {
		pattern := "%" + f.Search + "%"
		add("(titre ILIKE ? OR description ILIKE ?)", pattern, pattern)
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

var _ ServiceRepository = (*PgServiceRepo)(nil)
