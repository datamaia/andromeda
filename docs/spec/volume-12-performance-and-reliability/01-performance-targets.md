# 01 — Performance Targets

This chapter formalizes Andromeda's performance commitments as non-functional requirements.
It is the single normative home of the product's latency, throughput, and resource budgets.
The targets bound to Volume 1 success metrics (SM-06 startup, SM-07 interaction latency,
SM-08 streaming latency, SM-09 memory usage; Volume 1, chapter 06) preserve or strengthen
those product-level values — per Volume 1 metric governance, this volume MUST NOT weaken
them. Per that governance, targets are measured from MVP onward and gate at the phase named
in each requirement; measurement without gating starts as soon as the benchmark suite
(chapter [03](03-benchmarks-and-operational-limits.md)) exists.

## Reference machines

Volume 1 defines product-level reference conditions; this volume owns the formal test
environments derived from them ([ADR-160](../annexes/adr/ADR-160.md)). Two reference
machines are defined. Every NFR in this volume binds to one or both.

| ID | Canonical machine | Definition |
|---|---|---|
| RM-1 | MacBook Air, Apple M2, 8-core CPU, 16 GB unified memory, 512 GB internal NVMe SSD | macOS reference machine: Apple Silicon laptop class, 16 GB unified memory, internal NVMe SSD, running the minimum supported macOS version per the Volume 3 platform matrix, on mains power, no thermal throttling active at measurement start |
| RM-2 | Virtual machine: 4 vCPU on a current x86_64 server CPU generation, 8 GB RAM, SSD-class storage | Linux reference machine: 4 vCPU / 8 GB x86_64 VM, storage sustaining ≥ 500 MB/s sequential read, reference distribution and kernel per the Volume 3 platform matrix |

Equivalence rule: a physical or virtual machine qualifies as an instance of a reference
machine when its single-core CPU score and sequential SSD read throughput are within ±10% of
the canonical machine's, measured by the calibration step of the benchmark harness. The
harness records the machine fingerprint (CPU model, memory, storage class, OS version) with
every result set.

Cross-machine rule: unless a requirement states per-machine values, its target and minimum
threshold apply on **both** reference machines. Reference network conditions, where a
measurement involves the network at all, are Volume 1's (50 Mbit/s downstream, 40 ms RTT);
the latency and resource budgets in this chapter never depend on network conditions because
provider time is excluded by method.

## Reference datasets

All datasets are generated deterministically from pinned seeds by the fixture generator in
the benchmark suite (chapter 03), so results are reproducible and no large blobs live in the
repository.

| ID | Dataset | Definition |
|---|---|---|
| DS-M | Reference workspace | ≈ 5,000 files, ≈ 1,000,000 lines of mixed-language text, Git history ≥ 5,000 commits — the Volume 1 reference repository made concrete |
| DS-L | Large workspace | ≥ 100,000 files, ≥ 10,000,000 lines, Git history ≥ 50,000 commits |
| DS-F | File-size series | Text files of 1 MiB, 10 MiB, and 100 MiB; one 1 GiB binary file |
| DS-MEM | Memory corpus | 100,000 active Memory Records, size distribution median 1 KiB / p99 32 KiB, mixed layers and provenance |
| DS-SOAK | Long-session script | Deterministic 8-hour scripted session against DS-M: ≥ 500 turns via the mock provider, ≥ 2,000 tool invocations (filesystem, git, terminal), periodic diff reviews and index updates |

## Measurement methodology

The following rules apply to every requirement in this volume; each requirement's
measurement method refines, never contradicts, them.

1. **Percentiles.** A stated percentile (p95, p99) is computed over the iteration population
   of one benchmark run: ≥ 50 iterations for operation-level benchmarks (matching the SM-06
   method), ≥ 1,000 for micro-benchmarks (overhead per task, per chunk). Warm-up iterations
   are excluded and their count is recorded.
2. **Cold and warm.** A *cold start* is the first invocation after the harness has purged
   Andromeda's own caches and the OS page cache for the binary and databases (platform purge
   mechanics are a harness concern; feasibility on CI runners is tracked as an open
   question). A *warm start* is any subsequent invocation in the same iteration sequence.
3. **Provider isolation.** Any measurement crossing ProviderPort uses the mock provider: a
   test double emitting scripted responses and stream chunks with configurable pacing. This
   isolates Andromeda-added overhead from model inference time, which Andromeda does not
   control (SM-08 method).
