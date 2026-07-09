package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSON(t *testing.T) {
	t.Parallel()
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	cases := []struct {
		name    string
		body    string
		wantOK  bool
		wantFld string
	}{
		{"ok", `{"name":"alice","age":30}`, true, ""},
		{"empty", ``, false, ""},
		{"syntax", `{`, false, ""},
		{"unknown field", `{"name":"alice","age":30,"extra":true}`, false, "extra"},
		{"bad type", `{"name":"alice","age":"old"}`, false, "age"},
		{"two objects", `{"name":"a","age":1}{"name":"b","age":2}`, false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
			rr := httptest.NewRecorder()
			var got payload
			err := readJSON(rr, req, &got)
			if tc.wantOK {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				if got.Name != "alice" || got.Age != 30 {
					t.Fatalf("decoded=%+v", got)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, ErrBadRequest) {
				t.Fatalf("expected ErrBadRequest, got %v", err)
			}
			if tc.wantFld != "" {
				var de *DomainError
				if !errors.As(err, &de) || de.Field != tc.wantFld {
					t.Fatalf("expected field %q, got err=%v", tc.wantFld, err)
				}
			}
		})
	}
}

func TestParseIDParamAndWriteJSON(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]int{"id": id})
	})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/items/42", nil))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("code=%d", rr.Code)
	}
	var got map[string]int
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil || got["id"] != 42 {
		t.Fatalf("decode=%v got=%v", err, got)
	}

	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/items/nope", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("bad id code=%d", rr.Code)
	}
}

func TestParsePageParams(t *testing.T) {
	t.Parallel()
	cases := []struct {
		url     string
		want    pageParams
		wantErr bool
	}{
		{"/x", pageParams{Limit: 20, Offset: 0}, false},
		{"/x?limit=50&offset=10", pageParams{Limit: 50, Offset: 10}, false},
		{"/x?limit=0", pageParams{}, true},
		{"/x?limit=101", pageParams{}, true},
		{"/x?limit=abc", pageParams{}, true},
		{"/x?offset=-1", pageParams{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.url, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			got, err := parsePageParams(req)
			if tc.wantErr {
				if !errors.Is(err, ErrBadRequest) {
					t.Fatalf("expected BadRequest, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got=%+v, want=%+v", got, tc.want)
			}
		})
	}
}

func TestRequireUserID(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if _, err := requireUserID(req); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized without ctx user, got %v", err)
	}

	req.Header.Set("X-UserID", "42")
	rr := httptest.NewRecorder()
	authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := requireUserID(r)
		if err != nil || got != 42 {
			t.Fatalf("uid=%d err=%v", got, err)
		}
	})).ServeHTTP(rr, req)
}
