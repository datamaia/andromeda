//go:build unix

package pal

// unixCredentialStore implements CredentialStore over the OS keychain via zalando/go-keyring
// (macOS Keychain through /usr/bin/security; Linux Secret Service through D-Bus). Large secrets are
// transparently chunked and reassembled — the shared logic lives in credentials_keyring.go.
type unixCredentialStore struct{}

// NewCredentialStore returns the OS-keychain CredentialStore (ADR-014).
func NewCredentialStore() CredentialStore { return unixCredentialStore{} }

// unixChunkSize is the max base64 characters per keychain item. It is deliberately small because
// go-keyring's macOS backend caps the entire `security add-generic-password … -w <val>` command at
// 4096 bytes AND base64-encodes the value a second time (≈ ×1.34); 2000 leaves over a kilobyte of
// headroom for the service/account names, chunk suffix, and shell quoting.
const unixChunkSize = 2000

func (unixCredentialStore) Get(service, account string) ([]byte, error) {
	return keyringGet(service, account)
}

func (unixCredentialStore) Set(service, account string, secret []byte) error {
	return keyringSet(unixChunkSize, service, account, secret)
}

func (unixCredentialStore) Delete(service, account string) error {
	return keyringDelete(service, account)
}

func (unixCredentialStore) Available() bool { return keyringAvailable() }
