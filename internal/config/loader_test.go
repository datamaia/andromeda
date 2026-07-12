package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// fakeDirs is a pal.ConfigDirs backed by a fixed base directory, for hermetic tests.
type fakeDirs struct{ base string }

func (f fakeDirs) ConfigHome() (string, error) { return filepath.Join(f.base, "config"), nil }
func (f fakeDirs) DataHome() (string, error)   { return filepath.Join(f.base, "data"), nil }
func (f fakeDirs) CacheHome() (string, error)  { return filepath.Join(f.base, "cache"), nil }
func (f fakeDirs) RuntimeDir() (string, error) { return filepath.Join(f.base, "run"), nil }

func TestLoadLayersWithPrecedence(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	root := t.TempDir()

	// Global sets the theme to light; project overrides max_iterations.
	mustWrite(t, filepath.Join(base, "config", ConfigFileName), "[tui.theme]\nmode = \"light\"\n")
	mustWrite(t, filepath.Join(root, ConfigFileName), "[agent.loop]\nmax_iterations = 7\n")

	m, err := Load(ctx, fakeDirs{base: base}, root)
	if err != nil {
		t.Fatal(err)
	}
	res, err := m.Resolve(ctx, ports.ConfigQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Values["tui.theme.mode"] != "light" || res.Sources["tui.theme.mode"] != SourceGlobal {
		t.Errorf("theme mode = %v (%s)", res.Values["tui.theme.mode"], res.Sources["tui.theme.mode"])
	}
	if res.Values["agent.loop.max_iterations"] != int64(7) || res.Sources["agent.loop.max_iterations"] != SourceProject {
		t.Errorf("max_iterations = %v (%s)", res.Values["agent.loop.max_iterations"], res.Sources["agent.loop.max_iterations"])
	}
	// Default that no layer overrode retains its source.
	if res.Sources["logging.level"] != SourceDefaults {
		t.Errorf("logging.level source = %s, want defaults", res.Sources["logging.level"])
	}
}

func TestLoadMissingFilesAreSkipped(t *testing.T) {
	ctx := context.Background()
	m, err := Load(ctx, fakeDirs{base: t.TempDir()}, t.TempDir())
	if err != nil {
		t.Fatalf("missing files must not error: %v", err)
	}
	res, _ := m.Resolve(ctx, ports.ConfigQuery{})
	if res.Values["agent.loop.max_iterations"] != int64(50) {
		t.Errorf("default max_iterations = %v, want 50", res.Values["agent.loop.max_iterations"])
	}
}

func TestLoadMalformedFileErrors(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	mustWrite(t, filepath.Join(base, "config", ConfigFileName), "x = = broken")
	if _, err := Load(ctx, fakeDirs{base: base}, ""); err == nil {
		t.Fatal("expected an error for malformed global config")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
