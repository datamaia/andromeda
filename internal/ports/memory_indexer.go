package ports

import "context"

// MemoryStorePort is the Memory Manager's public face: persistent memory across the layers of
// Volume 7, backed by the databases of ADR-028. Contract owner: Volume 7. Errors: E-MEM.
type MemoryStorePort interface {
	Ingest(ctx context.Context, records []MemoryRecordDraft) ([]ULID, error)
	Retrieve(ctx context.Context, q MemoryQuery) ([]MemoryRecord, error)
	Rank(ctx context.Context, q MemoryQuery, candidates []ULID) ([]RankedMemory, error)
	Expire(ctx context.Context, policy ExpirePolicy) (ExpireReport, error)
	Delete(ctx context.Context, ids []ULID) error
	Export(ctx context.Context, q MemoryQuery) (Stream[MemoryRecord], error)
}

// MemoryRecordDraft is an unpersisted memory record with provenance and retention attributes.
type MemoryRecordDraft struct {
	Layer      string // "session" | "workspace" | "long_term" | "semantic" | "episodic"
	Content    string
	Provenance string
	Source     string
	Retention  string
}

// MemoryRecord is a persisted memory record.
type MemoryRecord struct {
	ID         ULID
	Layer      string
	Content    string
	Provenance string
	Source     string
	Status     string // Volume 2 chapter 09 recorded-status vocabulary
	CreatedAt  string // RFC 3339 UTC
}

// MemoryQuery selects records by layer, scope, provenance, time, and content.
type MemoryQuery struct {
	Layers   []string
	Scope    string
	Text     string
	Semantic bool
	Limit    int
}

// RankedMemory is a candidate scored against a query.
type RankedMemory struct {
	ID    ULID
	Score float64
}

// ExpirePolicy parameterizes a retention pass.
type ExpirePolicy struct {
	Layer     string
	OlderThan string // RFC 3339 UTC or a duration expression
}

// ExpireReport summarizes a retention pass.
type ExpireReport struct {
	Archived int
	Expired  int
}

// IndexerPort provides queryable indexes over workspace content — lexical and semantic
// (ADR-020) — with the Index lifecycle frozen in Volume 2 chapter 09. Contract owner:
// Volume 7. Errors: E-IDX.
type IndexerPort interface {
	Build(ctx context.Context, spec IndexSpec) (ULID, error)
	Update(ctx context.Context, indexID ULID, changes []PathChange) error
	Query(ctx context.Context, indexID ULID, q IndexQuery) ([]IndexHit, error)
	Invalidate(ctx context.Context, indexID ULID, scope InvalidateScope) error
	Status(ctx context.Context, indexID ULID) (IndexStatus, error)
}

// IndexSpec declares an index for a workspace scope.
type IndexSpec struct {
	WorkspaceID ULID
	Kind        string // "lexical" | "semantic"
	Include     []Path
	Exclude     []Path
}

// PathChange is an incremental change to index.
type PathChange struct {
	Path Path
	Kind string // "created" | "modified" | "deleted" | "renamed"
}

// IndexQuery is a lexical or semantic search with budgets.
type IndexQuery struct {
	Text       string
	Semantic   bool
	MaxResults int
	MaxLatency int // milliseconds hint
}

// IndexHit is one search result.
type IndexHit struct {
	Path       Path
	StartLine  int
	EndLine    int
	Score      float64
	Generation int64
}

// InvalidateScope names what to invalidate (paths or the whole index).
type InvalidateScope struct {
	Whole bool
	Paths []Path
}

// IndexStatus is the current state, generation, coverage, and staleness of one index.
type IndexStatus struct {
	State      string // Index lifecycle states (Volume 2 chapter 09)
	Generation int64
	Coverage   int
	Stale      bool
}
