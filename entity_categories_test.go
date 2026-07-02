package main

import "testing"

func TestIsValidCategorie(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{"Informatique", true},
		{"Jardinage", true},
		{"Demenagement", true},
		{"Autre", true},
		{"informatique", false}, // sensible à la casse
		{"", false},
		{"Inconnue", false},
		{" Informatique", false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if got := IsValidCategorie(tc.in); got != tc.want {
				t.Fatalf("IsValidCategorie(%q)=%v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCategoriesNotEmpty(t *testing.T) {
	t.Parallel()
	if len(Categories) == 0 {
		t.Fatal("Categories ne doit pas être vide")
	}
}
