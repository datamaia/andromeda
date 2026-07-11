package ports

import "context"

// GitPort is the Git operations surface, implemented by the Git Engine per ADR-025 (system
// git >= 2.40 behind an encapsulated adapter). Operation semantics and hosting integrations
// are Volume 11's. Errors: E-GIT. Mutating methods require a permission decision first.
type GitPort interface {
	Version(ctx context.Context) (GitVersion, error)
	Status(ctx context.Context, repo RepoRef) (RepoStatus, error)
	Diff(ctx context.Context, repo RepoRef, spec DiffSpec) (Stream[DiffHunk], error)
	Stage(ctx context.Context, repo RepoRef, paths []Path) error
	Unstage(ctx context.Context, repo RepoRef, paths []Path) error
	Commit(ctx context.Context, repo RepoRef, spec CommitSpec) (CommitID, error)
	Log(ctx context.Context, repo RepoRef, spec LogSpec) (Stream[CommitInfo], error)
	Show(ctx context.Context, repo RepoRef, rev Revision) (CommitDetail, error)
	ListBranches(ctx context.Context, repo RepoRef) ([]BranchInfo, error)
	CreateBranch(ctx context.Context, repo RepoRef, spec BranchSpec) error
	SwitchBranch(ctx context.Context, repo RepoRef, name string) error
	ApplyPatch(ctx context.Context, repo RepoRef, patch PatchDocument) (GitApplyReport, error)
	WorktreeAdd(ctx context.Context, repo RepoRef, spec WorktreeSpec) (WorktreeInfo, error)
	WorktreeList(ctx context.Context, repo RepoRef) ([]WorktreeInfo, error)
	WorktreeRemove(ctx context.Context, repo RepoRef, path Path) error
}

// RepoRef references a repository by its root path.
type RepoRef struct {
	Root Path
}

// GitVersion is the detected system git version.
type GitVersion struct {
	Raw   string
	Major int
	Minor int
}

// RepoStatus summarizes working-tree status.
type RepoStatus struct {
	Branch    string
	Staged    []Path
	Unstaged  []Path
	Untracked []Path
	Clean     bool
}

// DiffSpec parameterizes a diff.
type DiffSpec struct {
	From   Revision
	To     Revision
	Staged bool
	Paths  []Path
}

// DiffHunk is one streamed diff hunk.
type DiffHunk struct {
	Path   Path
	Header string
	Lines  []string
}

// CommitSpec parameterizes a commit.
type CommitSpec struct {
	Message string
	Author  string
	Signoff bool
}

// CommitID is a commit hash.
type CommitID = string

// Revision is a git revision expression (hash, ref, range endpoint).
type Revision = string

// LogSpec parameterizes a log query.
type LogSpec struct {
	Rev   Revision
	Max   int
	Paths []Path
}

// CommitInfo is one streamed log entry.
type CommitInfo struct {
	ID      CommitID
	Author  string
	Date    string
	Subject string
}

// CommitDetail is a full commit view.
type CommitDetail struct {
	Info  CommitInfo
	Files []Path
}

// BranchInfo describes a branch.
type BranchInfo struct {
	Name     string
	Current  bool
	Upstream string
}

// BranchSpec parameterizes branch creation.
type BranchSpec struct {
	Name string
	From Revision
}

// PatchDocument is a reviewed patch to apply.
type PatchDocument struct {
	Unified string
}

// GitApplyReport is the per-file result of applying a patch (atomically or not at all).
type GitApplyReport struct {
	Applied bool
	Files   []Path
	Rejects []string
}

// WorktreeSpec parameterizes worktree creation.
type WorktreeSpec struct {
	Path   Path
	Branch string
}

// WorktreeInfo describes a worktree.
type WorktreeInfo struct {
	Path   Path
	Branch string
	Head   CommitID
}

// WorkspacePort provides workspace and project discovery and lifecycle. Behavior is owned by
// Volume 4; entities by Volume 2. Errors: E-AGT.
type WorkspacePort interface {
	Discover(ctx context.Context, start Path) (WorkspaceCandidate, error)
	Open(ctx context.Context, root Path, opts OpenOptions) (WorkspaceHandle, error)
	Snapshot(ctx context.Context, ws WorkspaceHandle) (WorkspaceSnapshot, error)
	Close(ctx context.Context, ws WorkspaceHandle) error
}

// WorkspaceCandidate is what discovery found without opening.
type WorkspaceCandidate struct {
	Root      Path
	Found     bool
	HasMarker bool // .andromeda/ present
	IsRepo    bool
}

// OpenOptions tunes workspace open.
type OpenOptions struct {
	CreateIfMissing bool
}

// WorkspaceHandle references an open workspace.
type WorkspaceHandle struct {
	ID   ULID
	Root Path
}

// WorkspaceSnapshot is a consistent read-only description of workspace state at a point in
// time — the input for run reproducibility (SM-12) and context assembly.
type WorkspaceSnapshot struct {
	WorkspaceID      ULID
	Root             Path
	Projects         []Path
	ConfigProfile    string
	IndexGenerations map[ULID]int64
	VCSSummary       string
	TakenAt          string // RFC 3339 UTC
}
