package main

import (
	"errors"
	"net/http"
)

// errorResponse est le format unique des erreurs HTTP exposées aux clients.
type errorResponse struct {
	Error string `json:"error"`
	Field string `json:"field,omitempty"`
}

// writeError mappe les erreurs métier (sentinelles + DomainError) vers des
// codes HTTP cohérents. Les handlers ne doivent jamais décider eux-mêmes du
// status code d'une erreur métier : ils appellent cette fonction.
func writeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	status := http.StatusInternalServerError
	msg := "erreur interne"
	field := ""

	var de *DomainError
	if errors.As(err, &de) {
		msg = de.Message
		field = de.Field
	} else if errors.Is(err, ErrInternal) {
		msg = ErrInternal.Error()
	}

	switch {
	case errors.Is(err, ErrBadRequest):
		status = http.StatusBadRequest
		if msg == "erreur interne" {
			msg = ErrBadRequest.Error()
		}
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
		if msg == "erreur interne" {
			msg = ErrUnauthorized.Error()
		}
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
		if msg == "erreur interne" {
			msg = ErrForbidden.Error()
		}
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
		if msg == "erreur interne" {
			msg = ErrNotFound.Error()
		}
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
		if msg == "erreur interne" {
			msg = ErrConflict.Error()
		}
	}

	writeJSON(w, status, errorResponse{Error: msg, Field: field})
}
