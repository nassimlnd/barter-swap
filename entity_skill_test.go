package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSkillNiveau_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		n    SkillNiveau
		want bool
	}{
		{NiveauDebutant, true},
		{NiveauIntermediaire, true},
		{NiveauExpert, true},
		{"unknown", false},
		{"", false},
		{"DEBUTANT", false}, // sensible à la casse
	}
	for _, tc := range cases {
		t.Run(string(tc.n), func(t *testing.T) {
			if got := tc.n.Valid(); got != tc.want {
				t.Fatalf("(%q).Valid()=%v, want %v", tc.n, got, tc.want)
			}
		})
	}
}

func TestValidateSkill(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		s      Skill
		wantOK bool
	}{
		{"ok", Skill{Nom: "Jardinage", Niveau: NiveauExpert}, true},
		{"nom vide", Skill{Nom: "", Niveau: NiveauExpert}, false},
		{"nom espaces", Skill{Nom: "   ", Niveau: NiveauExpert}, false},
		{"nom trop long", Skill{Nom: strings.Repeat("x", MaxSkillNomLength+1), Niveau: NiveauDebutant}, false},
		{"niveau invalide", Skill{Nom: "Cuisine", Niveau: "guru"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSkill(tc.s)
			gotOK := err == nil
			if gotOK != tc.wantOK {
				t.Fatalf("ValidateSkill : ok=%v, want %v (err=%v)", gotOK, tc.wantOK, err)
			}
			if !gotOK && !errors.Is(err, ErrBadRequest) {
				t.Fatalf("erreur non réductible à ErrBadRequest : %v", err)
			}
		})
	}
}

func TestValidateSkills_Duplicates(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		in     []Skill
		wantOK bool
	}{
		{"vide", nil, true},
		{"un seul", []Skill{{Nom: "Jardinage", Niveau: NiveauExpert}}, true},
		{"deux distincts", []Skill{
			{Nom: "Jardinage", Niveau: NiveauExpert},
			{Nom: "Cuisine", Niveau: NiveauDebutant},
		}, true},
		{"doublon exact", []Skill{
			{Nom: "Jardinage", Niveau: NiveauExpert},
			{Nom: "Jardinage", Niveau: NiveauDebutant},
		}, false},
		{"doublon casse différente", []Skill{
			{Nom: "Jardinage", Niveau: NiveauExpert},
			{Nom: "jardinage", Niveau: NiveauDebutant},
		}, false},
		{"un invalide", []Skill{
			{Nom: "Jardinage", Niveau: NiveauExpert},
			{Nom: "", Niveau: NiveauDebutant},
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSkills(tc.in)
			gotOK := err == nil
			if gotOK != tc.wantOK {
				t.Fatalf("ok=%v, want %v (err=%v)", gotOK, tc.wantOK, err)
			}
		})
	}
}
