# 03 — Components: Core and Runtime

Chapters 03–06 specify all glossary components except the PAL (chapter
[07](07-platform-abstraction-layer.md)) with one uniform structure: an 18-aspect table plus
short prose. The table fixes each component's **boundary contract** — what it is responsible
for, which ports it implements or consumes, what it may and may not depend on, and its phase.
Behavioral detail belongs to each component's owning volume; a table row that names a volume
delegates to it. "Allowed dependencies" lists components and ports beyond the universal set —
every component may use the Core Domain, its injected ports, Logging, TelemetryPort,
EventBusPort, and SchedulerPort without restating them. Dependency statements here refine, and
never relax, the chapter [01](01-architecture-overview.md) matrix. This chapter covers the L0
core and the L2 execution engines.

## Core Domain

| Aspect | Specification |
|---|---|
| Responsibility | Entities, value types, integrity invariants, and canonical state enums of Volume 2; pure domain computations (state transition legality, ordering rules, ID handling per ADR-027) |
| Boundaries | L0. No I/O, no clocks read directly (time enters as parameters), no provider/platform/persistence specifics; opaque `adapter_metadata` passes through uninterpreted |
| Public API | Domain types and pure functions consumed by every layer; the `sdk/` module mirrors the extension-facing subset |
| Internal API | None — the package tree is the API |
| Allowed dependencies | Standard library only |
| Prohibited dependencies | Everything else: ports, engines, adapters, PAL, drivers, third-party libraries |
| Inputs | Values passed by callers |
| Outputs | Values, transition-legality verdicts, invariant-violation errors |
| Events emitted | None directly (emitting is an engine concern) |
| Errors | Invariant violations as typed domain errors; no E-area of its own — violations surface through the caller's family |
| States | Defines all canonical state enums (Volume 2 chapter 09); holds no runtime state itself |
| Persistence | None — the Persistence Layer maps domain types per Volume 2 chapter 10 |
| Concurrency | Immutable values and pure functions; safe everywhere by construction |
| Security | No secrets, no side effects; hosts redaction-classification metadata on types (which fields are safe to log) consumed by Logging per Volume 9/10 rules |
| Observability | Not instrumented (pure); callers instrument use |
| Testing | Highest bar in the codebase: SM-14's stricter core coverage; property tests for state machines and ordering (ADR-017) |
| Extensibility | Closed to third parties; new entities only via the Volume 2 change procedure |
| Phase | Core |

The Core Domain is deliberately boring: types, rules, and nothing that can fail for
environmental reasons. Its one architectural subtlety is transition legality — the enums are
frozen in Volume 2 while full machines belong to owning volumes, so the Core Domain encodes
*which transitions exist* as data usable by every engine, and owning volumes attach guards
and side effects behind their own APIs. That split lets the Persistence Layer enforce
`CHECK` constraints and lets the TUI render legal next actions without importing any engine.

## Runtime

| Aspect | Specification |
|---|---|
| Responsibility | Composition of the engines into the single execution surface (PRD-001): session lifecycle, run intake and orchestration hand-off, the application API drivers call |
| Boundaries | L2 facade; contains no planning, tool, or provider logic — it routes between engines that do. Behavior owner: Volume 4 |
| Public API | The Runtime application API (start/resume/attach sessions, submit runs, query state) consumed by CLI, TUI, and IPC server; consumes SessionStorePort, WorkspacePort, ConfigPort, PermissionPort |
| Internal API | Engine registration and lifecycle hooks (start/drain/stop) used by the composition root |
| Allowed dependencies | All L2 engines; ports listed above |
| Prohibited dependencies | Adapters; drivers; PAL |
| Inputs | Driver commands; resolved configuration; restored session/run snapshots |
| Outputs | Run/session handles, streamed run activity to drivers, persisted state via ports |
| Events emitted | `session.*` family (Volume 4 mints names) |
| Errors | E-AGT family (Volume 4); wiring failures at startup are E-ARCH-001 |
| States | Drives the Session machine (`created`, `active`, `suspended`, `ended`, `failed` — frozen names); run states are the Agent Engine's to drive |
| Persistence | Session rows and run linkage via SessionStorePort; incremental per the Volume 2 write discipline |
| Concurrency | One supervision root per session; each run owns a task-group subtree (chapter 08); no naked goroutines (ADR-023) |
| Security | Entry-point authority: every driver request is bound to a workspace + session identity before any engine sees it; permission context established here |
| Observability | Root span per run (TelemetryPort); session/run lifecycle events; startup component-set record (FR-ARCH-002) |
| Testing | Integration harness of Volume 13: scripted sessions over contract doubles; crash-recovery suites (SM-11) |
| Extensibility | None directly; extension enters through the engines' surfaces |
| Phase | Core |

