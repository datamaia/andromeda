# 02 — Port Interfaces

This chapter is the spine of the corpus's parallel authoring: it defines the **18 public port
interfaces** through which Andromeda's engines reach infrastructure and through which
extensions reach Andromeda. **Port names, method names, and signatures defined here are
frozen** (FR-ARCH-003). The volume named as each port's *contract owner* specifies the full
behavioral contract — semantics, state machines, retries, limits, configuration — and MUST do
so under exactly these names; it MUST NOT rename ports or methods, change parameter or result
arity, or move a method between ports. Additive evolution (new methods, new optional fields on
contract types) follows the compatibility rules of chapter
[09](09-deployment-update-extensibility-compatibility.md).

## Conventions

All ports obey the following rules; they are part of every port's contract and are not
restated per port:

1. **Context first, always.** Every method takes a `Context` as its first parameter and MUST
   honor its cancellation and deadline (ADR-023): return promptly with a cancellation error,
   release resources, and abort or detach side-effect-free work. A method MUST NOT outlive its
   context except where its contract explicitly defines detached completion (none in this
   chapter do).
2. **Typed errors by area.** Each port declares its error family by `E-<AREA>` name; concrete
   codes are minted by the contract owner under the ADR-016 envelope. Adapters MUST map
   underlying failures into the declared family — a caller never sees a raw driver, HTTP, or
   OS error through a port.
3. **Streams.** Streaming results use the `Stream[T]` shape below. Items arrive in order;
   `Next` blocks until an item, end of stream, context cancellation, or failure; `Close` is
   idempotent and cancels production. Every stream MUST be closed by its consumer; adapters
   MUST bound internal buffering (ADR-023 backpressure).
4. **Identifiers.** Entity references cross ports as ULIDs (ADR-027); natural keys and
   external identifiers appear only where the method semantics say so.
5. **No leakage.** Port signatures use only L1 contract types and L0 domain types. SDK types
   of wrapped libraries (MCP SDK per ADR-010, keyring, SQLite driver) MUST NOT appear in any
   port signature.
6. **Thread safety.** Every port implementation MUST be safe for concurrent use by multiple
   goroutines unless a method's semantics state a serialization requirement.

```pseudo
// Shared contract primitives (package internal/ports).

type ULID = string            // 26-char canonical ULID (ADR-027)
type JSON = bytes             // canonical JSON document (Volume 2, chapter 10)

type Stream[T] interface {
    Next(ctx Context) (T, error)   // ErrEndOfStream terminates; ctx cancellation aborts
    Close() error                  // idempotent; cancels production, releases resources
}

// PortError is the in-process view of the ADR-016 error envelope.
type PortError struct {
    Code        string   // stable E-<AREA>-NNN
    Category    string
    Severity    string
    Message     string   // user-facing, redacted
    Detail      string   // technical, safe-to-log
    Retryable   bool
    CorrelationID ULID
    Cause       error
}
```

## Port ownership map

| Port | Implemented by | Primary consumers | Contract owner | Error family |
|---|---|---|---|---|
| ProviderPort | Provider Layer adapters | Agent Engine, Context Manager, Indexing Engine | Volume 5 | E-PROV |
| AuthPort | Authentication Layer | Provider Layer, CLI/TUI auth commands | Volume 5 | E-AUTH |
| ToolPort | Tool implementations (built-in, plugin, MCP-bridged) via Tool Runtime | Execution Engine, Agent Engine | Volume 6 | E-TOOL |
| TerminalPort | Terminal Engine | Tool Runtime (terminal tools), Git Engine excluded (uses PAL directly, ADR-025) | Volume 6 | E-TOOL |
| MemoryStorePort | Memory Manager over Persistence Layer | Agent Engine, Context Manager, CLI/TUI | Volume 7 | E-MEM |
| IndexerPort | Indexing Engine | Context Manager, Memory Manager, CLI/TUI | Volume 7 | E-IDX |
| EventBusPort | Event Bus | every component (publish); TUI, Observability, IPC bridge (subscribe) | Volume 10 (envelope, delivery) | E-OBS |
| PermissionPort | Permission Manager | Tool Runtime, Sandbox Engine, Git Engine, IPC server, all side-effecting paths | Volume 9 | E-SEC |
| SecretStorePort | Secret Store | Authentication Layer, Provider Layer, MCP Runtime | Volume 9 | E-SEC |
| SandboxPort | Sandbox Engine | Tool Runtime, Terminal Engine, Plugin Runtime | Volume 9 | E-SEC |
| ConfigPort | Configuration Manager | every component (via injection at composition) | Volume 10 | E-CFG |
| SessionStorePort | Persistence Layer | Runtime, Agent Engine, Workflow Engine, CLI/TUI | Volume 10 (storage; run semantics Volume 4) | E-CFG |
| GitPort | Git Engine | Tool Runtime (Git tools), Workspace Engine, CLI/TUI | Volume 11 | E-GIT |
| WorkspacePort | Workspace Engine | Runtime, CLI/TUI, Indexing Engine | Volume 4 | E-AGT |
| SchedulerPort | Task Scheduler | every component that runs concurrent work | Volume 3 (chapter 08); pool budgets Volume 12 | E-ARCH |
| UpdaterPort | Updater | CLI (update commands), scheduled checks | Volume 14 | E-REL |
| PackagePort | Package Manager | CLI (package commands), Updater (extension compatibility checks) | Volume 6 (extension packages); release artifacts Volume 14 | E-PLUG |
| TelemetryPort | Telemetry | Observability, engines via instrumentation helpers | Volume 10 | E-OBS |

## ProviderPort

The single contract every model provider adapter implements (Principle 1). One `ProviderPort`
value represents one configured Provider; model selection travels in requests. The Provider
Layer's router itself implements `ProviderPort`, so consumers are indifferent to routing and
fallback (Volume 5).

