package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"
)

// App rassemble les dépendances partagées par les couches HTTP et SQL.
// Les use cases reçoivent les ports (interfaces) ; App reste le composant racine
// instancié dans main et passé à chaque adapter.
type App struct {
	DB     *sql.DB
	Logger *slog.Logger
	Now    func() time.Time
}

// buildHTTPHandler construit la chaîne complète : router + middlewares.
// Les jalons ultérieurs y branchent les handlers de chaque ressource.
func (a *App) buildHTTPHandler() http.Handler {
	mux := a.buildRouter()
	return a.withMiddlewares(mux)
}
