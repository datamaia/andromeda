//go:build windows

package pal

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/zalando/go-keyring"
)

// This file provides the Windows backends of the platform surfaces (Volume 3 chapter 07,
// Windows-future encapsulation). Windows is a v2-phase target; these implementations let the
// codebase cross-compile and run on Windows without scattering platform checks elsewhere.

// --- Paths ---

type windowsPaths struct{}

// NewPaths returns the Windows Paths surface.
func NewPaths() Paths { return windowsPaths{} }

func (windowsPaths) Abs(p string) (string, error) { return filepath.Abs(p) }
func (windowsPaths) Clean(p string) string        { return filepath.Clean(p) }
func (windowsPaths) Join(e ...string) string      { return filepath.Join(e...) }
func (windowsPaths) Home() (string, error)        { return os.UserHomeDir() }

// --- ConfigDirs ---

type windowsConfigDirs struct{}

// NewConfigDirs returns the Windows ConfigDirs surface, resolving to the standard known folders
// (%AppData% for config, %LocalAppData% for data/cache) with an andromeda subdirectory.
func NewConfigDirs() ConfigDirs { return windowsConfigDirs{} }

func (windowsConfigDirs) ConfigHome() (string, error) {
	base, err := os.UserConfigDir() // %AppData%
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "andromeda"), nil
}

func (windowsConfigDirs) DataHome() (string, error) {
	if x := os.Getenv("LocalAppData"); x != "" {
		return filepath.Join(x, "andromeda"), nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "andromeda"), nil
}

func (windowsConfigDirs) CacheHome() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "andromeda", "cache"), nil
}

func (windowsConfigDirs) RuntimeDir() (string, error) {
	return filepath.Join(os.TempDir(), "andromeda"), nil
}

// --- TempFiles ---

type windowsTempFiles struct{}

// NewTempFiles returns the Windows TempFiles surface.
func NewTempFiles() TempFiles { return windowsTempFiles{} }

func (windowsTempFiles) TempDir(pattern string) (string, error) { return os.MkdirTemp("", pattern) }
func (windowsTempFiles) TempFile(pattern string) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	name := f.Name()
	return name, f.Close()
}

// --- FileLocking ---

type windowsFileLocking struct{}

// NewFileLocking returns the Windows FileLocking surface. It uses an exclusive-create lock file
// with bounded retry (Windows lacks flock); the lock file is removed on release.
func NewFileLocking() FileLocking { return windowsFileLocking{} }

func (windowsFileLocking) Acquire(ctx context.Context, path string, _ bool) (FileLock, error) {
	lockPath := path + ".lock"
	deadline := time.Now().Add(10 * time.Second)
	for {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
		if err == nil {
			_ = f.Close()
			return &windowsFileLock{path: lockPath}, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, errors.New("pal: timed out acquiring file lock")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

type windowsFileLock struct{ path string }

func (l *windowsFileLock) Release() error { return os.Remove(l.path) }

// --- CredentialStore ---

type windowsCredentialStore struct{}

// NewCredentialStore returns the Windows CredentialStore surface, backed by the Windows
// Credential Manager via zalando/go-keyring.
func NewCredentialStore() CredentialStore { return windowsCredentialStore{} }

func (windowsCredentialStore) Get(service, account string) ([]byte, error) {
	enc, err := keyring.Get(service, account)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(enc)
}

func (windowsCredentialStore) Set(service, account string, secret []byte) error {
	return keyring.Set(service, account, base64.StdEncoding.EncodeToString(secret))
}

func (windowsCredentialStore) Delete(service, account string) error {
	return keyring.Delete(service, account)
}

func (windowsCredentialStore) Available() bool {
	_, err := keyring.Get("andromeda.probe", "availability")
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}
