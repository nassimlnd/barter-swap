package main

import "net/http"

func (a *App) handleGetUserStats(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	stats, err := a.Stats.UserStats(r.Context(), callerID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, userStatsToDTO(stats))
}
