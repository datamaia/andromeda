# Checklists

Three review checklists consolidated at Phase C from the owning volumes. Every item is a
yes/no question answerable with recorded evidence, and names the requirement, NFR, or ADR it
verifies. This annex mints nothing and restates no normative text: on any doubt about an
item's exact meaning, the cited source governs.

Handling a "no": in the release checklist, a failing item follows the Volume 13 waiver
policy (non-waivable items are marked; everything else needs a recorded, expiring,
maintainer-approved waiver). In the security and architecture review checklists there is no
waiver path — a "no" is a finding that becomes a tracked defect or a change-procedure
proposal before the review can conclude.

## 1. Release checklist

Sources: [Volume 13, chapter 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md)
(gate ladder, qualification pipeline S1–S6, evidence bundle, waiver policy) and
[Volume 14](../volume-14-distribution/00-index.md) (artifacts, integrity metadata, channels,
versioning, rollback). The pipeline stages run in order; a release is publishable only when
every item below is "yes" or carries a valid waiver.

### Qualification gates (S1–S4)

| # | Check (yes/no) | Verifies |
|---|---|---|
| R-01 | Was the candidate built by the release pipeline for every Tier 1 target (macOS arm64; Ubuntu 22.04, Debian 12, Fedora 39 on x86_64 and arm64)? | FR-REL-001; FR-PORT-004 |
| R-02 | Did the full test pyramid plus acceptance suite pass on every Tier 1 platform, with per-platform results attached to the release (S1)? | FR-TEST-009; NFR-PORT-001 (SM-17) |
| R-03 | Did the offline suite pass under OS-level network disablement, including the local-provider path and mirror fixtures (S2)? **Non-waivable** | FR-TEST-005; FR-PROV-085; FR-REL-004 (SM-05) |
| R-04 | Did the security gates pass: CodeQL, dependency audit, and secret scanning green (S2)? Secret scanning is **non-waivable** | NFR-SEC-001 (SM-16a) |
| R-05 | Did permission-mediation enforcement pass (no unmediated side-effect path)? **Non-waivable** | NFR-SEC-002 (SM-16b) |
| R-06 | Did the compatibility suite pass on the full Tier A terminal set, with no probe-versus-matrix mismatch? | NFR-TUI-070; Volume 8, chapter 12 |
| R-07 | Did migration fixtures pass, including the newer-database clean-refusal case (S3)? Migration-refusal semantics are **non-waivable** | ADR-029; FR-CFG-008; FR-CFG-009 |
| R-08 | Did the automated N−1 → N upgrade test pass within budget (S3)? | FR-REL-006; NFR-REL-001 (SM-18) |
| R-09 | Did the offline rollback test pass, including binary–database pairing behavior (S3)? | FR-REL-008; ADR-192; NFR-REL-002 (SM-19) |
| R-10 | Were all benchmark results within the Volume 12 budgets on both reference machines (S4)? | FR-TEST-009 stage S4; Volume 12, chapter 01 |

### Artifact integrity (S5)

| # | Check (yes/no) | Verifies |
|---|---|---|
| R-11 | Does every artifact follow the fixed naming grammar and declared package-format set? | ADR-190 |
| R-12 | Are published checksums present and re-verified for every artifact? **Non-waivable** | FR-REL-002 |
| R-13 | Are cosign signatures present and valid exactly where signing is enabled, and do the release notes state the signing/notarization status (macOS notarization PENDING VALIDATION per V14-OQ-1)? | FR-REL-002; ADR-013 |
| R-14 | Is a per-artifact SBOM attached and well-formed? | FR-REL-002; ADR-013 |
| R-15 | Does SLSA provenance exist for every artifact and reference the qualification bundle digest? | FR-REL-002; ADR-013; FR-TEST-009 |
| R-16 | Did install, uninstall (data preserved by default), and per-artifact smoke runs pass on Tier 1 platforms, including Homebrew tap and shell-installer channels? | FR-REL-003; FR-REL-010 |
| R-17 | Is the Linux release binary statically linked, with the clean-machine run succeeding without undeclared prerequisites? | NFR-PORT-003 |

