package main

import "time"

// ExchangeStatus — états du cycle de vie d'un échange. Synchronisé avec la
// contrainte CHECK dans migrations/001_schema.sql.
//
// Transitions valides :
//
//	pending   -> accepted   (owner accepte)
//	pending   -> rejected   (owner refuse)
//	pending   -> cancelled  (requester annule avant acceptation)
//	accepted  -> completed  (owner confirme la prestation rendue)
//	accepted  -> cancelled  (l'une des deux parties annule après acceptation)
type ExchangeStatus string

const (
	StatusPending   ExchangeStatus = "pending"
	StatusAccepted  ExchangeStatus = "accepted"
	StatusRejected  ExchangeStatus = "rejected"
	StatusCancelled ExchangeStatus = "cancelled"
	StatusCompleted ExchangeStatus = "completed"
)

// Valid renvoie true si s est l'un des cinq statuts acceptés. Utilisé pour
// valider un filtre ?status= en query string.
func (s ExchangeStatus) Valid() bool {
	switch s {
	case StatusPending, StatusAccepted, StatusRejected, StatusCancelled, StatusCompleted:
		return true
	}
	return false
}

// IsActive : un échange "actif" (pending ou accepted) verrouille le service
// concerné (cf. index unique partiel uq_exchange_active_per_service).
func (s ExchangeStatus) IsActive() bool {
	return s == StatusPending || s == StatusAccepted
}

// IsTerminal : pas de transition possible depuis cet état.
func (s ExchangeStatus) IsTerminal() bool {
	return s == StatusRejected || s == StatusCancelled || s == StatusCompleted
}

// Exchange représente une demande d'échange entre un requester et un owner
// pour un service donné. credits est un snapshot du coût au moment de la
// demande : il reste figé même si le service est modifié plus tard.
type Exchange struct {
	ID          int
	ServiceID   int
	RequesterID int
	OwnerID     int
	Credits     int
	Status      ExchangeStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CanAccept : transition pending -> accepted possible.
func (e *Exchange) CanAccept() bool { return e.Status == StatusPending }

// CanReject : transition pending -> rejected possible.
func (e *Exchange) CanReject() bool { return e.Status == StatusPending }

// CanComplete : transition accepted -> completed possible.
func (e *Exchange) CanComplete() bool { return e.Status == StatusAccepted }

// CanCancel : transition pending -> cancelled OU accepted -> cancelled.
func (e *Exchange) CanCancel() bool {
	return e.Status == StatusPending || e.Status == StatusAccepted
}

// InvolvesUser : true si userID est requester ou owner. Utilisé pour autoriser
// l'accès en lecture/écriture sur l'échange (sans s'occuper du rôle précis).
func (e *Exchange) InvolvesUser(userID int) bool {
	return e.RequesterID == userID || e.OwnerID == userID
}
