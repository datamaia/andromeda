# 06 — Components: Interface and Delivery

The last component group: the driver surfaces users touch, the delivery machinery that
installs and updates the product, the public SDK, and the observability trio. Table
conventions are those of chapter [03](03-components-core.md); universal dependencies are not
restated.

## CLI

| Aspect | Specification |
|---|---|
| Responsibility | The `andromeda` command-line interface: command grammar, argument parsing (cobra/pflag per ADR-005), non-interactive execution, structured JSON output, exit-code mapping (ADR-016), TUI launch. Behavior owner: Volume 8 |
| Boundaries | L4 driver: translates commands into Runtime API calls and renders results; contains no business logic, no direct adapter access, no persistence of its own |
| Public API | The command grammar (Volume 8 owns names and structure); consumes the Runtime API, ConfigPort (flag layer), UpdaterPort, PackagePort, GitPort/MemoryStorePort/IndexerPort for their command families |
| Internal API | Output renderers (human/JSON), exit-code mapper, prompt fallbacks for approvals in plain-terminal mode |
| Allowed dependencies | L2 engines via the Runtime API; ports; cobra/pflag |
| Prohibited dependencies | L3 adapters directly; TUI internals (launching the TUI is a hand-off, not an import of its widgets) |
| Inputs | argv, stdin, environment, terminal capabilities |
| Outputs | Human-readable output, structured JSON output (schema per Volume 8, stability per SM-20), exit codes 0–9 (ADR-016), TUI hand-off |
| Events emitted | `cli.*` family (Volume 8 mints names) |
| Errors | E-CLI family (Volume 8); usage errors map to exit code 2 |
| States | Stateless per invocation |
| Persistence | None (session state flows through the Runtime) |
| Concurrency | One command per process invocation; streaming output rendered as it arrives; signal handling per chapter 08 shutdown semantics |
| Security | Non-interactive mode never prompts — policy-unresolved permissions deny with exit code 5 (PRD-009); JSON output honors redaction rules |
| Observability | Command execution events; every command declares its producible exit codes (Volume 0 rule) |
| Testing | Grammar golden tests; JSON schema conformance tests; non-interactive matrix in CI (UC-07) per Volume 13 |
| Extensibility | Commands are a named extension surface (Principle 6): plugin/package-contributed commands per Volumes 6/8 |
| Phase | MVP |

## TUI

| Aspect | Specification |
|---|---|
| Responsibility | The interactive terminal interface (Bubble Tea v2 stack, ADR-006): session interaction with streaming, plan/task visibility, diff review, permission prompts, panels, theming per ADR-026. Behavior owner: Volume 8 |
| Boundaries | L4 driver over the same Runtime API as the CLI (PRD-001, PRD-009 parity); renders state from events and API reads — owns presentation state only |
| Public API | None to other components (nothing imports the TUI); consumes the Runtime API, EventBusPort subscriptions, PermissionPort approval presenter registration, TerminalPort (embedded command panes per Volume 8) |
| Internal API | Model/update/view components, layout manager, theme engine (design tokens per ADR-026) |
| Allowed dependencies | L2 via Runtime API; ports; bubbletea/lipgloss/bubbles v2 |
| Prohibited dependencies | L3 adapters directly; CLI internals |
| Inputs | Key/mouse/resize events, Runtime state, event subscriptions, terminal capability probes |
| Outputs | Rendered frames (truecolor→monochrome tiers per ADR-026), approval decisions relayed to the Permission Manager, user commands to the Runtime |
| Events emitted | `tui.*` family (Volume 8 mints names) |
| Errors | E-TUI family (Volume 8) |
| States | Presentation state only; degraded-terminal fallbacks per Volume 8 |
| Persistence | UI preferences via ConfigPort; nothing authoritative |
| Concurrency | Single render loop (Bubble Tea model); event subscriptions with bounded buffers so a busy run cannot stall input handling (SM-07 budget) |
| Security | Displays permission prompts with precise scopes (PRD-008); quiet modes reduce presentation, never recording (Volume 1 tension rule 1) |
| Observability | Interaction latency metrics (SM-07); render errors evented |
| Testing | teatest/v2 golden-frame tests (ADR-017); terminal capability matrix; interaction replay for latency budgets |
| Extensibility | TUI panels are an extension surface where viable (Principle 6, phased per Volume 8) |
| Phase | MVP |

