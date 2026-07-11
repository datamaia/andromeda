# Volume 7 — Memory, Context, and Indexing

**Status:** Complete · **Owner:** Memory Manager / Context Manager / Indexing Engine (Volume 7)

Volume 7 specifies Andromeda's knowledge subsystem: the memory model and its full lifecycle
(Memory Manager behind MemoryStorePort), the Context Manager that selects, budgets, and
records exactly what each model request contains, and the Indexing Engine (behind
IndexerPort) that builds and incrementally maintains lexical and semantic indexes — including
the full Index state machine. Per Volume 0, chapter 03, this volume mints all `MEM`, `CTX`,
and `IDX` identifiers, the `[memory]`/`[context]`/`[index]` configuration key content, the
`memory.*`/`context.*`/`index.*` event names, and ADRs 085–089.

Foundations assumed: Volume 0 (conventions), Volume 1 (objectives, phases, offline
guarantees), Volume 2 (Memory Record, Context Item, Index, Embedding shapes; frozen Index
states; recorded status vocabularies), Volume 3 (MemoryStorePort and IndexerPort signatures;
component boundaries; ADR-020/028 storage topology).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Memory Model](01-memory-model.md) | Two-axis taxonomy (layer × kind, ADR-085), visibility and isolation rules, FR-MEM-001 (memory model), provenance and trust (FR-MEM-002), secret exclusion (FR-MEM-003), NFR-MEM-001, RISK-MEM-001 |
| [02 — Memory Lifecycle](02-memory-lifecycle.md) | Ingestion/normalization, dedup/conflict/supersession versioning, consolidation, retention/expiration, deletion cascade, export, offline operation (FR-MEM-004..010), NFR-MEM-002, RISK-MEM-002, E-MEM catalog, `memory.*` events, `[memory]` keys |
| [03 — Context Manager](03-context-manager.md) | Sources and priority tiers (ADR-086), assembly (FR-CTX-001), token budgeting (FR-CTX-002, ADR-087), dedup/compression (FR-CTX-003), freshness/trust/conflicts (FR-CTX-004), pinning/exclusion (FR-CTX-005), tool-result/large/binary handling (FR-CTX-006), snapshots and reproducibility (FR-CTX-007), NFR-CTX-001, RISK-CTX-001, E-CTX catalog, `context.*` events, `[context]` keys |
| [04 — Indexing Engine](04-indexing-engine.md) | Default indexes, scope/ignore rules, engine contract (FR-IDX-001), chunking (FR-IDX-002, ADR-088), embeddings (FR-IDX-003, ADR-020/089), incremental updates and invalidation (FR-IDX-004), operation without embeddings/Internet (FR-IDX-005), NFR-IDX-001/002, RISK-IDX-001/002, E-IDX catalog, `index.*` events, `[index]` keys |
| [05 — Index State Machine](05-index-state-machine.md) | The full Index machine over the frozen states: transition table T1–T13, twelve machine elements, queryability by state, FR-IDX-006 |
| [99 — Volume Register](99-volume-register.md) | Everything this volume minted: requirements, ADRs, error codes, events, config keys, glossary additions, assumptions, open questions |

## Reading guide

1. Chapter 01 fixes vocabulary every other chapter uses (layers, kinds, trust ordering).
2. Chapters 02 and 04 are the behavioral contracts behind MemoryStorePort and IndexerPort
   respectively; Volume 3 chapter 02 holds the frozen signatures they elaborate.
3. Chapter 03 consumes both and is the sole authority on what enters a model request;
   Volume 4's engines call it, Volume 8 renders its records.
4. Chapter 05 is normative for every state the word "index" carries anywhere in the corpus.
