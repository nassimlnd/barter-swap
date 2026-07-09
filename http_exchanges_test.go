package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func createServiceForTest(t *testing.T, app *App, providerID int, categorie string) int {
	t.Helper()
	body := `{"titre":"Service test","description":"desc","categorie":"` + categorie + `","duree_minutes":60,"credits":1,"ville":"Paris"}`
	rr := doReq(app, http.MethodPost, "/api/services", body, itoa(providerID))
	if rr.Code != http.StatusCreated {
		t.Fatalf("create service: status=%d body=%s", rr.Code, rr.Body.String())
	}
	var svc serviceDTO
	if err := json.NewDecoder(rr.Body).Decode(&svc); err != nil {
		t.Fatalf("decode service: %v", err)
	}
	return svc.ID
}

func TestExchangesAPI_CreateListGetAndConflict(t *testing.T) {
	app := newTestApp(t)
	providerID := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})
	requesterID := createUserWithSkills(t, app, "requester", nil)
	otherID := createUserWithSkills(t, app, "other", nil)
	serviceID := createServiceForTest(t, app, providerID, "Jardinage")

	rr := doReq(app, http.MethodPost, "/api/exchanges", `{"service_id":`+itoa(serviceID)+`}`, itoa(requesterID))
	if rr.Code != http.StatusCreated {
		t.Fatalf("create exchange status=%d body=%s", rr.Code, rr.Body.String())
	}
	var ex exchangeDTO
	if err := json.NewDecoder(rr.Body).Decode(&ex); err != nil {
		t.Fatalf("decode exchange: %v", err)
	}
	if ex.ID == 0 || ex.Status != "pending" || ex.RequesterID != requesterID || ex.OwnerID != providerID {
		t.Fatalf("bad exchange: %+v", ex)
	}

	// Conflit DB via index unique partiel : même service déjà pending.
	rr = doReq(app, http.MethodPost, "/api/exchanges", `{"service_id":`+itoa(serviceID)+`}`, itoa(otherID))
	if rr.Code != http.StatusConflict {
		t.Fatalf("second exchange status=%d want 409 body=%s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodGet, "/api/exchanges?status=pending", ``, itoa(requesterID))
	if rr.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", rr.Code, rr.Body.String())
	}
	var page paginatedResponse[exchangeDTO]
	if err := json.NewDecoder(rr.Body).Decode(&page); err != nil {
		t.Fatalf("decode page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != ex.ID {
		t.Fatalf("bad list page: %+v", page)
	}

	rr = doReq(app, http.MethodGet, "/api/exchanges/"+itoa(ex.ID), ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("provider get status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodGet, "/api/exchanges/"+itoa(ex.ID), ``, itoa(otherID))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("other get status=%d want 403 body=%s", rr.Code, rr.Body.String())
	}
}

func TestExchangesAPI_Errors(t *testing.T) {
	app := newTestApp(t)
	providerID := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})
	requesterID := createUserWithSkills(t, app, "requester", nil)
	serviceID := createServiceForTest(t, app, providerID, "Jardinage")

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		userID string
		want   int
	}{
		{"create no auth", http.MethodPost, "/api/exchanges", `{"service_id":1}`, "", http.StatusUnauthorized},
		{"service id required", http.MethodPost, "/api/exchanges", `{"service_id":0}`, itoa(requesterID), http.StatusBadRequest},
		{"self exchange", http.MethodPost, "/api/exchanges", `{"service_id":` + itoa(serviceID) + `}`, itoa(providerID), http.StatusBadRequest},
		{"list no auth", http.MethodGet, "/api/exchanges", ``, "", http.StatusUnauthorized},
		{"invalid status", http.MethodGet, "/api/exchanges?status=weird", ``, itoa(requesterID), http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := doReq(app, tc.method, tc.path, tc.body, tc.userID)
			if rr.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, tc.want, rr.Body.String())
			}
		})
	}
}

