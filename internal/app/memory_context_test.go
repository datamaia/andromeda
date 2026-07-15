package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectMemoryAndFolding(t *testing.T) {
	root := t.TempDir()
	if projectMemory(root) != "" {
		t.Error("absent memory index should yield empty")
	}
	dir := filepath.Join(root, ".andromeda", "memory")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "MEMORY.md"), []byte("# Workspace memory\n\n- [Fact](0001-fact.md) — 0001"), 0o600); err != nil {
		t.Fatal(err)
	}
	mem := projectMemory(root)
	if !strings.Contains(mem, "Fact") {
		t.Fatalf("projectMemory = %q", mem)
	}
	sys := withMemory("BASE", mem)
	if !strings.HasPrefix(sys, "BASE") || !strings.Contains(sys, "Workspace memory index") || !strings.Contains(sys, "Fact") {
		t.Errorf("withMemory = %q", sys)
	}
	if withMemory("BASE", "") != "BASE" {
		t.Error("empty memory should not alter the base prompt")
	}
}
