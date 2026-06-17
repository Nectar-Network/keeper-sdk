package soroban

import (
	"errors"
	"fmt"
	"testing"
)

func TestParseContractCode(t *testing.T) {
	cases := []struct {
		in   string
		want uint32
		ok   bool
	}{
		{"HostError: Error(Contract, #4)", 4, true},
		{"submit sim: Error(Contract, #10) trapped", 10, true},
		{"Error(Contract, #5)", 5, true},
		{"value #7 standalone", 7, true},
		{"no contract code present", 0, false},
		{"tx 1a2b3c4d failed: AAAABQ== resultXdr", 0, false}, // base64/hash, no #N token
	}
	for _, c := range cases {
		got, ok := ParseContractCode(c.in)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("ParseContractCode(%q) = (%d,%v), want (%d,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestBroadcastErrorNotRetryable(t *testing.T) {
	be := &BroadcastError{Hash: "abcdef0123456789", Err: errors.New("timed out")}
	if isRetryable(be) {
		t.Error("a BroadcastError must never be retryable (double-execution hazard)")
	}
	if isRetryable(fmt.Errorf("await: %w", be)) {
		t.Error("a wrapped BroadcastError must never be retryable")
	}
	// A plain transient error that did NOT broadcast stays retryable.
	if !isRetryable(errors.New("connection timed out")) {
		t.Error("plain timeout (pre-broadcast) should remain retryable")
	}
}

func TestShortHash(t *testing.T) {
	if got := shortHash("abcdef0123456789"); got != "abcdef01" {
		t.Errorf("shortHash long = %q, want abcdef01", got)
	}
	if got := shortHash("abc"); got != "abc" { // must not panic on short input
		t.Errorf("shortHash short = %q, want abc", got)
	}
	if got := shortHash(""); got != "" {
		t.Errorf("shortHash empty = %q, want empty", got)
	}
}
