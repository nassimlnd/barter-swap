package main

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateReviewInput(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		note    int
		comment string
		wantOK  bool
	}{
		{"note 1", 1, "", true},
		{"note 5", 5, "super", true},
		{"note 0", 0, "", false},
		{"note 6", 6, "", false},
		{"note négative", -1, "", false},
		{"commentaire trop long", 3, strings.Repeat("c", MaxReviewCommentaireSize+1), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateReviewInput(tc.note, tc.comment)
			gotOK := err == nil
			if gotOK != tc.wantOK {
				t.Fatalf("ok=%v, want %v (err=%v)", gotOK, tc.wantOK, err)
			}
			if !gotOK && !errors.Is(err, ErrBadRequest) {
				t.Fatalf("erreur non réductible à ErrBadRequest : %v", err)
			}
		})
	}
}
