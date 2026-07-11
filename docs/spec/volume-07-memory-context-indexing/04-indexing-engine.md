# 04 — Indexing Engine

This chapter defines the Indexing Engine: construction, incremental maintenance, and querying
of the lexical and semantic indexes over workspace content and memory records. The engine
implements **IndexerPort** (`Build`, `Update`, `Query`, `Invalidate`, `Status` — frozen names,
Volume 3 chapter 02) and drives the Index state machine of chapter
[05](05-index-state-machine.md). Entity shapes and invariants are Volume 2 chapter 07 (Index,
Embedding; INV-IDX-01..04, INV-EMB-01..04). Storage follows ADR-028: index metadata rows in
the workspace database (`content_indexes`), index *data* — posting lists and vectors — in the
rebuildable cache database `.andromeda/index.db`, whose loss is never an integrity error
(INV-IDX-02). Vector search is exact in-process cosine similarity per ADR-020, with the
recorded design assumption of at most 100,000 chunks per workspace.

## Default indexes

On workspace open with `index.enabled = true`, the engine declares (creating metadata rows in
state `created` when absent):

| Name | Kind | Covers | Exists when |
|---|---|---|---|
| `files-lexical` | `lexical` | Workspace file content within scope rules | always |
| `files-semantic` | `semantic` | Chunked workspace file content (`file_chunk` embeddings) | a semantic embedding provider is configured (ADR-089) |
| `memory-semantic` | `semantic` | Session/workspace-layer Memory Records (`memory_record` embeddings) | a semantic embedding provider is configured |

Names are reserved: user-declared additional indexes (via Volume 8 `index` commands, by name)
MUST NOT collide with them (INV-IDX-01 uniqueness). `long_term` memory is not covered by any
workspace index — Index rows are workspace-scoped (Volume 2); global-scope semantic indexing
is an open question in the volume register, and long-term memory retrieval remains lexical
through MemoryStorePort until it is resolved.

## Scope and ignore rules

An index's `scope_config` resolves to the set of indexable files. The engine MUST apply, in
order: (1) always-excluded paths — `.andromeda/`, `.git/`, and the index cache itself; (2)
`.gitignore` semantics parsed by the engine (the format is public and stable; the engine does
not shell out to git for this — GitPort is not among its allowed dependencies, Volume 3); (3)
`.andromedaignore` at the workspace root, same pattern syntax, applied after `.gitignore` and
able to re-include with `!` patterns; (4) configuration include/exclude globs
(`index.include` / `index.exclude`); (5) the per-file byte cap `index.max_file_bytes`; (6)
binary detection per FR-CTX-006's rule (NUL byte in the first 8192 bytes or configured binary
extension) — binary files are never chunked or embedded.

## Requirements

### FR-IDX-001 — Indexing engine

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Indexing Engine (Volume 7)
- Affected components: Indexing Engine, Persistence Layer, Context Manager, Memory Manager, Workspace Engine, CLI, TUI
- Dependencies: ADR-020, ADR-028, ADR-089; FR-ARCH-003 (IndexerPort freeze); Volume 2 chapter 07; FR-IDX-006 (machine)
- Related risks: RISK-IDX-001, RISK-IDX-002

#### Description

The Indexing Engine MUST implement IndexerPort over the default and user-declared indexes:
`Build` declares an index (metadata row, state `created`) and runs a full build as a
supervised background task (SchedulerPort, by name), returning the Index ULID immediately;
`Update` applies explicit `PathChange` batches incrementally (FR-IDX-004); `Query` answers
lexical queries (token and phrase matching with path filters) and semantic queries (exact
cosine top-k per ADR-020) with budgets (max results, max latency hint) and returns hits
carrying path or memory reference, span, score, and **generation**; `Invalidate` marks
scopes stale (FR-IDX-004); `Status` reports state, generation, coverage (document count,
chunk count), and staleness metrics. Every committed build or update increments the index's
monotonic generation; queries are served only from the last committed generation — never
from an in-progress mutation (torn reads are structurally impossible). At most one mutator
(build or update) runs per index at a time; concurrent mutation attempts are rejected with
E-IDX-007. Lexical indexing MUST be fully self-contained (engine-owned posting lists in the
cache database; no external index service, no network).

#### Motivation

Retrieval quality gates context quality (FR-CTX-001 tier 7), and search is a direct user
surface (Volume 8 commands). One engine with one port keeps lexical and semantic retrieval
behind a single contract that the Context and Memory Managers can consume without knowing
the backing (ADR-020 reversal plan depends on exactly this).

#### Actors

Workspace Engine (open-time declaration, change feeds), Context Manager and Memory Manager
(queries), user (index commands), Task Scheduler (supervised builds).

#### Preconditions

Workspace open; `index.enabled = true`; for semantic indexes, an embedding provider
configured per ADR-089.

#### Main flow

1. `Build` writes the metadata row, transitions per chapter 05, and enumerates in-scope
   files via WorkspacePort.
