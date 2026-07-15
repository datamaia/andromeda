package permstore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/storage"
)

func TestAddLoadRoundtrip(t *testing.T) {
	root := t.TempDir()
	if _, err := Add(root, Allow, "git status"); err != nil {
		t.Fatal(err)
	}
	if _, err := Add(root, Allow, "go test ./..."); err != nil {
		t.Fatal(err)
	}
	if _, err := Add(root, Deny, "git push --force"); err != nil {
		t.Fatal(err)
	}
	// The file must live under the marker dir and carry both lists.
	if _, err := os.Stat(filepath.Join(root, storage.MarkerDir, FileName)); err != nil {
		t.Fatalf("permissions file not written: %v", err)
	}
	r, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Allow) != 2 || len(r.Deny) != 1 {
		t.Fatalf("rules = %+v", r)
	}
	if r.Deny[0] != "git push --force" {
		t.Fatalf("deny = %v", r.Deny)
	}
}

func TestAddDedupsAndRemove(t *testing.T) {
	root := t.TempDir()
	_, _ = Add(root, Allow, "git status")
	_, _ = Add(root, Allow, "git status") // duplicate
	r, _ := Load(root)
	if len(r.Allow) != 1 {
		t.Fatalf("duplicate not collapsed: %v", r.Allow)
	}
	if _, err := Remove(root, Allow, "git status"); err != nil {
		t.Fatal(err)
	}
	r, _ = Load(root)
	if len(r.Allow) != 0 {
		t.Fatalf("remove failed: %v", r.Allow)
	}
}

func TestLoadAbsentIsEmpty(t *testing.T) {
	r, err := Load(t.TempDir())
	if err != nil {
		t.Fatalf("absent file should not error: %v", err)
	}
	if len(r.Allow) != 0 || len(r.Deny) != 0 {
		t.Fatalf("expected empty rules, got %+v", r)
	}
}

func TestBlankCommandIsNoop(t *testing.T) {
	root := t.TempDir()
	if _, err := Add(root, Allow, "   "); err != nil {
		t.Fatal(err)
	}
	r, _ := Load(root)
	if len(r.Allow) != 0 {
		t.Fatalf("blank command should be ignored, got %v", r.Allow)
	}
}
