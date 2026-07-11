package git

import (
	"context"
	"strings"

	"github.com/datamaia/andromeda/internal/ports"
)

// WorktreeAdd creates a new worktree.
func (e *Engine) WorktreeAdd(ctx context.Context, repo ports.RepoRef, spec ports.WorktreeSpec) (ports.WorktreeInfo, error) {
	args := []string{"worktree", "add"}
	if spec.Branch != "" {
		args = append(args, "-b", spec.Branch)
	}
	args = append(args, spec.Path)
	if _, err := e.run(ctx, repo, args...); err != nil {
		return ports.WorktreeInfo{}, err
	}
	return ports.WorktreeInfo{Path: spec.Path, Branch: spec.Branch}, nil
}

// WorktreeList lists worktrees (git worktree list --porcelain).
func (e *Engine) WorktreeList(ctx context.Context, repo ports.RepoRef) ([]ports.WorktreeInfo, error) {
	out, err := e.run(ctx, repo, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	var list []ports.WorktreeInfo
	var cur ports.WorktreeInfo
	flush := func() {
		if cur.Path != "" {
			list = append(list, cur)
		}
		cur = ports.WorktreeInfo{}
	}
	for _, ln := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(ln, "worktree "):
			flush()
			cur.Path = strings.TrimPrefix(ln, "worktree ")
		case strings.HasPrefix(ln, "HEAD "):
			cur.Head = strings.TrimPrefix(ln, "HEAD ")
		case strings.HasPrefix(ln, "branch "):
			cur.Branch = strings.TrimPrefix(strings.TrimPrefix(ln, "branch "), "refs/heads/")
		}
	}
	flush()
	return list, nil
}

// WorktreeRemove removes a worktree.
func (e *Engine) WorktreeRemove(ctx context.Context, repo ports.RepoRef, path ports.Path) error {
	_, err := e.run(ctx, repo, "worktree", "remove", path)
	return err
}
