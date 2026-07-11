# 99 — Volume 7 Register

Machine-parseable register of everything Volume 7 minted, per Volume 0 chapters 02 and 03.
Merged into the Volume 0 registers at consolidation.

## Requirements index

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-MEM-001 | Memory model | MVP | MemoryStorePort contract suite; isolation and ordering-determinism property tests; offline degraded-path suite |
| FR-MEM-002 | Provenance, trust, and source attribution | MVP | Stamping/ordering unit tests; adversarial supersession tests; CLI/TUI provenance inspection tests |
| FR-MEM-003 | Secret and sensitive-content exclusion | MVP | Redaction-gate fixture tests; bypass-attempt property tests over all write paths |
| FR-MEM-004 | Ingestion and normalization | MVP | Ingest contract tests (transactionality, cancellation); mode-matrix tests; normalization determinism property tests |
| FR-MEM-005 | Deduplication, conflict resolution, and supersession versioning | MVP | Chain-linearity and idempotency property tests; head-contention concurrency tests; trust-guard adversarial tests |
| FR-MEM-006 | Compression and summarization (consolidation) | Beta | Integration tests with scripted provider doubles; offline deferral tests; archive-link integrity property tests |
| FR-MEM-007 | Retention and expiration | MVP | Idempotency/forward-only property tests; time-travel fixtures; cascade suite |
| FR-MEM-008 | Deletion and purge cascade | MVP | NFR-MEM-002 measurement suite; crash-injection cascade tests; permission enforcement tests |
| FR-MEM-009 | Export and portability | MVP | Round-trip suites; determinism tests; permission and encryption fixture tests |
| FR-MEM-010 | Offline memory operation | MVP | Offline suite (OS-level disablement) over the full operation surface; egress capture |
| FR-CTX-001 | Context assembly | MVP | Determinism/property tests; golden assembly fixtures; replay divergence tests; source-failure injection |
| FR-CTX-002 | Token budgeting and model window enforcement | MVP | Budget arithmetic unit tests; counting-unavailable contract-double tests; NFR-CTX-001 measurement |
| FR-CTX-003 | Deduplication and compression of context | MVP | Dedup/ladder property tests; containment golden fixtures; no-inference instrumentation assertions |
| FR-CTX-004 | Freshness, trust, provenance, and conflict detection | MVP | Conflict-matrix fixtures; staleness injection tests; adversarial trust tests |
| FR-CTX-005 | User pinning and exclusion | MVP | Command integration tests; assembly filter property tests; resume round-trips; cap unit tests |
| FR-CTX-006 | Tool-result, large-file, and binary-content handling | MVP | Transformation golden fixtures; detection determinism property tests; byte-level binary assertions |
| FR-CTX-007 | Context snapshots and reproducibility | MVP | Record-completeness validator (SM-12 method); replay divergence tests; crash injection between assembly and send |
| FR-IDX-001 | Indexing engine | MVP | IndexerPort contract suite; generation-consistency tests under concurrency; rebuild equivalence tests; scope fixtures; ADR-020 scale benchmarks |
| FR-IDX-002 | Chunking | MVP | Golden chunk fixtures; cross-platform determinism tests; boundary property tests |
| FR-IDX-003 | Embeddings and semantic index construction | MVP | Provider-double contract tests (capability refusal, dimension mismatch, retry); hash-gating property tests; cost-attribution tests |
| FR-IDX-004 | Incremental updates and invalidation | MVP | Watch-feed integration tests with change storms; staleness property tests; invalidation exclusion tests; cache-drop equivalence |
| FR-IDX-005 | Operation without embeddings and without Internet | MVP | Offline suite over lexical and both semantic conditions; egress capture; drain-on-reconnect tests |
| FR-IDX-006 | Index state machine conformance and recovery | MVP | State-machine property suite over randomized operation/fault sequences; crash-injection harness; recovery idempotence tests |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-MEM-001 | Memory privacy and locality | MVP | Egress capture during offline and instrumented-online suites; static dependency check |
| NFR-MEM-002 | Deletion completeness | MVP | Deletion suite with exhaustive port queries and database scans; crash-injection variant |
| NFR-CTX-001 | Context window compliance | MVP | Overflow classification over integration/E2E suites; seeded infeasibility fault matrix |
| NFR-IDX-001 | Offline indexing guarantee | MVP | Offline suite with egress capture; static ProviderPort-reference verification on lexical paths |
| NFR-IDX-002 | Generation integrity under interruption | MVP | Crash/cancel-injection harness with generation-consistency verification |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-MEM-001 | Memory poisoning and stale-knowledge drift | High | Open |
| RISK-MEM-002 | Unbounded memory and embedding growth | Medium | Open |
| RISK-CTX-001 | Critical-content eviction under budget pressure | High | Open |
| RISK-IDX-001 | Corpus scale beyond the ADR-020 assumption | Medium | Open — ANN successor path PENDING VALIDATION per ADR-020 |
| RISK-IDX-002 | Embedding-space drift across provider model revisions | Medium | Open — vector-stability assumption recorded below |

