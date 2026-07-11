# 05 — Components: Infrastructure

Infrastructure components implement the ports the engines consume and enforce the safety
boundaries the product promises. Table conventions are those of chapter
[03](03-components-core.md); universal dependencies (Core Domain, injected ports, Logging,
TelemetryPort, EventBusPort, SchedulerPort) are not restated. Except for Event Bus, Task
Scheduler, Permission Manager, and Policy Engine (L2 per the chapter 01 classification), the
components here are L3 adapters.

## Configuration Manager

| Aspect | Specification |
|---|---|
| Responsibility | Implementing ConfigPort: loading `andromeda.toml` layers, environment mapping (`ANDROMEDA_*`), precedence resolution, typed validation (ADR-008, ADR-024), source attribution, watching, and configuration migrations. Behavior owner: Volume 10 |
| Boundaries | Owns the mechanics of configuration; the schema *content* of each TOML table belongs to its area owner (Volume 0 chapter 03); does not act on values — consumers do |
| Public API | Implements ConfigPort; consumed by every component via composition |
| Internal API | Layer loaders, precedence merger, schema registry into which area owners contribute table schemas |
| Allowed dependencies | PAL (Config Directories, Filesystem, File Locking surfaces); go-toml/v2, jsonschema/v6 (pinned) |
| Prohibited dependencies | Engines; other adapters; viper (explicitly rejected by ADR-005) |
| Inputs | Configuration files, environment variables, invocation flags relayed by drivers, watch triggers |
| Outputs | Resolved configuration with per-value source attribution; validation reports; change notifications |
| Events emitted | `config.*` family (Volume 10 mints names) |
| Errors | E-CFG family (Volume 10); exit code 3 at the CLI boundary |
| States | Stateless service over versioned configuration documents |
| Persistence | Reads global config dir + `.andromeda/` project layer (ADR-022); writes only for migrations and explicit set commands, atomically via PAL |
| Concurrency | Immutable resolved snapshots; watches deliver new snapshots — consumers never observe torn config |
| Security | Config files may name secret references, never secrets; permission-relevant keys are policy-validated (Volume 9); source attribution defeats config-injection ambiguity |
| Observability | Resolution diagnostics show every value's source (ADR-022 risk mitigation); validation findings are complete, not first-error |
| Testing | Precedence property tests; golden resolution fixtures per platform; invalid-config suites (`toml invalid` examples per Volume 0) |
| Extensibility | Area owners extend the schema registry; no third-party config sources at MVP |
| Phase | Core |

Configuration is resolved once per invocation into an immutable snapshot with attribution —
"which file/env/flag set this" is queryable, which converts most support questions into
diagnostics output. N-1 configuration compatibility with migrations is part of the chapter
[09](09-deployment-update-extensibility-compatibility.md) strategy.

## Persistence Layer

| Aspect | Specification |
|---|---|
| Responsibility | All SQLite access (ADR-007): schema DDL per Volume 2 conventions, transactions with aggregate discipline, repositories per entity, migrations per ADR-029, backup hooks, implementing SessionStorePort. Operational behavior owner: Volume 10 |
| Boundaries | The only component that speaks SQL; enforces the Volume 2 write discipline (one aggregate, one transaction; incremental persistence); does not interpret domain semantics beyond mapping |
| Public API | Implements SessionStorePort; repository API consumed by Memory Manager, Audit Log, catalog owners; migration/backup API for the composition root and CLI |
| Internal API | Connection pools per database (workspace/global/index caches per ADR-028), statement builders, migration runner |
| Allowed dependencies | modernc.org/sqlite (pinned, ADR-007); PAL (Filesystem, File Locking, Config Directories) |
| Prohibited dependencies | Engines; all other adapters; any component embedding SQL against it (Volume 2 write discipline rule 1) |
| Inputs | Domain values to persist; queries; migration triggers |
| Outputs | Persisted rows, query results, migration/backup reports, optimistic-concurrency conflicts |
| Events emitted | `storage.*` family (Volume 10 mints names) |
| Errors | Storage classes of E-CFG (Volume 10); integrity failures follow ADR-029 with exit code 9 |
| States | Stateless service; carries each database's schema version (`user_version` + `schema_migrations`) |
| Persistence | Is persistence: WAL mode, foreign keys on, per-connection pragmas per ADR-007/Volume 2 chapter 10 |
| Concurrency | WAL single-writer/multi-reader per database; write serialization per aggregate; `revision`-based optimistic concurrency with retry semantics per Volume 10 |
| Security | Databases live under user-owned directories with restrictive permissions (PAL); no credential material ever in the workspace DB (ADR-028 rule 2); redaction applied before persisting where Volume 9 gates |
| Observability | Query/transaction timing metrics; migration events; WAL checkpoint stats for Volume 12 budgets |
| Testing | Migration up-chain tests from every historical schema; crash-consistency tests (kill during commit, SM-11); driver-assumption conformance tests (Volume 2 register assumption) |
| Extensibility | Storage is a named extension surface (Principle 6) *beneath ports* — alternative backends implement the same ports; the SQL core itself is not pluggable |
| Phase | Core |

