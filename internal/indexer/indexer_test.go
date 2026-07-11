package indexer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestBuildQueryReady(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.go"), "package main\n// the provider router\nfunc handle() {}\n")
	writeFile(t, filepath.Join(root, "b.txt"), "the provider layer routes requests\n")
	writeFile(t, filepath.Join(root, "c.md"), "unrelated documentation content\n")

	e := New()
	id, err := e.Build(ctx, ports.IndexSpec{WorkspaceID: "ws", Include: []ports.Path{root}})
	if err != nil {
		t.Fatal(err)
	}
	st, _ := e.Status(ctx, id)
	if st.State != StateReady || st.Coverage != 3 {
		t.Fatalf("status = %+v", st)
	}
	hits, err := e.Query(ctx, id, ports.IndexQuery{Text: "provider"})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("hits for 'provider' = %d, want 2 (%v)", len(hits), hits)
	}
	if hits[0].Generation != st.Generation {
		t.Errorf("hit generation = %d, want %d", hits[0].Generation, st.Generation)
	}
}

func TestUpdateReindexesChangedFile(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	f := filepath.Join(root, "a.txt")
	writeFile(t, f, "alpha beta\n")
	e := New()
	id, _ := e.Build(ctx, ports.IndexSpec{Include: []ports.Path{root}})

	if hits, _ := e.Query(ctx, id, ports.IndexQuery{Text: "gamma"}); len(hits) != 0 {
		t.Fatal("gamma should not be present yet")
	}
	writeFile(t, f, "gamma delta\n")
	if err := e.Update(ctx, id, []ports.PathChange{{Path: f, Kind: "modified"}}); err != nil {
		t.Fatal(err)
	}
	if hits, _ := e.Query(ctx, id, ports.IndexQuery{Text: "gamma"}); len(hits) != 1 {
		t.Errorf("gamma should be indexed after update")
	}
	if hits, _ := e.Query(ctx, id, ports.IndexQuery{Text: "alpha"}); len(hits) != 0 {
		t.Errorf("stale term alpha should be gone after re-index")
	}
}

func TestExcludeAndBinarySkipped(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "keep.txt"), "keepterm here\n")
	writeFile(t, filepath.Join(root, "vendor", "skip.txt"), "keepterm hidden\n")
	os.WriteFile(filepath.Join(root, "bin.dat"), []byte{0, 1, 2, 0}, 0o600)

	e := New()
	id, _ := e.Build(ctx, ports.IndexSpec{Include: []ports.Path{root}, Exclude: []ports.Path{"vendor"}})
	hits, _ := e.Query(ctx, id, ports.IndexQuery{Text: "keepterm"})
	if len(hits) != 1 || filepath.Base(hits[0].Path) != "keep.txt" {
		t.Fatalf("exclude/binary handling wrong: %v", hits)
	}
}

func TestInvalidateWholeMarksStale(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), "x\n")
	e := New()
	id, _ := e.Build(ctx, ports.IndexSpec{Include: []ports.Path{root}})
	if err := e.Invalidate(ctx, id, ports.InvalidateScope{Whole: true}); err != nil {
		t.Fatal(err)
	}
	st, _ := e.Status(ctx, id)
	if !st.Stale || st.State != StateStale {
		t.Errorf("status after invalidate = %+v", st)
	}
}

func TestUnknownIndex(t *testing.T) {
	e := New()
	if _, err := e.Status(context.Background(), "nope"); err == nil {
		t.Error("expected error for unknown index")
	}
}
