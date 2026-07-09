package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

func createUserWithSkills(t *testing.T, app *App, pseudo string, skills []skillDTO) int {
	t.Helper()
	rr := doReq(app, http.MethodPost, "/api/users", `{"pseudo":"`+pseudo+`"}`, "")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create user %s: status=%d body=%s", pseudo, rr.Code, rr.Body.String())
	}
	var u userDTO
	if err := json.NewDecoder(rr.Body).Decode(&u); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	body, _ := json.Marshal(replaceSkillsRequest{Skills: skills})
	rr = doReq(app, http.MethodPut, "/api/users/"+itoa(u.ID)+"/skills", string(body), itoa(u.ID))
	if rr.Code != http.StatusOK {
		t.Fatalf("replace skills: status=%d body=%s", rr.Code, rr.Body.String())
	}
	return u.ID
}

func itoa(v int) string { return strconv.Itoa(v) }

func TestServicesAPI_CRUDAndFilters(t *testing.T) {
	app := newTestApp(t)
	uid := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}, {Nom: "Cuisine", Niveau: "debutant"}})

	rr := doReq(app, http.MethodPost, "/api/services", `{"titre":"Tondre pelouse","description":"Aide jardin","categorie":"Jardinage","duree_minutes":60,"credits":1,"ville":"Paris"}`, itoa(uid))
	if rr.Code != http.StatusCreated {
		t.Fatalf("create service status=%d body=%s", rr.Code, rr.Body.String())
	}
	var svc serviceDTO
	if err := json.NewDecoder(rr.Body).Decode(&svc); err != nil {
		t.Fatalf("decode service: %v", err)
	}
	if svc.ID == 0 || svc.ProviderID != uid || !svc.Actif || svc.Categorie != "Jardinage" {
		t.Fatalf("bad service dto: %+v", svc)
	}

	rr = doReq(app, http.MethodGet, "/api/services?categorie=Jardinage&ville=Paris&search=pelouse&limit=10", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("list services status=%d body=%s", rr.Code, rr.Body.String())
	}
	var page paginatedResponse[serviceDTO]
	if err := json.NewDecoder(rr.Body).Decode(&page); err != nil {
		t.Fatalf("decode page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != svc.ID {
		t.Fatalf("bad page: %+v", page)
	}

	rr = doReq(app, http.MethodPut, "/api/services/"+itoa(svc.ID), `{"titre":"Cuisine maison","description":"Cours simple","categorie":"Cuisine","duree_minutes":90,"credits":2,"ville":"Lyon","actif":true}`, itoa(uid))
	if rr.Code != http.StatusOK {
		t.Fatalf("update service status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodDelete, "/api/services/"+itoa(svc.ID), ``, itoa(uid))
	if rr.Code != http.StatusNoContent {
		t.Fatalf("delete service status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodGet, "/api/services/"+itoa(svc.ID), ``, "")
	if rr.Code != http.StatusNotFound {
		t.Fatalf("get deleted status=%d want 404 body=%s", rr.Code, rr.Body.String())
	}
}

func TestServicesAPI_Errors(t *testing.T) {
	app := newTestApp(t)
	uid := createUserWithSkills(t, app, "provider", []skillDTO{{Nom: "Jardinage", Niveau: "expert"}})

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		userID string
		want   int
	}{
		{"create no auth", http.MethodPost, "/api/services", `{}`, "", http.StatusUnauthorized},
		{"skill not owned", http.MethodPost, "/api/services", `{"titre":"Cours","categorie":"Cuisine","duree_minutes":60,"credits":1}`, itoa(uid), http.StatusBadRequest},
		{"invalid category list", http.MethodGet, "/api/services?categorie=Inconnue", ``, "", http.StatusBadRequest},
		{"invalid pagination", http.MethodGet, "/api/services?limit=999", ``, "", http.StatusBadRequest},
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
