package main

import "context"

type StatsUsecase struct {
	Users UserRepository
	Stats StatsRepository
}

func (uc *StatsUsecase) UserStats(ctx context.Context, callerID, userID int) (UserStats, error) {
	if callerID != userID {
		return UserStats{}, ErrForbidden
	}
	if _, err := uc.Users.GetByID(ctx, userID); err != nil {
		return UserStats{}, err
	}
	return uc.Stats.UserStats(ctx, userID)
}
