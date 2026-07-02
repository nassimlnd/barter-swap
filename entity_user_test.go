package main

import (
	"errors"
	"strings"
	"testing"
)

func TestUser_CanAfford(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		balance int
		cost    int
		want    bool
	}{
		{"exact", 10, 10, true},
		{"above", 50, 10, true},
		{"below", 5, 10, false},
		{"zero balance, zero cost", 0, 0, true},
		{"negative cost rejected", 10, -1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := &User{CreditBalance: tc.balance}
			if got := u.CanAfford(tc.cost); got != tc.want {
				t.Fatalf("CanAfford(%d) avec balance=%d : got %v, want %v",
					tc.cost, tc.balance, got, tc.want)
			}
		})
	}
}

func TestUser_HasSkill(t *testing.T) {
	t.Parallel()
	u := &User{Skills: []Skill{
		{Nom: "Jardinage", Niveau: NiveauExpert},
		{Nom: "Cuisine", Niveau: NiveauDebutant},
	}}
	cases := []struct {
		query string
		want  bool
	}{
		{"Jardinage", true},
		{"jardinage", true}, // insensible à la casse
		{"  Cuisine  ", true},
		{"Bricolage", false},
		{"", false},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			if got := u.HasSkill(tc.query); got != tc.want {
				t.Fatalf("HasSkill(%q)=%v, want %v", tc.query, got, tc.want)
			}
		})
	}
}

func TestValidateUserInput(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		pseudo  string
		bio     string
		ville   string
		wantOK  bool
		wantFld string // champ attendu en cas d'erreur
	}{
		{"ok minimal", "alice", "", "", true, ""},
		{"ok complet", "bob", "ma bio", "Lyon", true, ""},
		{"pseudo vide", "", "", "", false, "pseudo"},
		{"pseudo espaces", "   ", "", "", false, "pseudo"},
		{"pseudo trop long", strings.Repeat("a", MaxPseudoLength+1), "", "", false, "pseudo"},
		{"bio trop longue", "alice", strings.Repeat("b", MaxBioLength+1), "", false, "bio"},
		{"ville trop longue", "alice", "", strings.Repeat("c", MaxVilleLength+1), false, "ville"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateUserInput(tc.pseudo, tc.bio, tc.ville)
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
				t.Fatalf("attendu erreur réductible à ErrBadRequest, got %v", err)
			}
			var de *DomainError
			if !errors.As(err, &de) {
				t.Fatalf("attendu *DomainError, got %T", err)
			}
			if de.Field != tc.wantFld {
				t.Fatalf("Field=%q, want %q", de.Field, tc.wantFld)
			}
		})
	}
}
