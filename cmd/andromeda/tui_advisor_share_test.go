package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/settingstore"
)

func TestAdvisorConsults(t *testing.T) {
	s := &tuiSession{ctx: context.Background(), wd: t.TempDir(), cfg: tuiConfig{provider: "groq", model: "small"},
		prov: cannedProvider{reply: "consider a bounded queue"}}
	s.history = []ports.Message{textMessage("user", "how to rate-limit?")}

	out := s.advisorAction(context.Background(), "should I use a semaphore?")
	if !strings.Contains(out, "advisor · small") || !strings.Contains(out, "bounded queue") {
		t.Fatalf("advisor consult: %q", out)
	}
	// The consultation must NOT pollute the conversation history.
	if len(s.history) != 1 {
		t.Fatalf("advisor should not touch history, got %d msgs", len(s.history))
	}
}

func TestAdvisorModelConfig(t *testing.T) {
	wd := t.TempDir()
	s := &tuiSession{ctx: context.Background(), wd: wd, cfg: tuiConfig{provider: "groq", model: "small"}}

	if got := s.advisorAction(context.Background(), ""); !strings.Contains(got, "usage") || !strings.Contains(got, "small") {
		t.Fatalf("usage: %q", got)
	}
	if got := s.advisorAction(context.Background(), "model big-brain"); !strings.Contains(got, "big-brain") {
		t.Fatalf("set model: %q", got)
	}
	if st, _ := settingstore.Load(wd); st.AdvisorModel != "big-brain" {
		t.Fatalf("advisor model not persisted: %+v", st)
	}
	if got := s.advisorModel(); got != "big-brain" {
		t.Fatalf("advisorModel = %q", got)
	}
	if got := s.advisorAction(context.Background(), "model"); !strings.Contains(got, "cleared") {
		t.Fatalf("clear model: %q", got)
	}
	if st, _ := settingstore.Load(wd); st.AdvisorModel != "" {
		t.Fatal("advisor model should be cleared")
	}
}

func TestAdvisorGoalIncludesContext(t *testing.T) {
	h := []ports.Message{textMessage("user", "build a cache"), textMessage("assistant", "ok")}
	g := advisorGoal(h, "is LRU right?")
	if !strings.Contains(g, "build a cache") || !strings.Contains(g, "is LRU right?") {
		t.Fatalf("advisor goal: %q", g)
	}
	if plain := advisorGoal(nil, "just this"); plain != "just this" {
		t.Fatalf("empty-history goal should be the bare question: %q", plain)
	}
}

func TestShareUploadsAndRemembers(t *testing.T) {
	wd := t.TempDir()
	s := &tuiSession{ctx: context.Background(), wd: wd}

	var gotArgs []string
	var gotStdin string
	orig := ghRunner
	ghRunner = func(_ context.Context, stdin string, args ...string) (string, error) {
		gotArgs, gotStdin = args, stdin
		return "https://gist.github.com/me/abc123", nil
	}
	defer func() { ghRunner = orig }()

	out := s.shareAction([]string{"you ▸ hi", "andromeda ▸ hello"})
	if !strings.Contains(out, "gist.github.com/me/abc123") || !strings.Contains(out, "/unshare") {
		t.Fatalf("share output: %q", out)
	}
	if !strings.Contains(gotStdin, "hello") || len(gotArgs) == 0 || gotArgs[0] != "gist" {
		t.Fatalf("gh not invoked correctly: args=%v", gotArgs)
	}
	// Secret by default: the args must NOT request a public gist.
	for _, a := range gotArgs {
		if a == "--public" {
			t.Fatal("share must create a secret gist, not public")
		}
	}
	if st, _ := settingstore.Load(wd); st.LastGist != "abc123" {
		t.Fatalf("gist id not remembered: %+v", st)
	}
}

func TestShareNoGH(t *testing.T) {
	s := &tuiSession{ctx: context.Background(), wd: t.TempDir()}
	orig := ghRunner
	ghRunner = func(context.Context, string, ...string) (string, error) { return "", errNoGH }
	defer func() { ghRunner = orig }()
	if got := s.shareAction([]string{"x"}); !strings.Contains(got, "GitHub CLI") {
		t.Fatalf("no-gh share: %q", got)
	}
}

func TestUnshareDeletes(t *testing.T) {
	wd := t.TempDir()
	s := &tuiSession{ctx: context.Background(), wd: wd}

	// Nothing shared yet.
	if got := s.unshareAction(context.Background()); !strings.Contains(got, "nothing shared") {
		t.Fatalf("empty unshare: %q", got)
	}
	// Seed a shared gist id, then delete it.
	_ = settingstore.Save(wd, settingstore.Settings{LastGist: "abc123"})
	var deleted string
	orig := ghRunner
	ghRunner = func(_ context.Context, _ string, args ...string) (string, error) {
		if len(args) >= 3 && args[0] == "gist" && args[1] == "delete" {
			deleted = args[2]
		}
		return "", nil
	}
	defer func() { ghRunner = orig }()

	if got := s.unshareAction(context.Background()); !strings.Contains(got, "deleted") {
		t.Fatalf("unshare: %q", got)
	}
	if deleted != "abc123" {
		t.Fatalf("gh delete got %q", deleted)
	}
	if st, _ := settingstore.Load(wd); st.LastGist != "" {
		t.Fatal("LastGist should be cleared after unshare")
	}
}

func TestGistID(t *testing.T) {
	if got := gistID("https://gist.github.com/me/abc123"); got != "abc123" {
		t.Fatalf("gistID = %q", got)
	}
	if got := gistID(""); got != "" {
		t.Fatalf("empty gistID = %q", got)
	}
}

func TestGHErrorSurfacesStderr(t *testing.T) {
	// A plain error passes through; an ExitError's stderr is surfaced.
	if got := ghError(errors.New("boom")); got.Error() != "boom" {
		t.Fatalf("ghError plain: %q", got)
	}
}
