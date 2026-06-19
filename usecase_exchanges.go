package main

import "context"

type ExchangeUsecase struct {
	Users     UserRepository
	Services  ServiceRepository
	Exchanges ExchangeRepository
	Credits   CreditRepository
	Tx        TxRunner
}

type CreateExchangeInput struct {
	RequesterID int
	ServiceID   int
}

func (uc *ExchangeUsecase) Create(ctx context.Context, in CreateExchangeInput) (*Exchange, error) {
	service, err := uc.Services.GetByID(ctx, in.ServiceID)
	if err != nil {
		return nil, err
	}
	if !service.IsBookable() {
		return nil, ErrServiceInactive
	}
	if service.ProviderID == in.RequesterID {
		return nil, ErrSelfExchange
	}
	requester, err := uc.Users.GetByID(ctx, in.RequesterID)
	if err != nil {
		return nil, err
	}
	if !requester.CanAfford(service.Credits) {
		return nil, ErrInsufficientCredits
	}

	e := &Exchange{
		ServiceID:   service.ID,
		RequesterID: in.RequesterID,
		OwnerID:     service.ProviderID,
		Credits:     service.Credits,
		Status:      StatusPending,
	}
	id, err := uc.Exchanges.Insert(ctx, e)
	if err != nil {
		return nil, err
	}
	return uc.Exchanges.GetByID(ctx, id)
}

func (uc *ExchangeUsecase) Get(ctx context.Context, callerID, id int) (*Exchange, error) {
	e, err := uc.Exchanges.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !e.InvolvesUser(callerID) {
		return nil, ErrForbidden
	}
	return e, nil
}

func (uc *ExchangeUsecase) List(ctx context.Context, callerID int, status ExchangeStatus, limit, offset int) ([]Exchange, int, error) {
	if status != "" && !status.Valid() {
		return nil, 0, ErrInvalidStatus
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 || offset < 0 {
		return nil, 0, ErrInvalidPagination
	}
	return uc.Exchanges.List(ctx, ExchangeFilter{UserID: callerID, Status: status, Limit: limit, Offset: offset})
}

func (uc *ExchangeUsecase) Accept(ctx context.Context, callerID, id int) (*Exchange, error) {
	if err := uc.Tx.InTx(ctx, func(ctx context.Context) error {
		e, err := uc.Exchanges.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if e.OwnerID != callerID {
			return ErrUnauthorizedExchange
		}
		if !e.CanAccept() {
			return ErrInvalidStatus
		}
		requester, err := uc.Users.GetByID(ctx, e.RequesterID)
		if err != nil {
			return err
		}
		if !requester.CanAfford(e.Credits) {
			return ErrInsufficientCredits
		}
		if err := uc.Exchanges.UpdateStatus(ctx, id, StatusPending, StatusAccepted); err != nil {
			return err
		}
		if err := uc.Users.AdjustBalance(ctx, e.RequesterID, -e.Credits); err != nil {
			return err
		}
		_, err = uc.Credits.Insert(ctx, creditTx(e.RequesterID, e.ID, -e.Credits, CreditSpend))
		return err
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, callerID, id)
}

func (uc *ExchangeUsecase) Reject(ctx context.Context, callerID, id int) (*Exchange, error) {
	if err := uc.Tx.InTx(ctx, func(ctx context.Context) error {
		e, err := uc.Exchanges.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if e.OwnerID != callerID {
			return ErrUnauthorizedExchange
		}
		if !e.CanReject() {
			return ErrInvalidStatus
		}
		return uc.Exchanges.UpdateStatus(ctx, id, StatusPending, StatusRejected)
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, callerID, id)
}

func (uc *ExchangeUsecase) Complete(ctx context.Context, callerID, id int) (*Exchange, error) {
	if err := uc.Tx.InTx(ctx, func(ctx context.Context) error {
		e, err := uc.Exchanges.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if e.OwnerID != callerID {
			return ErrUnauthorizedExchange
		}
		if !e.CanComplete() {
			return ErrInvalidStatus
		}
		if err := uc.Exchanges.UpdateStatus(ctx, id, StatusAccepted, StatusCompleted); err != nil {
			return err
		}
		if err := uc.Users.AdjustBalance(ctx, e.OwnerID, e.Credits); err != nil {
			return err
		}
		_, err = uc.Credits.Insert(ctx, creditTx(e.OwnerID, e.ID, e.Credits, CreditEarn))
		return err
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, callerID, id)
}

func (uc *ExchangeUsecase) Cancel(ctx context.Context, callerID, id int) (*Exchange, error) {
	var original Exchange
	if err := uc.Tx.InTx(ctx, func(ctx context.Context) error {
		e, err := uc.Exchanges.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if !e.InvolvesUser(callerID) {
			return ErrUnauthorizedExchange
		}
		if !e.CanCancel() {
			return ErrInvalidStatus
		}
		original = *e
		if err := uc.Exchanges.UpdateStatus(ctx, id, e.Status, StatusCancelled); err != nil {
			return err
		}
		if e.Status == StatusAccepted {
			if err := uc.Users.AdjustBalance(ctx, e.RequesterID, e.Credits); err != nil {
				return err
			}
			_, err = uc.Credits.Insert(ctx, creditTx(e.RequesterID, e.ID, e.Credits, CreditRefund))
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if original.OwnerID == callerID {
		return uc.Get(ctx, original.OwnerID, id)
	}
	return uc.Get(ctx, original.RequesterID, id)
}

func creditTx(userID, exchangeID, amount int, typ CreditTxType) *CreditTransaction {
	id := exchangeID
	return &CreditTransaction{UserID: userID, ExchangeID: &id, Montant: amount, Type: typ}
}
