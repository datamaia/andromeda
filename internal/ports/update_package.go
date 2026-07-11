package ports

import "context"

// UpdaterPort is self-update of the installed product (MVP item 23). Full update semantics,
// channels, and the Update process machine are Volume 14's. Errors: E-REL. Apply MUST refuse
// to run when Verify has not passed for the same artifact set.
type UpdaterPort interface {
	Check(ctx context.Context) (UpdateCheckResult, error)
	Download(ctx context.Context, rel ReleaseRef) (Stream[DownloadProgress], error)
	Verify(ctx context.Context, rel ReleaseRef) (VerificationReport, error)
	Apply(ctx context.Context, rel ReleaseRef) (UpdateApplyReport, error)
	Rollback(ctx context.Context) (RollbackReport, error)
}

// UpdateCheckResult reports channel status; the only method that needs network.
type UpdateCheckResult struct {
	Status  string // "up_to_date" | "update_available" (frozen Update states)
	Current string
	Latest  string
	Channel string
}

// ReleaseRef references a release artifact set.
type ReleaseRef struct {
	Version string
	Channel string
}

// DownloadProgress is streamed download progress.
type DownloadProgress struct {
	BytesDone  int64
	BytesTotal int64
}

// UpdateApplyReport is the result of applying an update (atomic replace-or-restore).
type UpdateApplyReport struct {
	Applied     bool
	FromVersion string
	ToVersion   string
}

// RollbackReport is the result of restoring the previously retained version.
type RollbackReport struct {
	RolledBack bool
	ToVersion  string
}

// PackagePort manages extension packages: resolve, install, verify, remove. The Package
// installation machine is Volume 6's. Errors: E-PLUG.
type PackagePort interface {
	Resolve(ctx context.Context, req PackageRequest) (ResolutionPlan, error)
	Install(ctx context.Context, plan ResolutionPlan) (Stream[InstallEvent], error)
	Verify(ctx context.Context, pkg PackageRef) (VerificationReport, error)
	Remove(ctx context.Context, pkg PackageRef, opts RemoveOptions) (RemoveReport, error)
}

// PackageRequest asks to install a package by name and constraint.
type PackageRequest struct {
	Name       string
	Constraint string
	Source     string
}

// PackageRef references an installed package.
type PackageRef struct {
	Name    string
	Version string
}

// ResolutionPlan is a concrete, side-effect-free installation plan.
type ResolutionPlan struct {
	Packages  []PackageRef
	Sources   map[string]string
	Checksums map[string]string
}

// InstallEvent is one streamed install-progress event.
type InstallEvent struct {
	State   string // frozen Package installation states (Volume 2 chapter 09)
	Package PackageRef
	Message string
}

// RemoveOptions tunes removal.
type RemoveOptions struct {
	Purge bool
}

// RemoveReport is the result of removing a package.
type RemoveReport struct {
	Removed bool
}