4. **Exclusions.** Resident set size and CPU budgets cover the Andromeda process only —
   child processes it supervises (tools, plugins, MCP servers, git) and any local model
   server are accounted separately by the same harness but are not counted against the main
   process budgets, matching SM-09's definition.
5. **Clocks.** All intervals use the platform monotonic clock; TUI intervals are measured
   from instrumented event timestamps (input event enqueued → frame flushed to the
   terminal), the SM-07 method.
6. **Records.** Every benchmark run records machine fingerprint, dataset IDs, product
   version, and raw samples, and emits `perf.benchmark.completed` (envelope per Volume 10).

```mermaid
flowchart LR
    A[user submit] --> B[context assembly]
    B --> C[provider request written]
    C --> D[first provider chunk received]
    D --> E[chunk rendered or emitted]
    E --> F[tool call decided]
    F --> G[tool Execute begins]
```

**Prose for the diagram.** The flowchart marks the instrumented measurement points on one
interactive turn. NFR-PERF-006 (first-token overhead) measures A→C: everything Andromeda adds
before the provider can start working, including context assembly. The provider's own
inference time, C→D, is excluded everywhere by the provider-isolation rule. NFR-PERF-007
(streaming update) measures D→E per chunk. NFR-PERF-008 (tool dispatch) measures F→G:
validation, non-interactive permission evaluation, sandbox preparation, and event emission,
excluding any human approval wait and the tool's own execution. The constraint the diagram
encodes: Andromeda's budgets cover exactly the segments Andromeda controls.

## Target overview

The table maps every mandated target area to its requirement. Values below are summaries;
the requirement text is normative.

| Area | Requirement | Headline target (p95 unless noted) |
|---|---|---|
| Cold start | NFR-PERF-001 | CLI version ≤ 150 ms |
| Warm start | NFR-PERF-002 | CLI version ≤ 80 ms |
| TUI start | NFR-PERF-003 | Interactive ≤ 400 ms |
| TUI render | NFR-PERF-004 | Full frame ≤ 20 ms |
| Input latency | NFR-PERF-005 | Input→render ≤ 40 ms |
| First token | NFR-PERF-006 | Added overhead ≤ 200 ms |
| Streaming update | NFR-PERF-007 | Added overhead ≤ 30 ms/chunk |
| Tool dispatch | NFR-PERF-008 | In-process ≤ 20 ms; subprocess ≤ 100 ms |
| Filesystem scan | NFR-PERF-009 | DS-M ≤ 500 ms; DS-L ≤ 10 s |
| Git status | NFR-PERF-010 | DS-M ≤ 300 ms; overhead ≤ 30 ms |
| Indexing | NFR-PERF-011 | DS-M lexical build ≤ 45 s; incremental ≤ 300 ms |
| Memory retrieval | NFR-PERF-012 | Retrieve ≤ 75 ms over DS-MEM |
| Search | NFR-PERF-013 | Lexical ≤ 100 ms; semantic ≤ 300 ms |
| Patch generation | NFR-PERF-014 | 10-file patch ≤ 150 ms |
| Diff rendering | NFR-PERF-015 | 1,000-line diff ≤ 80 ms |
| Session restore | NFR-PERF-016 | Resume ≤ 3 s |
| RAM | NFR-PERF-017 | Idle ≤ 250 MB; 8 h ≤ 500 MB |
| CPU | NFR-PERF-018 | Idle ≤ 0.5% core avg; streaming ≤ 35% core |
| Disk | NFR-PERF-019 | Install ≤ 120 MB; cache ≤ 1.0× text |
| Concurrency | NFR-PERF-020 | 4 runs / 16 tool invocations on RM-1 |
| Large repositories | NFR-PERF-021 | DS-L open ≤ 2 s; interactive budgets hold |
| Large files | NFR-PERF-022 | 100 MiB read streamed, RSS delta ≤ 32 MiB |
| Long sessions | NFR-PERF-023 | Hour-8 latency ≤ 1.1× hour-1; zero leaks |

## Requirements