### Decision, versioning, and publication (S6)

| # | Check (yes/no) | Verifies |
|---|---|---|
| R-18 | Is the evidence bundle complete, schema-valid, and its decision "qualified" (no unmeasured gate counted as passed)? | FR-TEST-009; E-TEST-005 |
| R-19 | Are all waivers recorded, approved, unexpired, visible in release notes — and is no gate waived for a third consecutive release? | Volume 13 waiver policy; NFR-TEST-006 |
| R-20 | Is the version-bump class (major/minor/patch) consistent with the contract-diff result over all public contracts? | FR-REL-012; NFR-REL-003 (SM-20); ADR-015 |
| R-21 | Are the changelog and release notes present, accurate against the commit range, and free of AI/vendor attribution in commit messages? | FR-REL-015; ADR-015 |
| R-22 | Is the deprecation ledger reconciled (every removal preceded by its announced window)? | FR-REL-013 |
| R-23 | Is the release placed on the correct branch and support window, with only closed-list backports on release branches? | FR-REL-014; ADR-193 |
| R-24 | Is release metadata correct for channel subscription semantics (channel maturity, upgrade-path admissibility, yank status)? | FR-REL-005; ADR-191 |
| R-25 | Was publication executed only through tooling that refuses candidates without a valid bundle? | FR-TEST-009; NFR-TEST-006 |

## 2. Security review checklist

Source: [Volume 9](../volume-09-security/00-index.md) — threat model (chapters 01–04),
permission model (chapter 05), sandbox (chapter 06), secrets (chapter 07), audit and
incident response (chapter 08), approval machine (chapter 09). Applies to any security
review: of the whole product at a phase gate, or of a feature/extension surface before
merge.

### Threat coverage

| # | Check (yes/no) | Verifies |
|---|---|---|
| S-01 | Is every one of the 27 cataloged threats (RISK-SEC-001 through RISK-SEC-027) either addressed by the reviewed change's controls or recorded as not applicable, with reasons? | Volume 9, chapters 01–04 |
| S-02 | Does each applicable threat entry remain complete (Asset, Actor, Vector, Preconditions, Impact, Prevention, Response, Recovery, Residual risk, Tests) after the change? | Volume 0 risk template; Volume 9 register |
| S-03 | Are new inputs from untrusted origins (files, tool results, MCP content, model output, memory, index hits) labeled as data and never treated as privileged instructions? | RISK-SEC-001, RISK-SEC-002; Volume 9 trust boundaries |
| S-04 | Do the threat tests named in each applicable entry's Tests row exist and run in a gating suite? | RISK-SEC-001..027 Tests rows; Volume 13 security suites |

### Permission bindings

| # | Check (yes/no) | Verifies |
|---|---|---|
| S-05 | Is every dangerous action bound to exactly the frozen permission names (13-member enum: `read` … `system_modification`), with no ad-hoc permission invented? | FR-SEC-100 |
| S-06 | Are grant scopes drawn only from the ten frozen scope qualifiers and decisions only from the seven-member decision enum? | FR-SEC-100 |
| S-07 | Does evaluation follow deny-overrides with the ask-forcing tier (deny > ask > allow > default-ask), with no specificity or recency ordering introduced? | ADR-121; FR-SEC-103 |
| S-08 | Do grants persist, expire, and revoke per specification, with revocation taking effect on the next evaluation and surviving backup/restore correctly? | FR-SEC-104 |
| S-09 | Do non-interactive and headless paths resolve permissions policy-only — no prompt, deterministic denial (exit code 5) when unresolved, at parity with interactive semantics? | FR-SEC-105 |
| S-10 | Does every permission decision, grant, and revocation produce its audit record and event (`permission.decision.recorded`, `permission.grant.*`)? | FR-SEC-100; FR-SEC-111 |
| S-11 | Do approvals follow the full lifecycle machine (request, decision, expiry, cancellation) with the 10-minute default timeout and prompt text naming the concrete effect? | FR-SEC-113; RISK-SEC-027 |

