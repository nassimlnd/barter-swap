package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestStatsAPI_UserStats(t *testing.T) {
	app := newTestApp(t)
	providerID, requesterID, _, exchangeID := createCompletedExchange(t, app)
	rr := doReq(app, http.MethodPost, "/api/exchanges/"+itoa(exchangeID)+"/review", `{"note":4,"commentaire":"Bien"}`, itoa(requesterID))
	if rr.Code != http.StatusCreated {
		t.Fatalf("review: %d %s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodGet, "/api/users/"+itoa(providerID)+"/stats", ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("stats status=%d body=%s", rr.Code, rr.Body.String())
	}
	var stats userStatsDTO
	if err := json.NewDecoder(rr.Body).Decode(&stats); err != nil {
		t.Fatalf("decode stats: %v", err)
	}
	if stats.UserID != providerID || stats.ServicesActifs != 1 || stats.EchangesCompletes != 1 || stats.CreditBalance != 11 || stats.NoteMoyenne != 4 || stats.NbAvis != 1 || stats.TotalGagne != 1 {
		t.Fatalf("bad stats: %+v", stats)
	}

	rr = doReq(app, http.MethodGet, "/api/users/"+itoa(requesterID)+"/stats", ``, itoa(requesterID))
	if rr.Code != http.StatusOK {
		t.Fatalf("requester stats status=%d body=%s", rr.Code, rr.Body.String())
	}
	var requesterStats userStatsDTO
	_ = json.NewDecoder(rr.Body).Decode(&requesterStats)
	if requesterStats.TotalDepense != 1 || requesterStats.CreditBalance != 9 {
		t.Fatalf("bad requester stats: %+v", requesterStats)
	}
}

func TestStatsAPI_Errors(t *testing.T) {
	app := newTestApp(t)
	uid := createUserWithSkills(t, app, "alice", nil)
	cases := []struct {
		name   string
		path   string
		userID string
		want   int
	}{
		{"no auth", "/api/users/" + itoa(uid) + "/stats", "", http.StatusUnauthorized},
		{"forbidden", "/api/users/" + itoa(uid) + "/stats", "999", http.StatusForbidden},
		{"not found", "/api/users/999/stats", "999", http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := doReq(app, http.MethodGet, tc.path, ``, tc.userID)
			if rr.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, tc.want, rr.Body.String())
			}
		})
	}
}
