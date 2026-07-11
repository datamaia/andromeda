# 99 — Volume 12 Register

Machine-parseable register of everything Volume 12 minted, per Volume 0 chapters 02 and 03.
Merged into the Volume 0 registers at consolidation.

## Requirements index

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-PERF-001 | Deadline and timeout baseline | MVP | Fault-injection suite; static audit of external call sites |
| FR-PERF-002 | Backpressure and overload shedding | MVP | Saturation fixtures; flooding tests; load scenario in benchmark suite |
| FR-PERF-003 | Degraded operation modes | MVP | Per-mode fault-injection conformance suite; offline suite; event-pairing audit |
| FR-PERF-004 | Resource watchdog and recovery objectives | MVP | Disk-fill/memory-pressure fixtures; restart-time measurement; watchdog-failure fixture |
| FR-PERF-005 | Benchmark suite and regression gating | MVP | Gate-evaluator self-test; CI configuration review; release artifact audit |
| FR-PERF-006 | Operational limits enforcement | MVP | Per-limit conformance fixtures; hot-path overhead benchmarks |
| NFR-PERF-001 | CLI cold start | v1 | Benchmark suite (operation layer), release gate |
| NFR-PERF-002 | CLI warm start | v1 | Benchmark suite (operation layer), release gate |
| NFR-PERF-003 | TUI startup to interactive | v1 | Benchmark suite (operation layer), release gate |
| NFR-PERF-004 | TUI frame render time | v1 | Instrumented render benchmark, release gate |
| NFR-PERF-005 | TUI input latency | v1 | Scripted interaction replay, release gate |
| NFR-PERF-006 | First-token overhead | v1 | Instrumented turn benchmark with mock provider, release gate |
| NFR-PERF-007 | Streaming update overhead | v1 | Mock streaming provider benchmark, release gate |
| NFR-PERF-008 | Tool dispatch overhead | v1 | Tool Runtime micro-benchmark, release gate |
| NFR-PERF-009 | Filesystem scan | v1 | Workspace scan benchmark on DS-M/DS-L, release gate |
| NFR-PERF-010 | Git status latency | v1 | Paired git benchmark, release gate |
| NFR-PERF-011 | Indexing throughput | v1 | IndexerPort benchmarks, release gate |
| NFR-PERF-012 | Memory retrieval latency | v1 | Memory Manager benchmark over DS-MEM, release gate |
| NFR-PERF-013 | Search latency | v1 | Index query benchmarks, release gate |
| NFR-PERF-014 | Patch generation | v1 | Patch benchmark with apply-clean assertion, release gate |
| NFR-PERF-015 | Diff rendering | v1 | Instrumented TUI diff benchmark, release gate |
| NFR-PERF-016 | Session restore | v1 | Crash-injection restore benchmark, release gate |
| NFR-PERF-017 | Memory budget | v1 | Process accounting (idle and soak), release gate |
| NFR-PERF-018 | CPU budget | v1 | Process accounting (idle and streaming), release gate |
| NFR-PERF-019 | Disk budget | v1 | Filesystem accounting after install/soak/build/shutdown, release gate |
| NFR-PERF-020 | Concurrency capacity and scheduler overhead | v1 | Load scenario plus scheduler micro-benchmark, release gate |
| NFR-PERF-021 | Large-repository operation | v1 | DS-L suite re-run, release gate |
| NFR-PERF-022 | Large-file handling | v1 | DS-F benchmarks with process accounting, release gate |
| NFR-PERF-023 | Long-session stability | v1 | DS-SOAK replay with hourly checkpoints and leak gates |
| NFR-PERF-024 | Offline operation | MVP | Offline test suite (Volume 13), MVP exit gate |
| NFR-PERF-025 | Tool-call reliability | v1 | Suite outcome classification plus fault injection |
| NFR-PERF-026 | Session recovery success | v1 | Crash-injection suite, ≥ 200 injections per release |
| NFR-PERF-027 | Crash-free operation | v1 | Soak/E2E crash accounting; panic-injection fixtures |
| NFR-PERF-028 | Degradation responsiveness | Beta | Instrumented per-mode fault injection |
| RISK-PERF-001 | Performance targets unattainable on reference hardware | — | Risk register review at phase gates; nightly trend monitoring |
| RISK-PERF-002 | Benchmark environment variance masks or fakes regressions | — | Risk register review at phase gates; calibration failure monitoring |
| RISK-PERF-003 | Degradation matrix complexity outgrows testing | — | Risk register review at phase gates; event-pairing audits |
| RISK-PERF-004 | Real workspaces exceed declared limits | — | Risk register review at phase gates; enforcement-counter diagnostics |

