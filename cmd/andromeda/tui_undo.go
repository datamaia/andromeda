package main

import (
	"context"

	"github.com/datamaia/andromeda/internal/checkpoint"
)

// checkpointBeforeTurn snapshots the working tree before an agent turn so /undo can revert the
// turn's file changes. It runs on the UI thread (the caller, startAgentRun, invokes it before
// spawning the run goroutine), so the undo/redo stacks are only ever touched from one goroutine and
// need no locking. Best-effort: a non-git workspace or a git hiccup simply yields no checkpoint.
func (s *tuiSession) checkpointBeforeTurn() {
	if !checkpoint.Available(s.wd) {
		return
	}
	tree, err := checkpoint.Snapshot(s.wd)
	if err != nil || tree == "" {
		return
	}
	// Skip a no-op checkpoint (nothing changed since the last one) so /undo steps over real changes.
	if n := len(s.undoStack); n > 0 && s.undoStack[n-1] == tree {
		return
	}
	s.undoStack = append(s.undoStack, tree)
	s.redoStack = nil // a fresh change invalidates the redo history
}

// undoAction backs /undo: it restores the working tree to the snapshot taken before the last agent
// turn, first snapshotting the current state so /redo can return to it.
func (s *tuiSession) undoAction(_ context.Context) string {
	if !checkpoint.Available(s.wd) {
		return "undo needs a git repository in the workspace"
	}
	if len(s.undoStack) == 0 {
		return "nothing to undo"
	}
	if cur, err := checkpoint.Snapshot(s.wd); err == nil && cur != "" {
		s.redoStack = append(s.redoStack, cur)
	}
	target := s.undoStack[len(s.undoStack)-1]
	s.undoStack = s.undoStack[:len(s.undoStack)-1]
	if err := checkpoint.Restore(s.wd, target); err != nil {
		return "undo failed: " + err.Error()
	}
	return "undo · reverted the workspace to before the last change (redo with /redo)"
}

// redoAction backs /redo: it re-applies a change undone by /undo, snapshotting the current state so
// /undo can step back again.
func (s *tuiSession) redoAction(_ context.Context) string {
	if !checkpoint.Available(s.wd) {
		return "redo needs a git repository in the workspace"
	}
	if len(s.redoStack) == 0 {
		return "nothing to redo"
	}
	if cur, err := checkpoint.Snapshot(s.wd); err == nil && cur != "" {
		s.undoStack = append(s.undoStack, cur)
	}
	target := s.redoStack[len(s.redoStack)-1]
	s.redoStack = s.redoStack[:len(s.redoStack)-1]
	if err := checkpoint.Restore(s.wd, target); err != nil {
		return "redo failed: " + err.Error()
	}
	return "redo · re-applied the change (undo again with /undo)"
}
