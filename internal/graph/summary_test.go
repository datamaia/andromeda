package graph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileSummary(t *testing.T) {
	root := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(root, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	if got := fileSummary(write("a.md", "---\nname: x\ndescription: The A doc\n---\nbody")); got != "The A doc" {
		t.Fatalf("frontmatter description: %q", got)
	}
	if got := fileSummary(write("b.md", "# Hello World\n\ntext")); got != "Hello World" {
		t.Fatalf("heading fallback: %q", got)
	}
	if got := fileSummary(write("c.go", "// Package graph does things.\npackage graph")); got != "Package graph does things." {
		t.Fatalf("comment de-mark: %q", got)
	}
	if got := fileSummary(write("d.bin", string([]byte{0, 1, 2, 3, 4}))); got != "" {
		t.Fatalf("binary should yield empty, got %q", got)
	}
	if got := fileSummary(filepath.Join(root, "missing")); got != "" {
		t.Fatalf("missing file should yield empty, got %q", got)
	}
}

func TestPopulateSummaries(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "readme.md"), []byte("# Title\nhi"), 0o600); err != nil {
		t.Fatal(err)
	}
	g := &Graph{Nodes: []Node{
		{ID: "project", Kind: "project"},
		{ID: "f/readme.md", Path: "readme.md", Kind: "doc"},
		{ID: "d/x", Path: "x", Kind: "directory"},
	}}
	g.populateSummaries(root)
	if g.Nodes[1].Summary != "Title" {
		t.Fatalf("doc node summary = %q", g.Nodes[1].Summary)
	}
	if g.Nodes[0].Summary != "" || g.Nodes[2].Summary != "" {
		t.Fatal("non-file nodes should have no summary")
	}
}

func TestClip(t *testing.T) {
	if got := clip("hello", 10); got != "hello" {
		t.Fatalf("short unchanged: %q", got)
	}
	if got := clip("hello world", 5); got != "hello…" {
		t.Fatalf("clipped: %q", got)
	}
}
