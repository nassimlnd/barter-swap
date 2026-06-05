// Package main — couche entités : erreurs métier.
//
// Ce fichier appartient à la couche entité (Clean Architecture). Il n'importe
// que la stdlib (errors). Aucune dépendance vers net/http, database/sql,
// encoding/json. Les codes HTTP sont mappés dans http_errors.go via errors.Is.
package main

import "errors"

// ─── Kind sentinels ─────────────────────────────────────────────────────────
// Chaque erreur retournée par les couches entité / usecase / repo doit pouvoir
// se réduire (via errors.Is) à l'une de ces catégories. Le mapping vers un
// code HTTP est centralisé dans http_errors.go.
var (
	ErrNotFound     = errors.New("ressource introuvable")
	ErrBadRequest   = errors.New("requête invalide")
	ErrUnauthorized = errors.New("authentification requise")
	ErrForbidden    = errors.New("action non autorisée")
	ErrConflict     = errors.New("conflit d'état")
	ErrInternal     = errors.New("erreur interne")
)

// DomainError porte un message lisible pour le client et un Kind sentinel pour
// le mapping HTTP. Field est optionnel et permet de renvoyer l'attribut fautif
// dans la réponse JSON ({"error":"...","field":"pseudo"}).
type DomainError struct {
	Kind    error
	Message string
	Field   string
}

func (d *DomainError) Error() string {
	if d.Field != "" {
		return d.Field + ": " + d.Message
	}
	return d.Message
}

func (d *DomainError) Unwrap() error { return d.Kind }

// newDomainErr construit une DomainError sans champ associé.
func newDomainErr(kind error, msg string) *DomainError {
	return &DomainError{Kind: kind, Message: msg}
}

// newFieldErr construit une DomainError liée à un champ spécifique (utile pour
// les validations de payload).
func newFieldErr(kind error, field, msg string) *DomainError {
	return &DomainError{Kind: kind, Message: msg, Field: field}
}

// ─── Erreurs métier nommées ─────────────────────────────────────────────────
// Préférer ces sentinels à des fmt.Errorf ad-hoc pour les règles partagées
// (utilisées dans les handlers + tests). Pour les validations de champs
// spécifiques (pseudo vide, etc.), utilisez newFieldErr inline.
var (
	ErrInsufficientCredits  = newDomainErr(ErrBadRequest, "crédits insuffisants")
	ErrSelfExchange         = newDomainErr(ErrBadRequest, "échange impossible avec soi-même")
	ErrServiceInactive      = newDomainErr(ErrBadRequest, "service inactif")
	ErrServiceAlreadyBooked = newDomainErr(ErrConflict, "un échange est déjà en cours pour ce service")
	ErrAlreadyReviewed      = newDomainErr(ErrConflict, "avis déjà déposé pour cet échange")
	ErrInvalidStatus        = newDomainErr(ErrBadRequest, "transition invalide depuis le statut courant")
	ErrInvalidNote          = newDomainErr(ErrBadRequest, "note invalide (doit être entre 1 et 5)")
	ErrSkillNotOwned        = newDomainErr(ErrBadRequest, "compétence non déclarée par le fournisseur")
	ErrCannotReview         = newDomainErr(ErrForbidden, "seuls les participants à l'échange peuvent le noter")
	ErrExchangeNotCompleted = newDomainErr(ErrBadRequest, "l'échange doit être terminé pour être noté")
	ErrUnauthorizedExchange = newDomainErr(ErrForbidden, "action non autorisée sur cet échange")
	ErrPseudoTaken          = newDomainErr(ErrConflict, "pseudo déjà utilisé")
	ErrInvalidCategorie     = newDomainErr(ErrBadRequest, "catégorie inconnue")
	ErrSelfReview           = newDomainErr(ErrBadRequest, "impossible de se noter soi-même")
	ErrInvalidSkillNiveau   = newDomainErr(ErrBadRequest, "niveau invalide (doit être 'debutant', 'intermediaire' ou 'expert')")
	ErrInvalidPagination    = newDomainErr(ErrBadRequest, "paramètres de pagination invalides")
)
