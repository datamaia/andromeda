package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/tui"
	"github.com/spf13/cobra"
)

func TestMapChoice(t *testing.T) {
	cases := []struct {
		in   tui.ApprovalChoice
		out  core.DecisionOutcome
		kind core.PermissionDecisionKind
	}{
		{tui.ApproveOnce, core.OutcomeAllow, core.DecisionAllowOnce},
		{tui.ApproveSession, core.OutcomeAllow, core.DecisionAllowForSession},
		{tui.ApproveWorkspace, core.OutcomeAllow, core.DecisionAllowForWorkspace},
		{tui.AlwaysDeny, core.OutcomeDeny, core.DecisionAlwaysDeny},
		{tui.RejectOnce, core.OutcomeDeny, core.DecisionDenyOnce},
	}
	for _, c := range cases {
		gotOut, gotKind := mapChoice(c.in)
		if gotOut != c.out || gotKind != c.kind {
			t.Errorf("mapChoice(%v) = (%v,%v), want (%v,%v)", c.in, gotOut, gotKind, c.out, c.kind)
		}
	}
}

func TestActionLabel(t *testing.T) {
	cases := map[core.Permission]string{
		core.PermWrite:            "write",
		core.PermExecute:          "run command",
		core.PermGitMutation:      "git mutation",
		core.PermProcessSpawn:     "spawn process",
		core.PermNetwork:          "network request",
		core.PermCredentialAccess: "read credential",
	}
	for p, want := range cases {
		if got := actionLabel(p); got != want {
			t.Errorf("actionLabel(%v) = %q, want %q", p, got, want)
		}
	}
	// An unknown permission falls back to its own string form.
	if got := actionLabel(core.Permission("custom")); got != "custom" {
		t.Errorf("actionLabel(custom) = %q", got)
	}
}

func TestSubjectLabelAndApprovalDetail(t *testing.T) {
	if got := subjectLabel(""); got != "(unspecified)" {
		t.Errorf("empty subject = %q", got)
	}
	if got := subjectLabel("/tmp/x"); got != "/tmp/x" {
		t.Errorf("subject = %q", got)
	}
	detail := approvalDetail(ports.PermissionQuery{Permission: core.PermWrite, Subject: "/tmp/x"})
	if !strings.Contains(detail, "write") || !strings.Contains(detail, "/tmp/x") {
		t.Errorf("approvalDetail = %q", detail)
	}
}

func TestMessageTextAndHistoryEntries(t *testing.T) {
	msgs := []ports.Message{
		{Role: "system", Parts: []ports.ContentPart{{Type: "text", Text: "sys"}}},
		{Role: "user", Parts: []ports.ContentPart{{Type: "text", Text: "hello"}, {Type: "text", Text: " there"}}},
		{Role: "assistant", Parts: []ports.ContentPart{{Text: "hi"}}},               // empty Type counts as text
		{Role: "user", Parts: []ports.ContentPart{{Type: "text", Text: "   "}}},     // whitespace-only is dropped
		{Role: "tool", Parts: []ports.ContentPart{{Type: "text", Text: "ignored"}}}, // non-user/assistant dropped
	}
	if got := messageText(msgs[1]); got != "hello there" {
		t.Errorf("messageText = %q", got)
	}
	entries := historyEntries(msgs)
	if len(entries) != 2 {
		t.Fatalf("historyEntries = %d, want 2: %+v", len(entries), entries)
	}
	if entries[0].Role != "user" || entries[0].Text != "hello there" {
		t.Errorf("entry[0] = %+v", entries[0])
	}
	if entries[1].Role != "agent" || entries[1].Text != "hi" {
		t.Errorf("entry[1] = %+v", entries[1])
	}
}

