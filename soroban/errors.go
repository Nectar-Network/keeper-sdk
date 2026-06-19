package soroban

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// rpcErrorMessage extracts a human-readable error from the JSON-RPC "error"
// field, which nodes return inconsistently: an {code,message} object, a bare
// string, an empty string, null, or absent. Returns "" when there is no real
// error so a successful response with `"error":""` is not treated as a failure.
func rpcErrorMessage(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "null" || s == `""` || s == "{}" {
		return ""
	}
	var obj struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &obj) == nil && (obj.Message != "" || obj.Code != 0) {
		if obj.Message == "" {
			return fmt.Sprintf("code %d", obj.Code)
		}
		if obj.Code != 0 {
			return fmt.Sprintf("%s (code %d)", obj.Message, obj.Code)
		}
		return obj.Message
	}
	var str string
	if json.Unmarshal(raw, &str) == nil {
		if strings.TrimSpace(str) == "" {
			return ""
		}
		return str
	}
	return s
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
