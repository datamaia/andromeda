package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/core"
)

func writeSkill(t *testing.T, manifest, prompt string) string {
	t.Helper()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ManifestFile), []byte(manifest), 0o600)
	if prompt != "" {
		os.WriteFile(filepath.Join(dir, "prompt.md"), []byte(prompt), 0o600)
	}
	return dir
}

func TestLoadValidSkill(t *testing.T) {
	dir := writeSkill(t, `
name = "refactor"
version = "1.0.0"
description = "Refactor code safely"
prompt = "prompt.md"
required_tools = ["fs_read", "fs_write"]
required_capabilities = ["tool_calling"]
`, "You refactor code carefully.")
	s, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Manifest.Name != "refactor" || s.Prompt != "You refactor code carefully." {
		t.Fatalf("skill = %+v", s)
	}
}

func TestLoadRejectsIncompleteManifest(t *testing.T) {
	dir := writeSkill(t, `description = "no name or version"`, "")
	if _, err := Load(dir); err == nil {
		t.Fatal("expected an error for a manifest without name/version")
	}
}

func TestResolveSatisfied(t *testing.T) {
	dir := writeSkill(t, `
name = "s"
version = "1"
prompt = "prompt.md"
required_tools = ["fs_read"]
required_capabilities = ["tool_calling"]
`, "system prompt here")
	s, _ := Load(dir)
	res := Resolve(s, []string{"fs_read", "fs_write"}, core.Capabilities{core.CapToolCalling, core.CapChat})
	if !res.OK || res.SystemPrompt != "system prompt here" {
		t.Fatalf("resolution = %+v", res)
	}
}

func TestResolveReportsMissing(t *testing.T) {
	dir := writeSkill(t, `
name = "s"
version = "1"
required_tools = ["fs_write", "http_get"]
required_capabilities = ["vision"]
`, "")
	s, _ := Load(dir)
	res := Resolve(s, []string{"fs_read"}, core.Capabilities{core.CapChat})
	if res.OK {
		t.Fatal("resolution should not be OK")
	}
	if len(res.MissingTools) != 2 || len(res.MissingCaps) != 1 {
		t.Fatalf("missing = tools %v caps %v", res.MissingTools, res.MissingCaps)
	}
}

func TestLoadMissingManifest(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestLoadDirMarkdownSkill(t *testing.T) {
	dir := t.TempDir()
	md := "---\nname: review\ndescription: Review a diff\nversion: 2.0.0\ntools: [fs_read, fs_diff]\n---\nYou review diffs carefully."
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(md), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Manifest.Name != "review" || s.Manifest.Version != "2.0.0" || s.Manifest.Description != "Review a diff" {
		t.Fatalf("manifest = %+v", s.Manifest)
	}
	if s.Prompt != "You review diffs carefully." {
		t.Fatalf("prompt = %q", s.Prompt)
	}
	if len(s.Manifest.RequiredTools) != 2 {
		t.Fatalf("tools = %v", s.Manifest.RequiredTools)
	}
}

func TestLoadDirMarkdownNameFallsBackToDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-skill")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# Do the thing\nStep one, step two."), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Manifest.Name != "my-skill" {
		t.Fatalf("name = %q, want dir fallback", s.Manifest.Name)
	}
	if s.Manifest.Description != "Do the thing" { // synthesized from the first body line
		t.Fatalf("description = %q", s.Manifest.Description)
	}
}

func TestLoadDirFallsBackToToml(t *testing.T) {
	dir := writeSkill(t, "name = \"t\"\nversion = \"1.0.0\"\n", "")
	s, err := LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if s.Manifest.Name != "t" {
		t.Fatalf("name = %q", s.Manifest.Name)
	}
}

func TestDiscoverAcrossDirsAndDedup(t *testing.T) {
	root := t.TempDir()
	write := func(base, name, body string) {
		dir := filepath.Join(root, base, "skills", name)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(".agents", "alpha", "---\nname: alpha\ndescription: from agents\n---\nA")
	write(".claude", "beta", "---\nname: beta\ndescription: from claude\n---\nB")
	write(".claude", "alpha-dup", "---\nname: alpha\ndescription: dup, should lose\n---\nA2")

	got := Discover(root)
	if len(got) != 2 {
		t.Fatalf("discovered %d skills, want 2: %+v", len(got), got)
	}
	byName := map[string]Discovered{}
	for _, d := range got {
		byName[d.Manifest.Name] = d
	}
	if byName["alpha"].Source != ".agents" {
		t.Fatalf("alpha source = %q, want .agents (precedence)", byName["alpha"].Source)
	}
	if byName["beta"].Source != ".claude" {
		t.Fatalf("beta source = %q, want .claude", byName["beta"].Source)
	}
}
