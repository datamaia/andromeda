# 04 — Glossary

One vocabulary for the whole corpus. A concept has exactly one name; a name refers to exactly
one concept. New terms are added via each volume's `99-volume-register.md` and merged here during
consolidation. Entity semantics are owned by Volume 2; component semantics by Volume 3; this
glossary gives the corpus-wide one-line meaning.

## Domain entities

| Term | Meaning |
|---|---|
| Workspace | The root working environment Andromeda operates in: a directory tree plus its Andromeda state (`.andromeda/`), settings, and indexes. |
| Project | A logical unit inside a workspace with its own configuration profile and metadata (commonly a repository). |
| Session | A bounded interactive or non-interactive engagement with Andromeda, holding runs, context, and session memory; persistable and resumable. |
| Agent | An autonomous executor that plans and acts through tools under a profile, permissions, and observability. |
| Agent Profile | A named, versioned configuration of an agent: model/provider selection, prompts, tool set, permission defaults, and behavioral parameters. |
| Run | One top-level execution of an agent or workflow within a session, from intake to terminal state. |
| Turn | One request/response exchange inside a run. |
| Message | A single unit of conversation content (user, agent, system, or tool) inside a turn. |
| Plan | A structured, inspectable set of intended steps produced by the Planner for a run. |
| Task | A unit of executable work derived from a plan, with its own state machine. |
| Tool | A named, versioned, schema-typed capability an agent can invoke, with declared permissions and limits. |
| Tool Invocation | One call of a tool with concrete inputs, under a granted permission set. |
| Tool Result | The typed output (or error) of a tool invocation. |
| Approval | A recorded human decision granting or denying a requested action or permission. |
| Permission | A grant to perform a class of side-effecting action within a scope (see Volume 9). |
| Artifact | A durable output produced by a run (file, patch, report, export). |
| File Change | A recorded modification to a file (create, edit, delete, rename) attributable to a run. |
| Patch | A reviewable diff representing one or more file changes. |
| Command Execution | A recorded execution of a terminal command, with its inputs, environment policy, and outcome. |
| Provider | An adapter-backed source of model inference (cloud service or local server). |
| Model | A concrete inference target exposed by a provider, with declared capabilities. |
| Capability | A declared, machine-checkable ability of a model/provider (e.g., tool calling, streaming, vision); closed enum owned by Volume 5. |
| Credential | A secret used to authenticate against a provider or service; stored per Volume 9. |
| Authentication Session | The state of an authenticated identity against a provider (tokens, expiry, refresh). |
| Workflow | A declared, stateful orchestration of agents, tools, and approvals (see Volume 4). |
| Workflow Run | One execution instance of a workflow. |
| Skill | A packaged, versioned unit of procedural knowledge (prompts + required tools/capabilities) loadable by agents. |
| Plugin | An external extension process integrated through the plugin runtime and Andromeda Runtime Protocol. |
| MCP Server | An external Model Context Protocol server offering tools, resources, or prompts. |
| MCP Client Connection | Andromeda's managed connection to one MCP server. |
| Memory Record | A persisted unit of memory (session, workspace, long-term, semantic, episodic) with provenance and retention. |
| Context Item | One candidate unit of content assembled into a model request by the Context Manager. |
| Index | A queryable structure over workspace content (lexical or semantic). |
| Embedding | A vector representation of content used by semantic indexing and retrieval. |
| Event | A structured, versioned occurrence emitted on the event bus (envelope owned by Volume 10). |
| Trace | A correlated tree of spans describing one run's execution across components. |
| Metric | A named quantitative measurement emitted by the runtime. |
| Cost Record | An accounting entry for tokens/spend attributed to a run, provider, and model. |
| Audit Record | An immutable entry documenting a security-relevant action (who, what, when, under which permission). |
| Configuration Profile | A named set of configuration values selectable at global, project, or invocation level. |
| Package | A distributable unit (binary, plugin, skill) with version, checksum, and signature metadata. |
| Extension | Any third-party addition: tool, plugin, skill, provider adapter, exporter, or command. |
| Release | A published, versioned distribution of Andromeda with artifacts, notes, and provenance. |

