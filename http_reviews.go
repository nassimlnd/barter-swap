package main

import "net/http"

type createReviewRequest struct {
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire"`
}

func (a *App) handleCreateReview(w http.ResponseWriter, r *http.Request) {
	callerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	exchangeID, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	var req createReviewRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	rev, err := a.Reviews.Create(r.Context(), CreateReviewInput{ExchangeID: exchangeID, AuthorID: callerID, Note: req.Note, Commentaire: req.Commentaire})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, reviewToDTO(rev))
}

func (a *App) handleListUserReviews(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := parsePageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}
	reviews, err := a.Reviews.ListByUser(r.Context(), id, page.Limit, page.Offset)
	if err != nil {
		writeError(w, err)
		return
	}
	dtos := make([]reviewDTO, 0, len(reviews))
	for i := range reviews {
		dtos = append(dtos, reviewToDTO(&reviews[i]))
	}
	writeJSON(w, http.StatusOK, paginatedResponse[reviewDTO]{Items: dtos, Total: len(dtos), Limit: page.Limit, Offset: page.Offset})
}

func (a *App) handleListServiceReviews(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	page, err := parsePageParams(r)
	if err != nil {
		writeError(w, err)
		return
	}
	reviews, err := a.Reviews.ListByService(r.Context(), id, page.Limit, page.Offset)
	if err != nil {
		writeError(w, err)
		return
	}
	dtos := make([]reviewDTO, 0, len(reviews))
	for i := range reviews {
		dtos = append(dtos, reviewToDTO(&reviews[i]))
	}
	writeJSON(w, http.StatusOK, paginatedResponse[reviewDTO]{Items: dtos, Total: len(dtos), Limit: page.Limit, Offset: page.Offset})
}