2. Content is chunked (FR-IDX-002); lexical postings always; embeddings per FR-IDX-003
   for semantic indexes.
3. The build commits as one new generation; state `ready`; `index.build.completed` emits
   with coverage figures.
4. `Query` serves hits from the committed generation with scores and the generation
   number.

#### Alternative flows

- Build cancelled or timed out: prior committed generation (if any) remains served; the
  machine transitions per chapter 05; partial work is discarded (never a torn
  generation).
- Query against an index in `failed` or without a committed generation: E-IDX-005; the
  Context Manager degrades per FR-CTX-001.

#### Edge cases

- Empty scope builds succeed with zero documents (state `ready`, empty generation).
- Chunk count reaching `index.max_chunks` during build: behavior per
  `index.on_scale_exceeded` — `degrade` (index the first N chunks by deterministic path
  order, mark coverage partial, emit E-IDX-006 as a warning event) or `refuse` (build
  fails with E-IDX-006). Default `degrade`.
- Cache database missing or corrupted at open: metadata rows flip to `stale`/`failed` per
  chapter 05 recovery and a rebuild is scheduled — never exit code 9 (ADR-028 rule 5).

#### Inputs

`IndexSpec` (name, kind, scope config, embedding space for semantic), `PathChange`
batches, `IndexQuery` values, invalidation scopes.

#### Outputs

Index ULIDs, generations, `IndexHit` sets, `IndexStatus` reports, `index.*` events,
posting lists and vectors in `index.db`.

#### States

Canonical Index states `created`, `building`, `ready`, `updating`, `stale`, `failed`,
`removed` (frozen); full machine in chapter 05 (FR-IDX-006).

#### Errors

E-IDX-001 (build failure), E-IDX-005 (not queryable), E-IDX-006 (scale limit),
E-IDX-007 (concurrent mutation); catalog below.

#### Constraints

One mutator per index; generation monotonicity; query-from-committed-only; INV-IDX-01..04
and INV-EMB-01..04 enforced structurally; the engine reads files only within scope rules
and never follows symlinks out of the workspace (PAL filesystem semantics, by name).

#### Security

Index data derives exclusively from content the workspace grants; ignore rules exclude
paths the user marked out; memory embeddings inherit the FR-MEM-003 gate (gated content
never reaches the embedder). Index poisoning as an attack is Volume 9's threat catalog
(by name); the structural control here is provenance: hits always name their source and
generation.

#### Observability

`index.build.started`/`completed`/`failed`, `index.update.completed`,
`index.scope.invalidated`, `index.state.changed` events; coverage, staleness, chunk-count
(ADR-020 observability), and query latency metrics.

#### Performance

Query and build budgets are Volume 12's; the ADR-020 scale assumption (≤ 100k chunks)
bounds semantic scan cost; `Status` and `Query` never block behind mutators.

#### Compatibility

Pure-Go storage per ADR-007/020 on all Tier 1 platforms; index data never migrates —
layout changes rebuild (Volume 2 `index_schema_version` rule).

#### Acceptance criteria

- Given a workspace with in-scope files, when `Build` completes, then `Status` reports
  `ready` with a generation ≥ 1 and coverage equal to the enumerated scope (main case).
- Given a query during an in-progress update, when it runs, then results come entirely
  from the prior committed generation and carry its number (isolation case).
- Given a second `Update` while one runs, when submitted, then E-IDX-007 returns and the
  first mutator is unaffected (concurrency case).
- Given a corrupted cache database, when the workspace opens, then affected indexes
  rebuild automatically, no integrity error surfaces, and `state.db` is untouched
  (recovery case, INV-IDX-02).
- Given `.andromedaignore` excluding a path, when built, then no content from that path
  exists in postings or vectors (scope/negative case).
- Given any state transition, when events are inspected, then `index.state.changed`
  carries from/to states and the trigger (observability case).

#### Verification method

IndexerPort contract suite (Volume 13); generation-consistency tests under concurrent
update+query; rebuild-from-nothing equivalence tests; scope-rule fixtures; ADR-020 scale
benchmarks.

#### Traceability

PRD-003, PRD-006; ADR-020, ADR-028, ADR-089; INV-IDX-01..04; FR-CTX-001; FR-IDX-002..006.

### FR-IDX-002 — Chunking

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Indexing Engine (Volume 7)
- Affected components: Indexing Engine
- Dependencies: ADR-087 (estimation), ADR-088; FR-IDX-001
- Related risks: RISK-IDX-001

#### Description

Text content entering any index MUST be chunked per ADR-088: chunks target
`index.chunk.target_tokens` (default 400, estimated per ADR-087), hard-capped at
`index.chunk.max_tokens` (default 512), with `index.chunk.overlap_tokens` (default 40) of
trailing overlap between consecutive chunks. Chunk boundaries MUST prefer, in order:
Markdown/heading boundaries, blank-line boundaries, top-level indentation decreases, and
line boundaries; a chunk never splits inside a line. Each chunk carries a locator — path,
byte offsets, line span — and a content hash; the locator format is stable and is the
`source_ref` for `file_chunk` embeddings (Volume 2). Memory records chunk as whole records
(they are capped at ingestion, FR-MEM-004, below the chunk maximum). Chunking MUST be
deterministic: identical bytes produce identical chunk sets on every platform and run.

