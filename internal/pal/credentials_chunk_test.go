//go:build unix

package pal

import (
	"errors"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

// A secret far larger than the keychain's per-item command limit must round-trip via transparent
// chunking. Regression for E-SEC-021: an OAuth token bundle (several JWTs) exceeded the macOS
// go-keyring 4 KB command cap and failed to store. The in-memory mock backend makes the chunk
// split/reassemble logic deterministic on every platform, including headless CI.
func TestChunkedCredentialRoundTrip(t *testing.T) {
	keyring.MockInit()
	cs := NewCredentialStore()
	if !cs.Available() {
		t.Fatal("mock keyring backend should report available")
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

	// Overwriting a chunked secret with a smaller one exercises the prior-chunk cleanup path, then
	// Delete removes the (now single-item) value.
	if err := cs.Set(svc, acct, []byte("short")); err != nil {
		t.Fatalf("re-Set: %v", err)
	}
	if got, _ := cs.Get(svc, acct); string(got) != "short" {
		t.Fatalf("shrunk secret = %q, want short", got)
	}
	if err := cs.Delete(svc, acct); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := cs.Get(svc, acct); !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("after Delete, Get = %v, want ErrCredentialNotFound", err)
	}
}

// A missing item surfaces as the platform-neutral sentinel, not the backend's raw error, so the
// Secret Store's not-found handling works regardless of OS keychain.
func TestMissingCredentialIsNotFound(t *testing.T) {
	keyring.MockInit()
	cs := NewCredentialStore()
	_, err := cs.Get("andromeda:test.absent", "does-not-exist")
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("missing item = %v, want ErrCredentialNotFound", err)
	}
}

// A small secret stores as a single item (the non-chunked path) and round-trips unchanged.
func TestSmallCredentialRoundTrip(t *testing.T) {
	keyring.MockInit()
	cs := NewCredentialStore()
	const svc, acct = "andromeda:test.small", "key:default"
	if err := cs.Set(svc, acct, []byte("short-secret")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	t.Cleanup(func() { _ = cs.Delete(svc, acct) })
	got, err := cs.Get(svc, acct)
	if err != nil || string(got) != "short-secret" {
		t.Fatalf("round-trip = %q, %v", got, err)
	}
}
