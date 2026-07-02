package main

import "testing"

func TestExchangeStatus_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s    ExchangeStatus
		want bool
	}{
		{StatusPending, true},
		{StatusAccepted, true},
		{StatusRejected, true},
		{StatusCancelled, true},
		{StatusCompleted, true},
		{"unknown", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := tc.s.Valid(); got != tc.want {
			t.Errorf("(%q).Valid()=%v, want %v", tc.s, got, tc.want)
		}
	}
}

func TestExchangeStatus_IsActive_IsTerminal(t *testing.T) {
	t.Parallel()
	cases := []struct {
		s        ExchangeStatus
		active   bool
		terminal bool
	}{
		{StatusPending, true, false},
		{StatusAccepted, true, false},
		{StatusRejected, false, true},
		{StatusCancelled, false, true},
		{StatusCompleted, false, true},
	}
	for _, tc := range cases {
		t.Run(string(tc.s), func(t *testing.T) {
			if got := tc.s.IsActive(); got != tc.active {
				t.Errorf("IsActive : got %v, want %v", got, tc.active)
			}
			if got := tc.s.IsTerminal(); got != tc.terminal {
				t.Errorf("IsTerminal : got %v, want %v", got, tc.terminal)
			}
			// Mutuellement exclusifs
			if tc.active && tc.terminal {
				t.Errorf("statut ne peut être à la fois actif ET terminal")
			}
		})
	}
}

func TestExchange_Transitions(t *testing.T) {
	t.Parallel()
	cases := []struct {
		status      ExchangeStatus
		canAccept   bool
		canReject   bool
		canComplete bool
		canCancel   bool
	}{
		{StatusPending, true, true, false, true},
		{StatusAccepted, false, false, true, true},
		{StatusRejected, false, false, false, false},
		{StatusCancelled, false, false, false, false},
		{StatusCompleted, false, false, false, false},
	}
	for _, tc := range cases {
		t.Run(string(tc.status), func(t *testing.T) {
			e := &Exchange{Status: tc.status}
			if got := e.CanAccept(); got != tc.canAccept {
				t.Errorf("CanAccept : got %v, want %v", got, tc.canAccept)
			}
			if got := e.CanReject(); got != tc.canReject {
				t.Errorf("CanReject : got %v, want %v", got, tc.canReject)
			}
			if got := e.CanComplete(); got != tc.canComplete {
				t.Errorf("CanComplete : got %v, want %v", got, tc.canComplete)
			}
			if got := e.CanCancel(); got != tc.canCancel {
				t.Errorf("CanCancel : got %v, want %v", got, tc.canCancel)
			}
		})
	}
}

func TestExchange_InvolvesUser(t *testing.T) {
	t.Parallel()
	e := &Exchange{RequesterID: 1, OwnerID: 2}
	cases := []struct {
		userID int
		want   bool
	}{
		{1, true},
		{2, true},
		{3, false},
		{0, false},
	}
	for _, tc := range cases {
		if got := e.InvolvesUser(tc.userID); got != tc.want {
			t.Errorf("InvolvesUser(%d)=%v, want %v", tc.userID, got, tc.want)
		}
	}
}
