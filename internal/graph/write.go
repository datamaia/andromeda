package graph

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/datamaia/andromeda/internal/ontology"
)

// Write (re)generates the graph artifacts under <root>/.andromeda/graph/: graph.json (the node/edge
// model), a set of Markdown notes (index.md + groups/*.md), and a manifest.json content hash. It
// replaces the notes directory wholesale so stale group files never linger, and returns the built
// graph along with the output directory.
func Write(root string, m *ontology.Model) (*Graph, string, error) {
	g := Build(m)
	g.populateSummaries(root)
	dir := Dir(root)

	// Reset the group notes directory so removed groups don't leave orphan files behind.
	if err := os.RemoveAll(filepath.Join(dir, "groups")); err != nil {
		return nil, "", err
	}
	if err := os.MkdirAll(filepath.Join(dir, "groups"), 0o700); err != nil {
		return nil, "", err
	}

	if err := atomicWrite(filepath.Join(dir, "graph.json"), g.JSON()); err != nil {
		return nil, "", err
	}
	for rel, content := range renderMarkdown(m) {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			return nil, "", err
		}
		if err := atomicWrite(p, []byte(content)); err != nil {
			return nil, "", err
		}
	}
	if err := atomicWrite(filepath.Join(dir, "manifest.json"), g.manifest()); err != nil {
		return nil, "", err
	}
	return g, dir, nil
}

// populateSummaries reads a short head excerpt from each text file node and stores it as the node's
// Summary, so the viewer can show an Obsidian-style preview on hover. Binary, oversized, and unreadable
// files are skipped (Summary stays empty). Deterministic given file contents.
func (g *Graph) populateSummaries(root string) {
	for i := range g.Nodes {
		n := &g.Nodes[i]
		if n.Path == "" || !summarizableKind(n.Kind) {
			continue
		}
		if s := fileSummary(filepath.Join(root, filepath.FromSlash(n.Path))); s != "" {
			n.Summary = s
		}
	}
}

// summarizableKind reports whether a node kind is text worth previewing.
func summarizableKind(kind string) bool {
	switch kind {
	case "doc", "code", "config", "data":
		return true
	}
	return false
}

// fileSummary returns a one-line preview from the head of a text file: a frontmatter description when
// present, else the first meaningful (de-marked, non-empty) line, capped in length. Binary files and
// read errors yield "".
func fileSummary(path string) string {
	data, err := os.ReadFile(path) //nolint:gosec // path is a workspace-relative file under root
	if err != nil {
		return ""
	}
	if len(data) > 8192 {
		data = data[:8192]
	}
	if bytes.IndexByte(data, 0) >= 0 { // NUL byte ⇒ treat as binary, skip
		return ""
	}
	inFront := false
	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if i == 0 && line == "---" { // YAML frontmatter opens
			inFront = true
			continue
		}
		if inFront {
			if line == "---" {
				inFront = false
			} else if v, ok := strings.CutPrefix(line, "description:"); ok {
				return clip(strings.Trim(strings.TrimSpace(v), `"'`), 160)
			}
			continue
		}
		if s := strings.TrimSpace(strings.TrimLeft(line, "#>/*-! \t")); s != "" {
			return clip(s, 160)
		}
	}
	return ""
}

// clip truncates s to at most n runes, appending an ellipsis when it was longer.
func clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + "…"
}

// Remove deletes the generated graph directory. Missing is not an error.
func Remove(root string) error {
	err := os.RemoveAll(Dir(root))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Dir returns the graph output directory for a workspace root.
func Dir(root string) string { return filepath.Join(root, markerDir, graphSubdir) }

// Stats renders a short human summary of the graph for CLI/TUI output.
func (g *Graph) Stats() string {
	files, dirs := 0, 0
	for _, n := range g.Nodes {
		switch n.Kind {
		case "directory":
			dirs++
		case "project":
		default:
			files++
		}
	}
	return fmt.Sprintf("%d nodes (%d files, %d directories), %d edges",
		len(g.Nodes), files, dirs, len(g.Edges))
}

// manifest is a deterministic JSON sidecar: counts plus a content hash over the node ids, so a
// caller can detect whether a re-scan would change the graph.
func (g *Graph) manifest() []byte {
	ids := make([]string, len(g.Nodes))
	for i, n := range g.Nodes {
		ids[i] = n.ID
	}
	sort.Strings(ids)
	h := sha256.New()
	for _, id := range ids {
		_, _ = fmt.Fprintf(h, "%s\n", id)
	}
	man := struct {
		Name      string `json:"name"`
		NodeCount int    `json:"nodeCount"`
		EdgeCount int    `json:"edgeCount"`
		Hash      string `json:"hash"`
	}{g.Name, len(g.Nodes), len(g.Edges), hex.EncodeToString(h.Sum(nil))}
	data, _ := json.MarshalIndent(man, "", "  ")
	return append(data, '\n')
}

// atomicWrite writes data to path via a temp file + rename, with owner-only perms.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
