package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "t@e.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestUndoRedoCycle(t *testing.T) {
	repo := gitRepo(t)
	af := filepath.Join(repo, "a.txt")
	if err := os.WriteFile(af, []byte("v1"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := &tuiSession{ctx: context.Background(), wd: repo}

	// Simulate an agent turn: checkpoint the pre-turn state, then "the agent" edits the file.
	s.checkpointBeforeTurn()
	if len(s.undoStack) != 1 {
		t.Fatalf("expected one checkpoint, got %d", len(s.undoStack))
	}
	if err := os.WriteFile(af, []byte("v2-edited"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Undo reverts to v1 and stocks the redo stack.
	if got := s.undoAction(context.Background()); !strings.Contains(got, "reverted") {
		t.Fatalf("undo: %q", got)
	}
	if b, _ := os.ReadFile(af); string(b) != "v1" {
		t.Fatalf("undo did not revert: %q", b)
	}
	if len(s.redoStack) != 1 || len(s.undoStack) != 0 {
		t.Fatalf("stacks after undo: undo=%d redo=%d", len(s.undoStack), len(s.redoStack))
	}

	// Redo re-applies v2.
	if got := s.redoAction(context.Background()); !strings.Contains(got, "re-applied") {
		t.Fatalf("redo: %q", got)
	}
	if b, _ := os.ReadFile(af); string(b) != "v2-edited" {
		t.Fatalf("redo did not re-apply: %q", b)
	}
}

func TestUndoRedoEmptyAndNoGit(t *testing.T) {
	// Nothing captured yet.
	repo := gitRepo(t)
	s := &tuiSession{ctx: context.Background(), wd: repo}
	if got := s.undoAction(context.Background()); !strings.Contains(got, "nothing to undo") {
		t.Fatalf("empty undo: %q", got)
	}
	if got := s.redoAction(context.Background()); !strings.Contains(got, "nothing to redo") {
		t.Fatalf("empty redo: %q", got)
	}
	// Non-git workspace.
	ng := &tuiSession{ctx: context.Background(), wd: t.TempDir()}
	if got := ng.undoAction(context.Background()); !strings.Contains(got, "git repository") {
		t.Fatalf("no-git undo: %q", got)
	}
	// A checkpoint in a non-git workspace is a silent no-op.
	ng.checkpointBeforeTurn()
	if len(ng.undoStack) != 0 {
		t.Fatal("non-git checkpoint should not stack anything")
	}
}

func TestCheckpointDedupesNoOp(t *testing.T) {
	repo := gitRepo(t)
	if err := os.WriteFile(filepath.Join(repo, "a.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := &tuiSession{ctx: context.Background(), wd: repo}
	s.checkpointBeforeTurn()
	s.checkpointBeforeTurn() // nothing changed → identical tree → must not double-stack
	if len(s.undoStack) != 1 {
		t.Fatalf("no-op checkpoint should be deduped, got %d entries", len(s.undoStack))
	}
}
