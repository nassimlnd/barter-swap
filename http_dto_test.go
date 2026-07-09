package main

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	t.Parallel()
	if got := formatTime(time.Time{}); got != "" {
		t.Fatalf("zero time -> %q, want empty", got)
	}
	tm := time.Date(2026, 5, 30, 14, 23, 0, 0, time.FixedZone("CEST", 2*60*60))
	if got := formatTime(tm); got != "2026-05-30T12:23:00Z" {
		t.Fatalf("formatTime=%q", got)
	}
}

func TestDTOConversions(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	u := &User{ID: 1, Pseudo: "alice", Bio: "bio", Ville: "Paris", CreditBalance: 10, CreatedAt: now,
		Skills: []Skill{{Nom: "Jardinage", Niveau: NiveauExpert}}}
	ud := userToDTO(u)
	if ud.ID != 1 || ud.Pseudo != "alice" || ud.CreatedAt != "2026-01-15T12:00:00Z" || len(ud.Skills) != 1 {
		t.Fatalf("bad user dto: %+v", ud)
	}

	s := &Service{ID: 2, ProviderID: 1, Titre: "Aide", Categorie: "Jardinage", DureeMinutes: 60, Credits: 1, Actif: true, CreatedAt: now}
	if got := serviceToDTO(s); got.ID != 2 || !got.Actif || got.CreatedAt == "" {
		t.Fatalf("bad service dto: %+v", got)
	}

	e := &Exchange{ID: 3, ServiceID: 2, RequesterID: 4, OwnerID: 1, Status: StatusPending, CreatedAt: now, UpdatedAt: now}
	if got := exchangeToDTO(e); got.Status != "pending" || got.UpdatedAt == "" {
		t.Fatalf("bad exchange dto: %+v", got)
	}

	r := &Review{ID: 4, ExchangeID: 3, AuthorID: 4, TargetID: 1, Note: 5, Commentaire: "top", CreatedAt: now}
	if got := reviewToDTO(r); got.Note != 5 || got.CreatedAt == "" {
		t.Fatalf("bad review dto: %+v", got)
	}

	stats := userStatsToDTO(UserStats{UserID: 1, ServicesActifs: 2, CreditBalance: 10})
	if stats.UserID != 1 || stats.ServicesActifs != 2 || stats.CreditBalance != 10 {
		t.Fatalf("bad stats dto: %+v", stats)
	}
}
