# Annex — Traceability Matrix

**Status:** Consolidated (Phase C). This annex is the corpus-wide, requirement-level
traceability table behind the master matrix of Volume 0, chapter
[09](../volume-00-conventions/09-master-traceability-matrix.md): every `FR-*`, `NFR-*`,
and `RISK-*` identifier in the corpus, one row each, grouped by owning volume. This annex
mints nothing and renames nothing.

Totals: 482 requirements — 266 functional (FR), 102 non-functional (NFR), and 114 risks (RISK).

Generation rule: rows are reproduced from each volume's `99-volume-register.md`
`## Requirements index` table (which the spec linter cross-checks against the defining
chapters), in register order, with cell content verbatim; the ID column links to the
chapter that defines the requirement. Registers index risk rows in one of two shapes:
under a `Phase | Verification method` header (reproduced as-is; some risks carry a
phase, most carry `—`) or under a `Severity | Status` header — those rows render `—`
for Phase and carry the register's cells as `Severity: …; Status: …` in the
Verification-method column (the phase-gate risk-register review of Volume 0, chapter
07, is where risk severity, mitigation, and status are tracked corpus-wide). `PRD-*`
objective rows are not requirements and are indexed in Volume 0, chapter 09 instead.
Volumes 0, 2, and 15 define no FR, NFR, or RISK identifiers and therefore have no
table.

## Volume 1 — Vision, Problem, Scope, and Product

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [RISK-PRD-001](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Provider API drift and deprecation | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-002](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Provider access-policy changes restricting third-party clients | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-003](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Maintainer bus factor | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-004](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Scope creep versus MVP viability | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-005](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Security incident caused by agent actions | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-006](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Local-model capability gaps | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-007](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | MCP ecosystem instability | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-008](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Competitive pace of funded incumbents | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-009](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Contributor onboarding friction | — | Risk register review at phase gates | Volume 1 |
| [RISK-PRD-010](../volume-01-vision-and-product/07-constraints-dependencies-risks.md) | Public-contract churn breaking extensions | — | Risk register review at phase gates | Volume 1 |

## Volume 3 — System Architecture

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-ARCH-001](../volume-03-architecture/01-architecture-overview.md) | Layered dependency rule | Core | depguard + import-graph test in CI (ADR-033); release audit | Volume 3 |
| [FR-ARCH-002](../volume-03-architecture/01-architecture-overview.md) | Ports-and-adapters composition | Core | Static adapter-name checks; contract suites over doubles; startup failure injection | Volume 3 |
| [FR-ARCH-003](../volume-03-architecture/02-port-interfaces.md) | Port name and signature freeze | Core | Spec-lint and consolidation audit; contract-diff tooling in CI | Volume 3 |
| [FR-ARCH-004](../volume-03-architecture/02-port-interfaces.md) | Context propagation and cancellation on all ports | Core | Per-method cancellation cases in contract suites; leak and fault-injection tests | Volume 3 |
| [FR-ARCH-005](../volume-03-architecture/08-processes-concurrency-ipc.md) | Bounded process model | Core | Process-tree termination tests per platform; spawn-path static audit | Volume 3 |
| [FR-ARCH-006](../volume-03-architecture/08-processes-concurrency-ipc.md) | Supervised concurrency | Core | Naked-goroutine scan; scheduler contract suite; leak gates | Volume 3 |
| [FR-ARCH-007](../volume-03-architecture/08-processes-concurrency-ipc.md) | External IPC control surface | MVP | IPC conformance suite; cross-user security tests; CLI-parity tests | Volume 3 |
| [FR-ARCH-008](../volume-03-architecture/08-processes-concurrency-ipc.md) | Headless operating mode | Beta | Headless E2E suite; interactive/headless record-parity comparison | Volume 3 |
| [FR-ARCH-009](../volume-03-architecture/08-processes-concurrency-ipc.md) | Crash recovery and resumable state | MVP | SM-11 crash-injection suite; recovery idempotence tests; corruption fixtures | Volume 3 |
| [FR-ARCH-010](../volume-03-architecture/08-processes-concurrency-ipc.md) | Graceful shutdown ordering | MVP | Shutdown-ordering integration tests; escalation and wedged-child fixtures | Volume 3 |
| [FR-ARCH-011](../volume-03-architecture/09-deployment-update-extensibility-compatibility.md) | Extension through versioned public contracts | Core | Per-surface conformance suites; SM-16(b) matrix over extension origins; compatibility fixtures | Volume 3 |
| [FR-PORT-001](../volume-03-architecture/07-platform-abstraction-layer.md) | Platform encapsulation | Core | Prohibited-construct scanner in CI; PAL conformance suite | Volume 3 |
| [FR-PORT-002](../volume-03-architecture/07-platform-abstraction-layer.md) | PAL surface completeness and portability of signatures | Core | Per-surface conformance suites per Tier 1 platform; signature review | Volume 3 |
| [FR-PORT-003](../volume-03-architecture/07-platform-abstraction-layer.md) | Directory resolution through the PAL with XDG semantics | Core | Golden path-resolution tests per platform; container CI job | Volume 3 |
| [FR-PORT-004](../volume-03-architecture/07-platform-abstraction-layer.md) | Platform support matrix conformance | MVP | Tier 1 release gates; SM-17 measurement; startup-refusal tests | Volume 3 |
| [NFR-ARCH-001](../volume-03-architecture/01-architecture-overview.md) | Dependency-rule enforcement in CI | Core | Required PR checks (depguard + graph test); release audit | Volume 3 |
| [NFR-ARCH-002](../volume-03-architecture/02-port-interfaces.md) | Port contract stability | Beta | Contract-diff per release; Volume 14 release audit | Volume 3 |
| [NFR-ARCH-003](../volume-03-architecture/08-processes-concurrency-ipc.md) | Shutdown deadline | MVP | Instrumented shutdown timing in suites and field diagnostics | Volume 3 |
| [NFR-ARCH-004](../volume-03-architecture/08-processes-concurrency-ipc.md) | Leak-free termination | Core | Goroutine-leak gates per suite; post-shutdown process scans | Volume 3 |
| [NFR-PORT-001](../volume-03-architecture/07-platform-abstraction-layer.md) | Tier 1 behavioral parity | MVP | Full acceptance suite per Tier 1 platform per release (SM-17) | Volume 3 |
| [NFR-PORT-002](../volume-03-architecture/07-platform-abstraction-layer.md) | PAL conformance coverage | Core | Surface-to-suite inventory check in CI | Volume 3 |
| [NFR-PORT-003](../volume-03-architecture/07-platform-abstraction-layer.md) | Single-binary deliverable with bounded prerequisites | MVP | Linkage inspection; clean-machine install tests | Volume 3 |
| [NFR-PORT-004](../volume-03-architecture/07-platform-abstraction-layer.md) | Platform-conditional code containment | Core | Automated scanner on every PR | Volume 3 |
| [RISK-ARCH-001](../volume-03-architecture/01-architecture-overview.md) | Layering erosion under delivery pressure | — | Severity: High; Status: Open | Volume 3 |
| [RISK-ARCH-002](../volume-03-architecture/02-port-interfaces.md) | Port interface churn during parallel authoring | — | Severity: High; Status: Open | Volume 3 |
| [RISK-ARCH-003](../volume-03-architecture/08-processes-concurrency-ipc.md) | Task Scheduler as a critical single point | — | Severity: High; Status: Open | Volume 3 |
| [RISK-ARCH-004](../volume-03-architecture/08-processes-concurrency-ipc.md) | Recovery divergence between recorded and actual state | — | Severity: High; Status: Open | Volume 3 |
| [RISK-PORT-001](../volume-03-architecture/07-platform-abstraction-layer.md) | macOS Intel viability gap | — | Severity: Medium; Status: Open — PENDING VALIDATION | Volume 3 |
| [RISK-PORT-002](../volume-03-architecture/07-platform-abstraction-layer.md) | Linux isolation primitive fragmentation | — | Severity: High; Status: Open — PENDING VALIDATION | Volume 3 |
| [RISK-PORT-003](../volume-03-architecture/07-platform-abstraction-layer.md) | PAL abstraction leaks blocking the Windows phase | — | Severity: High; Status: Open | Volume 3 |

## Volume 4 — Agent Runtime

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-AGT-001](../volume-04-agent-runtime/01-agent-engine.md) | Agent loop | MVP | Loop tests over provider doubles; replay divergence (SM-12); crash injection (SM-11); UC-01 E2E; audit-chain tests (SM-13) | Volume 4 |
| [FR-AGT-002](../volume-04-agent-runtime/01-agent-engine.md) | Turn handling and the message part vocabulary | MVP | Pipeline unit tests; message-part property tests; streaming/non-streaming contract tests; record validators | Volume 4 |
| [FR-AGT-003](../volume-04-agent-runtime/01-agent-engine.md) | Run interruption, pause, and resume | MVP | SM-11 crash-injection at randomized points; pause/resume integration; double-resume races; approval-expiry fixtures | Volume 4 |
| [FR-AGT-004](../volume-04-agent-runtime/01-agent-engine.md) | Sub-agent delegation | Beta | Delegation integration tests; permission-narrowing enforcement; audit-chain resolution; depth/budget property tests | Volume 4 |
| [FR-AGT-005](../volume-04-agent-runtime/01-agent-engine.md) | Run budget enforcement | MVP | Budget property tests; late-usage fixtures; scripted-spend integration; accounting consistency validators | Volume 4 |
| [FR-AGT-006](../volume-04-agent-runtime/01-agent-engine.md) | Workspace lifecycle over WorkspacePort | MVP | Filesystem fixtures (markers, nesting, permissions); concurrent-open tests; registry repair tests; Tier 1 matrix | Volume 4 |
| [FR-AGT-007](../volume-04-agent-runtime/02-planner.md) | Plan production | MVP | Golden decomposition tests; validation property tests; structured/text parity; failure-double integration | Volume 4 |
| [FR-AGT-008](../volume-04-agent-runtime/02-planner.md) | Plan revision | MVP | Revision protocol integration; transactional supersession (crash between steps); lineage validators; bound-breach tests | Volume 4 |
| [FR-AGT-009](../volume-04-agent-runtime/02-planner.md) | Plan inspection and approval interplay | MVP | Gate enforcement tests; mode × session-kind matrix; audit-chain validators; approval expiry fixtures | Volume 4 |
| [FR-AGT-010](../volume-04-agent-runtime/03-execution-engine.md) | Task scheduling and dispatch | MVP | Deterministic DAG execution; dispatch-refusal tests; cancellation storms; audit-chain validators; capacity matrix | Volume 4 |
| [FR-AGT-011](../volume-04-agent-runtime/03-execution-engine.md) | Task retry policy | MVP | Fault-injection matrices (retryable × side-effect); backoff property tests; gated-confirmation integration; exhaustion fixtures | Volume 4 |
| [FR-AGT-012](../volume-04-agent-runtime/03-execution-engine.md) | Cancellation, skip, and error propagation | MVP | Cancellation storms; wedged-child fixtures; cancel-vs-complete races; reason validators; leak gates | Volume 4 |
| [FR-AGT-013](../volume-04-agent-runtime/04-prompt-engine.md) | Versioned prompt templates and registry | MVP | Registry unit tests (precedence, shadowing, rejection); trust-gate integration; snapshot stability; golden diagnostics | Volume 4 |
| [FR-AGT-014](../volume-04-agent-runtime/04-prompt-engine.md) | Deterministic rendering with provenance | MVP | Cross-platform determinism property tests; slot/schema failure matrices; provenance round-trips in replay suite | Volume 4 |
| [FR-AGT-015](../volume-04-agent-runtime/05-core-state-machines.md) | Canonical machine enforcement | Core | Property-based machine tests; illegal-write attempts; replay validation fixtures; transition/event parity validators | Volume 4 |
| [NFR-AGT-001](../volume-04-agent-runtime/05-core-state-machines.md) | State transition legality under load | MVP | Property/race/crash suites with record validation; zero-violation gate per mainline commit | Volume 4 |
| [NFR-AGT-002](../volume-04-agent-runtime/05-core-state-machines.md) | Resume fidelity | MVP | SM-11-method crash-injection plus scripted resume; record diffing; confirmation assertions | Volume 4 |
| [NFR-AGT-003](../volume-04-agent-runtime/04-prompt-engine.md) | Prompt render determinism | MVP | ≥ 100 repeated renders per template per Tier 1 platform; golden render audit per release | Volume 4 |
| [RISK-AGT-001](../volume-04-agent-runtime/01-agent-engine.md) | Non-terminating or runaway agent loop | — | Mitigations tracked via limit/budget enforcement tests and soak runs | Volume 4 |
| [RISK-AGT-002](../volume-04-agent-runtime/02-planner.md) | Plan–execution divergence | — | Audit-chain resolution (SM-13); dispatch-refusal counters | Volume 4 |
| [RISK-AGT-003](../volume-04-agent-runtime/05-core-state-machines.md) | Resumption ambiguity across the crash boundary | — | SM-11 campaigns; pre/post record diffing; confirmation assertions | Volume 4 |
| [RISK-AGT-004](../volume-04-agent-runtime/04-prompt-engine.md) | Prompt template drift and override injection | — | Trust-gate tests; golden render diffs; override event audits | Volume 4 |
| [FR-WF-001](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) | Specification-driven development workflow | Beta | Workflow E2E suite over scripted doubles; gate-profile matrix; loop-budget and denial fixtures; audit-chain validation | Volume 4 |
| [FR-WF-002](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) | Workflow definition format and validation | Beta | Golden valid/invalid fixture suite; TOML→JSON round-trips; immutability and idempotence tests; trust-gating integration tests | Volume 4 |
| [FR-WF-003](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) | Workflow execution semantics | Beta | Deterministic engine tests over doubles; parallelism/serialization property tests; step-boundary crash injection; headless parity tests | Volume 4 |
| [FR-WF-004](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) | Workflow approval gates | Beta | Gate outcome matrix tests; headless prompt-absence tests; audit-chain checks; expiry timing tests | Volume 4 |
| [FR-WF-005](../volume-04-agent-runtime/07-workflow-run-state-machine.md) | Workflow Run state machine conformance | Beta | Transition-table property tests; event/transition matching audits; revision race tests | Volume 4 |
| [FR-WF-006](../volume-04-agent-runtime/07-workflow-run-state-machine.md) | Workflow interruption, recovery, and resume | Beta | Crash injection at randomized boundaries and mid-step; resume Approval fixtures; integrity-corruption fixtures; cross-machine resume tests | Volume 4 |
| [FR-WF-007](../volume-04-agent-runtime/07-workflow-run-state-machine.md) | Workflow cancellation and rollback | Beta | Cancellation storms; compensation-order property tests; failed-compensation fixtures; restore-point round-trips over GitPort doubles | Volume 4 |
| [FR-WF-008](../volume-04-agent-runtime/08-skill-engine-runtime.md) | Skill application in runs and workflows | Beta | Snapshot determinism and replay tests; gating matrix tests; mid-run mutation tests; composition-order golden tests | Volume 4 |
| [FR-WF-009](../volume-04-agent-runtime/09-task-scheduler.md) | Workflow scheduling and supervision | Beta | Scheduler-integration suite with cancellation storms and saturation fixtures; leak gates; waiting-state footprint tests; panic injection | Volume 4 |
| [FR-WF-010](../volume-04-agent-runtime/09-task-scheduler.md) | Durable timers and timeout enforcement | Beta | Timer property tests under injected clocks; kill-restart fixtures across deadlines; grant-vs-expiry races; pause/resume budget golden tests | Volume 4 |
| [NFR-WF-001](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) | Workflow format public-contract stability | Beta | Contract-diff of the definition JSON Schema per release; Volume 14 release audit | Volume 4 |
| [NFR-WF-002](../volume-04-agent-runtime/07-workflow-run-state-machine.md) | Workflow resume fidelity | Beta | Crash-injection suite with automated post-resume `step_states` audit | Volume 4 |
| [NFR-WF-003](../volume-04-agent-runtime/09-task-scheduler.md) | Workflow orchestration overhead | Beta | Instrumented no-op workflow benchmark; parked-run heap accounting on reference hardware | Volume 4 |
| [RISK-WF-001](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) | Process overhead deters SDD adoption | — | Severity: Medium; Status: Open | Volume 4 |
| [RISK-WF-002](../volume-04-agent-runtime/07-workflow-run-state-machine.md) | Irreversible external side effects defeat rollback | — | Severity: High; Status: Open | Volume 4 |
| [RISK-WF-003](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) | Approval-gate fatigue erodes the safety model | — | Severity: High; Status: Open | Volume 4 |

