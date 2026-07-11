# 99 — Volume 13 Register

Machine-parseable register of everything Volume 13 minted, per Volume 0 chapters 02 and 03.
Merged into the Volume 0 registers at consolidation.

## Requirements index

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-TEST-001 | Test pyramid and suite organization | MVP | Distribution report at T4; CI duration metrics; review checklist |
| FR-TEST-002 | Test-to-requirement traceability annotations | MVP | Traceability scanner in CI; T4 gate audit |
| FR-TEST-003 | Closed test-type catalog and classification | MVP | Classification check in T0; T4 audit against executed CI jobs |
| FR-TEST-004 | Port contract test kits | Core | Kit inventory check; kits in T0; consolidation audit |
| FR-TEST-005 | Offline test suite and network sentinel | MVP | Offline suite at T1/T3; harness self-tests; T4 report audit |
| FR-TEST-006 | Test doubles and test providers | MVP | Contract kits over fakes; ADR-033 import-graph check; live-lane cross-checks |
| FR-TEST-007 | Determinism controls and flaky-test quarantine | MVP | Nightly determinism lane; quarantine repository check; T4 accounting |
| FR-TEST-008 | Test data and secret handling | MVP | Secret-scanning checks; generator determinism tests; sanitizer tests |
| FR-TEST-009 | Release qualification pipeline and evidence bundle | MVP | Pipeline self-tests; bundle schema validation; T4 release audit |
| NFR-TEST-001 | Merge-gate wall-clock budget | MVP | Weekly CI duration report; T4 evaluation |
| NFR-TEST-002 | Suite determinism and order independence | MVP | Nightly determinism lane (50 repeats, shuffled) |
| NFR-TEST-003 | Coverage thresholds | MVP | Coverage merge gate; per-release report (SM-14) |
| NFR-TEST-004 | Mutation score on scoped packages | Beta | ADR-175 mutation lane report at phase gates |
| NFR-TEST-005 | Flake rate and quarantine dwell time | MVP | CI triage records and quarantine registry, weekly |
| NFR-TEST-006 | Qualification completeness | MVP | Post-publication audit job; phase-gate audit |
| RISK-TEST-001 | Pyramid inversion | — | Risk register review at phase gates |
| RISK-TEST-002 | Flaky tests eroding gate credibility | — | Risk register review at phase gates |
| RISK-TEST-003 | Coverage gaming | — | Risk register review at phase gates |
| RISK-TEST-004 | Test-double divergence from real systems | — | Risk register review at phase gates |
| RISK-TEST-005 | Secret or sensitive-data leakage through test assets | — | Risk register review at phase gates |
| RISK-TEST-006 | Gate erosion under release pressure | — | Risk register review at phase gates |

## ADRs minted

| ADR | Title | Status |
|---|---|---|
| ADR-175 | Mutation Testing: Scoped Packages, Phase-Gated Cadence, Tooling Pending Validation | Accepted |
| ADR-176 | Provider Test Doubles: Scripted Fakes First, Sanitized Recordings Second, Live Lanes Non-Gating | Accepted |
| ADR-177 | Flaky-Test Quarantine: Build-Tag Exile with a Time Box, No Blind Retries | Accepted |

## Error codes minted

| Code | Name | Exit code | Telemetry event |
|---|---|---|---|
| E-TEST-001 | Fixture integrity failure | 9 | `test.fixture.failed` |
| E-TEST-002 | Recorded interaction replay divergence | 1 | `test.replay.diverged` |
| E-TEST-003 | Scenario script rejected | 3 | `test.scenario.rejected` |
| E-TEST-004 | Hermeticity violation | 1 | `test.hermeticity.violated` |
| E-TEST-005 | Qualification evidence incomplete | 9 | `test.evidence.rejected` |
| E-TEST-006 | Gate evaluation failure | 1 | `test.gate.errored` |

## Events minted

All use the Volume 10 event envelope; payload summaries in chapter 04.

| Event | Emitted when |
|---|---|
| `test.suite.completed` | A classified suite finishes in any tier |
| `test.gate.evaluated` | A gate computes a result |
| `test.qualification.completed` | Qualification stage S6 finishes for a candidate |
| `test.flake.quarantined` | A quarantine change lands |
| `test.fixture.failed` | E-TEST-001 raised |
| `test.replay.diverged` | E-TEST-002 raised |
| `test.scenario.rejected` | E-TEST-003 raised |
| `test.hermeticity.violated` | E-TEST-004 raised |
| `test.evidence.rejected` | E-TEST-005 raised |
| `test.gate.errored` | E-TEST-006 raised |

