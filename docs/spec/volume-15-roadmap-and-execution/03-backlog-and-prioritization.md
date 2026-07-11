# 03 — Backlog and Prioritization

This chapter fixes how work is ordered: the prioritization model that combines the Volume
0 priority scale with the Volume 1 phases, the consolidated backlog for v1, v2, and
Future, and the risk-driven rules that may reorder any of it. It mints nothing; every
item cited is owned elsewhere.

## Prioritization model

Two orthogonal dimensions order every work item:

- **Phase** (Volume 1 chapter 05, carried by every requirement) answers *when the
  capability is committed*. Phases order the macro plan: no committed work targets a
  later phase while its own phase's exit criteria are unmet, except where the dependency
  graph of [chapter 02](02-epics-milestones-sequencing.md) requires early foundations.
- **Priority** (Volume 0 chapter 02) answers *how urgently an item is handled within its
  phase*: `P0` (blocking), `P1` (high), `P2` (normal), `P3` (low). On the development
  platform, priority is the `priority/*` label and Project field (Volume 11 chapter 05),
  where `p0` means drop-everything.

Combined semantics — the scheduling contract for triage:

| Priority | Within the active phase | Outside the active phase |
|---|---|---|
| P0 | Blocks the phase gate; blocks the release train; work stops until resolved. Examples: a red required check on `main`, a Critical/High security finding (NFR-SEC-001 class), a broken MVP-minimum item | Only security and integrity classes may be P0 outside the active phase (they jump phases by rule R1 below) |
| P1 | Scheduled into the current or next iteration; a phase gate MAY pass with open P1 items only with a recorded waiver in the gate evidence (waiver policy per Volume 13 chapter 04) | Enters the first iteration of its phase |
| P2 | Ordinary iteration planning by epic order | Backlog, ordered by epic |
| P3 | Scheduled only when it rides along with adjacent work | Backlog; candidates for `good-first-issue` curation |

Rules:

1. Priority never overrides phase for *feature* work: a P1 v2 feature does not start
   before v1 ships. Priority does override iteration order *within* a phase.
2. Severity (`severity/*`, bugs and security findings) maps to priority at triage:
   `critical` → P0, `high` → P1 by default, adjustable downward only with a recorded
   justification on the issue.
3. Priorities are set at triage and revisited at every phase gate; silent priority decay
   is prevented by the nightly audit surfacing stale `status/triage` items (FR-GH-008
   automation).

## Risk-driven ordering rules

Risk entries (`RISK-*`) carry no phase; they reorder the backlog through the following
standing rules. Each rule names the risks that motivate it; the risk registers of the
owning volumes remain authoritative for mitigation content.

| Rule | Ordering effect | Motivating risks |
|---|---|---|
| R1 — Security and integrity jump the queue | A finding that realizes (or materially raises the probability of) a security or state-integrity risk is scheduled ahead of feature work in any phase, at P0/P1 per severity | RISK-PRD-005, RISK-SEC-001 through RISK-SEC-027, RISK-AUTH-002, RISK-CFG-003, RISK-UX-079, RISK-REL-001 |
| R2 — Enforcement before construction | Mechanical guardrails (dependency checks, contract kits, traceability validators, leak scans) are built before the code volumes they guard grow — the cost of retrofit rises with every merge | RISK-ARCH-001, RISK-ARCH-002, RISK-GH-002, RISK-TEST-001, RISK-TEST-003 |
| R3 — De-risk external drift with fixtures first | Work against external surfaces (providers, git versions, hosting APIs, MCP revisions) lands its recorded-fixture and equivalence suites in the same epic as the integration, never later | RISK-PROV-001, RISK-PROV-080, RISK-GIT-001, RISK-GIT-004, RISK-MCP-002, RISK-TOOL-003 |
| R4 — Contract freeze discipline | As Beta exit approaches, changes to public contracts are ordered before all other Beta work, so the freeze (SM-20 regime) happens on a tested surface rather than a moving one | RISK-PRD-010, RISK-CLI-003, RISK-PLUG-002 |
| R5 — Scope-creep containment | Any proposal that adds MVP scope after this volume sequences it is ordered *behind* the existing MVP set and requires the change procedure with recorded justification (Volume 1 MVP discipline) | RISK-PRD-004 |
| R6 — Recovery paths before load | Crash-injection, rollback, and degraded-mode work is scheduled before the features that depend on those paths are declared done — recovery is not a hardening afterthought | RISK-ARCH-004, RISK-AGT-003, RISK-REL-001, RISK-REL-003, RISK-GIT-002, RISK-PERF-003 |
| R7 — Ecosystem trust before ecosystem growth | Signature verification, trust gating, and descriptor pinning land before any push for third-party extension adoption; growth without the trust rails converts adoption into exposure | RISK-PLUG-001, RISK-PLUG-003, RISK-SKILL-001, RISK-MCP-001, RISK-TOOL-001, RISK-TOOL-002 |
| R8 — Measure before optimizing | Benchmark and metric infrastructure (FR-PERF-005, ADR-160/ADR-161) precedes performance work; optimization without calibrated baselines is unverifiable | RISK-PERF-001, RISK-PERF-002 |
| R9 — Sustainability guard | Work that concentrates knowledge in one person (release pipeline, signing, security response) is paired with documentation and a second operator before the phase gate that depends on it | RISK-PRD-003, RISK-REL-002, RISK-REL-004 |