## ADRs minted

Volume 7 block allocation is 085–099 (Volume 0 chapter 03); this volume used 085–089; 090–099
remain permanent gaps unless minted by later amendment.

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-085](../annexes/adr/ADR-085.md) | Two-axis memory taxonomy: storage layer × memory kind | Accepted | Layer (`session`/`workspace`/`long_term`) × closed kind vocabulary (`episodic`, `semantic`, `procedural`, `preference`, `decision`); the eight mandated memory concepts map onto the grid |
| [ADR-086](../annexes/adr/ADR-086.md) | Deterministic tiered context assembly with refusal-over-drop | Accepted | Eight fixed priority tiers; tiers 1–3 mandatory with E-CTX-001 refusal instead of silent drop; deterministic intra-tier ordering; no inference in the pipeline |
| [ADR-087](../annexes/adr/ADR-087.md) | Token counting: official provider counting first, calibrated byte heuristic fallback | Accepted | `CountTokens` when officially offered (cached); else `ceil(bytes/4)+8` with 1.15 inflation on estimates plus a configurable budget safety margin; local tokenizer adoption PENDING VALIDATION |
| [ADR-088](../annexes/adr/ADR-088.md) | Chunking: deterministic structure-preferring windows with overlap | Accepted | 400-token target / 512 cap / 40 overlap; boundary preference headings → blank lines → indentation → lines; never mid-line; geometry changes rebuild via `index_schema_version` |
| [ADR-089](../annexes/adr/ADR-089.md) | Semantic indexing is explicit opt-in; lexical is the always-on baseline | Accepted | Lexical unconditional and offline-guaranteed; semantic only with explicit provider+model configuration; embedding traffic exempt from routing fallback |

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-MEM-001 | Memory record validation failure | 1 (2 for CLI-argument causes) |
| E-MEM-002 | Redaction gate refusal | 5 when a confirmation was denied; otherwise 1 |
| E-MEM-003 | Memory store unavailable | 9 for integrity/migration failures; otherwise 1 |
| E-MEM-004 | Memory record not found | 1 |
| E-MEM-005 | Memory maintenance or transfer failure | 1 |
| E-MEM-006 | Invalid retention policy | 3 |
| E-CTX-001 | Context budget infeasible | 1 |
| E-CTX-002 | Context source read failure | 1 when a turn fails; otherwise none (recorded degradation) |
| E-CTX-003 | Context snapshot persistence failure | 9 for integrity failures; otherwise 1 |
| E-CTX-004 | Invalid pin or exclusion | 2 |
| E-IDX-001 | Index build failure | 1 |
| E-IDX-002 | Embedding acquisition unavailable | 1 (7 when the command's sole purpose was the provider-backed operation) |
| E-IDX-003 | Embedding space violation | 1 |
| E-IDX-004 | Index cache corruption | none (background rebuild) |
| E-IDX-005 | Index not queryable | 1 |
| E-IDX-006 | Index scale limit reached | none when degrading; 1 when refusing |
| E-IDX-007 | Concurrent index mutation rejected | 1 when a CLI command was rejected |

## Events minted

Per the Volume 0 event grammar; envelope, ordering, delivery, persistence, retention, and
redaction semantics per Volume 10. All payloads are content-free (ULIDs, counts, figures).

| Event | Emitted by | Meaning |
|---|---|---|
| `memory.record.ingested` | Memory Manager | A new memory record persisted (FR-MEM-004) |
| `memory.record.superseded` | Memory Manager | A supersession committed (FR-MEM-005) |
| `memory.record.archived` | Memory Manager | Retention or consolidation archived a record (FR-MEM-006/007) |
| `memory.record.expired` | Memory Manager | Retention expired a record (FR-MEM-007) |
| `memory.record.deleted` | Memory Manager | Deletion tombstoned a record (FR-MEM-008) |
| `memory.retention.completed` | Memory Manager | A retention pass finished (FR-MEM-007) |
| `memory.consolidation.completed` | Memory Manager | A consolidation pass finished (FR-MEM-006) |
| `memory.ingestion.refused` | Memory Manager | The redaction/validation gate refused content (FR-MEM-003, E-MEM-001/002) |
| `memory.export.completed` | Memory Manager | An export stream closed complete (FR-MEM-009) |
| `context.assembly.completed` | Context Manager | A context snapshot persisted and the set was handed off (FR-CTX-001/007) |
| `context.budget.exceeded` | Context Manager | Assembly refused an infeasible mandatory set (FR-CTX-002, E-CTX-001) |
| `context.conflict.detected` | Context Manager | Conflicting candidates were found (FR-CTX-004) |
| `context.pin.changed` | Context Manager | A pin was added or removed (FR-CTX-005) |
| `context.exclusion.changed` | Context Manager | An exclusion was added or removed (FR-CTX-005) |
| `index.build.started` | Indexing Engine | A full build began (FR-IDX-001) |
| `index.build.completed` | Indexing Engine | A build committed its generation (FR-IDX-001) |
| `index.build.failed` | Indexing Engine | A build failed (E-IDX-001) |
| `index.update.completed` | Indexing Engine | An incremental update committed (FR-IDX-004) |
| `index.scope.invalidated` | Indexing Engine | A scope or whole index was invalidated (FR-IDX-004) |
| `index.state.changed` | Indexing Engine | Any Index machine transition (chapter 05) |

## Config keys minted

Key content owned by this volume; schema, precedence, env-var mapping, and validation by
Volume 10 (single-home matrix).

| Table | Keys |
|---|---|
| `[memory]` | `enabled`, `max_content_bytes` |
| `[memory.ingestion]` | `mode` |
| `[memory.retention]` | `session_days`, `workspace_days`, `long_term_days`, `archive_before_expire`, `archive_grace_days`, `purge_after_days`, `importance_protect_threshold` |
| `[memory.ranking]` | `weight_relevance`, `weight_recency`, `weight_importance`, `weight_trust` |
| `[memory.consolidation]` | `enabled`, `max_records_per_pass`, `max_tokens_per_pass` |
| `[context.budget]` | `reserve_output_tokens`, `safety_margin_ratio` |
| `[context.history]` | `max_turns`, `max_tokens_ratio` |
| `[context.tool_results]` | `max_tokens_per_result` |
| `[context.files]` | `max_file_bytes`, `excerpt_tokens` |
| `[context.compaction]` | `min_tokens` |
| `[context.pinning]` | `max_ratio` |
| `[context.summarization]` | `use_memory_summaries` |
| `[context.models.<name>]` | `max_input_tokens` (tighten-only per-model override) |
| `[index]` | `enabled`, `include`, `exclude`, `max_file_bytes`, `max_chunks`, `on_scale_exceeded` |
| `[index.chunk]` | `target_tokens`, `max_tokens`, `overlap_tokens` |
| `[index.semantic]` | `provider`, `model`, `batch_size` |
| `[index.watch]` | `enabled`, `debounce_ms` |
| `[index.stale]` | `max_pending_changes`, `max_age_seconds` |
| `[index.timeouts]` | `build_seconds`, `update_seconds` |

## Glossary additions

| Term | One-line meaning |
|---|---|
| Memory layer | The storage/visibility axis of a Memory Record: `session`, `workspace`, or `long_term` (values frozen by Volume 2; semantics chapter 01). |
| Memory kind | The closed content-class axis of a Memory Record: `episodic`, `semantic`, `procedural`, `preference`, `decision` (ADR-085). |
| Supersession chain | The linear, acyclic version history formed when a memory record replaces another (FR-MEM-005); heads are retrievable, history on request. |
| Consolidation | The Beta lifecycle pass that summarizes clusters of episodic records into semantic/procedural records and archives the sources (FR-MEM-006). |
| Mandatory set | Context tiers 1–3 (system/skill prompts, pins, current intent); assembly refuses rather than drops them (ADR-086). |
| Compaction ladder | The ordered, non-generative shrinking steps applied to borderline context candidates before exclusion (FR-CTX-003). |
| Context snapshot | The persisted Context Item rows plus assembly metadata that make a turn's request reproducible (FR-CTX-007). |
| Index generation | The monotonic counter of committed build/update results for one index; queries are served from exactly one committed generation (FR-IDX-001). |
| Embedding space | The immutable provider+model+dimensions declaration of a semantic index (Volume 2 attribute; INV-IDX-03; ADR-089). |
| Staleness threshold | The pending-change count/age bounds beyond which a `ready` index transitions to `stale` (FR-IDX-004). |
| Chunk locator | The stable path + byte-offset + line-span reference identifying one chunk (ADR-088); the `source_ref` of `file_chunk` embeddings. |

## Assumptions

Local list per Volume 0, chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | Embedding providers return stable vectors for identical input under a pinned model name within one provider revision; same-name silent revisions are rare enough that golden retrieval fixtures catch them | Volume 13 semantic retrieval golden suites per release; provider deprecation monitoring (Volume 5) | Rebuild affected semantic indexes as new embedding spaces (RISK-IDX-002 procedure); tighten space pinning to provider-versioned identifiers where offered |
| Technical assumption | The ADR-087 heuristic (`ceil(bytes/4)+8`, ×1.15) plus the default 5% safety margin over-estimates true token counts for the supported model families' typical code-and-prose content | NFR-CTX-001 overflow classification across counted and estimated models | Recalibrate constants by superseding ADR-087; raise the default safety margin |
| Technical assumption | The ≤100,000-chunk workspace assumption of ADR-020 holds for the target user base at MVP | Chunk-count metrics per workspace (ADR-020 observability); Volume 12 benchmarks | ADR-020 review conditions trigger the ANN evaluation; `on_scale_exceeded` behavior bounds the interim |
| Product hypothesis | Assisted ingestion (agent-proposed, session-layer auto-accepted, durable layers consent-gated) is the right default authoring balance for memory | Beta feedback channels (Volume 15); refusal/confirmation telemetry under consent | Ship `explicit` as default via configuration change; no contract impact |

## Open questions

Entries follow Volume 0, chapter 08; none blocks authoring. Every PENDING VALIDATION
occurrence in this volume maps to a row here (or to the originating ADR's register entry,
noted).

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V7-OQ-1 | Local exact tokenizers per model family (availability, licenses, fidelity) — adoption PENDING VALIDATION per ADR-087 alternative 1 | ADR-087; chapter 03 (FR-CTX-002) | No — official counting + heuristic fallback are the specified paths | Per-family validation spike before any adoption; supersede/extend ADR-087 | Open |
| V7-OQ-2 | ANN acceleration (e.g., `sqlite-vec`) remains PENDING VALIDATION per ADR-020, referenced by chapter 04 and RISK-IDX-001 | ADR-020; chapter 04 | No — exact cosine is the MVP algorithm | ADR-020 review conditions (scale or latency triggers) | Open |
| V7-OQ-3 | Application-level at-rest encryption of memory databases — viability of an encryption path within ADR-007's pure-Go constraint is PENDING VALIDATION; posture until then is OS file permissions plus optional age-encrypted exports (FR-MEM-009) | Chapter 02 (encryption posture) | No — classified Future; no committed requirement depends on it | Validation spike on pure-Go-conflicting mechanisms; ADR if adopted | Open |
| V7-OQ-4 | Semantic indexing of `long_term` (global) memory: Index rows are workspace-scoped by Volume 2, so global memory retrieval is lexical-only; extending index scope to the global database would need a Volume 2 amendment | Chapter 04 (default indexes) | No — long-term memory remains fully retrievable lexically | Assess demand at Beta; propose a Volume 2 amendment through the change procedure if warranted | Open |

## Cross-volume references

Load-bearing references made by this volume, by name (requirement-level cross-links are
upgraded at consolidation):

| Referenced | Where used |
|---|---|
| MemoryStorePort, IndexerPort, ProviderPort (`Embed`, `CountTokens`, `Capabilities`), SchedulerPort, WorkspacePort, SessionStorePort (Volume 3 chapter 02, frozen) | Chapters 01–05 contract elaboration |
| Memory Record, Context Item, Index, Embedding shapes and INV-MEM/CTXI/IDX/EMB invariants (Volume 2 chapter 07); Index state enum and recorded status vocabularies (Volume 2 chapter 09) | Throughout |
| ADR-007, ADR-014, ADR-016, ADR-019, ADR-020, ADR-022, ADR-023, ADR-027, ADR-028, ADR-029 (foundation decisions) | Storage, errors, concurrency, export encryption |
| Volume 5: capability enum (`embeddings`, `chat`), provider connection states, error normalization, adapter catalog | Chapters 02–04 |
| Volume 9: redaction model and detection classes, permission model (`read`, `write` permission names), Audit Log, threat catalog (memory/index poisoning, secret exfiltration) | Chapters 01–04 security sections |
| Volume 10: event envelope and delivery semantics, configuration schema/precedence/validation, replay regime (SM-12 formalization), storage rules | Event/config/reproducibility sections |
| Volume 8: `memory`, `context`, `index` command families; TUI context/memory views; confirmation conventions | User-facing operations |
| Volume 12: latency/throughput budgets for retrieval, assembly overhead, indexing; reference hardware | Performance sections and NFR test environments |
| Volume 13: offline suite, crash-injection harnesses, golden fixtures, property suites | Verification methods |
| Volume 4: Agent Engine/Planner/Workflow Engine as assembly callers; Prompt Engine rendering; run record append discipline | Chapters 02–03 |
