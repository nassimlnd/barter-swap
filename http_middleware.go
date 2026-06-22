package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

// ctxKey est un type local pour stocker des valeurs dans context.Context.
// Indispensable pour éviter les collisions et satisfaire le linter (govet).
type ctxKey int

const (
	ctxRequestID ctxKey = iota
	ctxUserID
)

// middleware est la signature standard ; chain les compose dans l'ordre fourni
// (le premier middleware est le plus externe, donc le premier à voir la requête).
type middleware func(http.Handler) http.Handler

func chain(h http.Handler, mws ...middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// withMiddlewares applique la chaîne canonique à un handler donné.
// Ordre choisi :
//  1. recover : capture les panics de toute la chaîne en aval
//  2. requestID : génère un ID propagé en contexte + header de réponse
//  3. logging : log structuré (method, path, status, durée, request_id)
//  4. cors : headers + shortcircuit OPTIONS
//  5. auth : extrait X-UserID en contexte (ne rejette pas, les handlers décident)
func (a *App) withMiddlewares(h http.Handler) http.Handler {
	return chain(h,
		a.recoverMW,
		requestIDMW,
		a.loggingMW,
		corsMW,
		authMW,
	)
}

func (a *App) recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				a.Logger.Error("panic recovered",
					"err", fmt.Sprint(rec),
					"stack", string(debug.Stack()),
					"path", r.URL.Path,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"erreur interne"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requestIDMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			rid = newRequestID()
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := context.WithValue(r.Context(), ctxRequestID, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *App) loggingMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		rid, _ := r.Context().Value(ctxRequestID).(string)
		a.Logger.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"dur_ms", time.Since(start).Milliseconds(),
			"request_id", rid,
			"remote", r.RemoteAddr,
		)
	})
}

func corsMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-UserID, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMW lit le header X-UserID et l'injecte dans le contexte si valide.
// Ne rejette PAS la requête : c'est aux handlers protégés d'exiger un userID.
// Cette politique évite de répéter une whitelist de routes publiques.
func authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if v := r.Header.Get("X-UserID"); v != "" {
			if id, err := strconv.Atoi(v); err == nil && id > 0 {
				ctx := context.WithValue(r.Context(), ctxUserID, id)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// userIDFromContext renvoie l'ID utilisateur authentifié (via X-UserID) et un
// booléen indiquant la présence. Utilisé par les handlers qui exigent un caller.
func userIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(ctxUserID).(int)
	return id, ok && id > 0
}

// statusRecorder enveloppe un ResponseWriter pour capturer le code HTTP final
// (loggué dans loggingMW). Implémente uniquement ce dont on a besoin.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// newRequestID génère un ID hex de 16 octets via crypto/rand (sans dep externe).
func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback déterministe : timestamp nanoseconde. Acceptable en cas
		// d'épuisement entropie, ce qui ne devrait jamais arriver en pratique.
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(b[:])
}