### NFR-PERF-001 — CLI cold start

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: Wall-clock time from CLI process start to first byte of stdout for the version command, cold start, p95 over ≥ 50 iterations
- Target: ≤ 150 ms p95 on RM-1 and RM-2
- Minimum threshold: ≤ 200 ms p95 (SM-06(a); weakening requires the Volume 0 change procedure)
- Measurement method: Benchmark harness spawns the real binary, timestamps process start and first stdout byte, purges caches per the cold-start rule between iterations; 50 iterations per platform per release
- Test environment: RM-1 and RM-2, DS-M present but not required by the command
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / CLI (Volume 8) for the measured surface
- Dependencies: FR-CLI-001; ADR-005
- Risks: RISK-PERF-001
- Acceptance criteria: Given the benchmark harness on either reference machine, when the cold-start benchmark runs, then p95 ≤ 200 ms (release gate) and the trend against the 150 ms target is reported; a p95 above threshold blocks the release.

### NFR-PERF-002 — CLI warm start

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: Wall-clock time from CLI process start to first byte of stdout for the version command, warm start, p95 over ≥ 50 iterations
- Target: ≤ 80 ms p95 on RM-1 and RM-2
- Minimum threshold: ≤ 120 ms p95
- Measurement method: Same harness as NFR-PERF-001 without cache purging; iterations 2..N of each sequence
- Test environment: RM-1 and RM-2
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: NFR-PERF-001
- Risks: RISK-PERF-001
- Acceptance criteria: Given a warmed environment, when the version command runs 50 times, then p95 ≤ 120 ms and warm p95 is strictly below the same run's cold p95.

### NFR-PERF-003 — TUI startup to interactive

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: Wall-clock time from TUI process start to interactive prompt — first full frame rendered and input accepted — in an indexed DS-M workspace, warm start, p95 over ≥ 50 iterations
- Target: ≤ 400 ms p95 on RM-1 and RM-2
- Minimum threshold: ≤ 500 ms p95 (SM-06(b))
- Measurement method: Instrumented TUI timestamps (process start, first frame flush, input-loop ready) under the scripted replay harness; the workspace index is `ready` before measurement
- Test environment: RM-1 and RM-2, DS-M with a ready lexical index
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / TUI (Volume 8) for the measured surface
- Dependencies: FR-TUI-001; ADR-006
- Risks: RISK-PERF-001
- Acceptance criteria: Given DS-M with a ready index, when the TUI starts 50 times, then p95 to interactive ≤ 500 ms (gate) with the 400 ms target trend reported; negative case: with a `building` index, startup does not block on indexing and still meets the threshold, showing index status honestly.

### NFR-PERF-004 — TUI frame render time

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: Time to compose and flush one full-screen frame at 120×40, truecolor tier, during active streaming, p95 over ≥ 1,000 frames
- Target: ≤ 20 ms p95
- Minimum threshold: ≤ 33 ms p95 (sustains ≥ 30 rendered frames per second under load)
- Measurement method: Instrumented render timestamps (frame build start → terminal write completed) during the DS-SOAK streaming segments, replayed with the mock provider
- Test environment: RM-1 and RM-2, DS-M, terminal 120×40 truecolor
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / TUI (Volume 8)
- Dependencies: FR-TUI-001; NFR-PERF-007
- Risks: RISK-PERF-001
- Acceptance criteria: Given active streaming at the mock provider's reference pacing, when 1,000 frames render, then p95 ≤ 33 ms and no frame exceeds 100 ms (recorded as p100); at 80×24 the same thresholds hold.

### NFR-PERF-005 — TUI input latency

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: Keystroke or navigation action to completed screen update, excluding model inference time, p95 over a scripted interaction replay of ≥ 1,000 inputs
- Target: ≤ 40 ms p95
- Minimum threshold: ≤ 50 ms p95 (SM-07)
- Measurement method: Instrumented TUI event timestamps (input event enqueued → resulting frame flushed) under scripted interaction replay per the SM-07 method, on DS-M and repeated on DS-L
- Test environment: RM-1 and RM-2, DS-M and DS-L
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / TUI (Volume 8)
- Dependencies: FR-TUI-001; NFR-PERF-004
- Risks: RISK-PERF-001
- Acceptance criteria: Given the scripted replay, when 1,000 inputs execute on DS-M, then p95 ≤ 50 ms; given DS-L, then the same threshold holds (NFR-PERF-021(c)); Volume 8's activity-indicator rule covers any operation legitimately exceeding it.

