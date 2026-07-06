package main

import (
	"strings"
	"testing"
)

func TestBuildServiceWhere(t *testing.T) {
	t.Parallel()
	where, args := buildServiceWhere(ServiceFilter{OnlyActif: true, Categorie: "Jardinage", Ville: "Paris", Search: "pelouse"})
	for _, want := range []string{"actif = TRUE", "categorie = $1", "ville ILIKE $2", "titre ILIKE $3", "description ILIKE $4"} {
		if !strings.Contains(where, want) {
			t.Fatalf("where=%q missing %q", where, want)
		}
	}
	if len(args) != 4 {
		t.Fatalf("args len=%d", len(args))
	}
}

func TestBuildExchangeWhere(t *testing.T) {
	t.Parallel()
	where, args := buildExchangeWhere(ExchangeFilter{UserID: 7, Status: StatusPending})
	if !strings.Contains(where, "requester_id = $1") || !strings.Contains(where, "owner_id = $1") || !strings.Contains(where, "status = $2") {
		t.Fatalf("bad where=%q", where)
	}
	if len(args) != 2 || args[0] != 7 || args[1] != StatusPending {
		t.Fatalf("bad args=%v", args)
	}
}
