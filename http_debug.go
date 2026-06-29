package main

import (
	"net/http"
)

type reconcileRow struct {
	UserID         int    `json:"user_id"`
	Pseudo         string `json:"pseudo"`
	CreditBalance  int    `json:"credit_balance"`
	JournalBalance int    `json:"journal_balance"`
}

type reconcileResponse struct {
	OK           bool           `json:"ok"`
	Incoherences []reconcileRow `json:"incoherences"`
}

func (a *App) handleDebugReconcile(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.QueryContext(r.Context(), `
        SELECT u.id, u.pseudo, u.credit_balance, COALESCE(SUM(ct.montant), 0)::int AS journal_balance
        FROM users u
        LEFT JOIN credit_transactions ct ON ct.user_id = u.id
        GROUP BY u.id, u.pseudo, u.credit_balance
        HAVING u.credit_balance <> COALESCE(SUM(ct.montant), 0)
        ORDER BY u.id
    `)
	if err != nil {
		writeError(w, ErrInternal)
		return
	}
	defer rows.Close()

	var out []reconcileRow
	for rows.Next() {
		var row reconcileRow
		if err := rows.Scan(&row.UserID, &row.Pseudo, &row.CreditBalance, &row.JournalBalance); err != nil {
			writeError(w, ErrInternal)
			return
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		writeError(w, ErrInternal)
		return
	}
	writeJSON(w, http.StatusOK, reconcileResponse{OK: len(out) == 0, Incoherences: out})
}

func (a *App) handleDebugSeed(w http.ResponseWriter, r *http.Request) {
	if _, err := a.DB.ExecContext(r.Context(), `
        INSERT INTO users (pseudo, bio, ville, credit_balance) VALUES
            ('alice', 'Jardinière amateur', 'Paris', 10),
            ('bob', 'Besoin de coups de main ponctuels', 'Paris', 10),
            ('carla', 'Cuisine familiale', 'Lyon', 10)
        ON CONFLICT (pseudo) DO NOTHING;

        INSERT INTO credit_transactions (user_id, montant, type)
        SELECT u.id, 10, 'welcome'
        FROM users u
        WHERE u.pseudo IN ('alice', 'bob', 'carla')
          AND NOT EXISTS (
              SELECT 1 FROM credit_transactions ct
              WHERE ct.user_id = u.id AND ct.type = 'welcome'
          );

        INSERT INTO skills (user_id, nom, niveau)
        SELECT u.id, 'Jardinage', 'expert' FROM users u WHERE u.pseudo = 'alice'
        ON CONFLICT (user_id, nom) DO NOTHING;

        INSERT INTO skills (user_id, nom, niveau)
        SELECT u.id, 'Cuisine', 'expert' FROM users u WHERE u.pseudo = 'carla'
        ON CONFLICT (user_id, nom) DO NOTHING;

        INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville)
        SELECT u.id, 'Tondre une pelouse', 'Aide jardinage pour petits espaces', 'Jardinage', 60, 1, 'Paris'
        FROM users u WHERE u.pseudo = 'alice'
          AND NOT EXISTS (SELECT 1 FROM services s WHERE s.titre = 'Tondre une pelouse' AND s.provider_id = u.id);

        INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville)
        SELECT u.id, 'Cours de cuisine maison', 'Recettes simples et économiques', 'Cuisine', 90, 2, 'Lyon'
        FROM users u WHERE u.pseudo = 'carla'
          AND NOT EXISTS (SELECT 1 FROM services s WHERE s.titre = 'Cours de cuisine maison' AND s.provider_id = u.id);
    `); err != nil {
		writeError(w, ErrInternal)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "seeded"})
}