### Sandbox

| # | Check (yes/no) | Verifies |
|---|---|---|
| S-12 | Is every child process launched exclusively through SandboxPort (no direct spawn path anywhere outside the sanctioned families)? | FR-SEC-101; FR-ARCH-005 |
| S-13 | Is each execution subject mapped to exactly one of the five sandbox tiers (`process`, `tool`, `workflow`, `plugin`, `mcp_server`), with effective policy = tier ∩ declaration ∩ grants? | ADR-122; FR-SEC-106 |
| S-14 | Is the effective containment level recorded per execution, and does isolation degradation surface observably (`sandbox.containment.degraded`) — never silently? | ADR-021; FR-SEC-101 |
| S-15 | Are environments filtered to the allowlist and resolved secrets scrubbed from every sandboxed environment? | FR-SEC-107 |
| S-16 | Do filesystem policy checks resolve symlinks before acting, confine to declared roots, use per-sandbox temp directories, and clean up on teardown (including crash-orphan sweeps)? | FR-SEC-108 |
| S-17 | Do command allow/denylists apply as configured, and do resource limits (CPU, memory, processes, open files) enforce with defined violation behavior? | FR-SEC-101; FR-SEC-106; `[sandbox]` keys |

### Secrets and redaction

| # | Check (yes/no) | Verifies |
|---|---|---|
| S-18 | Is there no plaintext secret material at rest anywhere the change touches (canary scan green)? | FR-SEC-102; NFR-AUTH-001; ADR-014 |
| S-19 | Is the age-encrypted fallback store used only under explicit recorded consent, never auto-enabled on keychain failure? | FR-SEC-110; E-PORT-003 |
| S-20 | Does every sink (logs, errors, events, traces, memory records, TUI/CLI output) render through redaction, with registry exact-match plus structural and heuristic layers, failing closed? | FR-SEC-109; ADR-124; NFR-SEC-004; NFR-AUTH-002 |
| S-21 | Is secret access mediated by the `credential_access` permission and audited (`secret.accessed`)? | FR-SEC-102; FR-SEC-100 |
| S-22 | Are authentication mechanisms limited to officially documented ones, with the prohibited-mechanism gate intact? | FR-AUTH-001 |

### Audit and incident readiness

| # | Check (yes/no) | Verifies |
|---|---|---|
| S-23 | Does every audited action in the closed catalog yield exactly one Audit Record, written before the effect, including denied and failed outcomes? | FR-SEC-111; ADR-125 |
| S-24 | Is the audit chain verifiable (head check on open, full walk on demand) and does tampering trigger freeze-preserve-acknowledge with E-SEC-013 semantics? | ADR-123; NFR-SEC-005 |
| S-25 | Do security-relevant failures fail closed: evaluation failure denies, unwritable audit blocks the action (E-SEC-014), unappliable sandbox policy refuses launch, redaction failure withholds payloads? | ADR-125 |
| S-26 | Are incident triggers wired (update verification failure, chain break, leak detection), deduplicated, and do they produce `security.incident.opened`/`closed` with exportable evidence? | FR-SEC-112 |
| S-27 | Is the disclosure path current (SECURITY.md placement per Volume 11; first-response tracking)? | NFR-SEC-003 (SM-16c) |

## 3. Architecture review checklist

Source: [Volume 3](../volume-03-architecture/00-index.md) — layering and dependency matrix
(chapter 01, ADR-030/ADR-033), ports (chapter 02), components (chapters 03–06), PAL
(chapter 07), processes and concurrency (chapter 08). Applies to architecture reviews of
substantial changes and to the release audit.

### Layering and dependency matrix

