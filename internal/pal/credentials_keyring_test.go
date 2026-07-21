//go:build unix || windows

package pal

import (
	"errors"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"
)

// A secret larger than the Windows Credential Manager's 2560-byte blob cap must round-trip via the
// shared chunking when stored with a small chunk size. Regression for E-SEC-021: the Windows backend
// stored ChatGPT OAuth token bundles unchunked and failed once the base64 value crossed 2560 bytes.
// The literal 1000 mirrors windowsChunkSize (which is only defined on the windows build) so this
// runs on the Linux CI runner and exercises the exact multi-chunk path Windows uses.
func TestKeyringChunkedSmallChunkSize(t *testing.T) {
	keyring.MockInit()
	const smallChunk = 1000
	const svc, acct = "andromeda:test.wincred", "oauth_token:default"
	big := []byte(strings.Repeat("tok3n_", 900)) // ~5.4 KB -> base64 ~7.2 KB -> several 1000-char chunks

	if err := keyringSet(smallChunk, svc, acct, big); err != nil {
		t.Fatalf("keyringSet with the Windows chunk size must chunk transparently: %v", err)
	}
	t.Cleanup(func() { _ = keyringDelete(svc, acct) })

	got, err := keyringGet(svc, acct)
	if err != nil {
		t.Fatalf("keyringGet: %v", err)
	}
	if string(got) != string(big) {
		t.Fatalf("round-trip mismatch: got %d bytes, want %d", len(got), len(big))
	}

	// Overwrite with a value smaller than one chunk to exercise the prior-chunk cleanup path.
	if err := keyringSet(smallChunk, svc, acct, []byte("small")); err != nil {
		t.Fatalf("re-Set: %v", err)
	}
	if got, _ := keyringGet(svc, acct); string(got) != "small" {
		t.Fatalf("shrunk secret = %q, want small", got)
	}
	if err := keyringDelete(svc, acct); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := keyringGet(svc, acct); !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("after Delete, Get = %v, want ErrCredentialNotFound", err)
	}
}
