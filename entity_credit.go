package main

import "time"

// CreditTxType — type d'opération dans le journal des crédits.
// Synchronisé avec la contrainte CHECK dans credit_transactions.type.
type CreditTxType string

const (
	// CreditWelcome : crédits offerts à l'inscription (+10). ExchangeID == nil.
	CreditWelcome CreditTxType = "welcome"

	// CreditSpend : débit lors de l'acceptation d'un échange (montant négatif).
	CreditSpend CreditTxType = "spend"

	// CreditEarn : crédit définitif à la complétion d'un échange (montant positif).
	CreditEarn CreditTxType = "earn"

	// CreditRefund : restitution au requester en cas de cancel/reject d'un
	// échange déjà accepté (montant positif).
	CreditRefund CreditTxType = "refund"
)

// Valid renvoie true si t est l'un des types acceptés.
func (t CreditTxType) Valid() bool {
	switch t {
	case CreditWelcome, CreditSpend, CreditEarn, CreditRefund:
		return true
	}
	return false
}

// CreditTransaction — ligne du journal append-only. La somme des montants par
// utilisateur doit toujours égaler users.credit_balance (vérifiable via le
// bonus /debug/reconcile).
type CreditTransaction struct {
	ID         int
	UserID     int
	ExchangeID *int // nil uniquement pour CreditWelcome
	Montant    int  // signé : négatif = débit, positif = crédit
	Type       CreditTxType
	CreatedAt  time.Time
}
