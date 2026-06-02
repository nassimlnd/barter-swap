package main

import "strings"

// SkillNiveau est l'enum des niveaux de maîtrise déclarables.
// Les valeurs sont sans accent pour rester compatibles avec une CHECK
// contrainte SQL non ambigüe (ASCII pur).
type SkillNiveau string

const (
	NiveauDebutant      SkillNiveau = "debutant"
	NiveauIntermediaire SkillNiveau = "intermediaire"
	NiveauExpert        SkillNiveau = "expert"

	MaxSkillNomLength = 80
)

// Valid renvoie true si n est l'un des trois niveaux acceptés.
func (n SkillNiveau) Valid() bool {
	switch n {
	case NiveauDebutant, NiveauIntermediaire, NiveauExpert:
		return true
	}
	return false
}

// Skill — une compétence déclarée par un utilisateur.
type Skill struct {
	UserID int
	Nom    string
	Niveau SkillNiveau
}

// ValidateSkill valide nom + niveau avant insertion.
func ValidateSkill(s Skill) error {
	nom := strings.TrimSpace(s.Nom)
	if nom == "" {
		return newFieldErr(ErrBadRequest, "nom", "requis")
	}
	if len(nom) > MaxSkillNomLength {
		return newFieldErr(ErrBadRequest, "nom", "trop long")
	}
	if !s.Niveau.Valid() {
		return ErrInvalidSkillNiveau
	}
	return nil
}

// ValidateSkills valide un lot entier ET vérifie qu'il n'y a pas de doublons
// par nom (insensible à la casse). Utilisé par PUT /api/users/{id}/skills.
func ValidateSkills(skills []Skill) error {
	seen := make(map[string]struct{}, len(skills))
	for _, s := range skills {
		if err := ValidateSkill(s); err != nil {
			return err
		}
		key := strings.ToLower(strings.TrimSpace(s.Nom))
		if _, dup := seen[key]; dup {
			return newFieldErr(ErrBadRequest, "skills", "compétence en double: "+s.Nom)
		}
		seen[key] = struct{}{}
	}
	return nil
}