#### Motivation

Chunk geometry determines both retrieval precision and embedding cost; deterministic,
structure-preferring boundaries keep hits aligned with how humans and models read code and
prose, and keep re-index diffs minimal (only changed chunks re-embed, FR-IDX-004).

#### Actors

Indexing Engine (chunker); embeddings path (consumer); Context Manager (hit spans).

#### Preconditions

File within scope rules and byte cap; text (binary excluded per chapter rule).

#### Main flow

1. Content splits at preferred boundaries into windows ≤ target, growing to max where
   boundaries are sparse.
2. Overlap is appended from the previous chunk's tail.
3. Locators and hashes compute; chunks feed postings and (semantic) the embedding queue.

#### Alternative flows

- Single lines exceeding the max token cap (minified files): the line becomes one
  oversized chunk, truncated at the max for embedding purposes with the truncation
  recorded in the chunk metadata; lexical postings index the full line.

#### Edge cases

- Empty files produce zero chunks.
- Files with no structural boundaries chunk on line windows.
- CRLF and LF content chunk identically after canonical newline handling; hashes are
  computed over the file's actual bytes (staleness detection remains byte-faithful,
  INV-EMB-02).

#### Inputs

File bytes, memory record content, chunk configuration.

#### Outputs

Chunk sets with locators and hashes.

#### States

Not applicable — pure function of content and configuration.

#### Errors

None minted — unreadable files are build-item failures aggregated under E-IDX-001 detail.

#### Constraints

Determinism; no mid-line splits; locator stability across releases (a locator format
change is an index layout change → rebuild per `index_schema_version`).

#### Security

Chunking never widens scope; oversized-line truncation is recorded, not silent.

#### Observability

Chunk counts, oversized-line truncation occurrences, and per-file chunking contributions
appear in build reports and the chunk-count metric (ADR-020 observability); chunking emits
no events of its own — build-level events (FR-IDX-001) carry the aggregates.

#### Performance

Single pass per file, O(bytes); chunk-count contribution to the ADR-020 budget is
observable per file in build reports.

#### Compatibility

Identical output across Tier 1 platforms (verified by golden fixtures).

#### Acceptance criteria

- Given a Markdown file with headings, when chunked, then no chunk crosses a heading
  boundary unless a single section exceeds the max (boundary case).
- Given identical bytes on macOS and Linux, when chunked, then chunk sets and hashes are
  identical (determinism case).
- Given a one-line 1 MB minified file within the byte cap, when chunked, then lexical
  postings cover it, its embedding input is truncated at the max, and the truncation is
  recorded (edge case).
- Given a changed middle section of a file, when re-chunked, then chunks before the
  change retain identical hashes (incremental-friendliness case).

#### Verification method

Golden chunk fixtures across content types; cross-platform determinism tests; property
tests on boundary rules (Volume 13).

#### Traceability

ADR-087, ADR-088; FR-IDX-001, FR-IDX-003, FR-IDX-004; INV-EMB-02.

### FR-IDX-003 — Embeddings and semantic index construction

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Indexing Engine (Volume 7)
- Affected components: Indexing Engine, Provider Layer (via ProviderPort), Persistence Layer
- Dependencies: ADR-019, ADR-020, ADR-089; FR-IDX-001, FR-IDX-002; Volume 2 INV-IDX-03, INV-EMB-01..04
- Related risks: RISK-IDX-001, RISK-IDX-002

#### Description

Semantic indexes acquire vectors exclusively through `ProviderPort.Embed` against the
embedding provider and model named in the index's `embedding_space` — which is fixed at
index creation from `index.semantic.provider` / `index.semantic.model` and immutable
thereafter (INV-IDX-03): changing either creates a new index and retires the old one. The
provider MUST declare the `embeddings` capability (Volume 5 enum); absence is E-IDX-002 at
build declaration time, not silent degradation (Principle 2). Embedding requests batch up
to `index.semantic.batch_size` chunks (default 64, and never above the provider's declared
batch limit); transient failures retry per batch up to 3 attempts with exponential backoff
(base 1 s, factor 4, jittered) before the mutation fails with E-IDX-002. Vectors persist in
`index.db` with dimensions validated against the embedding space (INV-EMB-01; mismatch is
E-IDX-003 and fails the mutation — a provider returning wrong dimensions is a contract
violation, never stored). Re-embedding occurs only for chunks whose content hash changed
(INV-EMB-02/04); unchanged chunks carry their vectors across generations. Queries embed the
query text through the same embedding space and score by exact cosine per ADR-020.

#### Motivation

