package main

// UserStats — projection des statistiques d'un utilisateur. Construite par
// une requête agrégée unique côté repo (cf. repo_stats_postgres.go). Pas
// d'identité, pas de mutation : c'est une vue en lecture seule.
type UserStats struct {
	UserID            int
	ServicesActifs    int
	EchangesCompletes int
	CreditBalance     int
	NoteMoyenne       float64
	NbAvis            int
	TotalGagne        int
	TotalDepense      int
}
