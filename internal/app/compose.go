package app

import (
	"context"
	"errors"
	"os"

	"github.com/datamaia/andromeda/internal/config"
	"github.com/datamaia/andromeda/internal/pal"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/secret"
)

// keychainBackend adapts a pal.CredentialStore to secret.Backend, translating the platform-neutral
// pal.ErrCredentialNotFound into secret.ErrNotFound so the Secret Store recognizes a missing item
// (e.g. an as-yet-unwritten reference index) as empty rather than as a backend failure.
type keychainBackend struct{ cs pal.CredentialStore }

func (k keychainBackend) Get(service, account string) ([]byte, error) {
	b, err := k.cs.Get(service, account)
	if errors.Is(err, pal.ErrCredentialNotFound) {
		return nil, secret.ErrNotFound
	}
	return b, err
}

func (k keychainBackend) Set(service, account string, v []byte) error {
	return k.cs.Set(service, account, v)
}

func (k keychainBackend) Delete(service, account string) error {
	if err := k.cs.Delete(service, account); err != nil {
		if errors.Is(err, pal.ErrCredentialNotFound) {
			return secret.ErrNotFound
		}
		return err
	}
	return nil
}

func (k keychainBackend) Available() bool { return k.cs.Available() }

// SecretStore builds the Secret Store, preferring the OS keychain and falling back to the
// age-encrypted file store when a keychain is unavailable and a passphrase is provided via
// ANDROMEDA_SECRET_PASSPHRASE. Returns an error when neither backend is usable.
func SecretStore() (ports.SecretStorePort, error) {
	kc := pal.NewCredentialStore()
	if kc.Available() {
		return secret.NewStore(keychainBackend{cs: kc}), nil
	}
	pass := os.Getenv("ANDROMEDA_SECRET_PASSPHRASE")
	if pass == "" {
		return nil, &ports.PortError{
			Code: "E-SEC-022", Category: "security", Severity: "error",
			Message: "no OS keychain available; set ANDROMEDA_SECRET_PASSPHRASE to use the encrypted file fallback",
		}
	}
	dirs := pal.NewConfigDirs()
	dataHome, err := dirs.DataHome()
	if err != nil {
		return nil, err
	}
	fb, err := secret.NewFileBackend(dataHome+"/secrets.age", pass)
	if err != nil {
		return nil, err
	}
	return secret.NewStore(fb), nil
}

// LoadedConfig resolves configuration for the current workspace and returns the effective
// values with source attribution.
func LoadedConfig(ctx context.Context, workspaceRoot string) (ports.ResolvedConfig, error) {
	dirs := pal.NewConfigDirs()
	m, err := config.Load(ctx, dirs, workspaceRoot)
	if err != nil {
		return ports.ResolvedConfig{}, err
	}
	return m.Resolve(ctx, ports.ConfigQuery{Scope: "workspace"})
}
