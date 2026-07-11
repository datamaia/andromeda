package indexer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// Frozen Index lifecycle states (Volume 2, chapter 09).
const (
	StateCreated  = "created"
	StateBuilding = "building"
	StateReady    = "ready"
	StateUpdating = "updating"
	StateStale    = "stale"
	StateFailed   = "failed"
	StateRemoved  = "removed"
)

// Engine implements ports.IndexerPort with in-memory lexical indexes.
type Engine struct {
	mu      sync.Mutex
	indexes map[core.ULID]*index
	// maxBytes bounds per-file indexing (binary/huge files are skipped).
	maxBytes int64
}

type index struct {
	spec       ports.IndexSpec
	state      string
	generation int64
	// postings maps a term to the set of file paths containing it.
	postings map[string]map[string]struct{}
	files    map[string]struct{}
}

// New returns an Indexing Engine.
func New() *Engine {
	return &Engine{indexes: map[core.ULID]*index{}, maxBytes: 1 << 20}
}

var _ ports.IndexerPort = (*Engine)(nil)

// Build declares and fully builds a lexical index for a workspace scope.
func (e *Engine) Build(ctx context.Context, spec ports.IndexSpec) (core.ULID, error) {
	id := core.NewULID()
	idx := &index{spec: spec, state: StateBuilding, postings: map[string]map[string]struct{}{}, files: map[string]struct{}{}}
	e.mu.Lock()
	e.indexes[id] = idx
	e.mu.Unlock()

	if err := e.rebuild(ctx, idx); err != nil {
		e.mu.Lock()
		idx.state = StateFailed
		e.mu.Unlock()
		return id, err
	}
	e.mu.Lock()
	idx.state = StateReady
	idx.generation++
	e.mu.Unlock()
	return id, nil
}

func (e *Engine) rebuild(ctx context.Context, idx *index) error {
	roots := idx.spec.Include
	if len(roots) == 0 {
		roots = []ports.Path{"."}
	}
	postings := map[string]map[string]struct{}{}
	files := map[string]struct{}{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable entries
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if d.IsDir() {
				if isExcluded(path, idx.spec.Exclude) || d.Name() == ".git" || d.Name() == ".andromeda" {
					return filepath.SkipDir
				}
				return nil
			}
			if isExcluded(path, idx.spec.Exclude) {
				return nil
			}
			e.indexFile(path, postings, files)
			return nil
		})
		if err != nil {
			return err
		}
	}
	idx.postings = postings
	idx.files = files
	return nil
}

func (e *Engine) indexFile(path string, postings map[string]map[string]struct{}, files map[string]struct{}) {
	info, err := os.Stat(path)
	if err != nil || info.Size() > e.maxBytes {
		return
	}
	data, err := os.ReadFile(path) //nolint:gosec // indexing workspace files is the engine's job
	if err != nil || !looksTextual(data) {
		return
	}
	files[path] = struct{}{}
	for term := range tokenize(string(data)) {
		set := postings[term]
		if set == nil {
			set = map[string]struct{}{}
			postings[term] = set
		}
		set[path] = struct{}{}
	}
}

// Update incrementally re-indexes changed paths.
func (e *Engine) Update(ctx context.Context, indexID core.ULID, changes []ports.PathChange) error {
	e.mu.Lock()
	idx, ok := e.indexes[indexID]
	e.mu.Unlock()
	if !ok {
		return idxErr("E-IDX-001", "unknown index")
	}
	e.mu.Lock()
	idx.state = StateUpdating
	e.mu.Unlock()
	for _, c := range changes {
		// Remove old postings for this path, then re-index if it still exists.
		removePath(idx, c.Path)
		if c.Kind != "deleted" {
			e.indexFile(c.Path, idx.postings, idx.files)
		}
	}
	e.mu.Lock()
	idx.state = StateReady
	idx.generation++
	e.mu.Unlock()
	return nil
}

// Query runs a lexical search, returning hits with the index generation.
func (e *Engine) Query(ctx context.Context, indexID core.ULID, q ports.IndexQuery) ([]ports.IndexHit, error) {
	e.mu.Lock()
	idx, ok := e.indexes[indexID]
	e.mu.Unlock()
	if !ok {
		return nil, idxErr("E-IDX-001", "unknown index")
	}
	terms := tokenize(q.Text)
	scores := map[string]int{}
	for t := range terms {
		for path := range idx.postings[t] {
			scores[path]++
		}
	}
	var hits []ports.IndexHit
	for path, score := range scores {
		hits = append(hits, ports.IndexHit{
			Path:       path,
			Score:      float64(score) / float64(max(1, len(terms))),
			Generation: idx.generation,
		})
	}
	sortHits(hits)
	if q.MaxResults > 0 && len(hits) > q.MaxResults {
		hits = hits[:q.MaxResults]
	}
	return hits, nil
}

// Invalidate marks a scope stale (forcing rebuild); dropping the whole index is always legal.
func (e *Engine) Invalidate(ctx context.Context, indexID core.ULID, scope ports.InvalidateScope) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	idx, ok := e.indexes[indexID]
	if !ok {
		return idxErr("E-IDX-001", "unknown index")
	}
	if scope.Whole {
		idx.state = StateStale
		return nil
	}
	for _, p := range scope.Paths {
		removePath(idx, p)
	}
	idx.state = StateStale
	return nil
}

// Status returns the current state, generation, and coverage of an index.
func (e *Engine) Status(ctx context.Context, indexID core.ULID) (ports.IndexStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	idx, ok := e.indexes[indexID]
	if !ok {
		return ports.IndexStatus{}, idxErr("E-IDX-001", "unknown index")
	}
	return ports.IndexStatus{
		State:      idx.state,
		Generation: idx.generation,
		Coverage:   len(idx.files),
		Stale:      idx.state == StateStale,
	}, nil
}

func idxErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "index", Severity: "error", Message: msg}
}

func removePath(idx *index, path string) {
	delete(idx.files, path)
	for term, set := range idx.postings {
		delete(set, path)
		if len(set) == 0 {
			delete(idx.postings, term)
		}
	}
}

func isExcluded(path string, excludes []ports.Path) bool {
	for _, ex := range excludes {
		if strings.Contains(path, ex) {
			return true
		}
	}
	return false
}

func looksTextual(data []byte) bool {
	n := len(data)
	if n > 8000 {
		n = 8000
	}
	for _, b := range data[:n] {
		if b == 0 {
			return false
		}
	}
	return true
}

func tokenize(s string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, f := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(r == '_' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}) {
		if len(f) > 1 {
			set[f] = struct{}{}
		}
	}
	return set
}
