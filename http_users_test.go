package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func doReq(app *App, method, path, body string, userID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if userID != "" {
		req.Header.Set("X-UserID", userID)
	}
	rr := httptest.NewRecorder()
	app.buildHTTPHandler().ServeHTTP(rr, req)
	return rr
}

func TestUsersAPI_CreateGetUpdateSkills(t *testing.T) {
	app := newTestApp(t)

	rr := doReq(app, http.MethodPost, "/api/users", `{"pseudo":"alice","bio":"hello","ville":"Paris"}`, "")
	if rr.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", rr.Code, rr.Body.String())
	}
	var created userDTO
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ID != 1 || created.Pseudo != "alice" || created.CreditBalance != WelcomeCredits {
		t.Fatalf("bad created user: %+v", created)
	}

	// Le journal doit contenir la transaction welcome.
	sum, err := app.Users.Credits.SumByUser(t.Context(), created.ID)
	if err != nil {
		t.Fatalf("sum credits: %v", err)
	}
	if sum != WelcomeCredits {
		t.Fatalf("credit journal sum=%d, want %d", sum, WelcomeCredits)
	}

	rr = doReq(app, http.MethodGet, "/api/users/1", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodPut, "/api/users/1", `{"pseudo":"alice2","bio":"updated","ville":"Lyon"}`, "1")
	if rr.Code != http.StatusOK {
		t.Fatalf("update status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodPut, "/api/users/1/skills", `{"skills":[{"nom":"Jardinage","niveau":"expert"},{"nom":"Cuisine","niveau":"debutant"}]}`, "1")
	if rr.Code != http.StatusOK {
		t.Fatalf("replace skills status=%d body=%s", rr.Code, rr.Body.String())
	}
	var skills []skillDTO
	if err := json.NewDecoder(rr.Body).Decode(&skills); err != nil {
		t.Fatalf("decode skills: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("skills len=%d, want 2", len(skills))
	}

	rr = doReq(app, http.MethodGet, "/api/users/1/skills", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("get skills status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestUsersAPI_Errors(t *testing.T) {
	app := newTestApp(t)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
		userID string
		want   int
	}{
		{"pseudo vide", http.MethodPost, "/api/users", `{"pseudo":""}`, "", http.StatusBadRequest},
		{"get not found", http.MethodGet, "/api/users/999", ``, "", http.StatusNotFound},
		{"update no auth", http.MethodPut, "/api/users/1", `{"pseudo":"a"}`, "", http.StatusUnauthorized},
		{"skills no auth", http.MethodPut, "/api/users/1/skills", `{"skills":[]}`, "", http.StatusUnauthorized},
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

func TestUsersAPI_DuplicatePseudo(t *testing.T) {
	app := newTestApp(t)
	body := []byte(`{"pseudo":"alice"}`)
	rr := doReq(app, http.MethodPost, "/api/users", string(body), "")
	if rr.Code != http.StatusCreated {
		t.Fatalf("first create status=%d body=%s", rr.Code, rr.Body.String())
	}
	rr = doReq(app, http.MethodPost, "/api/users", string(bytes.TrimSpace(body)), "")
	if rr.Code != http.StatusConflict {
		t.Fatalf("second create status=%d want 409 body=%s", rr.Code, rr.Body.String())
	}
}