Embedding-space discipline is what makes similarity numbers meaningful (vectors from
different spaces are not comparable), and hash-gated re-embedding is what makes incremental
updates affordable in both money and time.

#### Actors

Indexing Engine (batching, storage, scoring), Provider Layer (Embed), user (space
configuration).

#### Preconditions

Embedding provider configured, declaring `embeddings`; index metadata declared.

#### Main flow

1. Chunks needing vectors (new or hash-changed) queue in deterministic order.
2. Batches embed via ProviderPort; vectors validate and persist.
3. The mutation commits its generation; query-time embedding uses the same space.

#### Alternative flows

- Provider unreachable (connection state not `available`, Volume 5 machine): a full build
  fails with E-IDX-002; an incremental update leaves the index `stale` with pending
  changes queued (chapter 05) — lexical service continues either way (FR-IDX-005).
- Cancellation mid-mutation: per the port contract, the prior generation stands.

#### Edge cases

- Zero-delta updates (all hashes unchanged) commit a generation without any Embed call.
- A remote provider's cost accrues to Cost Records attributed to the indexing task
  (Volume 2 accounting); cost visibility is a Principle 7 obligation.
- Query embedding failure (provider down at query time): the semantic query fails with
  E-IDX-002 detail `query embedding unavailable`; callers degrade per FR-CTX-001; cached
  chunk vectors are unaffected.

#### Inputs

Chunk sets with hashes, embedding space, batch configuration, query texts.

#### Outputs

Validated vectors in `index.db`; cosine-scored hits; cost records for remote providers.

#### States

Mutations drive the chapter 05 machine; vectors are immutable rows (INV-EMB-04).

#### Errors

E-IDX-002 (acquisition failure), E-IDX-003 (space mismatch); catalog below.

#### Constraints

One embedding space per index forever; no vector normalization tricks that change scoring
semantics between generations; batch order deterministic; no Embed calls for gated memory
content (FR-MEM-003 precedes indexing).

#### Security

Chunk content leaves the machine only when the configured embedding provider is remote —
an explicit, named configuration choice (ADR-089); the engine MUST refuse to fall back to
any other provider for embedding (no silent provider substitution; Volume 5 routing
fallback is disabled for embedding requests by contract here).

#### Observability

Embed batch counts, retry counts, and token/cost figures per mutation in build/update
events; chunk-count metrics per ADR-020's observability commitment.

#### Performance

Batching bounds request counts; cosine scan cost bounded by `index.max_chunks` (ADR-020);
budgets per Volume 12.

#### Compatibility

Provider-agnostic through ProviderPort and the capability enum; local and remote embedding
providers behave identically at this contract.

#### Acceptance criteria

- Given a configured provider without the `embeddings` capability, when a semantic build
  is declared, then E-IDX-002 returns at declaration and no request is sent (negative
  case).
- Given a provider returning vectors of wrong dimensionality, when a batch lands, then
  E-IDX-003 fails the mutation and no such vector exists in `index.db` (contract case).
- Given an update where 2 of 100 chunks changed, when it completes, then exactly the 2
  changed chunks were re-embedded (hash-gating case).
- Given a remote embedding space, when any mutation runs, then all Embed traffic went to
  the configured provider only, and cost records exist for it (security/cost case).
- Given transient provider failures, when a batch retries within policy and then
  succeeds, then the mutation completes and retry counts appear in the event
  (resilience/observability case).

#### Verification method

Provider-double contract tests (capability refusal, dimension mismatch, retry/backoff);
hash-gating property tests; cost-attribution integration tests (Volume 13).

#### Traceability

PRD-002, PRD-003; ADR-019, ADR-020, ADR-089; INV-IDX-03, INV-EMB-01..04; FR-IDX-005;
RISK-IDX-002.

### FR-IDX-004 — Incremental updates and invalidation

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Indexing Engine (Volume 7)
- Affected components: Indexing Engine, Workspace Engine
- Dependencies: FR-IDX-001..003; FR-IDX-006 (machine)
- Related risks: RISK-IDX-001

#### Description

The engine MUST maintain indexes incrementally from explicit `PathChange` feeds: the
Workspace Engine's file watches (debounced by `index.watch.debounce_ms`, default 500) and
File Change records from runs (both by name; the port carries only `PathChange` batches).
An `Update` recomputes affected files only: changed files re-chunk, changed chunks
re-embed (FR-IDX-003), deleted files' postings and vectors drop, renames drop+add. Pending
changes accumulate per index while a mutator runs or the provider is unavailable; when
pending count exceeds `index.stale.max_pending_changes` (default 500) or pending age
exceeds `index.stale.max_age_seconds` (default 3600), the index transitions `ready` →
`stale` (chapter 05) — still queryable, marked out of date, with staleness metrics in
`Status` and hit generations telling consumers exactly how old results are.
`Invalidate` marks a path scope or the whole index stale and forces recomputation of that
scope before its entries are served again; whole-index invalidation schedules a full
rebuild. Dropping the cache database entirely MUST always be legal and equivalent to
whole-index invalidation of every index (INV-IDX-02).

