package indexer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// BenchmarkSemanticQuery measures semantic search latency: in-process cosine similarity over a
// small embedded index. Relates to NFR-PERF-013 (search latency). Uses the deterministic
// bagEmbedder from semantic_test.go. Micro-benchmark tier, Volume 12 chapter 03.
func BenchmarkSemanticQuery(b *testing.B) {
	ctx := context.Background()
	root := b.TempDir()
	for i := 0; i < 20; i++ {
		p := filepath.Join(root, fmt.Sprintf("doc%d.md", i))
		if err := os.WriteFile(p, []byte("the provider router dispatches model requests for the deploy pipeline\n"), 0o600); err != nil {
			b.Fatal(err)
		}
	}

	emb := bagEmbedder{vocab: []string{"provider", "router", "model", "deploy", "pipeline", "recipe", "bread", "flour", "sqlite", "migration"}}
	e := NewSemantic(emb)
	id, err := e.Build(ctx, ports.IndexSpec{Kind: "semantic", Include: []ports.Path{root}})
	if err != nil {
		b.Fatal(err)
	}

	q := ports.IndexQuery{Text: "which file is about the model provider?", MaxResults: 5}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := e.Query(ctx, id, q); err != nil {
			b.Fatal(err)
		}
	}
}
