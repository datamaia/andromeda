// Package checkpoint snapshots and restores a workspace's working tree using git plumbing, without
// touching the user's index, branch, or stash. It backs the interactive /undo and /redo commands: a
// snapshot is taken before each agent turn, and /undo restores the previous one. Snapshots are git
// tree objects (built through a throwaway index), so they are cheap and de-duplicated by git's
// object store. It requires a git repository; callers check Available first. Layer L3 (a small,
// git-backed engine).
package checkpoint

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Available reports whether root is inside a git working tree (and git is on PATH).
func Available(root string) bool {
	_, err := runGit(root, nil, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// Snapshot captures the entire working tree (tracked and untracked, excluding .gitignored files) as
// a git tree object and returns its SHA. It stages into a throwaway index file, so the user's real
// index is never modified.
func Snapshot(root string) (string, error) {
	idxPath, cleanup, err := tempIndex()
	if err != nil {
		return "", err
	}
	defer cleanup()
	env := indexEnv(idxPath)
	// An empty starting index + "add -A" stages the whole working tree as the snapshot content.
	if _, err := runGit(root, env, "add", "-A"); err != nil {
		return "", err
	}
	out, err := runGit(root, env, "write-tree")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Restore overwrites the working tree to match a snapshot tree: edited files are reverted, deleted
// files are recreated, and files created after the snapshot (which the snapshot does not contain)
// are removed — excluding .gitignored files, which are never touched.
func Restore(root, tree string) error {
	if strings.TrimSpace(tree) == "" {
		return errors.New("empty checkpoint")
	}
	idxPath, cleanup, err := tempIndex()
	if err != nil {
		return err
	}
	defer cleanup()
	env := indexEnv(idxPath)
	// Load the snapshot into the throwaway index and record which files it holds.
	if _, err := runGit(root, env, "read-tree", tree); err != nil {
		return err
	}
	snapOut, err := runGit(root, env, "ls-files")
	if err != nil {
		return err
	}
	// Write every snapshot file back to the working tree (reverting edits, restoring deletions).
	if _, err := runGit(root, env, "checkout-index", "-a", "-f"); err != nil {
		return err
	}
	// Remove files that exist now but were not in the snapshot — i.e. created after it. -o lists
	// untracked and -c tracked, both excluding .gitignored, so ignored files are left alone.
	snap := lineSet(snapOut)
	curOut, err := runGit(root, nil, "ls-files", "-o", "-c", "--exclude-standard")
	if err != nil {
		return err
	}
	for _, f := range strings.Split(curOut, "\n") {
		f = strings.TrimSpace(f)
		if f == "" || snap[f] {
			continue
		}
		_ = os.Remove(filepath.Join(root, filepath.FromSlash(f)))
	}
	return nil
}

// tempIndex returns a path for a throwaway git index file and a cleanup func. The file itself is
// created by git on first use; we only reserve a unique name.
func tempIndex() (string, func(), error) {
	f, err := os.CreateTemp("", "andromeda-ckpt-*.idx")
	if err != nil {
		return "", func() {}, err
	}
	path := f.Name()
	_ = f.Close()
	_ = os.Remove(path) // git writes it fresh; a stale empty file would be treated as a corrupt index
	return path, func() { _ = os.Remove(path) }, nil
}

func indexEnv(idxPath string) []string {
	return append(os.Environ(), "GIT_INDEX_FILE="+idxPath)
}

// runGit runs a git subcommand in root with an optional environment, returning stdout. git's stderr
// is surfaced on failure so the caller can report why.
func runGit(root string, env []string, args ...string) (string, error) {
	cmd := exec.Command("git", args...) //nolint:gosec // args are fixed plumbing verbs + a git SHA
	cmd.Dir = root
	if env != nil {
		cmd.Env = env
	}
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %s", args[0], strings.TrimSpace(string(ee.Stderr)))
		}
		return "", err
	}
	return string(out), nil
}

// lineSet splits git output into a set of non-empty trimmed lines.
func lineSet(s string) map[string]bool {
	set := map[string]bool{}
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			set[l] = true
		}
	}
	return set
}
