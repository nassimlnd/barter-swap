package main

import "time"

const (
	MinNote                  = 1
	MaxNote                  = 5
	MaxReviewCommentaireSize = 1000
)

// Review — avis laissé après complétion d'un échange.
// Règles métier :
//   - 1 avis maximum par auteur par échange (UNIQUE en DB)
//   - Auteur != cible (CHECK en DB)
//   - L'échange doit être en status 'completed' (vérifié dans l'usecase)
type Review struct {
	ID          int
	ExchangeID  int
	AuthorID    int
	TargetID    int
	Note        int
	Commentaire string
	CreatedAt   time.Time
}

// ValidateReviewInput valide la note et la taille du commentaire.
func ValidateReviewInput(note int, commentaire string) error {
	if note < MinNote || note > MaxNote {
		return ErrInvalidNote
	}
	if len(commentaire) > MaxReviewCommentaireSize {
		return newFieldErr(ErrBadRequest, "commentaire", "trop long")
	}
	return nil
}
