package main

import (
	"context"
	"database/sql"
)

type PgStatsRepo struct{ db *sql.DB }

func NewPgStatsRepo(db *sql.DB) *PgStatsRepo { return &PgStatsRepo{db: db} }

func (r *PgStatsRepo) UserStats(ctx context.Context, userID int) (UserStats, error) {
	conn := dbConn(ctx, r.db)
	var s UserStats
	err := conn.QueryRowContext(ctx, `
        SELECT
            u.id,
            (SELECT COUNT(*) FROM services sv WHERE sv.provider_id = u.id AND sv.actif = TRUE) AS services_actifs,
            (SELECT COUNT(*) FROM exchanges e WHERE e.status = 'completed' AND (e.requester_id = u.id OR e.owner_id = u.id)) AS echanges_completes,
            u.credit_balance,
            COALESCE((SELECT AVG(r.note)::float8 FROM reviews r WHERE r.target_id = u.id), 0) AS note_moyenne,
            (SELECT COUNT(*) FROM reviews r WHERE r.target_id = u.id) AS nb_avis,
            COALESCE((SELECT SUM(ct.montant) FROM credit_transactions ct WHERE ct.user_id = u.id AND ct.type = 'earn'), 0) AS total_gagne,
            COALESCE((SELECT -SUM(ct.montant) FROM credit_transactions ct WHERE ct.user_id = u.id AND ct.type = 'spend'), 0) AS total_depense
        FROM users u
        WHERE u.id = $1
    `, userID).Scan(&s.UserID, &s.ServicesActifs, &s.EchangesCompletes, &s.CreditBalance, &s.NoteMoyenne, &s.NbAvis, &s.TotalGagne, &s.TotalDepense)
	if err != nil {
		return UserStats{}, mapNoRows(err)
	}
	return s, nil
}

var _ StatsRepository = (*PgStatsRepo)(nil)
