package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestIDMW(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", "known")
	rr := httptest.NewRecorder()
	requestIDMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Context().Value(ctxRequestID); got != "known" {
			t.Fatalf("ctx request id=%v", got)
		}
	})).ServeHTTP(rr, req)
	if got := rr.Header().Get("X-Request-ID"); got != "known" {
		t.Fatalf("header request id=%q", got)
	}
}

func TestCorsMWOptions(t *testing.T) {
	t.Parallel()
	rr := httptest.NewRecorder()
	called := false
	corsMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })).ServeHTTP(rr, httptest.NewRequest(http.MethodOptions, "/", nil))
	if called {
		t.Fatal("next should not be called for OPTIONS")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("code=%d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Access-Control-Allow-Headers"), "X-UserID") {
		t.Fatalf("missing cors headers")
	}
}

func TestRecoverMW(t *testing.T) {
	t.Parallel()
	app := &App{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Now: time.Now}
	rr := httptest.NewRecorder()
	app.recoverMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { panic("boom") })).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("code=%d", rr.Code)
	}
}

func TestLoggingMWAndStatusRecorder(t *testing.T) {
	t.Parallel()
	app := &App{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Now: time.Now}
	rr := httptest.NewRecorder()
	h := app.loggingMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/x", nil))
	if rr.Code != http.StatusCreated {
		t.Fatalf("code=%d", rr.Code)
	}
}
