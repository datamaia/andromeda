package git

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// initRepo creates a temporary git repository with one commit and returns its root.
func initRepo(t *testing.T) ports.RepoRef {
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
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "initial commit")
	return ports.RepoRef{Root: dir}
}

func TestVersionMeetsFloor(t *testing.T) {
	v, err := New("").Version(context.Background())
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if v.Major < MinGitMajor {
		t.Fatalf("git major %d below floor", v.Major)
	}
}

func TestStatusCleanThenDirty(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	e := New("")
	st, err := e.Status(ctx, repo)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Clean || st.Branch != "main" {
		t.Fatalf("status = %+v, want clean on main", st)
	}
	// Modify a file → unstaged; create a new file → untracked.
	os.WriteFile(filepath.Join(repo.Root, "README.md"), []byte("# changed\n"), 0o600)
	os.WriteFile(filepath.Join(repo.Root, "new.txt"), []byte("x\n"), 0o600)
	st2, _ := e.Status(ctx, repo)
	if st2.Clean {
		t.Fatal("expected dirty status")
	}
	if len(st2.Unstaged) == 0 || len(st2.Untracked) == 0 {
		t.Fatalf("status = %+v", st2)
	}
}

func TestStageCommitLog(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	e := New("")
	os.WriteFile(filepath.Join(repo.Root, "feature.go"), []byte("package x\n"), 0o600)
	if err := e.Stage(ctx, repo, []ports.Path{"feature.go"}); err != nil {
		t.Fatal(err)
	}
	id, err := e.Commit(ctx, repo, ports.CommitSpec{Message: "feat: add feature"})
	if err != nil {
		t.Fatal(err)
	}
	if len(id) < 7 {
		t.Fatalf("commit id looks wrong: %q", id)
	}
	st, err := e.Log(ctx, repo, ports.LogSpec{Max: 10})
	if err != nil {
		t.Fatal(err)
	}
	var subjects []string
	for {
		c, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		subjects = append(subjects, c.Subject)
	}
	if len(subjects) != 2 || subjects[0] != "feat: add feature" {
		t.Fatalf("log subjects = %v", subjects)
	}
}

func TestDiffAndShow(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	e := New("")
	os.WriteFile(filepath.Join(repo.Root, "README.md"), []byte("# hello\nmore\n"), 0o600)
	st, err := e.Diff(ctx, repo, ports.DiffSpec{})
	if err != nil {
		t.Fatal(err)
	}
	h, err := st.Next(ctx)
	if err != nil {
		t.Fatalf("expected a diff hunk: %v", err)
	}
	if h.Path != "README.md" {
		t.Errorf("diff path = %q", h.Path)
	}

	head, _ := e.Show(ctx, repo, "HEAD")
	if head.Info.Subject != "initial commit" {
		t.Errorf("show HEAD subject = %q", head.Info.Subject)
	}
	if len(head.Files) == 0 {
		t.Error("show should list changed files")
	}
}

func TestBranchesAndSwitch(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	e := New("")
	if err := e.CreateBranch(ctx, repo, ports.BranchSpec{Name: "feature/x"}); err != nil {
		t.Fatal(err)
	}
	branches, err := e.ListBranches(ctx, repo)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, b := range branches {
		names[b.Name] = true
	}
	if !names["main"] || !names["feature/x"] {
		t.Fatalf("branches = %v", branches)
	}
	if err := e.SwitchBranch(ctx, repo, "feature/x"); err != nil {
		t.Fatal(err)
	}
	st, _ := e.Status(ctx, repo)
	if st.Branch != "feature/x" {
		t.Errorf("after switch, branch = %q", st.Branch)
	}
}

func TestApplyPatch(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	e := New("")
	patch := "diff --git a/README.md b/README.md\n" +
		"index 0000000..1111111 100644\n" +
		"--- a/README.md\n" +
		"+++ b/README.md\n" +
		"@@ -1 +1,2 @@\n" +
		" # hello\n" +
		"+added by patch\n"
	rep, err := e.ApplyPatch(ctx, repo, ports.PatchDocument{Unified: patch})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !rep.Applied {
		t.Fatal("patch not applied")
	}
	data, _ := os.ReadFile(filepath.Join(repo.Root, "README.md"))
	if string(data) != "# hello\nadded by patch\n" {
		t.Errorf("patched content = %q", data)
	}
}

func TestWorktrees(t *testing.T) {
	ctx := context.Background()
	repo := initRepo(t)
	e := New("")
	wtPath := filepath.Join(t.TempDir(), "wt")
	if _, err := e.WorktreeAdd(ctx, repo, ports.WorktreeSpec{Path: wtPath, Branch: "wt-branch"}); err != nil {
		t.Fatal(err)
	}
	list, err := e.WorktreeList(ctx, repo)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 2 {
		t.Fatalf("worktree list = %v, want >= 2", list)
	}
	if err := e.WorktreeRemove(ctx, repo, wtPath); err != nil {
		t.Fatal(err)
	}
}

func TestErrorMapsToE_GIT(t *testing.T) {
	ctx := context.Background()
	e := New("")
	_, err := e.Status(ctx, ports.RepoRef{Root: t.TempDir()}) // not a repo
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code == "" || pe.Category != "git" {
		t.Fatalf("want E-GIT PortError, got %v", err)
	}
}
