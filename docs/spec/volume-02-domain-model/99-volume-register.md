# 99 — Volume 2 Register

Machine-parseable register of everything Volume 2 minted, per Volume 0 chapters 02 and 03.
Merged into the Volume 0 registers at consolidation.

## Requirements index

Volume 2 mints no `FR`, `NFR`, or `RISK` identifiers and no error codes — it is definitional.
Behavioral requirements over the entities defined here are minted by the owning volumes
(Volume 0, chapter 03 area ownership); integrity invariants are chapter-local `INV-*` labels,
referenced by volume + label, not corpus identifiers.

| ID | Title | Phase | Verification method |
|---|---|---|---|
| — | *(none minted by this volume)* | — | — |

## Entity catalog

All 43 domain entities of the Volume 0 glossary, with aggregate assignment (chapter 01),
statefulness (chapter 09), and persistence home (chapter 10). "Machine" = canonical state
enum frozen in chapter 09; "Status" = recorded status vocabulary (chapter 09, second table);
"No" = stateless or immutable record. "Per scope" = workspace or global database following
the entity's `scope` attribute (ADR-028).

| Entity | Aggregate | Stateful? | Persisted in |
|---|---|---|---|
| Workspace | Workspace (root) | No | Workspace DB `workspace` + global registry `workspaces` |
| Project | Workspace | No | Workspace DB `projects` |
| Configuration Profile | Workspace (global rows: standalone catalog) | No | Per scope, `configuration_profiles` |
| Session | Session (root) | Machine | Workspace DB `sessions` |
| Agent | Run | Machine | Workspace DB `agents` |
| Agent Profile | Agent Profile (root) | No | Per scope, `agent_profiles` (builtin: in binary) |
| Run | Run (root) | Machine | Workspace DB `runs` |
| Turn | Run | Status | Workspace DB `turns` |
| Message | Run | No | Workspace DB `messages` |
| Plan | Run | Machine | Workspace DB `plans` |
| Task | Run | Machine | Workspace DB `tasks` |
| Tool | Tool (root) | No | Per provider scope, `tools` (builtin: in binary) |
| Tool Invocation | Run | Machine | Workspace DB `tool_invocations` |
| Tool Result | Run | Status | Workspace DB `tool_results` |
| Approval | Approval (root) | Machine | Workspace DB `approvals` (global for standing global requests) |
| Permission | Permission (root) | No | Per scope, `permissions` |
| Artifact | Run | No | Workspace DB `artifacts` + content under `.andromeda/artifacts/` |
| File Change | Run | No | Workspace DB `file_changes` |
| Patch | Run | Status | Workspace DB `patches` |
| Command Execution | Run | Status | Workspace DB `command_executions` |
| Provider | Provider (root) | Machine | Global DB `providers` |
| Model | Provider | No | Global DB `models` |
| Capability | Value vocabulary (no aggregate) | No | Not persisted (enum in code, Volume 5) |
| Credential | Credential (root) | Status | Global DB only, `credentials` |
| Authentication Session | Credential | Machine | Global DB only, `auth_sessions` |
| Workflow | Workflow (root) | No | Per scope, `workflows` (builtin: in binary) |
| Workflow Run | Workflow Run (root) | Machine | Workspace DB `workflow_runs` |
| Skill | Skill (root) | No | Per scope, `skills` (builtin: in binary) |
| Plugin | Plugin (root) | Machine | Per scope, `plugins` |
| MCP Server | MCP Server (root) | No | Per scope, `mcp_servers` |
| MCP Client Connection | MCP Server | Machine | Workspace DB `mcp_connections` |
| Package | Package (root) | Machine | Per scope, `packages` |
| Extension | Extension (root) | No | Per scope, `extensions` |
| Release | Release (root) | Machine | Global DB `releases` |
| Memory Record | Memory Record (root) | Status | Workspace DB (`session`/`workspace` layers) or global DB (`long_term`), `memory_records` |
| Context Item | Run | No | Workspace DB `context_items` |
| Index | Index (root) | Machine | Workspace DB `content_indexes` (metadata); data in `.andromeda/index.db` |
| Embedding | Index | No | Index cache DB `embeddings` |
| Event | Event (root) | No | Workspace or global DB `events` (per context) |
| Trace | Trace (root) | Status | Workspace DB `traces` + `trace_spans` |
| Metric | Metric (definition) | No | Definitions in code; optional samples `metric_points` |
| Cost Record | Cost Record (root) | No | Workspace DB `cost_records` |
| Audit Record | Audit Record (root) | No | Workspace or global DB `audit_records` (per context) |

