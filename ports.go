// ports.go — interfaces (ports) consommées par les use cases.
//
// Couche : ports / boundaries (Clean Architecture).
// Imports autorisés : context, types du domaine (entity_*). PAS d'import vers
// database/sql, net/http, encoding/json : ce fichier doit pouvoir être
// implémenté par n'importe quel adapter (Postgres, MySQL, in-memory, mock).
package main

import "context"

// TxRunner abstrait la gestion transactionnelle. Permet aux use cases de
// composer plusieurs opérations atomiques sans connaître *sql.Tx directement.
// L'implémentation Postgres injecte la transaction en cours dans le contexte ;
// les repos détectent cette présence et utilisent la même connexion.
type TxRunner interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// ─── User & Skill ──────────────────────────────────────────────────────────

type UserRepository interface {
	// Insert insère un utilisateur (avec credit_balance par défaut) et renvoie
	// l'ID généré. Renvoie ErrPseudoTaken si le pseudo est déjà utilisé.
	Insert(ctx context.Context, u *User) (int, error)
	// GetByID renvoie l'utilisateur (sans ses skills, à charger séparément).
	// Renvoie ErrNotFound si absent.
	GetByID(ctx context.Context, id int) (*User, error)
	// UpdateProfile met à jour pseudo, bio et ville. Renvoie ErrPseudoTaken si
	// le nouveau pseudo est déjà utilisé.
	UpdateProfile(ctx context.Context, id int, pseudo, bio, ville string) error
	// AdjustBalance ajoute delta à credit_balance (peut être négatif).
	// La contrainte CHECK rejette les soldes négatifs : renvoie ErrInsufficientCredits.
	AdjustBalance(ctx context.Context, userID, delta int) error
}

type SkillRepository interface {
	ListByUser(ctx context.Context, userID int) ([]Skill, error)
	// ReplaceForUser supprime puis ré-insère le lot complet. Atomique si
	// appelé dans une TxRunner.InTx (recommandé).
	ReplaceForUser(ctx context.Context, userID int, skills []Skill) error
}

// ─── Service ───────────────────────────────────────────────────────────────

// ServiceFilter regroupe les paramètres de listing de services.
type ServiceFilter struct {
	Categorie string // optionnel, "" = pas de filtre
	Ville     string // optionnel
	Search    string // recherche ILIKE sur titre + description
	OnlyActif bool   // true par défaut côté usecase
	Limit     int    // par défaut 20, cap à 100
	Offset    int
}

type ServiceRepository interface {
	Insert(ctx context.Context, s *Service) (int, error)
	GetByID(ctx context.Context, id int) (*Service, error)
	Update(ctx context.Context, id int, s *Service) error
	Delete(ctx context.Context, id int) error
	// List renvoie la page + le total (utile pour la pagination).
	List(ctx context.Context, f ServiceFilter) (items []Service, total int, err error)
}

// ─── Exchange ──────────────────────────────────────────────────────────────

type ExchangeFilter struct {
	UserID int            // requester OU owner ; 0 = pas de filtre
	Status ExchangeStatus // "" = tous statuts
	Limit  int
	Offset int
}

type ExchangeRepository interface {
	Insert(ctx context.Context, e *Exchange) (int, error)
	GetByID(ctx context.Context, id int) (*Exchange, error)
	// UpdateStatus met à jour le statut si l'état courant correspond à 'from'.
	// Renvoie ErrConflict si l'état a changé entre-temps (RowsAffected=0).
	UpdateStatus(ctx context.Context, id int, from, to ExchangeStatus) error
	List(ctx context.Context, f ExchangeFilter) (items []Exchange, total int, err error)
}

// ─── Credit ────────────────────────────────────────────────────────────────

type CreditRepository interface {
	Insert(ctx context.Context, tx *CreditTransaction) (int, error)
	ListByUser(ctx context.Context, userID int, limit, offset int) ([]CreditTransaction, error)
	// SumByUser renvoie la somme des montants pour un utilisateur (utilisée
	// par /debug/reconcile bonus pour vérifier l'égalité avec credit_balance).
	SumByUser(ctx context.Context, userID int) (int, error)
}

// ─── Review ────────────────────────────────────────────────────────────────

type ReviewRepository interface {
	// Insert crée un avis. Renvoie ErrAlreadyReviewed sur violation d'unicité
	// (un seul avis par auteur par échange).
	Insert(ctx context.Context, r *Review) (int, error)
	ListByUser(ctx context.Context, targetUserID int, limit, offset int) ([]Review, error)
	ListByService(ctx context.Context, serviceID int, limit, offset int) ([]Review, error)
	ExistsByAuthor(ctx context.Context, exchangeID, authorID int) (bool, error)
}

// ─── Stats ─────────────────────────────────────────────────────────────────

type StatsRepository interface {
	UserStats(ctx context.Context, userID int) (UserStats, error)
}