## Updater

| Aspect | Specification |
|---|---|
| Responsibility | Implementing UpdaterPort: release checking against channels, artifact download, checksum/signature/provenance verification (ADR-013), atomic apply, offline rollback. Behavior owner: Volume 14 |
| Boundaries | Updates the product binary and bundled assets only; extension packages are the Package Manager's; release publishing is CI's (Volume 11/14) |
| Public API | Implements UpdaterPort; consumed by CLI update commands and the scheduled check |
| Internal API | Channel metadata client, artifact cache, verifier (cosign/checksums per ADR-013), apply/rollback transactions over PAL |
| Allowed dependencies | PAL (Installer, Updater, Filesystem, File Locking, Temporary Files surfaces); ConfigPort; PermissionPort (update application is a side-effecting action) |
| Prohibited dependencies | Engines; Package Manager; provider adapters |
| Inputs | Release metadata, artifacts, channel configuration, user consent |
| Outputs | Update state progression, applied versions, rollback restoration, update history rows |
| Events emitted | `update.*` family (Volume 14 mints names) |
| Errors | E-REL family (Volume 14) |
| States | Drives the Update process machine (`checking`, `up_to_date`, `update_available`, `downloading`, `verifying`, `applying`, `applied`, `failed`, `rolled_back` — frozen names) |
| Persistence | Release rows and update history in the global DB (ADR-028); retained previous version for offline rollback (SM-19) |
| Concurrency | One update operation machine-wide (lock via PAL); check is concurrent-safe and rate-limited |
| Security | Verify-before-apply is unconditional; signature enforcement per the Volume 1 signing-viability status; downgrade protection per Volume 14 policy; never partial binaries (atomic replace-or-restore via PAL) |
| Observability | Update events per state transition; SM-18/SM-19 timing metrics |
| Testing | N−1→N update tests per release (SM-18); rollback tests offline (SM-19); corrupted/tampered artifact rejection tests |
| Extensibility | Channels and sources are configuration (Volume 14); no third-party update code |
| Phase | MVP (basic update per MVP item 23) |

## Package Manager

| Aspect | Specification |
|---|---|
| Responsibility | Implementing PackagePort: resolving, installing, verifying, and removing extension packages (plugins, skills, MCP server bundles) through the frozen installation states. Behavior owner: Volume 6 (packages); artifact/registry formats shared with Volume 14 |
| Boundaries | Manages package artifacts and registrations; the extensions inside packages are governed by their runtimes (Plugin Runtime, Skill Engine, MCP Runtime) after registration; product updates are the Updater's |
| Public API | Implements PackagePort; consumed by CLI package commands and drivers |
| Internal API | Source resolvers, dependency solver, staging/rollback transactions, manifest validation |
| Allowed dependencies | PAL (Filesystem, Temporary Files, File Locking); PermissionPort; Policy Engine verdicts (trust/signature policy per Volume 9); Persistence Layer (package/extension rows) |
| Prohibited dependencies | Engines; Plugin Runtime/Skill Engine/MCP Runtime internals (they observe registrations; installation does not reach into them) |
| Inputs | Package requests, package archives with manifests/checksums/signatures, trust policy verdicts |
| Outputs | Resolution plans, installed packages with Extension registry rows, verification reports, removal tombstones |
| Events emitted | `package.*` family (Volume 6 mints names) |
| Errors | E-PLUG family (Volume 6) |
| States | Drives the Package installation machine (`requested`, `resolving`, `downloading`, `verifying`, `staged`, `installing`, `installed`, `failed`, `rolled_back`, `removed` — frozen names) |
| Persistence | Package/Extension rows per scope (ADR-028); staged content isolated until `installing` |
| Concurrency | One mutating operation per package; independent packages install concurrently under scheduler pools |
| Security | Verification precedes activation (`staged` gate); trust policy decides signature requirements; nothing partially active on failure (frozen-state guarantee) |
| Observability | Installation events per state; provenance recorded per package |
| Testing | Resolution conflict matrices; kill-during-install recovery tests; tampered-package rejection tests |
| Extensibility | Package sources/registries are configuration (Volume 6/14 phasing); marketplace expansions are post-v1 (Volume 1) |
| Phase | Beta |

