package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// A fresh workspace gets the full stack: AGENTS.md, a complete andromeda.toml, and the .agents/ and
// .andromeda/ trees.
func TestScaffoldProjectCreatesStack(t *testing.T) {
	wd := t.TempDir()
	// Marker file so the AGENTS.md build/test commands are pre-filled for Go.
	if err := os.WriteFile(filepath.Join(wd, "go.mod"), []byte("module x\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	summary := scaffoldProject(wd, "openai-chatgpt", "gpt-5.5")

	for _, p := range []string{
		"AGENTS.md", "andromeda.toml",
		".agents/README.md", ".agents/skills", ".agents/commands", ".agents/mcp",
		".andromeda/README.md", ".andromeda/memory",
	} {
		if _, err := os.Stat(filepath.Join(wd, p)); err != nil {
			t.Errorf("expected %s to exist: %v", p, err)
		}
	}

	agents := read(t, filepath.Join(wd, "AGENTS.md"))
	if !strings.Contains(agents, "go test ./...") {
		t.Error("AGENTS.md should pre-fill Go commands when go.mod is present")
	}
	if !strings.Contains(agents, filepath.Base(wd)) {
		t.Error("AGENTS.md should mention the project name")
	}

	toml := read(t, filepath.Join(wd, "andromeda.toml"))
	for _, want := range []string{`default = "openai-chatgpt"`, `model   = "gpt-5.5"`, "[permission]", "[mcp]", "[plugins]", "allow = ["} {
		if !strings.Contains(toml, want) {
			t.Errorf("andromeda.toml missing %q", want)
		}
	}
	if !strings.Contains(summary, "AGENTS.md") || !strings.Contains(summary, "created") {
		t.Errorf("summary should report what was created: %q", summary)
	}
}

// Re-running on a complete workspace is a no-op: nothing is recreated or clobbered.
func TestScaffoldProjectIdempotent(t *testing.T) {
	wd := t.TempDir()
	scaffoldProject(wd, "p", "m")
	// Personalize AGENTS.md; a second run must not overwrite it.
	custom := "# my own AGENTS\n\ncustom guidance\n"
	if err := os.WriteFile(filepath.Join(wd, "AGENTS.md"), []byte(custom), 0o600); err != nil {
		t.Fatal(err)
	}
	summary := scaffoldProject(wd, "p", "m")
	if read(t, filepath.Join(wd, "AGENTS.md")) != custom {
		t.Error("second /init clobbered a user-edited AGENTS.md")
	}
	if !strings.Contains(summary, "kept") {
		t.Errorf("second run should report kept files: %q", summary)
	}
}

// An andromeda.toml predating the new sections is augmented in place — new sections appended, the
// user's existing values preserved verbatim.
func TestScaffoldProjectUpgradesExistingTOML(t *testing.T) {
	wd := t.TempDir()
	legacy := "[provider]\ndefault = \"ollama\"\nmodel   = \"llama3\"\n\n[agent]\nmax_iterations = 7\n"
	if err := os.WriteFile(filepath.Join(wd, "andromeda.toml"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}
	summary := scaffoldProject(wd, "p", "m")

	toml := read(t, filepath.Join(wd, "andromeda.toml"))
	// User content preserved.
	if !strings.Contains(toml, `default = "ollama"`) || !strings.Contains(toml, "max_iterations = 7") {
		t.Error("upgrade clobbered existing values")
	}
	// Missing sections appended.
	for _, want := range []string{"[permission]", "[mcp]", "[plugins]"} {
		if !strings.Contains(toml, want) {
			t.Errorf("upgrade did not append %q", want)
		}
	}
	// Exactly one [provider] header (no duplication).
	if n := strings.Count(toml, "[provider]"); n != 1 {
		t.Errorf("provider section duplicated: %d", n)
	}
	if !strings.Contains(summary, "updated") {
		t.Errorf("summary should report the toml upgrade: %q", summary)
	}
}

func TestDetectCommands(t *testing.T) {
	wd := t.TempDir()
	// Unknown stack → placeholders.
	if b, _, _ := detectCommands(wd); !strings.Contains(b, "build command") {
		t.Errorf("unknown stack should yield a placeholder, got %q", b)
	}
	if err := os.WriteFile(filepath.Join(wd, "Cargo.toml"), []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, test, _ := detectCommands(wd); !strings.Contains(test, "cargo test") {
		t.Errorf("Cargo.toml should map to cargo test, got %q", test)
	}
}