#### Motivation

Full rebuilds on every save are unaffordable; unbounded staleness silently rots retrieval.
Explicit change feeds with bounded staleness give freshness with a defined, observable
degradation mode.

#### Actors

Workspace Engine (feeds), Indexing Engine (updates), user (manual invalidate/rebuild
commands, Volume 8).

#### Preconditions

Index has a committed generation; change feed active.

#### Main flow

1. Debounced changes batch into an `Update`.
2. Affected-only recomputation runs as the single mutator; a new generation commits;
   `index.update.completed` emits with delta figures.

#### Alternative flows

- Update failure with intact prior generation: index → `stale` with the failed batch
  re-queued (chapter 05 retry policy); service continues on the prior generation.
- Invalidation of a scope currently pending changes: scopes merge; recomputation covers
  the union.

#### Edge cases

- Changes to out-of-scope paths are discarded without state effect.
- A change storm exceeding the pending cap mid-update: the running update commits its
  batch; the index lands in `stale` and a catch-up update schedules immediately.
- Watch loss (platform watcher overflow): the Workspace Engine signals it; the engine
  treats it as whole-index invalidation (correctness over cleverness).

#### Inputs

`PathChange` batches, invalidation scopes, staleness thresholds.

#### Outputs

New generations, staleness transitions and metrics, events.

#### States

`ready` ↔ `updating`, `ready`/`updating` → `stale`, per the chapter 05 machine.

#### Errors

E-IDX-002 (embedding path), E-IDX-004 (cache corruption discovered during update),
E-IDX-007 (mutator conflict).

#### Constraints

Affected-only recomputation; generation atomicity; staleness thresholds enforced;
invalidated entries never served until recomputed.

#### Security

Change feeds carry paths only; scope rules re-apply on every update (a newly ignored path
drops from the index on its next change or invalidation).

#### Observability

`index.update.completed` with add/change/delete/re-embed counts and duration;
`index.scope.invalidated` with scope and reason; staleness gauges.

#### Performance

Update cost proportional to delta size; debounce bounds update frequency; budgets per
Volume 12.

#### Compatibility

Watcher differences are the PAL's concern (by name); the engine consumes uniform
`PathChange` batches everywhere.

#### Acceptance criteria

- Given one modified file among 5,000, when the update completes, then only that file's
  chunks recomputed and the generation incremented once (affected-only case).
- Given pending changes beyond the cap, when the threshold trips, then the index is
  `stale`, still answers queries with the old generation, and `Status` reports pending
  figures (staleness case).
- Given `Invalidate` on a path scope, when a query targets that scope before
  recomputation, then no entry from the invalidated scope is served (invalidation case).
- Given the cache file deleted while running, when the next operation touches it, then
  affected indexes rebuild per recovery rules and no data loss occurs outside the cache
  (INV-IDX-02 case).
- Given a rename, when updated, then old-path entries are gone and new-path entries
  present with identical content hashes (rename case).

#### Verification method

Watch-feed integration tests with change storms; staleness threshold property tests;
invalidation-service exclusion tests; cache-drop equivalence tests (Volume 13).

#### Traceability

PRD-010 (durable, honest state); INV-IDX-02; FR-IDX-001, FR-IDX-003, FR-IDX-006.

### FR-IDX-005 — Operation without embeddings and without Internet

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Indexing Engine (Volume 7)
- Affected components: Indexing Engine, Context Manager
- Dependencies: ADR-089; FR-IDX-001, FR-IDX-003; Volume 1 Principle 3 (offline guarantee item 3); Volume 2 INV-IDX-04
- Related risks: RISK-IDX-002

#### Description

The engine MUST be fully functional with **no embedding provider configured**: lexical
indexes build, update, and answer queries with no degradation, and semantic indexes simply
do not exist (their declaration is refused with E-IDX-002 naming the missing
configuration; nothing pretends to be semantic — Principle 2's no-silent-simulation
applied to retrieval). It MUST be fully functional **offline**: lexical operations always
(INV-IDX-04); semantic operations whenever the embedding provider is local; with a remote
provider offline, committed semantic generations keep serving queries whose query-vector
can be produced (they cannot — query embedding needs the provider — so semantic queries
fail with E-IDX-002 and consumers degrade to lexical per FR-CTX-001), and incremental
updates queue into `stale` per FR-IDX-004. Degradations MUST be reported: `Status` names
the unavailable capability, hits never masquerade across kinds, and the Context Manager's
assembly metadata records the retrieval mode used (FR-CTX-004). This requirement realizes
offline guarantee item 3 ("indexing files") and is measured by NFR-IDX-001.

#### Motivation

Local-first is identity (PRD-003): a developer on a plane gets full lexical search and
whatever semantic capability their local models provide, with honest reporting instead of
mysterious emptiness.

#### Actors

Indexing Engine; Context Manager (degradation consumer); offline suite.

