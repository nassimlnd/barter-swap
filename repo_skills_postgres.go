package main

import (
	"context"
	"database/sql"
)

type PgSkillRepo struct{ db *sql.DB }

func NewPgSkillRepo(db *sql.DB) *PgSkillRepo { return &PgSkillRepo{db: db} }

func (r *PgSkillRepo) ListByUser(ctx context.Context, userID int) ([]Skill, error) {
	conn := dbConn(ctx, r.db)
	rows, err := conn.QueryContext(ctx, `
        SELECT user_id, nom, niveau
        FROM skills
        WHERE user_id = $1
        ORDER BY nom
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Skill
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.UserID, &s.Nom, &s.Niveau); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PgSkillRepo) ReplaceForUser(ctx context.Context, userID int, skills []Skill) error {
	conn := dbConn(ctx, r.db)
	if _, err := conn.ExecContext(ctx, `DELETE FROM skills WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, s := range skills {
		if _, err := conn.ExecContext(ctx, `
            INSERT INTO skills (user_id, nom, niveau)
            VALUES ($1, $2, $3)
        `, userID, s.Nom, s.Niveau); err != nil {
			return mapPgError(err)
		}
	}
	return nil
}

var _ SkillRepository = (*PgSkillRepo)(nil)
