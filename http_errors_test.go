package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
		wantField  string
	}{
		{"bad request", newFieldErr(ErrBadRequest, "pseudo", "requis"), http.StatusBadRequest, "requis", "pseudo"},
		{"unauthorized", ErrUnauthorized, http.StatusUnauthorized, "authentification requise", ""},
		{"forbidden", ErrForbidden, http.StatusForbidden, "action non autorisée", ""},
		{"not found", ErrNotFound, http.StatusNotFound, "ressource introuvable", ""},
		{"conflict", ErrServiceAlreadyBooked, http.StatusConflict, "un échange est déjà en cours pour ce service", ""},
		{"internal default", errors.New("boom"), http.StatusInternalServerError, "erreur interne", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			writeError(rr, tc.err)
			if rr.Code != tc.wantStatus {
				t.Fatalf("status=%d, want %d", rr.Code, tc.wantStatus)
			}
			var got errorResponse
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got.Error != tc.wantError || got.Field != tc.wantField {
				t.Fatalf("body=%+v, want error=%q field=%q", got, tc.wantError, tc.wantField)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type=%q", ct)
			}
		})
	}
}

func TestWriteErrorNil(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	writeError(rr, nil)
	if rr.Code != http.StatusOK || rr.Body.Len() != 0 {
		t.Fatalf("nil error should not write response, got code=%d body=%q", rr.Code, rr.Body.String())
	}
}
