package builtin

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/git"
)

func initBuiltinRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# hello\n"), 0o600)
	run("add", ".")
	run("commit", "-m", "initial commit")
	return dir
}

func TestGitExecStatusAndCommit(t *testing.T) {
	dir := initBuiltinRepo(t)
	tool := NewGitExec(git.New(""))

	// Read: status on a clean repo.
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"operation": "status", "repo": dir}))
	if outcome != "success" {
		t.Fatalf("status failed: %s", text)
	}

	// Mutation: create a file, stage it, and commit through git_exec.
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("data\n"), 0o600)
	if o, tx := runTool(t, tool, mustJSON(map[string]any{"operation": "stage", "repo": dir, "args": map[string]any{"paths": []string{"new.txt"}}})); o != "success" {
		t.Fatalf("stage failed: %s", tx)
	}
	outcome, text = runTool(t, tool, mustJSON(map[string]any{"operation": "commit", "repo": dir, "args": map[string]any{"message": "add new.txt"}}))
	if outcome != "success" {
		t.Fatalf("commit failed: %s", text)
	}
	var res struct {
		Result struct {
			Commit string `json:"commit"`
		} `json:"result"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Result.Commit == "" {
		t.Fatalf("commit id missing: %s", text)
	}
}

func TestGitExecBranchAndLog(t *testing.T) {
	dir := initBuiltinRepo(t)
	tool := NewGitExec(git.New(""))

	if o, tx := runTool(t, tool, mustJSON(map[string]any{"operation": "branch_create", "repo": dir, "args": map[string]any{"name": "feature"}})); o != "success" {
		t.Fatalf("branch_create failed: %s", tx)
	}
	if o, tx := runTool(t, tool, mustJSON(map[string]any{"operation": "branch_switch", "repo": dir, "args": map[string]any{"name": "feature"}})); o != "success" {
		t.Fatalf("branch_switch failed: %s", tx)
	}
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"operation": "log", "repo": dir, "args": map[string]any{"max": 5}}))
	if outcome != "success" {
		t.Fatalf("log failed: %s", text)
	}
	var res struct {
		Result []map[string]any `json:"result"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if len(res.Result) == 0 {
		t.Fatalf("log returned no commits: %s", text)
	}
}

func TestGitExecMutationRequestsGitMutationPermission(t *testing.T) {
	tool := NewGitExec(git.New(""))
	qs, err := tool.Resources([]byte(`{"operation":"commit","repo":"."}`))
	if err != nil {
		t.Fatal(err)
	}
	var sawMutation bool
	for _, q := range qs {
		if q.Permission == "git_mutation" {
			sawMutation = true
		}
	}
	if !sawMutation {
		t.Fatal("commit must request the git_mutation permission")
	}
	// Read operations must not request git_mutation.
	qs, _ = tool.Resources([]byte(`{"operation":"status","repo":"."}`))
	for _, q := range qs {
		if q.Permission == "git_mutation" {
			t.Fatal("status must not request git_mutation")
		}
	}
}

func TestGitExecUnsupportedOperation(t *testing.T) {
	tool := NewGitExec(git.New(""))
	if outcome, _ := runTool(t, tool, `{"operation":"nuke","repo":"."}`); outcome != "error" {
		t.Fatal("unsupported operation should error")
	}
}
