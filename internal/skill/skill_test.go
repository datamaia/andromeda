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
