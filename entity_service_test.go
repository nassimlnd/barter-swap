package main

import (
	"errors"
	"strings"
	"testing"
)

func TestService_IsBookable(t *testing.T) {
	t.Parallel()
	cases := []struct {
		actif bool
		want  bool
	}{
		{true, true},
		{false, false},
	}
	for _, tc := range cases {
		s := &Service{Actif: tc.actif}
		if got := s.IsBookable(); got != tc.want {
			t.Fatalf("IsBookable() avec actif=%v : got %v, want %v", tc.actif, got, tc.want)
		}
	}
}

func TestValidateServiceInput(t *testing.T) {
	t.Parallel()
	type in struct {
		titre, description, categorie, ville string
		duree, credits                       int
	}
	good := in{"Cours de jardinage", "Désherbage et plantations", "Jardinage", "Lyon", 60, 1}

	cases := []struct {
		name    string
		mut     func(*in)
		wantOK  bool
		wantFld string
	}{
		{"ok", func(i *in) {}, true, ""},
		{"titre vide", func(i *in) { i.titre = "" }, false, "titre"},
		{"titre espaces", func(i *in) { i.titre = "  " }, false, "titre"},
		{"titre trop long", func(i *in) { i.titre = strings.Repeat("t", MaxServiceTitreLength+1) }, false, "titre"},
		{"description trop longue", func(i *in) { i.description = strings.Repeat("d", MaxServiceDescriptionLength+1) }, false, "description"},
		{"categorie inconnue", func(i *in) { i.categorie = "Inconnue" }, false, ""},
		{"ville trop longue", func(i *in) { i.ville = strings.Repeat("v", MaxServiceVilleLength+1) }, false, "ville"},
		{"duree nulle", func(i *in) { i.duree = 0 }, false, "duree_minutes"},
		{"duree négative", func(i *in) { i.duree = -1 }, false, "duree_minutes"},
		{"credits nuls", func(i *in) { i.credits = 0 }, false, "credits"},
		{"credits négatifs", func(i *in) { i.credits = -5 }, false, "credits"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			i := good
			tc.mut(&i)
			err := ValidateServiceInput(i.titre, i.description, i.categorie, i.ville, i.duree, i.credits)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("attendu nil, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("attendu une erreur, got nil")
			}
			if !errors.Is(err, ErrBadRequest) {
				t.Fatalf("erreur non réductible à ErrBadRequest : %v", err)
			}
			if tc.wantFld != "" {
				var de *DomainError
				if !errors.As(err, &de) {
					t.Fatalf("attendu *DomainError, got %T", err)
				}
				if de.Field != tc.wantFld {
					t.Fatalf("Field=%q, want %q", de.Field, tc.wantFld)
				}
			}
		})
	}
}