The layer's non-negotiable rule is exclusivity: every byte of authoritative state flows
through it, which is what makes the ADR-029 migration procedure, the ADR-028 topology, and
crash-consistency guarantees (PRD-010) enforceable in one place. Its repositories accept and
return Core Domain types, keeping SQL invisible above L3.

## Event Bus

| Aspect | Specification |
|---|---|
| Responsibility | Implementing EventBusPort: in-process typed publish/subscribe with per-subscriber bounded buffers and declared per-family overflow policy (ADR-012). Envelope/delivery contract owner: Volume 10 |
| Boundaries | Distribution only: no persistence (Event records are written by publishers via the Persistence Layer), no cross-process transport (the IPC bridge subscribes like any consumer) |
| Public API | Implements EventBusPort; consumed corpus-wide |
| Internal API | Topic registry with Go-typed payload bindings; buffer/overflow machinery |
| Allowed dependencies | None beyond the universal set — pure channels (ADR-012); L2 |
| Prohibited dependencies | External brokers (prohibited by ADR-012); adapters; PAL |
| Inputs | Enveloped events; subscriptions |
| Outputs | Delivered events; overflow/drop counters |
| Events emitted | Meta-events about itself minted by Volume 10 (overflow signaling per family) |
| Errors | E-OBS family (Volume 10) |
| States | Stateless besides subscription tables |
| Persistence | None (see Boundaries) |
| Concurrency | Publishers never block on slow subscribers (bounded buffers + policy); ordering per topic per publisher; shutdown drains per chapter 08 ordering |
| Security | In-process only; events crossing IPC pass Volume 9/10 redaction at the bridge |
| Observability | Per-topic throughput, buffer occupancy, drops by policy — the inputs to ADR-012's backpressure risk detection |
| Testing | Race and load tests per ADR-017; overflow-policy property tests; leak tests on subscription churn |
| Extensibility | New topics/payloads by area owners through the typed registry; no dynamic third-party topics in-process (plugins receive bridged events per Volume 6) |
| Phase | Core |

## Task Scheduler