### NFR-PERF-006 — First-token overhead

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: Andromeda-added time from user submit to the provider request's first byte written to the transport — including context assembly with a warm index and memory store — p95 over ≥ 50 turns
- Target: ≤ 200 ms p95
- Minimum threshold: ≤ 300 ms p95
- Measurement method: Instrumented timestamps at submit and at request write, mock provider, DS-M with ready index and DS-MEM loaded; context assembly budget per Volume 7's contract is included in the measured interval
- Test environment: RM-1 and RM-2, DS-M, DS-MEM
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Context Manager (Volume 7) for the dominant interval
- Dependencies: FR-CTX-001; FR-AGT-001
- Risks: RISK-PERF-001
- Acceptance criteria: Given a warm workspace, when 50 turns are submitted, then p95 submit-to-request-write ≤ 300 ms; negative case: with a cold index the request still dispatches (degraded retrieval per Volume 7) and the interval is recorded separately, not hidden in the gated population.

### NFR-PERF-007 — Streaming update overhead

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: Andromeda-added time per streamed chunk from provider chunk receipt to rendered output (TUI) or emitted structured output (CLI), p95 over ≥ 1,000 chunks
- Target: ≤ 30 ms p95
- Minimum threshold: ≤ 50 ms p95 (SM-08)
- Measurement method: Instrumented timestamps with the mock streaming provider at reference pacing (20 chunks/s) and at stress pacing (200 chunks/s), per the SM-08 method; both TUI and CLI `--json` surfaces measured
- Test environment: RM-1 and RM-2, DS-M
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: FR-PROV-001; FR-TUI-001; FR-CLI-001
- Risks: RISK-PERF-001
- Acceptance criteria: Given reference pacing, when 1,000 chunks stream, then p95 added overhead ≤ 50 ms on both surfaces; given stress pacing at 200 chunks/s, then rendering coalesces per frame without unbounded queue growth and the TUI stays within NFR-PERF-004.

### NFR-PERF-008 — Tool dispatch overhead

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: Time from the agent's tool-call decision to the tool's `Execute` entry — input validation, non-interactive permission evaluation, sandbox preparation, event emission — excluding approval wait and tool execution, p95 over ≥ 1,000 dispatches
- Target: ≤ 20 ms p95 for in-process built-in tools; ≤ 100 ms p95 for subprocess dispatch under MVP process-level sandboxing
- Minimum threshold: ≤ 50 ms p95 in-process; ≤ 200 ms p95 subprocess
- Measurement method: Micro-benchmark through the Tool Runtime with a no-op built-in tool and a no-op subprocess tool; permission decisions pre-resolved by policy so no interaction occurs; budgets for OS-level sandbox tiers are PENDING VALIDATION until the ADR-021 mechanism per platform is selected (open question in the [volume register](99-volume-register.md))
- Test environment: RM-1 and RM-2, DS-M
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Tool Runtime (Volume 6)
- Dependencies: FR-TOOL-001; FR-SEC-100; FR-SEC-101; ADR-021
- Risks: RISK-PERF-001
- Acceptance criteria: Given pre-resolved permissions, when 1,000 no-op dispatches run, then in-process p95 ≤ 50 ms and subprocess p95 ≤ 200 ms; permission and observability case: every dispatch in the population carries its permission decision reference and emits its invocation events — dispatch overhead is measured with full mediation on, never bypassed.

### NFR-PERF-009 — Filesystem scan

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: Wall-clock time for a full ignore-rule-aware workspace walk (enumeration and metadata, no content read), p95 over ≥ 50 scans
- Target: ≤ 500 ms p95 on DS-M; ≤ 10 s p95 on DS-L
- Minimum threshold: ≤ 1 s p95 on DS-M; ≤ 20 s p95 on DS-L
- Measurement method: Benchmark harness invokes the Workspace Engine scan path with warm and cold OS caches recorded separately; gated population is warm
- Test environment: RM-1 and RM-2, DS-M and DS-L
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Workspace Engine (Volume 4)
- Dependencies: ADR-022
- Risks: RISK-PERF-001, RISK-PERF-004
- Acceptance criteria: Given DS-M, when 50 warm scans run, then p95 ≤ 1 s; given DS-L, then p95 ≤ 20 s and the scan runs on the `background` pool without breaching NFR-PERF-005 for a concurrently interactive session.

