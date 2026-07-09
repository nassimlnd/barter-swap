package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func createCompletedExchange(t *testing.T, app *App) (providerID, requesterID, serviceID, exchangeID int) {
	t.Helper()
	providerID = createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})
	requesterID = createUserWithSkills(t, app, "requester", nil)
	serviceID = createServiceForTest(t, app, providerID, "Jardinage")
	rr := doReq(app, http.MethodPost, "/api/exchanges", `{"service_id":`+itoa(serviceID)+`}`, itoa(requesterID))
	if rr.Code != http.StatusCreated {
		t.Fatalf("create exchange: %d %s", rr.Code, rr.Body.String())
	}
	var ex exchangeDTO
	_ = json.NewDecoder(rr.Body).Decode(&ex)
	exchangeID = ex.ID
	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(exchangeID)+"/accept", ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("accept: %d %s", rr.Code, rr.Body.String())
	}
	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(exchangeID)+"/complete", ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", rr.Code, rr.Body.String())
	}
	return providerID, requesterID, serviceID, exchangeID
}

func TestReviewsAPI_CreateAndList(t *testing.T) {
	app := newTestApp(t)
	providerID, requesterID, serviceID, exchangeID := createCompletedExchange(t, app)

	rr := doReq(app, http.MethodPost, "/api/exchanges/"+itoa(exchangeID)+"/review", `{"note":5,"commentaire":"Excellent"}`, itoa(requesterID))
	if rr.Code != http.StatusCreated {
		t.Fatalf("create review: %d %s", rr.Code, rr.Body.String())
	}
	var rev reviewDTO
	if err := json.NewDecoder(rr.Body).Decode(&rev); err != nil {
		t.Fatalf("decode review: %v", err)
	}
	if rev.AuthorID != requesterID || rev.TargetID != providerID || rev.Note != 5 {
		t.Fatalf("bad review: %+v", rev)
	}

	rr = doReq(app, http.MethodGet, "/api/users/"+itoa(providerID)+"/reviews", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list user reviews: %d %s", rr.Code, rr.Body.String())
	}
	var userPage paginatedResponse[reviewDTO]
	_ = json.NewDecoder(rr.Body).Decode(&userPage)
	if userPage.Total != 1 || userPage.Items[0].ID != rev.ID {
		t.Fatalf("bad user review page: %+v", userPage)
	}

	rr = doReq(app, http.MethodGet, "/api/services/"+itoa(serviceID)+"/reviews", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list service reviews: %d %s", rr.Code, rr.Body.String())
	}
	var servicePage paginatedResponse[reviewDTO]
	_ = json.NewDecoder(rr.Body).Decode(&servicePage)
	if servicePage.Total != 1 || servicePage.Items[0].ID != rev.ID {
		t.Fatalf("bad service review page: %+v", servicePage)
	}

	rr = doReq(app, http.MethodPost, "/api/exchanges/"+itoa(exchangeID)+"/review", `{"note":4}`, itoa(requesterID))
	if rr.Code != http.StatusConflict {
		t.Fatalf("duplicate review status=%d want 409 body=%s", rr.Code, rr.Body.String())
	}
}

func TestReviewsAPI_Errors(t *testing.T) {
	app := newTestApp(t)
	providerID := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})
	requesterID := createUserWithSkills(t, app, "requester", nil)
	serviceID := createServiceForTest(t, app, providerID, "Jardinage")
	rr := doReq(app, http.MethodPost, "/api/exchanges", `{"service_id":`+itoa(serviceID)+`}`, itoa(requesterID))
	var ex exchangeDTO
	_ = json.NewDecoder(rr.Body).Decode(&ex)

	cases := []struct {
		name   string
		body   string
		userID string
		want   int
	}{
		{"no auth", `{"note":5}`, "", http.StatusUnauthorized},
		{"invalid note", `{"note":6}`, itoa(requesterID), http.StatusBadRequest},
		{"not completed", `{"note":5}`, itoa(requesterID), http.StatusBadRequest},
		{"non participant", `{"note":5}`, "999", http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := doReq(app, http.MethodPost, "/api/exchanges/"+itoa(ex.ID)+"/review", tc.body, tc.userID)
			if rr.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, tc.want, rr.Body.String())
			}
		})
	}
}