| Aspect | Specification |
|---|---|
| Responsibility | Implementing SchedulerPort: supervised execution of all concurrent work — named bounded pools, structured groups with errgroup semantics, panic capture, cancellation wiring, lifecycle observability (ADR-023). Contract owner: this volume (chapter 08); pool budgets: Volume 12 |
| Boundaries | Owns goroutines; owns no domain logic. The scheduler's work items (`SchedTaskID`) are distinct from the domain Task entity |
| Public API | Implements SchedulerPort; consumed by every component running concurrent work |
| Internal API | Pool implementations, supervision tree bookkeeping |
| Allowed dependencies | errgroup (the sanctioned L2 exception); nothing else beyond the universal set |
| Prohibited dependencies | Everything else; conversely, components MUST NOT spawn goroutines outside it (ADR-023 "no naked goroutines") |
| Inputs | Task specs, group specs, cancellations, pool configuration |
| Outputs | Task outcomes, panics-as-errors with stacks, stats |
| Events emitted | `scheduler.*` family (this volume mints names; see chapter 08) |
| Errors | E-ARCH family: E-ARCH-005 on bounded-pool rejection (chapter 08) |
| States | Work items: submitted → running → terminal (outcome vocabulary in chapter 08); not a Volume 2 entity machine |
| Persistence | None — supervision state is process-local; durable work state belongs to the entities the work advances |
| Concurrency | Is the concurrency model; unbounded queues prohibited (ADR-023) |
| Security | Panic containment prevents one run's failure from killing the process; per-pool bounds prevent resource-exhaustion denial of service from any single feature |
| Observability | The `Stats` surface; per-pool saturation metrics feeding Volume 12; task lifecycle events with correlation IDs |
| Testing | Leak checks (zero goroutines after shutdown, NFR-ARCH-004); cancellation storm tests; deadlock detection under cyclic-load fixtures |
| Extensibility | Pool set and bounds are configuration (Volume 12 budgets); no third-party schedulers |
| Phase | Core |

## Indexing Engine

| Aspect | Specification |
|---|---|
| Responsibility | Implementing IndexerPort: lexical and semantic index build/update/query over workspace content, embedding storage per ADR-020, staleness tracking. Behavior owner: Volume 7 |
| Boundaries | Owns index structures and the index cache DB; obtains embeddings via ProviderPort `Embed` (through a Volume 7-defined acquisition path that respects offline operation); does not decide relevance for context (Context Manager ranks) |
| Public API | Implements IndexerPort; consumed by Context Manager, Memory Manager, drivers (search commands) |
| Internal API | Chunkers, lexical index structures, cosine-similarity search (ADR-020), watcher integration |
| Allowed dependencies | Persistence Layer (index metadata rows; cache DB access), WorkspacePort (content enumeration), ProviderPort (embeddings only), PAL Filesystem |
| Prohibited dependencies | Engines other than via ports; provider adapters directly |
| Inputs | Workspace content, path change notifications, queries, invalidations |
| Outputs | Index hits with spans/scores/generations; status reports; embeddings in `.andromeda/index.db` |
| Events emitted | `index.*` family (Volume 7 mints names) |
| Errors | E-IDX family (Volume 7) |
| States | Drives the Index machine (`created`, `building`, `ready`, `updating`, `stale`, `failed`, `removed` — frozen names) |
| Persistence | Index metadata in the workspace DB; index data and embeddings in the rebuildable cache DB — corruption triggers rebuild, never exit code 9 (ADR-028 rule 4, INV-IDX-02) |
| Concurrency | Builds/updates as supervised background tasks, one mutator per index; queries concurrent against the last consistent generation |
| Security | Indexes only what workspace read permissions allow; respects ignore rules (Volume 7); embeddings of sensitive paths excluded per policy |
| Observability | Build/update progress events; staleness and coverage metrics; query latency metrics against Volume 12 budgets |
| Testing | Generation-consistency tests under concurrent update+query; rebuild-from-nothing equivalence tests; ADR-020 scale benchmarks (≤ 100k chunks) |
| Extensibility | Indexers are a named extension surface (Principle 6): alternative implementations register beneath IndexerPort per Volume 7 |
| Phase | MVP (lexical and semantic per ADR-020) |

## Workspace Engine

