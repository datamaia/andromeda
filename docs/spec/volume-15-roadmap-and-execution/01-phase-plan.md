# 01 — Phase Plan

This chapter consolidates the delivery plan across the whole corpus. The phase
*definitions* — Core, MVP, Beta, v1, v2, Future, Out of Scope — are owned by Volume 1
[chapter 05](../volume-01-vision-and-product/05-scope-and-phases.md) (single-home matrix,
Volume 0 chapter 03); this chapter does not redefine them. What this chapter adds:

1. The **capability set per phase**, aggregated from the Requirements index of every
   volume register (Volumes 1–14), grouped by area and cited by requirement ID.
2. The **MVP conformance check**: the aggregated MVP set compared against the
   twenty-seven-item MVP minimum of Volume 1 chapter 05, with every drift finding recorded
   as an open question in this volume's register — never silently corrected.
3. **Entry criteria, exit criteria, and quality gates** per phase, operationalizing the
   Volume 1 definitions with the concrete gates minted by Volumes 11, 13, and 14.

Method: the phase of a requirement is authoritative in its minting volume's register. This
chapter is a projection of those registers; if a register and this chapter ever disagree,
the register wins and this chapter is corrected through the change procedure (Volume 0
chapter 10). Risks (`RISK-*`) carry no phase and are excluded from the capability tables;
they drive ordering in [chapter 03](03-backlog-and-prioritization.md).

## Core

Core is the architectural nucleus: contracts and primitives that everything else consumes.
Core is not a release; every Core item ships inside the MVP but is built and stabilized
first (Volume 1 chapter 05).

| Area | Capability set | Requirement IDs |
|---|---|---|
| Product (Vol 1) | The thirteen product objectives that anchor traceability | PRD-001, PRD-002, PRD-003, PRD-004, PRD-005, PRD-006, PRD-007, PRD-008, PRD-009, PRD-010, PRD-011, PRD-012, PRD-013 |
| Architecture (Vol 3) | Layered dependency rule, ports-and-adapters composition, frozen port names, context propagation and cancellation, bounded process model, supervised concurrency, versioned public contracts; CI dependency enforcement; leak-free termination | FR-ARCH-001, FR-ARCH-002, FR-ARCH-003, FR-ARCH-004, FR-ARCH-005, FR-ARCH-006, FR-ARCH-011, NFR-ARCH-001, NFR-ARCH-004 |
| Platform abstraction (Vol 3) | PAL encapsulation, surface completeness, XDG directory resolution; conformance coverage; platform-conditional code containment | FR-PORT-001, FR-PORT-002, FR-PORT-003, NFR-PORT-002, NFR-PORT-004 |
| Agent runtime (Vol 4) | Canonical state-machine enforcement over the frozen Volume 2 vocabularies | FR-AGT-015 |
| Providers (Vol 5) | Provider contract (keystone), adapter declaration and registration, capability declaration and matrix, error normalization; official-mechanisms-only authentication gate (keystone) | FR-PROV-001, FR-PROV-002, FR-PROV-010, FR-PROV-050, FR-AUTH-001 |
| Tools (Vol 6) | Tool contract (keystone), declaration and payload schema validation, naming/namespaces/resolution | FR-TOOL-001, FR-TOOL-002, FR-TOOL-003 |
| CLI (Vol 8) | CLI grammar (keystone), runtime mediation and driver parity, global flags and invocation modes, structured JSON output for every command, stream discipline | FR-CLI-001, FR-CLI-002, FR-CLI-005, FR-CLI-006, FR-CLI-007 |
| Security (Vol 9) | Permission model (keystone), evaluation precedence and inheritance | FR-SEC-100, FR-SEC-103 |
| Observability (Vol 10) | Event envelope (keystone), event delivery and subscription semantics | FR-OBS-001, FR-OBS-005 |
| Development platform (Vol 11) | Repository structure, branching rules, PR process, AI-provenance and commit-message enforcement, issue taxonomy, label taxonomy, quality pipelines and required checks, workflow security posture | FR-GH-002, FR-GH-003, FR-GH-004, FR-GH-005, FR-GH-006, FR-GH-007, FR-GH-009, FR-GH-012 |
| Testing (Vol 13) | Port contract test kits | FR-TEST-004 |
| Distribution (Vol 14) | Semantic versioning of the product and public contracts | FR-REL-012 |

