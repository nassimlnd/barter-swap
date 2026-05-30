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

	// Les handlers métier seront branchés ici au fil des jalons :
	//   mux.HandleFunc("POST /api/users", a.handleCreateUser)
	//   mux.HandleFunc("GET /api/users/{id}", a.handleGetUser)
	//   ...

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
