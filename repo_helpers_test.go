package main

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapPgError(t *testing.T) {
	t.Parallel()
	plain := errors.New("plain")
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"nil", nil, nil},
		{"non pg", plain, plain},
		{"pseudo unique", &pgconn.PgError{Code: pgUniqueViolation, ConstraintName: "users_pseudo_key"}, ErrPseudoTaken},
		{"exchange active unique", &pgconn.PgError{Code: pgUniqueViolation, ConstraintName: "uq_exchange_active_per_service"}, ErrServiceAlreadyBooked},
		{"review unique", &pgconn.PgError{Code: pgUniqueViolation, ConstraintName: "reviews_exchange_id_author_id_key"}, ErrAlreadyReviewed},
		{"other unique", &pgconn.PgError{Code: pgUniqueViolation, ConstraintName: "other"}, ErrConflict},
		{"balance check", &pgconn.PgError{Code: pgCheckViolation, ConstraintName: "users_credit_balance_check"}, ErrInsufficientCredits},
		{"other check", &pgconn.PgError{Code: pgCheckViolation, ConstraintName: "some_check"}, ErrBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapPgError(tc.err)
			if tc.want == nil {
				if got != nil {
					t.Fatalf("got %v, want nil", got)
				}
				return
			}
			if !errors.Is(got, tc.want) {
				t.Fatalf("got %v, want errors.Is(..., %v)", got, tc.want)
			}
		})
	}
}