**Core entry criteria** (per item): the owning specification volume is authored and
lint-clean; the contract surfaces the item exposes are defined (Volume 1 chapter 05). At
Phase C consolidation this holds for Volumes 0–14.

**Core exit criteria** (per item): implemented, covered by unit and contract tests, and
consumed by at least one dependent component. Additionally, corpus-wide:

- The repository skeleton, branch protection, and required checks of FR-GH-002/003/004/009
  are live before any product code merges — process gates precede the code they gate.
- The ADR-033 dependency checks (depguard, import-graph test, prohibited-construct
  scanner) pass on every merge from the first mainline commit (NFR-ARCH-001).
- Every frozen port (Volume 3 chapter 02) has its FR-TEST-004 contract kit before the
  first adapter of that port merges.

**Core quality gates:** the Volume 13 T0 (merge) tier — gofmt and golangci-lint (ADR-018),
`scripts/spec_lint.py` with zero errors on `docs/spec/` paths, dependency-rule checks
(ADR-033), unit and contract suites, the traceability validators of FR-GH-001 (activated
progressively as chains form), and the commit-message/provenance checks of FR-GH-005.
Coverage is measured from the first merge; the enforced floor binds at MVP per
NFR-TEST-003. A Core contract may change freely until MVP exit; from MVP exit onward,
changes follow the deprecation rules bound to SM-20 (Volume 1 chapter 05 discipline).

## MVP

The MVP is the first public, usable, installable release, functional for the end-to-end
jobs UC-01, UC-02, UC-03, UC-09, and UC-11 (Volume 1 chapters 03/05). The aggregated MVP
capability set, by area:

