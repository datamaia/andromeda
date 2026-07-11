package memory

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	db, err := storage.OpenWorkspaceDB(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return New(db)
}

func TestIngestAndRetrieve(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	ids, err := s.Ingest(ctx, []ports.MemoryRecordDraft{
		{Layer: "session", Content: "the deploy pipeline uses goreleaser", Source: "run-1"},
		{Layer: "workspace", Content: "prefer table-driven tests"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("ingested %d", len(ids))
	}

	got, err := s.Retrieve(ctx, ports.MemoryQuery{Text: "goreleaser"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Source != "run-1" {
		t.Fatalf("retrieve = %+v", got)
	}

	byLayer, _ := s.Retrieve(ctx, ports.MemoryQuery{Layers: []string{"workspace"}})
	if len(byLayer) != 1 || byLayer[0].Layer != "workspace" {
		t.Fatalf("layer filter = %+v", byLayer)
	}
}

func TestRankByTermOverlap(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	ids, _ := s.Ingest(ctx, []ports.MemoryRecordDraft{
		{Layer: "session", Content: "sqlite migrations run forward only"},
		{Layer: "session", Content: "the cat sat on the mat"},
	})
	ranked, err := s.Rank(ctx, ports.MemoryQuery{Text: "sqlite migrations"}, ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(ranked) != 2 || ranked[0].ID != ids[0] || ranked[0].Score <= ranked[1].Score {
		t.Fatalf("rank = %+v", ranked)
	}
}

func TestExpireAndDelete(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	ids, _ := s.Ingest(ctx, []ports.MemoryRecordDraft{{Layer: "session", Content: "ephemeral"}})
	rep, err := s.Expire(ctx, ports.ExpirePolicy{Layer: "session"})
	if err != nil {
		t.Fatal(err)
	}
	if rep.Expired != 1 {
		t.Errorf("expired = %d", rep.Expired)
	}
	// Expired records are not returned by active retrieval.
	if got, _ := s.Retrieve(ctx, ports.MemoryQuery{Text: "ephemeral"}); len(got) != 0 {
		t.Errorf("expired record still retrieved: %+v", got)
	}
	if err := s.Delete(ctx, ids); err != nil {
		t.Fatal(err)
	}
}

func TestExportStreams(t *testing.T) {
	ctx := context.Background()
	s := newStore(t)
	s.Ingest(ctx, []ports.MemoryRecordDraft{{Layer: "session", Content: "a"}, {Layer: "session", Content: "b"}})
	st, err := s.Export(ctx, ports.MemoryQuery{})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	var n int
	for {
		_, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 2 {
		t.Errorf("exported %d, want 2", n)
	}
}
