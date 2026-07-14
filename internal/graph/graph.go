// Package graph turns a workspace's structural ontology into a visual, navigable graph. It derives
// a node/edge model from internal/ontology, writes a set of human-readable Markdown notes plus a
// deterministic graph.json under <root>/.andromeda/graph/, and can serve a small self-contained
// force-directed viewer over localhost. Like the ontology package, generation is fully
// deterministic — the same tree yields byte-for-byte identical artifacts, with no timestamps.
package graph

import (
	"encoding/json"
	"strings"

	"github.com/datamaia/andromeda/internal/ontology"
)

// markerDir is the workspace-local Andromeda surface (ADR-028). Graph artifacts live under
// <root>/.andromeda/graph/. Defined locally to keep this package free of heavy imports.
const markerDir = ".andromeda"

// graphSubdir is the directory (under the marker dir) that holds generated graph files.
const graphSubdir = "graph"

// Node is a vertex in the workspace graph: the project root, a directory, or a file.
type Node struct {
	ID    string `json:"id"`    // stable identity: "project", "d/<dir>", or "f/<path>"
	Label string `json:"label"` // display name (base name)
	Kind  string `json:"kind"`  // project | directory | code | doc | config | data | asset | other
	Path  string `json:"path"`  // workspace-relative path ("" for the project node)
	Group string `json:"group"` // top-level directory, for clustering/coloring ("" at the root)
	Size  int64  `json:"size"`  // byte size for files (0 for project/directories)
}

// Edge is a directed containment relationship: From contains To.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Rel  string `json:"rel"` // always "contains" for now
}

// Graph is a serializable node/edge view of a workspace, derived from an ontology model.
type Graph struct {
	Name  string `json:"name"`
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

const projectID = "project"

// Build derives a graph from an ontology model. It is deterministic: nodes and edges follow the
// model's already-sorted directories and files (project first, then directories, then files).
func Build(m *ontology.Model) *Graph {
	g := &Graph{Name: m.Name}
	g.Nodes = append(g.Nodes, Node{ID: projectID, Label: m.Name, Kind: "project"})

	for _, d := range m.Dirs {
		g.Nodes = append(g.Nodes, Node{
			ID:    dirID(d),
			Label: baseName(d),
			Kind:  "directory",
			Path:  d,
			Group: topLevel(d),
		})
		g.Edges = append(g.Edges, Edge{From: parentID(d), To: dirID(d), Rel: "contains"})
	}

	for _, f := range m.Files {
		g.Nodes = append(g.Nodes, Node{
			ID:    fileID(f.Path),
			Label: f.Name,
			Kind:  f.Kind,
			Path:  f.Path,
			Group: topLevel(f.Path),
			Size:  f.Size,
		})
		g.Edges = append(g.Edges, Edge{From: parentID(f.Path), To: fileID(f.Path), Rel: "contains"})
	}
	return g
}

// JSON renders the graph as pretty-printed, deterministic JSON (trailing newline).
func (g *Graph) JSON() []byte {
	data, _ := json.MarshalIndent(g, "", "  ")
	return append(data, '\n')
}

// --- identity & path helpers ---------------------------------------------

func fileID(p string) string { return "f/" + p }
func dirID(p string) string  { return "d/" + p }

// parentID is the containing node's ID: the parent directory, or the project when at the root.
func parentID(p string) string {
	parent := dirOf(p)
	if parent == "." || parent == "" {
		return projectID
	}
	return dirID(parent)
}

// dirOf is filepath.Dir over slash-separated paths.
func dirOf(p string) string {
	i := strings.LastIndexByte(p, '/')
	if i < 0 {
		return "."
	}
	return p[:i]
}

func baseName(p string) string { return p[strings.LastIndexByte(p, '/')+1:] }

// topLevel returns the first path segment — the top-level group a node belongs to.
func topLevel(p string) string {
	if i := strings.IndexByte(p, '/'); i >= 0 {
		return p[:i]
	}
	return p
}