| Area | Capability set | Requirement IDs |
|---|---|---|
| Architecture (Vol 3) | External IPC control surface; crash recovery and resumable state; graceful shutdown ordering and deadline; platform matrix conformance, Tier 1 parity, single-binary deliverable | FR-ARCH-007, FR-ARCH-009, FR-ARCH-010, FR-PORT-004, NFR-ARCH-003, NFR-PORT-001, NFR-PORT-003 |
| Agent runtime (Vol 4) | Agent loop (keystone); turn handling; interruption/pause/resume; run budgets; workspace lifecycle; plan production, revision, and approval interplay; task scheduling, retry, cancellation; versioned prompt templates and deterministic rendering; transition legality, resume fidelity, render determinism | FR-AGT-001, FR-AGT-002, FR-AGT-003, FR-AGT-005, FR-AGT-006, FR-AGT-007, FR-AGT-008, FR-AGT-009, FR-AGT-010, FR-AGT-011, FR-AGT-012, FR-AGT-013, FR-AGT-014, NFR-AGT-001, NFR-AGT-002, NFR-AGT-003 |
| Providers (Vol 5) | Capability negotiation and degradation; model discovery and change notification; streaming, tool-calling normalization, structured outputs; token and cost accounting; timeouts/rate limits/retries, circuit breaker, routing, guarded fallback; adapter catalog with the MVP seed (generic OpenAI-compatible, Anthropic, Ollama); local and offline provider operation; accounting completeness and normalization coverage | FR-PROV-011, FR-PROV-012, FR-PROV-013, FR-PROV-020, FR-PROV-021, FR-PROV-022, FR-PROV-030, FR-PROV-031, FR-PROV-040, FR-PROV-041, FR-PROV-042, FR-PROV-043, FR-PROV-080, FR-PROV-081, FR-PROV-082, FR-PROV-083, FR-PROV-084, FR-PROV-085, NFR-PROV-003, NFR-PROV-004 |
| Authentication (Vol 5) | API keys; multiple profiles; Secret Store-backed storage and resolution; rotation and revocation; no plaintext at rest; redaction in every sink | FR-AUTH-002, FR-AUTH-008, FR-AUTH-009, FR-AUTH-011, NFR-AUTH-001, NFR-AUTH-002 |
| Tools (Vol 6) | Registration/availability/enablement; permission mediation of every invocation; limits, sandbox placement, teardown; built-in catalog MVP subset (filesystem, terminal, git); Tool Invocation machine conformance; contract conformance and record completeness | FR-TOOL-004, FR-TOOL-005, FR-TOOL-006, FR-TOOL-007, FR-TOOL-008, NFR-TOOL-001, NFR-TOOL-002 |
| Memory (Vol 7) | Memory model (keystone); provenance and trust; secret exclusion; ingestion; dedup/conflict/supersession; retention; deletion cascade; export; offline operation; privacy/locality and deletion completeness | FR-MEM-001, FR-MEM-002, FR-MEM-003, FR-MEM-004, FR-MEM-005, FR-MEM-007, FR-MEM-008, FR-MEM-009, FR-MEM-010, NFR-MEM-001, NFR-MEM-002 |
| Context (Vol 7) | Context assembly (keystone); token budgeting; dedup/compression; freshness/trust/conflicts; pinning and exclusion; tool-result and binary handling; snapshots and reproducibility; window compliance | FR-CTX-001, FR-CTX-002, FR-CTX-003, FR-CTX-004, FR-CTX-005, FR-CTX-006, FR-CTX-007, NFR-CTX-001 |
| Indexing (Vol 7) | Indexing engine (keystone); chunking; embeddings; incremental updates; operation without embeddings and without Internet; machine conformance and recovery; offline guarantee and generation integrity | FR-IDX-001, FR-IDX-002, FR-IDX-003, FR-IDX-004, FR-IDX-005, FR-IDX-006, NFR-IDX-001, NFR-IDX-002 |
| CLI (Vol 8) | Root command and TUI hand-off; verbosity modes; non-interactive and CI modes; confirmations; environment variables; shell completion; core/platform/data/maintenance command families; error presentation, capability adaptation, progress reporting; help completeness; prompt-free non-interactive operation | FR-CLI-003, FR-CLI-008, FR-CLI-009, FR-CLI-010, FR-CLI-011, FR-CLI-012, FR-CLI-013, FR-CLI-014, FR-CLI-015, FR-CLI-016, FR-UX-001, FR-UX-002, FR-UX-003, NFR-CLI-002, NFR-CLI-003 |
| TUI (Vol 8) | TUI shell (keystone); panels, navigation, keyboard/mouse, resize; streaming render pipeline; theming and design tokens with degradation tiers and light fallback; splash; core and platform screens; command palette; help overlay; error center; no-color and glyph fallback; SSH/multiplexer operation; interaction patterns (streaming, progress, modals, toasts, view states, copy/paste, data navigation); small-terminal completeness, determinism, contrast, reachability, feedback deadline, memory ceiling | FR-TUI-001, FR-TUI-002, FR-TUI-003, FR-TUI-004, FR-TUI-005, FR-TUI-006, FR-TUI-007, FR-TUI-008, FR-TUI-009, FR-UX-040, FR-UX-041, FR-UX-042, FR-UX-043, FR-TUI-060, FR-TUI-061, FR-TUI-063, FR-TUI-064, FR-TUI-066, FR-TUI-067, FR-TUI-068, FR-UX-070, FR-UX-071, FR-UX-072, FR-UX-073, FR-UX-074, FR-UX-075, FR-UX-076, NFR-TUI-001, NFR-TUI-002, NFR-UX-040, NFR-TUI-069, NFR-UX-077, NFR-UX-078 |
| Security (Vol 9) | Sandbox (keystone) with tiers, environment/secret filtering, filesystem policy; secret storage (keystone) with fallback-store consent; grant persistence/expiry/revocation; non-interactive policy-only enforcement; redaction at every sink; Audit Log; incident-response hooks; approval lifecycle; mediation coverage, leak prevention, chain integrity, decision latency | FR-SEC-101, FR-SEC-102, FR-SEC-104, FR-SEC-105, FR-SEC-106, FR-SEC-107, FR-SEC-108, FR-SEC-109, FR-SEC-110, FR-SEC-111, FR-SEC-112, FR-SEC-113, NFR-SEC-002, NFR-SEC-004, NFR-SEC-005, NFR-SEC-006 |
| Configuration (Vol 10) | Precedence (keystone); documents/locations/loading; profiles; env-var mapping; overrides; typed validation; schema versioning and migration; database backups/retention/locking; error reporting; secret detection and redaction; resolution latency/determinism, validation completeness, redaction effectiveness | FR-CFG-001, FR-CFG-002, FR-CFG-003, FR-CFG-004, FR-CFG-005, FR-CFG-007, FR-CFG-008, FR-CFG-009, FR-CFG-010, FR-CFG-011, NFR-CFG-001, NFR-CFG-002, NFR-CFG-003, NFR-CFG-004 |
| Observability (Vol 10) | Structured logging with rotation, retention, redaction; event persistence and export; traces and OTel mapping; metric registry; cost rollups; strict local/remote separation; telemetry consent lifecycle, catalog, queue and deletion; side-effect traceability; zero-egress default | FR-OBS-002, FR-OBS-003, FR-OBS-004, FR-OBS-006, FR-OBS-007, FR-OBS-008, FR-OBS-009, FR-OBS-010, FR-OBS-011, FR-OBS-012, FR-OBS-013, NFR-OBS-003, NFR-OBS-006 |
| Git (Vol 11) | Git Engine (keystone): discovery and version gating; status/diff/log/show/blame fidelity; staging and commits; branches/tags/worktrees; feature passthrough (hooks, ignore, signing, submodules, sparse, LFS); output fidelity | FR-GIT-001, FR-GIT-002, FR-GIT-003, FR-GIT-004, FR-GIT-005, FR-GIT-008, NFR-GIT-001 |
| Development platform (Vol 11) | Traceability automation (keystone); Projects/milestones/roadmap operation; security scanning pipelines; release/upgrade/docs pipelines; traceability completeness and PR feedback latency | FR-GH-001, FR-GH-008, FR-GH-010, FR-GH-011, NFR-GH-001, NFR-GH-002 |
| Performance and reliability (Vol 12) | Deadline/timeout baseline; backpressure and shedding; degraded modes; resource watchdog; benchmark suite and regression gating; operational limits; offline operation | FR-PERF-001, FR-PERF-002, FR-PERF-003, FR-PERF-004, FR-PERF-005, FR-PERF-006, NFR-PERF-024 |
| Testing (Vol 13) | Pyramid and suite organization; traceability annotations; type catalog; offline suite and network sentinel; doubles and test providers; determinism and quarantine; test data and secret handling; release qualification pipeline; merge-gate budget, determinism, coverage thresholds, flake rate, qualification completeness | FR-TEST-001, FR-TEST-002, FR-TEST-003, FR-TEST-005, FR-TEST-006, FR-TEST-007, FR-TEST-008, FR-TEST-009, NFR-TEST-001, NFR-TEST-002, NFR-TEST-003, NFR-TEST-005, NFR-TEST-006 |
| Distribution (Vol 14) | Release pipeline (keystone); integrity metadata (checksums, signatures, SBOM, provenance); Homebrew tap, shell installer, Linux packages; air-gapped installation; update check/download/verify/apply; rollback; layout and ownership detection; uninstall with data preservation; explicit purge; support windows and branches; changelog; machine conformance and history | FR-REL-001, FR-REL-002, FR-REL-003, FR-REL-004, FR-REL-005, FR-REL-006, FR-REL-008, FR-REL-009, FR-REL-010, FR-REL-011, FR-REL-014, FR-REL-015, FR-REL-016 |