Local (non-corpus) identifier series used by this volume: `RM-N` (reference machines) and
`DS-*` (reference datasets), both defined in chapter 01; degraded-mode identifiers
(`offline`, `provider_limited`, `index_unavailable`, `no_embeddings`, `secret_fallback`,
`low_disk`, `low_memory`, `sandbox_reduced`, chapter 02); and operational-limit identifiers
(chapter 03 tables). These are chapter-local labels, not corpus identifiers; other volumes
reference them by naming the volume and chapter.

## ADRs minted

| ADR | Title | Status |
|---|---|---|
| [ADR-160](../annexes/adr/ADR-160.md) | Reference machines, reference datasets, and measurement methodology for performance verification | Accepted |
| [ADR-161](../annexes/adr/ADR-161.md) | Benchmark tooling and regression gating policy | Accepted |
| [ADR-162](../annexes/adr/ADR-162.md) | Closed, evented degraded-mode catalog; no silent degradation | Accepted |

ADR numbers 163–174 of this volume's block are unallocated and remain permanent gaps unless
used by later amendments of this volume.

## Error codes minted

| Code | Name | Severity | Exit code |
|---|---|---|---|
| E-PERF-001 | Performance budget exceeded | Warning | none (diagnostic) |
| E-PERF-002 | Operational limit exceeded | Error | 1 (foreground); 6 (tool surface) |
| E-PERF-003 | Resource exhaustion refusal | Error | 1 (foreground) |
| E-PERF-004 | Benchmark regression gate failure | Error | 1 (harness process) |

## Events minted

Envelope, delivery, persistence, and retention per Volume 10 (referenced, not restated).

| Event | Emitted when |
|---|---|
| `perf.budget.exceeded` | Instrumented operation exceeds its declared budget by > 2× in a running instance |
| `perf.limit.enforced` | Any operational-limit enforcement (refuse, truncate, narrow, shed) |
| `perf.overload.shed` | Load-shedding action under sustained saturation |
| `perf.degradation.entered` | Degraded-mode entry |
| `perf.degradation.exited` | Degraded-mode exit |
| `perf.benchmark.completed` | Benchmark run finished with stored results |
| `perf.benchmark.regressed` | Relative regression beyond the warning or failure band |

## Config keys minted

None. Volume 12 owns no TOML table (Volume 0, chapter 03). Operational-limit and pool values
in chapter 03 are normative defaults; where a limit is user-configurable, the owning area
mints its key in its own table (owners named per row in the chapter 03 tables) with this
volume's value as the default. Limits without a named key are fixed at MVP.

## Glossary additions

| Term | One-line meaning |
|---|---|
| Reference machine (RM-1, RM-2) | One of the two canonical test machines of Volume 12, chapter 01, with a ±10% calibration equivalence rule (ADR-160). |
| Reference dataset (DS-*) | A deterministic, seed-generated benchmark corpus defined in Volume 12, chapter 01 (DS-M, DS-L, DS-F, DS-MEM, DS-SOAK). |
| Mock provider | A ProviderPort test double emitting scripted responses and stream chunks with configurable pacing, used to isolate Andromeda-added overhead from model inference time. |
| Degraded mode | A named, evented, queryable reduced-service condition from the closed catalog of Volume 12, chapter 02 (ADR-162). |
| Operational limit | A normative bound from Volume 12, chapter 03 enforced by one of four declared enforcement classes, always with an enforcement event. |
| Enforcement class | One of refuse, truncate, virtualize/narrow, shed/queue — the closed set of behaviors at an operational limit. |
| Rolling baseline | The median of a benchmark's last 5 mainline nightly results on the same runner class, used by the relative regression gate (ADR-161). |
| Resource watchdog | The runtime sampler of disk, memory, database size, and saturation that trips degraded modes and pre-run refusals (Volume 12, chapter 02). |