func TestExchangesAPI_AcceptCompleteTransfersCredits(t *testing.T) {
	app := newTestApp(t)
	providerID := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})
	requesterID := createUserWithSkills(t, app, "requester", nil)
	serviceID := createServiceForTest(t, app, providerID, "Jardinage")

	rr := doReq(app, http.MethodPost, "/api/exchanges", `{"service_id":`+itoa(serviceID)+`}`, itoa(requesterID))
	if rr.Code != http.StatusCreated {
		t.Fatalf("create exchange: %d %s", rr.Code, rr.Body.String())
	}
	var ex exchangeDTO
	_ = json.NewDecoder(rr.Body).Decode(&ex)

	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(ex.ID)+"/accept", ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("accept: %d %s", rr.Code, rr.Body.String())
	}
	_ = json.NewDecoder(rr.Body).Decode(&ex)
	if ex.Status != "accepted" {
		t.Fatalf("status=%s want accepted", ex.Status)
	}
	requester, _ := app.Users.Get(t.Context(), requesterID)
	provider, _ := app.Users.Get(t.Context(), providerID)
	if requester.CreditBalance != 9 || provider.CreditBalance != 10 {
		t.Fatalf("after accept balances requester=%d provider=%d", requester.CreditBalance, provider.CreditBalance)
	}

	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(ex.ID)+"/complete", ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("complete: %d %s", rr.Code, rr.Body.String())
	}
	_ = json.NewDecoder(rr.Body).Decode(&ex)
	if ex.Status != "completed" {
		t.Fatalf("status=%s want completed", ex.Status)
	}
	requester, _ = app.Users.Get(t.Context(), requesterID)
	provider, _ = app.Users.Get(t.Context(), providerID)
	if requester.CreditBalance != 9 || provider.CreditBalance != 11 {
		t.Fatalf("after complete balances requester=%d provider=%d", requester.CreditBalance, provider.CreditBalance)
	}
	requesterSum, _ := app.Users.Credits.SumByUser(t.Context(), requesterID)
	providerSum, _ := app.Users.Credits.SumByUser(t.Context(), providerID)
	if requesterSum != requester.CreditBalance || providerSum != provider.CreditBalance {
		t.Fatalf("journal mismatch requester sum=%d balance=%d provider sum=%d balance=%d", requesterSum, requester.CreditBalance, providerSum, provider.CreditBalance)
	}
}

func TestExchangesAPI_CancelAcceptedRefundsRequester(t *testing.T) {
	app := newTestApp(t)
	providerID := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})
	requesterID := createUserWithSkills(t, app, "requester", nil)
	serviceID := createServiceForTest(t, app, providerID, "Jardinage")

	rr := doReq(app, http.MethodPost, "/api/exchanges", `{"service_id":`+itoa(serviceID)+`}`, itoa(requesterID))
	var ex exchangeDTO
	_ = json.NewDecoder(rr.Body).Decode(&ex)
	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(ex.ID)+"/accept", ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("accept: %d %s", rr.Code, rr.Body.String())
	}
	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(ex.ID)+"/cancel", ``, itoa(requesterID))
	if rr.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", rr.Code, rr.Body.String())
	}
	_ = json.NewDecoder(rr.Body).Decode(&ex)
	if ex.Status != "cancelled" {
		t.Fatalf("status=%s want cancelled", ex.Status)
	}
	requester, _ := app.Users.Get(t.Context(), requesterID)
	if requester.CreditBalance != 10 {
		t.Fatalf("requester balance=%d want 10", requester.CreditBalance)
	}
	requesterSum, _ := app.Users.Credits.SumByUser(t.Context(), requesterID)
	if requesterSum != requester.CreditBalance {
		t.Fatalf("journal mismatch sum=%d balance=%d", requesterSum, requester.CreditBalance)
	}
}

func TestExchangesAPI_InvalidTransitions(t *testing.T) {
	app := newTestApp(t)
	providerID := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})
	requesterID := createUserWithSkills(t, app, "requester", nil)
	serviceID := createServiceForTest(t, app, providerID, "Jardinage")
	rr := doReq(app, http.MethodPost, "/api/exchanges", `{"service_id":`+itoa(serviceID)+`}`, itoa(requesterID))
	var ex exchangeDTO
	_ = json.NewDecoder(rr.Body).Decode(&ex)

	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(ex.ID)+"/complete", ``, itoa(providerID))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("complete pending status=%d want 400 body=%s", rr.Code, rr.Body.String())
	}
	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(ex.ID)+"/reject", ``, itoa(providerID))
	if rr.Code != http.StatusOK {
		t.Fatalf("reject pending status=%d body=%s", rr.Code, rr.Body.String())
	}
	rr = doReq(app, http.MethodPut, "/api/exchanges/"+itoa(ex.ID)+"/accept", ``, itoa(providerID))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("accept rejected status=%d want 400 body=%s", rr.Code, rr.Body.String())
	}
}
