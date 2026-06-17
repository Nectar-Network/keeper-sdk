package soroban

import (
	"fmt"
	"regexp"
	"strconv"
)

// BroadcastError wraps any failure that occurs once a transaction has (or may
// have) been broadcast to the network — i.e. anything at or after Send. The
// transaction may already be executing or committed, so the caller must NOT
// rebuild and resubmit it with a fresh sequence number: doing so would
// double-execute a non-idempotent operation (e.g. draw vault capital twice or
// fill an auction twice). InvokeWithRetry treats this as non-retryable; callers
// reconcile against on-chain state instead (the keeper's stale-draw recovery
// makes the vault whole on the next cycle).
type BroadcastError struct {
	Hash string
	Err  error
}

func (e *BroadcastError) Error() string {
	if e.Hash != "" {
		return fmt.Sprintf("tx %s broadcast but unconfirmed: %v", shortHash(e.Hash), e.Err)
	}
	return fmt.Sprintf("tx possibly broadcast (not safe to resubmit): %v", e.Err)
}

func (e *BroadcastError) Unwrap() error { return e.Err }

// shortHash returns the first 8 chars of a tx hash for logging, guarding against
// a short/empty hash from a misbehaving RPC (a bare hash[:8] would panic).
func shortHash(h string) string {
	if len(h) < 8 {
		return h
	}
	return h[:8]
}

var (
	// Canonical Soroban contract-error rendering, e.g. "Error(Contract, #4)".
	contractCodeRe = regexp.MustCompile(`Error\(Contract,\s*#(\d+)\)`)
	// Fallback: a bare "#N" code token when the canonical form is absent.
	looseCodeRe = regexp.MustCompile(`#(\d+)\b`)
)

// ParseContractCode extracts the numeric contract error code from a Soroban
// error message (simulate/invoke). It prefers the canonical
// "Error(Contract, #N)" form and falls back to a bare "#N" token, returning
// (code, true) on success. Callers match on the integer code against named
// constants rather than substring-scanning free text — which is both robust
// against incidental matches (a "#42" id, a base64 result blob) and far harder
// for an adversarial RPC to spoof than a variant-name substring.
func ParseContractCode(msg string) (uint32, bool) {
	if m := contractCodeRe.FindStringSubmatch(msg); m != nil {
		if n, err := strconv.ParseUint(m[1], 10, 32); err == nil {
			return uint32(n), true
		}
	}
	if m := looseCodeRe.FindStringSubmatch(msg); m != nil {
		if n, err := strconv.ParseUint(m[1], 10, 32); err == nil {
			return uint32(n), true
		}
	}
	return 0, false
}
