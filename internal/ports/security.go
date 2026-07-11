package ports

import (
	"context"

	"github.com/datamaia/andromeda/internal/core"
)

// PermissionPort is the single decision path for side-effecting actions (Principle 8).
// Decision semantics, the permission enum, scopes, and policy interplay are Volume 9's.
// Errors: E-SEC. Denial is a decision (a Decision value), not an error.
type PermissionPort interface {
	Check(ctx context.Context, req PermissionQuery) (Decision, error)
	Request(ctx context.Context, req PermissionRequest) (Decision, error)
	RecordDecision(ctx context.Context, rec DecisionRecord) error
}

// PermissionQuery describes an action to evaluate against standing grants and policies.
type PermissionQuery struct {
	Permission core.Permission
	Scope      core.PermissionScope
	Subject    string // the concrete resource: path, host, tool name, repository, ...
}

// PermissionRequest is a query that may raise an interactive Approval where permitted.
type PermissionRequest struct {
	Query       PermissionQuery
	Interactive bool
	Reason      string
}

// Decision is the resolved outcome of a permission evaluation.
type Decision struct {
	Outcome    core.DecisionOutcome
	Kind       core.PermissionDecisionKind
	ApprovalID ULID // set when an Approval was raised
}

// DecisionRecord is a decision produced elsewhere (e.g. policy pre-resolution) to persist.
type DecisionRecord struct {
	Query    PermissionQuery
	Decision Decision
	Actor    string
}

// SecretStorePort holds credential material behind references (ADR-014). Only references
// cross other ports; only this port touches material. Contract owner: Volume 9. Errors:
// E-SEC. All methods are local-only — no network (Principle 3).
type SecretStorePort interface {
	Get(ctx context.Context, ref SecretRef) (SecretValue, error)
	Set(ctx context.Context, ref SecretRef, value SecretValue, meta SecretMeta) error
	Delete(ctx context.Context, ref SecretRef) error
	List(ctx context.Context, scope SecretScope) ([]SecretRef, error)
}

// SecretRef references credential material without exposing it.
type SecretRef struct {
	Namespace string
	Name      string
}

// SecretValue is a zeroize-on-release wrapper; callers MUST NOT persist or log it.
type SecretValue struct {
	material []byte
}

// NewSecretValue wraps raw material.
func NewSecretValue(b []byte) SecretValue { return SecretValue{material: b} }

// Bytes returns the underlying material for immediate use; do not retain the slice.
func (v SecretValue) Bytes() []byte { return v.material }

// Zero clears the material in place.
func (v *SecretValue) Zero() {
	for i := range v.material {
		v.material[i] = 0
	}
}

// SecretMeta is non-secret metadata about a credential.
type SecretMeta struct {
	Kind      string
	Provider  string
	CreatedAt string // RFC 3339 UTC
}

// SecretScope selects references to enumerate.
type SecretScope struct {
	Namespace string
}

// SandboxPort applies isolation policy to anything Andromeda executes (ADR-021). Policy
// content and the containment model are Volume 9's; the layered mechanism (process-level at
// MVP, OS-level from Beta/v1) is fixed by ADR-021. Errors: E-SEC.
type SandboxPort interface {
	Prepare(ctx context.Context, spec SandboxSpec) (SandboxHandle, error)
	ApplyPolicy(ctx context.Context, sb SandboxHandle, policy SandboxPolicy) error
	ExecuteIn(ctx context.Context, sb SandboxHandle, cmd CommandSpec) (ExecutionID, error)
	Teardown(ctx context.Context, sb SandboxHandle) error
}

// SandboxSpec describes the execution environment to allocate.
type SandboxSpec struct {
	WorkingDir Path
	EnvAllow   []string // deny-by-default passthrough allowlist (ADR-021)
}

// SandboxHandle references a prepared sandbox; its observable state includes the effective
// containment level (ADR-021).
type SandboxHandle struct {
	ID               ULID
	ContainmentLevel string // "process" | "os" — the effective, recorded level
}

// SandboxPolicy is the effective policy: path rules, network stance, resource limits, and
// isolation mechanism selection.
type SandboxPolicy struct {
	ReadPaths     []Path
	WritePaths    []Path
	NetworkPolicy string // "deny" | "allow" | "restricted"
	CPULimit      int
	MemLimitMB    int
	TimeLimitSec  int
	CommandAllow  []string
	CommandDeny   []string
}
