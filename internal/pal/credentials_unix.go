//go:build unix

package pal

import (
	"encoding/base64"
	"errors"

	"github.com/zalando/go-keyring"
)

// unixCredentialStore implements CredentialStore over the OS keychain via zalando/go-keyring
// (macOS Keychain through /usr/bin/security; Linux Secret Service through D-Bus). Byte secrets
// are base64-encoded because the backend stores strings.
type unixCredentialStore struct{}

// NewCredentialStore returns the OS-keychain CredentialStore (ADR-014).
func NewCredentialStore() CredentialStore { return unixCredentialStore{} }

func (unixCredentialStore) Get(service, account string) ([]byte, error) {
	enc, err := keyring.Get(service, account)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(enc)
}

func (unixCredentialStore) Set(service, account string, secret []byte) error {
	return keyring.Set(service, account, base64.StdEncoding.EncodeToString(secret))
}

func (unixCredentialStore) Delete(service, account string) error {
	return keyring.Delete(service, account)
}

// Available probes the backend: a missing item (ErrNotFound) means the backend works; any
// other error means it is unavailable (no D-Bus Secret Service, locked keychain, headless CI).
func (unixCredentialStore) Available() bool {
	_, err := keyring.Get("andromeda.probe", "availability")
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}
