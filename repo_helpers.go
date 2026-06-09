package main

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgUniqueViolation = "23505"
	pgCheckViolation  = "23514"
)

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case pgUniqueViolation:
		name := pgErr.ConstraintName
		switch {
		case name == "users_pseudo_key":
			return ErrPseudoTaken
		case name == "uq_exchange_active_per_service" || strings.Contains(name, "exchange_active"):
			return ErrServiceAlreadyBooked
		case name == "reviews_exchange_id_author_id_key":
			return ErrAlreadyReviewed
		default:
			return ErrConflict
		}
	case pgCheckViolation:
		if strings.Contains(pgErr.ConstraintName, "credit_balance") {
			return ErrInsufficientCredits
		}
		return ErrBadRequest
	default:
		return err
	}
}
