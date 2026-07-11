# Volume 12 — Performance and Reliability

**Status:** Complete · **Owner:** Performance and Reliability (Volume 12)

Volume 12 turns the product-level performance and reliability commitments of Volume 1 into
verifiable engineering budgets. It formalizes the success metrics assigned to this volume
(SM-05 through SM-11, Volume 1, chapter 06) as `NFR-PERF-*` requirements with concrete
percentiles, reference hardware, datasets, thresholds, measurement methods, and phases; it
defines the runtime's timeout, backpressure, degradation, and recovery behavior; and it
specifies the benchmark suite and the operational limits the implementation enforces. Per
Volume 0, chapter 03, this volume mints all `PERF` identifiers and owns the pool sizes and
saturation budgets that Volume 3, chapter 08 delegates to it.

Foundations assumed: Volume 0 (conventions, templates, ID taxonomy), Volume 1 (objectives,
success metrics, reference conditions, phases), Volume 2 (entities and canonical states),
Volume 3 (ports, components, process/concurrency model).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Performance Targets](01-performance-targets.md) | Reference machines RM-1/RM-2 and reference datasets, measurement methodology, and NFR-PERF-001 through NFR-PERF-023: cold/warm start, TUI render and input latency, first token, streaming, tool dispatch, filesystem scan, git status, indexing, memory retrieval, search, patch generation, diff rendering, session restore, RAM, CPU, disk, concurrency, large repositories, large files, long sessions |
| [02 — Reliability and Degradation](02-reliability-and-degradation.md) | Availability model, timeout baseline, backpressure and overload shedding, degraded-mode catalog, resource watchdog and recovery objectives; NFR-PERF-024 through NFR-PERF-028 (offline operation, tool-call reliability, session recovery, crash-free operation, degradation responsiveness) |
| [03 — Benchmarks and Operational Limits](03-benchmarks-and-operational-limits.md) | Benchmark suite definition, baselines and regression gating, benchmark inventory; operational limits tables including scheduler pool budgets; E-PERF error codes |
| [99 — Volume Register](99-volume-register.md) | Everything this volume minted: requirements, ADRs, error codes, events, glossary additions, assumptions, open questions |

## Reading guide

1. Chapter 01 is the normative target set: every latency, throughput, and resource budget the
   product commits to, each bound to a reference machine and dataset defined there.
2. Chapter 02 defines how the runtime behaves when budgets cannot be met: bounded queues,
   finite deadlines, explicit degraded modes, and recovery objectives. Degradation is always
   evented and visible, never silent ([ADR-162](../annexes/adr/ADR-162.md)).
3. Chapter 03 defines how compliance is proven: the benchmark suite that measures every NFR in
   chapter 01, the regression gate ([ADR-161](../annexes/adr/ADR-161.md)), and the hard limits
   the runtime enforces at the boundaries of its budgets.