#### Preconditions

None — this is the baseline.

#### Main flow

1. Offline or embedding-less operation proceeds for all lexical paths identically to
   online operation.
2. Semantic paths either work (local provider) or fail/queue explicitly per the rules
   above.

#### Alternative flows

- Provider becomes reachable again: queued changes drain (`stale` → `updating` →
  `ready`, chapter 05); no user action required.

#### Edge cases

- First workspace open offline with a remote embedding space configured: `files-semantic`
  stays `created` (build refused E-IDX-002 with detail `provider unreachable`); it builds
  automatically on the first open with reachability.
- Mixed fleets (local chat model, remote embedding model): chat works offline; semantic
  indexing queues — independence of the two paths is required.

#### Inputs

The full port surface under offline and embedding-less conditions.

#### Outputs

Full lexical service; explicit semantic failures/queues; degradation reporting.

#### States

`stale` accumulation and drain per chapter 05; no offline-specific states.

#### Errors

E-IDX-002 with distinguishing detail (`not configured` vs `provider unreachable`); never
a generic failure that hides the cause.

#### Constraints

No network attempts on lexical paths (structural: lexical code has no ProviderPort
access); no queued content transmission besides the explicit embedding queue whose
destination the user configured (ADR-089).

#### Security

Offline operation proves the no-hidden-egress property for the indexing subsystem
(NFR-IDX-001 companion to NFR-MEM-001).

#### Observability

Degradation reasons in `Status` and events; offline suite results per release.

#### Performance

Lexical-only operation meets the same Volume 12 lexical budgets online and offline.

#### Compatibility

Identical across Tier 1 platforms; local providers (Ollama adapter, OpenAI-compatible
local servers — Volume 5 catalog, by name) exercise the semantic-offline path in CI.

#### Acceptance criteria

- Given no embedding provider configured, when the workspace opens and builds run, then
  `files-lexical` reaches `ready`, no semantic index exists, and search commands work
  (baseline case).
- Given all interfaces disabled with a local embedding provider, when semantic build and
  query run, then both succeed with zero egress (offline-semantic case, NFR-IDX-001
  method).
- Given a remote embedding provider offline, when a semantic query runs, then E-IDX-002
  with `provider unreachable` returns and the same query served lexically succeeds
  (degradation case).
- Given reachability restored, when the queue drains, then the index returns to `ready`
  without user action (recovery case).

#### Verification method

Offline suite (OS-level disablement) covering lexical build/update/query and both
semantic conditions; egress capture; drain-on-reconnect integration tests (Volume 13).

#### Traceability

PRD-003; Volume 1 Principle 3 item 3; INV-IDX-04; ADR-089; NFR-IDX-001; FR-CTX-001.

### NFR-IDX-001 — Offline indexing guarantee

- Category: Portability
- Priority: P0
- Phase: MVP
- Metric: Pass rate of lexical index operations (build, update, query, invalidate, status) and local-provider semantic operations under the offline condition; network connection attempts observed from the Indexing Engine during lexical operations
- Target: 100% pass; 0 connection attempts on lexical paths
- Minimum threshold: 100% / 0 (identity property; offline guarantee item 3 binds at MVP exit per Volume 1)
- Measurement method: Volume 13 offline suite with OS-level interface disablement over the Volume 12 reference repository; egress capture; static verification that lexical code paths hold no ProviderPort reference
- Test environment: Volume 12 reference hardware, offline condition
- Measurement frequency: every CI run; gated at MVP exit and every release
- Owner: Indexing Engine (Volume 7)
- Dependencies: FR-IDX-005; INV-IDX-04
- Risks: RISK-IDX-002
- Acceptance criteria: The offline suite passes all lexical checks and local-provider semantic checks with zero observed egress from indexing code; remote-provider semantic checks fail with the specified E-IDX-002 details and lexical fallback succeeds.

### NFR-IDX-002 — Generation integrity under interruption

- Category: Reliability
- Priority: P0
- Phase: MVP
- Metric: Torn or mixed-generation query results, and indexes left unqueryable with an intact prior generation, per 1,000 injected interruptions (crash, kill, cancellation, timeout) of builds and updates
- Target: 0
- Minimum threshold: 0 (the port contract makes this absolute: consistent prior generation or `failed`)
- Measurement method: Crash/cancel-injection harness interrupting mutations at randomized points, followed by consistency verification: every query result set maps to exactly one committed generation; recovery per chapter 05 lands in the specified states
- Test environment: Volume 12 reference hardware; Volume 13 fixture workspaces
- Measurement frequency: every release; reduced matrix per CI run
- Owner: Indexing Engine (Volume 7)
- Dependencies: FR-IDX-001, FR-IDX-004, FR-IDX-006
- Risks: RISK-IDX-001
- Acceptance criteria: All injected interruptions leave each index either serving its intact prior generation (states `ready`/`stale`) or in `failed`/`created` with no partial generation observable; zero mixed-generation result sets across the harness.

