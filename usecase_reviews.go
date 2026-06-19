package main

import "context"

type ReviewUsecase struct {
	Exchanges ExchangeRepository
	Reviews   ReviewRepository
}

type CreateReviewInput struct {
	ExchangeID  int
	AuthorID    int
	Note        int
	Commentaire string
}

func (uc *ReviewUsecase) Create(ctx context.Context, in CreateReviewInput) (*Review, error) {
	if err := ValidateReviewInput(in.Note, in.Commentaire); err != nil {
		return nil, err
	}
	ex, err := uc.Exchanges.GetByID(ctx, in.ExchangeID)
	if err != nil {
		return nil, err
	}
	if !ex.InvolvesUser(in.AuthorID) {
		return nil, ErrCannotReview
	}
	if ex.Status != StatusCompleted {
		return nil, ErrExchangeNotCompleted
	}
	targetID := ex.OwnerID
	if in.AuthorID == ex.OwnerID {
		targetID = ex.RequesterID
	}
	if targetID == in.AuthorID {
		return nil, ErrSelfReview
	}
	r := &Review{ExchangeID: in.ExchangeID, AuthorID: in.AuthorID, TargetID: targetID, Note: in.Note, Commentaire: in.Commentaire}
	id, err := uc.Reviews.Insert(ctx, r)
	if err != nil {
		return nil, err
	}
	r.ID = id
	return r, nil
}

func (uc *ReviewUsecase) ListByUser(ctx context.Context, targetUserID, limit, offset int) ([]Review, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 || offset < 0 {
		return nil, ErrInvalidPagination
	}
	return uc.Reviews.ListByUser(ctx, targetUserID, limit, offset)
}

func (uc *ReviewUsecase) ListByService(ctx context.Context, serviceID, limit, offset int) ([]Review, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 || offset < 0 {
		return nil, ErrInvalidPagination
	}
	return uc.Reviews.ListByService(ctx, serviceID, limit, offset)
}