```pseudo
interface ProviderPort {
    Chat(ctx Context, req ChatRequest) (ChatResponse, error)
    ChatStream(ctx Context, req ChatRequest) (Stream[ChatEvent], error)
    Embed(ctx Context, req EmbedRequest) (EmbedResponse, error)
    DiscoverModels(ctx Context) ([]ModelDescriptor, error)
    Capabilities(ctx Context, model string) (CapabilitySet, error)
    CountTokens(ctx Context, req TokenCountRequest) (TokenCount, error)
}
```

Method semantics:

- `Chat` — one non-streaming inference turn: messages, tool declarations, and parameters in;
  a complete response (content parts, tool calls, usage accounting) out.
- `ChatStream` — the same request surface with incremental delivery. `ChatEvent` is a tagged
  union (content delta, tool-call delta, usage, terminal event); the stream ends with exactly
  one terminal event carrying final usage. Consumers cancel by context or `Close`.
- `Embed` — batch embedding for the Indexing Engine and Memory Manager; input batching limits
  are declared via `Capabilities`.
- `DiscoverModels` — enumerates the models this provider currently exposes, with external
  model identifiers and declared capabilities; used to populate the Model catalog (Volume 2).
- `Capabilities` — the declared `CapabilitySet` for one model (closed enum, Volume 5).
  Behavior keys off this, never off model names (Principle 2).
- `CountTokens` — counts tokens for content against a named model, exactly when the provider
  offers an official counting mechanism; otherwise returns the E-PROV "capability
  unavailable" error class so the Context Manager falls back to its estimation strategy
  (Volume 7).

Errors: E-PROV family (Volume 5): connectivity, authentication hand-off, rate limiting,
capability unavailability, malformed provider responses. Retryability per code; retry/fallback
*policy* lives above the adapter, in the Provider Layer router.

Concurrency and cancellation: adapters MUST support concurrent in-flight requests per
provider up to configured limits; cancelling the context MUST abort the underlying HTTP
request (ADR-019) and, for `ChatStream`, end the stream with a cancellation error. Usage
accounting for cancelled streams reports tokens consumed up to abort, when determinable.

## ToolPort

The uniform execution contract for every tool, regardless of origin (Principle 4). The Tool
Runtime mediates all access: agents never hold a `ToolPort` directly — the Execution Engine
requests invocation through the Tool Runtime, which performs validation, permission
evaluation, sandbox placement, and observability, then drives this port.

```pseudo
interface ToolPort {
    Describe(ctx Context) (ToolDescriptor, error)
    Validate(ctx Context, input JSON) (ValidationResult, error)
    Execute(ctx Context, req ToolExecuteRequest) (Stream[ToolEvent], error)
    Cancel(ctx Context, invocationID ULID) error
}
```

Method semantics:

- `Describe` — the full tool declaration of Principle 4: identity, version, input/output
  schemas, permission declaration, timeouts, limits, origin, trust level (contract in
  Volume 6). Stable for the lifetime of a registration; changes require re-registration.
- `Validate` — input validation against the input schema plus tool-specific semantic checks,
  without side effects. The Tool Runtime MUST call it (or an equivalent cached schema check)
  before `Execute`; tools MUST NOT rely on callers having done so.
- `Execute` — runs one Tool Invocation (`req` carries the invocation ULID, validated input,
  granted permission set, and effective limits). The stream emits ordered `ToolEvent` items:
  progress, partial output, log lines, artifacts references, and exactly one terminal event
  that becomes the Tool Result (`success` or `error`, Volume 2). Streaming exists so the TUI
  shows live tool activity (PRD-008).
- `Cancel` — best-effort cooperative cancellation of a running invocation by ID; the
  invocation still terminates through its stream with the `cancelled` outcome (Tool Invocation
  states, Volume 2 chapter 09).

Errors: E-TOOL family (Volume 6): validation failure, execution failure, timeout, limit
breach, sandbox refusal surfaced from E-SEC. A tool error is data (a Tool Result), not a
transport failure; `Execute` returns a Go-level error only when the invocation could not be
started or the stream broke.

Concurrency and cancellation: one `Execute` per invocation ULID; a tool declares its own
concurrency limit in `Describe` and the Tool Runtime enforces it. Context cancellation MUST
terminate the sandboxed process tree (chapter 07, Process Trees surface) within the teardown
budget Volume 6 sets.

## MemoryStorePort

The Memory Manager's public face: persistent memory across the layers defined by Volume 7
(session, workspace, long-term), backed by the databases of ADR-028.

```pseudo
interface MemoryStorePort {
    Ingest(ctx Context, records []MemoryRecordDraft) ([]ULID, error)
    Retrieve(ctx Context, q MemoryQuery) ([]MemoryRecord, error)
    Rank(ctx Context, q MemoryQuery, candidates []ULID) ([]RankedMemory, error)
    Expire(ctx Context, policy ExpirePolicy) (ExpireReport, error)
    Delete(ctx Context, ids []ULID) error
    Export(ctx Context, q MemoryQuery) (Stream[MemoryRecord], error)
}
```

Method semantics:

- `Ingest` — writes memory records with provenance and retention attributes; returns minted
  ULIDs. Idempotency and deduplication rules are Volume 7's.
- `Retrieve` — query by layer, scope, provenance, time, and content (lexical or semantic via
  IndexerPort under the hood); returns full records.
- `Rank` — scores an explicit candidate set against a query; exists so the Context Manager
  can re-rank a working set without re-querying.
- `Expire` — applies a retention policy pass; returns what was archived/expired (Memory
  Record `status` vocabulary, Volume 2 chapter 09). Invoked by retention scheduling (Volume 7)
  and manual commands.
- `Delete` — hard deletion honoring INV-MEM cascade rules (Volume 2); audit precedence
  applies (records of the deletion persist even when content is removed).