## Architecture components

| Term | Meaning |
|---|---|
| Core Domain | Pure domain model and invariants; no I/O, no provider or platform specifics. |
| Runtime | The composed engine layer that executes sessions, runs, and workflows. |
| Agent Engine | Drives the agent loop: planning, acting, observing, iterating. |
| Planner | Produces and revises plans from goals and context. |
| Execution Engine | Executes plan tasks, dispatching tools and managing task states. |
| Context Manager | Selects, ranks, budgets, and assembles context items for model requests. |
| Memory Manager | Ingests, stores, retrieves, and expires memory records. |
| Provider Layer | Common provider contract plus per-provider adapters. |
| Authentication Layer | Credential acquisition, storage integration, refresh, rotation, revocation. |
| Tool Runtime | Registers, validates, sandboxes, executes, and observes tools. |
| Plugin Runtime | Manages plugin processes and their protocol lifecycle. |
| Workflow Engine | Executes workflow definitions as state machines with approvals and resumability. |
| Skill Engine | Loads, validates, composes, and applies skills. |
| Prompt Engine | Renders versioned prompt templates with context and profile parameters. |
| MCP Runtime | Manages MCP client connections, discovery, and lifecycle. |
| Configuration Manager | Loads, validates, migrates, and resolves configuration with precedence. |
| Telemetry | Collects and exports metrics/usage signals under consent policy. |
| Logging | Structured, redacted, local-first logs. |
| Observability | Correlates logs, events, traces, metrics, and cost records. |
| Indexing Engine | Builds and incrementally updates lexical/semantic indexes. |
| Workspace Engine | Discovers, opens, snapshots, and manages workspaces and projects. |
| Sandbox Engine | Applies isolation policies to tool/command execution. |
| Permission Manager | Evaluates permission requests against grants, scopes, and policies. |
| Git Engine | Encapsulated Git operations (status, diff, commit, branch, worktree, …). |
| Terminal Engine | PTY management and command execution with capture and limits. |
| CLI | The `andromeda` command-line interface. |
| TUI | The interactive terminal user interface. |
| Updater | Checks, downloads, verifies, applies, and rolls back updates. |
| Package Manager | Installs, verifies, updates, and removes packages/extensions. |
| Extension SDK | Public contracts for building tools, plugins, skills, and adapters. |
| Persistence Layer | SQLite-backed storage for sessions, memory, indexes, audit, and state. |
| Event Bus | In-process typed publish/subscribe channel for events. |
| Task Scheduler | Schedules and supervises concurrent tasks with cancellation and backpressure. |
| Platform Abstraction Layer (PAL) | Encapsulates OS-specific behavior (filesystem, processes, signals, PTY, credential store, …). |
| Secret Store | The credential-storage abstraction over OS keychains and the encrypted-file fallback. |
| Audit Log | Append-only store of audit records. |
| Policy Engine | Evaluates configured policies (permissions, telemetry, provider routing constraints). |

## Product and process terms

| Term | Meaning |
|---|---|
| AI Engineering Harness | The product category Andromeda belongs to: a runtime + tool environment for engineering with AI agents. |
| Andromeda Runtime Protocol | The JSON-RPC 2.0-based protocol between Andromeda and plugin processes (see the decision register, chapter 06). |
| Phase | A delivery stage: `Core`, `MVP`, `Beta`, `v1`, `v2`, `Future`, `Out of Scope` (defined in Volume 1). |
| Keystone requirement | A hub requirement pre-listed for cross-volume reference. |
| Local-first | Core functionality operates without Internet when models, tools, indexes, memory, and repositories are local. |
| Vendor-agnostic | No provider-specific logic outside that provider's adapter. |
| Model-agnostic | Behavior driven by declared capabilities, not by assumptions about specific models. |
| Design token | A named brand value (color, typography role) from the brand definition; the closed set is specified in Volume 8. |
| Volume register | The `99-volume-register.md` file at the end of each content volume, listing everything that volume minted. |
| Single-home matrix | The table in chapter 03 assigning each cross-cutting topic exactly one authoritative volume. |
| Spec linter | `scripts/spec_lint.py`, the mechanical enforcement of Volume 0. |

