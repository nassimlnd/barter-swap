package main

import (
	"context"
	"net/http"
)

type createExchangeRequest struct {
	ServiceID int `json:"service_id"`
}

func (a *App) handleCreateExchange(w http.ResponseWriter, r *http.Request) {
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var req createExchangeRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	if req.ServiceID <= 0 {
		writeError(w, newFieldErr(ErrBadRequest, "service_id", "requis"))
		return
	}
	e, err := a.Exchanges.Create(r.Context(), CreateExchangeInput{RequesterID: callerID, ServiceID: req.ServiceID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, exchangeToDTO(e))
}

func (a *App) handleGetExchange(w http.ResponseWriter, r *http.Request) {
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := a.Exchanges.Get(r.Context(), callerID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exchangeToDTO(e))
}

func (a *App) handleListExchanges(w http.ResponseWriter, r *http.Request) {
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := parsePageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}
	status := ExchangeStatus(r.URL.Query().Get("status"))
	items, total, err := a.Exchanges.List(r.Context(), callerID, status, page.Limit, page.Offset)
	if err != nil {
		writeError(w, err)
		return
	}
	dtos := make([]exchangeDTO, 0, len(items))
	for i := range items {
		dtos = append(dtos, exchangeToDTO(&items[i]))
	}
	writeJSON(w, http.StatusOK, paginatedResponse[exchangeDTO]{Items: dtos, Total: total, Limit: page.Limit, Offset: page.Offset})
}

func (a *App) handleAcceptExchange(w http.ResponseWriter, r *http.Request) {
	a.handleExchangeAction(w, r, a.Exchanges.Accept)
}

func (a *App) handleRejectExchange(w http.ResponseWriter, r *http.Request) {
	a.handleExchangeAction(w, r, a.Exchanges.Reject)
}

func (a *App) handleCompleteExchange(w http.ResponseWriter, r *http.Request) {
	a.handleExchangeAction(w, r, a.Exchanges.Complete)
}

func (a *App) handleCancelExchange(w http.ResponseWriter, r *http.Request) {
	a.handleExchangeAction(w, r, a.Exchanges.Cancel)
}

func (a *App) handleExchangeAction(w http.ResponseWriter, r *http.Request, action func(ctx context.Context, callerID, id int) (*Exchange, error)) {
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	e, err := action(r.Context(), callerID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exchangeToDTO(e))
}
