package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestDebugSeedAndReconcile(t *testing.T) {
	app := newTestApp(t)

	rr := doReq(app, http.MethodPost, "/debug/seed", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("seed status=%d body=%s", rr.Code, rr.Body.String())
	}
	var seed map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&seed); err != nil {
		t.Fatalf("decode seed: %v", err)
	}
	if seed["status"] != "seeded" {
		t.Fatalf("bad seed response: %+v", seed)
	}

	// Idempotence : deuxième appel seed ne doit ni échouer ni doubler les welcome.
	rr = doReq(app, http.MethodPost, "/debug/seed", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("seed idempotent status=%d body=%s", rr.Code, rr.Body.String())
	}

	rr = doReq(app, http.MethodGet, "/debug/reconcile", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("reconcile status=%d body=%s", rr.Code, rr.Body.String())
	}
	var rec reconcileResponse
	if err := json.NewDecoder(rr.Body).Decode(&rec); err != nil {
		t.Fatalf("decode reconcile: %v", err)
	}
	if !rec.OK || len(rec.Incoherences) != 0 {
		t.Fatalf("expected clean reconcile, got %+v", rec)
	}
}

func TestDebugReconcileDetectsMismatch(t *testing.T) {
	app := newTestApp(t)
	uid := createUserWithSkills(t, app, "alice", nil)
	if _, err := app.DB.ExecContext(t.Context(), `UPDATE users SET credit_balance = credit_balance + 99 WHERE id = $1`, uid); err != nil {
		t.Fatalf("corrupt balance: %v", err)
	}

	rr := doReq(app, http.MethodGet, "/debug/reconcile", ``, "")
	if rr.Code != http.StatusOK {
		t.Fatalf("reconcile status=%d body=%s", rr.Code, rr.Body.String())
	}
	var rec reconcileResponse
	_ = json.NewDecoder(rr.Body).Decode(&rec)
	if rec.OK || len(rec.Incoherences) != 1 || rec.Incoherences[0].UserID != uid {
		t.Fatalf("expected one mismatch for uid=%d, got %+v", uid, rec)
	}
}
