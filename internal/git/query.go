package git

import (
	"context"
	"strings"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

const (
	unitSep   = "\x1f"
	recordSep = "\x1e"
)

// Diff streams the diff hunks for a spec. Output is parsed and delivered as a slice-backed
// stream; true incremental streaming for very large diffs is a later refinement.
func (e *Engine) Diff(ctx context.Context, repo ports.RepoRef, spec ports.DiffSpec) (ports.Stream[ports.DiffHunk], error) {
	args := []string{"diff"}
	if spec.Staged {
		args = append(args, "--staged")
	}
	if spec.From != "" {
		args = append(args, spec.From)
	}
	if spec.To != "" {
		args = append(args, spec.To)
	}
	if len(spec.Paths) > 0 {
		args = append(args, "--")
		args = append(args, spec.Paths...)
	}
	out, err := e.run(ctx, repo, args...)
	if err != nil {
		return nil, err
	}
	return streams.Slice(parseDiff(out)), nil
}

func parseDiff(out string) []ports.DiffHunk {
	if strings.TrimSpace(out) == "" {
		return nil
	}
	var hunks []ports.DiffHunk
	blocks := strings.Split(out, "\ndiff --git ")
	for i, b := range blocks {
		if b == "" {
			continue
		}
		if i > 0 {
			b = "diff --git " + b
		}
		lines := strings.Split(b, "\n")
		h := ports.DiffHunk{Header: lines[0], Lines: lines[1:], Path: pathFromDiffHeader(lines[0])}
		hunks = append(hunks, h)
	}
	return hunks
}

func pathFromDiffHeader(header string) string {
	// "diff --git a/path b/path"
	fields := strings.Fields(header)
	if len(fields) >= 4 {
		return strings.TrimPrefix(fields[2], "a/")
	}
	return ""
}

// Log streams commit summaries.
func (e *Engine) Log(ctx context.Context, repo ports.RepoRef, spec ports.LogSpec) (ports.Stream[ports.CommitInfo], error) {
	format := "--format=%H" + unitSep + "%an" + unitSep + "%aI" + unitSep + "%s" + recordSep
	args := []string{"log", format}
	if spec.Max > 0 {
		args = append(args, "-n", intToStr(spec.Max))
	}
	if spec.Rev != "" {
		args = append(args, spec.Rev)
	}
	if len(spec.Paths) > 0 {
		args = append(args, "--")
		args = append(args, spec.Paths...)
	}
	out, err := e.run(ctx, repo, args...)
	if err != nil {
		return nil, err
	}
	var commits []ports.CommitInfo
	for _, rec := range strings.Split(out, recordSep) {
		rec = strings.Trim(rec, "\n")
		if rec == "" {
			continue
		}
		f := strings.Split(rec, unitSep)
		if len(f) < 4 {
			continue
		}
		commits = append(commits, ports.CommitInfo{ID: f[0], Author: f[1], Date: f[2], Subject: f[3]})
	}
	return streams.Slice(commits), nil
}

// Show returns a full commit view.
func (e *Engine) Show(ctx context.Context, repo ports.RepoRef, rev ports.Revision) (ports.CommitDetail, error) {
	format := "--format=%H" + unitSep + "%an" + unitSep + "%aI" + unitSep + "%s"
	out, err := e.run(ctx, repo, "show", "--no-patch", format, rev)
	if err != nil {
		return ports.CommitDetail{}, err
	}
	f := strings.Split(strings.TrimSpace(out), unitSep)
	var d ports.CommitDetail
	if len(f) >= 4 {
		d.Info = ports.CommitInfo{ID: f[0], Author: f[1], Date: f[2], Subject: f[3]}
	}
	files, err := e.run(ctx, repo, "show", "--name-only", "--format=", rev)
	if err == nil {
		for _, ln := range strings.Split(strings.TrimSpace(files), "\n") {
			if ln != "" {
				d.Files = append(d.Files, ln)
			}
		}
	}
	return d, nil
}

// ListBranches lists local branches.
func (e *Engine) ListBranches(ctx context.Context, repo ports.RepoRef) ([]ports.BranchInfo, error) {
	format := "--format=%(refname:short)" + unitSep + "%(HEAD)" + unitSep + "%(upstream:short)"
	out, err := e.run(ctx, repo, "branch", format)
	if err != nil {
		return nil, err
	}
	var branches []ports.BranchInfo
	for _, ln := range strings.Split(strings.TrimSpace(out), "\n") {
		if ln == "" {
			continue
		}
		f := strings.Split(ln, unitSep)
		bi := ports.BranchInfo{Name: f[0]}
		if len(f) > 1 {
			bi.Current = strings.TrimSpace(f[1]) == "*"
		}
		if len(f) > 2 {
			bi.Upstream = f[2]
		}
		branches = append(branches, bi)
	}
	return branches, nil
}

// CreateBranch creates a branch (without switching).
func (e *Engine) CreateBranch(ctx context.Context, repo ports.RepoRef, spec ports.BranchSpec) error {
	args := []string{"branch", spec.Name}
	if spec.From != "" {
		args = append(args, spec.From)
	}
	_, err := e.run(ctx, repo, args...)
	return err
}

// SwitchBranch switches the working tree to a branch.
func (e *Engine) SwitchBranch(ctx context.Context, repo ports.RepoRef, name string) error {
	_, err := e.run(ctx, repo, "switch", name)
	return err
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
