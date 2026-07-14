// Package ontology builds a deterministic structural ontology of a workspace: it enumerates the
// repository's files (honoring .gitignore), derives cheap, reproducible facts about each one, and
// emits a Turtle (.ttl) description of how files, directories, and data relate. The output is a
// fast navigation surface — a map an AI (or a person) can read to understand how a repo is
// organized before touching it. Generation is fully deterministic: the same tree yields byte-for-
// byte identical output, with no timestamps or randomness. See ttl.go for serialization.
package ontology

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// markerDir is the workspace-local Andromeda surface (ADR-028). Ontology artifacts live under
// <root>/.andromeda/ontology/. Defined locally to keep this package free of heavy imports.
const markerDir = ".andromeda"

// ontologySubdir is the directory (under the marker dir) that holds generated ontology files.
const ontologySubdir = "ontology"

// maxSummaryScan bounds how many bytes of a file are read to derive its cheap summary.
const maxSummaryScan = 8 << 10

// maxWalkFiles caps the fallback directory walk so a pathological tree cannot hang a scan.
const maxWalkFiles = 20000

// FileNode is a deterministic structural fact record for a single file.
type FileNode struct {
	Path     string // workspace-relative, slash-separated
	Dir      string // parent directory (workspace-relative, slash-separated; "." at the root)
	Name     string // base name
	Ext      string // extension without the dot, lowercased ("" if none)
	Language string // human-readable language/format ("Go", "Markdown", …; "" if unknown)
	Kind     string // coarse category: code | doc | config | data | asset | other
	Size     int64  // size in bytes (0 if unstatable)
	Summary  string // cheap, content-derived one-liner (Go package, first MD heading, …)
}

// Model is the deterministic ontology of a workspace: its files, the directories that group them,
// and language counts. All slices are sorted so serialization is reproducible.
type Model struct {
	Root      string         // absolute workspace root
	Name      string         // project name (base name of Root)
	Dirs      []string       // sorted unique directories (workspace-relative), excluding "."
	Files     []FileNode     // sorted by Path
	Languages map[string]int // language -> file count
}

// Scan reads the workspace at root and returns its deterministic ontology. It enumerates files via
// `git ls-files` when root is a git repo (fast, respects .gitignore) and otherwise falls back to a
// bounded directory walk that skips VCS, dependency, and Andromeda-internal directories. Results are
// sorted, so the returned Model is identical for identical inputs.
func Scan(ctx context.Context, root string) (*Model, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	rels := enumerate(ctx, abs)
	sort.Strings(rels)

	m := &Model{Root: abs, Name: filepath.Base(abs), Languages: map[string]int{}}
	dirSet := map[string]struct{}{}
	for _, rel := range rels {
		rel = filepath.ToSlash(rel)
		if rel == "" {
			continue
		}
		node := describe(abs, rel)
		m.Files = append(m.Files, node)
		if node.Language != "" {
			m.Languages[node.Language]++
		}
		for _, d := range ancestorDirs(rel) {
			dirSet[d] = struct{}{}
		}
	}
	for d := range dirSet {
		m.Dirs = append(m.Dirs, d)
	}
	sort.Strings(m.Dirs)
	return m, nil
}

// enumerate returns workspace-relative file paths, preferring git and falling back to a walk.
func enumerate(ctx context.Context, root string) []string {
	if out, err := exec.CommandContext(ctx, "git", "-C", root, "ls-files").Output(); err == nil { //nolint:gosec // fixed 'git' command; root is the workspace path
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			return filterInternal(lines)
		}
	}
	var files []string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", markerDir, "node_modules", "vendor", ".venv", "target", "dist", "build":
				return filepath.SkipDir
			}
			return nil
		}
		if rel, err := filepath.Rel(root, p); err == nil {
			files = append(files, filepath.ToSlash(rel))
		}
		if len(files) >= maxWalkFiles {
			return filepath.SkipAll
		}
		return nil
	})
	return files
}

// filterInternal drops Andromeda-internal paths (the .andromeda surface) from a git listing so the
// ontology never describes its own generated artifacts.
func filterInternal(paths []string) []string {
	out := paths[:0:0]
	for _, p := range paths {
		if p == markerDir || strings.HasPrefix(p, markerDir+"/") {
			continue
		}
		out = append(out, p)
	}
	return out
}

// ancestorDirs returns every directory that encloses rel (workspace-relative), excluding ".".
func ancestorDirs(rel string) []string {
	dir := path0(rel)
	var dirs []string
	for dir != "." && dir != "" {
		dirs = append(dirs, dir)
		dir = path0(dir)
	}
	return dirs
}

// path0 is filepath.Dir over slash-separated paths (avoids OS separator surprises on the model).
func path0(p string) string {
	i := strings.LastIndexByte(p, '/')
	if i < 0 {
		return "."
	}
	return p[:i]
}

// describe derives the deterministic facts for one workspace-relative file.
func describe(root, rel string) FileNode {
	name := rel[strings.LastIndexByte(rel, '/')+1:]
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	lang, kind := classify(name, ext)
	n := FileNode{Path: rel, Dir: path0(rel), Name: name, Ext: ext, Language: lang, Kind: kind}
	if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err == nil {
		n.Size = info.Size()
	}
	n.Summary = summarize(root, rel, ext)
	return n
}

// summarize reads a bounded prefix of the file to extract a cheap, deterministic one-liner: the
// Go package name, or a Markdown document's first heading. Anything else gets no summary.
func summarize(root, rel, ext string) string {
	switch ext {
	case "go", "md", "markdown":
	default:
		return ""
	}
	f, err := os.Open(filepath.Join(root, filepath.FromSlash(rel))) //nolint:gosec // workspace-relative path from a git/walk enumeration
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, maxSummaryScan)
	nRead, _ := f.Read(buf)
	sc := bufio.NewScanner(bytes.NewReader(buf[:nRead]))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch ext {
		case "go":
			if rest, ok := strings.CutPrefix(line, "package "); ok {
				if pkg := strings.Fields(rest); len(pkg) > 0 {
					return "package " + pkg[0]
				}
			}
		case "md", "markdown":
			if h, ok := strings.CutPrefix(line, "# "); ok {
				return strings.TrimSpace(h)
			}
		}
	}
	return ""
}