At every phase gate, the risk-register review (Volume 1 verification method for the
RISK-PRD series) re-evaluates these rules' application and records reorderings in the
gate evidence.

## v1 backlog by area

The v1 phase content is the set of requirements phased `v1` in the Volumes 1–14 registers
(consolidated in [chapter 01](01-phase-plan.md)) plus the functional tranches the owning
volumes classify v1 by name. Ordered by area, with the risk rules that shape ordering:

| Order | Area | Backlog content | Ordering rationale |
|---|---|---|---|
| 1 | Performance gates (Vol 12) | Bring NFR-PERF-001 through NFR-PERF-023 and NFR-PERF-025, NFR-PERF-026, NFR-PERF-027 to gating status: calibrated runners, full DS-* dataset coverage, regression gates armed per ADR-161 | R8 — everything else in v1 is measured against these gates |
| 2 | Security posture (Vol 9) | NFR-SEC-001 (zero known critical/high vulnerabilities gating publication), NFR-SEC-003 (disclosure first-response ≤ 3 business days; process in [chapter 04](04-open-source-governance.md)) | R1, R9 |
| 3 | Reproducibility (Vol 10) | NFR-OBS-004: run-record completeness and replay divergence as release gates (SM-12) | R6 |
| 4 | Distribution guarantees (Vol 14) | NFR-REL-001 (update time), NFR-REL-002 (rollback time), NFR-REL-003 (public-contract stability, SM-20) as measured release commitments | R4 |
| 5 | Local-model conformance (Vol 5) | NFR-PROV-002: conformance suite per release on ≥ 2 local serving paths (SM-04) | R3; PRD-003 |
| 6 | Extension ecosystem maturity (Vol 6) | NFR-MCP-001, NFR-MCP-002 (conformance and interop as gates), NFR-PLUG-001 (plugin creation time) | R7 |
| 7 | Enterprise authentication (Vol 5) | FR-AUTH-005 (service accounts, managed identity) with its per-provider validation spikes (Volume 5 register V5B-OQ-3) | R3 |
| 8 | Hosting breadth (Vol 11) | The GitLab tool tranche per ADR-147 (v1 phase; parity validation per Volume 11 register V11-OQ-3) | R3 |
| 9 | Catalog tranches (Vols 5, 6) | v1-classified provider adapters (Volume 5 chapter 09 catalog) and integration tools (Volume 6 chapter 03, ADR-074), each landing with its conformance fixtures | R3, R7 |
| 10 | Sandbox depth (Vol 9) | OS-level isolation claims where the ADR-021 per-platform validations succeed (Volume 9 register V9B-OQ-1/V9B-OQ-2); otherwise the process-level floor remains the honest claim | R1; honesty over breadth |

## v2 backlog (candidate, unconfirmed)

Per Volume 1 chapter 05, v2 scope is confirmed by the change procedure when v1 ships.
Candidate ordering:

| Order | Item | Source of the candidacy |
|---|---|---|
| 1 | Native Windows 11 support (x86_64 first): PAL backend spike, then surface-by-surface implementation against the Volume 3 Windows-future rules | Volume 1 chapter 05; RISK-PORT-003 requires the validation spike before commitment |
| 2 | v2 catalog tranches: remaining provider adapters and integration tools classified v2 by Volumes 5/6 | ADR-065 generic-adapter-first keeps reachability meanwhile |
| 3 | WASM in-process plugin isolation as a third tool channel | ADR-009/ADR-073 review-condition spike (Volume 6 register V6A-OQ-3) |
| 4 | Marketplace/distribution expansions classified post-v1 | Volume 1 chapter 05; predicated on ADR-080's registry mechanism operating |

## Future backlog

Uncommitted; items leave `Future` only through the change procedure. Listed so the
backlog has one consolidated home:

- Curated extension marketplace as a registry-kind source (ADR-080); depends on the
  official registry-index decision (Volume 6 register V6B-OQ-4).
- Hosted Linux package repositories (apt/rpm repos) beyond attached package files
  (ADR-190).
- Application-level at-rest encryption for memory databases (Volume 7 register V7-OQ-3).
- Local exact tokenizers per model family (ADR-087 alternative; Volume 7 register
  V7-OQ-1).
- ANN retrieval acceleration if corpus scale breaches the ADR-020 assumption (Volume 7
  register V7-OQ-2; RISK-IDX-001).
- Additional Unix-like platforms per the Volume 1 platform matrix; macOS Intel if its
  viability window closes (V1-OQ-2 fallback).
- Multi-step rollback beyond one retained version (ADR-192 reversal plan).

## Backlog operation

The single operational backlog is the **Andromeda Roadmap** project (Volume 11 chapter
05): this chapter's tables are its sequencing narrative, and there is deliberately no
second roadmap document to drift. Intake flows through the issue forms; triage assigns
type, area, phase, priority, and severity; epics fan out requirements to issues; the
nightly traceability audit (FR-GH-001) keeps the chain honest. Changes to *this chapter*
— reordering areas, promoting candidates, confirming v2 scope — follow the Volume 0
change procedure and are decided per the governance model of
[chapter 04](04-open-source-governance.md).
