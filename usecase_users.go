package main

import (
	"context"
	"strings"
)

type UserUsecase struct {
	Users   UserRepository
	Skills  SkillRepository
	Credits CreditRepository
	Tx      TxRunner
}

type CreateUserInput struct {
	Pseudo string
	Bio    string
	Ville  string
}

type UpdateUserInput struct {
	Pseudo string
	Bio    string
	Ville  string
}

func (uc *UserUsecase) Create(ctx context.Context, in CreateUserInput) (*User, error) {
	if err := ValidateUserInput(in.Pseudo, in.Bio, in.Ville); err != nil {
		return nil, err
	}
	u := &User{
		Pseudo:        strings.TrimSpace(in.Pseudo),
		Bio:           in.Bio,
		Ville:         in.Ville,
		CreditBalance: WelcomeCredits,
	}

	if err := uc.Tx.InTx(ctx, func(ctx context.Context) error {
		id, err := uc.Users.Insert(ctx, u)
		if err != nil {
			return err
		}
		u.ID = id
		_, err = uc.Credits.Insert(ctx, &CreditTransaction{
			UserID:  id,
			Montant: WelcomeCredits,
			Type:    CreditWelcome,
		})
		return err
	}); err != nil {
		return nil, err
	}
	return uc.Get(ctx, u.ID)
}

func (uc *UserUsecase) Get(ctx context.Context, id int) (*User, error) {
	u, err := uc.Users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	sk, err := uc.Skills.ListByUser(ctx, id)
	if err != nil {
		return nil, err
	}
	u.Skills = sk
	return u, nil
}

func (uc *UserUsecase) Update(ctx context.Context, callerID, id int, in UpdateUserInput) (*User, error) {
	if callerID != id {
		return nil, ErrForbidden
	}
	if err := ValidateUserInput(in.Pseudo, in.Bio, in.Ville); err != nil {
		return nil, err
	}
	if err := uc.Users.UpdateProfile(ctx, id, strings.TrimSpace(in.Pseudo), in.Bio, in.Ville); err != nil {
		return nil, err
	}
	return uc.Get(ctx, id)
}

func (uc *UserUsecase) ListSkills(ctx context.Context, userID int) ([]Skill, error) {
	// Vérifie l'existence de l'utilisateur pour retourner 404 plutôt que [] sur
	// un user inexistant.
	if _, err := uc.Users.GetByID(ctx, userID); err != nil {
		return nil, err
	}
	return uc.Skills.ListByUser(ctx, userID)
}

func (uc *UserUsecase) ReplaceSkills(ctx context.Context, callerID, userID int, skills []Skill) ([]Skill, error) {
	if callerID != userID {
		return nil, ErrForbidden
	}
	if err := ValidateSkills(skills); err != nil {
		return nil, err
	}
	for i := range skills {
		skills[i].UserID = userID
		skills[i].Nom = strings.TrimSpace(skills[i].Nom)
	}
	if err := uc.Tx.InTx(ctx, func(ctx context.Context) error {
		if _, err := uc.Users.GetByID(ctx, userID); err != nil {
			return err
		}
		return uc.Skills.ReplaceForUser(ctx, userID, skills)
	}); err != nil {
		return nil, err
	}
	return uc.Skills.ListByUser(ctx, userID)
}
