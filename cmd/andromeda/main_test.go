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

func TestRootRunHelpSucceeds(t *testing.T) {
	if code := run([]string{}); code != exitOK {
		t.Fatalf("expected exit %d for bare invocation, got %d", exitOK, code)
	}
}
