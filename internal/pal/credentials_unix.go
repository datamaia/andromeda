//go:build unix

package pal

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zalando/go-keyring"
)

// unixCredentialStore implements CredentialStore over the OS keychain via zalando/go-keyring
// (macOS Keychain through /usr/bin/security; Linux Secret Service through D-Bus). Byte secrets
// are base64-encoded because the backend stores strings. Because go-keyring caps a single macOS
// item at ~4 KB of command line, large secrets (OAuth token bundles with several JWTs) are split
// into numbered chunks and transparently reassembled.
type unixCredentialStore struct{}

// NewCredentialStore returns the OS-keychain CredentialStore (ADR-014).
func NewCredentialStore() CredentialStore { return unixCredentialStore{} }

// chunkSize is the max base64 characters stored per keychain item. It is deliberately small
// because go-keyring's macOS backend caps the *entire* `security add-generic-password … -w <val>`
// command at 4096 bytes AND base64-encodes the value a second time itself (≈ ×1.34). So one chunk
// costs ~18 + chunkSize×1.34 command bytes; at 2000 that is ~2700, leaving well over a kilobyte of
// headroom for the service name, account name, chunk suffix, and shell quoting — enough that even
// long provider/profile references never push a single item over the limit.
const chunkSize = 2000

// chunkHeaderPrefix marks an item whose value is a chunk count rather than base64 data. A ":" never
// appears in base64, so the prefix is unambiguous.
const chunkHeaderPrefix = "chunks:"

func chunkAccount(account string, i int) string { return fmt.Sprintf("%s#chunk%d", account, i) }

func (unixCredentialStore) Get(service, account string) ([]byte, error) {
	head, err := keyring.Get(service, account)
	if err != nil {
		return nil, mapNotFound(err)
	}
	if n, ok := parseChunkHeader(head); ok {
		var b strings.Builder
		for i := 0; i < n; i++ {
			part, err := keyring.Get(service, chunkAccount(account, i))
			if err != nil {
				return nil, mapNotFound(err)
			}
			b.WriteString(part)
		}
		return base64.StdEncoding.DecodeString(b.String())
	}
	return base64.StdEncoding.DecodeString(head)
}

// mapNotFound normalizes the backend's missing-item error to the platform-neutral sentinel so the
// Secret Store's not-found handling works regardless of which OS keychain reported it.
func mapNotFound(err error) error {
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrCredentialNotFound
	}
	return err
}

func (unixCredentialStore) Set(service, account string, secret []byte) error {
	enc := base64.StdEncoding.EncodeToString(secret)
	// Best-effort cleanup of any prior chunks so a shrinking secret leaves nothing dangling.
	if head, err := keyring.Get(service, account); err == nil {
		if n, ok := parseChunkHeader(head); ok {
			for i := 0; i < n; i++ {
				_ = keyring.Delete(service, chunkAccount(account, i))
			}
		}
	}
	if len(enc) <= chunkSize {
		return keyring.Set(service, account, enc)
	}
	n := (len(enc) + chunkSize - 1) / chunkSize
	for i := 0; i < n; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(enc) {
			end = len(enc)
		}
		if err := keyring.Set(service, chunkAccount(account, i), enc[start:end]); err != nil {
			return err
		}
	}
	return keyring.Set(service, account, fmt.Sprintf("%s%d", chunkHeaderPrefix, n))
}

func (unixCredentialStore) Delete(service, account string) error {
	if head, err := keyring.Get(service, account); err == nil {
		if n, ok := parseChunkHeader(head); ok {
			for i := 0; i < n; i++ {
				_ = keyring.Delete(service, chunkAccount(account, i))
			}
		}
	}
	return mapNotFound(keyring.Delete(service, account))
}

// Available probes the backend: a missing item (ErrNotFound) means the backend works; any
// other error means it is unavailable (no D-Bus Secret Service, locked keychain, headless CI).
func (unixCredentialStore) Available() bool {
	_, err := keyring.Get("andromeda.probe", "availability")
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}

// parseChunkHeader returns the chunk count when value is a "chunks:N" header.
func parseChunkHeader(value string) (int, bool) {
	if !strings.HasPrefix(value, chunkHeaderPrefix) {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(value, chunkHeaderPrefix))
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
