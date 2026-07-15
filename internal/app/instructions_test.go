package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectMapsDetectsArtifacts(t *testing.T) {
	dir := t.TempDir()
	if got := projectMaps(dir); got != "" {
		t.Fatalf("no maps should yield empty, got %q", got)
	}
	// Create the ontology map.
	ttl := filepath.Join(dir, ".andromeda", "ontology")
	if err := os.MkdirAll(ttl, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ttl, "project.ttl"), []byte("@prefix x: <y> ."), 0o600); err != nil {
		t.Fatal(err)
	}
	got := projectMaps(dir)
	if !strings.Contains(got, "project.ttl") {
		t.Fatalf("ontology should be detected, got %q", got)
	}
	// withMaps folds it into the system prompt with the navigation instruction.
	sys := withMaps("BASE", got)
	if !strings.Contains(sys, "BASE") || !strings.Contains(sys, "navigate the") || !strings.Contains(sys, "project.ttl") {
		t.Fatalf("withMaps output = %q", sys)
	}
	if withMaps("BASE", "") != "BASE" {
		t.Fatal("empty maps should not alter the base prompt")
	}
}

func TestProjectInstructionsReadsAgentsFile(t *testing.T) {
	dir := t.TempDir()
	if got := projectInstructions(dir); got != "" {
		t.Fatalf("no AGENTS.md should yield empty, got %q", got)
	}
	body := "# AGENTS.md\n\nAlways run gofmt.\n"
	if err := os.WriteFile(filepath.Join(dir, AgentsFileName), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	got := projectInstructions(dir)
	if !strings.Contains(got, "Always run gofmt.") {
		t.Fatalf("instructions = %q", got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Errorf("instructions should be trimmed, got %q", got)
	}
}

func TestProjectInstructionsCapsSize(t *testing.T) {
	dir := t.TempDir()
	big := strings.Repeat("x", maxInstructionsBytes+5000)
	if err := os.WriteFile(filepath.Join(dir, AgentsFileName), []byte(big), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := len(projectInstructions(dir)); got > maxInstructionsBytes {
		t.Fatalf("instructions not capped: %d bytes", got)
	}
}

func TestComposeSystem(t *testing.T) {
	// No instructions leaves the base untouched.
	if got := composeSystem("BASE", ""); got != "BASE" {
		t.Errorf("empty instructions changed base: %q", got)
	}
	// Base first, then a labeled AGENTS.md block.
	got := composeSystem("BASE", "do the thing")
	if !strings.HasPrefix(got, "BASE") || !strings.Contains(got, AgentsFileName) || !strings.Contains(got, "do the thing") {
		t.Errorf("composed system = %q", got)
	}
	// An empty base yields just the instructions block (no leading blank identity).
	if got := composeSystem("", "only project"); strings.HasPrefix(got, "\n") {
		t.Errorf("empty base should not lead with a newline: %q", got)
	}
}
