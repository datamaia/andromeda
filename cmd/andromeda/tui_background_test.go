package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/storage"
)

func TestBackgroundLaunches(t *testing.T) {
	wd := t.TempDir()
	s := &tuiSession{ctx: context.Background(), wd: wd, cfg: tuiConfig{provider: "groq", model: "small"}}

	// Stub the spawn so no real process starts (the test binary must never be re-executed).
	var launched *exec.Cmd
	orig := startBackground
	startBackground = func(cmd *exec.Cmd) (int, error) { launched = cmd; return 4321, nil }
	defer func() { startBackground = orig }()

	// Usage on empty goal.
	if got := s.backgroundAction(context.Background(), ""); !strings.Contains(got, "usage") {
		t.Fatalf("empty: %q", got)
	}
	// Default grants write only.
	out := s.backgroundAction(context.Background(), "add a test")
	if !strings.Contains(out, "started background task") || !strings.Contains(out, "grants: write") ||
		strings.Contains(out, "exec") {
		t.Fatalf("default background: %q", out)
	}
	args := strings.Join(launched.Args, " ")
	if !strings.Contains(args, "run add a test") || !strings.Contains(args, "--allow-write") ||
		strings.Contains(args, "--allow-exec") || !strings.Contains(args, "--model small") {
		t.Fatalf("launch args: %v", launched.Args)
	}
	// A log file is created under the marker dir.
	logDir := filepath.Join(wd, storage.MarkerDir, "background")
	entries, _ := os.ReadDir(logDir)
	if len(entries) == 0 {
		t.Fatal("expected a background log file")
	}

	// --exec opts into command execution.
	out = s.backgroundAction(context.Background(), "--exec deploy it")
	if !strings.Contains(out, "write + exec") {
		t.Fatalf("exec grant: %q", out)
	}
	if !strings.Contains(strings.Join(launched.Args, " "), "--allow-exec") {
		t.Fatalf("exec args: %v", launched.Args)
	}
}

func TestAutofixPRDiagnoses(t *testing.T) {
	s := &tuiSession{ctx: context.Background(), wd: t.TempDir()}
	orig := ghRunner
	defer func() { ghRunner = orig }()

	// Failing checks on an explicit PR number → a fix goal is returned.
	ghRunner = func(_ context.Context, _ string, args ...string) (string, error) {
		if strings.Contains(strings.Join(args, " "), "statusCheckRollup") {
			return "lint\nunit-tests\n", nil
		}
		return "", nil
	}
	goal, status := s.autofixPRAction(context.Background(), "42")
	if !strings.Contains(goal, "lint") || !strings.Contains(goal, "unit-tests") ||
		!strings.Contains(goal, "#42") || !strings.Contains(goal, "push") {
		t.Fatalf("goal: %q", goal)
	}
	if !strings.Contains(status, "2 failing checks") || !strings.Contains(status, "#42") {
		t.Fatalf("status: %q", status)
	}

	// Green CI → no goal.
	ghRunner = func(context.Context, string, ...string) (string, error) { return "", nil }
	goal, status = s.autofixPRAction(context.Background(), "42")
	if goal != "" || !strings.Contains(status, "green") {
		t.Fatalf("green: goal=%q status=%q", goal, status)
	}

	// No PR number and none for the branch.
	ghRunner = func(context.Context, string, ...string) (string, error) {
		return "", &exec.ExitError{} // gh pr view --json number fails
	}
	goal, status = s.autofixPRAction(context.Background(), "")
	if goal != "" || !strings.Contains(status, "no PR") {
		t.Fatalf("no-pr: goal=%q status=%q", goal, status)
	}

	// No gh installed.
	ghRunner = func(context.Context, string, ...string) (string, error) { return "", errNoGH }
	if _, status = s.autofixPRAction(context.Background(), "42"); !strings.Contains(status, "GitHub CLI") {
		t.Fatalf("no-gh: %q", status)
	}
}

func TestPluralHelper(t *testing.T) {
	if plural(1, "check") != "1 check" || plural(3, "check") != "3 checks" {
		t.Fatalf("plural: %q / %q", plural(1, "check"), plural(3, "check"))
	}
}