| Aspect | Specification |
|---|---|
| Responsibility | Implementing WorkspacePort: workspace/project discovery, `.andromeda/` initialization, workspace database attachment, snapshots, file watching, close/detach. Behavior owner: Volume 4 (entities: Volume 2) |
| Boundaries | Owns the workspace boundary and its state directory; delegates VCS awareness to GitPort and persistence to the Persistence Layer |
| Public API | Implements WorkspacePort; watch feed consumed by Indexing Engine; consumed by Runtime and drivers |
| Internal API | Root-detection heuristics, `.andromeda/` layout manager, watcher abstraction over PAL |
| Allowed dependencies | Persistence Layer, GitPort (repository detection/summary), ConfigPort, PAL (Filesystem, Paths, File Locking) |
| Prohibited dependencies | Engines; provider adapters; Indexing Engine (it consumes the watch feed, not vice versa) |
| Inputs | Paths, open options, filesystem events |
| Outputs | Workspace handles, snapshots, change feeds, registry updates in the global DB |
| Events emitted | `workspace.*` family (Volume 4 mints names) |
| Errors | E-AGT family (Volume 4) |
| States | Workspace/Project are stateless catalog entities; open/attached state is process-local |
| Persistence | Workspace identity row + projects in `.andromeda/state.db`; machine-wide registry in the global DB (ADR-028) |
| Concurrency | Concurrent opens governed by Volume 10 single-writer database rules; watches deliver through bounded subscriptions |
| Security | Path policy anchor: the workspace root is the default permission scope boundary (Volume 9); symlink escapes handled per PAL Paths surface rules |
| Observability | Open/close/snapshot events; watch lag metrics |
| Testing | Discovery fixtures (nested repos, worktrees, symlinks); concurrent-open contention tests; snapshot determinism tests |
| Extensibility | Discovery heuristics configurable (Volume 4); no third-party workspace types at MVP |
| Phase | MVP |

## Git Engine

| Aspect | Specification |
|---|---|
| Responsibility | Implementing GitPort by shelling out to system git ≥ 2.40 behind one encapsulated adapter (ADR-025): subprocess invocation, environment filtering, porcelain/NUL-format parsing, timeout/cancellation, error mapping. Behavior owner: Volume 11 |
| Boundaries | The only component that invokes git or parses git output (ADR-025); repository *policy* (what may be committed when) belongs to callers and Volume 11 |
| Public API | Implements GitPort; consumed by Tool Runtime (Git tools), Workspace Engine, drivers |
| Internal API | git subprocess harness (argv building, version gating), format parsers |
| Allowed dependencies | PAL (Processes, Process Trees, Filesystem, Temporary Files); go-git only for individually PENDING-VALIDATION read-only fast paths per ADR-025 |
| Prohibited dependencies | TerminalPort (its subprocess needs are direct PAL, not PTY); engines; other adapters |
| Inputs | GitPort requests; git stdout/stderr; user git config/hooks (honored by construction, ADR-025) |
| Outputs | Parsed statuses/diffs/logs; commit IDs; File Change/Patch-relevant data for run records |
| Events emitted | `git.*` family (Volume 11 mints names) |
| Errors | E-GIT family (Volume 11); version-floor failures as configuration errors (exit code 3) |
| States | Stateless per operation |
| Persistence | None of its own; results persist via callers into run records |
| Concurrency | Read operations concurrent; mutations per repository serialized; every subprocess context-bounded (FR-ARCH-004) |
| Security | Environment filtering per ADR-021 applies to git subprocesses (credential helpers still function per user config); mutating operations require caller-held permission decisions; no shell interpolation — argv arrays only |
| Observability | Per-operation spans with git exit codes and timing; version detection logged at startup |
| Testing | Parser golden tests per pinned git versions; equivalence tests for any go-git fast path (ADR-025); large-repo benchmarks (Volume 12) |
| Extensibility | Git hosting integrations (GitHub et al.) build *above* GitPort per Volume 11; the subprocess core is not pluggable |
| Phase | MVP (status, diff, stage, commit, branch, log per MVP item 12; worktrees per Volume 11 phasing) |

## Terminal Engine

