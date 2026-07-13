package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// PKCE holds a Proof Key for Code Exchange pair (RFC 7636) for the OAuth authorization-code flow
// used by browser logins (ADR-063). The verifier is kept locally and sent only on token exchange;
// only the S256 challenge is placed in the authorization URL.
type PKCE struct {
	Verifier  string
	Challenge string
	Method    string // always "S256"
}

// NewPKCE generates a fresh S256 PKCE pair from a cryptographically random verifier.
func NewPKCE() (PKCE, error) {
	verifier, err := randomURLSafe(32)
	if err != nil {
		return PKCE{}, err
	}
	sum := sha256.Sum256([]byte(verifier))
	return PKCE{
		Verifier:  verifier,
		Challenge: base64.RawURLEncoding.EncodeToString(sum[:]),
		Method:    "S256",
	}, nil
}

// RandomState returns a URL-safe random value for the OAuth state parameter (CSRF protection).
func RandomState() (string, error) { return randomURLSafe(24) }

// randomURLSafe returns n random bytes encoded as unpadded base64url.
func randomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
