package graph

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ontology"
)

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

func sampleModel(t *testing.T) (*ontology.Model, string) {
	t.Helper()
	root := t.TempDir()
	writeFile(t, root, "main.go", "package main\n")
	writeFile(t, root, "README.md", "# Proj\n")
	writeFile(t, root, "internal/util/util.go", "package util\n")
	writeFile(t, root, "assets/logo.png", "\x89PNG")
	m, err := ontology.Scan(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	return m, root
}

func TestBuildGraphContainment(t *testing.T) {
	m, _ := sampleModel(t)
	g := Build(m)

	byID := map[string]Node{}
	for _, n := range g.Nodes {
		byID[n.ID] = n
	}
	if _, ok := byID["project"]; !ok {
		t.Fatal("missing project node")
	}
	if n := byID["f/internal/util/util.go"]; n.Kind != "code" || n.Group != "internal" {
		t.Errorf("util.go node = %+v", n)
	}
	if n := byID["f/assets/logo.png"]; n.Kind != "asset" || n.Group != "assets" {
		t.Errorf("logo.png node = %+v", n)
	}
	// Every edge connects two real nodes; root-level files hang off the project.
	for _, e := range g.Edges {
		if _, ok := byID[e.From]; !ok {
			t.Errorf("edge from unknown node %q", e.From)
		}
		if _, ok := byID[e.To]; !ok {
			t.Errorf("edge to unknown node %q", e.To)
		}
	}
	if p := parentID("main.go"); p != "project" {
		t.Errorf("root file parent = %q, want project", p)
	}
	if p := parentID("internal/util/util.go"); p != "d/internal/util" {
		t.Errorf("nested file parent = %q", p)
	}
}

func TestGraphJSONIsDeterministic(t *testing.T) {
	m, _ := sampleModel(t)
	a, b := Build(m).JSON(), Build(m).JSON()
	if string(a) != string(b) {
		t.Fatal("graph JSON is not deterministic")
	}
	var parsed Graph
	if err := json.Unmarshal(a, &parsed); err != nil {
		t.Fatalf("graph JSON does not round-trip: %v", err)
	}
	if len(parsed.Nodes) == 0 || len(parsed.Edges) == 0 {
		t.Fatal("parsed graph is empty")
	}
}

func TestWriteProducesArtifactsAndRemove(t *testing.T) {
	m, root := sampleModel(t)
	g, dir, err := Write(root, m)
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"graph.json", "manifest.json", "index.md", "groups/internal.md", "groups/assets.md"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected artifact %s: %v", rel, err)
		}
	}
	idx, err := os.ReadFile(filepath.Join(dir, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(idx), "```mermaid") || !strings.Contains(string(idx), "workspace map") {
		t.Errorf("index.md missing expected sections:\n%s", idx)
	}
	if g.Stats() == "" {
		t.Error("empty stats")
	}

	// Rewriting after deleting a group drops that group's stale note.
	if err := os.Remove(filepath.Join(root, "assets", "logo.png")); err != nil {
		t.Fatal(err)
	}
	m2, _ := ontology.Scan(context.Background(), root)
	if _, _, err := Write(root, m2); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "groups", "assets.md")); !os.IsNotExist(err) {
		t.Error("stale groups/assets.md should have been removed on rewrite")
	}

	if err := Remove(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("graph dir should be gone after Remove")
	}
	if err := Remove(root); err != nil {
		t.Errorf("second Remove errored: %v", err)
	}
}
