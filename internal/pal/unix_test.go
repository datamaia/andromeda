//go:build unix

package pal

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDirsUnderAndromeda(t *testing.T) {
	cd := NewConfigDirs()
	for name, fn := range map[string]func() (string, error){
		"ConfigHome": cd.ConfigHome,
		"DataHome":   cd.DataHome,
		"CacheHome":  cd.CacheHome,
		"RuntimeDir": cd.RuntimeDir,
	} {
		got, err := fn()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if filepath.Base(got) != "andromeda" {
			t.Errorf("%s = %q, want it to end in /andromeda", name, got)
		}
	}
}

func TestDataHomeHonorsXDG(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdgtest")
	got, err := NewConfigDirs().DataHome()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "/tmp/xdgtest/") {
		t.Errorf("DataHome = %q, want it under the XDG_DATA_HOME override", got)
	}
}

func TestTempFilesAndPaths(t *testing.T) {
	dir, err := NewTempFiles().TempDir("andromeda-test-")
	if err != nil {
		t.Fatal(err)
	}
	abs, err := NewPaths().Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if abs != dir {
		t.Errorf("Abs(%q) = %q, want the same absolute path", dir, abs)
	}
}

func TestPathsHelpers(t *testing.T) {
	p := NewPaths()
	if got := p.Clean("/a/b/../c"); got != "/a/c" {
		t.Errorf("Clean = %q, want /a/c", got)
	}
	if got := p.Join("a", "b", "c"); got != "a/b/c" {
		t.Errorf("Join = %q, want a/b/c", got)
	}
	if _, err := p.Home(); err != nil {
		t.Errorf("Home: %v", err)
	}
}

func TestRuntimeDirHonorsXDG(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "/tmp/xdgrun")
	got, err := NewConfigDirs().RuntimeDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/xdgrun/andromeda" {
		t.Errorf("RuntimeDir = %q, want /tmp/xdgrun/andromeda", got)
	}
}

func TestTempFileCreated(t *testing.T) {
	name, err := NewTempFiles().TempFile("andromeda-f-")
	if err != nil {
		t.Fatal(err)
	}
	if name == "" {
		t.Error("expected a temp file path")
	}
}

func TestFileLockingExclusive(t *testing.T) {
	dir, err := NewTempFiles().TempDir("andromeda-lock-")
	if err != nil {
		t.Fatal(err)
	}
	lockPath := filepath.Join(dir, "lock")
	fl := NewFileLocking()
	l, err := fl.Acquire(context.Background(), lockPath, true)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	// Re-acquire after release must succeed.
	l2, err := fl.Acquire(context.Background(), lockPath, true)
	if err != nil {
		t.Fatalf("re-acquire: %v", err)
	}
	if err := l2.Release(); err != nil {
		t.Fatalf("release 2: %v", err)
	}
}
