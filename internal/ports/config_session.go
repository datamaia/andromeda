package ports

import "context"

// ConfigPort resolves configuration with precedence, validation, and watching. Schema,
// precedence order, and migration are Volume 10's single-home topics. Errors: E-CFG (exit
// code 3 at the CLI boundary). Resolution never blocks on the network.
type ConfigPort interface {
	Resolve(ctx context.Context, q ConfigQuery) (ResolvedConfig, error)
	Validate(ctx context.Context, doc ConfigDocument) (ValidationReport, error)
	Watch(ctx context.Context, sel ConfigSelector) (Stream[ConfigChange], error)
}

// ConfigQuery selects the scope to resolve.
type ConfigQuery struct {
	Scope       string // "global" | "workspace" | "project" | "invocation"
	WorkspaceID ULID
}

// ResolvedConfig is the effective configuration with per-value source attribution.
type ResolvedConfig struct {
	Values  map[string]any
	Sources map[string]string // key -> the layer that supplied it
}

// ConfigDocument is a configuration document to validate without applying.
type ConfigDocument struct {
	Format string // "toml"
	Raw    []byte
}

// ValidationReport carries all validation findings, not just the first.
type ValidationReport struct {
	Valid    bool
	Findings []ConfigFinding
}

// ConfigFinding is one validation finding with its location.
type ConfigFinding struct {
	Key      string
	Message  string
	Code     string // E-CFG-NNN
	Severity string
}

// ConfigSelector selects keys to watch.
type ConfigSelector struct {
	Keys     []string
	Prefixes []string
}

// ConfigChange is a resolved change delta.
type ConfigChange struct {
	Key      string
	OldValue any
	NewValue any
	Source   string
}

// SessionStorePort persists and restores Sessions and Runs — the durability behind PRD-010.
// Storage mechanics are Volume 10's; run/turn semantics are Volume 4's. Errors: E-CFG;
// corruption and migration failures follow ADR-029 (exit code 9).
type SessionStorePort interface {
	SaveSession(ctx context.Context, s SessionSnapshot) error
	LoadSession(ctx context.Context, id ULID) (SessionSnapshot, error)
	ListSessions(ctx context.Context, f SessionFilter) ([]SessionSummary, error)
	AppendRunRecords(ctx context.Context, runID ULID, batch []RunRecord) error
	LoadRun(ctx context.Context, id ULID) (RunSnapshot, error)
	ListRuns(ctx context.Context, f RunFilter) ([]RunSummary, error)
	MarkInterrupted(ctx context.Context, scope InterruptScope) ([]ULID, error)
}

// SessionSnapshot is a session row and its state, with optimistic concurrency via Revision.
type SessionSnapshot struct {
	ID        ULID
	State     string // canonical Session states (Volume 2 chapter 09)
	Revision  int64
	CreatedAt string
	Data      JSON
}

// SessionFilter filters session listings.
type SessionFilter struct {
	State string
	Limit int
}

// SessionSummary is a compact session listing row.
type SessionSummary struct {
	ID        ULID
	State     string
	CreatedAt string
}

// RunRecord is one appended run record (turn, message, transition, tool invocation, ...).
type RunRecord struct {
	Kind    string
	Payload JSON
}

// RunSnapshot reconstructs a run's full state.
type RunSnapshot struct {
	ID      ULID
	State   string // canonical Run states (Volume 2 chapter 09)
	Records []RunRecord
}

// RunFilter filters run listings.
type RunFilter struct {
	SessionID ULID
	State     string
	Limit     int
}

// RunSummary is a compact run listing row.
type RunSummary struct {
	ID        ULID
	SessionID ULID
	State     string
}

// InterruptScope selects runs/tasks to mark interrupted during crash recovery.
type InterruptScope struct {
	WorkspaceID ULID
	SessionID   ULID
}
