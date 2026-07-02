package main

import (
	"errors"
	"testing"
)

func TestDomainError_Error(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  *DomainError
		want string
	}{
		{"sans field", newDomainErr(ErrBadRequest, "msg"), "msg"},
		{"avec field", newFieldErr(ErrBadRequest, "pseudo", "requis"), "pseudo: requis"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Fatalf("Error()=%q, want %q", got, tc.want)
			}
		})
	}
}

func TestDomainError_UnwrapsToKind(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		kind error
	}{
		{"InsufficientCredits -> BadRequest", ErrInsufficientCredits, ErrBadRequest},
		{"ServiceAlreadyBooked -> Conflict", ErrServiceAlreadyBooked, ErrConflict},
		{"AlreadyReviewed -> Conflict", ErrAlreadyReviewed, ErrConflict},
		{"InvalidStatus -> BadRequest", ErrInvalidStatus, ErrBadRequest},
		{"CannotReview -> Forbidden", ErrCannotReview, ErrForbidden},
		{"PseudoTaken -> Conflict", ErrPseudoTaken, ErrConflict},
		{"newFieldErr unwraps", newFieldErr(ErrBadRequest, "x", "y"), ErrBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, tc.kind) {
				t.Fatalf("errors.Is(%v, %v) = false, want true", tc.err, tc.kind)
			}
		})
	}
}

func TestDomainError_DistinctSentinels(t *testing.T) {
	t.Parallel()
	// Vérifie que deux DomainError de Kind différent ne sont pas confondus.
	if errors.Is(ErrServiceAlreadyBooked, ErrBadRequest) {
		t.Fatalf("ErrServiceAlreadyBooked ne doit PAS être BadRequest")
	}
	if errors.Is(ErrInsufficientCredits, ErrConflict) {
		t.Fatalf("ErrInsufficientCredits ne doit PAS être Conflict")
	}
}