### NFR-PERF-010 — Git status latency

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: (a) End-to-end `GitPort.Status` wall-clock time; (b) Andromeda-added overhead versus raw `git status --porcelain=v2` on the same repository — both p95 over ≥ 50 invocations
- Target: (a) ≤ 300 ms p95 on DS-M, ≤ 3 s p95 on DS-L; (b) ≤ 30 ms p95
- Minimum threshold: (a) ≤ 500 ms p95 on DS-M, ≤ 6 s p95 on DS-L; (b) ≤ 75 ms p95
- Measurement method: Paired benchmark: raw system git invocation and `GitPort.Status` through the Git Engine (ADR-025) on identical repository state; overhead is the paired difference
- Test environment: RM-1 and RM-2, DS-M and DS-L with realistic dirty state (100 modified files)
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Git Engine (Volume 11)
- Dependencies: ADR-025
- Risks: RISK-PERF-001, RISK-PERF-004
- Acceptance criteria: Given DS-M with 100 modified files, when 50 status calls run, then end-to-end p95 ≤ 500 ms and paired overhead p95 ≤ 75 ms; the DS-L thresholds hold on the large dataset.

### NFR-PERF-011 — Indexing throughput

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: (a) Full lexical index build wall-clock time on DS-M; (b) incremental index update latency after one changed file, p95; (c) chunking-and-persistence throughput in chunks/s, excluding embedding computation
- Target: (a) ≤ 45 s; (b) ≤ 300 ms p95; (c) ≥ 1,000 chunks/s
- Minimum threshold: (a) ≤ 90 s; (b) ≤ 1 s p95; (c) ≥ 400 chunks/s
- Measurement method: IndexerPort `Build`/`Update` benchmarks on DS-M; builds run on the `background` pool as in production; embedding computation excluded by using the mock embedding fixture (ADR-020 storage path exercised for real)
- Test environment: RM-1 and RM-2, DS-M; DS-L build time reported non-gated under NFR-PERF-021
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Indexing Engine (Volume 7)
- Dependencies: FR-IDX-001; ADR-020
- Risks: RISK-PERF-001
- Acceptance criteria: Given DS-M, when a full lexical build runs, then it completes ≤ 90 s while a concurrent interactive session keeps NFR-PERF-005; given one file change, then the index returns to `ready` within 1 s p95; error case: a cancelled build leaves the prior generation queryable (Volume 7 integrity rules), verified in the same suite.

### NFR-PERF-012 — Memory retrieval latency

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: (a) `MemoryStorePort.Retrieve` latency for a lexical query over DS-MEM, p95; (b) `MemoryStorePort.Rank` latency for 256 candidates, p95 — each over ≥ 1,000 queries
- Target: (a) ≤ 75 ms p95; (b) ≤ 30 ms p95
- Minimum threshold: (a) ≤ 150 ms p95; (b) ≤ 75 ms p95
- Measurement method: Benchmark against the Memory Manager with DS-MEM loaded in the workspace database (ADR-028); query mix defined by the fixture script (by layer, scope, time, content)
- Test environment: RM-1 and RM-2, DS-MEM
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Memory Manager (Volume 7)
- Dependencies: FR-MEM-001; ADR-007, ADR-028
- Risks: RISK-PERF-001
- Acceptance criteria: Given DS-MEM, when the query mix runs 1,000 times, then Retrieve p95 ≤ 150 ms and Rank p95 ≤ 75 ms; observability case: retrieval latency is emitted as a metric so the Context Manager's budget accounting (Volume 7) can consume it.

### NFR-PERF-013 — Search latency

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: (a) Lexical `IndexerPort.Query` latency on a ready DS-M index, p95; (b) semantic query latency by exact cosine similarity over 100,000 chunks — the ADR-020 ceiling — excluding query-embedding computation, p95; each over ≥ 1,000 queries
- Target: (a) ≤ 100 ms p95; (b) ≤ 300 ms p95
- Minimum threshold: (a) ≤ 250 ms p95; (b) ≤ 600 ms p95
- Measurement method: Query benchmark with fixture query sets against ready indexes; semantic path uses precomputed query vectors so only Andromeda's scan-and-score work is measured
- Test environment: RM-1 and RM-2, DS-M indexed; semantic corpus filled to 100,000 chunks
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Indexing Engine (Volume 7)
- Dependencies: FR-IDX-001; ADR-020
- Risks: RISK-PERF-001
- Acceptance criteria: Given ready indexes at the stated sizes, when 1,000 queries of each kind run, then p95 ≤ 250 ms lexical and ≤ 600 ms semantic; negative case: a query against a `stale` index is answered with staleness flagged per Volume 7 and its latency is excluded from the gated population.

### NFR-PERF-014 — Patch generation