- `Export` — streams matching records as canonical JSON entity documents (Volume 2 chapter
  10 export forms), for user data portability.

Errors: E-MEM family (Volume 7). Concurrency: safe for concurrent readers/writers;
consistency per aggregate rules (Memory Record is its own aggregate). Cancellation aborts
scans mid-stream; `Ingest` is transactional per batch — a cancelled batch writes nothing.

## IndexerPort

Queryable indexes over workspace content — lexical and semantic (ADR-020) — with the Index
lifecycle frozen in Volume 2 chapter 09 (`created`, `building`, `ready`, `updating`, `stale`,
`failed`, `removed`).

```pseudo
interface IndexerPort {
    Build(ctx Context, spec IndexSpec) (ULID, error)
    Update(ctx Context, indexID ULID, changes []PathChange) error
    Query(ctx Context, indexID ULID, q IndexQuery) ([]IndexHit, error)
    Invalidate(ctx Context, indexID ULID, scope InvalidateScope) error
    Status(ctx Context, indexID ULID) (IndexStatus, error)
}
```

Method semantics:

- `Build` — declares and fully builds an index for a workspace scope; returns the Index ULID
  immediately after the metadata row exists, with progress observable via `Status` and
  `index.*` events. Long-running; runs as a supervised background task (SchedulerPort).
- `Update` — incremental update from explicit path changes (from Workspace Engine watches or
  File Change records); queryability during update follows Volume 7 rules.
- `Query` — lexical or semantic search with budgets (max results, max latency hint); hits
  carry paths, spans, scores, and index generation for staleness-aware consumers.
- `Invalidate` — marks a scope (paths or whole index) stale, forcing rebuild before those
  entries are served again; dropping the whole cache is always legal (INV-IDX-02: caches are
  rebuildable, never data loss).
- `Status` — current state, generation, coverage, and staleness metrics for one index.

Errors: E-IDX family (Volume 7). Concurrency: queries are concurrent; at most one build or
update per index at a time (enforced by the engine). Cancellation of `Build`/`Update` leaves
the index in a consistent prior generation or `failed` — never a torn generation.

## EventBusPort

The in-process typed publish/subscribe channel of ADR-012. The envelope, ordering, delivery
semantics, and per-family overflow policy are Volume 10's; the port fixes the shape.

```pseudo
interface EventBusPort {
    Publish(ctx Context, event Event) error
    Subscribe(ctx Context, sel TopicSelector, opts SubscribeOptions) (Subscription, error)
}

type Subscription interface {
    Events() Stream[Event]     // bounded per-subscriber buffer (ADR-012)
    Close() error
}
```

Method semantics:

- `Publish` — emits one enveloped event to its topic (name grammar
  `<area>.<noun>.<verb-past>`, Volume 0 chapter 03). Publishing MUST NOT block on slow
  subscribers: per-subscriber bounded buffers apply the family's declared overflow policy
  (drop-oldest, block, or coalesce — Volume 10).
- `Subscribe` — registers a subscriber for a topic selector (exact names or area prefixes)
  with options (buffer size within bounds, replay-from-persisted where the family supports
  it). The subscription ends when closed or when its context is cancelled.

Errors: E-OBS family (Volume 10); publishing to a closed bus and malformed envelopes are the
principal classes. Concurrency: fully concurrent; ordering is guaranteed per topic per
publisher, as Volume 10 specifies. Delivery is at-most-once in-process, with persistence of
Event records (Volume 2) providing the durable trail.

## PermissionPort

The single decision path for side-effecting actions (Principle 8). Decision semantics, the
permission enum, scopes, and policy interplay are Volume 9's.

```pseudo
interface PermissionPort {
    Check(ctx Context, req PermissionQuery) (Decision, error)
    Request(ctx Context, req PermissionRequest) (Decision, error)
    RecordDecision(ctx Context, rec DecisionRecord) error
}
```

Method semantics:

- `Check` — non-interactive evaluation of an action against standing grants and policies;
  never prompts; returns `allow`, `deny`, or `ask` guidance (vocabulary owned by Volume 9).
  Callers on non-interactive paths treat anything but `allow` as denial (PRD-009).
- `Request` — the full decision flow: evaluates as `Check`, and where interaction is
  permitted and required, raises an Approval (states per Volume 2 chapter 09) through the
  active driver, blocking on the user's decision or its expiry. Returns the final decision
  with the Approval ULID.
- `RecordDecision` — appends a decision record for flows where the decision was produced
  elsewhere (e.g., policy pre-resolution at run start); every decision — grant or denial —
  is persisted and auditable (PRD-006).

Errors: E-SEC family (Volume 9). Denial is a *decision*, not an error: it returns as a
`Decision` value; errors are reserved for evaluation failure. Concurrency: concurrent checks
are safe; `Request` serializes per subject where Volume 9 requires single outstanding
approvals. Cancellation of a pending `Request` cancels the Approval (`cancelled` terminal
state).

## SecretStorePort

Credential material behind references (ADR-014): OS keychains with the encrypted-file
fallback. Only references (`secret_ref`) cross other ports; only this port touches material.

```pseudo
interface SecretStorePort {
    Get(ctx Context, ref SecretRef) (SecretValue, error)
    Set(ctx Context, ref SecretRef, value SecretValue, meta SecretMeta) error
    Delete(ctx Context, ref SecretRef) error
    List(ctx Context, scope SecretScope) ([]SecretRef, error)
}
```

Method semantics:

- `Get` — resolves a reference to material; every access is audit-logged (Volume 9).
  `SecretValue` is a zeroize-on-release wrapper; callers MUST NOT persist or log it.
- `Set` — creates or replaces material under a reference with metadata (kind, provider
  association, creation time); replacement is how rotation lands (AuthPort drives it).