The Runtime exists so that "everything drives the same runtime" (PRD-001) is a checkable
property: CLI one-shot runs, TUI interactive sessions, and IPC-driven headless automation all
enter through this one API with identical permission, persistence, and observability
behavior (PRD-009). It is also the recovery orchestrator's home: on startup it asks
SessionStorePort's `MarkInterrupted` for orphaned work and applies the chapter 08 recovery
procedure before accepting new runs.

## Agent Engine

| Aspect | Specification |
|---|---|
| Responsibility | The agent loop: plan–act–observe iteration per run — turn assembly, model interaction, tool-call handling, delegation to sub-agents, run state driving. Behavior owner: Volume 4 |
| Boundaries | Owns Run/Turn/Agent progression; does not select context (Context Manager), produce plans (Planner), execute tasks (Execution Engine), or touch providers directly beyond ProviderPort |
| Public API | Run execution API consumed by the Runtime and Workflow Engine; consumes ProviderPort, PermissionPort, MemoryStorePort, SessionStorePort; consumes Planner, Execution Engine, Context Manager, Prompt Engine (L2 peers) |
| Internal API | Turn assembly pipeline and delegation protocol (Volume 4) |
| Allowed dependencies | L2 peers named above; their ports |
| Prohibited dependencies | Adapters (any); TerminalPort/GitPort/ToolPort directly — all acting goes through the Execution Engine and Tool Runtime |
| Inputs | Run goals, agent profiles (snapshotted per INV-AG-03), context assemblies, tool results, approvals |
| Outputs | Turns, messages, tool-call requests, run outcomes, cost accounting inputs |
| Events emitted | `run.*`, `turn.*`, `agent.*` families (Volume 4 mints names) |
| Errors | E-AGT family (Volume 4) |
| States | Drives Run (`pending` … `cancelled`) and Agent (`instantiated` … `failed`) machines with frozen names; Turn recorded status |
| Persistence | Appends run records incrementally via SessionStorePort at every transition and Record append (PRD-010) |
| Concurrency | One driver goroutine per run within the run's task group; provider streams and tool executions run as supervised child tasks; cancellation per FR-ARCH-004 |
| Security | Never bypasses PermissionPort; provider/model changes are announced and recorded (Principle 7); capability checks before capability use (Principle 2) |
| Observability | Spans per turn; token/cost accounting to Cost Records; every state transition an event |
| Testing | Deterministic loop tests over scripted ProviderPort doubles; replay-mode divergence tests (SM-12) |
| Extensibility | Agent behavior extends via Agent Profiles, Skills, and prompts — not by third-party code in the engine |
| Phase | MVP |

The Agent Engine is the product's heart, and the architecture's job is to keep it small: it
orchestrates ports and peers, holding no provider specifics (Principle 1), no context
heuristics, and no execution mechanics. Everything it does is reconstructable from persisted
records — the loop appends before it acts, which is what makes SM-11 recovery and SM-12
replay honest claims rather than aspirations.

## Planner