| # | Check (yes/no) | Verifies |
|---|---|---|
| A-01 | Is every package assigned to exactly one layer in the layer manifest, and is the manifest current with the change? | ADR-033; ADR-031 |
| A-02 | Do the generated depguard rules and the import-graph test pass — zero PROHIBITED imports per the chapter 01 dependency matrix? | FR-ARCH-001; NFR-ARCH-001; ADR-030 |
| A-03 | Do engines (L2) hold only port interface values and import no L3 adapter package? | FR-ARCH-001; ADR-030 |
| A-04 | Is all construction and wiring confined to the composition root (`cmd/andromeda`), with adapters importable only there? | FR-ARCH-002 |
| A-05 | Does the SDK module mirror public contracts without importing `internal/`? | ADR-031; FR-ARCH-011 |

### Port contract stability

| # | Check (yes/no) | Verifies |
|---|---|---|
| A-06 | Are the 18 frozen port names and method signatures unchanged (contract-diff clean), with any intended change carried through the Volume 0 change procedure? | FR-ARCH-003; NFR-ARCH-002 |
| A-07 | Does every port method accept a context and honor cancellation, verified by the port's contract test kit? | FR-ARCH-004; FR-TEST-004 |
| A-08 | Do all port implementations (real adapters and fakes) pass the same contract kit? | FR-TEST-004; FR-ARCH-002 |
| A-09 | Are extension surfaces reachable only through versioned public contracts (SDK, ARP, MCP, ports) — no bypass into internals? | FR-ARCH-011 |

### PAL encapsulation

| # | Check (yes/no) | Verifies |
|---|---|---|
| A-10 | Does the prohibited-construct scan report zero occurrences of `runtime.GOOS`, OS build tags, `syscall`/`golang.org/x/sys`, or platform path/signal/mode literals outside `internal/pal`? | FR-PORT-001; NFR-PORT-004 |
| A-11 | Do exactly the 19 PAL surfaces exist, each with `Probe`, portable signatures, a declared degradation policy, and a maintained Windows-future mapping? | FR-PORT-002 |
| A-12 | Is every config/data/state/cache/log/runtime path resolved solely through the Config Directories surface (no hardcoded platform directory literals)? | FR-PORT-003; ADR-022 |
| A-13 | Does the PAL conformance suite pass on every Tier 1 platform for every surface a shipped feature consumes? | NFR-PORT-002; NFR-PORT-001 |

### Processes, concurrency, and lifecycle

| # | Check (yes/no) | Verifies |
|---|---|---|
| A-14 | Does every child process belong to one of the four sanctioned families with exactly one supervisor, and is the whole process tree terminated on cancellation? | FR-ARCH-005 |
| A-15 | Is all concurrency supervised (no naked goroutines; errgroup/context structure; bounded pools with backpressure)? | FR-ARCH-006; ADR-023 |
| A-16 | Does shutdown follow the defined ordering within its deadline, with forced-shutdown escalation recorded (E-ARCH-006)? | FR-ARCH-010; NFR-ARCH-003 |
| A-17 | Are termination paths leak-free (goroutine-leak gates and post-shutdown process scans green)? | NFR-ARCH-004 |
| A-18 | Does crash recovery reconcile recorded state against actual state with idempotent recovery steps (E-ARCH-007 on failure)? | FR-ARCH-009 |
| A-19 | Do IPC endpoints enforce same-user peer verification before request parsing, and does headless mode operate policy-only over that surface? | FR-ARCH-007; FR-ARCH-008; ADR-032 |

### State-machine completeness

| # | Check (yes/no) | Verifies |
|---|---|---|
| A-20 | Does every owned state machine use exactly the frozen state names of Volume 2, chapter 09 — no renamed, added, or repurposed states outside the change procedure? | Volume 2, chapter 09; FR-ARCH-003 discipline |
| A-21 | Does every owned machine define all mandated elements: initial state, terminal states, transitions, events, guards, side effects, persistence, recovery, timeouts, cancellation, retries, and errors? | Volume 0 state-machine rules; owning volumes' machine chapters |
| A-22 | Are machine transitions the only writers of their status fields, with transitions persisted per the Volume 2 write discipline and recovery paths tested by crash injection? | Volume 2, chapter 10; FR-ARCH-009 |
