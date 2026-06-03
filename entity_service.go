package main

import (
	"strings"
	"time"
)

const (
	MaxServiceTitreLength       = 120
	MaxServiceDescriptionLength = 2000
	MaxServiceVilleLength       = 100
)

// Service — une annonce publiée par un utilisateur (provider) proposant un
// service contre des crédits-temps.
type Service struct {
	ID           int
	ProviderID   int
	Titre        string
	Description  string
	Categorie    string
	DureeMinutes int
	Credits      int
	Ville        string
	Actif        bool
	CreatedAt    time.Time
}

// IsBookable : un service ne peut être réservé que s'il est actif.
// (Le verrou "un seul échange actif par service" est en plus géré au niveau DB
// par l'index unique partiel.)
func (s *Service) IsBookable() bool {
	return s.Actif
}

// ValidateServiceInput vérifie les champs métier avant insertion ou update.
// Le couplage entre service.Categorie et les compétences déclarées par le
// provider est vérifié dans l'usecase (qui charge l'user concerné).
func ValidateServiceInput(titre, description, categorie, ville string, dureeMinutes, credits int) error {
	titre = strings.TrimSpace(titre)
	if titre == "" {
		return newFieldErr(ErrBadRequest, "titre", "requis")
	}
	if len(titre) > MaxServiceTitreLength {
		return newFieldErr(ErrBadRequest, "titre", "trop long")
	}
	if len(description) > MaxServiceDescriptionLength {
		return newFieldErr(ErrBadRequest, "description", "trop longue")
	}
	if !IsValidCategorie(categorie) {
		return ErrInvalidCategorie
	}
	if len(ville) > MaxServiceVilleLength {
		return newFieldErr(ErrBadRequest, "ville", "trop longue")
	}
	if dureeMinutes <= 0 {
		return newFieldErr(ErrBadRequest, "duree_minutes", "doit être positif")
	}
	if credits <= 0 {
		return newFieldErr(ErrBadRequest, "credits", "doit être positif")
	}
	return nil
}