- Category: Performance
- Priority: P2
- Phase: v1
- Metric: Time to produce a reviewable Patch from recorded file changes: (a) 10 files / 500 changed lines; (b) 100 files / 10,000 changed lines — p95 over ≥ 50 generations
- Target: (a) ≤ 150 ms p95; (b) ≤ 1.5 s p95
- Minimum threshold: (a) ≤ 400 ms p95; (b) ≤ 3 s p95
- Measurement method: Benchmark drives the patch-generation path over fixture change sets on DS-M; output correctness is asserted (apply-clean check) so speed never trades against fidelity
- Test environment: RM-1 and RM-2, DS-M
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: ADR-025
- Risks: RISK-PERF-001
- Acceptance criteria: Given the fixture change sets, when 50 patches generate per size class, then p95 within thresholds and every generated patch applies cleanly to the pre-change tree.

### NFR-PERF-015 — Diff rendering

- Category: Performance
- Priority: P2
- Phase: v1
- Metric: (a) First full paint of a 1,000-line diff in the TUI diff view, p95; (b) first paint of a 100,000-line diff through virtualization — only visible hunks materialized — p95; each over ≥ 50 renders
- Target: (a) ≤ 80 ms p95; (b) ≤ 250 ms p95
- Minimum threshold: (a) ≤ 150 ms p95; (b) ≤ 500 ms p95
- Measurement method: Instrumented TUI render timestamps over fixture diffs; scrolling through the virtualized diff must sustain NFR-PERF-005 input latency, measured in the same run
- Test environment: RM-1 and RM-2, terminal 120×40
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / TUI (Volume 8)
- Dependencies: FR-TUI-001; NFR-PERF-005
- Risks: RISK-PERF-001
- Acceptance criteria: Given the fixture diffs, when each renders 50 times, then first-paint p95 within thresholds; given continuous scrolling of the 100,000-line diff, then input latency stays within NFR-PERF-005 and memory stays within NFR-PERF-017(a) plus 50 MB.

### NFR-PERF-016 — Session restore

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: (a) Time from restore command to interactive resumed session — recovery procedure, snapshot load, first frame — for an interrupted DS-SOAK session, p95; (b) `ListSessions` over 1,000 persisted sessions, p95
- Target: (a) ≤ 3 s p95; (b) ≤ 100 ms p95
- Minimum threshold: (a) ≤ 5 s p95 (SM-11(b)); (b) ≤ 250 ms p95
- Measurement method: Crash-injection harness (SM-11 method: kill −9 at randomized points) followed by timed restore; the interval includes FR-ARCH-009 recovery when the restore is the first start after the crash
- Test environment: RM-1 and RM-2, DS-M workspace with DS-SOAK-generated session history
- Measurement frequency: Every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Runtime (Volume 4 semantics, Volume 10 storage)
- Dependencies: FR-ARCH-009; ADR-007, ADR-028
- Risks: RISK-PERF-001
- Acceptance criteria: Given a session interrupted by kill −9, when restored, then the session is interactive ≤ 5 s p95 with zero persisted-turn loss (NFR-PERF-026 covers the success-rate half); given 1,000 sessions, listing completes ≤ 250 ms p95.

### NFR-PERF-017 — Memory budget

- Category: Performance
- Priority: P0
- Phase: v1
- Metric: (a) Resident set size of an idle interactive session in DS-M, excluding child processes and any local model server; (b) RSS at the end of the DS-SOAK 8-hour scripted session
- Target: (a) ≤ 250 MB; (b) ≤ 500 MB
- Minimum threshold: (a) ≤ 300 MB (SM-09(a)); (b) ≤ 600 MB (SM-09(b))
- Measurement method: Process accounting in the benchmark harness per the SM-09 method; idle sampled after 5 minutes of no activity with watchers running; soak sampled at hourly checkpoints
- Test environment: RM-1 and RM-2, DS-M, DS-SOAK
- Measurement frequency: Idle nightly; soak per release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: FR-ARCH-006; ADR-023
- Risks: RISK-PERF-001
- Acceptance criteria: Given an idle session, when sampled, then RSS ≤ 300 MB; given the completed soak, then RSS ≤ 600 MB and hourly checkpoints show no unbounded monotonic growth (each hour-over-hour delta ≤ 40 MB after hour 2).

