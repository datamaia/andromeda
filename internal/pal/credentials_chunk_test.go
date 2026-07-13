//go:build unix

package pal

import (
	"errors"
	"strings"
	"testing"
)

// A secret far larger than the keychain's per-item command limit must round-trip via transparent
// chunking. Regression for E-SEC-021: an OAuth token bundle (several JWTs) exceeded the macOS
// go-keyring 4 KB command cap and failed to store. Skips where no OS keychain is present (CI).
func TestChunkedCredentialRoundTrip(t *testing.T) {
	cs := NewCredentialStore()
	if !cs.Available() {
		t.Skip("no OS keychain available")
	}
	const svc, acct = "andromeda:test.chunk", "oauth_token:default"
	big := []byte(strings.Repeat("Xy9_", 5000)) // 20 KB → ~27 KB base64 → many chunks

	if err := cs.Set(svc, acct, big); err != nil {
		t.Fatalf("Set of a large secret must chunk transparently: %v", err)
	}
	t.Cleanup(func() { _ = cs.Delete(svc, acct) })

	got, err := cs.Get(svc, acct)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != string(big) {
		t.Fatalf("round-trip mismatch: got %d bytes, want %d", len(got), len(big))
	}
}

// A missing item surfaces as the platform-neutral sentinel, not the backend's raw error, so the
// Secret Store's not-found handling works regardless of OS keychain.
func TestMissingCredentialIsNotFound(t *testing.T) {
	cs := NewCredentialStore()
	if !cs.Available() {
		t.Skip("no OS keychain available")
	}
	_, err := cs.Get("andromeda:test.absent", "does-not-exist")
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("missing item = %v, want ErrCredentialNotFound", err)
	}
}
