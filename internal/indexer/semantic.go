package indexer

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// Embedder produces embedding vectors for a batch of texts (backed by a ProviderPort.Embed in
// production, ADR-020). It is injected so the semantic index is testable without a live model.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// SemanticEngine is an in-memory semantic index over workspace files: it embeds file contents
// and answers queries by cosine similarity. Vectors are a rebuildable cache (INV-IDX-02).
type SemanticEngine struct {
	embedder Embedder
	maxBytes int64

	mu      sync.Mutex
	indexes map[core.ULID]*semIndex
}

type semIndex struct {
	spec       ports.IndexSpec
	state      string
	generation int64
	vectors    map[string][]float32 // path -> embedding
}

// NewSemantic returns a semantic Indexing Engine over the given embedder.
func NewSemantic(embedder Embedder) *SemanticEngine {
	return &SemanticEngine{embedder: embedder, maxBytes: 1 << 20, indexes: map[core.ULID]*semIndex{}}
}

var _ ports.IndexerPort = (*SemanticEngine)(nil)

// Build embeds every in-scope file and stores its vector.
func (e *SemanticEngine) Build(ctx context.Context, spec ports.IndexSpec) (core.ULID, error) {
	id := core.NewULID()
	idx := &semIndex{spec: spec, state: StateBuilding, vectors: map[string][]float32{}}
	e.mu.Lock()
	e.indexes[id] = idx
	e.mu.Unlock()

	var paths []string
	var texts []string
	roots := spec.Include
	if len(roots) == 0 {
		roots = []ports.Path{"."}
	}
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if d.IsDir() {
				if d.Name() == ".git" || d.Name() == ".andromeda" || isExcluded(path, spec.Exclude) {
					return filepath.SkipDir
				}
				return nil
			}
			if isExcluded(path, spec.Exclude) {
				return nil
			}
			info, err := d.Info()
			if err != nil || info.Size() > e.maxBytes {
				return nil
			}
			data, err := os.ReadFile(path) //nolint:gosec // indexing workspace files
			if err != nil || !looksTextual(data) {
				return nil
			}
			paths = append(paths, path)
			texts = append(texts, string(data))
			return nil
		})
	}
	if len(texts) > 0 {
		vecs, err := e.embedder.Embed(ctx, texts)
		if err != nil {
			e.setState(idx, StateFailed)
			return id, &ports.PortError{Code: "E-IDX-010", Category: "index", Message: "embedding failed", Detail: err.Error(), Cause: err}
		}
		for i := range paths {
			if i < len(vecs) {
				idx.vectors[paths[i]] = vecs[i]
			}
		}
	}
	e.mu.Lock()
	idx.state = StateReady
	idx.generation++
	e.mu.Unlock()
	return id, nil
}

// Query embeds the query text and returns the most cosine-similar files.
func (e *SemanticEngine) Query(ctx context.Context, indexID core.ULID, q ports.IndexQuery) ([]ports.IndexHit, error) {
	e.mu.Lock()
	idx, ok := e.indexes[indexID]
	e.mu.Unlock()
	if !ok {
		return nil, idxErr("E-IDX-001", "unknown index")
	}
	qv, err := e.embedder.Embed(ctx, []string{q.Text})
	if err != nil || len(qv) == 0 {
		return nil, idxErr("E-IDX-010", "query embedding failed")
	}
	var hits []ports.IndexHit
	for path, v := range idx.vectors {
		hits = append(hits, ports.IndexHit{Path: path, Score: cosine(qv[0], v), Generation: idx.generation})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if q.MaxResults > 0 && len(hits) > q.MaxResults {
		hits = hits[:q.MaxResults]
	}
	return hits, nil
}

// Update re-embeds changed files.
func (e *SemanticEngine) Update(ctx context.Context, indexID core.ULID, changes []ports.PathChange) error {
	e.mu.Lock()
	idx, ok := e.indexes[indexID]
	e.mu.Unlock()
	if !ok {
		return idxErr("E-IDX-001", "unknown index")
	}
	for _, c := range changes {
		if c.Kind == "deleted" {
			e.mu.Lock()
			delete(idx.vectors, c.Path)
			e.mu.Unlock()
			continue
		}
		data, err := os.ReadFile(c.Path) //nolint:gosec // indexing workspace files
		if err != nil {
			continue
		}
		vecs, err := e.embedder.Embed(ctx, []string{string(data)})
		if err == nil && len(vecs) > 0 {
			e.mu.Lock()
			idx.vectors[c.Path] = vecs[0]
			e.mu.Unlock()
		}
	}
	e.mu.Lock()
	idx.generation++
	e.mu.Unlock()
	return nil
}

// Invalidate marks a semantic index stale.
func (e *SemanticEngine) Invalidate(_ context.Context, indexID core.ULID, scope ports.InvalidateScope) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	idx, ok := e.indexes[indexID]
	if !ok {
		return idxErr("E-IDX-001", "unknown index")
	}
	if scope.Whole {
		idx.vectors = map[string][]float32{}
	} else {
		for _, p := range scope.Paths {
			delete(idx.vectors, p)
		}
	}
	idx.state = StateStale
	return nil
}

// Status returns the semantic index state.
func (e *SemanticEngine) Status(_ context.Context, indexID core.ULID) (ports.IndexStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	idx, ok := e.indexes[indexID]
	if !ok {
		return ports.IndexStatus{}, idxErr("E-IDX-001", "unknown index")
	}
	return ports.IndexStatus{State: idx.state, Generation: idx.generation, Coverage: len(idx.vectors), Stale: idx.state == StateStale}, nil
}

func (e *SemanticEngine) setState(idx *semIndex, s string) {
	e.mu.Lock()
	idx.state = s
	e.mu.Unlock()
}

// cosine returns the cosine similarity of two vectors (0 for mismatched/empty).
func cosine(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// ProviderEmbedder adapts a ports.ProviderPort into an Embedder, backing the semantic index
// with a real model's embeddings (ADR-020, ADR-019).
type ProviderEmbedder struct {
	Provider ports.ProviderPort
	Model    string
}

// Embed embeds texts via the provider's Embed method.
func (p ProviderEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	resp, err := p.Provider.Embed(ctx, ports.EmbedRequest{Model: p.Model, Inputs: texts})
	if err != nil {
		return nil, err
	}
	return resp.Vectors, nil
}
