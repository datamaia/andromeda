package indexer

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// bagEmbedder is a deterministic test embedder: a bag-of-words vector over a fixed vocabulary,
// so texts sharing words have high cosine similarity — enough to exercise semantic retrieval.
type bagEmbedder struct{ vocab []string }

func (b bagEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, txt := range texts {
		terms := tokenize(txt)
		v := make([]float32, len(b.vocab))
		for j, w := range b.vocab {
			if _, ok := terms[w]; ok {
				v[j] = 1
			}
		}
		out[i] = v
	}
	return out, nil
}

func TestSemanticRetrieval(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "providers.md"), "the provider router dispatches model requests\n")
	writeFile(t, filepath.Join(root, "cooking.md"), "recipe for bread with flour and yeast\n")

	emb := bagEmbedder{vocab: []string{"provider", "router", "model", "recipe", "bread", "flour"}}
	e := NewSemantic(emb)
	id, err := e.Build(ctx, ports.IndexSpec{Kind: "semantic", Include: []ports.Path{root}})
	if err != nil {
		t.Fatal(err)
	}
	st, _ := e.Status(ctx, id)
	if st.State != StateReady || st.Coverage != 2 {
		t.Fatalf("status = %+v", st)
	}

	hits, err := e.Query(ctx, id, ports.IndexQuery{Text: "which file is about the model provider?", MaxResults: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 || filepath.Base(hits[0].Path) != "providers.md" {
		t.Fatalf("top semantic hit = %v", hits)
	}
	if hits[0].Score <= hits[len(hits)-1].Score && len(hits) > 1 {
		t.Errorf("hits not sorted by score: %v", hits)
	}
}

func TestSemanticUpdateAndInvalidate(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	f := filepath.Join(root, "a.md")
	writeFile(t, f, "provider content\n")
	emb := bagEmbedder{vocab: []string{"provider", "recipe"}}
	e := NewSemantic(emb)
	id, _ := e.Build(ctx, ports.IndexSpec{Kind: "semantic", Include: []ports.Path{root}})

	writeFile(t, f, "recipe content\n")
	if err := e.Update(ctx, id, []ports.PathChange{{Path: f, Kind: "modified"}}); err != nil {
		t.Fatal(err)
	}
	hits, _ := e.Query(ctx, id, ports.IndexQuery{Text: "recipe", MaxResults: 1})
	if len(hits) != 1 || hits[0].Score == 0 {
		t.Errorf("expected the updated vector to match 'recipe': %v", hits)
	}

	if err := e.Invalidate(ctx, id, ports.InvalidateScope{Whole: true}); err != nil {
		t.Fatal(err)
	}
	st, _ := e.Status(ctx, id)
	if !st.Stale || st.Coverage != 0 {
		t.Errorf("status after invalidate = %+v", st)
	}
}

func TestCosine(t *testing.T) {
	if v := cosine([]float32{1, 0}, []float32{1, 0}); v < 0.999 {
		t.Errorf("identical vectors cosine = %v, want ~1", v)
	}
	if v := cosine([]float32{1, 0}, []float32{0, 1}); v != 0 {
		t.Errorf("orthogonal vectors cosine = %v, want 0", v)
	}
	if v := cosine([]float32{1}, []float32{1, 2}); v != 0 {
		t.Errorf("mismatched-length cosine = %v, want 0", v)
	}
}