- `Delete` — removes material; the corresponding Credential row is tombstoned per Volume 2.
- `List` — enumerates references and metadata in scope; never returns material.

Errors: E-SEC family (Volume 9): backend unavailable (E-PORT-003 surfaces beneath it when the
platform lacks a keychain and the fallback is declined), reference not found, access denied
by the OS store. Concurrency: safe; the encrypted-file fallback serializes writes via PAL
File Locking. All methods are local-only — no network (Principle 3).

## SandboxPort

Isolation policy application for anything Andromeda executes (ADR-021). Policy content and
the containment model are Volume 9's; the layered mechanism (process-level controls at MVP,
OS-level isolation from Beta/v1) is fixed by ADR-021.

```pseudo
interface SandboxPort {
    Prepare(ctx Context, spec SandboxSpec) (SandboxHandle, error)
    ApplyPolicy(ctx Context, sb SandboxHandle, policy SandboxPolicy) error
    ExecuteIn(ctx Context, sb SandboxHandle, cmd CommandSpec) (ExecutionID, error)
    Teardown(ctx Context, sb SandboxHandle) error
}
```

Method semantics:

- `Prepare` — allocates an execution environment: working directory policy, filtered
  environment (deny-by-default passthrough, ADR-021), resource limit skeleton. Returns a
  handle usable for one or more executions per its spec.
- `ApplyPolicy` — binds or tightens the effective policy (path rules, network stance,
  resource limits, isolation mechanism selection). The **effective containment level is part
  of the handle's observable state** and is recorded per execution (ADR-021).
- `ExecuteIn` — starts a command inside the sandbox; execution I/O is then driven through
  TerminalPort using the returned `ExecutionID`. The launch path for tools, plugins, and
  terminal commands is exclusively this method — direct process spawning outside the Sandbox
  Engine is a defect (ADR-009 risk note).
- `Teardown` — terminates the full process tree (PAL Process Trees surface), releases
  resources, and finalizes the execution records. Idempotent.

Errors: E-SEC family (Volume 9): policy violation refusals, mechanism unavailability
(degradation to a weaker layer is explicit and observable, never silent), teardown failures.
Concurrency: handles are single-owner; executions within a handle may be concurrent when the
spec allows. Context cancellation triggers teardown semantics.

## ConfigPort

Configuration resolution with precedence, validation, and watching. Schema, precedence order,
and migration of configuration are Volume 10's single-home topics.

```pseudo
interface ConfigPort {
    Resolve(ctx Context, q ConfigQuery) (ResolvedConfig, error)
    Validate(ctx Context, doc ConfigDocument) (ValidationReport, error)
    Watch(ctx Context, sel ConfigSelector) (Stream[ConfigChange], error)
}
```

Method semantics:

- `Resolve` — produces the effective configuration for a scope (global, workspace, project,
  invocation) by applying Volume 10's precedence over `andromeda.toml` layers, environment
  (`ANDROMEDA_*`), and invocation flags. Every resolved value carries its **source
  attribution** (which layer supplied it) for diagnostics.
- `Validate` — validates a configuration document against the typed schema (ADR-008,
  ADR-024) without applying it; returns all findings, not just the first.
- `Watch` — change notification for a selector; emits the delta and the new resolved values.
  Consumers that cannot re-resolve live (most engines) receive changes at defined
  reconfiguration points owned by their volumes.

Errors: E-CFG family (Volume 10), mapped to exit code 3 at the CLI boundary. Concurrency:
reads are concurrent and cheap (cached resolution); watch streams follow Stream rules.
Resolution never blocks on the network.

## GitPort

The Git operations surface, implemented by the Git Engine per ADR-025 (system git ≥ 2.40
behind an encapsulated adapter). Operation semantics, output fidelity rules, and extended
hosting integrations are Volume 11's.

```pseudo
interface GitPort {
    Version(ctx Context) (GitVersion, error)
    Status(ctx Context, repo RepoRef) (RepoStatus, error)
    Diff(ctx Context, repo RepoRef, spec DiffSpec) (Stream[DiffHunk], error)
    Stage(ctx Context, repo RepoRef, paths []Path) error
    Unstage(ctx Context, repo RepoRef, paths []Path) error
    Commit(ctx Context, repo RepoRef, spec CommitSpec) (CommitID, error)
    Log(ctx Context, repo RepoRef, spec LogSpec) (Stream[CommitInfo], error)
    Show(ctx Context, repo RepoRef, rev Revision) (CommitDetail, error)
    ListBranches(ctx Context, repo RepoRef) ([]BranchInfo, error)
    CreateBranch(ctx Context, repo RepoRef, spec BranchSpec) error
    SwitchBranch(ctx Context, repo RepoRef, name string) error
    ApplyPatch(ctx Context, repo RepoRef, patch PatchDocument) (ApplyReport, error)
    WorktreeAdd(ctx Context, repo RepoRef, spec WorktreeSpec) (WorktreeInfo, error)
    WorktreeList(ctx Context, repo RepoRef) ([]WorktreeInfo, error)
    WorktreeRemove(ctx Context, repo RepoRef, path Path) error
}
```

Method semantics, grouped: `Version` detects and gates on the ADR-025 floor at startup.
`Status`/`Diff`/`Log`/`Show`/`ListBranches` are read-only queries parsed from porcelain and
NUL-terminated formats; `Diff` and `Log` stream because repositories are large. `Stage`,
`Unstage`, `Commit`, `CreateBranch`, `SwitchBranch`, `ApplyPatch`, and the worktree methods
mutate the repository and are side-effecting: the Tool Runtime and drivers MUST hold a
permission decision (PermissionPort) before invoking them, and each produces File Change /
Patch / Command Execution records per Volume 2 attribution invariants. `ApplyPatch` applies a
reviewed Patch (status vocabulary `proposed`→`applied`, Volume 2) atomically or not at all,
returning per-file results.