| Aspect | Specification |
|---|---|
| Responsibility | Producing and revising Plans: goal decomposition into Task graphs with dependencies, revision on new information, plan-approval interplay. Behavior owner: Volume 4 |
| Boundaries | Produces and revises Plan/Task structures; does not execute tasks and does not talk to tools; model access only via ProviderPort requests assembled with Context Manager and Prompt Engine |
| Public API | Planning API consumed by the Agent Engine and Workflow Engine |
| Internal API | Decomposition strategies and revision heuristics (Volume 4) |
| Allowed dependencies | Context Manager, Prompt Engine (L2); ProviderPort, PermissionPort (plan-approval gates via Agent Engine) |
| Prohibited dependencies | Adapters; ToolPort/TerminalPort/GitPort; Execution Engine internals |
| Inputs | Run goal, workspace snapshot, prior plan versions, execution feedback |
| Outputs | Plan versions with Task graphs (acyclic per INV-TASK-02); revision rationales |
| Events emitted | `plan.*` family (Volume 4 mints names) |
| Errors | E-AGT family (Volume 4) |
| States | Drives the Plan machine (`draft`, `proposed`, `approved`, `executing`, `revising`, `completed`, `superseded`, `abandoned` — frozen names) |
| Persistence | Plans and Tasks persist in the Run aggregate via SessionStorePort; at most one non-terminal plan per run (INV-PLAN-02) |
| Concurrency | Planning for one run is serial; revision may overlap execution only through the `revising` state protocol (Volume 4) |
| Security | Plans are inspectable before execution (PRD-005 gate point); plan-approval requests flow through PermissionPort |
| Observability | Plan versions and revision events fully persisted — the audit trail for "why did it do that" |
| Testing | Golden decomposition tests over recorded fixtures; property tests on graph acyclicity |
| Extensibility | Planning strategies configurable per Agent Profile (Volume 4); no third-party planner injection at MVP |
| Phase | MVP |

The Planner's inspectable output is a product feature (MVP item 4): plans render in the TUI,
gate on approval where policy requires, and their revision history explains agent behavior
after the fact. Architecturally it is a peer of the Agent Engine, not its subroutine, because
the Workflow Engine also plans (workflow steps expand to plans) and both callers need the
same persistence and approval semantics.

## Execution Engine

| Aspect | Specification |
|---|---|
| Responsibility | Executing approved plan Tasks: dependency-order scheduling, dispatching tool invocations through the Tool Runtime, task state driving, retry policy application, error propagation. Behavior owner: Volume 4 |
| Boundaries | Executes Tasks; does not decide *what* to do (Planner/Agent Engine) or *whether* it is permitted (Permission Manager via Tool Runtime); does not run subprocesses itself |
| Public API | Task execution API consumed by the Agent Engine and Workflow Engine; consumes Tool Runtime (L2), SchedulerPort, SessionStorePort |
| Internal API | Retry/backoff machinery, dependency resolver (Volume 4) |
| Allowed dependencies | Tool Runtime (L2); ports above |
| Prohibited dependencies | Adapters; ToolPort directly (always via Tool Runtime); ProviderPort |
| Inputs | Approved Plans, task definitions, tool results, cancellation signals |
| Outputs | Task outcomes, tool invocation requests, propagated errors to the Agent Engine |
| Events emitted | `task.*` family (Volume 4 mints names) |
| Errors | E-AGT family (Volume 4); tool failures arrive as E-TOOL data and map to task outcomes per Volume 4 rules |
| States | Drives the Task machine (`pending`, `ready`, `running`, `blocked`, `awaiting_approval`, `interrupted`, `completed`, `failed`, `cancelled`, `skipped` — frozen names) |
| Persistence | Task transitions persist via SessionStorePort as they occur; interrupted tasks are never assumed complete (PRD-010) |
| Concurrency | Parallel task execution within plan dependency constraints, bounded by run-level pools (chapter 08); each task a supervised scheduler task |
| Security | Every side-effecting dispatch flows through the Tool Runtime's permission mediation; the engine itself has no ambient authority |
| Observability | Span per task; transition events; retry attempts recorded with causes |
| Testing | Deterministic DAG execution tests; fault injection on tool doubles; cancellation storms (Volume 13) |
| Extensibility | Retry/concurrency policy via configuration (Volume 4 tables); no third-party executors |
| Phase | MVP |