## Canonical state enums

Frozen in chapter 09 for the whole corpus; owning volumes define the full machines with
exactly these names. Terminal states are marked `*`.

| Entity | States | Full machine owner |
|---|---|---|
| Session | `created`, `active`, `suspended`, `ended`\*, `failed`\* | Volume 4 |
| Run | `pending`, `planning`, `running`, `awaiting_approval`, `paused`, `interrupted`, `completed`\*, `failed`\*, `cancelled`\* | Volume 4 |
| Plan | `draft`, `proposed`, `approved`, `executing`, `revising`, `completed`\*, `superseded`\*, `abandoned`\* | Volume 4 |
| Task | `pending`, `ready`, `running`, `blocked`, `awaiting_approval`, `interrupted`, `completed`\*, `failed`\*, `cancelled`\*, `skipped`\* | Volume 4 |
| Agent | `instantiated`, `idle`, `thinking`, `acting`, `waiting`, `terminated`\*, `failed`\* | Volume 4 |
| Workflow Run | `pending`, `running`, `awaiting_approval`, `paused`, `interrupted`, `completed`\*, `failed`\*, `cancelled`\* | Volume 4 |
| Tool Invocation | `requested`, `awaiting_approval`, `approved`, `executing`, `succeeded`\*, `failed`\*, `denied`\*, `timed_out`\*, `cancelled`\* | Volume 6 |
| Plugin | `registered`, `starting`, `running`, `stopping`, `stopped`, `failed`, `disabled`, `removed`\* | Volume 6 |
| MCP Client Connection | `configured`, `connecting`, `initializing`, `ready`, `reconnecting`, `disconnected`, `failed`, `disabled`, `removed`\* | Volume 6 |
| Package (installation) | `requested`, `resolving`, `downloading`, `verifying`, `staged`, `installing`, `installed`, `failed`, `rolled_back`, `removed`\* | Volume 6 |
| Approval | `requested`, `granted`\*, `denied`\*, `expired`\*, `cancelled`\* | Volume 9 |
| Authentication Session | `unauthenticated`, `authenticating`, `active`, `refreshing`, `expired`, `failed`, `revoked`\* | Volume 5 |
| Provider (connection) | `configured`, `verifying`, `available`, `degraded`, `unavailable`, `disabled`, `removed`\* | Volume 5 |
| Index | `created`, `building`, `ready`, `updating`, `stale`, `failed`, `removed`\* | Volume 7 |
| Release | `drafted`, `candidate`, `published`, `superseded`, `yanked`\* | Volume 14 |
| Update (process) | `checking`, `up_to_date`\*, `update_available`, `downloading`, `verifying`, `applying`, `applied`\*, `failed`\*, `rolled_back`\* | Volume 14 |

Recorded status vocabularies (chapter 09, not machines): Turn `status`; Tool Result `status`;
Patch `status`; Command Execution `outcome`; Credential `status`; Memory Record `status`;
Trace `status`.

## ADRs minted

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-027](../annexes/adr/ADR-027.md) | Entity identifiers: ULIDs as primary IDs | Accepted | Every persisted entity row has a ULID primary key — 26-char uppercase Crockford base32, monotonic per process, never semantic beyond identity and creation-time ordering |
| [ADR-028](../annexes/adr/ADR-028.md) | Database topology: one workspace-local database plus one global database | Accepted | Authoritative state splits into `.andromeda/state.db` (per workspace) and `<data_dir>/andromeda/global.db` (per machine); credentials global-only; rebuildable caches in separate cache databases |
| [ADR-029](../annexes/adr/ADR-029.md) | Schema evolution: forward-only migrations, no downgrade support | Accepted | Numbered immutable forward migrations with pre-migration backup and integrity checks; failures and future-schema encounters refuse cleanly with exit code 9; recovery via backup restore |

## Glossary additions