Errors: E-GIT family (Volume 11): missing/old git (configuration error), repository state
conflicts, apply rejections, subprocess failures. Concurrency: read methods are concurrent;
mutating methods on one repository serialize within the engine. Cancellation kills the git
subprocess and reports partial completion honestly.

## TerminalPort

PTY-backed command execution with streaming capture, signals, and resize — the Terminal
Engine's contract (Volume 6). All executions enter through SandboxPort (`ExecuteIn`); the
`ExecutionID` links the two ports.

```pseudo
interface TerminalPort {
    Execute(ctx Context, spec CommandSpec) (ExecutionID, error)   // convenience: Prepare+ExecuteIn with default policy
    Stream(ctx Context, id ExecutionID) (Stream[TerminalChunk], error)
    Write(ctx Context, id ExecutionID, input bytes) error
    Signal(ctx Context, id ExecutionID, sig SignalName) error
    Resize(ctx Context, id ExecutionID, cols int, rows int) error
    Wait(ctx Context, id ExecutionID) (CommandOutcome, error)
}
```

Method semantics:

- `Execute` — starts a command under the default sandbox policy for the calling context;
  callers needing custom policy use SandboxPort directly. PTY versus pipe mode is part of
  `CommandSpec`.
- `Stream` — ordered output chunks (stdout/stderr tagged, or merged PTY byte stream) with
  capture limits from the effective policy; truncation is marked, never silent.
- `Write` — sends input to the command's stdin/PTY (interactive tools, REPL-style flows).
- `Signal` — delivers a signal by portable name (`interrupt`, `terminate`, `kill`, …); the
  PAL maps names to platform mechanics (chapter 07, Signals surface).
- `Resize` — PTY window size change, propagated to the child.
- `Wait` — blocks until termination; returns the Command Execution outcome
  (`succeeded`, `failed`, `timed_out`, `killed` — Volume 2 recorded status) with exit code,
  signal, timing, and resource usage.

Errors: E-TOOL family (Volume 6). Concurrency: one writer at a time per execution; `Stream`
may have exactly one consumer; `Signal`/`Resize`/`Wait` are concurrent-safe. Context
cancellation of `Wait` does not kill the command; killing is explicit (`Signal`) or arrives
via sandbox teardown and timeout policy.

## WorkspacePort

Workspace and project discovery and lifecycle (behavior owned by Volume 4; entities by
Volume 2).

```pseudo
interface WorkspacePort {
    Discover(ctx Context, start Path) (WorkspaceCandidate, error)
    Open(ctx Context, root Path, opts OpenOptions) (WorkspaceHandle, error)
    Snapshot(ctx Context, ws WorkspaceHandle) (WorkspaceSnapshot, error)
    Close(ctx Context, ws WorkspaceHandle) error
}
```

Method semantics:

- `Discover` — walks upward from a starting path to locate the workspace root (`.andromeda/`
  marker, repository root heuristics per Volume 4), reporting what it found without opening.
- `Open` — opens (creating on first use, when options say so) the workspace: initializes or
  attaches `.andromeda/`, opens the workspace database (ADR-028), registers the workspace in
  the global registry, and yields the handle other components receive.
- `Snapshot` — a consistent, read-only description of workspace state at a point in time:
  project layout, active configuration profile references, index generations, VCS summary —
  the input for run reproducibility (SM-12) and context assembly.
- `Close` — detaches cleanly: flushes state, releases locks (PAL File Locking), emits the
  closing event.

Errors: E-AGT family (Volume 4): not a workspace, unreadable root, database open/migration
failures surfaced from beneath (integrity failures keep exit code 9 semantics per ADR-029).
Concurrency: multiple concurrent `andromeda` processes against one workspace are governed by
the single-writer rules Volume 10 defines over the workspace database; `Open` fails cleanly
when exclusivity rules would be violated.

## SessionStorePort

Persistence and restoration of Sessions and Runs — the durability behind PRD-010. Storage
mechanics are Volume 10's; run/turn semantics are Volume 4's.

```pseudo
interface SessionStorePort {
    SaveSession(ctx Context, s SessionSnapshot) error
    LoadSession(ctx Context, id ULID) (SessionSnapshot, error)
    ListSessions(ctx Context, f SessionFilter) ([]SessionSummary, error)
    AppendRunRecords(ctx Context, runID ULID, batch []RunRecord) error
    LoadRun(ctx Context, id ULID) (RunSnapshot, error)
    ListRuns(ctx Context, f RunFilter) ([]RunSummary, error)
    MarkInterrupted(ctx Context, scope InterruptScope) ([]ULID, error)
}
```

Method semantics:

- `SaveSession`/`LoadSession`/`ListSessions` — session rows and their state transitions
  (canonical Session states, Volume 2 chapter 09), with optimistic concurrency via `revision`.
- `AppendRunRecords` — the incremental persistence path: turns, messages, plan/task
  transitions, tool invocations and results, file changes — appended transactionally per
  batch, at minimum at every canonical state transition and Record append (Volume 2 chapter
  10 write discipline). This is the hot write path; its latency budget is Volume 12's.
- `LoadRun` — reconstructs a run's full state for resumption or inspection: current states,
  record history, accounting.
- `ListRuns` — filtered summaries (by session, state, time) for CLI/TUI listings.
- `MarkInterrupted` — crash recovery support (chapter 08): atomically marks every
  non-terminal run/task in scope as `interrupted` (frozen state names) and returns them.
  Interrupted work is never assumed complete (PRD-010).

Errors: E-CFG family (storage, Volume 10); corruption and migration failures follow ADR-029
(exit code 9). Concurrency: appends for one run serialize per the run's single-writer
discipline (Volume 4); reads are concurrent. Cancelled batch appends write nothing.

