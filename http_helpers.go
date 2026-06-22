package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const maxJSONBodySize = 1 << 20 // 1 MiB

// readJSON lit un payload JSON strict : taille limitée, champs inconnus
// refusés, et un seul objet JSON autorisé dans le body.
func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return jsonDecodeError(err)
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return newDomainErr(ErrBadRequest, "le body doit contenir un seul objet JSON")
	}
	return nil
}

func jsonDecodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.As(err, &syntaxErr):
		return newDomainErr(ErrBadRequest, fmt.Sprintf("JSON invalide à l'octet %d", syntaxErr.Offset))
	case errors.As(err, &typeErr):
		if typeErr.Field != "" {
			return newFieldErr(ErrBadRequest, typeErr.Field, "type JSON invalide")
		}
		return newDomainErr(ErrBadRequest, "type JSON invalide")
	case errors.As(err, &maxBytesErr):
		return newDomainErr(ErrBadRequest, "body JSON trop volumineux")
	case errors.Is(err, io.EOF):
		return newDomainErr(ErrBadRequest, "body JSON requis")
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		field := strings.TrimPrefix(err.Error(), "json: unknown field ")
		field = strings.Trim(field, "\"")
		return newFieldErr(ErrBadRequest, field, "champ inconnu")
	default:
		return newDomainErr(ErrBadRequest, "JSON invalide")
	}
}

// writeJSON encode payload en JSON avec Content-Type uniforme.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseIDParam(r *http.Request, name string) (int, error) {
	v := r.PathValue(name)
	id, err := strconv.Atoi(v)
	if err != nil || id <= 0 {
		return 0, newFieldErr(ErrBadRequest, name, "identifiant invalide")
	}
	return id, nil
}

// requireUserID extrait X-UserID depuis le contexte. Les handlers protégés
// l'utilisent pour renvoyer 401 si l'appelant n'est pas authentifié.
func requireUserID(r *http.Request) (int, error) {
	uid, ok := userIDFromContext(r.Context())
	if !ok {
		return 0, ErrUnauthorized
	}
	return uid, nil
}

type pageParams struct {
	Limit  int
	Offset int
}

func parsePageParams(r *http.Request) (pageParams, error) {
	q := r.URL.Query()
	limit := 20
	offset := 0
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 || v > 100 {
			return pageParams{}, ErrInvalidPagination
		}
		limit = v
	}
	if raw := q.Get("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			return pageParams{}, ErrInvalidPagination
		}
		offset = v
	}
	return pageParams{Limit: limit, Offset: offset}, nil
}

type paginatedResponse[T any] struct {
	Items  []T `json:"items"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}