## Config keys minted

None. Volume 13 owns no `andromeda.toml` table (Volume 0, chapter 03 configuration-table
ownership); test and CI behavior is configured in repository tooling and pipeline
definitions (Volume 11), not in product configuration.

## Glossary additions

| Term | One-line meaning |
|---|---|
| Contract test kit | The reusable per-port test package that any implementation (real adapter or fake) must pass (FR-TEST-004). |
| Scenario script | A versioned, deterministic script driving the fake provider: content, chunk pacing, tool calls, injected errors, usage figures (ADR-176). |
| Network sentinel | Test-build instrumentation at the dialer seam that records and fails on outbound connection attempts in hermetic suites (FR-TEST-005). |
| Quarantine (test) | Exclusion of a flaky test from gating suites via the `quarantine` build tag with a dated, issue-linked comment and a 14-day time box (ADR-177). |
| Mutation score | Killed mutants ÷ (killed + surviving mutants) after recorded equivalent-mutant exclusions (ADR-175). |
| Qualification evidence bundle | The schema-validated JSON record of everything that qualified a release candidate, retained with the release (FR-TEST-009). |
| Gate tier | One of T0 (merge) / T1 (trunk) / T2 (scheduled) / T3 (release qualification) / T4 (phase gate), fixing when a suite runs and what it blocks (Volume 13, chapter 01). |
| Golden file | A versioned expected-output file compared byte-exactly, updated only via an explicit flag recorded in the diff (Volume 13, chapter 02). |

## Assumptions

Local list per Volume 0, chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | CI runners (Volume 11 pipelines) can run Linux network namespaces for OS-level offline isolation | First implementation of the offline harness on the chosen runners | Offline isolation on Linux falls back to sentinel-only with a recorded finding; a self-hosted or containerized lane is added |
| Technical assumption | A local model server (Ollama with a pinned model) is runnable within CI resource limits for the offline and local-conformance lanes | First SM-04/SM-05 lane bring-up | Offline lane runs against the local fake serving path for Andromeda-side checks; SM-04 model-quality runs move to reference hardware outside hosted CI |
| Technical assumption | Timed generative budgets (fuzz ≥ 15 min/target at T2) fit scheduled-lane capacity | Lane runtime observation over the first month | Budgets are re-balanced across the week per target set, keeping per-target monthly totals |

## Open questions

| Question / pending item | Marked as | Blocking? | Resolution path |
|---|---|---|---|
| Concrete mutation-testing tool for Go meeting the ADR-175 selection criteria | PENDING VALIDATION (ADR-175; NFR-TEST-004) | No — gate definition, score formula, and fallback are fixed; evaluation before Beta entry | Evaluate candidate tools against the fixed criteria; amend ADR-175 with the selected, pinned tool |
| OS-level network isolation mechanism for the offline suite on macOS CI runners | PENDING VALIDATION (FR-TEST-005) | No — Linux isolation plus the platform-neutral sentinel carry MVP; macOS lane runs sentinel-only until validated | Validate a documented macOS isolation mechanism (or dedicated runner configuration) before Beta; update FR-TEST-005's compatibility notes |

## Cross-volume references

- Volume 1: SM-14 formalized here (NFR-TEST-003); suite mechanics defined here for SM-04,
  SM-05, SM-10, SM-11, SM-15, SM-16, SM-17, SM-18, SM-19 while their NFR formalization stays
  with the volumes named in Volume 1, chapter 06; MVP minimum items 24–25; UC-01/UC-09
  journeys.
- Volume 3: port freeze (FR-ARCH-003) and cancellation contract (FR-ARCH-004) enforced by
  FR-TEST-004 kits; NFR-ARCH-002 contract-diff evidence flows into the chapter 04 bundle.
- Volume 5: provider contract and adapter declarations verified by provider contract tests;
  SM-04 suite content.
- Volume 6: tool contract, MCP conformance suite content, SDK distribution of contract kits
  and doubles.
- Volume 9: permission/sandbox suite semantics; redaction rules asserted by security tests.
- Volume 10: event envelope for all `test.*` events; storage/retention of evidence bundles.
- Volume 11: CI pipeline definitions, secret scanning, traceability automation consuming the
  FR-TEST-002 map; error-catalog consistency check placement (ADR-016).
- Volume 12: performance/load/stress/soak budgets and reference environments evaluated at S4.
- Volume 14: release pipeline hand-off, publication refusal without a qualified bundle,
  upgrade/rollback semantics under test.
