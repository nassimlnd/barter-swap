package main

import (
	"context"
	"strings"
)

type ServiceUsecase struct {
	Users    UserRepository
	Skills   SkillRepository
	Services ServiceRepository
}

type CreateServiceInput struct {
	ProviderID   int
	Titre        string
	Description  string
	Categorie    string
	DureeMinutes int
	Credits      int
	Ville        string
}

type UpdateServiceInput struct {
	Titre        string
	Description  string
	Categorie    string
	DureeMinutes int
	Credits      int
	Ville        string
	Actif        bool
}

func (uc *ServiceUsecase) Create(ctx context.Context, in CreateServiceInput) (*Service, error) {
	if err := ValidateServiceInput(in.Titre, in.Description, in.Categorie, in.Ville, in.DureeMinutes, in.Credits); err != nil {
		return nil, err
	}
	provider, err := uc.Users.GetByID(ctx, in.ProviderID)
	if err != nil {
		return nil, err
	}
	skills, err := uc.Skills.ListByUser(ctx, in.ProviderID)
	if err != nil {
		return nil, err
	}
	provider.Skills = skills
	if !provider.HasSkill(in.Categorie) {
		return nil, ErrSkillNotOwned
	}

	s := &Service{
		ProviderID:   in.ProviderID,
		Titre:        strings.TrimSpace(in.Titre),
		Description:  in.Description,
		Categorie:    in.Categorie,
		DureeMinutes: in.DureeMinutes,
		Credits:      in.Credits,
		Ville:        in.Ville,
		Actif:        true,
	}
	id, err := uc.Services.Insert(ctx, s)
	if err != nil {
		return nil, err
	}
	return uc.Services.GetByID(ctx, id)
}

func (uc *ServiceUsecase) Get(ctx context.Context, id int) (*Service, error) {
	return uc.Services.GetByID(ctx, id)
}

func (uc *ServiceUsecase) Update(ctx context.Context, callerID, id int, in UpdateServiceInput) (*Service, error) {
	if err := ValidateServiceInput(in.Titre, in.Description, in.Categorie, in.Ville, in.DureeMinutes, in.Credits); err != nil {
		return nil, err
	}
	current, err := uc.Services.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.ProviderID != callerID {
		return nil, ErrForbidden
	}
	provider, err := uc.Users.GetByID(ctx, callerID)
	if err != nil {
		return nil, err
	}
	skills, err := uc.Skills.ListByUser(ctx, callerID)
	if err != nil {
		return nil, err
	}
	provider.Skills = skills
	if !provider.HasSkill(in.Categorie) {
		return nil, ErrSkillNotOwned
	}
	s := &Service{
		ProviderID:   callerID,
		Titre:        strings.TrimSpace(in.Titre),
		Description:  in.Description,
		Categorie:    in.Categorie,
		DureeMinutes: in.DureeMinutes,
		Credits:      in.Credits,
		Ville:        in.Ville,
		Actif:        in.Actif,
	}
	if err := uc.Services.Update(ctx, id, s); err != nil {
		return nil, err
	}
	return uc.Services.GetByID(ctx, id)
}

func (uc *ServiceUsecase) Delete(ctx context.Context, callerID, id int) error {
	s, err := uc.Services.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if s.ProviderID != callerID {
		return ErrForbidden
	}
	return uc.Services.Delete(ctx, id)
}

func (uc *ServiceUsecase) List(ctx context.Context, f ServiceFilter) ([]Service, int, error) {
	if f.Categorie != "" && !IsValidCategorie(f.Categorie) {
		return nil, 0, ErrInvalidCategorie
	}
	if f.Limit <= 0 {
		f.Limit = 20
	}
	if f.Limit > 100 || f.Offset < 0 {
		return nil, 0, ErrInvalidPagination
	}
	f.OnlyActif = true
	return uc.Services.List(ctx, f)
}