The Execution Engine turns a Plan into supervised concurrent work without owning any of the
dangerous parts: the Tool Runtime holds the permission boundary, the Task Scheduler holds the
goroutines, and the engine holds only ordering, retries, and state truth. This is the
component where `interrupted` semantics matter most — on crash recovery it is the Execution
Engine that walks resurrected tasks and re-derives what may resume versus what must be
re-approved (chapter 08).

## Context Manager

| Aspect | Specification |
|---|---|
| Responsibility | Selecting, ranking, budgeting, and assembling Context Items for model requests: candidate gathering from memory/index/workspace/history, token budgeting per model, truncation and packing strategy. Behavior owner: Volume 7 |
| Boundaries | Assembles context; does not persist memory (Memory Manager), build indexes (Indexing Engine), or render prompts (Prompt Engine — the Context Manager supplies material, the Prompt Engine shapes it) |
| Public API | Context assembly API consumed by Agent Engine, Planner, and Workflow Engine; consumes MemoryStorePort, IndexerPort, WorkspacePort, ProviderPort (`CountTokens`, `Capabilities`) |
| Internal API | Ranking pipelines and budget solvers (Volume 7) |
| Allowed dependencies | Ports above; Prompt Engine as downstream consumer only |
| Prohibited dependencies | Adapters; ToolPort; direct provider adapter knowledge (token counting only through ProviderPort) |
| Inputs | Run/turn intent, candidate sources, model capability and budget data |
| Outputs | Ordered, budgeted Context Item sets persisted per turn (Volume 2), with inclusion/exclusion attribution |
| Events emitted | `context.*` family (Volume 7 mints names) |
| Errors | E-CTX family (Volume 7) |
| States | Stateless service; Context Items are Record entities |
| Persistence | Context Items persist in the Run aggregate (workspace DB) per Volume 2; exclusion records per Volume 7's diagnostics decision (V2-OQ-2) |
| Concurrency | Assembly for one turn is serial; candidate gathering fans out as supervised parallel queries with per-source deadlines |
| Security | Redaction gates apply before content enters context where Volume 9 requires; context never includes Secret Store material |
| Observability | Context state is user-visible (Principle 7): what was included, what was excluded, and why, per turn |
| Testing | Budget property tests (never exceeds model budget); golden assembly fixtures; degradation tests when sources are slow/unavailable |
| Extensibility | Ranking strategies configurable (Volume 7); custom indexers plug in beneath via IndexerPort |
| Phase | MVP |

The Context Manager is the quality lever of the whole product — and a discipline problem,
because every component "just wants to add one thing" to context. The architecture therefore
gives it sole authority over assembly: nothing else concatenates content into model requests.
Its per-turn persistence of what was considered and rejected is what makes context state
transparent (Principle 7) and regressions diagnosable.

## Memory Manager

