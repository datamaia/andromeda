package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestNewPKCEProducesValidS256Pair(t *testing.T) {
	p, err := NewPKCE()
	if err != nil {
		t.Fatal(err)
	}
	if p.Method != "S256" {
		t.Errorf("method = %q, want S256", p.Method)
	}
	// RFC 7636: verifier length 43..128 chars for a 32-byte base64url value.
	if len(p.Verifier) < 43 || len(p.Verifier) > 128 {
		t.Errorf("verifier length %d out of range", len(p.Verifier))
	}
	// The challenge must be base64url(SHA256(verifier)).
	sum := sha256.Sum256([]byte(p.Verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	if p.Challenge != want {
		t.Errorf("challenge = %q, want %q", p.Challenge, want)
	}
}

func TestPKCEAndStateAreUnique(t *testing.T) {
	a, _ := NewPKCE()
	b, _ := NewPKCE()
	if a.Verifier == b.Verifier {
		t.Error("verifiers should be unique across calls")
	}
	s1, _ := RandomState()
	s2, _ := RandomState()
	if s1 == "" || s1 == s2 {
		t.Errorf("states should be non-empty and unique: %q %q", s1, s2)
	}
}
