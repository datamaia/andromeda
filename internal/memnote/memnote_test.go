package memnote

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddListGet(t *testing.T) {
	root := t.TempDir()

	a, err := Add(root, "Fix the auth bug", []string{"auth", "bug"}, "The token was truncated.")
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != "0001" {
		t.Errorf("first id = %q, want 0001", a.ID)
	}
	b, err := Add(root, "Release process", []string{"release"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != "0002" {
		t.Errorf("second id = %q, want 0002", b.ID)
	}

	notes, err := List(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 || notes[0].ID != "0002" {
		t.Fatalf("list should be newest-first, got %+v", notes)
	}

	got, err := Get(root, "1") // accepts unpadded id
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Fix the auth bug" || strings.Join(got.Tags, ",") != "auth,bug" || got.Body != "The token was truncated." {
		t.Errorf("roundtrip = %+v", got)
	}

	// The file lives under .andromeda/memory with a readable slug, and the index exists.
	if !strings.HasPrefix(got.Slug(), "0001-fix-the-auth-bug") {
		t.Errorf("slug = %q", got.Slug())
	}
	if _, err := os.Stat(filepath.Join(Dir(root), "MEMORY.md")); err != nil {
		t.Errorf("index not written: %v", err)
	}
	idx, _ := os.ReadFile(filepath.Join(Dir(root), "MEMORY.md"))
	if !strings.Contains(string(idx), "Fix the auth bug") || !strings.Contains(string(idx), "Release process") {
		t.Errorf("index missing entries:\n%s", idx)
	}
}

func TestSearchUpdateDelete(t *testing.T) {
	root := t.TempDir()
	_, _ = Add(root, "Auth token chunking", []string{"auth", "keychain"}, "chunk large secrets")
	_, _ = Add(root, "TUI banner", []string{"tui"}, "ANSI shadow wordmark")

	hits, err := Search(root, "keychain")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Title != "Auth token chunking" {
		t.Fatalf("search by tag = %+v", hits)
	}
	if hits, _ := Search(root, "wordmark"); len(hits) != 1 { // matches body
		t.Errorf("search by body failed: %+v", hits)
	}
	if all, _ := Search(root, ""); len(all) != 2 {
		t.Errorf("empty query should return all, got %d", len(all))
	}

	if err := Update(root, "1", "chunk large secrets across N keychain entries"); err != nil {
		t.Fatal(err)
	}
	if n, _ := Get(root, "1"); !strings.Contains(n.Body, "across N keychain") {
		t.Errorf("update did not persist: %q", n.Body)
	}

	if err := Delete(root, "2"); err != nil {
		t.Fatal(err)
	}
	if notes, _ := List(root); len(notes) != 1 {
		t.Errorf("delete failed, %d notes remain", len(notes))
	}
	// Delete is idempotent.
	if err := Delete(root, "2"); err != nil {
		t.Errorf("second delete errored: %v", err)
	}
}

func TestListMissingDirIsEmpty(t *testing.T) {
	notes, err := List(t.TempDir())
	if err != nil || notes != nil {
		t.Errorf("missing dir should yield no notes, got %v err=%v", notes, err)
	}
}
