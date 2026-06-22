package main

import "time"

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

type skillDTO struct {
	Nom    string `json:"nom"`
	Niveau string `json:"niveau"`
}

func skillToDTO(s Skill) skillDTO {
	return skillDTO{Nom: s.Nom, Niveau: string(s.Niveau)}
}

func skillsToDTO(skills []Skill) []skillDTO {
	out := make([]skillDTO, 0, len(skills))
	for _, s := range skills {
		out = append(out, skillToDTO(s))
	}
	return out
}

type userDTO struct {
	ID            int        `json:"id"`
	Pseudo        string     `json:"pseudo"`
	Bio           string     `json:"bio,omitempty"`
	Ville         string     `json:"ville,omitempty"`
	Skills        []skillDTO `json:"skills,omitempty"`
	CreditBalance int        `json:"credit_balance"`
	CreatedAt     string     `json:"created_at"`
}

func userToDTO(u *User) userDTO {
	return userDTO{
		ID:            u.ID,
		Pseudo:        u.Pseudo,
		Bio:           u.Bio,
		Ville:         u.Ville,
		Skills:        skillsToDTO(u.Skills),
		CreditBalance: u.CreditBalance,
		CreatedAt:     formatTime(u.CreatedAt),
	}
}

type serviceDTO struct {
	ID           int    `json:"id"`
	ProviderID   int    `json:"provider_id"`
	Titre        string `json:"titre"`
	Description  string `json:"description,omitempty"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville,omitempty"`
	Actif        bool   `json:"actif"`
	CreatedAt    string `json:"created_at"`
}

func serviceToDTO(s *Service) serviceDTO {
	return serviceDTO{
		ID:           s.ID,
		ProviderID:   s.ProviderID,
		Titre:        s.Titre,
		Description:  s.Description,
		Categorie:    s.Categorie,
		DureeMinutes: s.DureeMinutes,
		Credits:      s.Credits,
		Ville:        s.Ville,
		Actif:        s.Actif,
		CreatedAt:    formatTime(s.CreatedAt),
	}
}

type exchangeDTO struct {
	ID          int    `json:"id"`
	ServiceID   int    `json:"service_id"`
	RequesterID int    `json:"requester_id"`
	OwnerID     int    `json:"owner_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func exchangeToDTO(e *Exchange) exchangeDTO {
	return exchangeDTO{
		ID:          e.ID,
		ServiceID:   e.ServiceID,
		RequesterID: e.RequesterID,
		OwnerID:     e.OwnerID,
		Status:      string(e.Status),
		CreatedAt:   formatTime(e.CreatedAt),
		UpdatedAt:   formatTime(e.UpdatedAt),
	}
}

type reviewDTO struct {
	ID          int    `json:"id"`
	ExchangeID  int    `json:"exchange_id"`
	AuthorID    int    `json:"author_id"`
	TargetID    int    `json:"target_id"`
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func reviewToDTO(r *Review) reviewDTO {
	return reviewDTO{
		ID:          r.ID,
		ExchangeID:  r.ExchangeID,
		AuthorID:    r.AuthorID,
		TargetID:    r.TargetID,
		Note:        r.Note,
		Commentaire: r.Commentaire,
		CreatedAt:   formatTime(r.CreatedAt),
	}
}

type userStatsDTO struct {
	UserID            int     `json:"user_id"`
	ServicesActifs    int     `json:"services_actifs"`
	EchangesCompletes int     `json:"echanges_completes"`
	CreditBalance     int     `json:"credit_balance"`
	NoteMoyenne       float64 `json:"note_moyenne"`
	NbAvis            int     `json:"nb_avis"`
	TotalGagne        int     `json:"total_gagne"`
	TotalDepense      int     `json:"total_depense"`
}

func userStatsToDTO(s UserStats) userStatsDTO {
	return userStatsDTO{
		UserID:            s.UserID,
		ServicesActifs:    s.ServicesActifs,
		EchangesCompletes: s.EchangesCompletes,
		CreditBalance:     s.CreditBalance,
		NoteMoyenne:       s.NoteMoyenne,
		NbAvis:            s.NbAvis,
		TotalGagne:        s.TotalGagne,
		TotalDepense:      s.TotalDepense,
	}
}
