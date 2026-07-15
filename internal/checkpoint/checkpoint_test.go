package checkpoint

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// newRepo creates a git repo in a temp dir with an initial file, returning the root.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=t@e.com",
			"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=t@e.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "t@e.com")
	run("config", "user.name", "Test")
	return dir
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, root, rel string) (string, bool) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(b), true
}

func TestAvailable(t *testing.T) {
	if !Available(newRepo(t)) {
		t.Fatal("a git repo should be Available")
	}
	if Available(t.TempDir()) {
		t.Fatal("a non-repo should not be Available")
	}
}

// The core safety property: a snapshot faithfully reverts edits, restores deletions, and removes
// files created after it — without disturbing .gitignored files.
func TestSnapshotRestoreEditDeleteCreate(t *testing.T) {
	root := newRepo(t)
	write(t, root, "keep.txt", "original")
	write(t, root, "gone.txt", "to be deleted")
	write(t, root, ".gitignore", "secret.env\n")
	write(t, root, "secret.env", "DO NOT TOUCH")

	snap, err := Snapshot(root)
	if err != nil {
		t.Fatal(err)
	}
	if snap == "" {
		t.Fatal("snapshot returned empty tree")
	}

	// Mutate: edit keep, delete gone, create fresh, and change the ignored file.
	write(t, root, "keep.txt", "EDITED")
	if err := os.Remove(filepath.Join(root, "gone.txt")); err != nil {
		t.Fatal(err)
	}
	write(t, root, "fresh.txt", "created after snapshot")
	write(t, root, "secret.env", "CHANGED-BY-USER")

	if err := Restore(root, snap); err != nil {
		t.Fatal(err)
	}

	// Edit reverted.
	if got, _ := read(t, root, "keep.txt"); got != "original" {
		t.Errorf("keep.txt = %q, want reverted to original", got)
	}
	// Deletion restored.
	if got, ok := read(t, root, "gone.txt"); !ok || got != "to be deleted" {
		t.Errorf("gone.txt not restored: %q (present=%v)", got, ok)
	}
	// Post-snapshot creation removed.
	if _, ok := read(t, root, "fresh.txt"); ok {
		t.Error("fresh.txt (created after snapshot) should have been removed")
	}
	// Ignored file untouched — never in the snapshot, never removed or reverted.
	if got, ok := read(t, root, "secret.env"); !ok || got != "CHANGED-BY-USER" {
		t.Errorf("gitignored secret.env should be left alone, got %q (present=%v)", got, ok)
	}
}

func TestRestoreEmptyTreeErrors(t *testing.T) {
	if err := Restore(newRepo(t), "   "); err == nil {
		t.Fatal("restoring an empty checkpoint should error")
	}
}

// Two snapshots of identical trees produce the same SHA (git content-addressing), which the driver
// relies on to skip no-op checkpoints.
func TestSnapshotStableForUnchangedTree(t *testing.T) {
	root := newRepo(t)
	write(t, root, "a.txt", "x")
	s1, err := Snapshot(root)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := Snapshot(root)
	if err != nil {
		t.Fatal(err)
	}
	if s1 != s2 {
		t.Fatalf("identical trees should snapshot to the same SHA: %q vs %q", s1, s2)
	}
}