## Consolidated additions

Terms minted by Volumes 1–14 through their `99-volume-register.md` files, merged here at
consolidation: deduplicated (one entry per term; "Effective containment level" was recorded
by both Volume 9 fragments and appears once) and sorted alphabetically. Each meaning names
its defining volume, chapter, requirement, or ADR.

| Term | Meaning |
|---|---|
| Accessible output mode | The linear, append-only rendering profile for assistive and transcript use (ADR-111, FR-TUI-065). |
| Action registry | The single TUI catalog of named, context-predicated, permission-annotated operations that palette, quick actions, keymap, and help render from (ADR-110). |
| Adapter | An L3 implementation of a port against an external system (provider, SQLite, git, OS); importable only by the composition root. |
| Adapter Declaration | The static, versioned, registration-validated manifest of every provider fact an adapter claims (Volume 5, chapter 01). |
| Agent loop | The single plan–act–observe iteration cycle the Agent Engine drives for every run (FR-AGT-001) |
| Aggregate | A consistency boundary of entities with one root; intra-aggregate invariants commit in one transaction, cross-aggregate references are by ULID only (Volume 2, chapter 01). |
| Aggregate root | The entity that owns an aggregate's identity and lifecycle; deleting it deletes or tombstones its members (Volume 2, chapter 01). |
| Allowlist construction | Building telemetry payloads field-by-field from the catalog so prohibited data is unrepresentable rather than filtered (ADR-141) |
| Approval gate | A `gate` step raising an Approval with subject kind `workflow_gate`; passed only by a terminal `granted` decision (INV-WFR-04). |
| ARP handshake | The host-initiated `arp.initialize` exchange negotiating the protocol version and validating declared surfaces against the manifest (chapter 08). |
| Ask-forcing rule | An `effect = "ask"` policy rule that pins interactive review even where a standing allow matches (ADR-121 tier 2). |
| Attack vector | The channel through which a threat reaches an asset (model context, tool surface, command/path, credential handling, supply chain, inference channel, human channel). |
| Audited action | An entry of the closed chapter 08 catalog that MUST yield exactly one Audit Record, including denied/failed outcomes. |
| Authentication Profile | A named configuration unit binding a provider slug to a credential label with selection options (ADR-062). |
| Backport | A cherry-picked, qualification-listed fix released as a patch on a supported release branch (FR-REL-014). |
| Budget slice | The portion of run budgets explicitly allocated to a delegated sub-agent |
| Canonical JSON | Volume 2 chapter 10's serialization form: UTF-8, sorted `snake_case` keys, omitted absent optionals, RFC 3339 UTC timestamps, `schema_version` field — the reproducible input for hashing and export. |
| Canonical state enum | The frozen state-name set of a stateful entity, owned by Volume 2 chapter 09; full machines are owned by the entity's area volume. |
| Capability provenance | The class recording why a capability value is present: `declared`, `discovered`, `configured`, `verified`, or masked by `refuted` (Volume 5, chapter 02). |
| Catalog entity | A long-lived registry-style entity with natural-key uniqueness and in-place updates tracked by a `revision` counter (Volume 2, chapter 01). |
| Change-request vocabulary | The provider-neutral operation set over pull requests and merge requests implemented by the `github` and `gitlab` tools (Volume 11, chapter 02; ADR-147). |
| Chunk locator | The stable path + byte-offset + line-span reference identifying one chunk (ADR-088); the `source_ref` of `file_chunk` embeddings. |
| Collected-data catalog | The versioned, user-facing document enumerating every field remote telemetry may carry; its version is the consent version (chapter 06) |
| Compact / explicit form | The two `ANDROMEDA_*` addressing forms of FR-CFG-004: schema-matched single-underscore names, and `__`-separated path segments. |
| Compaction ladder | The ordered, non-generative shrinking steps applied to borderline context candidates before exclusion (FR-CTX-003). |
| Compensation | A step's declared undo action, executed during rollback in reverse completion order (ADR-054). |
| Composition report | The structured record of every skill contribution, its source skill/version, and every override applied in one activation set (chapter 07). |
| Composition root | `cmd/andromeda` — the only production code that constructs and wires engines with adapters (FR-ARCH-002). |
| Configuration layer | One of the ten precedence-ordered value sources of FR-CFG-001, from CLI flags down to built-in defaults. |
| Confirmation tier | The three-level destructiveness-bound confirmation scheme: none / default-No modal / typed name (FR-UX-072). |
| Confused deputy | A component (including the model) that misuses its own granted authority when misled by untrusted input. |
| Consent record | The global-database row created by the interactive telemetry confirmation (consent version, product version, timestamp, endpoint, scope) |
| Consolidation | The Beta lifecycle pass that summarizes clusters of episodic records into semantic/procedural records and archives the sources (FR-MEM-006). |
| Context snapshot | The persisted Context Item rows plus assembly metadata that make a turn's request reproducible (FR-CTX-007). |
| Contract test kit | The reusable per-port test package that any implementation (real adapter or fake) must pass (FR-TEST-004). |
| Cost rollup | A daily aggregate of Cost Records keyed by day, provider, model, workspace, and cost basis; bases and currencies never merge (chapter 05) |
| Degradation strategy | The per-capability rule applied when a required capability is absent: `refuse`, `report_unavailable`, `substitute` (opt-in), or `reroute` (Volume 5, chapter 02). |
| Degradation tier | One of `truecolor` / `ansi256` / `ansi16` / `none` — the color capability level frames render at (FR-UX-041) |
| Degraded mode | A named, evented, queryable reduced-service condition from the closed catalog of Volume 12, chapter 02 (ADR-162). |
| Dependency matrix | The chapter 01 table declaring every ALLOWED/PROHIBITED import direction between layers (FR-ARCH-001). |
| Derived palette | The closed set of theme values derived from the five brand tokens (chapter 08) |
| Descriptor pinning | SHA-256 digests of approved MCP tool descriptors persisted at exposure approval; later drift suspends exposure (E-MCP-008). |
| Destructive confirmation | A CLI-local consent checkpoint for operations that destroy or overwrite user-visible state; distinct from permission Approvals (FR-CLI-010). |
| Detector registry | The versioned, compiled-in set of named secret detectors with `definite` (blocking) and `heuristic` (warning) classes. |
| Direct-execution plan | A one-task Plan produced for goals classified single-step, preserving the every-run-has-a-plan invariant |
| Driver | An L4 entry-point surface (CLI, TUI, IPC server) that steers the Runtime and is imported by nothing. |
| Durable timer | A workflow deadline persisted as an absolute UTC instant, enforced by guarded idempotent firing with a ≤ 30 s sweep bound (ADR-051). |
| Effective capability set | The per-model capability set resolved from declared/discovered/configured provenance with refutation masking (Volume 5, chapter 02). |
| Effective containment level | The observable, per-execution record of which sandbox layer actually applied (ADR-021). |
| Effective timeout | min(declared timeout or configured default, configured cap), recorded per invocation. |
| Effects classification | A step's declared side-effect class (`none`/`workspace`/`external`) driving resume and rollback rules (ADR-050). |
| Embedding space | The immutable provider+model+dimensions declaration of a semantic index (Volume 2 attribute; INV-IDX-03; ADR-089). |
| Enforcement class | One of refuse, truncate, virtualize/narrow, shed/queue — the closed set of behaviors at an operational limit. |
| Ephemeral credential | Environment-sourced credential material used for the process lifetime and never persisted (FR-AUTH-002). |
| Event family | The delivery classification of an event name (`lifecycle`, `action`, `progress`, `security`, `telemetry`) that fixes its overflow policy and default buffer (chapter 04) |
| Exposure approval | The recorded Approval that makes an MCP server's discovered surfaces agent-visible (chapter 06). |
| Extension mount (`x` namespace) | The reserved command group under which all extension-contributed commands appear (ADR-104). |
| Fallback chain | A configured, ordered set of fallback targets with trigger classes and guard parameters, the only mechanism by which fallback occurs (Volume 5, chapter 05). |
| Fallback store | The age-encrypted credential file backend, active only under explicit consent (ADR-014; FR-SEC-110). |
| Gate profile | The configured selection (`strict`/`standard`/`minimal`) of which SDD stage gates require human approval. |
| Gate tier | One of T0 (merge) / T1 (trunk) / T2 (scheduled) / T3 (release qualification) / T4 (phase gate), fixing when a suite runs and what it blocks (Volume 13, chapter 01). |
| Generic-adapter-first | The ADR-065 rule that OpenAI-compatible services are reachable via the generic adapter before any dedicated adapter ships. |
| Global database | The authoritative SQLite database at `<data_dir>/andromeda/global.db`, one per user per machine, holding machine-level state including all credential metadata (ADR-028). |
| Glyph set | One of the two closed chrome character inventories, `unicode` or `ascii`, with declared parity (ADR-112). |
| Go-to chord | A `g`-prefixed two-key sequence addressing a screen directly (FR-TUI-004) |
| Golden file | A versioned expected-output file compared byte-exactly, updated only via an explicit flag recorded in the diff (Volume 13, chapter 02). |
| Grant scope | The lifetime attachment of a Permission row: `invocation`, `run`, `session`, `workspace`, or `global` (Volume 2 enum). |
| Guard rules F1–F8 | The ordered normative conditions every fallback step must pass: explicit chains, egress, capability, policy, cost, stream boundary, no auth masking, announcement (Volume 5, chapter 05). |
| Headless mode | The Beta operating mode of the same binary driven solely through the IPC surface with policy-only permissions (ADR-032, FR-ARCH-008). |
| Incident record | The paired `incident.recorded` audit actions (with `security.incident.*` events) marking an open or closed security incident. |
| Index cache database | The non-authoritative, rebuildable SQLite database at `.andromeda/index.db` holding index data and embeddings; its loss is never an integrity error (Volume 2, chapters 07/10). |
| Index generation | The monotonic counter of committed build/update results for one index; queries are served from exactly one committed generation (FR-IDX-001). |
| Installation owner | The detected manager of the installed binary — `self`, `homebrew`, or `package` — deciding whether self-update proceeds or defers (FR-REL-009). |
| Invocation pipeline | The fixed validate → permission → sandbox → execute → record order every invocation traverses. |
| Invocation-mode record | The immutable per-invocation resolution of interactivity, CI mode, color, output format, and verbosity that every pipeline stage consults (chapter 01). |
| Iteration boundary | The loop checkpoint between turns where budgets, limits, and revision triggers are evaluated |
| Layer manifest | The machine-readable package-to-layer assignment from which ADR-033 enforcement is generated. |
| Layout class | One of `wide` / `standard` / `compact` — the size-derived arrangement a screen renders in (FR-TUI-002) |
| Least exposure | Sending or exposing the minimum context and surface necessary to a lower-trust domain. |
| Locality rule | The ADR-066 classification: an endpoint is local exactly when it resolves to loopback or a Unix domain socket. |
| Management frame | The shared list-pane/detail-pane layout every platform screen instantiates (chapter 10). |
| Mandatory set | Context tiers 1–3 (system/skill prompts, pins, current intent); assembly refuses rather than drops them (ADR-086). |
| Memory kind | The closed content-class axis of a Memory Record: `episodic`, `semantic`, `procedural`, `preference`, `decision` (ADR-085). |
| Memory layer | The storage/visibility axis of a Memory Record: `session`, `workspace`, or `long_term` (values frozen by Volume 2; semantics chapter 01). |
| Mock provider | A ProviderPort test double emitting scripted responses and stream chunks with configurable pacing, used to isolate Andromeda-added overhead from model inference time. |
| Mutation score | Killed mutants ÷ (killed + surviving mutants) after recorded equivalent-mutant exclusions (ADR-175). |
| MVP provider seed | The three MVP provider adapters: generic OpenAI-compatible, Anthropic (cloud), and Ollama (local) (Volume 1, chapter 05). |
| Network sentinel | Test-build instrumentation at the dialer seam that records and fails on outbound connection attempts in hermetic suites (FR-TEST-005). |
| Offer rule | The ADR-191 selection: the highest non-yanked release whose channel maturity ≥ the subscriber's channel and whose upgrade path admits the installed version. |
| Offline guarantee list | The eleven operations (Volume 1, chapter 04, Local First) that MUST work with no Internet connection when models, tools, indexes, memory, and repositories are local. |
| Operation class | The Git operation catalog's classification — read-only, additive, history, destructive, remote — driving permission and confirmation requirements (Volume 11, chapter 01). |
| Operational limit | A normative bound from Volume 12, chapter 03 enforced by one of four declared enforcement classes, always with an enforcement event. |
| Package source | A configured acquisition endpoint of kind `registry`, `git`, `archive`, or `path`, consulted by discovery and resolution (chapter 09). |
| PAL surface | One of the 19 platform abstractions of chapter 07, each with per-OS backends, capability probes, and a declared degradation policy. |
| Panel | A rectangular region of a screen with title, border, optional viewport, and focus state |
| Persist-then-publish | The delivery discipline whereby lifecycle/action events append transactionally with the writes they describe before bus publication (ADR-137) |
| Phase gate | The audit at a phase boundary (Volume 1, chapter 05) verifying exit criteria, bound metrics, and the risk register before the next phase begins. |
| Port interface | One of the 18 frozen L1 interfaces of chapter 02 through which engines reach infrastructure and extensions reach Andromeda. |
| Pricing table | User-maintained per-(provider, model) price configuration with mandatory `source` and `effective_date`, the only source of cost estimates (Volume 5, chapter 04; ADR-058). |
| Process family | One of the four sanctioned child-process categories (plugins, MCP stdio servers, sandboxed tool/terminal children, git operations), each with exactly one supervisor (FR-ARCH-005). |
| Protected tables | The tables runtime overrides may never touch: `[permissions]`, `[sandbox]`, `[security]`, `[auth]`, `[telemetry]`. |
| Provenance label | The mandatory PR-level `ai/none` / `ai/assisted` / `ai/generated` label recording how a change was produced (Volume 11, chapter 04; ADR-015). |
| Provider Router | The Provider Layer composite that implements ProviderPort, owning routing, fallback, retries, pacing, and breakers (Volume 5, chapters 01/05). |
| Qualification evidence bundle | The schema-validated JSON record of everything that qualified a release candidate, retained with the release (FR-TEST-009). |
| Quarantine (test) | Exclusion of a flaky test from gating suites via the `quarantine` build tag with a dated, issue-linked comment and a 14-day time box (ADR-177). |
| Re-anchor | The acknowledged restart of a tampered audit chain: a new sentinel segment referencing the break and the evidence-export digest. |
| Ready computation | The Execution Engine's derivation of dispatchable tasks from dependency and plan state |
| Record entity | An append-only, immutable entity (messages, events, audit records, …); mutation of history is a defect (Volume 2, chapter 01). |
| Recorded status | A frozen outcome/classification vocabulary on an entity attribute without a governed state machine (Volume 2, chapter 09). |
| Redaction registry | The in-process set of resolved secret material (and encodings) that every sink scrubs before emission (ADR-124). |
| Redaction token | The ASCII placeholder (`[redacted:<class>]`, `[redacted:sha256:<8 hex>]`) that replaces a value in every redacted rendering. |
| Reference dataset (DS-*) | A deterministic, seed-generated benchmark corpus defined in Volume 12, chapter 01 (DS-M, DS-L, DS-F, DS-MEM, DS-SOAK). |
| Reference machine (RM-1, RM-2) | One of the two canonical test machines of Volume 12, chapter 01, with a ±10% calibration equivalence rule (ADR-160). |
| Registry index | The schema-versioned JSON document a `registry` source serves: packages with versions, kinds, locators, checksums, and signature references. |
| Release mirror | A filesystem or HTTPS root with an `index.json` and artifact files, serving the full update flow to air-gapped sites via `[update].source` (FR-REL-004). |
| Render profile | The resolved per-process rendering facts: color tier, glyph set, motion, and accessibility flags (ADR-114). |
| Render provenance | The `(namespace, name, version, parameter_hash)` tuple recorded per rendered prompt on its consuming turn |
| Replay mode | An execution mode that re-traverses a recorded run's decision and tool sequence from its run record without re-invoking providers (formalized in Volumes 4 and 10). |
| Reserved namespace | A built-in tool namespace closed to third-party registration (ADR-070). |
| Reserved root key | A root-level `andromeda.toml` key owned by the schema itself: `config_version`, `include`, `default_profile`. |
| Residual risk | The exposure that remains after a threat's named controls are applied. |
| Resolution plan | The deterministic recorded output of `PackagePort.Resolve` — exact versions, sources, checksums, signature references — and the only input `Install` accepts. |
| Resource watchdog | The runtime sampler of disk, memory, database size, and saturation that trips degraded modes and pre-run refusals (Volume 12, chapter 02). |
| Restore point | The git branch/commit anchor SDD records before its first side-effecting stage, used by implementation compensation. |
| Result envelope | The fixed eight-field JSON document (`schema`, `command`, `ok`, `exit_code`, `data`, `error`, `warnings`, `meta`) every `--json` command emits (FR-CLI-006). |
| Resume Approval | The explicit human decision required before an interrupted side-effecting step re-dispatches. |
| Retained version | The previously installed binary (plus verification metadata and supported schema versions) kept by the Updater to enable offline rollback (ADR-192). |
| Rolling baseline | The median of a benchmark's last 5 mainline nightly results on the same runner class, used by the relative regression gate (ADR-161). |
| Run record | The persisted, correlation-ID-linked record of a run — configuration snapshot, provider/model identity, prompts references, tool-invocation sequence and results — sufficient for audit and replay (formalized in Volumes 4 and 10). |
| Runtime override | A session-scoped, in-memory, never-persisted key assignment issued through TUI or IPC surfaces (FR-CFG-005). |
| Safety ref | A local-only ref under `refs/andromeda/safety/<ulid>` recorded before every destructive Git operation, restorable offline within its retention window (ADR-146). |
| Sandbox tier | One of five base isolation profiles keyed by subject kind: `process`, `tool`, `workflow`, `plugin`, `mcp_server` (ADR-122). |
| Scenario script | A versioned, deterministic script driving the fake provider: content, chunk pacing, tool calls, injected errors, usage figures (ADR-176). |
| Scope qualifier | One of the ten frozen dimensions (`session`, `workspace`, `command`, `tool`, `provider`, `host`, `path`, `domain`, `repository`, `organization`) bounding what a grant covers. |
| Screen | One full-frame TUI view dedicated to a concern (session, plan, diff, …); exactly one is active (ADR-107) |
| Secret reference | An opaque handle (`secret_ref`, `token_ref`) into the Secret Store; the only representation of secret material the domain model permits (Volume 2, chapter 05). |
| Sensitivity class | The per-key schema marker (`public`/`sensitive`) governing whether a value may ever render in any sink (FR-CFG-011). |
| Shared verb vocabulary | The closed subcommand verb set (`list`, `show`, `add`, `remove`, `install`, `uninstall`, `enable`, `disable`, `test`, `status`, `validate`, `search`, `export`) of ADR-100. |
| Single-flight refresh | The FR-AUTH-010 discipline: at most one renewal exchange in flight per Authentication Session, all consumers awaiting its outcome. |
| Skill activation | Making a registered skill effective for a session, profile, or run after its requirements resolve (FR-SKILL-002). |
| Skill snapshot | The per-run record of resolved skill names, versions, hashes, and application order, immutable for the run (ADR-052). |
| Source attribution | The per-key record of which layer and concrete source (file:line, profile, variable, flag) supplied every resolved value. |
| Spillover Artifact | The content-addressed Artifact holding tool output beyond the inline cap (ADR-071). |
| Splash screen | The identity screen rendering the mascot, wordmark, and tagline per `tui.splash` policy (FR-UX-043) |
| Stage | A named step of the builtin `spec-driven-dev` workflow (fourteen stages, ADR-049). |
| Staging area | The scope-local directory where acquired archives are verified and extracted before installation; cleaned on every failure path. |
| Staleness threshold | The pending-change count/age bounds beyond which a `ready` index transitions to `stale` (FR-IDX-004). |
| Stream document | One NDJSON line wrapping a Volume 10 event envelope, emitted by streaming commands before their result envelope (ADR-101). |
| Success metric (SM) | A product-level measurable commitment in Volume 1, chapter 06, formalized as one or more NFRs by the named owning volume. |
| Supersession chain | The linear, acyclic version history formed when a memory record replaces another (FR-MEM-005); heads are retrievable, history on request. |
| Supervision tree | The chapter 08 context/group hierarchy under which all concurrent work runs (ADR-023, FR-ARCH-006). |
| Support window | The published period during which a release line receives defined fix classes (ADR-193). |
| Teardown budget | The grace-plus-kill window bounding termination escalation after cancel or timeout. |
| Threat actor | A party (human, process, or the confused-deputy model) capable of realizing a threat. |
| Tier 1 platform | A platform on which releases are built, tested, and supported, and whose full acceptance suite gates releases (Volume 1, chapter 05). |
| Tier 2 platform | A platform on which releases are built and smoke-tested; defects are accepted but do not gate releases (Volume 1, chapter 05). |
| Toast | A non-focus-stealing, severity-classed notice with bounded stacking and error double-recording (FR-UX-073). |
| TOCTOU | Time-of-check-to-time-of-use: a race where a checked resource changes before it is used (relevant to symlink handling). |
| Tombstone | A retained entity row whose content or subject was removed, preserving attribution and hashes after pruning or unregistration (Volume 2, chapters 04/06/10). |
| Tool Declaration | The single self-describing JSON contract document every tool registers (chapter 01). |
| ToolEvent | The five-kind ordered stream union a tool emits during execution: `progress`, `log`, `output_delta`, `artifact`, `result`. |
| Traceability chain report | The per-release artifact resolving every commit in the release range to its PR, issue, and requirement IDs (Volume 11, chapter 07). |
| Trust boundary | A separation between domains of differing trust where a security control MUST sit. |
| ULID | Universally Unique Lexicographically Sortable Identifier — 48-bit millisecond timestamp + 80 random bits, canonically 26 characters of uppercase Crockford base32 (ADR-027). |
| Untrusted-content labeling | Marking ingested content (files, tool results, memory, index hits, provider output) as data, never as privileged instruction. |
| Update channel | One of the closed set `stable`, `rc`, `beta`, `nightly`: a maturity filter over the single SemVer-ordered release line (ADR-191). |
| Update history | The append-and-update global-database record of every Update machine instance (FR-REL-016). |
| Update lock | The machine-wide PAL file lock serializing mutating Updater operations (E-REL-007). |
| Update notice | The throttled, suppressible post-command stderr line announcing a cached newer release (chapter 06). |
| Usage report | The per-request token/cost accounting structure populated exclusively from official provider accounting (Volume 5, chapter 04). |
| View state | One of the six canonical per-panel UI states: `loading`, `content`, `empty`, `error`, `offline`, `degraded` (FR-UX-074). |
| Workflow policy check | The required CI check mechanically enforcing the ADR-149 workflow security posture (Volume 11, chapter 06). |
| Workflow step | One declared unit of a workflow definition: kind `agent`, `gate`, or `transform`, with dependencies, criteria, timeout, retry, and routing. |
| Workspace database | The authoritative SQLite database at `.andromeda/state.db`, one per Workspace, holding all workspace-scoped state (ADR-028). |