## SchedulerPort

Supervised concurrency — the ADR-023 model as a contract. The full behavioral contract lives
in chapter [08](08-processes-concurrency-ipc.md) of this volume; pool sizes and shed policies
are budgeted by Volume 12.

```pseudo
interface SchedulerPort {
    Submit(ctx Context, spec TaskSpec) (TaskHandle, error)
    NewGroup(ctx Context, spec GroupSpec) (TaskGroup, error)
    Cancel(ctx Context, id SchedTaskID, reason CancelReason) error
    Stats(ctx Context) (SchedulerStats, error)
}

type TaskGroup interface {
    Go(ctx Context, spec TaskSpec) (TaskHandle, error)
    Wait(ctx Context) error          // first-error propagation, joined shutdown (errgroup semantics)
    Cancel(reason CancelReason)
}

type TaskHandle interface {
    Await(ctx Context) (TaskOutcome, error)
    ID() SchedTaskID
}
```

Method semantics:

- `Submit` — schedules one supervised unit of work onto a named pool (`spec.Pool`); the
  scheduler owns panic capture, lifecycle events, and cancellation wiring. When the pool's
  queue bound is reached, behavior follows the pool's declared policy: block with deadline or
  reject with E-ARCH-005.
- `NewGroup` — creates a structured group bound to a parent context: first error cancels the
  group; `Wait` joins all members. Groups nest; a run's entire concurrent footprint hangs off
  one root group (chapter 08).
- `Cancel` — cancels a task or group subtree by ID with a recorded reason (user interrupt,
  timeout, budget, shutdown).
- `Stats` — pool occupancy, queue depths, task counts by state — the observability surface
  for NFR-ARCH-004 and Volume 12 saturation metrics.

Errors: E-ARCH family (this volume, chapter 08): submission rejection (E-ARCH-005), unknown
pool, scheduler shut down. Concurrency: the port *is* the concurrency model; note
`SchedTaskID` identifies scheduler work items, distinct from the domain Task entity.

## UpdaterPort

Self-update of the installed product (MVP item 23). Full update semantics, channels, and the
Update process machine (frozen states, Volume 2 chapter 09) are Volume 14's.

```pseudo
interface UpdaterPort {
    Check(ctx Context) (UpdateCheckResult, error)
    Download(ctx Context, rel ReleaseRef) (Stream[DownloadProgress], error)
    Verify(ctx Context, rel ReleaseRef) (VerificationReport, error)
    Apply(ctx Context, rel ReleaseRef) (ApplyReport, error)
    Rollback(ctx Context) (RollbackReport, error)
}
```

Method semantics: `Check` queries release metadata for the configured channel and reports
`up_to_date` or `update_available` (frozen Update states); it is the only method that
requires network and MUST fail cleanly offline. `Download` fetches artifacts with resumable
progress; `Verify` validates checksums and — when signing is enabled per the Volume 1 signing
viability note — signatures and provenance (ADR-013); `Apply` swaps the installed binary via
the PAL Installer/Updater surfaces with atomic replace-or-restore semantics; `Rollback`
restores the previously retained version offline (SM-19). `Apply` MUST refuse to run when
`Verify` has not passed for the same artifact set.

Errors: E-REL family (Volume 14). Concurrency: at most one update operation machine-wide
(PAL File Locking on the update lock); `Check` is concurrent-safe. Cancellation between
`Apply` steps never leaves a half-replaced installation (the PAL surface contract).

## PackagePort

Extension package management: resolve, install, verify, remove — the Package installation
machine (frozen states, Volume 2 chapter 09) is owned by Volume 6.

```pseudo
interface PackagePort {
    Resolve(ctx Context, req PackageRequest) (ResolutionPlan, error)
    Install(ctx Context, plan ResolutionPlan) (Stream[InstallEvent], error)
    Verify(ctx Context, pkg PackageRef) (VerificationReport, error)
    Remove(ctx Context, pkg PackageRef, opts RemoveOptions) (RemoveReport, error)
}
```

Method semantics: `Resolve` turns a package request (name, version constraint, source) into a
concrete plan — versions, dependencies, download sources, checksums — without side effects.
`Install` executes a plan through the frozen installation states (`resolving` …
`installed`), streaming progress events; failure at any step leaves nothing partially active
(`failed`/`rolled_back` states). `Verify` re-checks an installed package's integrity
(checksums, signature state, manifest consistency). `Remove` uninstalls and tombstones the
Package row (`removed`), deregistering its Extensions.

Errors: E-PLUG family (Volume 6): resolution conflicts, verification failures, registry
errors. Concurrency: one mutating operation per package at a time; installs of independent
packages may proceed concurrently under scheduler pools. Trust and signature policy come from
Volume 9 via the Policy Engine.

## AuthPort

Provider authentication flows: acquisition, refresh, revocation, and rotation of credentials
— always via official mechanisms (Volume 1 constraint). Flow semantics are Volume 5's;
storage is Volume 9's through SecretStorePort.

```pseudo
interface AuthPort {
    Authenticate(ctx Context, spec AuthSpec) (AuthenticationHandle, error)
    Refresh(ctx Context, h AuthenticationHandle) (AuthenticationHandle, error)
    Revoke(ctx Context, h AuthenticationHandle) error
    Rotate(ctx Context, credentialID ULID) (RotationReport, error)
    ListProfiles(ctx Context) ([]AuthProfile, error)
}
```

Method semantics:

- `Authenticate` — runs the flow for a provider's official mechanism (API key intake, OAuth
  device/browser flows, local-server no-auth): drives the Authentication Session machine
  (`unauthenticated` → `authenticating` → `active`, frozen names) and stores material behind
  Secret Store references only.
- `Refresh` — renews token material (`refreshing` state) before or upon expiry; returns the
  updated handle. Adapters call it through the Authentication Layer, never directly against
  provider endpoints.