### NFR-PERF-018 — CPU budget

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: (a) Average CPU of the idle Andromeda process (no active run, watchers active) over a 5-minute window, as a fraction of one core; (b) CPU of the main process during an active streaming turn at reference pacing, excluding children, p95 of 1-second samples
- Target: (a) ≤ 0.5%; (b) ≤ 35% of one core
- Minimum threshold: (a) ≤ 1%; (b) ≤ 60% of one core
- Measurement method: Process accounting during idle and DS-SOAK streaming segments; children and model servers measured separately per the exclusion rule
- Test environment: RM-1 and RM-2, DS-M
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: ADR-023
- Risks: RISK-PERF-001
- Acceptance criteria: Given an idle session, when sampled for 5 minutes, then average CPU ≤ 1% of one core — idle polling loops that busy-wait are defects; given streaming at reference pacing, then main-process CPU p95 ≤ 60% of one core.

### NFR-PERF-019 — Disk budget

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: (a) Installed product footprint (binary plus bundled assets); (b) workspace state growth (`.andromeda/` excluding index cache) after the DS-SOAK session under default retention; (c) index cache size as a multiple of indexed text bytes on DS-M; (d) temporary-file residue after orderly shutdown
- Target: (a) ≤ 120 MB; (b) ≤ 200 MB; (c) ≤ 1.0×; (d) 0 bytes
- Minimum threshold: (a) ≤ 150 MB; (b) ≤ 500 MB; (c) ≤ 1.5×; (d) 0 bytes
- Measurement method: Filesystem accounting by the harness after install, after DS-SOAK completion, after full DS-M index build, and after orderly shutdown; retention defaults per the owning volumes (memory retention Volume 7, log rotation Volume 10)
- Test environment: RM-1 and RM-2, DS-M, DS-SOAK
- Measurement frequency: Every release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: ADR-022, ADR-028
- Risks: RISK-PERF-001, RISK-PERF-004
- Acceptance criteria: Given a fresh install, when measured, then footprint ≤ 150 MB; given the completed soak, then workspace state ≤ 500 MB and the index cache ≤ 1.5× indexed text; given orderly shutdown, then zero temporary residue remains in the runtime directories.

### NFR-PERF-020 — Concurrency capacity and scheduler overhead

- Category: Scalability
- Priority: P1
- Phase: v1
- Metric: (a) Concurrent capacity: number of simultaneously active runs and tool invocations sustained while NFR-PERF-005 and NFR-PERF-007 remain within thresholds; (b) Task Scheduler submission-to-start overhead on an unloaded pool, p95 over ≥ 10,000 tasks
- Target: (a) ≥ 4 concurrent runs with ≥ 16 concurrent tool invocations on RM-1, ≥ 2 runs with ≥ 8 invocations on RM-2; (b) ≤ 250 µs p95
- Minimum threshold: (a) as target — capacity below it is a defect; (b) ≤ 1 ms p95
- Measurement method: Load scenario driving parallel scripted runs via the mock provider and no-op tools while the interactive replay measures latency; scheduler micro-benchmark per ADR-023's review condition; pool budgets under test are the chapter 03 values
- Test environment: RM-1 and RM-2, DS-M
- Measurement frequency: Nightly on mainline; every release; gate at v1
- Owner: Performance and Reliability (Volume 12) / Task Scheduler (Volume 3, chapter 08 contract)
- Dependencies: FR-ARCH-006; ADR-023; E-ARCH-005
- Risks: RISK-PERF-001
- Acceptance criteria: Given the stated load on RM-1, when the interactive replay runs concurrently, then input and streaming latency stay within thresholds and no `interactive`-pool submission is rejected; given the micro-benchmark, then supervision overhead p95 ≤ 1 ms — a sustained breach triggers the ADR-023 review condition.

### NFR-PERF-021 — Large-repository operation

- Category: Scalability
- Priority: P1
- Phase: v1
- Metric: On DS-L (≥ 100,000 files): (a) workspace open time, p95; (b) TUI startup to interactive, p95; (c) compliance of NFR-PERF-005 and NFR-PERF-007 on DS-L; scan, git status, and indexing per their own DS-L rows
- Target: (a) ≤ 2 s p95; (b) ≤ 1.5 s p95; (c) unchanged thresholds
- Minimum threshold: (a) ≤ 4 s p95; (b) ≤ 3 s p95; (c) unchanged thresholds
- Measurement method: The chapter's benchmarks re-run against DS-L; operations that scale with repository size (full index build) run on the `background` pool and are reported without gating, while interactive budgets are gated unchanged
- Test environment: RM-1 and RM-2, DS-L
- Measurement frequency: Every release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: NFR-PERF-005, NFR-PERF-007, NFR-PERF-009, NFR-PERF-010, NFR-PERF-011
- Risks: RISK-PERF-004
- Acceptance criteria: Given DS-L, when the suite runs, then open ≤ 4 s p95, TUI start ≤ 3 s p95, and interactive latency thresholds hold unchanged; negative case: exceeding the 200,000-file default indexing scope limit (chapter 03) narrows indexing scope with an explicit `perf.limit.enforced` event rather than degrading interactive latency.

