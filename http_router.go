package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// buildRouter déclare toutes les routes de l'API. Les patterns utilisent la
// syntaxe du ServeMux de Go 1.22 : `METHOD /path/{param}`.
func (a *App) buildRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", a.handleHealth)

	mux.HandleFunc("POST /api/users", a.handleCreateUser)
	mux.HandleFunc("GET /api/users/{id}", a.handleGetUser)
	mux.HandleFunc("PUT /api/users/{id}", a.handleUpdateUser)
	mux.HandleFunc("GET /api/users/{id}/skills", a.handleGetUserSkills)
	mux.HandleFunc("PUT /api/users/{id}/skills", a.handleReplaceUserSkills)

	mux.HandleFunc("GET /api/services", a.handleListServices)
	mux.HandleFunc("POST /api/services", a.handleCreateService)
	mux.HandleFunc("GET /api/services/{id}", a.handleGetService)
	mux.HandleFunc("PUT /api/services/{id}", a.handleUpdateService)
	mux.HandleFunc("DELETE /api/services/{id}", a.handleDeleteService)

	mux.HandleFunc("POST /api/exchanges", a.handleCreateExchange)
	mux.HandleFunc("GET /api/exchanges", a.handleListExchanges)
	mux.HandleFunc("GET /api/exchanges/{id}", a.handleGetExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", a.handleAcceptExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", a.handleRejectExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", a.handleCompleteExchange)
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", a.handleCancelExchange)
	mux.HandleFunc("POST /api/exchanges/{id}/review", a.handleCreateReview)
	mux.HandleFunc("GET /api/users/{id}/reviews", a.handleListUserReviews)
	mux.HandleFunc("GET /api/users/{id}/stats", a.handleGetUserStats)
	mux.HandleFunc("GET /api/services/{id}/reviews", a.handleListServiceReviews)

	mux.HandleFunc("POST /debug/seed", a.handleDebugSeed)
	mux.HandleFunc("GET /debug/reconcile", a.handleDebugReconcile)

	return mux
}

// handleHealth retourne 200 si l'API est vivante et la DB joignable.
// Codé directement ici (et non dans un http_health.go dédié) parce qu'il n'a
// aucune logique métier : c'est purement de la plomberie.
func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	resp := map[string]string{
		"status": "ok",
		"db":     "up",
	}
	status := http.StatusOK

	if err := a.DB.PingContext(ctx); err != nil {
		resp["db"] = "down"
		resp["status"] = "degraded"
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}
