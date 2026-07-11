package workspace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/git"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
)

func newEngine(t *testing.T) *Engine {
	t.Helper()
	gdb, err := storage.OpenGlobalDB(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { gdb.Close() })
	return New(gdb, git.New(""))
}

func TestOpenCreatesMarkerAndRegisters(t *testing.T) {
	ctx := context.Background()
	e := newEngine(t)
	root := t.TempDir()

	h, err := e.Open(ctx, root, ports.OpenOptions{CreateIfMissing: true})
	if err != nil {
		t.Fatal(err)
	}
	if h.ID == "" {
		t.Fatal("expected a workspace id")
	}
	if !isDir(filepath.Join(root, storage.MarkerDir)) {
		t.Error(".andromeda marker not created")
	}
	// Registered in the global registry.
	var count int
	e.global.SQL().QueryRow(`SELECT COUNT(*) FROM workspaces WHERE root = ?`, mustAbs(t, root)).Scan(&count)
	if count != 1 {
		t.Errorf("workspace registered %d times, want 1", count)
	}
	if err := e.Close(ctx, h); err != nil {
		t.Fatal(err)
	}
}

func TestOpenWithoutMarkerRefusedUnlessCreate(t *testing.T) {
	ctx := context.Background()
	e := newEngine(t)
	_, err := e.Open(ctx, t.TempDir(), ports.OpenOptions{CreateIfMissing: false})
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-AGT-010" {
		t.Fatalf("want E-AGT-010, got %v", err)
	}
}

func TestDiscoverFindsMarkerUpward(t *testing.T) {
	ctx := context.Background()
	e := newEngine(t)
	root := t.TempDir()
	if _, err := e.Open(ctx, root, ports.OpenOptions{CreateIfMissing: true}); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	cand, err := e.Discover(ctx, nested)
	if err != nil {
		t.Fatal(err)
	}
	if !cand.Found || !cand.HasMarker || cand.Root != mustAbs(t, root) {
		t.Fatalf("discover = %+v", cand)
	}
}

func TestDiscoverFindsGitRepo(t *testing.T) {
	ctx := context.Background()
	e := newEngine(t)
	root := t.TempDir()
	cmd := exec.Command("git", "init", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	cand, err := e.Discover(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	if !cand.Found || !cand.IsRepo {
		t.Fatalf("discover repo = %+v", cand)
	}
}

func TestSnapshotIncludesVCSSummary(t *testing.T) {
	ctx := context.Background()
	e := newEngine(t)
	root := t.TempDir()
	// Make it a git repo with a commit so Status works.
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.email", "t@e.com"}, {"config", "user.name", "T"}} {
		c := exec.Command("git", args...)
		c.Dir = root
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	os.WriteFile(filepath.Join(root, "f.txt"), []byte("x"), 0o600)
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "init"}} {
		c := exec.Command("git", args...)
		c.Dir = root
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=T", "GIT_AUTHOR_EMAIL=t@e.com", "GIT_COMMITTER_NAME=T", "GIT_COMMITTER_EMAIL=t@e.com")
		c.CombinedOutput()
	}

	h, _ := e.Open(ctx, root, ports.OpenOptions{CreateIfMissing: true})
	snap, err := e.Snapshot(ctx, h)
	if err != nil {
		t.Fatal(err)
	}
	if snap.Root != mustAbs(t, root) {
		t.Errorf("snapshot root = %q", snap.Root)
	}
	if snap.VCSSummary == "" {
		t.Error("expected a VCS summary")
	}
}

func TestReopenKeepsSameID(t *testing.T) {
	ctx := context.Background()
	e := newEngine(t)
	root := t.TempDir()
	h1, _ := e.Open(ctx, root, ports.OpenOptions{CreateIfMissing: true})
	_ = e.Close(ctx, h1)
	h2, _ := e.Open(ctx, root, ports.OpenOptions{CreateIfMissing: true})
	if h1.ID != h2.ID {
		t.Errorf("reopen changed id: %s != %s", h1.ID, h2.ID)
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	a, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