## Extension SDK

| Aspect | Specification |
|---|---|
| Responsibility | The public contracts for building extensions: Go module (`sdk/`, ADR-003) mirroring the extension-facing subset of L1 — tool authoring kit, plugin-side ARP implementation (ADR-009), provider adapter kit, skill/workflow format helpers, conformance fixtures. Behavior owner: Volume 6 |
| Boundaries | A library shipped to third parties; runs in *their* processes (plugins) or as in-tree contributions; never imports `internal/` (ADR-031 rule 1) |
| Public API | The `andromeda-sdk` Go API; the ARP wire protocol as language-neutral contract (the SDK is a convenience, the protocol is the contract — ADR-009) |
| Internal API | Codegen/mirror tooling keeping SDK types equivalent to `internal/ports` (CI-enforced) |
| Allowed dependencies | Minimal allowlisted third-party set (ADR-002/ADR-003 dependency budget); standard library |
| Prohibited dependencies | Every `internal/` package; heavyweight frameworks |
| Inputs | Extension author code |
| Outputs | Conformant extensions: plugins speaking ARP, tools passing the Volume 13 contract suite, adapters passing provider conformance |
| Events emitted | None (library) |
| Errors | E-SDK family (Volume 6) for SDK-detected contract misuse |
| States | Stateless library |
| Persistence | None |
| Concurrency | SDK helpers follow the same context-first discipline (FR-ARCH-004) so extensions inherit cancellation correctness |
| Security | Ships the permission-declaration and schema tooling that make extensions safe-by-construction; SDK docs are normative about what extensions MUST NOT do (Volume 6) |
| Observability | Conformance fixtures include observability assertions (events/spans extensions must produce) |
| Testing | The SDK ships the conformance suites themselves (SM-02/SM-03 measure authoring time against them); mirror-equivalence check in CI |
| Extensibility | Is the extensibility product; versioned with subdirectory tags, SemVer per ADR-015 |
| Phase | Beta (contract subsets stabilize at Core/MVP inside the monorepo) |

## Telemetry