### NFR-PERF-022 — Large-file handling

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: (a) Peak incremental RSS while a tool reads the DS-F 100 MiB file through streaming I/O; (b) TUI file-viewer first paint on the DS-F 1 GiB file and its RSS increase; (c) conformance to the Volume 7 large-file context rule — files above the context ingestion threshold are never fully loaded into a model request
- Target: (a) ≤ 32 MiB; (b) ≤ 500 ms p95 first paint, ≤ 100 MiB RSS increase; (c) 0 violations
- Minimum threshold: (a) ≤ 64 MiB; (b) ≤ 1 s p95, ≤ 200 MiB; (c) 0 violations
- Measurement method: DS-F benchmarks with process accounting; the context rule is verified by asserting on assembled request sizes in instrumented runs over DS-F
- Test environment: RM-1 and RM-2, DS-F
- Measurement frequency: Every release; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: FR-CTX-001; NFR-PERF-017
- Risks: RISK-PERF-001
- Acceptance criteria: Given the 100 MiB file, when a filesystem tool reads it, then output is streamed with truncation marked per the tool contract and peak incremental RSS ≤ 64 MiB; given the 1 GiB file in the viewer, then first paint ≤ 1 s p95 via virtualization with RSS increase ≤ 200 MiB; negative case: an attempt to ingest the 100 MiB file into context is bounded by the Volume 7 threshold, never a full load.

### NFR-PERF-023 — Long-session stability

- Category: Reliability
- Priority: P0
- Phase: v1
- Metric: Over the DS-SOAK 8-hour scripted session: (a) hour-8 p95 input latency as a multiple of hour-1 p95; (b) hour-8 first-token overhead as a multiple of hour-1; (c) leaked goroutines and child processes at orderly shutdown; (d) RSS per NFR-PERF-017(b)
- Target: (a) ≤ 1.1×; (b) ≤ 1.1×; (c) 0/0; (d) within budget
- Minimum threshold: (a) ≤ 1.25×; (b) ≤ 1.25×; (c) 0/0 (NFR-ARCH-004 — no tolerance); (d) within budget
- Measurement method: DS-SOAK replay with hourly latency checkpoints; leak detection per the NFR-ARCH-004 instrumentation at final shutdown
- Test environment: RM-1 and RM-2, DS-M, DS-SOAK
- Measurement frequency: Per release; weekly on mainline; gate at v1
- Owner: Performance and Reliability (Volume 12)
- Dependencies: NFR-ARCH-004; NFR-PERF-005, NFR-PERF-006, NFR-PERF-017
- Risks: RISK-PERF-001
- Acceptance criteria: Given the completed soak, when hour-8 checkpoints are compared with hour-1, then latency multiples ≤ 1.25×; given final shutdown, then zero leaked goroutines and zero surviving children; error case: any crash during the soak fails the run outright and is counted against NFR-PERF-027.

## Risks

### RISK-PERF-001 — Performance targets unattainable on reference hardware

- Category: Technical
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Benchmark spike before MVP implementation (the Volume 1 assumption's validation path) covering cold start, streaming overhead, and the SQLite hot append path (ADR-007's benchmark-gated review); budgets carry target/threshold headroom so early misses surface as trend warnings before they become gate failures; profiling investment is prioritized by benchmark evidence, not intuition
- Detection: Nightly benchmark trends against targets from MVP onward; `perf.benchmark.regressed` events; phase-gate audits per Volume 1 metric governance
- Owner: Performance and Reliability (Volume 12)
- Status: Open

The Go implementation choice (ADR-001) and the pure-Go SQLite driver (ADR-007) were made with
these budgets explicitly in view, but no target in this chapter is proven until measured on
RM-1 and RM-2. The dual target/threshold structure exists so the project learns it is drifting
while correction is still cheap; ADR-007 names the switch to a CGO driver as the recorded
fallback if the persistence hot path is the blocker.
