package builtin

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// runTool drains a tool's event stream and returns the terminal outcome and its text payload.
func runTool(t *testing.T, tool ports.ToolPort, input string) (outcome, text string) {
	t.Helper()
	st, err := tool.Execute(context.Background(), ports.ToolExecuteRequest{Input: ports.JSON(input)})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatalf("stream: %v", err)
		}
		if ev.Terminal {
			outcome, text = ev.Outcome, ev.Text
		}
	}
	return outcome, text
}

func TestFSReplaceUnique(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	os.WriteFile(p, []byte("alpha beta gamma\n"), 0o600)

	outcome, text := runTool(t, FSReplace{}, `{"path":"`+p+`","old":"beta","new":"BETA"}`)
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	var res struct {
		Replacements int    `json:"replacements"`
		AfterHash    string `json:"after_hash"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Replacements != 1 || len(res.AfterHash) != 64 {
		t.Fatalf("result = %+v", res)
	}
	data, _ := os.ReadFile(p)
	if string(data) != "alpha BETA gamma\n" {
		t.Fatalf("file = %q", data)
	}
}

func TestFSReplaceNonUniqueRejected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	os.WriteFile(p, []byte("x x x\n"), 0o600)

	if outcome, _ := runTool(t, FSReplace{}, `{"path":"`+p+`","old":"x","new":"y"}`); outcome != "error" {
		t.Fatal("expected non-unique match to be rejected")
	}
	// replace_all succeeds and reports the count.
	outcome, text := runTool(t, FSReplace{}, `{"path":"`+p+`","old":"x","new":"y","replace_all":true}`)
	if outcome != "success" {
		t.Fatalf("replace_all failed: %s", text)
	}
	data, _ := os.ReadFile(p)
	if string(data) != "y y y\n" {
		t.Fatalf("file = %q", data)
	}
}

func TestFSReplaceRegex(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	os.WriteFile(p, []byte("id=42 id=7\n"), 0o600)
	outcome, _ := runTool(t, FSReplace{}, `{"path":"`+p+`","old":"id=[0-9]+","new":"id=X","replace_all":true,"regex":true}`)
	if outcome != "success" {
		t.Fatal("regex replace_all should succeed")
	}
	data, _ := os.ReadFile(p)
	if string(data) != "id=X id=X\n" {
		t.Fatalf("file = %q", data)
	}
}

func TestFSDiffThenPatchRoundTrip(t *testing.T) {
	dir := t.TempDir()
	left := filepath.Join(dir, "code.txt")
	original := "line1\nline2\nline3\nline4\n"
	target := "line1\nCHANGED2\nline3\nline4\nline5\n"
	os.WriteFile(left, []byte(original), 0o600)

	// Compute the diff between the file and the target content.
	_, diffText := runTool(t, FSDiff{}, mustJSON(map[string]any{"left": left, "right": "code.txt", "right_content": target}))
	var diffRes struct {
		Diff   string `json:"diff"`
		Binary bool   `json:"binary"`
	}
	_ = json.Unmarshal([]byte(diffText), &diffRes)
	if diffRes.Binary || diffRes.Diff == "" {
		t.Fatalf("unexpected diff result: %+v", diffRes)
	}

	// check_only must not modify the file.
	outcome, _ := runTool(t, FSPatch{}, mustJSON(map[string]any{"diff": diffRes.Diff, "root": dir, "check_only": true}))
	if outcome != "success" {
		t.Fatal("check_only should succeed")
	}
	if data, _ := os.ReadFile(left); string(data) != original {
		t.Fatal("check_only modified the file")
	}

	// Applying the diff reproduces the target exactly (round-trip property).
	outcome, patchText := runTool(t, FSPatch{}, mustJSON(map[string]any{"diff": diffRes.Diff, "root": dir}))
	if outcome != "success" {
		t.Fatalf("patch failed: %s", patchText)
	}
	got, _ := os.ReadFile(left)
	if string(got) != target {
		t.Fatalf("round-trip mismatch:\n got: %q\nwant: %q", got, target)
	}
}

func TestFSPatchContextMismatchRejected(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "code.txt")
	os.WriteFile(p, []byte("totally\ndifferent\ncontent\n"), 0o600)
	// A hunk whose context does not match the file must be rejected atomically.
	diff := "--- a/code.txt\n+++ b/code.txt\n@@ -1,2 +1,2 @@\n line1\n-line2\n+CHANGED\n"
	outcome, text := runTool(t, FSPatch{}, mustJSON(map[string]any{"diff": diff, "root": dir}))
	if outcome != "error" {
		t.Fatalf("expected rejection, got %s", text)
	}
	if data, _ := os.ReadFile(p); string(data) != "totally\ndifferent\ncontent\n" {
		t.Fatal("file was modified despite rejection")
	}
}

func TestFSDiffEqualFilesEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.txt")
	os.WriteFile(p, []byte("same\n"), 0o600)
	_, text := runTool(t, FSDiff{}, mustJSON(map[string]any{"left": p, "right": "a.txt", "right_content": "same\n"}))
	var res struct {
		Diff string `json:"diff"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Diff != "" {
		t.Fatalf("equal files should produce an empty diff, got %q", res.Diff)
	}
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