| Aspect | Specification |
|---|---|
| Responsibility | Implementing TelemetryPort over OpenTelemetry (ADR-011): metric/span collection, local-first sinks, optional OTLP export strictly under recorded consent. Behavior and consent policy owner: Volume 10 |
| Boundaries | Collection and export mechanics; *what* is measured is each component's declaration (Metric definitions, Volume 2); consent policy content is Volume 10's with Policy Engine evaluation |
| Public API | Implements TelemetryPort; consumed corpus-wide via instrumentation helpers |
| Internal API | OTel pipeline assembly, sink adapters (local files/DB per Volume 10), exporter gating |
| Allowed dependencies | OTel SDK (pinned); Policy Engine (consent verdicts); ConfigPort; PAL (Config Directories) |
| Prohibited dependencies | Engines; secret-adjacent components |
| Inputs | Metric samples, spans, flush requests, consent state |
| Outputs | Local telemetry stores; consented OTLP export; drop counters |
| Events emitted | `telemetry.*` family (Volume 10 mints names) |
| Errors | E-OBS family (Volume 10); telemetry failure never fails the observed operation (port rule) |
| States | Stateless pipeline; consent state is configuration |
| Persistence | Local sinks per Volume 10 retention; nothing leaves the machine without consent (Principle 9: local-first observability) |
| Concurrency | Non-blocking hot path with bounded buffers; flush bounded by context deadlines |
| Security | Redaction before export; consent is opt-in, recorded, revocable (Volume 10); export endpoints validated against config |
| Observability | Self-metrics (drops, buffer occupancy, export outcomes) |
| Testing | Consent-gating tests (0 egress without consent, aligned with SM-05's 0-network-attempts check); pipeline overflow tests |
| Extensibility | Telemetry exporters are a named extension surface (Principle 6) per Volume 10 |
| Phase | Beta (local metrics earlier as Volume 10 phases them) |

## Logging

| Aspect | Specification |
|---|---|
| Responsibility | Structured, redacted, local-first logs via log/slog JSON handlers (ADR-011): level policy, correlation ID injection, redaction filters, rotation/retention hooks. Behavior owner: Volume 10 |
| Boundaries | The logging pipeline; *what* components log is theirs; redaction *rules* are Volume 9/10's — the pipeline enforces them |
| Public API | slog handler wiring + helper API consumed corpus-wide (the sanctioned L2 exception of chapter 01) |
| Internal API | Handler chain (redact → correlate → encode → sink), sink management via PAL |
| Allowed dependencies | PAL (Config Directories, Filesystem, File Locking); ConfigPort |
| Prohibited dependencies | Engines; Event Bus (logs are not events; bridging is Observability's concern) |
| Inputs | Log records with structured attributes |
| Outputs | JSON log files under the platform log location (ADR-022); console output per driver mode |
| Events emitted | None (logs are the lower-level channel) |
| Errors | E-OBS family (Volume 10); logging failure degrades to stderr, never crashes the process |
| States | Stateless pipeline |
| Persistence | Log files with rotation/retention per Volume 10 |
| Concurrency | Concurrent-safe handlers; bounded buffering; no hot-path blocking |
| Security | Redaction is structural: types carry safe-to-log classification (Core Domain metadata) and the secret-leak test suite gates releases (Volume 13); logs never contain Secret Store material |
| Observability | Log volume/drop metrics; correlation IDs make logs joinable with traces and events (SM-13) |
| Testing | Redaction golden tests; secret-injection leak tests; rotation crash tests |
| Extensibility | Additional sinks per Volume 10 configuration; handler chain not third-party pluggable |
| Phase | MVP (MVP item 18) |

## Observability

| Aspect | Specification |
|---|---|
| Responsibility | Correlating the four record kinds — logs, events, traces/spans, metrics — plus Cost Records into one queryable, run-centric view (Principle 9, PRD-006): correlation ID discipline, the run inspection API behind UC-13 audit and cost queries. Behavior owner: Volume 10 |
| Boundaries | Correlation and query composition over data other components produced; owns no collection pipeline (Logging, Telemetry, Event Bus, Persistence Layer do) |
| Public API | Run/session inspection and accounting query API consumed by CLI/TUI (history, cost, audit views) and the IPC surface |
| Internal API | Correlation joiners across stores; cost aggregation per provider/model/run |
| Allowed dependencies | Persistence Layer (events, traces, cost records), Audit Log (query side), MemoryStorePort where Volume 10 defines memory-state views |
| Prohibited dependencies | Provider adapters; drivers |
| Inputs | Queries (by run, session, time, provider); persisted records |
| Outputs | Correlated views: what ran, what it did, what it touched, what it cost — during and after runs (PRD-006) |
| Events emitted | None of its own (it consumes) |
| Errors | E-OBS family (Volume 10) |
| States | Stateless query service |
| Persistence | None of its own; reads others' stores |
| Concurrency | Read-only, concurrent; live views subscribe to the Event Bus with bounded buffers |
| Security | Queries respect redaction and permission scope; audit queries delegate to the Audit Log's controlled surface |
| Observability | Self-describing: its query latencies are themselves metered (Volume 12 budgets) |
| Testing | The SM-13 audit-chain test lives here: enumerate side effects, resolve each to its full record chain, 0 orphans; correlation property tests |
| Extensibility | Views/exports per Volume 10; not third-party pluggable at MVP |
| Phase | MVP |

Two notes bind this chapter together. First, driver parity: the CLI and TUI consume the same
Runtime API with the same permission and persistence semantics — features may render
differently but MUST NOT exist in only one driver unless Volume 8 explicitly scopes them
(PRD-009). Second, the delivery pair: Updater (product) and Package Manager (extensions) are
deliberately separate machines with separate state enums, because their failure domains
differ — a broken extension must never require a product rollback, and vice versa.
