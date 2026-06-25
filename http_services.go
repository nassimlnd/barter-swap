package main

import "net/http"

type serviceRequest struct {
	Titre        string `json:"titre"`
	Description  string `json:"description"`
	Categorie    string `json:"categorie"`
	DureeMinutes int    `json:"duree_minutes"`
	Credits      int    `json:"credits"`
	Ville        string `json:"ville"`
	Actif        *bool  `json:"actif,omitempty"`
}

func (a *App) handleCreateService(w http.ResponseWriter, r *http.Request) {
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req serviceRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	s, err := a.Services.Create(r.Context(), CreateServiceInput{
		ProviderID:   callerID,
		Titre:        req.Titre,
		Description:  req.Description,
		Categorie:    req.Categorie,
		DureeMinutes: req.DureeMinutes,
		Credits:      req.Credits,
		Ville:        req.Ville,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, serviceToDTO(s))
}

func (a *App) handleGetService(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	s, err := a.Services.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, serviceToDTO(s))
}

func (a *App) handleUpdateService(w http.ResponseWriter, r *http.Request) {
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
	var req serviceRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	actif := true
	if req.Actif != nil {
		actif = *req.Actif
	}
	s, err := a.Services.Update(r.Context(), callerID, id, UpdateServiceInput{
		Titre:        req.Titre,
		Description:  req.Description,
		Categorie:    req.Categorie,
		DureeMinutes: req.DureeMinutes,
		Credits:      req.Credits,
		Ville:        req.Ville,
		Actif:        actif,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, serviceToDTO(s))
}

func (a *App) handleDeleteService(w http.ResponseWriter, r *http.Request) {
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
	if err := a.Services.Delete(r.Context(), callerID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) handleListServices(w http.ResponseWriter, r *http.Request) {
	page, err := parsePageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}
	q := r.URL.Query()
	items, total, err := a.Services.List(r.Context(), ServiceFilter{
		Categorie: q.Get("categorie"),
		Ville:     q.Get("ville"),
		Search:    q.Get("search"),
		Limit:     page.Limit,
		Offset:    page.Offset,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	dtos := make([]serviceDTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, serviceToDTO(&items[i]))
	}
	writeJSON(w, http.StatusOK, paginatedResponse[serviceDTO]{
		Items:  dtos,
		Total:  total,
		Limit:  page.Limit,
		Offset: page.Offset,
	})
}