| Aspect | Specification |
|---|---|
| Responsibility | Implementing TerminalPort: PTY-backed and pipe-backed command execution with streaming capture, input, signals, resize, and outcome recording. Behavior owner: Volume 6 |
| Boundaries | Owns PTY/process mechanics via PAL; policy (what may run, resource limits) arrives from SandboxPort handles; command *choice* belongs to callers |
| Public API | Implements TerminalPort; consumed by Tool Runtime (terminal tools) and TUI (interactive command panes per Volume 8) |
| Internal API | PTY session management, capture ring buffers, encoding handling |
| Allowed dependencies | SandboxPort (all launches), PAL (PTY, Processes, Signals, Process Trees, Shell) |
| Prohibited dependencies | Engines; GitPort; provider adapters |
| Inputs | Command specs, stdin writes, signals, resize geometry |
| Outputs | Output chunk streams (truncation always marked), Command Execution records (`succeeded`, `failed`, `timed_out`, `killed` — frozen vocabulary) |
| Events emitted | `terminal.*` family (Volume 6 mints names) |
| Errors | E-TOOL family (Volume 6) |
| States | Executions are process-local; outcomes persist as Command Execution records (Volume 2) |
| Persistence | Command Execution rows in the Run aggregate via callers |
| Concurrency | Many concurrent executions bounded by pools; one stream consumer and serialized writes per execution |
| Security | No launch outside a sandbox handle; environment deny-by-default (ADR-021); capture limits prevent output-flood exhaustion; full tree termination on cancel/timeout |
| Observability | Execution spans; captured output linked to records with redaction; resource usage per execution via PAL |
| Testing | PTY behavior matrix across Tier 1 platforms (conformance per chapter 07); signal/resize tests; capture-limit and encoding fixtures |
| Extensibility | None third-party; shells and defaults configurable per Volume 6/8 |
| Phase | MVP |

## Sandbox Engine

| Aspect | Specification |
|---|---|
| Responsibility | Implementing SandboxPort per ADR-021: layered containment — process-level controls at MVP (env filtering, path policy, resource limits, allow/denylists), OS-level isolation from Beta/v1 (Seatbelt, Landlock, bubblewrap — each PENDING VALIDATION per ADR-021); effective-containment observability. Model owner: Volume 9 |
| Boundaries | Owns mechanism selection and application; *policy content* comes from Permission Manager + Policy Engine decisions; owns no execution I/O (TerminalPort does) |
| Public API | Implements SandboxPort; consumed by Tool Runtime, Terminal Engine, Plugin Runtime, MCP Runtime |
| Internal API | Mechanism backends per platform behind one policy interface (ADR-021); capability probing |
| Allowed dependencies | PermissionPort, PAL (Sandbox, Processes, Process Trees, Signals, Temporary Files surfaces) |
| Prohibited dependencies | Engines; direct OS calls bypassing PAL |
| Inputs | Sandbox specs, policies, execution requests |
| Outputs | Sandbox handles, execution IDs, effective-containment level per execution (recorded, ADR-021) |
| Events emitted | `sandbox.*` family (Volume 9 mints names) |
| Errors | E-SEC family (Volume 9) |
| States | Handles: prepared → active → torn down (process-local); no Volume 2 machine |
| Persistence | Effective containment recorded with each Command Execution/Tool Invocation |
| Concurrency | Concurrent handles; per-handle execution rules from spec; teardown idempotent under concurrent cancellation |
| Security | The containment boundary itself; honesty rule: MUST NOT claim stronger isolation than the mechanism applied (ADR-021); mechanism degradation is explicit, observable, and policy-gated |
| Observability | Containment level per execution queryable (ADR-021); mechanism probe results at startup diagnostics |
| Testing | Escape regression suites per mechanism; env-filtering leak tests; per-platform mechanism validation gates before any isolation claim (ADR-021) |
| Extensibility | New mechanism backends (e.g., containers as optional backend per ADR-021 reversal plan) behind the same interface |
| Phase | MVP (process-level); Beta/v1 (OS-level isolation, PENDING VALIDATION per ADR-021) |

## Permission Manager