- `Revoke` — invalidates the session (`revoked`, terminal) and the provider-side grant where
  the official API supports it.
- `Rotate` — replaces a Credential's material (new key/token), updating dependent
  Authentication Sessions per Volume 5 flow rules; the Credential `status` vocabulary
  (`active`, `rotated`, `revoked`, `expired`) records the outcome.
- `ListProfiles` — enumerates configured authentication profiles (named credential+provider
  bindings) without any secret material.

Errors: E-AUTH family (Volume 5), mapped to exit code 4 at the CLI boundary. Concurrency:
refreshes for one Authentication Session serialize (single-flight); flows requiring user
interaction integrate with drivers the same way Approvals do. Secrets never appear in logs,
events, or errors (Volume 9 redaction).

## TelemetryPort

Metrics, traces, and spans under ADR-011 (OpenTelemetry, local-first sinks, consent-gated
export — consent policy owned by Volume 10).

```pseudo
interface TelemetryPort {
    EmitMetric(ctx Context, sample MetricSample) error
    StartSpan(ctx Context, name string, attrs Attributes) (Context, Span)
    Flush(ctx Context) error
}

type Span interface {
    SetAttribute(key string, value AttrValue)
    RecordError(err error)
    End(status SpanStatus)
}
```

Method semantics: `EmitMetric` records one sample against a registered Metric definition
(Volume 2); unknown metric names are defects surfaced as E-OBS errors in development builds
and dropped-with-counter in release builds. `StartSpan` opens a span nested under the span in
`ctx` and returns the derived context — this is how the Trace tree (one per run) is built;
span `status` values mirror the Trace recorded-status vocabulary (`ok`, `error`,
`interrupted`). `Flush` drains buffered telemetry to local sinks (and to OTLP export only
under recorded consent), bounded by the context deadline; shutdown calls it with the chapter
08 shutdown budget.

Errors: E-OBS family (Volume 10). Concurrency: all methods are concurrent-safe and
non-blocking on the hot path (bounded buffers, shed-with-counter on overflow); telemetry
failure MUST never fail the operation being observed.

## Requirements

### FR-ARCH-003 — Port name and signature freeze

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Core
- Source: Design
- Owner: Architecture (Volume 3)
- Affected components: all components; Extension SDK; Volumes 4–14 as contract elaborators
- Dependencies: FR-ARCH-002; ADR-015, ADR-030
- Related risks: RISK-ARCH-002

#### Description

The 18 port interfaces of this chapter — their names, method names, parameter and result
arities, and streaming shapes — are frozen for the corpus and the implementation. Owning
volumes MUST elaborate behavioral contracts under exactly these names and MUST NOT rename,
re-sign, split, or merge ports or methods. Changes to this chapter occur only through the
Volume 0 change procedure and, after MVP exit, additionally follow the public-contract
stability regime (ADR-015, SM-20): additive changes in minor releases, breaking changes only
in a major release after a deprecation window.

#### Motivation

Eleven volumes and the implementation are authored in parallel against these interfaces. A
rename or signature change after parallel authoring begins invalidates cross-volume
references silently — the corpus equivalent of an ABI break.

#### Actors

Volume authors; implementers; SDK consumers; the spec linter and contract-diff tooling.

#### Preconditions

This chapter is authored and lint-clean; owning volumes reference it.

#### Main flow

1. An owning volume specifies behavior for its port using the frozen names.
2. Implementation declares the interfaces in `internal/ports` textually consistent with the
   pseudocode here.
3. Contract tests (Volume 13) verify each adapter against the port contract.

#### Alternative flows

- A genuine contract gap is found: the owning volume proposes an additive method or type
  field through the change procedure; this chapter is amended; downstream volumes are
  notified via the consolidation process.

#### Edge cases

- A port method an owning volume decides not to support in a phase (e.g., OAuth-dependent
  flows PENDING VALIDATION per ADR-010) still exists in the interface and returns the
  declared "unsupported" error class for that family — presence is frozen even when behavior
  is phased.
- Contract *types* (request/response structs) may gain optional fields additively; required
  fields follow the breaking-change regime.

#### Inputs

Port definitions (this chapter); change proposals.

#### Outputs

Stable `internal/ports` package and `sdk/` mirrors; contract-diff reports per release.

#### States

Not applicable — the requirement governs contract artifacts, not runtime state.

#### Errors

Not applicable at runtime; violations are specification or CI defects.

#### Constraints

ADR-015 SemVer discipline; SM-20 targets (0 breaking changes outside a major release).

#### Security

Frozen contracts prevent security-relevant drift: PermissionPort and SandboxPort mediation
points cannot be quietly bypassed by re-signing methods without the change procedure.

#### Observability

Contract-diff tooling in CI compares released contract schemas per SM-20's measurement
method.

#### Performance

Not applicable.

#### Compatibility

This requirement *is* the compatibility anchor for extensions and later volumes; see chapter
09 for the versioning strategy it plugs into.

#### Acceptance criteria

- Given any Volume 4–14 chapter, when it references a port or method, then the name matches
  this chapter exactly (spec-lint cross-reference discipline; consolidation audit).
- Given two consecutive releases within a major line, when contract-diff runs over
  `internal/ports` and `sdk/`, then no removed or re-signed port method exists.
- Given a proposed breaking port change without a major-release vehicle, when review applies
  this requirement, then the change is rejected or re-shaped as additive.
- Negative case: given an adapter implementing a stale signature, when contract tests run,
  then compilation or the contract suite fails — drift cannot ship silently.

#### Verification method

Spec-lint name checks and consolidation audit (corpus side); compilation, contract test
suites, and contract-diff tooling in CI (implementation side); release audit per Volume 14.

#### Traceability

PRD-007; SM-20; ADR-015; FR-ARCH-002; RISK-ARCH-002.

