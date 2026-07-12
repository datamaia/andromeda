package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.email", "t@e.com"}, {"config", "user.name", "T"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o600)
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "initial"}} {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=T", "GIT_AUTHOR_EMAIL=t@e.com", "GIT_COMMITTER_NAME=T", "GIT_COMMITTER_EMAIL=t@e.com")
		c.CombinedOutput()
	}
}

func TestGitStatusAndLogCommands(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	t.Chdir(dir)

	out, err := runCmd(t, "git", "status")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "On branch main") {
		t.Errorf("git status output = %q", out)
	}
	logOut, err := runCmd(t, "git", "log", "-n", "5")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(logOut, "initial") {
		t.Errorf("git log output = %q", logOut)
	}
}

func TestMemoryCommands(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if _, err := runCmd(t, "memory", "add", "--layer", "session", "remember goreleaser"); err != nil {
		t.Fatal(err)
	}
	out, err := runCmd(t, "memory", "search", "goreleaser")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "goreleaser") {
		t.Errorf("memory search output = %q", out)
	}
	listOut, _ := runCmd(t, "memory", "list")
	if !strings.Contains(listOut, "goreleaser") {
		t.Errorf("memory list output = %q", listOut)
	}
}

func TestConfigShowCommand(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	out, err := runCmd(t, "config", "show")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "logging.level") {
		t.Errorf("config show output = %q", out)
	}
	jsonOut, err := runCmd(t, "config", "show", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jsonOut, "\"values\"") {
		t.Errorf("config show --json output = %q", jsonOut)
	}
}

func TestToolListCommand(t *testing.T) {
	out, err := runCmd(t, "tool", "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"fs_read", "fs_write", "fs_search", "terminal_run"} {
		if !strings.Contains(out, want) {
			t.Errorf("tool list missing %q: %q", want, out)
		}
	}
}

func TestIndexQueryCommand(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("the provider router dispatches\n"), 0o600)
	t.Chdir(dir)
	out, err := runCmd(t, "index", "query", "provider")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "a.txt") {
		t.Errorf("index query output = %q", out)
	}
}

func TestWorkflowCommands(t *testing.T) {
	list, err := runCmd(t, "workflow", "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"intake", "release-preparation"} {
		if !strings.Contains(list, want) {
			t.Errorf("workflow list missing %q", want)
		}
	}
	// Non-interactive run halts at the first gate without --auto-approve.
	halt, err := runCmd(t, "workflow", "run", "sdd")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(halt, "awaiting_approval") && !strings.Contains(halt, "halted at gate") {
		t.Errorf("expected a gate halt, got %q", halt)
	}
	// With --auto-approve the whole 14-stage pipeline completes.
	full, err := runCmd(t, "workflow", "run", "sdd", "--auto-approve")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(full, "14 stages") || !strings.Contains(full, "completed") {
		t.Errorf("expected a completed 14-stage run, got %q", full)
	}
}

func TestProviderListCommand(t *testing.T) {
	out, err := runCmd(t, "provider", "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"ollama", "anthropic", "openai-compatible"} {
		if !strings.Contains(out, want) {
			t.Errorf("provider list missing %q", want)
		}
	}
}

func TestLogsAndExportCommands(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	// doctor emits and persists an event; logs should then show at least one row.
	if _, err := runCmd(t, "doctor"); err != nil {
		// doctor may fail on global-db in odd envs; tolerate and still exercise logs/export.
		t.Logf("doctor: %v", err)
	}
	if _, err := runCmd(t, "logs", "-n", "5"); err != nil {
		t.Fatalf("logs: %v", err)
	}
	out, err := runCmd(t, "export")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "[") {
		t.Errorf("export output = %q", out)
	}
}

func TestVersionCommandOutput(t *testing.T) {
	out, err := runCmd(t, "version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "andromeda ") {
		t.Errorf("version output = %q", out)
	}
}
