package settingstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/storage"
)

func TestLoadAbsentIsZero(t *testing.T) {
	s, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if s.AutoCompact {
		t.Fatalf("absent settings should be zero-valued, got %+v", s)
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	root := t.TempDir()
	if err := Save(root, Settings{AutoCompact: true}); err != nil {
		t.Fatal(err)
	}
	// The file lands under the .andromeda marker directory.
	if _, err := os.Stat(filepath.Join(root, storage.MarkerDir, FileName)); err != nil {
		t.Fatalf("settings file not written: %v", err)
	}
	got, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if !got.AutoCompact {
		t.Fatalf("roundtrip lost AutoCompact: %+v", got)
	}
	// No leftover temp file.
	if _, err := os.Stat(Path(root) + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temp file should be renamed away")
	}
}

func TestLoadMalformedErrors(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, storage.MarkerDir), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(Path(root), []byte("this = = not toml"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(root); err == nil {
		t.Fatal("malformed settings should error")
	}
}