| Aspect | Specification |
|---|---|
| Responsibility | Implementing MemoryStorePort: ingestion with provenance, layered storage (session/workspace/long-term), retrieval and ranking, retention and expiry, export. Behavior owner: Volume 7 |
| Boundaries | Owns Memory Record lifecycles; does not decide what enters a model request (Context Manager) and does not embed content itself (IndexerPort supplies semantic search) |
| Public API | Implements MemoryStorePort; consumed by Agent Engine, Context Manager, CLI/TUI memory commands |
| Internal API | Layer routing, provenance stamping, retention scheduling (Volume 7) |
| Allowed dependencies | Persistence Layer (via its repository API), IndexerPort, ConfigPort |
| Prohibited dependencies | Provider adapters; ProviderPort (embeddings arrive via IndexerPort); drivers |
| Inputs | Memory drafts from runs, queries, retention policies, deletion requests |
| Outputs | Memory Records, ranked results, expiry reports, export streams |
| Events emitted | `memory.*` family (Volume 7 mints names) |
| Errors | E-MEM family (Volume 7) |
| States | Memory Record `status` vocabulary (`active`, `archived`, `expired`, `deleted` — frozen); no machine |
| Persistence | Workspace DB for session/workspace layers; global DB for long-term (ADR-028); cascades per INV-MEM-04 |
| Concurrency | Concurrent reads/writes; retention passes run as scheduled background tasks; ingestion batches are transactional |
| Security | Provenance is mandatory (what run/tool produced a record); deletion honors audit precedence (INV-AUD-04); export excludes redacted content per Volume 9 |
| Observability | Memory state is user-inspectable (Principle 7); ingestion/expiry events; layer size metrics |
| Testing | Retention property tests; cross-layer isolation tests; export/import round-trips (Volume 13) |
| Extensibility | Storage strategy extension point (Principle 6) beneath the port — alternative memory backends implement MemoryStorePort |
| Phase | MVP |

Memory is architecturally split from context on lifetime grounds: memory is durable and
policy-governed (retention, privacy, export), while context is per-turn and disposable. The
port between them keeps that split honest — the Context Manager cannot reach around retention
policy, and memory backends can be swapped (an extension surface of PRD-007) without touching
assembly logic.

## Prompt Engine

| Aspect | Specification |
|---|---|
| Responsibility | Rendering versioned prompt templates: system/agent/task prompts composed from profile parameters, skill contributions, and Context Manager material, with deterministic output for identical inputs. Behavior owner: Volume 4 |
| Boundaries | Renders prompts; owns the template format and registry. Does not select context, call providers, or store skill packages (Skill Engine loads; Prompt Engine composes) |
| Public API | Rendering API consumed by Agent Engine, Planner, Workflow Engine; template registry API consumed by Skill Engine |
| Internal API | Template parsing/caching (Volume 4) |
| Allowed dependencies | ConfigPort; Skill Engine as upstream contributor |
| Prohibited dependencies | Adapters; ProviderPort; MemoryStorePort (material arrives from callers) |
| Inputs | Template identifiers + versions, profile parameters, context material, skill fragments |
| Outputs | Rendered prompt structures with template provenance (which template + version produced each part) |
| Events emitted | `prompt.*` family (Volume 4 mints names) |
| Errors | E-AGT family (Volume 4): unknown template, version conflicts, render failures |
| States | Stateless; templates are versioned assets (prompts are internal versioned assets per Volume 1 non-goal 5) |
| Persistence | Built-in templates in the binary; workspace/global template overrides via configuration (Volume 4/10); rendered prompts referenced from run records for SM-12 |
| Concurrency | Pure rendering; concurrent-safe; template cache invalidation on config watch |
| Security | Templates are code-adjacent assets: override sources are trust-gated per Volume 9; rendered output passes redaction classification |
| Observability | Prompt references (template + version + parameter hash) recorded per turn — full prompts reproducible without logging their content by default |
| Testing | Golden render tests per template version; determinism property tests |
| Extensibility | Prompts are a named extension surface (Principle 6): skills and packages contribute templates through the registry |
| Phase | MVP |

Deterministic, versioned rendering is what ties the loop together for reproducibility: given
the recorded template references and the persisted context items, a run's exact model inputs
are reconstructable (SM-12) without persisting bulky prompt text on every turn. The registry
design also gives skills a disciplined way to alter agent behavior — contributions compose
through declared slots rather than string concatenation, with rules in Volume 4.