### RISK-IDX-001 — Corpus scale beyond the ADR-020 assumption

- Category: Technical
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: `index.max_chunks` bound with declared `degrade`/`refuse` behavior (FR-IDX-001, E-IDX-006); chunk-count observability mandated by ADR-020; scope rules and byte caps keeping corpora lean; ADR-020 review conditions trigger the ANN evaluation when real workspaces exceed the assumption
- Detection: Chunk-count metrics per workspace; E-IDX-006 warning events; Volume 12 latency benchmarks at scale boundaries
- Owner: Indexing Engine (Volume 7)
- Status: Open — the ANN successor path is PENDING VALIDATION per ADR-020 (see the volume register)

Monorepo-scale workspaces can exceed 100k chunks, where exact cosine scans degrade
linearly. The bound converts silent degradation into a visible, configured behavior, and
ADR-020's reversal plan (re-backing the scorer behind IndexerPort) is the prepared exit.

### RISK-IDX-002 — Embedding-space drift across provider model revisions

- Category: Technical / external dependency
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Embedding space pinned per index with provider slug, model name, and dimensions (INV-IDX-03); dimension validation per batch (E-IDX-003); content-hash staleness keeps re-embedding explicit; a provider-side model revision that changes vector semantics is handled by creating a new index (new space) rather than mixing vectors
- Detection: Dimension mismatches (hard failure); retrieval-quality regression signals from conformance fixtures (Volume 13 semantic retrieval golden sets); provider deprecation notices surfaced by Volume 5's catalog maintenance
- Owner: Indexing Engine (Volume 7)
- Status: Open — provider-side vector stability across same-name model revisions is an assumption recorded in the volume register

Remote providers can silently revise the model behind a name; vectors from different
revisions may not be comparable even at identical dimensionality. Detection is imperfect
(same-dimension drift is invisible structurally), so the assumption is recorded and golden
retrieval fixtures provide a behavioral tripwire.

## Error catalog (E-IDX)

### E-IDX-001 — Index build failure

- Category: Execution
- Severity: Error
- User message: "Building the index '<name>' failed: <summary>."
- Technical message: index ULID, phase (enumerate/chunk/embed/commit), per-item failure list (bounded), error chain
- Cause: unreadable content beyond tolerance, cache write failure, embedding failure escalated from E-IDX-002, timeout, cancellation
- Safe-to-log data: index name/ULID, phase, counts, sanitized causes — no file content
- Recoverability: recoverable — rebuild after remediation; prior generation (if any) still serves
- Retry policy: no automatic retry of full builds; the next workspace open or a manual rebuild retries
- Recommended action: inspect the named phase; for provider causes see E-IDX-002 guidance
- Exit-code mapping: 1
- HTTP mapping: 500 over IPC
- Telemetry event: `index.build.failed`
- Security implications: none; paths appear per scope rules only

### E-IDX-002 — Embedding acquisition unavailable