## Volume 5 — Providers, Models, and Authentication

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-PROV-001](../volume-05-providers-and-auth/01-provider-contract.md) | Provider contract | Core | Provider conformance suite; ADR-033 dependency checks; contract tests incl. cancellation | Volume 5 |
| [FR-PROV-002](../volume-05-providers-and-auth/01-provider-contract.md) | Adapter declaration and registration | Core | Registry validation tests; declaration-vs-behavior conformance checks | Volume 5 |
| [FR-PROV-010](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) | Capability declaration and the capability matrix | Core | Conformance honesty checks (SM-04); ADR-033 name-branch scan; resolution unit tests | Volume 5 |
| [FR-PROV-011](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) | Capability negotiation, verification, and degradation | MVP | Conformance degradation scenarios; refutation fault injection | Volume 5 |
| [FR-PROV-012](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) | Model discovery and the model registry | MVP | Reconciliation unit tests; recorded-fixture discovery tests; offline suite | Volume 5 |
| [FR-PROV-013](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) | Provider and model change notification | MVP | Event-before-output integration tests; SM-13 audit-chain test | Volume 5 |
| [FR-PROV-020](../volume-05-providers-and-auth/03-streaming-toolcalling-structured-outputs.md) | Streaming contract | MVP | Paced-fixture conformance; fault injection; SM-08 benchmarks | Volume 5 |
| [FR-PROV-021](../volume-05-providers-and-auth/03-streaming-toolcalling-structured-outputs.md) | Tool-calling normalization | MVP | Round-trip corpus per adapter; malformed-output fault injection | Volume 5 |
| [FR-PROV-022](../volume-05-providers-and-auth/03-streaming-toolcalling-structured-outputs.md) | Structured outputs | MVP | Schema corpus across modes; violation fixtures; retry-ceiling tests | Volume 5 |
| [FR-PROV-030](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) | Token usage accounting | MVP | Accounting-honesty conformance; SM-13 chain; CountTokens contract tests | Volume 5 |
| [FR-PROV-031](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) | Cost accounting and pricing tables | MVP | Resolution/arithmetic unit tests; config validation tests; labeling tests | Volume 5 |
| [FR-PROV-040](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) | Timeouts, rate limits, and retries | MVP | Fault injection with fake clocks; abort-on-timeout contract tests | Volume 5 |
| [FR-PROV-041](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) | Circuit breaker and health verification | MVP | Scripted failure sequences; single-probe concurrency tests | Volume 5 |
| [FR-PROV-042](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) | Routing and selection | MVP | Selection unit tests; SM-12 replay determinism; permission-path tests | Volume 5 |
| [FR-PROV-043](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) | Fallback and its guard rules | MVP | Per-guard integration scenarios; headless policy tests; SM-13 verification | Volume 5 |
| [FR-PROV-050](../volume-05-providers-and-auth/06-error-normalization.md) | Provider error normalization | Core | Fault-injection corpus; leak checks; envelope completeness check | Volume 5 |
| [NFR-PROV-001](../volume-05-providers-and-auth/01-provider-contract.md) | Provider integration effort | Beta | Timed reference-integration exercise at phase gates (SM-01) | Volume 5 |
| [NFR-PROV-002](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) | Local-model conformance | v1 | Conformance suite per release on ≥ 2 local serving paths (SM-04) | Volume 5 |
| [NFR-PROV-003](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) | Accounting completeness | MVP | Record-completeness validator over suite runs | Volume 5 |
| [NFR-PROV-004](../volume-05-providers-and-auth/06-error-normalization.md) | Error normalization coverage | MVP | Leak-detection assertions; per-code fault reconciliation | Volume 5 |
| [RISK-PROV-001](../volume-05-providers-and-auth/01-provider-contract.md) | Provider API drift breaks adapters | — | Risk register review at phase gates; live conformance runs | Volume 5 |
| [RISK-PROV-002](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) | Capability misdeclaration | — | Risk register review at phase gates; honesty checks | Volume 5 |
| [RISK-PROV-003](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) | Stale or wrong pricing data misleads users | — | Risk register review at phase gates; basis metrics | Volume 5 |
| [RISK-PROV-004](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) | Fallback amplifies cost or data exposure | — | Risk register review at phase gates; SM-13 audit trail | Volume 5 |
| [FR-AUTH-001](../volume-05-providers-and-auth/07-authentication-layer.md) | Official-mechanisms-only authentication | Core | Registry validation tests; conformance prohibited-mechanism cases; adapter review checklist; CI static checks | Volume 5 |
| [FR-AUTH-002](../volume-05-providers-and-auth/07-authentication-layer.md) | API key authentication | MVP | Intake-path integration tests; canary leak scan; CI env-indirection test; permission tests | Volume 5 |
| [FR-AUTH-003](../volume-05-providers-and-auth/07-authentication-layer.md) | OAuth 2.0 authorization code flow with PKCE | Beta | Mock authorization-server integration tests; grant-type static check; leak scan | Volume 5 |
| [FR-AUTH-004](../volume-05-providers-and-auth/07-authentication-layer.md) | OAuth 2.0 device authorization grant | Beta | Mock device-flow integration tests; SSH manual test at Beta gate; leak scan | Volume 5 |
| [FR-AUTH-005](../volume-05-providers-and-auth/07-authentication-layer.md) | Service accounts and managed identity | v1 | Mock token-exchange tests; platform-conditional fakes; per-provider validation record check | Volume 5 |
| [FR-AUTH-006](../volume-05-providers-and-auth/07-authentication-layer.md) | Enterprise proxies and trust anchors | Beta | Local authenticating-proxy integration tests; no-bypass static check | Volume 5 |
| [FR-AUTH-007](../volume-05-providers-and-auth/07-authentication-layer.md) | Temporary credentials | Beta | Fake-clock unit tests; expiry sweep test | Volume 5 |
| [FR-AUTH-008](../volume-05-providers-and-auth/07-authentication-layer.md) | Multiple authentication profiles | MVP | Precedence-matrix unit tests; ambiguity cases; SM-12 record check | Volume 5 |
| [FR-AUTH-009](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Credential storage and resolution through the Secret Store | MVP | SecretStorePort ordering contract tests with crash injection; canary scans; audit completeness check | Volume 5 |
| [FR-AUTH-010](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Token refresh | Beta | Mock token-endpoint tests (single-flight, rotation-on-use, offline); fake-clock tests; leak scan | Volume 5 |
| [FR-AUTH-011](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Credential rotation and revocation | MVP | Atomicity and cascade integration tests; idempotence tests; endpoint failure injection; audit chain | Volume 5 |
| [FR-PROV-080](../volume-05-providers-and-auth/09-provider-adapters-catalog.md) | Adapter catalog and phase classification | MVP | Adapter conformance suites; registry and validation tests; release audit against catalog | Volume 5 |
| [FR-PROV-081](../volume-05-providers-and-auth/09-provider-adapters-catalog.md) | Generic OpenAI-compatible adapter | MVP | Conformance suite (fixtures + live local server); framing fault injection; offline suite | Volume 5 |
| [FR-PROV-082](../volume-05-providers-and-auth/09-provider-adapters-catalog.md) | Anthropic adapter | MVP | Conformance suite with recorded fixtures; scheduled live smoke; vendor-type leak check | Volume 5 |
| [FR-PROV-083](../volume-05-providers-and-auth/09-provider-adapters-catalog.md) | Ollama adapter | MVP | Conformance suite against pinned local Ollama; offline suite; dependency audit | Volume 5 |
| [FR-PROV-084](../volume-05-providers-and-auth/10-local-and-offline-operation.md) | Local provider operation | MVP | Offline suite (UC-09 journey); socket-activity assertion; locality unit tests; permission parity test | Volume 5 |
| [FR-PROV-085](../volume-05-providers-and-auth/10-local-and-offline-operation.md) | Offline behavior of the provider layer | MVP | Offline suite per SM-05 method; fault injection (portals, partial connectivity); retry-count assertions | Volume 5 |
| [NFR-AUTH-001](../volume-05-providers-and-auth/08-credential-lifecycle.md) | No plaintext credential material at rest | MVP | Canary-secret scan over all written artifacts, every gated CI run | Volume 5 |
| [NFR-AUTH-002](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Credential redaction in logs, events, errors, and memory records | MVP | Canary scan over observability channels with debug verbosity, every gated CI run | Volume 5 |
| [NFR-AUTH-003](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Credential resolution latency | Beta | Per-backend benchmark harness, p95, per release | Volume 5 |
| [RISK-AUTH-001](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Secret Store backend unavailability and fallback overuse | — | Severity: Medium; Status: Open | Volume 5 |
| [RISK-AUTH-002](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Credential leakage through logs, errors, or crash artifacts | — | Severity: High; Status: Open | Volume 5 |
| [RISK-AUTH-003](../volume-05-providers-and-auth/08-credential-lifecycle.md) | Provider authentication mechanism drift | — | Severity: Medium; Status: Open | Volume 5 |
| [RISK-PROV-080](../volume-05-providers-and-auth/09-provider-adapters-catalog.md) | Adapter catalog drift against provider realities | — | Severity: High; Status: Open | Volume 5 |
| [RISK-PROV-081](../volume-05-providers-and-auth/10-local-and-offline-operation.md) | Local model and server capability variance | — | Severity: High; Status: Open | Volume 5 |

## Volume 6 — Tools, MCP, Skills, and Plugins

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-TOOL-001](../volume-06-tools-mcp-skills-plugins/01-tool-sdk-and-contract.md) | Tool contract | Core | Tool contract conformance suite over all built-ins and SDK fixtures; registration fuzzing; mediation probes | Volume 6 |
| [FR-TOOL-002](../volume-06-tools-mcp-skills-plugins/01-tool-sdk-and-contract.md) | Declaration and payload schema validation | Core | Official JSON Schema vectors per draft; mutation and nonconformance fixtures; network-isolation assertion | Volume 6 |
| [FR-TOOL-003](../volume-06-tools-mcp-skills-plugins/01-tool-sdk-and-contract.md) | Tool naming, namespaces, and resolution | Core | Grammar fuzzing; collision matrices; alias fixtures; canonical-identity audit checks | Volume 6 |
| [FR-TOOL-004](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) | Tool registration, availability, and enablement | MVP | Provider-lifecycle integration tests; per-scope persistence tests; audit-chain resolution | Volume 6 |
| [FR-TOOL-005](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) | Permission mediation for every invocation | MVP | Per-mode permission matrix tests; unmediated-execution probes; expiry/revocation race tests | Volume 6 |
| [FR-TOOL-006](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) | Execution limits, sandbox placement, and teardown | MVP | Fault-injection suite (hang, fork-bomb, flood); teardown timing per platform; containment-level record assertions | Volume 6 |
| [FR-TOOL-007](../volume-06-tools-mcp-skills-plugins/03-builtin-tools-catalog.md) | Built-in tool catalog and phasing | MVP | Per-tool conformance and golden fixtures; offline suite; Tier 1 matrix; recorded-API contract tests | Volume 6 |
| [FR-TOOL-008](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) | Tool Invocation machine conformance | MVP | Transition-matrix property tests; crash injection at every boundary; race tests; audit-chain resolution | Volume 6 |
| [FR-SDK-001](../volume-06-tools-mcp-skills-plugins/01-tool-sdk-and-contract.md) | Extension SDK | Beta | CI tutorial walkthrough (SM-02 method); mirror-equivalence check; fixture self-tests; timed gate exercises | Volume 6 |
| [NFR-SDK-001](../volume-06-tools-mcp-skills-plugins/01-tool-sdk-and-contract.md) | Tool creation time (SM-02) | Beta | Timed phase-gate exercise; CI tutorial walkthrough; contribution-record sampling | Volume 6 |
| [NFR-TOOL-001](../volume-06-tools-mcp-skills-plugins/03-builtin-tools-catalog.md) | Built-in tool contract conformance | MVP | Conformance suite in CI per release across Tier 1 matrix; release audit attachment | Volume 6 |
| [NFR-TOOL-002](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) | Invocation record and event completeness | MVP | Audit-chain test over integration/E2E and crash-injection runs per release | Volume 6 |
| [RISK-TOOL-001](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) | Dishonest or over-broad tool declarations | — | Severity: Critical; Status: Open | Volume 6 |
| [RISK-TOOL-002](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) | Name shadowing and typosquatting across origins | — | Severity: High; Status: Open | Volume 6 |
| [RISK-TOOL-003](../volume-06-tools-mcp-skills-plugins/03-builtin-tools-catalog.md) | External service API drift breaking integration tools | — | Severity: High; Status: Open — PENDING VALIDATION per service | Volume 6 |
| [RISK-TOOL-004](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) | Output flooding and resource exhaustion through tools | — | Severity: High; Status: Open | Volume 6 |
| [FR-MCP-001](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) | MCP client support (keystone) | Beta | MCP conformance suite and interop job (SM-15); sandbox-launch, credential, and reconnection integration tests; secret-scan over sinks | Volume 6 |
| [FR-MCP-002](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) | MCP transports and connection establishment | Beta | Fixture-server integration for both transports; fault injection per phase; sandbox-policy assertions | Volume 6 |
| [FR-MCP-003](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) | Discovery and bridging of tools, resources, and prompts | Beta | Conformance fixtures (listings, pagination, listChanged); schema-validation units; bridge-teardown integration; interop suite (SM-15b) | Volume 6 |
| [FR-MCP-004](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) | MCP server authorization | Beta | Authorized fixture servers; secret-scan assertions; permission-denial tests; OAuth PENDING VALIDATION register cross-check | Volume 6 |
| [FR-MCP-005](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) | Connection health, server logs, and maintenance operations | Beta | Fault injection (probe failures, log floods); update/uninstall integration over fixture packages; redaction assertions | Volume 6 |
| [FR-MCP-006](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) | MCP trust gating and isolation | Beta | Pre-approval and post-drift enforcement probes; sandbox environment assertions; audit-chain tests (SM-16 pattern) | Volume 6 |
| [FR-MCP-007](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) | MCP conformance test program | Beta | Suite presence, coverage mapping, and gating wiring audited at phase gates (Volume 13 release qualification) | Volume 6 |
| [FR-MCP-008](../volume-06-tools-mcp-skills-plugins/10-state-machines.md) | MCP Client Connection machine conformance | Beta | Transition-matrix property tests; per-phase fault injection; crash-injection recovery; routing-refusal probes in every non-ready state | Volume 6 |
| [FR-SKILL-001](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) | Skill format and manifest (keystone) | Beta | Golden manifest corpus (valid + invalid); hash tamper tests; schema round-trip against the SDK mirror | Volume 6 |
| [FR-SKILL-002](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) | Skill loading, requirement resolution, and activation | Beta | Registry-gap integration tests; run-record inspection; resolver-determinism property tests | Volume 6 |
| [FR-SKILL-003](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) | Skill inheritance, composition, and overrides | Beta | Composition conflict matrix; determinism and cycle-detection property tests; golden composition reports | Volume 6 |
| [FR-SKILL-004](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) | Skill testing and fixtures | Beta | Runner self-tests; SDK template test in CI; offline-suite inclusion | Volume 6 |
| [FR-SKILL-005](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) | Skill distribution and deprecation | Beta | End-to-end package tests; signature policy matrix; deprecation selection tests; audit-chain inspection | Volume 6 |
| [FR-PLUG-001](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | Plugin runtime over the Andromeda Runtime Protocol (keystone) | Beta | ARP conformance fixtures (handshake, streaming, cancellation, reserved codes); chaos tests; cross-language smoke plugin; SM-03 timed exercise | Volume 6 |
| [FR-PLUG-002](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | ARP handshake, version negotiation, and method conformance | Beta | Version-parameterized conformance fixtures; SDK cross-tests; negotiation property tests | Volume 6 |
| [FR-PLUG-003](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | Plugin permission mediation and sandbox containment | Beta | Sandbox assertion fixtures (env, filesystem, network, process tree); audit-chain tests; manifest-bound consistency tests; secret-scan on debug captures | Volume 6 |
| [FR-PLUG-004](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | Plugin supervision, health, and restart | Beta | Chaos tests (kill, hang, crash-loop); restart-policy property tests; persisted-state recovery tests; leak checks | Volume 6 |
| [FR-PLUG-005](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) | Extension package operations | Beta | End-to-end operation tests over all four source kinds; crash injection per state; upgrade/rollback tests; cascade-removal tests; audit-chain inspection | Volume 6 |
| [FR-PLUG-006](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) | Package sources, discovery, and dependency resolution | Beta | Resolver property tests (determinism, bounds, intersection); multi-source integration incl. divergent checksums; offline-cache tests; permission-denial tests | Volume 6 |
| [FR-PLUG-007](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) | Package verification and integrity | Beta | Verification matrix (checksum × signature × policy × source flags); tamper fixtures; offline verification tests; audit-chain assertions | Volume 6 |
| [FR-PLUG-008](../volume-06-tools-mcp-skills-plugins/10-state-machines.md) | Plugin lifecycle machine conformance | Beta | Transition-matrix property tests; chaos suite; crash-injection reconciliation tests; event-sequence reconstruction | Volume 6 |
| [FR-PLUG-009](../volume-06-tools-mcp-skills-plugins/10-state-machines.md) | Package installation machine conformance | Beta | Transition-matrix property tests; crash injection at every boundary (fresh + upgrade); bundle-atomicity tests; staging-cleanup assertions | Volume 6 |
| [NFR-MCP-001](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) | MCP conformance pass rate (SM-15a) | v1 | Conformance suite in CI per release and per SDK bump; per-revision reporting | Volume 6 |
| [NFR-MCP-002](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) | MCP reference-server interoperation (SM-15b) | v1 | Weekly scheduled interop job; mandatory run with scorecard at release qualification | Volume 6 |
| [NFR-SKILL-001](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) | Skill validation and composition latency | Beta | Benchmark harness over the fixture skill corpus, p95, per release | Volume 6 |
| [NFR-PLUG-001](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | Plugin creation time (SM-03) | v1 | Timed phase-gate exercise against the SDK plugin template; CI tutorial walkthrough | Volume 6 |
| [NFR-PLUG-002](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | ARP invocation dispatch overhead | Beta | Instrumented no-op fixture plugin benchmark, 1000 invocations, p95, per release | Volume 6 |
| [NFR-PLUG-003](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) | Extension package operation latency | Beta | Benchmark harness over fixture packages and indexes, p95, per release | Volume 6 |
| [RISK-MCP-001](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) | Malicious or compromised MCP server | — | Severity: High; Status: Open | Volume 6 |
| [RISK-MCP-002](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) | MCP specification churn and revision skew | — | Severity: Medium; Status: Open | Volume 6 |
| [RISK-SKILL-001](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) | Malicious or manipulative skill content | — | Severity: High; Status: Open | Volume 6 |
| [RISK-PLUG-001](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | Malicious or compromised plugin | — | Severity: High; Status: Open | Volume 6 |
| [RISK-PLUG-002](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) | Protocol evolution breaking the plugin ecosystem | — | Severity: Medium; Status: Open | Volume 6 |
| [RISK-PLUG-003](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) | Extension supply-chain compromise | — | Severity: High; Status: Open | Volume 6 |

## Volume 7 — Memory, Context, and Indexing

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-MEM-001](../volume-07-memory-context-indexing/01-memory-model.md) | Memory model | MVP | MemoryStorePort contract suite; isolation and ordering-determinism property tests; offline degraded-path suite | Volume 7 |
| [FR-MEM-002](../volume-07-memory-context-indexing/01-memory-model.md) | Provenance, trust, and source attribution | MVP | Stamping/ordering unit tests; adversarial supersession tests; CLI/TUI provenance inspection tests | Volume 7 |
| [FR-MEM-003](../volume-07-memory-context-indexing/01-memory-model.md) | Secret and sensitive-content exclusion | MVP | Redaction-gate fixture tests; bypass-attempt property tests over all write paths | Volume 7 |
| [FR-MEM-004](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Ingestion and normalization | MVP | Ingest contract tests (transactionality, cancellation); mode-matrix tests; normalization determinism property tests | Volume 7 |
| [FR-MEM-005](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Deduplication, conflict resolution, and supersession versioning | MVP | Chain-linearity and idempotency property tests; head-contention concurrency tests; trust-guard adversarial tests | Volume 7 |
| [FR-MEM-006](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Compression and summarization (consolidation) | Beta | Integration tests with scripted provider doubles; offline deferral tests; archive-link integrity property tests | Volume 7 |
| [FR-MEM-007](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Retention and expiration | MVP | Idempotency/forward-only property tests; time-travel fixtures; cascade suite | Volume 7 |
| [FR-MEM-008](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Deletion and purge cascade | MVP | NFR-MEM-002 measurement suite; crash-injection cascade tests; permission enforcement tests | Volume 7 |
| [FR-MEM-009](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Export and portability | MVP | Round-trip suites; determinism tests; permission and encryption fixture tests | Volume 7 |
| [FR-MEM-010](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Offline memory operation | MVP | Offline suite (OS-level disablement) over the full operation surface; egress capture | Volume 7 |
| [FR-CTX-001](../volume-07-memory-context-indexing/03-context-manager.md) | Context assembly | MVP | Determinism/property tests; golden assembly fixtures; replay divergence tests; source-failure injection | Volume 7 |
| [FR-CTX-002](../volume-07-memory-context-indexing/03-context-manager.md) | Token budgeting and model window enforcement | MVP | Budget arithmetic unit tests; counting-unavailable contract-double tests; NFR-CTX-001 measurement | Volume 7 |
| [FR-CTX-003](../volume-07-memory-context-indexing/03-context-manager.md) | Deduplication and compression of context | MVP | Dedup/ladder property tests; containment golden fixtures; no-inference instrumentation assertions | Volume 7 |
| [FR-CTX-004](../volume-07-memory-context-indexing/03-context-manager.md) | Freshness, trust, provenance, and conflict detection | MVP | Conflict-matrix fixtures; staleness injection tests; adversarial trust tests | Volume 7 |
| [FR-CTX-005](../volume-07-memory-context-indexing/03-context-manager.md) | User pinning and exclusion | MVP | Command integration tests; assembly filter property tests; resume round-trips; cap unit tests | Volume 7 |
| [FR-CTX-006](../volume-07-memory-context-indexing/03-context-manager.md) | Tool-result, large-file, and binary-content handling | MVP | Transformation golden fixtures; detection determinism property tests; byte-level binary assertions | Volume 7 |
| [FR-CTX-007](../volume-07-memory-context-indexing/03-context-manager.md) | Context snapshots and reproducibility | MVP | Record-completeness validator (SM-12 method); replay divergence tests; crash injection between assembly and send | Volume 7 |
| [FR-IDX-001](../volume-07-memory-context-indexing/04-indexing-engine.md) | Indexing engine | MVP | IndexerPort contract suite; generation-consistency tests under concurrency; rebuild equivalence tests; scope fixtures; ADR-020 scale benchmarks | Volume 7 |
| [FR-IDX-002](../volume-07-memory-context-indexing/04-indexing-engine.md) | Chunking | MVP | Golden chunk fixtures; cross-platform determinism tests; boundary property tests | Volume 7 |
| [FR-IDX-003](../volume-07-memory-context-indexing/04-indexing-engine.md) | Embeddings and semantic index construction | MVP | Provider-double contract tests (capability refusal, dimension mismatch, retry); hash-gating property tests; cost-attribution tests | Volume 7 |
| [FR-IDX-004](../volume-07-memory-context-indexing/04-indexing-engine.md) | Incremental updates and invalidation | MVP | Watch-feed integration tests with change storms; staleness property tests; invalidation exclusion tests; cache-drop equivalence | Volume 7 |
| [FR-IDX-005](../volume-07-memory-context-indexing/04-indexing-engine.md) | Operation without embeddings and without Internet | MVP | Offline suite over lexical and both semantic conditions; egress capture; drain-on-reconnect tests | Volume 7 |
| [FR-IDX-006](../volume-07-memory-context-indexing/05-index-state-machine.md) | Index state machine conformance and recovery | MVP | State-machine property suite over randomized operation/fault sequences; crash-injection harness; recovery idempotence tests | Volume 7 |
| [NFR-MEM-001](../volume-07-memory-context-indexing/01-memory-model.md) | Memory privacy and locality | MVP | Egress capture during offline and instrumented-online suites; static dependency check | Volume 7 |
| [NFR-MEM-002](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Deletion completeness | MVP | Deletion suite with exhaustive port queries and database scans; crash-injection variant | Volume 7 |
| [NFR-CTX-001](../volume-07-memory-context-indexing/03-context-manager.md) | Context window compliance | MVP | Overflow classification over integration/E2E suites; seeded infeasibility fault matrix | Volume 7 |
| [NFR-IDX-001](../volume-07-memory-context-indexing/04-indexing-engine.md) | Offline indexing guarantee | MVP | Offline suite with egress capture; static ProviderPort-reference verification on lexical paths | Volume 7 |
| [NFR-IDX-002](../volume-07-memory-context-indexing/04-indexing-engine.md) | Generation integrity under interruption | MVP | Crash/cancel-injection harness with generation-consistency verification | Volume 7 |
| [RISK-MEM-001](../volume-07-memory-context-indexing/01-memory-model.md) | Memory poisoning and stale-knowledge drift | — | Severity: High; Status: Open | Volume 7 |
| [RISK-MEM-002](../volume-07-memory-context-indexing/02-memory-lifecycle.md) | Unbounded memory and embedding growth | — | Severity: Medium; Status: Open | Volume 7 |
| [RISK-CTX-001](../volume-07-memory-context-indexing/03-context-manager.md) | Critical-content eviction under budget pressure | — | Severity: High; Status: Open | Volume 7 |
| [RISK-IDX-001](../volume-07-memory-context-indexing/04-indexing-engine.md) | Corpus scale beyond the ADR-020 assumption | — | Severity: Medium; Status: Open — ANN successor path PENDING VALIDATION per ADR-020 | Volume 7 |
| [RISK-IDX-002](../volume-07-memory-context-indexing/04-indexing-engine.md) | Embedding-space drift across provider model revisions | — | Severity: Medium; Status: Open — vector-stability assumption recorded below | Volume 7 |

## Volume 8 — CLI and TUI

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-CLI-001](../volume-08-cli-and-tui/01-cli-architecture.md) | CLI command grammar | Core | Grammar golden tests over the full tree; argv fuzzing; SM-20 contract-diff; consolidation audit of chapters 03–06 | Volume 8 |
| [FR-CLI-002](../volume-08-cli-and-tui/01-cli-architecture.md) | Runtime mediation and driver parity | Core | ADR-033 dependency checks; CLI/IPC parity record comparison; permission mediation tests (SM-16b) | Volume 8 |
| [FR-CLI-003](../volume-08-cli-and-tui/01-cli-architecture.md) | Root command and TUI hand-off | MVP | teatest launch tests; pipe/CI matrix asserting exit 2 and clean stdout; `--version` alias golden | Volume 8 |
| [FR-CLI-004](../volume-08-cli-and-tui/01-cli-architecture.md) | Extension-contributed commands | Beta | SDK extension-command conformance suite; namespace-closure tests; permission mediation over extension origins | Volume 8 |
| [FR-CLI-005](../volume-08-cli-and-tui/02-cli-conventions.md) | Global flags and invocation modes | Core | Global-flag matrix over the tree; cancellation tests; ConfigPort attribution assertions | Volume 8 |
| [FR-CLI-006](../volume-08-cli-and-tui/02-cli-conventions.md) | Structured JSON output for every command | Core | Schema conformance matrix (success + failure per command); NDJSON strict parsing; SM-20 contract-diff; redaction leak tests | Volume 8 |
| [FR-CLI-007](../volume-08-cli-and-tui/02-cli-conventions.md) | Stream discipline: stdout, stderr, exit code | Core | Stream-classification matrix; EPIPE fault injection; redirection goldens | Volume 8 |
| [FR-CLI-008](../volume-08-cli-and-tui/02-cli-conventions.md) | Verbosity modes: quiet, verbose, debug | MVP | Recording-parity tests; redaction leak tests; per-level stderr goldens | Volume 8 |
| [FR-CLI-009](../volume-08-cli-and-tui/02-cli-conventions.md) | Non-interactive and CI modes | MVP | Decision-table unit tests; NFR-CLI-003 prompt-free matrix; policy fixture parity tests | Volume 8 |
| [FR-CLI-010](../volume-08-cli-and-tui/02-cli-conventions.md) | Confirmation behavior | MVP | PTY prompt-driving tests; non-interactive matrix; audit-record assertions; golden prompt texts | Volume 8 |
| [FR-CLI-011](../volume-08-cli-and-tui/02-cli-conventions.md) | Environment variables | MVP | Environment matrix; precedence tests vs. ConfigPort attribution; truthy-parser unit tests | Volume 8 |
| [FR-CLI-012](../volume-08-cli-and-tui/02-cli-conventions.md) | Shell completion | MVP | Per-shell completion harness in CI; dynamic-completion fixtures; silent-empty fault injection | Volume 8 |
| [FR-CLI-013](../volume-08-cli-and-tui/03-cli-commands-core.md) | Core command family behavior | MVP | Per-command goldens and schema conformance; PTY prompt tests; non-interactive matrix; UC-01/UC-07/UC-11 E2E | Volume 8 |
| [FR-CLI-014](../volume-08-cli-and-tui/04-cli-commands-platform.md) | Platform command family behavior | MVP | Grammar goldens; tampered-artifact and stopped-server fixtures; confirmation matrix; UC-12 E2E at Beta | Volume 8 |
| [FR-CLI-015](../volume-08-cli-and-tui/05-cli-commands-data.md) | Data command family behavior | MVP | Offline suite (SM-05); export schema conformance; permission-denial fixtures; redaction leak tests | Volume 8 |
| [FR-CLI-016](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) | Maintenance command family behavior | MVP | Doctor fixture matrix; update E2E with tampered artifacts and interrupts (SM-18/SM-19); cold-start benchmark (SM-06a); offline suite | Volume 8 |
| [FR-UX-001](../volume-08-cli-and-tui/02-cli-conventions.md) | Error presentation standard | MVP | Golden-format tests per error family; leak tests; JSON/human parity assertions | Volume 8 |
| [FR-UX-002](../volume-08-cli-and-tui/02-cli-conventions.md) | Terminal capability adaptation and paging | MVP | Byte-classification over TTY/pipe matrix; styled/plain parity diff; pager fault injection | Volume 8 |
| [FR-UX-003](../volume-08-cli-and-tui/02-cli-conventions.md) | Progress reporting outside the TUI | MVP | PTY/pipe capture with byte classification; heartbeat cadence with mock clocks; recording-parity assertions | Volume 8 |
| [NFR-CLI-001](../volume-08-cli-and-tui/02-cli-conventions.md) | Structured-output schema stability | Beta | Contract-diff of published schemas per release; CI validation of emitted output against schemas | Volume 8 |
| [NFR-CLI-002](../volume-08-cli-and-tui/02-cli-conventions.md) | Help and reference completeness | MVP | Automated help-coverage walk of the tree in CI; docs generation gate | Volume 8 |
| [NFR-CLI-003](../volume-08-cli-and-tui/02-cli-conventions.md) | Prompt-free non-interactive operation | MVP | Piped full-command matrix instrumented for TTY reads, with confirmation/approval fixtures | Volume 8 |
| [RISK-CLI-001](../volume-08-cli-and-tui/01-cli-architecture.md) | Grammar and surface sprawl across releases | — | Severity: High; Status: Open | Volume 8 |
| [RISK-CLI-002](../volume-08-cli-and-tui/01-cli-architecture.md) | TUI hand-off misdetection corrupting scripted output | — | Severity: Medium; Status: Open | Volume 8 |
| [RISK-CLI-003](../volume-08-cli-and-tui/02-cli-conventions.md) | Structured-output drift breaking automation | — | Severity: High; Status: Open | Volume 8 |
| [FR-TUI-001](../volume-08-cli-and-tui/07-tui-architecture.md) | TUI shell | MVP | teatest/v2 golden frames; PTY integration tests incl. panic/kill restoration; event assertions | Volume 8 |
| [FR-TUI-002](../volume-08-cli-and-tui/07-tui-architecture.md) | Panel system and layout manager | MVP | Golden frames per layout class; geometry property tests | Volume 8 |
| [FR-TUI-003](../volume-08-cli-and-tui/07-tui-architecture.md) | Navigation and focus model | MVP | Scripted interaction suites; focus-trap and typed-ahead-discard tests | Volume 8 |
| [FR-TUI-004](../volume-08-cli-and-tui/07-tui-architecture.md) | Keyboard command map | MVP | Scripted keystroke matrix; random-sequence fuzz for panics/unintended actions | Volume 8 |
| [FR-TUI-005](../volume-08-cli-and-tui/07-tui-architecture.md) | Mouse input | MVP | Synthetic mouse-event tests; PTY capture of reporting sequences | Volume 8 |
| [FR-TUI-006](../volume-08-cli-and-tui/07-tui-architecture.md) | Resize and small-terminal behavior | MVP | Resize scripts with frame capture; SIGWINCH integration; purity property test | Volume 8 |
| [FR-TUI-007](../volume-08-cli-and-tui/07-tui-architecture.md) | Runtime event rendering and streaming pipeline | MVP | Load tests with mock streaming provider; overflow/fault injection; byte-equality vs persisted records | Volume 8 |
| [FR-TUI-008](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) | Theme configuration and resolution | MVP | Configuration-matrix PTY tests; watch-repaint test; validation exit-code tests | Volume 8 |
| [FR-TUI-009](../volume-08-cli-and-tui/09-wireframes-core.md) | Core screen inventory and content contract | MVP | Golden-frame enumeration suite; permission-mediation instrumentation; fault-injection screen matrix | Volume 8 |
| [FR-UX-040](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) | Closed design-token vocabulary with fixed Danger | MVP | Golden-frame color scanning; color-literal grep gate; contrast recomputation script | Volume 8 |
| [FR-UX-041](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) | Token-to-ANSI degradation tiers | MVP | Per-tier golden frames; SGR-legality scanners | Volume 8 |
| [FR-UX-042](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) | Light-terminal fallback theme | MVP | Light-mode golden frames; scripted OSC 11 responder; contrast scanner | Volume 8 |
| [FR-UX-043](../volume-08-cli-and-tui/09-wireframes-core.md) | Splash and identity surfaces | MVP | Golden frames for splash variants; policy-matrix start tests; timing assertions | Volume 8 |
| [NFR-TUI-001](../volume-08-cli-and-tui/07-tui-architecture.md) | Small-terminal functional completeness | MVP | Scripted 80×24 traversal of all screens/actions; compact-class smoke traversal | Volume 8 |
| [NFR-TUI-002](../volume-08-cli-and-tui/07-tui-architecture.md) | Rendering determinism for golden-frame testing | MVP | Double-run frame diff across the golden matrix in CI | Volume 8 |
| [NFR-UX-040](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) | Contrast and non-color redundancy | MVP | Automated contrast computation over tables and scanned frames; danger-marker frame scan | Volume 8 |
| [RISK-TUI-001](../volume-08-cli-and-tui/07-tui-architecture.md) | Presentation-state monolith erodes maintainability and latency | MVP | Golden-suite runtime and SM-07 trend monitoring; review discipline | Volume 8 |
| [RISK-TUI-002](../volume-08-cli-and-tui/07-tui-architecture.md) | Event flood renders the TUI unresponsive or misleading | MVP | Load tests at 10× delta rates; overflow counters | Volume 8 |
| [RISK-TUI-003](../volume-08-cli-and-tui/09-wireframes-core.md) | Wireframe and implementation drift | MVP | FR-TUI-009 enumeration suite; consolidation/release audits | Volume 8 |
| [RISK-UX-040](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) | User-palette variance at the ansi16 tier breaks contrast or brand | MVP | Terminal-matrix testing; doctor tier reporting; user reports | Volume 8 |
| [FR-TUI-060](../volume-08-cli-and-tui/10-wireframes-platform.md) | Platform screen catalog and management frame | MVP | teatest golden frames per screen/tier/glyph set; registry-coverage test; driver-parity audit | Volume 8 |
| [FR-TUI-061](../volume-08-cli-and-tui/10-wireframes-platform.md) | Command palette | MVP | teatest interaction scripts; registry coverage; latency measurement per NFR-UX-077 | Volume 8 |
| [FR-TUI-062](../volume-08-cli-and-tui/10-wireframes-platform.md) | Quick actions | Beta | teatest ranking/staleness/confirmation scripts | Volume 8 |
| [FR-TUI-063](../volume-08-cli-and-tui/10-wireframes-platform.md) | Help overlay and keybinding reference | MVP | Registry-diff coverage test; golden frames; offline assertion; keymap-rebind test | Volume 8 |
| [FR-TUI-064](../volume-08-cli-and-tui/10-wireframes-platform.md) | Error center and recovery screens | MVP | Crash-injection suite (SM-11 method) via teatest; envelope golden tests; clipboard audit assertion | Volume 8 |
| [FR-TUI-065](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | Accessible output mode | Beta | Byte-stream sequence denylist; registry-parity traversal; AT validation pass (PENDING VALIDATION) | Volume 8 |
| [FR-TUI-066](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | No-color and monochrome operation | MVP | Byte classification; parity diff across tiers; attribute-free terminfo fixture | Volume 8 |
| [FR-TUI-067](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | Glyph tiers and Unicode fallback | MVP | Byte-inventory scans per set; resolution matrix unit tests; width fixtures | Volume 8 |
| [FR-TUI-068](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | SSH, multiplexer, and non-TTY operation | MVP | Remote-PTY golden suite over SSH/tmux/screen; clipboard fault scripts; CI refusal tests | Volume 8 |
| [FR-UX-070](../volume-08-cli-and-tui/11-interaction-patterns.md) | Streaming output rendering | MVP | Mock-stream teatest scripts; SM-08 harness; escape-injection fixture corpus | Volume 8 |
| [FR-UX-071](../volume-08-cli-and-tui/11-interaction-patterns.md) | Spinners and progress bars | MVP | Time-controlled teatest scripts; reduced-motion golden frames; stall injection | Volume 8 |
| [FR-UX-072](../volume-08-cli-and-tui/11-interaction-patterns.md) | Modal overlays and confirmation tiers | MVP | Per-tier interaction scripts; keystroke-buffer injection; z-order golden frames | Volume 8 |
| [FR-UX-073](../volume-08-cli-and-tui/11-interaction-patterns.md) | Toasts | MVP | Timing scripts (dismissal, queue, coalescing, modal pause); storm injection | Volume 8 |
| [FR-UX-074](../volume-08-cli-and-tui/11-interaction-patterns.md) | Canonical view states | MVP | Signal-injection fixtures; offline suite assertions; per-panel state scripts | Volume 8 |
| [FR-UX-075](../volume-08-cli-and-tui/11-interaction-patterns.md) | Copy and paste | MVP | Per-kind scripts; sanitization corpus; mechanism fakes; event payload assertions | Volume 8 |
| [FR-UX-076](../volume-08-cli-and-tui/11-interaction-patterns.md) | Data navigation: search, filtering, pagination, virtualization | MVP | Synthetic 100k-row stores; memory accounting; facet/smart-case unit tests; fetch-fault injection | Volume 8 |
| [NFR-TUI-069](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | Keyboard reachability and focus visibility | MVP | Registry traversal in teatest; color-stripped focus-marker analysis on golden corpus | Volume 8 |
| [NFR-TUI-070](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | Terminal compatibility conformance | Beta | Compatibility suite on the Tier A terminal set; probe-versus-matrix comparison per release | Volume 8 |
| [NFR-UX-077](../volume-08-cli-and-tui/11-interaction-patterns.md) | Interaction feedback deadline | MVP | SM-07 replay harness extended with feedback classification | Volume 8 |
| [NFR-UX-078](../volume-08-cli-and-tui/11-interaction-patterns.md) | Virtualized view memory ceiling | MVP | Benchmark-harness process accounting on 100 vs 100,000-record stores; render-state inspection | Volume 8 |
| [RISK-TUI-071](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | Assistive-technology experience remains impractical despite accessible output mode | — | Severity: High; Status: Open | Volume 8 |
| [RISK-TUI-072](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) | Character-width divergence corrupts layout in edge terminals | — | Severity: Medium; Status: Open | Volume 8 |
| [RISK-UX-079](../volume-08-cli-and-tui/11-interaction-patterns.md) | Clipboard exposure of sensitive content | — | Severity: High; Status: Open | Volume 8 |
| [RISK-UX-080](../volume-08-cli-and-tui/11-interaction-patterns.md) | Stale or masked view states misrepresenting reality | — | Severity: High; Status: Open | Volume 8 |

## Volume 9 — Security

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [RISK-SEC-001](../volume-09-security/02-threats-injection.md) | Prompt injection (direct) | — | Injection-corpus tests (no unmediated side effect); permission-matrix and approval-flow tests; audit-chain tests (SM-13) | Volume 9 |
| [RISK-SEC-002](../volume-09-security/02-threats-injection.md) | Indirect prompt injection | — | Indirect-injection fixtures over files/issues/tool results; provenance tests; egress-decision tests | Volume 9 |
| [RISK-SEC-003](../volume-09-security/02-threats-injection.md) | Tool injection | — | Input-schema mutation/fuzzing; sink-control tests; no-shell-interpolation assertion | Volume 9 |
| [RISK-SEC-004](../volume-09-security/02-threats-injection.md) | Tool poisoning | — | Descriptor-pinning re-consent tests; grant-scoping tests; declaration-diff conformance | Volume 9 |
| [RISK-SEC-005](../volume-09-security/02-threats-injection.md) | MCP poisoning | — | MCP conformance and trust-gating tests (SM-15); served-injection fixtures; context-exposure tests | Volume 9 |
| [RISK-SEC-006](../volume-09-security/02-threats-injection.md) | Malicious model output | — | Structured-output conformance/mutation; terminal-escape fixtures; tool-argument validation | Volume 9 |
| [RISK-SEC-007](../volume-09-security/02-threats-injection.md) | Memory poisoning | — | Provenance-tracking tests; poisoning fixtures with delete; secret-in-memory refusal tests | Volume 9 |
| [RISK-SEC-008](../volume-09-security/02-threats-injection.md) | Index poisoning | — | Index-hit provenance tests; invalidate-and-rebuild tests; untrusted-labeled retrieval tests | Volume 9 |
| [RISK-SEC-009](../volume-09-security/02-threats-injection.md) | Malicious files | — | Decompression-bomb/oversized fixtures; in-file injection fixtures; by-reference binary handling | Volume 9 |
| [RISK-SEC-010](../volume-09-security/03-threats-extensions-supply-chain.md) | Malicious plugins | — | Plugin isolation/teardown tests; permission-scoping tests; install-verification and ARP conformance | Volume 9 |
| [RISK-SEC-011](../volume-09-security/03-threats-extensions-supply-chain.md) | Malicious skills | — | Skill injection fixtures; required-tools gating tests; update-re-consent tests | Volume 9 |
| [RISK-SEC-012](../volume-09-security/03-threats-extensions-supply-chain.md) | Malicious repositories | — | Hostile-repo fixtures (no hook/script auto-exec); path/symlink containment; Git-mutation gating | Volume 9 |
| [RISK-SEC-013](../volume-09-security/03-threats-extensions-supply-chain.md) | Dependency attacks | — | Module-sum verification; CI dependency-audit gate (SM-16); SBOM diff; reproducible-build checks | Volume 9 |
| [RISK-SEC-014](../volume-09-security/03-threats-extensions-supply-chain.md) | CI compromise | — | Pipeline-permission audits; pinned-action verification; fork-PR isolation; provenance verification | Volume 9 |
| [RISK-SEC-015](../volume-09-security/03-threats-extensions-supply-chain.md) | Release compromise | — | Checksum verification; signature/provenance verification where enabled; yank-and-refuse tests | Volume 9 |
| [RISK-SEC-016](../volume-09-security/03-threats-extensions-supply-chain.md) | Update compromise | — | Verify-before-apply refusal tests; tampered-artifact fixtures; atomic-apply and offline-rollback tests | Volume 9 |
| [RISK-SEC-017](../volume-09-security/03-threats-extensions-supply-chain.md) | Compromised providers | — | Egress least-exposure tests; provider-change notification; fallback guard-rule tests (Volume 5) | Volume 9 |
| [RISK-SEC-018](../volume-09-security/03-threats-extensions-supply-chain.md) | Compromised local models | — | Local-provider conformance (SM-04); capability-honesty tests; output validation | Volume 9 |
| [RISK-SEC-019](../volume-09-security/04-threats-system.md) | Command injection | — | No-shell-interpolation fixtures; permission-matrix `execute` gating; sandbox env/limit/teardown tests | Volume 9 |
| [RISK-SEC-020](../volume-09-security/04-threats-system.md) | Path traversal | — | Traversal fixtures (rejection); canonicalization and per-path permission tests; working-dir confinement | Volume 9 |
| [RISK-SEC-021](../volume-09-security/04-threats-system.md) | Symlink attacks | — | Out-of-root symlink fixtures; TOCTOU race tests; resolved-target confinement tests | Volume 9 |
| [RISK-SEC-022](../volume-09-security/04-threats-system.md) | Secret exfiltration | — | Redaction-conformance; environment-filtering; egress-gating; secret-access audit tests | Volume 9 |
| [RISK-SEC-023](../volume-09-security/04-threats-system.md) | Credential theft | — | No-plaintext-at-rest tests; fallback encryption tests; zeroize-on-release; `credential_access` gating | Volume 9 |
| [RISK-SEC-024](../volume-09-security/04-threats-system.md) | Sandbox escape | — | SandboxPort-only launch audits; containment-level recording; policy enforcement; teardown process-tree tests | Volume 9 |
| [RISK-SEC-025](../volume-09-security/04-threats-system.md) | Privilege escalation | — | Scope-isolation tests; confused-deputy tests; grant-revocation/audit; approval-on-widening tests | Volume 9 |
| [RISK-SEC-026](../volume-09-security/04-threats-system.md) | Log leakage | — | Redaction-conformance over logs/errors/events/traces; secret-pattern scans; telemetry-content tests | Volume 9 |
| [RISK-SEC-027](../volume-09-security/04-threats-system.md) | Social engineering | — | Trusted-prompt effect/scope rendering tests; default-deny; approval-audit linkage; no-dangerous-default UI tests | Volume 9 |
| [FR-SEC-100](../volume-09-security/05-permission-model.md) | Permission model | Core | Permission-matrix unit tests; mediation probes (NFR-SEC-002); CLI/TUI parity fixtures; audit-chain resolution (SM-13 method) | Volume 9 |
| [FR-SEC-101](../volume-09-security/06-sandbox-specification.md) | Sandbox | MVP | Sandbox conformance suite: escape attempts, limit fault injection, teardown timing, orphan-sweep crash tests; containment honesty documentation audit | Volume 9 |
| [FR-SEC-102](../volume-09-security/07-credential-and-secret-management.md) | Secret storage | MVP | Per-platform backend integration tests; leak-hunt suite; INV-CRED-04 ordering crash tests; permission-denial probes; audit-chain resolution | Volume 9 |
| [FR-SEC-103](../volume-09-security/05-permission-model.md) | Evaluation precedence and inheritance | Core | Property-based candidate-permutation tests; golden decision tables; config-validation negatives; cross-platform determinism runs | Volume 9 |
| [FR-SEC-104](../volume-09-security/05-permission-model.md) | Grant persistence, expiry, and revocation | MVP | Persistence crash injection (decision→mint boundary); revocation races; ADR-029 backup/restore divergence tests; listing golden tests | Volume 9 |
| [FR-SEC-105](../volume-09-security/05-permission-model.md) | Non-interactive and policy-only enforcement | MVP | Mode-parity fixture suite; bypass-hunt over documented env vars/flags; prompt-free CI matrix (Volume 8); audit inspection | Volume 9 |
| [FR-SEC-106](../volume-09-security/06-sandbox-specification.md) | Sandbox tiers | MVP | Tier matrix integration tests (five subjects); workflow ceiling tests; plugin/MCP environment probes; record audits | Volume 9 |
| [FR-SEC-107](../volume-09-security/06-sandbox-specification.md) | Environment and secret filtering | MVP | Environment probes across tiers/platforms; planted-secret scrubs; single-construction-path static check; resolution fault injection | Volume 9 |
| [FR-SEC-108](../volume-09-security/06-sandbox-specification.md) | Filesystem policy, symlinks, temp directories, and cleanup | MVP | Path-policy fuzzing (traversal, links, case, depth) per platform; crash-injection sweep tests; teardown verification | Volume 9 |
| [FR-SEC-109](../volume-09-security/07-credential-and-secret-management.md) | Secret redaction at every sink | MVP | Leak-hunt suite with planted markers across all sinks; chunk-boundary property tests; schema field tests; redaction fault injection | Volume 9 |
| [FR-SEC-110](../volume-09-security/07-credential-and-secret-management.md) | Fallback store consent and lifecycle | MVP | Consent-flow integration tests; permission/corruption fixtures; keychain↔fallback migration round-trips; wording audit | Volume 9 |
| [FR-SEC-111](../volume-09-security/08-audit-and-incident-response.md) | Audit Log semantics | MVP | Chain property tests (append/verify/tamper); record-before-effect crash injection; retention/archive round-trips with offline re-verification; SM-13 suite | Volume 9 |
| [FR-SEC-112](../volume-09-security/08-audit-and-incident-response.md) | Incident response and disclosure hooks | MVP | Trigger simulation per trigger row; dedup tests; incident-surface integration tests; evidence-export verification | Volume 9 |
| [FR-SEC-113](../volume-09-security/09-approval-state-machine.md) | Approval lifecycle enforcement | MVP | Machine property tests; decision/expiry race injection; crash injection at persistence boundaries (SM-11 method); prompt fidelity goldens | Volume 9 |
| [NFR-SEC-001](../volume-09-security/08-audit-and-incident-response.md) | Release vulnerability posture (SM-16 a) | v1 | CodeQL + dependency audit + secret scanning gating releases; release audit attachment | Volume 9 |
| [NFR-SEC-002](../volume-09-security/05-permission-model.md) | Permission mediation coverage (SM-16 b) | MVP | Enforcement suite attempting unmediated side effects on every path; audit-chain resolution over instrumented runs | Volume 9 |
| [NFR-SEC-003](../volume-09-security/08-audit-and-incident-response.md) | Coordinated disclosure first response (SM-16 c) | v1 | Security-inbox tracking (Volume 15 process); quarterly and phase-gate review | Volume 9 |
| [NFR-SEC-004](../volume-09-security/06-sandbox-specification.md) | Secret leakage prevention across boundaries | MVP | Leak-hunt suite scanning all sinks for planted markers, all Tier 1 platforms, per merge and release | Volume 9 |
| [NFR-SEC-005](../volume-09-security/08-audit-and-incident-response.md) | Audit chain integrity and verification performance | MVP | Crash-injection chain continuity suite; 100k-record verification benchmark on reference hardware | Volume 9 |
| [NFR-SEC-006](../volume-09-security/05-permission-model.md) | Permission decision latency | MVP | Micro-benchmark at reference grant population (1,000 grants / 200 rules) per release | Volume 9 |

## Volume 10 — Configuration, Storage, and Observability

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-CFG-001](../volume-10-config-storage-observability/01-configuration-model.md) | Configuration precedence | MVP | Layer-matrix unit tests (every layer pair, both orders); determinism property tests; offline-suite zero-network assertion | Volume 10 |
| [FR-CFG-002](../volume-10-config-storage-observability/01-configuration-model.md) | Configuration documents, locations, and loading | MVP | Per-platform golden path tests (ADR-022); fault injection (unreadable/oversize/malformed); stray-file non-consultation test | Volume 10 |
| [FR-CFG-003](../volume-10-config-storage-observability/01-configuration-model.md) | Configuration Profiles as scope-bound layers | MVP | Selection-matrix unit tests (selector × scope × existence); persisted-profile integration tests; golden attribution fixtures | Volume 10 |
| [FR-CFG-004](../volume-10-config-storage-observability/01-configuration-model.md) | Environment variable mapping algorithm | MVP | Round-trip property tests over all schema keys; name fuzzing; ambiguity/duplicate matrices; Volume 8 environment-table integration | Volume 10 |
| [FR-CFG-005](../volume-10-config-storage-observability/01-configuration-model.md) | Invocation and runtime overrides | MVP | Protected-table matrix tests; TUI/IPC integration; watch-delta tests under concurrent edits; audit-chain event assertions | Volume 10 |
| [FR-CFG-006](../volume-10-config-storage-observability/01-configuration-model.md) | Include mechanism | Beta | Merge-order/cycle/depth/count/containment matrices incl. symlink escapes; include-graph fuzzing; attribution fixtures | Volume 10 |
| [FR-CFG-007](../volume-10-config-storage-observability/01-configuration-model.md) | Typed validation with complete findings | MVP | Seeded-defect corpus (100% recall per NFR-CFG-003); golden diagnostics; cross-key rule matrices; ADR-008 parser-differential tests | Volume 10 |
| [FR-CFG-008](../volume-10-config-storage-observability/01-configuration-model.md) | Configuration schema versioning, migration, and deprecation | MVP | Transform-chain tests (per-version and 1→current); rewrite crash/I-O fault injection; deprecation matrix (old/new/both) | Volume 10 |
| [FR-CFG-009](../volume-10-config-storage-observability/01-configuration-model.md) | Database backups, retention, and workspace locking | MVP | Crash injection at every step boundary; concurrent-process lock tests; retention audit-preservation property tests; disk-full injection | Volume 10 |
| [FR-CFG-010](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) | Configuration error reporting | MVP | Golden fixtures for every catalog entry (human + JSON); human/machine parity property test; redaction assertions | Volume 10 |
| [FR-CFG-011](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) | Secret detection and redaction in configuration surfaces | MVP | Seeded-secret corpus across all sinks (NFR-CFG-004 method); detector true/false-positive suites; write-path refusal tests | Volume 10 |
| [NFR-CFG-001](../volume-10-config-storage-observability/01-configuration-model.md) | Configuration resolution latency | MVP | Isolated benchmark harness, 50 iterations, p95, both reference machines; SM-06 budget decomposition | Volume 10 |
| [NFR-CFG-002](../volume-10-config-storage-observability/01-configuration-model.md) | Resolution determinism and snapshot completeness | MVP | Double-resolution hash property tests; SM-12 record-completeness validator over all suite runs | Volume 10 |
| [NFR-CFG-003](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) | Validation finding completeness | MVP | Seeded-defect corpus (≥ 50 configurations) with exact-match report assertions per merge | Volume 10 |
| [NFR-CFG-004](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) | Redaction effectiveness | MVP | Seeded-secret corpus; automated fragment scan of every sink output per merge and release | Volume 10 |
| [RISK-CFG-001](../volume-10-config-storage-observability/01-configuration-model.md) | Drift between schema, defaults, and published reference | — | Severity: Medium; Status: Open | Volume 10 |
| [RISK-CFG-002](../volume-10-config-storage-observability/01-configuration-model.md) | Users mispredict the effective value across ten layers | — | Severity: Medium; Status: Open | Volume 10 |
| [RISK-CFG-003](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) | Secrets committed to version control through workspace configuration | — | Severity: High; Status: Open | Volume 10 |
| [RISK-CFG-004](../volume-10-config-storage-observability/01-configuration-model.md) | Version-skew refusals in synced or downgraded environments | — | Severity: Medium; Status: Open | Volume 10 |
| [FR-OBS-001](../volume-10-config-storage-observability/04-events-and-envelope.md) | Event envelope | Core | Registry conformance over all suite-produced events; name-grammar lint in CI; SM-20 contract-diff of envelope and registry; negative-emission unit tests | Volume 10 |
| [FR-OBS-002](../volume-10-config-storage-observability/03-logging.md) | Structured logging pipeline | MVP | Log-conformance schema validation over suite output; correlation-join tests (SM-13 method); sink fault injection; Tier 1 matrix | Volume 10 |
| [FR-OBS-003](../volume-10-config-storage-observability/03-logging.md) | Log rotation and retention | MVP | Rotation unit tests; retention property tests (age × count × liveness); crash injection between rotation steps; disk-full fault injection | Volume 10 |
| [FR-OBS-004](../volume-10-config-storage-observability/03-logging.md) | Log redaction enforcement | MVP | Secret-leak canary suite (release gate); matcher property tests; handler fault injection; stderr capture audits | Volume 10 |
| [FR-OBS-005](../volume-10-config-storage-observability/04-events-and-envelope.md) | Event delivery and subscription semantics | Core | Burst/soak delivery suites with policy assertions; ordering property tests; replay boundary tests; shutdown drain tests; IPC bridge conformance | Volume 10 |
| [FR-OBS-006](../volume-10-config-storage-observability/04-events-and-envelope.md) | Event persistence, retention, and export | MVP | Crash injection around transactional appends (SM-11 method); retention property tests with audit-linked exclusions; export round-trips against payload schemas | Volume 10 |
| [FR-OBS-007](../volume-10-config-storage-observability/05-traces-metrics-costs.md) | Trace model and OpenTelemetry mapping | MVP | Trace-completeness validator (SM-12 component); tree-invariant property tests; crash injection with post-recovery audits; cross-boundary propagation tests; redaction leak tests | Volume 10 |
| [FR-OBS-008](../volume-10-config-storage-observability/05-traces-metrics-costs.md) | Metric registry and core catalog | MVP | Registry conformance over emitted samples; catalog presence test per release; cardinality scanners over snapshots; persistence fault injection | Volume 10 |
| [FR-OBS-009](../volume-10-config-storage-observability/05-traces-metrics-costs.md) | Cost observability: rollups, honesty, retention | MVP | Rollup property tests (idempotence, late arrivals, corrections, multi-currency); retention guards; split-basis honesty tests; divergence injection | Volume 10 |
| [FR-OBS-010](../volume-10-config-storage-observability/06-telemetry-and-consent.md) | Strict local/remote separation | MVP | SM-05 offline suite with network observation; no-exporter composition test; endpoint fault injection; policy-lock matrix | Volume 10 |
| [FR-OBS-011](../volume-10-config-storage-observability/06-telemetry-and-consent.md) | Telemetry consent lifecycle | MVP | Consent lifecycle integration tests (grant/revoke/version bump/endpoint change/lock/non-interactive refusal); backup-restore staleness tests; audit assertions | Volume 10 |
| [FR-OBS-012](../volume-10-config-storage-observability/06-telemetry-and-consent.md) | Collected-data catalog and prohibited data | MVP | Payload-enumeration conformance with prohibited-pattern scanners and canaries; catalog↔allowlist CI divergence check; extension-aggregation tests | Volume 10 |
| [FR-OBS-013](../volume-10-config-storage-observability/06-telemetry-and-consent.md) | Telemetry queue, export, and deletion | MVP | OTLP-double integration matrices (success/partial/outage/auth); queue-bound property tests; backoff timing with mock clocks; shutdown flush tests | Volume 10 |
| [NFR-OBS-001](../volume-10-config-storage-observability/03-logging.md) | Logging hot-path overhead | Beta | Handler-chain benchmark and suppressed-call microbenchmark per release on reference hardware | Volume 10 |
| [NFR-OBS-002](../volume-10-config-storage-observability/04-events-and-envelope.md) | Event publication overhead and delivery latency | Beta | Instrumented bus benchmark (subscriber counts 1/4/16, 500 events/s) per release; SM-07 regression cross-check | Volume 10 |
| [NFR-OBS-003](../volume-10-config-storage-observability/05-traces-metrics-costs.md) | Side-effect traceability | MVP | SM-13 automated audit-chain test per release (gating from MVP exit; 0 orphans) | Volume 10 |
| [NFR-OBS-004](../volume-10-config-storage-observability/05-traces-metrics-costs.md) | Run record completeness for replay | v1 | SM-12 record-completeness validator over all suite runs; replay divergence test per release | Volume 10 |
| [NFR-OBS-005](../volume-10-config-storage-observability/05-traces-metrics-costs.md) | Instrumentation overhead | Beta | Span/sample microbenchmarks plus paired instrumented/no-op E2E runs per release | Volume 10 |
| [NFR-OBS-006](../volume-10-config-storage-observability/06-telemetry-and-consent.md) | Zero-egress default posture | MVP | OS-level network observation across suites under default configuration; no-exporter composition test (gating from MVP exit) | Volume 10 |
| [RISK-OBS-001](../volume-10-config-storage-observability/03-logging.md) | Observability data volume degrades the host | — | Severity: Medium; Status: Open | Volume 10 |
| [RISK-OBS-002](../volume-10-config-storage-observability/04-events-and-envelope.md) | Correlation discontinuity across process boundaries | — | Severity: High; Status: Open | Volume 10 |
| [RISK-OBS-003](../volume-10-config-storage-observability/06-telemetry-and-consent.md) | Accidental telemetry egress | — | Severity: High; Status: Open | Volume 10 |

## Volume 11 — Git, GitHub, and Development Platforms

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-GIT-001](../volume-11-git-and-github/01-git-engine.md) | Git Engine | MVP | GitPort contract suite incl. cancellation; equivalence suite; permission-enforcement tests; Tier 1 matrix | Volume 11 |
| [FR-GIT-002](../volume-11-git-and-github/01-git-engine.md) | Repository discovery and version gating | MVP | Fixture-tree unit tests (worktrees, submodules, bare, symlinks); version-gate tests | Volume 11 |
| [FR-GIT-003](../volume-11-git-and-github/01-git-engine.md) | Read-query fidelity: status, diff, log, show, blame | MVP | Equivalence suite; streaming/cancellation contract tests; parser fuzzing | Volume 11 |
| [FR-GIT-004](../volume-11-git-and-github/01-git-engine.md) | Staging and commit creation | MVP | Hook fixtures; message byte-equivalence; attribution-refusal tests; permission tests | Volume 11 |
| [FR-GIT-005](../volume-11-git-and-github/01-git-engine.md) | Branches, tags, and worktrees | MVP | Dirty-tree/unmerged fixtures; safety-ref restore tests; permission tests | Volume 11 |
| [FR-GIT-006](../volume-11-git-and-github/01-git-engine.md) | Remote operations: fetch, pull, push | Beta | Local fixture remotes; lease-race tests; permission matrix; timeout injection | Volume 11 |
| [FR-GIT-007](../volume-11-git-and-github/01-git-engine.md) | History modification, conflicts, and recovery | Beta | Conflict-matrix tests; published-rewrite classification property tests; offline recovery round-trips | Volume 11 |
| [FR-GIT-008](../volume-11-git-and-github/01-git-engine.md) | Repository feature passthrough (hooks, ignore, signing, submodules, sparse checkout, LFS) | MVP | Feature fixtures; signed-commit verification; audit assertions | Volume 11 |
| [FR-GIT-009](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) | Hosting integration layer | Beta | Contract tests over official-API fixtures; permission enforcement; enterprise-endpoint tests | Volume 11 |
| [FR-GIT-010](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) | Change-request preparation flow | Beta | E2E over hosting doubles; duplicate-prevention; SDD stage integration | Volume 11 |
| [FR-GH-001](../volume-11-git-and-github/07-traceability-automation.md) | Traceability automation | MVP | Validator golden fixtures; fixture-repository chain E2E incl. release report; phase-gate sampling | Volume 11 |
| [FR-GH-002](../volume-11-git-and-github/03-repository-structure-and-branching.md) | Repository structure | Core | Structure-check CI job; CODEOWNERS/branch-protection audit; release qualification audit | Volume 11 |
| [FR-GH-003](../volume-11-git-and-github/03-repository-structure-and-branching.md) | Branching rules and branch hygiene | Core | Protection-configuration audit; validator tests; nightly audit assertions | Volume 11 |
| [FR-GH-004](../volume-11-git-and-github/04-pull-requests.md) | Pull request process | Core | Platform-configuration audit; validator tests; nightly audit; chain report | Volume 11 |
| [FR-GH-005](../volume-11-git-and-github/04-pull-requests.md) | AI provenance labeling and commit-message enforcement | Core | Golden message corpus (hook/CI parity); label-consistency tests; configuration audit | Volume 11 |
| [FR-GH-006](../volume-11-git-and-github/05-issues-projects-roadmap.md) | Issue taxonomy and intake forms | Core | Form-schema validation; submission walkthrough at phase gates; audit assertions | Volume 11 |
| [FR-GH-007](../volume-11-git-and-github/05-issues-projects-roadmap.md) | Label taxonomy as synchronized data | Core | Sync-tool tests against fixture state; workflow permission audit | Volume 11 |
| [FR-GH-008](../volume-11-git-and-github/05-issues-projects-roadmap.md) | Projects, milestones, and roadmap operation | MVP | Automation integration tests on fixture project; nightly audit; phase-gate review | Volume 11 |
| [FR-GH-009](../volume-11-git-and-github/06-github-actions.md) | Quality pipelines and required checks | Core | Workflow integration tests; protection audit; NFR-GH-002 measurement; policy self-test | Volume 11 |
| [FR-GH-010](../volume-11-git-and-github/06-github-actions.md) | Security scanning pipelines | MVP | Fixture-based scan tests; release-gate integration test; weekly report audit | Volume 11 |
| [FR-GH-011](../volume-11-git-and-github/06-github-actions.md) | Release, upgrade, and documentation pipelines | MVP | Snapshot dry runs; rc rehearsals; artifact verification commands; environment audit | Volume 11 |
| [FR-GH-012](../volume-11-git-and-github/06-github-actions.md) | Workflow security posture enforcement | Core | Rule-set unit tests over fixture workflows; self-application; phase-gate ADR-149 comparison | Volume 11 |
| [NFR-GIT-001](../volume-11-git-and-github/01-git-engine.md) | Git output fidelity | MVP | Equivalence suite, 0 divergences, per merge and per release across the git version matrix | Volume 11 |
| [NFR-GIT-002](../volume-11-git-and-github/01-git-engine.md) | Destructive-operation recoverability | Beta | Instrumented destructive-operation campaign with offline restoration, 100% | Volume 11 |
| [NFR-GH-001](../volume-11-git-and-github/06-github-actions.md) | Development traceability completeness | MVP | Nightly chain audit; release chain reports; phase-gate trend review | Volume 11 |
| [NFR-GH-002](../volume-11-git-and-github/06-github-actions.md) | Pull-request feedback latency | MVP | Check-run timestamp rollups, p85 targets, monthly and phase-gate review | Volume 11 |
| [RISK-GIT-001](../volume-11-git-and-github/01-git-engine.md) | Git version and output-format drift | — | Equivalence-suite failures; E-GIT-005 telemetry | Volume 11 |
| [RISK-GIT-002](../volume-11-git-and-github/01-git-engine.md) | User data loss through destructive operations | — | Class D audit records; NFR-GIT-002 campaign | Volume 11 |
| [RISK-GIT-003](../volume-11-git-and-github/01-git-engine.md) | Repository hooks as arbitrary code execution | — | Hook-execution audit trail; Volume 9 monitoring | Volume 11 |
| [RISK-GIT-004](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) | Hosting API drift, deprecation, or policy change | — | Contract-fixture failures; error telemetry trends | Volume 11 |
| [RISK-GH-001](../volume-11-git-and-github/06-github-actions.md) | CI or release workflow compromise | — | Policy-check findings; provenance verification; platform audit log review | Volume 11 |
| [RISK-GH-002](../volume-11-git-and-github/07-traceability-automation.md) | Traceability erosion through process bypass or decay | — | Nightly audit orphan/override reports; NFR-GH-001 trend | Volume 11 |
| [RISK-GH-003](../volume-11-git-and-github/06-github-actions.md) | Fork pull-request secret or privilege exposure | — | Policy check on workflow changes; nightly fork-trigger assertion | Volume 11 |

## Volume 12 — Performance and Reliability

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-PERF-001](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Deadline and timeout baseline | MVP | Fault-injection suite; static audit of external call sites | Volume 12 |
| [FR-PERF-002](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Backpressure and overload shedding | MVP | Saturation fixtures; flooding tests; load scenario in benchmark suite | Volume 12 |
| [FR-PERF-003](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Degraded operation modes | MVP | Per-mode fault-injection conformance suite; offline suite; event-pairing audit | Volume 12 |
| [FR-PERF-004](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Resource watchdog and recovery objectives | MVP | Disk-fill/memory-pressure fixtures; restart-time measurement; watchdog-failure fixture | Volume 12 |
| [FR-PERF-005](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) | Benchmark suite and regression gating | MVP | Gate-evaluator self-test; CI configuration review; release artifact audit | Volume 12 |
| [FR-PERF-006](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) | Operational limits enforcement | MVP | Per-limit conformance fixtures; hot-path overhead benchmarks | Volume 12 |
| [NFR-PERF-001](../volume-12-performance-and-reliability/01-performance-targets.md) | CLI cold start | v1 | Benchmark suite (operation layer), release gate | Volume 12 |
| [NFR-PERF-002](../volume-12-performance-and-reliability/01-performance-targets.md) | CLI warm start | v1 | Benchmark suite (operation layer), release gate | Volume 12 |
| [NFR-PERF-003](../volume-12-performance-and-reliability/01-performance-targets.md) | TUI startup to interactive | v1 | Benchmark suite (operation layer), release gate | Volume 12 |
| [NFR-PERF-004](../volume-12-performance-and-reliability/01-performance-targets.md) | TUI frame render time | v1 | Instrumented render benchmark, release gate | Volume 12 |
| [NFR-PERF-005](../volume-12-performance-and-reliability/01-performance-targets.md) | TUI input latency | v1 | Scripted interaction replay, release gate | Volume 12 |
| [NFR-PERF-006](../volume-12-performance-and-reliability/01-performance-targets.md) | First-token overhead | v1 | Instrumented turn benchmark with mock provider, release gate | Volume 12 |
| [NFR-PERF-007](../volume-12-performance-and-reliability/01-performance-targets.md) | Streaming update overhead | v1 | Mock streaming provider benchmark, release gate | Volume 12 |
| [NFR-PERF-008](../volume-12-performance-and-reliability/01-performance-targets.md) | Tool dispatch overhead | v1 | Tool Runtime micro-benchmark, release gate | Volume 12 |
| [NFR-PERF-009](../volume-12-performance-and-reliability/01-performance-targets.md) | Filesystem scan | v1 | Workspace scan benchmark on DS-M/DS-L, release gate | Volume 12 |
| [NFR-PERF-010](../volume-12-performance-and-reliability/01-performance-targets.md) | Git status latency | v1 | Paired git benchmark, release gate | Volume 12 |
| [NFR-PERF-011](../volume-12-performance-and-reliability/01-performance-targets.md) | Indexing throughput | v1 | IndexerPort benchmarks, release gate | Volume 12 |
| [NFR-PERF-012](../volume-12-performance-and-reliability/01-performance-targets.md) | Memory retrieval latency | v1 | Memory Manager benchmark over DS-MEM, release gate | Volume 12 |
| [NFR-PERF-013](../volume-12-performance-and-reliability/01-performance-targets.md) | Search latency | v1 | Index query benchmarks, release gate | Volume 12 |
| [NFR-PERF-014](../volume-12-performance-and-reliability/01-performance-targets.md) | Patch generation | v1 | Patch benchmark with apply-clean assertion, release gate | Volume 12 |
| [NFR-PERF-015](../volume-12-performance-and-reliability/01-performance-targets.md) | Diff rendering | v1 | Instrumented TUI diff benchmark, release gate | Volume 12 |
| [NFR-PERF-016](../volume-12-performance-and-reliability/01-performance-targets.md) | Session restore | v1 | Crash-injection restore benchmark, release gate | Volume 12 |
| [NFR-PERF-017](../volume-12-performance-and-reliability/01-performance-targets.md) | Memory budget | v1 | Process accounting (idle and soak), release gate | Volume 12 |
| [NFR-PERF-018](../volume-12-performance-and-reliability/01-performance-targets.md) | CPU budget | v1 | Process accounting (idle and streaming), release gate | Volume 12 |
| [NFR-PERF-019](../volume-12-performance-and-reliability/01-performance-targets.md) | Disk budget | v1 | Filesystem accounting after install/soak/build/shutdown, release gate | Volume 12 |
| [NFR-PERF-020](../volume-12-performance-and-reliability/01-performance-targets.md) | Concurrency capacity and scheduler overhead | v1 | Load scenario plus scheduler micro-benchmark, release gate | Volume 12 |
| [NFR-PERF-021](../volume-12-performance-and-reliability/01-performance-targets.md) | Large-repository operation | v1 | DS-L suite re-run, release gate | Volume 12 |
| [NFR-PERF-022](../volume-12-performance-and-reliability/01-performance-targets.md) | Large-file handling | v1 | DS-F benchmarks with process accounting, release gate | Volume 12 |
| [NFR-PERF-023](../volume-12-performance-and-reliability/01-performance-targets.md) | Long-session stability | v1 | DS-SOAK replay with hourly checkpoints and leak gates | Volume 12 |
| [NFR-PERF-024](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Offline operation | MVP | Offline test suite (Volume 13), MVP exit gate | Volume 12 |
| [NFR-PERF-025](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Tool-call reliability | v1 | Suite outcome classification plus fault injection | Volume 12 |
| [NFR-PERF-026](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Session recovery success | v1 | Crash-injection suite, ≥ 200 injections per release | Volume 12 |
| [NFR-PERF-027](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Crash-free operation | v1 | Soak/E2E crash accounting; panic-injection fixtures | Volume 12 |
| [NFR-PERF-028](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Degradation responsiveness | Beta | Instrumented per-mode fault injection | Volume 12 |
| [RISK-PERF-001](../volume-12-performance-and-reliability/01-performance-targets.md) | Performance targets unattainable on reference hardware | — | Risk register review at phase gates; nightly trend monitoring | Volume 12 |
| [RISK-PERF-002](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) | Benchmark environment variance masks or fakes regressions | — | Risk register review at phase gates; calibration failure monitoring | Volume 12 |
| [RISK-PERF-003](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) | Degradation matrix complexity outgrows testing | — | Risk register review at phase gates; event-pairing audits | Volume 12 |
| [RISK-PERF-004](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) | Real workspaces exceed declared limits | — | Risk register review at phase gates; enforcement-counter diagnostics | Volume 12 |

## Volume 13 — Testing and Quality

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-TEST-001](../volume-13-testing-and-quality/01-strategy-and-pyramid.md) | Test pyramid and suite organization | MVP | Distribution report at T4; CI duration metrics; review checklist | Volume 13 |
| [FR-TEST-002](../volume-13-testing-and-quality/01-strategy-and-pyramid.md) | Test-to-requirement traceability annotations | MVP | Traceability scanner in CI; T4 gate audit | Volume 13 |
| [FR-TEST-003](../volume-13-testing-and-quality/02-test-types-catalog.md) | Closed test-type catalog and classification | MVP | Classification check in T0; T4 audit against executed CI jobs | Volume 13 |
| [FR-TEST-004](../volume-13-testing-and-quality/02-test-types-catalog.md) | Port contract test kits | Core | Kit inventory check; kits in T0; consolidation audit | Volume 13 |
| [FR-TEST-005](../volume-13-testing-and-quality/02-test-types-catalog.md) | Offline test suite and network sentinel | MVP | Offline suite at T1/T3; harness self-tests; T4 report audit | Volume 13 |
| [FR-TEST-006](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Test doubles and test providers | MVP | Contract kits over fakes; ADR-033 import-graph check; live-lane cross-checks | Volume 13 |
| [FR-TEST-007](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Determinism controls and flaky-test quarantine | MVP | Nightly determinism lane; quarantine repository check; T4 accounting | Volume 13 |
| [FR-TEST-008](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Test data and secret handling | MVP | Secret-scanning checks; generator determinism tests; sanitizer tests | Volume 13 |
| [FR-TEST-009](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) | Release qualification pipeline and evidence bundle | MVP | Pipeline self-tests; bundle schema validation; T4 release audit | Volume 13 |
| [NFR-TEST-001](../volume-13-testing-and-quality/01-strategy-and-pyramid.md) | Merge-gate wall-clock budget | MVP | Weekly CI duration report; T4 evaluation | Volume 13 |
| [NFR-TEST-002](../volume-13-testing-and-quality/01-strategy-and-pyramid.md) | Suite determinism and order independence | MVP | Nightly determinism lane (50 repeats, shuffled) | Volume 13 |
| [NFR-TEST-003](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Coverage thresholds | MVP | Coverage merge gate; per-release report (SM-14) | Volume 13 |
| [NFR-TEST-004](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Mutation score on scoped packages | Beta | ADR-175 mutation lane report at phase gates | Volume 13 |
| [NFR-TEST-005](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Flake rate and quarantine dwell time | MVP | CI triage records and quarantine registry, weekly | Volume 13 |
| [NFR-TEST-006](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) | Qualification completeness | MVP | Post-publication audit job; phase-gate audit | Volume 13 |
| [RISK-TEST-001](../volume-13-testing-and-quality/01-strategy-and-pyramid.md) | Pyramid inversion | — | Risk register review at phase gates | Volume 13 |
| [RISK-TEST-002](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Flaky tests eroding gate credibility | — | Risk register review at phase gates | Volume 13 |
| [RISK-TEST-003](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Coverage gaming | — | Risk register review at phase gates | Volume 13 |
| [RISK-TEST-004](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Test-double divergence from real systems | — | Risk register review at phase gates | Volume 13 |
| [RISK-TEST-005](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) | Secret or sensitive-data leakage through test assets | — | Risk register review at phase gates | Volume 13 |
| [RISK-TEST-006](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) | Gate erosion under release pressure | — | Risk register review at phase gates | Volume 13 |

## Volume 14 — Distribution, Installation, and Updates

| ID | Title | Phase | Verification method | Owning volume |
|---|---|---|---|---|
| [FR-REL-001](../volume-14-distribution/01-distribution-channels.md) | Release pipeline and distribution channels | MVP | Release-audit CI job (inventory, grammar, channel mapping); Volume 13 release qualification | Volume 14 |
| [FR-REL-002](../volume-14-distribution/01-distribution-channels.md) | Integrity metadata: checksums, signatures, SBOM, provenance | MVP | Tamper-fixture suite; release-audit re-verification; SM-18 harness inclusion | Volume 14 |
| [FR-REL-003](../volume-14-distribution/01-distribution-channels.md) | Installation channels: Homebrew tap, shell installer, Linux packages | MVP | Per-release installation matrix on Tier 1 platforms; installer tamper fixtures | Volume 14 |
| [FR-REL-004](../volume-14-distribution/01-distribution-channels.md) | Air-gapped installation and offline update sources | MVP | Offline suite with mirror fixtures under OS-level network disablement | Volume 14 |
| [FR-REL-005](../volume-14-distribution/02-updater-and-rollback.md) | Update check, channel subscription, and notification | MVP | UpdaterPort contract tests with metadata fixtures; offline suite; egress capture | Volume 14 |
| [FR-REL-006](../volume-14-distribution/02-updater-and-rollback.md) | Download, verification, and consent-gated apply | MVP | SM-18 update suite; fault injection per state; lock contention tests | Volume 14 |
| [FR-REL-007](../volume-14-distribution/02-updater-and-rollback.md) | Update automation policy | Beta | Policy-matrix integration tests; idle-window scheduling tests; suspension fixtures | Volume 14 |
| [FR-REL-008](../volume-14-distribution/02-updater-and-rollback.md) | Rollback of the installed version | MVP | SM-19 offline rollback harness; corruption and schema-boundary fixtures | Volume 14 |
| [FR-REL-009](../volume-14-distribution/03-install-uninstall-data.md) | Installation layout and ownership detection | MVP | Per-owner fixtures; lazy-initialization filesystem snapshots | Volume 14 |
| [FR-REL-010](../volume-14-distribution/03-install-uninstall-data.md) | Uninstallation with data preservation by default | MVP | Uninstall matrix with before/after snapshots; broken-install fixtures | Volume 14 |
| [FR-REL-011](../volume-14-distribution/03-install-uninstall-data.md) | Explicit data removal | MVP | Purge fixtures with filesystem/credential snapshots; override-layout fixtures | Volume 14 |
| [FR-REL-012](../volume-14-distribution/04-versioning-support-backports.md) | Semantic versioning of the product and public contracts | Core | Release-audit bump-class vs contract-diff consistency; upgrade-test matrix | Volume 14 |
| [FR-REL-013](../volume-14-distribution/04-versioning-support-backports.md) | Deprecation policy | Beta | Deprecation-ledger reconciliation in release audit; warning emission tests | Volume 14 |
| [FR-REL-014](../volume-14-distribution/04-versioning-support-backports.md) | Support windows, release branches, and backports | MVP | Upgrade-path matrix; branch-content audits; support-status computation tests | Volume 14 |
| [FR-REL-015](../volume-14-distribution/04-versioning-support-backports.md) | Changelog and release notes | MVP | Release-audit note checks; changelog divergence test | Volume 14 |
| [FR-REL-016](../volume-14-distribution/05-state-machines.md) | Machine conformance and update history | MVP | State-machine property suite; crash-injection per state; event/history reconciliation | Volume 14 |
| [NFR-REL-001](../volume-14-distribution/02-updater-and-rollback.md) | Update time (SM-18) | v1 | Automated N−1 → N update test per release, per-state instrumentation, p95 | Volume 14 |
| [NFR-REL-002](../volume-14-distribution/02-updater-and-rollback.md) | Rollback time (SM-19) | v1 | Automated offline rollback test per release with egress capture, p95 | Volume 14 |
| [NFR-REL-003](../volume-14-distribution/04-versioning-support-backports.md) | Public-contract stability (SM-20) | v1 | Contract-diff tooling per release; audit against the deprecation ledger | Volume 14 |
| [RISK-REL-001](../volume-14-distribution/02-updater-and-rollback.md) | Failed or interrupted update leaves an unusable installation | — | Severity: High; Status: Open | Volume 14 |
| [RISK-REL-002](../volume-14-distribution/01-distribution-channels.md) | External signing and notarization dependencies | — | Severity: Medium; Status: Open — V14-OQ-1/V14-OQ-3 pending | Volume 14 |
| [RISK-REL-003](../volume-14-distribution/02-updater-and-rollback.md) | Binary–database version skew after rollback or workspace sync | — | Severity: Medium; Status: Open | Volume 14 |
| [RISK-REL-004](../volume-14-distribution/04-versioning-support-backports.md) | Support and compatibility obligations exceeding maintainer capacity | — | Severity: Medium; Status: Open | Volume 14 |