| Aspect | Specification |
|---|---|
| Responsibility | Implementing PermissionPort: evaluating permission queries against grants, scopes, and policies; running the Approval flow; recording every decision. Model owner: Volume 9 |
| Boundaries | Owns decision evaluation and records; the permission enum, scope semantics, and policy language are Volume 9's; interactive prompt *presentation* is the drivers' (Volume 8) |
| Public API | Implements PermissionPort; consumed by every side-effecting path (SM-16b: 100% mediation) |
| Internal API | Grant store access, policy evaluation hooks (Policy Engine), approval broker to drivers |
| Allowed dependencies | Policy Engine (L2), Persistence Layer (grants, approvals), Audit Log |
| Prohibited dependencies | Adapters other than via ports; drivers (the broker is an inversion — drivers register approval presenters with it) |
| Inputs | Permission queries/requests, policy verdicts, user approval decisions, revocations |
| Outputs | Decisions (allow/deny/ask vocabulary per Volume 9), Approval records, standing Permission grants |
| Events emitted | `permission.*` family (Volume 9 mints names) |
| Errors | E-SEC family (Volume 9); exit code 5 mapping for denials at the CLI boundary |
| States | Drives the Approval machine (`requested`, `granted`, `denied`, `expired`, `cancelled` — frozen names); Permission rows are append-only records (revocation appends, Volume 2) |
| Persistence | Permissions per scope; Approvals per Volume 2; every decision persisted before its effect proceeds |
| Concurrency | Concurrent checks; single outstanding interactive approval per subject per Volume 9; non-interactive contexts never block on prompts (PRD-009) |
| Security | The core enforcement point of Safe by Default (Principle 8); fail-closed: evaluation errors deny; decisions precede effects |
| Observability | Every decision evented and audit-logged with full context (PRD-006); denial statistics per SM-16 |
| Testing | Decision-table golden tests; fail-closed fault injection; the SM-16(b) unmediated-side-effect enforcement test |
| Extensibility | Policies extend decisions (Policy Engine); the decision engine itself is not pluggable |
| Phase | Core |

## Secret Store

| Aspect | Specification |
|---|---|
| Responsibility | Implementing SecretStorePort per ADR-014: OS keychain integration (macOS Keychain, Secret Service) with the opt-in age-encrypted file fallback; reference-based access; zeroization discipline. Storage model owner: Volume 9 |
| Boundaries | Owns secret material at rest and its access path; credential *metadata* (rows) is the Persistence Layer's (global DB); acquisition flows are the Authentication Layer's |
| Public API | Implements SecretStorePort; consumed by Authentication Layer, Provider Layer, MCP Runtime |
| Internal API | Backend selection (keychain vs fallback), age encryption for the fallback, PAL Credential Store bridge |
| Allowed dependencies | PAL (Credential Store, File Locking, Config Directories); zalando/go-keyring + age (pinned, ADR-014) |
| Prohibited dependencies | Engines; network anything — this component is strictly local |
| Inputs | Secret references, material to store, scope queries |
| Outputs | Material (zeroize-on-release wrappers), reference listings, audit trail entries |
| Events emitted | `secret.*` family (Volume 9 mints names; content-free) |
| Errors | E-SEC family (Volume 9); backend unavailability surfaces E-PORT-003 beneath it (chapter 07) |
| States | Stateless service; Credential `status` vocabulary lives on the metadata rows |
| Persistence | Material in the OS store or encrypted fallback file — never in any SQLite database (ADR-028); metadata global-only |
| Concurrency | Concurrent reads; fallback-file writes serialized via PAL File Locking |
| Security | Deny-by-default environment passthrough keeps material out of child processes (ADR-021); every access audit-logged; fallback is opt-in with explicit user acknowledgment (ADR-014) |
| Observability | Access events (who/what/when, never the value); backend-in-use visible in diagnostics |
| Testing | Backend matrix tests per platform (conformance, chapter 07); fallback round-trip and permission tests; leak tests over logs/events/exports |
| Extensibility | Additional backends behind the PAL Credential Store surface (future OS phases); not third-party pluggable at MVP |
| Phase | MVP |

## Audit Log

