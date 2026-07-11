package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/datamaia/andromeda/internal/ports"
)

// MinGitMajor and MinGitMinor are the ADR-025 floor (git >= 2.40).
const (
	MinGitMajor = 2
	MinGitMinor = 40
)

// Engine implements ports.GitPort by shelling out to the system git.
type Engine struct {
	gitPath string
}

// New returns a Git Engine using the given git executable (empty means "git" on PATH).
func New(gitPath string) *Engine {
	if gitPath == "" {
		gitPath = "git"
	}
	return &Engine{gitPath: gitPath}
}

var _ ports.GitPort = (*Engine)(nil)

// run executes git in the repository and returns stdout. Failures map to E-GIT with git's
// stderr as safe-to-log detail.
func (e *Engine) run(ctx context.Context, repo ports.RepoRef, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, e.gitPath, args...) //nolint:gosec // args are engine-controlled
	if repo.Root != "" {
		cmd.Dir = repo.Root
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), &ports.PortError{
			Code: "E-GIT-001", Category: "git", Severity: "error",
			Message: "git command failed", Detail: strings.TrimSpace(stderr.String()), Cause: err,
		}
	}
	return stdout.String(), nil
}

func (e *Engine) runInput(ctx context.Context, repo ports.RepoRef, stdin string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, e.gitPath, args...) //nolint:gosec // args are engine-controlled
	if repo.Root != "" {
		cmd.Dir = repo.Root
	}
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), &ports.PortError{
			Code: "E-GIT-001", Category: "git", Severity: "error",
			Message: "git command failed", Detail: strings.TrimSpace(stderr.String()), Cause: err,
		}
	}
	return stdout.String(), nil
}

// Version detects the system git version and gates on the ADR-025 floor.
func (e *Engine) Version(ctx context.Context) (ports.GitVersion, error) {
	out, err := e.run(ctx, ports.RepoRef{}, "--version")
	if err != nil {
		return ports.GitVersion{}, err
	}
	v := parseVersion(strings.TrimSpace(out))
	if v.Major < MinGitMajor || (v.Major == MinGitMajor && v.Minor < MinGitMinor) {
		return v, &ports.PortError{
			Code: "E-GIT-002", Category: "configuration", Severity: "error",
			Message: fmt.Sprintf("git %d.%d or newer is required", MinGitMajor, MinGitMinor),
			Detail:  v.Raw,
		}
	}
	return v, nil
}

func parseVersion(raw string) ports.GitVersion {
	v := ports.GitVersion{Raw: raw}
	fields := strings.Fields(raw) // "git version 2.50.1 ..."
	for _, f := range fields {
		if strings.Count(f, ".") >= 1 && f[0] >= '0' && f[0] <= '9' {
			parts := strings.Split(f, ".")
			v.Major, _ = strconv.Atoi(parts[0])
			if len(parts) > 1 {
				v.Minor, _ = strconv.Atoi(parts[1])
			}
			break
		}
	}
	return v
}

// Status parses `git status --porcelain=v1 -z` plus the current branch.
func (e *Engine) Status(ctx context.Context, repo ports.RepoRef) (ports.RepoStatus, error) {
	branch, err := e.run(ctx, repo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ports.RepoStatus{}, err
	}
	out, err := e.run(ctx, repo, "status", "--porcelain=v1", "-z")
	if err != nil {
		return ports.RepoStatus{}, err
	}
	st := ports.RepoStatus{Branch: strings.TrimSpace(branch)}
	for _, entry := range strings.Split(out, "\x00") {
		if len(entry) < 3 {
			continue
		}
		x, y, path := entry[0], entry[1], entry[3:]
		switch {
		case x == '?' && y == '?':
			st.Untracked = append(st.Untracked, path)
		default:
			if x != ' ' && x != 0 {
				st.Staged = append(st.Staged, path)
			}
			if y != ' ' && y != 0 {
				st.Unstaged = append(st.Unstaged, path)
			}
		}
	}
	st.Clean = len(st.Staged) == 0 && len(st.Unstaged) == 0 && len(st.Untracked) == 0
	return st, nil
}

// Stage stages paths (git add).
func (e *Engine) Stage(ctx context.Context, repo ports.RepoRef, paths []ports.Path) error {
	if len(paths) == 0 {
		return nil
	}
	_, err := e.run(ctx, repo, append([]string{"add", "--"}, paths...)...)
	return err
}

// Unstage removes paths from the index (git restore --staged).
func (e *Engine) Unstage(ctx context.Context, repo ports.RepoRef, paths []ports.Path) error {
	if len(paths) == 0 {
		return nil
	}
	_, err := e.run(ctx, repo, append([]string{"restore", "--staged", "--"}, paths...)...)
	return err
}

// Commit creates a commit and returns its hash.
func (e *Engine) Commit(ctx context.Context, repo ports.RepoRef, spec ports.CommitSpec) (ports.CommitID, error) {
	args := []string{"commit", "-m", spec.Message}
	if spec.Author != "" {
		args = append(args, "--author", spec.Author)
	}
	if spec.Signoff {
		args = append(args, "--signoff")
	}
	if _, err := e.run(ctx, repo, args...); err != nil {
		return "", err
	}
	out, err := e.run(ctx, repo, "rev-parse", "HEAD")
	return strings.TrimSpace(out), err
}

// ApplyPatch applies a unified patch atomically (git apply --index).
func (e *Engine) ApplyPatch(ctx context.Context, repo ports.RepoRef, patch ports.PatchDocument) (ports.GitApplyReport, error) {
	// Check first so a failure applies nothing.
	if _, err := e.runInput(ctx, repo, patch.Unified, "apply", "--check", "-"); err != nil {
		return ports.GitApplyReport{Applied: false, Rejects: []string{"patch does not apply cleanly"}}, err
	}
	if _, err := e.runInput(ctx, repo, patch.Unified, "apply", "--index", "-"); err != nil {
		return ports.GitApplyReport{Applied: false}, err
	}
	return ports.GitApplyReport{Applied: true}, nil
}
