// Package memnote is Andromeda's file-based workspace memory: a folder of Markdown notes under
// <root>/.andromeda/memory/ with a human- and agent-readable index (MEMORY.md). Each note is a small
// Markdown file with YAML-ish frontmatter — a low-cardinality consecutive id, a title, tags, and a
// creation date — so facts can be recalled by tag or keyword and linked from the index. It sits
// alongside AGENTS.md as the durable, inspectable memory a person edits and an agent can read (unlike
// the SQLite semantic store in internal/memory). No external YAML/Markdown dependencies: the tiny
// frontmatter format is parsed by hand.
package memnote

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	markerDir = ".andromeda"
	memSubdir = "memory"
	indexFile = "MEMORY.md"
)

// Note is a single workspace memory note.
type Note struct {
	ID      string   // zero-padded consecutive id, e.g. "0007"
	Title   string   // one-line title
	Tags    []string // free-form low-cardinality tags for recall
	Created string   // creation date, YYYY-MM-DD
	Body    string   // Markdown body (may be empty)
	slug    string   // filename base (without .md)
}

// Dir returns the memory directory for a workspace root.
func Dir(root string) string { return filepath.Join(root, markerDir, memSubdir) }

// Path is the on-disk file for a note.
func (n Note) Path(root string) string { return filepath.Join(Dir(root), n.slug+".md") }

// Slug exposes the note's filename base (for display).
func (n Note) Slug() string { return n.slug }

// List returns all notes, newest first (highest id first). Missing directory yields no notes.
func List(root string) ([]Note, error) {
	dir := Dir(root)
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var notes []Note
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == indexFile {
			continue
		}
		n, err := readNote(filepath.Join(dir, e.Name()))
		if err != nil {
			continue // skip unparseable files rather than fail the whole listing
		}
		notes = append(notes, n)
	}
	sort.Slice(notes, func(i, j int) bool { return notes[i].ID > notes[j].ID })
	return notes, nil
}

// Get returns the note with the given id.
func Get(root, id string) (Note, error) {
	id = normalizeID(id)
	notes, err := List(root)
	if err != nil {
		return Note{}, err
	}
	for _, n := range notes {
		if n.ID == id {
			return n, nil
		}
	}
	return Note{}, fmt.Errorf("no memory note %s", id)
}

// Add creates a note from a title (and optional tags/body), assigning the next consecutive id and
// today's date, writing the file and regenerating the index. Returns the stored note.
func Add(root, title string, tags []string, body string) (Note, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Note{}, fmt.Errorf("a memory note needs a title")
	}
	notes, err := List(root)
	if err != nil {
		return Note{}, err
	}
	n := Note{
		ID:      nextID(notes),
		Title:   title,
		Tags:    cleanTags(tags),
		Created: time.Now().UTC().Format("2006-01-02"),
		Body:    strings.TrimSpace(body),
	}
	n.slug = n.ID + "-" + slugify(title)
	if err := os.MkdirAll(Dir(root), 0o700); err != nil {
		return Note{}, err
	}
	if err := writeNote(root, n); err != nil {
		return Note{}, err
	}
	return n, writeIndex(root)
}

// Update replaces a note's body (title and tags are preserved) and refreshes the index.
func Update(root, id, body string) error {
	n, err := Get(root, id)
	if err != nil {
		return err
	}
	n.Body = strings.TrimSpace(body)
	if err := writeNote(root, n); err != nil {
		return err
	}
	return writeIndex(root)
}

// Delete removes a note and refreshes the index. A missing note is not an error.
func Delete(root, id string) error {
	n, err := Get(root, id)
	if err != nil {
		return nil
	}
	if err := os.Remove(n.Path(root)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return writeIndex(root)
}

// Search returns notes whose title, tags, or body contain the query (case-insensitive). An empty
// query returns everything.
func Search(root, query string) ([]Note, error) {
	notes, err := List(root)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return notes, nil
	}
	var out []Note
	for _, n := range notes {
		hay := strings.ToLower(n.Title + " " + strings.Join(n.Tags, " ") + " " + n.Body)
		if strings.Contains(hay, q) {
			out = append(out, n)
		}
	}
	return out, nil
}

// --- persistence ---------------------------------------------------------

func writeNote(root string, n Note) error {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "id: %s\n", n.ID)
	fmt.Fprintf(&b, "title: %s\n", n.Title)
	fmt.Fprintf(&b, "tags: %s\n", strings.Join(n.Tags, ", "))
	fmt.Fprintf(&b, "created: %s\n", n.Created)
	b.WriteString("---\n\n")
	if n.Body != "" {
		b.WriteString(n.Body)
		b.WriteString("\n")
	}
	return atomicWrite(n.Path(root), []byte(b.String()))
}

var frontmatterRe = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n?(.*)$`)

func readNote(path string) (Note, error) {
	data, err := os.ReadFile(path) //nolint:gosec // path comes from a directory listing under the marker dir
	if err != nil {
		return Note{}, err
	}
	m := frontmatterRe.FindSubmatch(data)
	if m == nil {
		return Note{}, fmt.Errorf("%s: missing frontmatter", filepath.Base(path))
	}
	n := Note{Body: strings.TrimSpace(string(m[2]))}
	for _, line := range strings.Split(string(m[1]), "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		switch strings.TrimSpace(key) {
		case "id":
			n.ID = val
		case "title":
			n.Title = val
		case "tags":
			n.Tags = cleanTags(strings.Split(val, ","))
		case "created":
			n.Created = val
		}
	}
	if n.ID == "" {
		return Note{}, fmt.Errorf("%s: note has no id", filepath.Base(path))
	}
	n.slug = strings.TrimSuffix(filepath.Base(path), ".md")
	return n, nil
}

// writeIndex regenerates MEMORY.md: a one-line-per-note index, newest first, so a person or an agent
// can scan what is remembered and open the specific note.
func writeIndex(root string) error {
	notes, err := List(root)
	if err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# Workspace memory\n\n")
	b.WriteString("Durable notes for this workspace, remembered by Andromeda. ")
	b.WriteString("One file per note under `.andromeda/memory/`; this index is generated.\n\n")
	if len(notes) == 0 {
		b.WriteString("_No notes yet._\n")
	}
	for _, n := range notes {
		tags := ""
		if len(n.Tags) > 0 {
			tags = " · " + strings.Join(n.Tags, ", ")
		}
		fmt.Fprintf(&b, "- [%s](%s.md) — %s%s (%s)\n", n.Title, n.slug, n.ID, tags, n.Created)
	}
	return atomicWrite(filepath.Join(Dir(root), indexFile), []byte(b.String()))
}

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
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

// --- helpers -------------------------------------------------------------

// nextID returns the next zero-padded consecutive id (one past the current maximum).
func nextID(notes []Note) string {
	maxID := 0
	for _, n := range notes {
		if v, err := strconv.Atoi(n.ID); err == nil && v > maxID {
			maxID = v
		}
	}
	return fmt.Sprintf("%04d", maxID+1)
}

// normalizeID accepts "7", "007", or "0007" and returns the zero-padded form.
func normalizeID(id string) string {
	if v, err := strconv.Atoi(strings.TrimSpace(id)); err == nil {
		return fmt.Sprintf("%04d", v)
	}
	return strings.TrimSpace(id)
}

func cleanTags(in []string) []string {
	var out []string
	for _, t := range in {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// slugify makes a short, filesystem-safe base name from a title.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = strings.Trim(s[:40], "-")
	}
	if s == "" {
		s = "note"
	}
	return s
}