### MVP conformance check against the Volume 1 minimum

Volume 1 chapter 05 fixes a twenty-seven-item MVP minimum ("the MVP includes, at
minimum"). The table maps each item to the requirement IDs that realize it, all of which
are phased Core or MVP in their owning registers. **Direction 1 result: no drift** — no
minimum item depends on a requirement phased later than MVP.

| # | MVP minimum item (Vol 1 ch 05) | Realizing requirement IDs (phase Core or MVP) |
|---|---|---|
| 1 | Functional CLI | FR-CLI-001, FR-CLI-002, FR-CLI-005, FR-CLI-006, FR-CLI-007 (Core); FR-CLI-003, FR-CLI-008, FR-CLI-009, FR-CLI-010, FR-CLI-011, FR-CLI-012, FR-CLI-013, FR-CLI-014, FR-CLI-015, FR-CLI-016, FR-UX-001, FR-UX-002, FR-UX-003 |
| 2 | Functional TUI | FR-TUI-001 through FR-TUI-009, FR-UX-040 through FR-UX-043, FR-TUI-060, FR-TUI-061, FR-TUI-063, FR-TUI-064, FR-TUI-066, FR-TUI-067, FR-TUI-068, FR-UX-070 through FR-UX-076 |
| 3 | Agent runtime | FR-AGT-001, FR-AGT-002, FR-AGT-005, FR-AGT-015 (Core) |
| 4 | Basic Planner | FR-AGT-007, FR-AGT-008, FR-AGT-009 |
| 5 | Execution Engine | FR-AGT-010, FR-AGT-011, FR-AGT-012 |
| 6 | Context Manager | FR-CTX-001 through FR-CTX-007, NFR-CTX-001 |
| 7 | Tool Runtime | FR-TOOL-001, FR-TOOL-002, FR-TOOL-003 (Core); FR-TOOL-004, FR-TOOL-005, FR-TOOL-006, FR-TOOL-008 |
| 8 | Permission Manager | FR-SEC-100, FR-SEC-103 (Core); FR-SEC-104, FR-SEC-105, FR-SEC-113 |
| 9 | Workspace Engine | FR-AGT-006 |
| 10 | Terminal | FR-TOOL-007 (terminal entry of the built-in catalog), FR-TOOL-006 |
| 11 | Filesystem tools | FR-TOOL-007 (filesystem entries of the built-in catalog) |
| 12 | Basic Git | FR-GIT-001, FR-GIT-002, FR-GIT-003, FR-GIT-004, FR-GIT-005, NFR-GIT-001 |
| 13 | Provider abstraction | FR-PROV-001, FR-PROV-002, FR-PROV-010, FR-PROV-050 (Core); FR-PROV-011, FR-PROV-080 |
| 14 | At least one cloud provider | FR-PROV-082 (Anthropic); FR-PROV-081 (generic OpenAI-compatible, endpoint-configured) |
| 15 | At least one local provider | FR-PROV-083 (Ollama); FR-PROV-084, FR-PROV-085 |
| 16 | Streaming | FR-PROV-020, FR-TUI-007, FR-UX-070; CLI stream documents per FR-CLI-006 |
| 17 | Configuration | FR-CFG-001 through FR-CFG-005, FR-CFG-007 through FR-CFG-011 |
| 18 | Logging | FR-OBS-002, FR-OBS-003, FR-OBS-004 |
| 19 | Session persistence | FR-ARCH-009, FR-AGT-003, NFR-AGT-002, FR-CFG-009 |
| 20 | macOS Tier 1 | FR-PORT-004, NFR-PORT-001 |
| 21 | Linux Tier 1 | FR-PORT-004, NFR-PORT-001 |
| 22 | Installation | FR-REL-003, FR-REL-009 |
| 23 | Basic update | FR-REL-005, FR-REL-006, FR-REL-016 |
| 24 | Unit and integration tests | FR-TEST-001, FR-TEST-002, FR-TEST-003, NFR-TEST-003 (SM-14 gate) |
| 25 | Main E2E test | FR-TEST-001 (E2E layer); the UC-01 journey named in FR-AGT-001's verification method |
| 26 | GitHub Actions | FR-GH-009 (Core), FR-GH-011, FR-GH-001 |
| 27 | Signed releases (when viable) | FR-REL-002; signing enablement PENDING VALIDATION per the Volume 1 signing viability note (register row V15-OQ-3 mirrors V1-OQ-1/V14-OQ-1) |

**Direction 2 result: widening drift found.** The aggregated MVP set is substantially
larger than the twenty-seven-item minimum. The minimum is a floor by its own wording, so
the widening is not a contradiction — but it is material schedule exposure under
RISK-PRD-004 (scope creep versus MVP viability), and per this volume's mandate it is
flagged, not fixed. The beyond-minimum MVP-phased groups are:

- **Memory subsystem** (FR-MEM-001 through FR-MEM-005, FR-MEM-007 through FR-MEM-010) and
  **indexing subsystem** (FR-IDX-001 through FR-IDX-006) — the minimum names only the
  Context Manager (item 6).
- **Sandbox and audit** (FR-SEC-101, FR-SEC-106 through FR-SEC-108, FR-SEC-111,
  FR-SEC-112) — the minimum names only the Permission Manager (item 8).
- **External IPC control surface** (FR-ARCH-007).
- **Traces, metrics, cost rollups, and the full telemetry consent stack** (FR-OBS-007
  through FR-OBS-013) — the minimum names only logging (item 18).
- **Cost accounting and resilience breadth** in the provider layer (FR-PROV-030,
  FR-PROV-031, FR-PROV-041, FR-PROV-042, FR-PROV-043).
- **TUI platform-screen and interaction breadth** (FR-TUI-060 through FR-TUI-068 except
  the Beta items, FR-UX-070 through FR-UX-076).
- **Distribution breadth** (FR-REL-004 air-gapped installation, FR-REL-008 rollback,
  FR-REL-010, FR-REL-011, FR-REL-014, FR-REL-015).
- **Performance/reliability operations** (FR-PERF-001 through FR-PERF-006).

This drift is recorded as open question V15-OQ-1 (register): each owning volume SHOULD
either confirm its MVP phasing as load-bearing (many of these items back MVP exit
criteria — e.g., FR-PROV-085 and FR-IDX-005 back the offline guarantee; FR-SEC-111 backs
SM-13) or re-phase the non-load-bearing remainder to Beta through the change procedure.
No register is edited by this volume.

### MVP entry criteria

- Core exit criteria met for every Core item the MVP consumes (Volume 1 chapter 05).
- The FR-TEST-009 qualification pipeline exists at least in skeleton (gates T0–T1 wired),
  so MVP work is measured from its first merge.

### MVP exit criteria

Volume 1 chapter 05 fixes the five exit conditions; operationally they bind to:

1. Every MVP-minimum item functional — verified per the conformance table above.
2. Acceptance suite and the UC-01 E2E journey green on all Tier 1 platforms
   (NFR-PORT-001, SM-17 binding at MVP exit).
3. Offline guarantee list green under the offline suite with a local provider
   (FR-TEST-005, NFR-PERF-024, FR-PROV-085, SM-05).
4. Installation and basic update working on macOS and Linux (FR-REL-003, FR-REL-005,
   FR-REL-006).
5. Releases produced by CI with checksums (FR-GH-011, FR-REL-001, FR-REL-002); signatures
   per the signing viability note.

### MVP quality gates

| Gate | Definition | Source |
|---|---|---|
| Lint clean | `scripts/spec_lint.py` zero errors on `docs/spec/`; gofmt/golangci-lint clean on code | Volume 0 chapter 10; ADR-018; FR-GH-009 |
| Coverage | ≥ 70% overall and ≥ 85% strict scope (Core Domain, ports, SDK contracts), ratchet never decreases | NFR-TEST-003 (SM-14) |
| Determinism and flake control | Nightly determinism lane; quarantine per ADR-177 with dwell-time bound | NFR-TEST-002, NFR-TEST-005 |
| Traceability | 0 orphan side effects (SM-13); development chain audit | NFR-OBS-003, NFR-GH-001 |
| Mediation and egress | 100% permission mediation of side effects; zero-egress default posture | NFR-SEC-002, NFR-OBS-006 |
| Release qualification | Gate tiers T0–T4 wired; publication refused without a complete evidence bundle | FR-TEST-009, NFR-TEST-006; Volume 14 release audit |

## Beta

Beta hardens and broadens: reliability targets become gates, public contracts stabilize,
and the extension surfaces are exercised by real third-party extensions.

| Area | Capability set | Requirement IDs |
|---|---|---|
| Architecture (Vol 3) | Headless operating mode; port contract stability regime | FR-ARCH-008, NFR-ARCH-002 |
| Agent runtime (Vol 4) | Sub-agent delegation; the full workflow engine and SDD (keystone), definition format, execution semantics, gates, machine conformance, recovery/resume, cancellation/rollback, skill application, scheduling, durable timers; format stability, resume fidelity, orchestration overhead | FR-AGT-004, FR-WF-001, FR-WF-002, FR-WF-003, FR-WF-004, FR-WF-005, FR-WF-006, FR-WF-007, FR-WF-008, FR-WF-009, FR-WF-010, NFR-WF-001, NFR-WF-002, NFR-WF-003 |
| Providers and auth (Vol 5) | Provider integration effort target; OAuth authorization-code and device grants; enterprise proxies and trust anchors; temporary credentials; token refresh; credential resolution latency | NFR-PROV-001, FR-AUTH-003, FR-AUTH-004, FR-AUTH-006, FR-AUTH-007, FR-AUTH-010, NFR-AUTH-003 |
| Extension surfaces (Vol 6) | Extension SDK (keystone) and tool-creation target; MCP client (keystone) with transports, discovery/bridging, authorization, health/maintenance, trust gating, conformance program, connection machine; skill format (keystone), loading, composition, testing, distribution; plugin runtime over ARP (keystone), handshake/negotiation, permission/sandbox containment, supervision, package operations, sources and resolution, verification, machine conformance; skill and plugin latency budgets | FR-SDK-001, NFR-SDK-001, FR-MCP-001, FR-MCP-002, FR-MCP-003, FR-MCP-004, FR-MCP-005, FR-MCP-006, FR-MCP-007, FR-MCP-008, FR-SKILL-001, FR-SKILL-002, FR-SKILL-003, FR-SKILL-004, FR-SKILL-005, NFR-SKILL-001, FR-PLUG-001, FR-PLUG-002, FR-PLUG-003, FR-PLUG-004, FR-PLUG-005, FR-PLUG-006, FR-PLUG-007, FR-PLUG-008, FR-PLUG-009, NFR-PLUG-002, NFR-PLUG-003 |
| Memory (Vol 7) | Compression and summarization (consolidation) | FR-MEM-006 |
| CLI and TUI (Vol 8) | Extension-contributed commands; structured-output schema stability; quick actions; accessible output mode; terminal compatibility conformance | FR-CLI-004, NFR-CLI-001, FR-TUI-062, FR-TUI-065, NFR-TUI-070 |
| Configuration and observability (Vol 10) | Include mechanism; logging hot-path, event bus, and instrumentation overhead budgets | FR-CFG-006, NFR-OBS-001, NFR-OBS-002, NFR-OBS-005 |
| Git and hosting (Vol 11) | Remote operations; history modification, conflicts, recovery; hosting integration layer (`github` tool); change-request preparation flow; destructive-operation recoverability | FR-GIT-006, FR-GIT-007, FR-GIT-009, FR-GIT-010, NFR-GIT-002 |
| Performance (Vol 12) | Degradation responsiveness budget | NFR-PERF-028 |
| Testing (Vol 13) | Mutation score on scoped packages | NFR-TEST-004 |
| Distribution (Vol 14) | Update automation policy; deprecation policy | FR-REL-007, FR-REL-013 |

### Beta entry criteria

- MVP shipped: MVP exit criteria met and the release published (Volume 1 chapter 05).
- The MVP-widening open question V15-OQ-1 dispositioned (confirmed or re-phased), so the
  Beta plan starts from an audited baseline.

### Beta exit criteria

1. The v1 requirement set (chapter 03 backlog of this volume) is feature-complete.
2. Public contracts frozen as release candidates: remaining changes additive or through
   deprecation (FR-REL-013); contract-diff evidence per NFR-ARCH-002 and the SM-20 regime.
3. Beta-bound metrics meet their targets on Tier 1 platforms: NFR-PROV-001, NFR-SDK-001,
   NFR-AUTH-003, NFR-SKILL-001, NFR-PLUG-002, NFR-PLUG-003, NFR-WF-001, NFR-WF-002,
   NFR-WF-003, NFR-CLI-001, NFR-TUI-070, NFR-OBS-001, NFR-OBS-002, NFR-OBS-005,
   NFR-PERF-028, NFR-TEST-004, NFR-GIT-002, NFR-ARCH-002.
4. Upgrade and rollback paths from every Beta build to the v1 candidate tested
   (FR-REL-006, FR-REL-008; Volume 13 upgrade matrices).

### Beta quality gates

All MVP gates continue at their ratcheted levels, plus: the mutation lane (NFR-TEST-004,
ADR-175) reports at phase gates; the MCP conformance suite and interop job run per release
(FR-MCP-007, preparing the v1-bound NFR-MCP-001/NFR-MCP-002); breaking changes are
permitted only with migration notes and a recorded decision (Volume 1 Beta discipline);
extension-surface conformance suites (ARP fixtures, skill corpus, SDK walkthroughs) gate
every release that touches those contracts.

## v1

v1 is the first stable release: SemVer guarantees attach to the public contracts
(FR-REL-012), and the measured targets deferred to v1 become release gates.

| Area | Capability set | Requirement IDs |
|---|---|---|
| Providers and auth (Vol 5) | Local-model conformance target; service accounts and managed identity | NFR-PROV-002, FR-AUTH-005 |
| Extension surfaces (Vol 6) | MCP conformance pass rate and reference-server interoperation; plugin creation time | NFR-MCP-001, NFR-MCP-002, NFR-PLUG-001 |
| Security (Vol 9) | Release vulnerability posture; coordinated disclosure first response | NFR-SEC-001, NFR-SEC-003 |
| Observability (Vol 10) | Run record completeness for replay | NFR-OBS-004 |
| Performance (Vol 12) | The full performance-target set as release gates: startup, TUI latency, first token, streaming, tool dispatch, filesystem/git/indexing/memory/search/patch/diff, session restore, RAM/CPU/disk, concurrency, large repositories and files, long sessions, tool-call reliability, recovery, crash-free operation | NFR-PERF-001 through NFR-PERF-023, NFR-PERF-025, NFR-PERF-026, NFR-PERF-027 |
| Distribution (Vol 14) | Update time, rollback time, public-contract stability as measured commitments | NFR-REL-001, NFR-REL-002, NFR-REL-003 |

Functional additions classified v1 by their owning volumes (referenced by name, phases per
the owning registers): the GitLab hosting tool of ADR-147; the v1 tranche of the provider
adapter catalog (Volume 5 chapter 09) and of the built-in integration tools (Volume 6
chapter 03, ADR-074); OS-level sandbox isolation claims where the ADR-021 validations
succeed.

**Entry:** Beta exit criteria met. **Exit:** v1 published and in maintenance under the
support policy (FR-REL-014, ADR-193; governance in [chapter 04](04-open-source-governance.md));
exit occurs when v2 ships or v1 support ends per that policy.

**v1 quality gates:** all Beta gates at their ratcheted levels; coverage ≥ 80% overall and
≥ 90% strict scope (NFR-TEST-003); the Volume 12 benchmark suite gating releases across
its full NFR set; NFR-SEC-001 (zero known critical/high vulnerabilities at publication)
gating every release; the release audit of Volume 14 verifying bump class against
contract-diff evidence (FR-REL-012).

## v2

Candidate v2 scope, recorded per Volume 1 chapter 05 and confirmed by the change procedure
when v1 ships:

- **Native Windows 11 support** (x86_64; arm64 subject to viability) — PAL backend
  mappings per the Volume 3 Windows-future rules (RISK-PORT-003; validation spike
  required).
- **WASM in-process plugin isolation** as a possible third tool channel — v2 candidate per
  ADR-009/ADR-073 review conditions.
- **Marketplace and distribution expansions** that Volumes 6 and 14 classify post-v1,
  where their PENDING VALIDATION items resolve in favor (register row V15-OQ-4 tracks the
  hand-off).
- The v2 tranche of the provider adapter catalog and integration tools per the owning
  volumes' catalogs.

No corpus requirement carries phase v2 today; v2 items are sequenced only after the
confirming change-procedure decision.

## Future

Desirable, uncommitted; no committed requirement may depend on these (Volume 1 chapter
05): the curated extension marketplace as a registry-kind source (ADR-080); hosted Linux
package repositories (ADR-190); application-level at-rest encryption of memory databases
(Volume 7 register, V7-OQ-3); local exact tokenizers per model family (ADR-087
alternative); ANN retrieval acceleration beyond the ADR-020 assumption; platforms beyond
the Volume 1 matrix (other Unix systems); macOS Intel Tier 2 if the capacity validation
fails its MVP-era window (falls to Future per V1-OQ-2).

## Out of Scope

Owned and enumerated by Volume 1 chapter 05; restated here only as the planning boundary:
a hosted Andromeda service; graphical clients; model training/fine-tuning/weight
management; full editor/IDE functionality; unofficial integrations of any kind; a
general-purpose consumer assistant; Windows versions earlier than Windows 11.
Reclassification requires the Volume 0 change procedure; this volume never schedules work
against an Out of Scope item.

## Phase-gate operation

Each phase boundary is a **phase gate** (Volume 1 glossary): an audit verifying exit
criteria, bound metrics, and the risk register before the next phase begins. Gates are
executed as the T4 tier of Volume 13 chapter 01, consume the qualification evidence
bundles of FR-TEST-009, and are recorded as `release` issues with the Volume 11
`phase/*` milestone mechanics. The gate review MUST cover: the exit-criteria checklist of
this chapter, the metric bindings listed per phase, every open `RISK-*` entry touching the
phase (chapter 03 ordering rules), and the open-questions registers of all volumes for
items whose resolution windows fall inside the phase.
