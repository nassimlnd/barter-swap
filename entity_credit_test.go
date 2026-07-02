package main

import "testing"

func TestCreditTxType_Valid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		t    CreditTxType
		want bool
	}{
		{CreditWelcome, true},
		{CreditSpend, true},
		{CreditEarn, true},
		{CreditRefund, true},
		{"unknown", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := tc.t.Valid(); got != tc.want {
			t.Errorf("(%q).Valid()=%v, want %v", tc.t, got, tc.want)
		}
	}
}