- Category: External dependency
- Severity: Error
- User message: "Semantic indexing is unavailable: <detail>." (details: `not configured`, `capability missing`, `provider unreachable`, `query embedding unavailable`)
- Technical message: embedding space, provider connection state, capability check result, retry/backoff history
- Cause: no configured embedding provider (ADR-089), provider without `embeddings` capability, provider unreachable, exhausted batch retries
- Safe-to-log data: provider slug, model name, detail class, retry counts — no chunk content
- Recoverability: recoverable — configure/restore the provider; lexical operation continues throughout
- Retry policy: batch-level 3 attempts with exponential backoff (FR-IDX-003); update-path failures queue as `stale` and drain on reachability
- Recommended action: per detail — configure `[index.semantic]`, choose a capable provider, or restore connectivity
- Exit-code mapping: 1 (7 when a CLI command's sole purpose was the provider-backed operation)
- HTTP mapping: 503 over IPC
- Telemetry event: within the `index.*` family with error envelope
- Security implications: refusal is explicit — no silent fallback to another provider (FR-IDX-003 security rule)

### E-IDX-003 — Embedding space violation

- Category: Internal defect / contract
- Severity: Error
- User message: "The embedding provider returned incompatible vectors; the index was not modified."
- Technical message: expected space (provider, model, dimensions), received dimensions, batch identity
- Cause: provider returning vectors inconsistent with the declared space (INV-EMB-01), or an attempt to mutate `embedding_space` (INV-IDX-03)
- Safe-to-log data: space declaration, received dimensions, batch index
- Recoverability: recoverable — the mutation failed atomically; investigate provider or create a new index for a new space
- Retry policy: none automatic (deterministic contract violation)
- Recommended action: verify provider/model configuration; a model change requires a new index (ADR-089 guidance)
- Exit-code mapping: 1
- HTTP mapping: 502 over IPC
- Telemetry event: within the `index.*` family with error envelope
- Security implications: prevents silent corruption of similarity semantics (RISK-IDX-002 control)

### E-IDX-004 — Index cache corruption

- Category: Storage
- Severity: Warning
- User message: "The search index cache was corrupted and is being rebuilt."
- Technical message: cache path, detection point (open/read/update), affected indexes
- Cause: torn cache writes from crashes, external file damage
- Safe-to-log data: cache path, affected index names, detection point
- Recoverability: fully recoverable — the cache is non-authoritative (INV-IDX-02); automatic rebuild
- Retry policy: automatic — affected indexes transition per chapter 05 recovery and rebuild
- Recommended action: none required; investigate recurring corruption (disk health)
- Exit-code mapping: none (never fails a command by itself; rebuild proceeds in background)
- HTTP mapping: not applicable
- Telemetry event: within the `index.*` family with error envelope
- Security implications: never exit code 9 — cache loss is not an integrity error (ADR-028 rule 5)

### E-IDX-005 — Index not queryable

- Category: Not found / state
- Severity: Warning
- User message: "The index '<name>' is not available for search right now: <state>."
- Technical message: index ULID/name, current state, generation availability, last error summary
- Cause: unknown index ULID, state `failed` without a committed generation, `created` never built, `removed`
- Safe-to-log data: index name/ULID, state, generation number
- Recoverability: recoverable — build/rebuild the index
- Retry policy: none automatic at query time; consumers degrade (FR-CTX-001)
- Recommended action: run the index build command or wait for the scheduled rebuild
- Exit-code mapping: 1
- HTTP mapping: 409 over IPC
- Telemetry event: within the `index.*` family with error envelope
- Security implications: existence reporting respects workspace scope

### E-IDX-006 — Index scale limit reached

- Category: Resource limit
- Severity: Warning (degrade) / Error (refuse)
- User message: "The workspace exceeds the configured indexing limit of <n> chunks; indexing <behavior>."
- Technical message: chunk count, limit, `on_scale_exceeded` behavior, coverage figures
- Cause: corpus beyond `index.max_chunks` (ADR-020 assumption boundary)
- Safe-to-log data: counts, limit, behavior, coverage ratio
- Recoverability: recoverable — narrow scope rules, raise the limit knowingly, or await the ANN path (ADR-020 review)
- Retry policy: none automatic
- Recommended action: tighten `index.include`/`index.exclude`; review ADR-020 guidance before raising the limit
- Exit-code mapping: none when degrading; 1 when configured to refuse
- HTTP mapping: not applicable / 507 over IPC when refusing
- Telemetry event: within the `index.*` family with error envelope
- Security implications: none

### E-IDX-007 — Concurrent index mutation rejected

- Category: Concurrency
- Severity: Warning
- User message: "The index '<name>' is already being built or updated."
- Technical message: index ULID, running mutation kind and start time, rejected request kind
- Cause: second build/update submitted while a mutator holds the index (FR-IDX-001 single-mutator rule)
- Safe-to-log data: index name, mutation kinds, timestamps
- Recoverability: recoverable — the running mutation completes; changes queue for the next update
- Retry policy: automatic re-queue of update batches; manual retry for explicit builds
- Recommended action: none; check `Status` for progress
- Exit-code mapping: 1 when a CLI command was rejected
- HTTP mapping: 409 over IPC
- Telemetry event: within the `index.*` family with error envelope
- Security implications: none

## Events minted (index family)

| Event | Emitted when | Payload (summary) |
|---|---|---|
| `index.build.started` | A full build begins (FR-IDX-001) | index ULID/name, kind, scope digest |
| `index.build.completed` | A build commits its generation | index ULID, generation, document/chunk counts, duration, embed batch/retry/cost figures |
| `index.build.failed` | A build fails (E-IDX-001) | index ULID, phase, error envelope reference |
| `index.update.completed` | An incremental update commits (FR-IDX-004) | index ULID, generation, add/change/delete/re-embed counts, duration |
| `index.scope.invalidated` | A scope or whole index is invalidated (FR-IDX-004, FR-MEM-008 cascade) | index ULID, scope digest, reason |
| `index.state.changed` | Any state transition of the chapter 05 machine | index ULID, from/to states, trigger |

## Configuration: `[index]` table

```toml
[index]
enabled = true
include = []                        # additional include globs; empty = whole workspace
exclude = []                        # additional exclude globs (applied after ignore files)
max_file_bytes = 1048576            # per-file indexing cap
max_chunks = 100000                 # ADR-020 assumption boundary
on_scale_exceeded = "degrade"       # "degrade" | "refuse" (E-IDX-006)

[index.chunk]
target_tokens = 400                 # ADR-088
max_tokens = 512
overlap_tokens = 40

[index.semantic]
provider = ""                       # empty = semantic indexing off (ADR-089)
model = ""
batch_size = 64

[index.watch]
enabled = true
debounce_ms = 500

[index.stale]
max_pending_changes = 500           # FR-IDX-004 staleness thresholds
max_age_seconds = 3600

[index.timeouts]
build_seconds = 1800                # chapter 05 timeout budgets
update_seconds = 300
```