func TestFormatMemory(t *testing.T) {
	if got := formatMemory(nil, errors.New("boom")); !strings.Contains(got, "boom") {
		t.Errorf("error case = %q", got)
	}
	if got := formatMemory(nil, nil); !strings.Contains(got, "no memories") {
		t.Errorf("empty case = %q", got)
	}
	recs := []ports.MemoryRecord{{Layer: "project", Content: "remember this"}}
	if got := formatMemory(recs, nil); !strings.Contains(got, "project") || !strings.Contains(got, "remember this") {
		t.Errorf("records case = %q", got)
	}
}

// listSessions writes a listing (or an empty-state note) without error. XDG_DATA_HOME is pointed at
// a temp dir so the test never touches or pollutes the real sessions store.
func TestListSessionsRuns(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	if err := listSessions(cmd); err != nil {
		t.Fatalf("listSessions: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("listSessions wrote nothing")
	}
}

// listFiles falls back to a bounded directory walk outside a git repo, returning workspace-relative
// paths and skipping dependency directories like node_modules.
func TestListFilesWalk(t *testing.T) {
	tmp := t.TempDir()
	must := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	must(os.WriteFile(filepath.Join(tmp, "a.go"), []byte("x"), 0o600))
	must(os.MkdirAll(filepath.Join(tmp, "sub"), 0o750))
	must(os.WriteFile(filepath.Join(tmp, "sub", "b.txt"), []byte("y"), 0o600))
	must(os.MkdirAll(filepath.Join(tmp, "node_modules", "dep"), 0o750))
	must(os.WriteFile(filepath.Join(tmp, "node_modules", "dep", "skip.js"), []byte("z"), 0o600))

	s := &tuiSession{wd: tmp}
	got := strings.Join(s.listFiles(context.Background()), "\n")
	if !strings.Contains(got, "a.go") || !strings.Contains(got, filepath.Join("sub", "b.txt")) {
		t.Errorf("expected a.go and sub/b.txt, got: %q", got)
	}
	if strings.Contains(got, "skip.js") {
		t.Errorf("node_modules should be skipped, got: %q", got)
	}
}

// gitBranch is empty outside a repo and a branch name inside one; contextAction reports the
// workspace, branch, and change count.
func TestGitBranchAndContextAction(t *testing.T) {
	tmp := t.TempDir()
	if got := gitBranch(context.Background(), tmp); got != "" {
		t.Errorf("non-repo branch = %q, want empty", got)
	}
	git := func(args ...string) error {
		c := exec.Command("git", append([]string{"-C", tmp}, args...)...)
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@e", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@e")
		return c.Run()
	}
	if err := git("init", "-q"); err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	_ = git("config", "user.email", "t@e")
	_ = git("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(tmp, "f.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	_ = git("add", ".")
	_ = git("commit", "-q", "-m", "init")

	if got := gitBranch(context.Background(), tmp); got == "" {
		t.Error("expected a branch name inside a git repo")
	}
	lines := strings.Join((&tuiSession{wd: tmp}).contextAction(context.Background()), "\n")
	for _, want := range []string{"workspace", "branch", "changes"} {
		if !strings.Contains(lines, want) {
			t.Errorf("contextAction missing %q: %q", want, lines)
		}
	}
}

// detectCommands maps each recognized ecosystem to its commands.
func TestDetectCommandsEcosystems(t *testing.T) {
	cases := map[string]string{"go.mod": "go test", "package.json": "npm test", "pyproject.toml": "pytest"}
	for marker, wantTest := range cases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, marker), []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, test, _ := detectCommands(dir); !strings.Contains(test, wantTest) {
			t.Errorf("%s → test %q, want %q", marker, test, wantTest)
		}
	}
}

// providerChoices exposes the catalog as TUI menu entries; it must be non-empty and carry ids.
func TestProviderChoices(t *testing.T) {
	choices := providerChoices()
	if len(choices) == 0 {
		t.Fatal("providerChoices should not be empty")
	}
	for _, c := range choices {
		if c.ID == "" {
			t.Errorf("provider choice missing id: %+v", c)
		}
	}
}
