//go:build unix || windows

package pal

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/zalando/go-keyring"
)

// This file holds the OS-keychain credential logic shared by the unix backend (macOS Keychain via
// /usr/bin/security; Linux Secret Service via D-Bus) and the Windows backend (Credential Manager),
// all reached through zalando/go-keyring. Byte secrets are base64-encoded because the backends store
// strings. Every backend caps a single item, so large secrets — OAuth token bundles with several
// JWTs — are split into numbered chunks and transparently reassembled. The per-platform cap is
// passed in as chunkSize (see unixChunkSize / windowsChunkSize), because the Windows Credential
// Manager's 2560-byte blob limit is much tighter than the macOS command-line limit.

// chunkHeaderPrefix marks an item whose value is a chunk count rather than base64 data. A ":" never
// appears in base64, so the prefix is unambiguous.
const chunkHeaderPrefix = "chunks:"

func chunkAccount(account string, i int) string { return fmt.Sprintf("%s#chunk%d", account, i) }

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

// mapNotFound normalizes the backend's missing-item error to the platform-neutral sentinel so the
// Secret Store's not-found handling works regardless of which OS keychain reported it.
func mapNotFound(err error) error {
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrCredentialNotFound
	}
	return err
}

// keyringGet reads a (possibly chunked) secret and base64-decodes it.
func keyringGet(service, account string) ([]byte, error) {
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

// keyringSet base64-encodes a secret and stores it, splitting the encoded value into chunkSize-char
// chunks when it exceeds a single item's cap. Any prior chunks are cleaned up first so a shrinking
// secret leaves nothing dangling.
func keyringSet(chunkSize int, service, account string, secret []byte) error {
	enc := base64.StdEncoding.EncodeToString(secret)
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

// keyringDelete removes a secret and any chunks it was split into.
func keyringDelete(service, account string) error {
	if head, err := keyring.Get(service, account); err == nil {
		if n, ok := parseChunkHeader(head); ok {
			for i := 0; i < n; i++ {
				_ = keyring.Delete(service, chunkAccount(account, i))
			}
		}
	}
	return mapNotFound(keyring.Delete(service, account))
}

// keyringAvailable probes the backend: a missing item (ErrNotFound) means it works; any other error
// means it is unavailable (no D-Bus Secret Service, locked keychain, headless CI).
func keyringAvailable() bool {
	_, err := keyring.Get("andromeda.probe", "availability")
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}
