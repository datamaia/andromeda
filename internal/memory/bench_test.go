package memory

import (
	"context"
	"fmt"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
)

// BenchmarkRetrieve measures memory retrieval latency over a populated store (term-overlap
// ranking across 200 records). Relates to NFR-PERF-012 (memory retrieval latency).
// Micro-benchmark tier, Volume 12 chapter 03.
func BenchmarkRetrieve(b *testing.B) {
	ctx := context.Background()
	db, err := storage.OpenWorkspaceDB(ctx, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	s := New(db)

	drafts := make([]ports.MemoryRecordDraft, 0, 200)
	for i := 0; i < 200; i++ {
		drafts = append(drafts, ports.MemoryRecordDraft{
			Layer:   "session",
			Content: fmt.Sprintf("record %d: the deploy pipeline uses goreleaser and sqlite migrations", i),
		})
	}
	if _, err := s.Ingest(ctx, drafts); err != nil {
		b.Fatal(err)
	}

	q := ports.MemoryQuery{Text: "goreleaser sqlite migrations"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Retrieve(ctx, q); err != nil {
			b.Fatal(err)
		}
	}
}
