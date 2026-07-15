package main

import (
	"context"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/ports"
)

// msg builds a one-part message for session test transcripts.
func msg(role, text string) ports.Message {
	return ports.Message{Role: role, Parts: []ports.ContentPart{{Type: "text", Text: text}}}
}

// convo is a small two-turn transcript used across the session tests.
func convo() []ports.Message {
	return []ports.Message{msg("user", "add a cache"), msg("assistant", "done")}
}

func TestBranchSnapshotsAndTree(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	s := &tuiSession{ctx: context.Background(), sessionID: "root", history: convo(),
		cfg: tuiConfig{provider: "groq", model: "x"}}
	s.persistSession("agent") // root must be on disk for the tree to nest the branch under it

	out := s.branchAction(context.Background())
	if !strings.Contains(out, "branched") {
		t.Fatalf("branch status: %q", out)
	}
	if s.sessionID != "root" {
		t.Fatalf("/branch must keep the user on the current session, got %q", s.sessionID)
	}
	sessions, err := app.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Fatalf("want root + branch, got %d", len(sessions))
	}
	var fork app.StoredSession
	for _, st := range sessions {
		if st.ID != "root" {
			fork = st
		}
	}
	if fork.Parent != "root" {
		t.Fatalf("branch parent = %q, want root", fork.Parent)
	}
	tree := s.sessionTreeAction(context.Background())
	if !strings.Contains(tree, "root") || !strings.Contains(tree, fork.ID) {
		t.Fatalf("tree missing lineage: %q", tree)
	}
	// The branch line must be indented under root (nested), and root marked current.
	if !strings.Contains(tree, "  • "+fork.ID) {
		t.Fatalf("branch not nested under root: %q", tree)
	}
	if !strings.Contains(tree, "◂ current") {
		t.Fatalf("current session not marked: %q", tree)
	}
}

func TestCloneSwitchesLiveSession(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	s := &tuiSession{ctx: context.Background(), sessionID: "root", history: convo(),
		cfg: tuiConfig{provider: "groq", model: "x"}}

	out := s.cloneAction(context.Background())
	if !strings.Contains(out, "cloned") {
		t.Fatalf("clone status: %q", out)
	}
	if s.sessionID == "root" || s.sessionID == "" {
		t.Fatalf("/clone must move to a fresh session id, got %q", s.sessionID)
	}
	sessions, err := app.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 { // the frozen original + the clone
		t.Fatalf("want original + clone, got %d", len(sessions))
	}
	for _, st := range sessions {
		if st.ID == s.sessionID && st.Parent != "root" {
			t.Fatalf("clone parent = %q, want root", st.Parent)
		}
	}
}

func TestBranchCloneEmptyHistory(t *testing.T) {
	s := &tuiSession{ctx: context.Background(), sessionID: "root"}
	if got := s.branchAction(context.Background()); !strings.Contains(got, "nothing to branch") {
		t.Fatalf("empty branch: %q", got)
	}
	if got := s.cloneAction(context.Background()); !strings.Contains(got, "nothing to clone") {
		t.Fatalf("empty clone: %q", got)
	}
}

func TestNoteFoldsIntoNextRun(t *testing.T) {
	s := &tuiSession{}
	if got := s.noteAction(context.Background(), "  "); !strings.Contains(got, "empty") {
		t.Fatalf("blank note: %q", got)
	}
	s.noteAction(context.Background(), "prefer table-driven tests")
	s.noteAction(context.Background(), "keep functions small")
	if len(s.pendingNotes) != 2 {
		t.Fatalf("pendingNotes = %v", s.pendingNotes)
	}
	folded := s.foldPendingNotes("write the parser")
	if !strings.Contains(folded, "prefer table-driven tests") ||
		!strings.Contains(folded, "keep functions small") ||
		!strings.Contains(folded, "write the parser") {
		t.Fatalf("folded goal missing content: %q", folded)
	}
	if len(s.pendingNotes) != 0 {
		t.Fatal("pending notes should clear after folding")
	}
	// With nothing queued the goal is returned verbatim.
	if got := s.foldPendingNotes("plain"); got != "plain" {
		t.Fatalf("verbatim goal changed: %q", got)
	}
}

func TestSessionsListResumeRemove(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	// Seed a saved session to resume into.
	saved := app.NewSessionID()
	if err := app.SaveSession(app.StoredSession{ID: saved, Title: "old work",
		Provider: "groq", Model: "x", UpdatedAt: nowRFC(), Messages: convo()}); err != nil {
		t.Fatal(err)
	}
	s := &tuiSession{ctx: context.Background(), sessionID: "current", history: convo(),
		cfg: tuiConfig{provider: "groq", model: "x"}}

	if list := s.sessionsAction(context.Background(), "list"); !strings.Contains(list, saved) {
		t.Fatalf("list missing saved session: %q", list)
	}
	// Resume swaps the live conversation and returns entries to re-seed the transcript.
	entries, ok, status := s.resumeSessionAction(context.Background(), saved)
	if !ok || s.sessionID != saved || len(entries) == 0 || !strings.Contains(status, "resumed") {
		t.Fatalf("resume: ok=%v id=%q entries=%d status=%q", ok, s.sessionID, len(entries), status)
	}
	// Unknown id fails cleanly without changing the live session.
	if _, ok, st := s.resumeSessionAction(context.Background(), "nope"); ok || !strings.Contains(st, "nope") {
		t.Fatalf("resume unknown: ok=%v status=%q", ok, st)
	}
	if s.sessionID != saved {
		t.Fatalf("failed resume must not move the session, got %q", s.sessionID)
	}
	// Removing the current session is refused; removing another succeeds.
	if got := s.sessionsAction(context.Background(), "rm "+saved); !strings.Contains(got, "you're in") {
		t.Fatalf("rm current should be refused: %q", got)
	}
	other := app.NewSessionID()
	_ = app.SaveSession(app.StoredSession{ID: other, UpdatedAt: nowRFC(), Messages: convo()})
	if got := s.sessionsAction(context.Background(), "rm "+other); !strings.Contains(got, "removed") {
		t.Fatalf("rm other: %q", got)
	}
}

func TestSessionsEmptyAndUsage(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	s := &tuiSession{ctx: context.Background(), sessionID: "cur"}
	if got := s.sessionsAction(context.Background(), "list"); !strings.Contains(got, "no saved sessions") {
		t.Fatalf("empty list: %q", got)
	}
	if got := s.sessionTreeAction(context.Background()); !strings.Contains(got, "no saved sessions") {
		t.Fatalf("empty tree: %q", got)
	}
	if got := s.sessionsAction(context.Background(), "rm"); !strings.Contains(got, "usage") {
		t.Fatalf("rm usage: %q", got)
	}
	if got := s.sessionsAction(context.Background(), "frob"); !strings.Contains(got, "usage") {
		t.Fatalf("unknown sub: %q", got)
	}
}

func TestShortStamp(t *testing.T) {
	if got := shortStamp("2026-07-15T09:30:00Z"); !strings.Contains(got, "2026-07-15") {
		t.Fatalf("shortStamp = %q", got)
	}
	if got := shortStamp("garbage"); got != "garbage" {
		t.Fatalf("shortStamp fallback = %q", got)
	}
}
