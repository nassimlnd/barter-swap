package main

import (
	"strings"
	"time"
)

// Constantes utilisateur — partagées entre validations, repo (limites
// d'insertion) et tests.
const (
	WelcomeCredits  = 10
	MaxPseudoLength = 50
	MaxBioLength    = 500
	MaxVilleLength  = 100
)

// User est l'entité racine : un compte BarterSwap doté d'un solde de crédits
// et d'une liste optionnelle de compétences (chargées à la demande).
type User struct {
	ID            int
	Pseudo        string
	Bio           string
	Ville         string
	Skills        []Skill
	CreditBalance int
	CreatedAt     time.Time
}

// CanAfford : l'utilisateur a-t-il assez de crédits pour bloquer 'cost' ?
// Utilisé avant la création d'un échange ou l'acceptation d'une demande.
func (u *User) CanAfford(cost int) bool {
	return cost >= 0 && u.CreditBalance >= cost
}

// HasSkill renvoie true si l'utilisateur a déclaré la compétence 'nom' (la
// recherche est insensible à la casse pour offrir un peu de souplesse côté
// API publique, mais l'insertion garde la casse du fournisseur).
func (u *User) HasSkill(nom string) bool {
	target := strings.ToLower(strings.TrimSpace(nom))
	for _, s := range u.Skills {
		if strings.ToLower(s.Nom) == target {
			return true
		}
	}
	return false
}

// ValidateUserInput vérifie les champs avant insertion ou mise à jour.
// Ne contrôle pas l'unicité du pseudo (responsabilité du repo : 23505).
func ValidateUserInput(pseudo, bio, ville string) error {
	pseudo = strings.TrimSpace(pseudo)
	if pseudo == "" {
		return newFieldErr(ErrBadRequest, "pseudo", "requis")
	}
	if len(pseudo) > MaxPseudoLength {
		return newFieldErr(ErrBadRequest, "pseudo", "trop long")
	}
	if len(bio) > MaxBioLength {
		return newFieldErr(ErrBadRequest, "bio", "trop longue")
	}
	if len(ville) > MaxVilleLength {
		return newFieldErr(ErrBadRequest, "ville", "trop longue")
	}
	return nil
}
