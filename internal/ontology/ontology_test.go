package ontology

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile is a test helper that creates a file (and parents) under root.
func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// sampleTree builds a small workspace and returns its root.
func sampleTree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, root, "README.md", "# My Project\n\nHello.\n")
	writeFile(t, root, "config.toml", "[x]\n")
	writeFile(t, root, "internal/util/util.go", "// doc\npackage util\n")
	writeFile(t, root, "assets/logo.png", "\x89PNG binary")
	writeFile(t, root, "docs/guide.md", "intro line\n# Guide\n")
	return root
}

func TestScanStructuralFacts(t *testing.T) {
	root := sampleTree(t)
	m, err := Scan(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) != 6 {
		t.Fatalf("file count = %d, want 6: %+v", len(m.Files), m.Files)
	}
	byPath := map[string]FileNode{}
	for _, f := range m.Files {
		byPath[f.Path] = f
	}
	if g := byPath["main.go"]; g.Language != "Go" || g.Kind != "code" || g.Summary != "package main" {
		t.Errorf("main.go = %+v", g)
	}
	if g := byPath["README.md"]; g.Language != "Markdown" || g.Kind != "doc" || g.Summary != "My Project" {
		t.Errorf("README.md = %+v", g)
	}
	if g := byPath["docs/guide.md"]; g.Summary != "Guide" || g.Dir != "docs" {
		t.Errorf("docs/guide.md = %+v", g)
	}
	if g := byPath["assets/logo.png"]; g.Kind != "asset" || g.Language != "Image" {
		t.Errorf("logo.png = %+v", g)
	}
	if g := byPath["internal/util/util.go"]; g.Summary != "package util" || g.Dir != "internal/util" {
		t.Errorf("util.go = %+v", g)
	}
	// Directories are the deduped, sorted set of ancestors.
	wantDirs := []string{"assets", "docs", "internal", "internal/util"}
	if strings.Join(m.Dirs, ",") != strings.Join(wantDirs, ",") {
		t.Errorf("dirs = %v, want %v", m.Dirs, wantDirs)
	}
	if m.Languages["Go"] != 2 || m.Languages["Markdown"] != 2 {
		t.Errorf("languages = %v", m.Languages)
	}
}

func TestTTLIsDeterministicAndWellFormed(t *testing.T) {
	root := sampleTree(t)
	m1, _ := Scan(context.Background(), root)
	m2, _ := Scan(context.Background(), root)
	ttl1, ttl2 := m1.TTL(), m2.TTL()
	if ttl1 != ttl2 {
		t.Fatal("TTL is not deterministic across scans")
	}
	// No timestamps or absolute paths leak in.
	if strings.Contains(ttl1, root) {
		t.Error("TTL leaks the absolute workspace path")
	}
	for _, want := range []string{
		"@prefix am: <https://andromedacli.com/ns#> .",
		"<project> a am:Project ;",
		"am:usesLanguage \"Go\"",
		"<f/main.go> a am:File ;",
		"am:language \"Go\"",
		"am:summary \"package main\"",
		"<d/internal/util> a am:Directory ;",
		"am:inDirectory <d/internal>",
		"am:contains",
	} {
		if !strings.Contains(ttl1, want) {
			t.Errorf("TTL missing %q", want)
		}
	}
	// Every file node points its containment at a real parent (project or a directory node).
	if strings.Count(ttl1, "a am:File ;") != 6 {
		t.Errorf("expected 6 file nodes, got %d", strings.Count(ttl1, "a am:File ;"))
	}
}

func TestIRIAndStringEscaping(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a file.md", "# Title \"quoted\"\n")
	m, _ := Scan(context.Background(), root)
	ttl := m.TTL()
	if !strings.Contains(ttl, "<f/a%20file.md>") {
		t.Errorf("space in path not IRI-escaped: %s", ttl)
	}
	if !strings.Contains(ttl, `am:summary "Title \"quoted\""`) {
		t.Errorf("quote in literal not escaped: %s", ttl)
	}
}

func TestWriteAndRemove(t *testing.T) {
	root := sampleTree(t)
	m, _ := Scan(context.Background(), root)
	ttlPath, err := Write(root, m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ttlPath); err != nil {
		t.Fatalf("ttl not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(Dir(root), "manifest.json")); err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
	if err := Remove(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(Dir(root)); !os.IsNotExist(err) {
		t.Error("ontology dir should be gone after Remove")
	}
	// Remove is idempotent.
	if err := Remove(root); err != nil {
		t.Errorf("second Remove errored: %v", err)
	}
}

func TestScanSkipsAndromedaSurface(t *testing.T) {
	root := sampleTree(t)
	writeFile(t, root, ".andromeda/ontology/project.ttl", "# old")
	writeFile(t, root, ".andromeda/memory/note.md", "x")
	m, _ := Scan(context.Background(), root)
	for _, f := range m.Files {
		if strings.HasPrefix(f.Path, ".andromeda/") {
			t.Errorf("scan should skip .andromeda, found %s", f.Path)
		}
	}
}