| Aspect | Specification |
|---|---|
| Responsibility | Append-only, hash-chained recording of security-relevant actions (Audit Records, Volume 2): permission decisions, secret accesses, sandbox launches, credential changes, policy changes, IPC administrative calls. Semantics owner: Volume 9 |
| Boundaries | Owns the audit append path and chain verification; retention obeys audit precedence (INV-AUD-04); it is not the general event stream (Event Bus) nor logging |
| Public API | Append + verify + query API consumed by Permission Manager, Secret Store, Sandbox Engine, Authentication Layer, IPC server, CLI audit commands |
| Internal API | Hash chaining, sequence management per database |
| Allowed dependencies | Persistence Layer (audit tables, workspace or global per context) |
| Prohibited dependencies | Everything else; nothing may write audit rows except through it |
| Inputs | Audit record drafts with actor/action/subject/permission context |
| Outputs | Committed chained records; verification reports; query results |
| Events emitted | `audit.*` family (Volume 9 mints names) |
| Errors | E-SEC family (Volume 9); chain-verification failures are integrity errors (exit code 9 semantics) |
| States | Append-only records; no machine |
| Persistence | `audit_records` tables per ADR-028 context; hash chain per INV-AUD-02 |
| Concurrency | Appends serialized per chain; reads concurrent |
| Security | Tamper-evidence via chaining; audit writes precede the actions they authorize where Volume 9 requires; redaction never removes attribution |
| Observability | Chain head and verification status in diagnostics; audit query surface feeds UC-13 |
| Testing | Chain property tests (any mutation detected); crash-during-append recovery tests; precedence-over-retention tests |
| Extensibility | None — the audit path is deliberately closed |
| Phase | MVP |

## Policy Engine

| Aspect | Specification |
|---|---|
| Responsibility | Evaluating configured policies: pre-approved permission classes for unattended runs (PRD-009), telemetry consent policy, provider routing constraints, package/skill trust policy. Policy language and permission semantics owner: Volume 9 (telemetry consent: Volume 10) |
| Boundaries | Pure evaluation over declared policy documents; it decides nothing alone — Permission Manager, Provider Layer routing, Telemetry, and Package Manager consume its verdicts within their own flows |
| Public API | Policy evaluation API consumed by Permission Manager, Provider Layer, Telemetry, Package Manager; consumes ConfigPort (policy documents) |
| Internal API | Policy compilation/caching, scope matchers |
| Allowed dependencies | ConfigPort; Core Domain |
| Prohibited dependencies | Adapters; PAL; anything side-effecting — evaluation is pure |
| Inputs | Policy documents (validated by Configuration Manager against Volume 9/10 schemas), evaluation queries |
| Outputs | Verdicts with matched-rule attribution (which policy line decided) |
| Events emitted | `policy.*` family (Volume 9 mints names) |
| Errors | E-SEC family (Volume 9); malformed policy is a configuration error (fail-closed: unevaluable policy denies) |
| States | Stateless over versioned policy snapshots |
| Persistence | Policies are configuration (per scope); verdicts persist only as part of consumers' decision records |
| Concurrency | Pure and concurrent-safe; recompilation on config watch |
| Security | Fail-closed evaluation; rule attribution makes every automated grant explainable (Principle 8: policies are recorded, scoped, revocable) |
| Observability | Verdict events with rule attribution; policy version in effect recorded per run (SM-12) |
| Testing | Policy decision tables; fail-closed fuzzing over malformed documents; attribution golden tests |
| Extensibility | Policies are a named extension surface (Principle 6): the policy vocabulary grows per Volume 9's ADR-gated process |
| Phase | MVP |

Across this chapter one pattern repeats deliberately: mechanism components (Sandbox Engine,
Secret Store, Git Engine, Terminal Engine) own *how*, while decision components (Permission
Manager, Policy Engine) own *whether*, and record entities (Audit Log) own *proof*. The
separation is what lets Volume 9 strengthen policy and Volume 12 tune mechanisms
independently, and it gives the SM-13 traceability chain exactly one place to look for each
question.
