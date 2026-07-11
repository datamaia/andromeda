//go:build unix

package pal

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
)

// This file provides the Unix reference implementations of the platform surfaces that the
// architecture skeleton needs (ADR-030: Unix is the reference behavior). The remaining
// surfaces are implemented by their owning epics. A future Windows backend implements the
// same interfaces in a windows-tagged file.

// unixPaths implements Paths.
type unixPaths struct{}

// NewPaths returns the Unix Paths surface.
func NewPaths() Paths { return unixPaths{} }

func (unixPaths) Abs(p string) (string, error) { return filepath.Abs(p) }
func (unixPaths) Clean(p string) string        { return filepath.Clean(p) }
func (unixPaths) Join(e ...string) string      { return filepath.Join(e...) }
func (unixPaths) Home() (string, error)        { return os.UserHomeDir() }

// unixConfigDirs implements ConfigDirs using the Go stdlib resolution, which follows the
// Apple-native mapping on macOS and XDG on Linux (ADR-022). adrg/xdg with explicit XDG_*
// honoring is adopted in EP-03.
type unixConfigDirs struct{}

// NewConfigDirs returns the Unix ConfigDirs surface.
func NewConfigDirs() ConfigDirs { return unixConfigDirs{} }

func (unixConfigDirs) ConfigHome() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "andromeda"), nil
}

func (unixConfigDirs) DataHome() (string, error) {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return filepath.Join(x, "andromeda"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "andromeda"), nil
}

func (unixConfigDirs) CacheHome() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "andromeda"), nil
}

func (unixConfigDirs) RuntimeDir() (string, error) {
	if x := os.Getenv("XDG_RUNTIME_DIR"); x != "" {
		return filepath.Join(x, "andromeda"), nil
	}
	return filepath.Join(os.TempDir(), "andromeda"), nil
}

// unixTempFiles implements TempFiles.
type unixTempFiles struct{}

// NewTempFiles returns the Unix TempFiles surface.
func NewTempFiles() TempFiles { return unixTempFiles{} }

func (unixTempFiles) TempDir(pattern string) (string, error) {
	return os.MkdirTemp("", pattern)
}

func (unixTempFiles) TempFile(pattern string) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	name := f.Name()
	return name, f.Close()
}

// unixFileLocking implements FileLocking with advisory flock(2).
type unixFileLocking struct{}

// NewFileLocking returns the Unix FileLocking surface.
func NewFileLocking() FileLocking { return unixFileLocking{} }

func (unixFileLocking) Acquire(ctx context.Context, path string, exclusive bool) (FileLock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	how := syscall.LOCK_SH
	if exclusive {
		how = syscall.LOCK_EX
	}
	if err := syscall.Flock(int(f.Fd()), how); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &unixFileLock{f: f}, nil
}

type unixFileLock struct{ f *os.File }

func (l *unixFileLock) Release() error {
	err := syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
	if cerr := l.f.Close(); err == nil {
		err = cerr
	}
	return err
}
