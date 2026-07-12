package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	root := newRootCommand()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if got := out.String(); !strings.HasPrefix(got, "andromeda ") {
		t.Fatalf("unexpected version output: %q", got)
	}
}

func TestRunUnknownCommandIsUsageError(t *testing.T) {
	if code := run([]string{"definitely-not-a-command"}); code != exitUsage {
		t.Fatalf("expected exit %d for unknown command, got %d", exitUsage, code)
	}
}

// FR-CLI-003: a bare invocation in a non-interactive context (go test never has a TTY;
// TERM=dumb makes that explicit) is a usage error (exit 2), not a TUI launch or help/exit 0.
func TestBareNonInteractiveExitsUsage(t *testing.T) {
	t.Setenv("TERM", "dumb")
	if code := run([]string{}); code != exitUsage {
		t.Fatalf("expected exit %d for bare non-interactive invocation, got %d", exitUsage, code)
	}
}