## Assumptions

Local list per Volume 0, chapter 05 (global `AS-NNN`/`HY-NNN` numbers are minted at
consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | A Go implementation over the pure-Go SQLite driver (ADR-007) meets the chapter 01 hot-path budgets (first-token overhead, session-restore, memory-retrieval latency) on RM-1 and RM-2 | Pre-MVP benchmark spike per the Volume 1 assumption's validation path; nightly trends from MVP | ADR-007's recorded fallback (CGO driver evaluation) and, in the limit, the ADR-001 reversal plan; affected NFR thresholds revisited through the change procedure |
| Technical assumption | At least one CI runner class can hold the ADR-160 ±10% calibration stably enough for gating | Calibration telemetry during CI bring-up (open question OQ-V12-2) | Dedicated benchmark runner is provisioned; gate cadence may drop to scheduled runs on that hardware |
| Technical assumption | Mock-provider pacing profiles are representative of real provider stream shapes for overhead measurement | Non-gated comparison of overhead distributions against instrumented real-provider sessions at each phase gate | Pacing profiles are recalibrated from recorded real traces; affected benchmarks reset baselines explicitly |

## Open questions

Entries follow Volume 0, chapter 08; global `OQ-NNN` numbers are minted at consolidation.
Items marked PENDING VALIDATION below are the register entries for every `PENDING
VALIDATION` occurrence in this volume.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| OQ-V12-1 | Tool-dispatch overhead budgets for OS-level sandbox tiers (PENDING VALIDATION: ADR-021 mechanism selection per platform) | Chapter 01, NFR-PERF-008 | No — MVP process-level budgets are defined; OS-level tiers are Beta/v1 | Volume 9 selects mechanisms per ADR-021; this volume then mints the tier budgets as an amendment to NFR-PERF-008 | Open |
| OQ-V12-2 | Whether GitHub-hosted CI runners can hold ADR-160 calibration for gating, or a dedicated benchmark runner is required (PENDING VALIDATION: runner hardware stability) | Chapter 03, RISK-PERF-002; ADR-161 | No — absolute thresholds gate on any calibrated machine; only gate placement is open | Measure calibration stability across runner samples during CI bring-up; decide runner strategy and record it | Open |
| OQ-V12-3 | OS page-cache purge mechanics for cold-start measurement on each platform and on CI (privileged operations may be unavailable) | Chapter 01, measurement methodology rule 2 | No — warm-start measurement and Andromeda-cache purging are unaffected | Harness implementation validates per-platform purge; if infeasible in CI, cold-start gating runs on the calibrated reference machines only | Open |

## Cross-volume references

Expectations this volume places on other volumes; consolidation verifies each.

| This volume defines | Volume expected to formalize / consume |
|---|---|
| Formalization of SM-05, SM-06, SM-07, SM-08, SM-09, SM-10, SM-11 as NFR-PERF-024, -001/-002/-003, -005, -007, -017, -025, -016/-026 | Volume 1 metric governance satisfied; Volume 15 phase gates consume the gate results |
| Scheduler pool sizes, queue bounds, and saturation policies (chapter 03) | Volume 3, chapter 08 delegates these values; Task Scheduler implements them |
| Timeout defaults table (chapter 02) | Volumes 5, 6, 8, 10, 11, 14 own per-operation semantics and any config keys; owner values prevail and reconcile at consolidation |
| Degraded-mode catalog and lifecycle events | Volume 8 presents indicators and doctor/status output; Volume 5 owns provider routing inside `provider_limited`; Volume 9 owns policy of `secret_fallback` and `sandbox_reduced`; Volume 10 carries the events |
| Operational limits with configurable-key ownership per row | Volumes 4, 6, 7, 8, 10 mint the named keys in their tables with these defaults |
| Benchmark suite and regression gate | Volume 11 hosts the CI pipelines; Volume 13 integrates suite execution into quality gates; Volume 14 blocks releases on the gate report |
| Offline-operation NFR (NFR-PERF-024) | Volume 13 specifies and executes the offline suite (SM-05 method) |
| Crash-injection populations for recovery metrics | Volume 13 owns the suites; Volume 4 owns resumption semantics under test |
| `perf.*` event names and payload summaries | Volume 10 envelope; consolidated event catalog |