| Term | One-line meaning |
|---|---|
| Aggregate | A consistency boundary of entities with one root; intra-aggregate invariants commit in one transaction, cross-aggregate references are by ULID only (Volume 2, chapter 01). |
| Aggregate root | The entity that owns an aggregate's identity and lifecycle; deleting it deletes or tombstones its members (Volume 2, chapter 01). |
| Record entity | An append-only, immutable entity (messages, events, audit records, …); mutation of history is a defect (Volume 2, chapter 01). |
| Catalog entity | A long-lived registry-style entity with natural-key uniqueness and in-place updates tracked by a `revision` counter (Volume 2, chapter 01). |
| Canonical state enum | The frozen state-name set of a stateful entity, owned by Volume 2 chapter 09; full machines are owned by the entity's area volume. |
| Recorded status | A frozen outcome/classification vocabulary on an entity attribute without a governed state machine (Volume 2, chapter 09). |
| ULID | Universally Unique Lexicographically Sortable Identifier — 48-bit millisecond timestamp + 80 random bits, canonically 26 characters of uppercase Crockford base32 (ADR-027). |
| Workspace database | The authoritative SQLite database at `.andromeda/state.db`, one per Workspace, holding all workspace-scoped state (ADR-028). |
| Global database | The authoritative SQLite database at `<data_dir>/andromeda/global.db`, one per user per machine, holding machine-level state including all credential metadata (ADR-028). |
| Index cache database | The non-authoritative, rebuildable SQLite database at `.andromeda/index.db` holding index data and embeddings; its loss is never an integrity error (Volume 2, chapters 07/10). |
| Canonical JSON | Volume 2 chapter 10's serialization form: UTF-8, sorted `snake_case` keys, omitted absent optionals, RFC 3339 UTC timestamps, `schema_version` field — the reproducible input for hashing and export. |
| Tombstone | A retained entity row whose content or subject was removed, preserving attribution and hashes after pruning or unregistration (Volume 2, chapters 04/06/10). |
| Secret reference | An opaque handle (`secret_ref`, `token_ref`) into the Secret Store; the only representation of secret material the domain model permits (Volume 2, chapter 05). |

## Assumptions

Local list per Volume 0, chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | The SQLite build shipped via `modernc.org/sqlite` (ADR-007) provides the JSON functions and pragmas chapter 10 relies on (`json_valid`, `user_version`, `integrity_check`, `foreign_key_check`) | Persistence Layer conformance tests (Volume 13) against the pinned driver | Affected `CHECK` constraints move into Persistence Layer validation code; conventions stand |
| Technical assumption | Per-turn persistence of Context Item rows (including excluded candidates) stays within the storage and latency budgets Volume 12 sets | Volume 12 storage/latency benchmarks on representative runs | Excluded-candidate persistence becomes opt-in diagnostics (model already permits earlier pruning, chapter 07) |
| Technical assumption | Embedding vectors stored as BLOBs in the index cache database support workspace-scale semantic retrieval within Volume 12 budgets, without a dedicated vector store | Volume 7 retrieval design + Volume 12 benchmarks | Index cache moves to a different on-disk layout; authoritative model is unaffected (INV-IDX-02 isolates it) |
| Technical assumption | A process-monotonic ULID source with crypto-grade randomness is implementable in pure Go within ADR-001 constraints | Generator conformance tests (Volume 13) | Identifier contract (ADR-027 rules 1–3) is preserved by an alternative generator design; format unchanged |

## Open questions

Entries follow Volume 0, chapter 08; none blocks authoring — each names its stable abstraction.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V2-OQ-1 | Will the MCP transport vocabulary (`stdio`, `streamable_http`) need extension as the MCP specification evolves? | Chapter 06, MCP Server | No — the enum is closed and extended via the change procedure | Volume 6 tracks the official MCP specification and proposes vocabulary changes as ADR-gated amendments | Open |
| V2-OQ-2 | Should excluded-candidate Context Items persist by default or only in a diagnostics mode? | Chapter 07, Context Item | No — the model supports both; chapter 07 allows earlier pruning of excluded rows | Volume 7 decides the default with Volume 12 budget data | Open |
| V2-OQ-3 | Is local persistence of metric samples (`metric_points`) enabled at MVP, or export-only aggregation? | Chapter 08, Metric; chapter 10 map | No — the definition shape is fixed either way | Volume 10 decides with the observability pipeline design | Open |
| V2-OQ-4 | Does long-term memory need per-identity partitioning in the global database (multiple OS users are already separate; multiple personas per user are not)? | Chapter 07, Memory Record | No — `long_term` layer semantics are Volume 7's to refine | Volume 7 memory taxonomy design | Open |