### FR-ARCH-004 — Context propagation and cancellation on all ports

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Core
- Source: Design
- Owner: Architecture (Volume 3)
- Affected components: all port implementations and consumers; Task Scheduler
- Dependencies: ADR-023; FR-ARCH-003
- Related risks: RISK-ARCH-003

#### Description

Every port method takes a `Context` as its first parameter. Implementations MUST honor
cancellation and deadlines end-to-end: cancelling a context MUST abort the underlying
operation (HTTP request, subprocess, database statement, stream production), release its
resources, and return an error in the cancellation class of the port's error family.
Consumers MUST derive request contexts from their run/task supervision tree (chapter 08) so
that user interrupts, timeouts, and budget cancellations reach every in-flight operation.

#### Motivation

Cancellation is user-visible product behavior: interrupting a run (UC-14, exit code 8
semantics per ADR-016) must actually stop provider requests, tool subprocesses, and index
builds. A single port that ignores its context reintroduces orphaned work and unbounded
resource use (ADR-023 problem statement).

#### Actors

All engines and adapters; Task Scheduler; users triggering interrupts.

#### Preconditions

The supervision tree of chapter 08 provides parent contexts.

#### Main flow

1. A consumer calls a port method with a context derived from its task group.
2. The adapter starts the operation, wiring cancellation to the underlying mechanism.
3. On normal completion, results return and resources release.

#### Alternative flows

- Cancellation: the context is cancelled mid-operation → the adapter aborts (kills the
  subprocess via teardown, cancels the HTTP request, closes the stream), releases resources,
  and returns the cancellation error class; partial effects are reported honestly per the
  port's semantics.
- Deadline: identical handling with the timeout error class; exit code 8 at the CLI boundary.

#### Edge cases

- Non-cancellable critical sections (SQLite commit in flight, update `Apply` step) MUST be
  short and bounded; the method returns as soon as the section ends and MUST NOT begin a new
  section after cancellation.
- Streams: producer-side cancellation surfaces on the consumer's next `Next` call; `Close`
  after cancellation remains idempotent.

#### Inputs

Contexts with cancellation/deadline; port method arguments.

#### Outputs

Normal results, or cancellation/timeout errors in the port's family.

#### States

Cancelled work lands in the frozen cancellation outcomes of the affected entities
(`cancelled` for runs/tasks/tool invocations; `killed` for command executions) — never
in success states.

#### Errors

Each port family's cancellation and timeout classes; E-ARCH-005 for scheduler-level
rejection.

#### Constraints

No port method may block indefinitely ignoring its context; polling loops MUST select on
context doneness.

#### Security

Cancellation MUST terminate sandboxed process trees (SandboxPort teardown) — abandoned
children retaining granted permissions would be a containment breach.

#### Observability

Cancellations emit the affected entities' transition events with the recorded reason;
scheduler stats expose cancellation counts.

#### Performance

Cancellation-to-quiescence latency budgets are Volume 12's; the structural requirement here
is that the path exists everywhere.

#### Compatibility

Uniform across platforms; PAL maps process termination per OS (chapter 07).

#### Acceptance criteria

- Given a run with an in-flight provider stream, a running sandboxed tool, and a background
  index update, when the user interrupts the run, then all three terminate through their
  cancellation classes, their entities record `cancelled`/`killed` outcomes, and no child
  process of the run's tree survives.
- Given a port method invoked with an already-cancelled context, when called, then it returns
  the cancellation error without starting side effects.
- Given Volume 13 leak tests, when a component shuts down after heavy cancellation load, then
  zero goroutines and zero child processes leak (NFR-ARCH-004).
- Observability case: given a cancelled tool invocation, when its records are inspected, then
  the transition event carries the cancellation reason and correlation IDs per Principle 9.

#### Verification method

Contract test suites include cancellation and pre-cancelled-context cases for every port
method; fault-injection and leak tests per Volume 13; SM-11 crash/interrupt suites.

#### Traceability

PRD-005, PRD-008, PRD-010; ADR-023, ADR-016; FR-ARCH-006; NFR-ARCH-004.

### NFR-ARCH-002 — Port contract stability

- Category: Maintainability
- Priority: P0
- Phase: Beta
- Metric: Breaking changes to `internal/ports` / `sdk/` contract surfaces shipped outside a major release; deprecation-window compliance for announced breaks
- Target: 0 breaking changes outside a major release; 100% of breaks preceded by ≥ 1 minor release of deprecation
- Minimum threshold: same as target (this is SM-20's regime applied to the port layer; no tolerance)
- Measurement method: contract-diff tooling in CI comparing exported port/type signatures between releases; release audit per Volume 14
- Test environment: CI on release candidates
- Measurement frequency: every release; audited at phase gates
- Owner: Architecture (Volume 3) / Volume 14 (release audit)
- Dependencies: FR-ARCH-003; ADR-015
- Risks: RISK-ARCH-002
- Acceptance criteria: Contract-diff report for each release shows zero removed/re-signed port methods within a major line; any additive change is documented in release notes with its owning volume's contract reference.

## Risks

### RISK-ARCH-002 — Port interface churn during parallel authoring

- Category: Process / technical
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: FR-ARCH-003 freeze with change-procedure-only amendments; owning volumes elaborate semantics without touching signatures; consolidation phase reconciles residual friction as recorded amendments rather than ad-hoc edits
- Detection: spec-lint cross-reference checks; consolidation audit comparing volume references against this chapter; contract-diff in CI once implementation exists
- Owner: Architecture (Volume 3)
- Status: Open

Eleven volumes elaborate these ports concurrently. The plausible failure is quiet divergence:
a volume "improves" a method name or return shape locally, and the corpus forks. The freeze
plus mechanical reference checking turns divergence into a lint error instead of a discovery
during implementation.
