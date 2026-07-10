# 10 — Persistence, Serialization, and Migrations

This chapter binds the domain model to storage and wire formats: where entities persist
(ADR-028), how tables and columns are shaped over SQLite per ADR-007, how entities serialize
as canonical JSON, and how schemas evolve (ADR-029). The Persistence Layer component contract
and operational storage behavior (backups, locking, size budgets, retention execution) are
owned by Volume 10; this chapter owns the *model-facing* conventions every schema and
serialization MUST follow.

## Databases

Decided by [ADR-028](../annexes/adr/ADR-028.md): Andromeda persists authoritative state in
exactly **two databases per installation-plus-workspace pair**, plus non-authoritative cache
databases.

| Database | Location | Scope | Holds |
|---|---|---|---|
| Workspace database | `.andromeda/state.db` (workspace root, ADR-022) | One per Workspace | Everything scoped to the workspace: the Workspace row, projects, sessions, runs and all Run aggregate members, workspace-scoped catalogs, workspace events/traces/cost/audit records, session/workspace memory |
| Global database | `<data_dir>/andromeda/global.db`, where `<data_dir>` is the platform data directory resolved per ADR-022 | One per user per machine | Machine-level state: workspace registry, providers and models, credentials and authentication sessions, global-scoped catalogs, releases and update history, long-term memory, global events/audit records |
| Index cache database | `.andromeda/index.db` | One per Workspace | Index data and embeddings — rebuildable, non-authoritative (INV-IDX-02) |

Rules:

1. Workspace state MUST NOT be written to the global database, and vice versa: the workspace
   database is self-contained so that copying a workspace directory moves its full history,
   and deleting `.andromeda/` severs nothing global except the registry row.
2. Credentials, authentication sessions, and their metadata live **only** in the global
   database (never workspace: INV-CRED-01 context — workspace exports must be structurally
   unable to include credential metadata).
3. Scoped catalog entities (Configuration Profile, Agent Profile, Workflow, Skill, Plugin,
   MCP Server, Package, Extension) persist in the database matching their `scope` attribute,
   under the same table name in both databases.
4. Cache databases carry no authoritative state; the Runtime MUST treat their absence or
   corruption as a rebuild trigger, never as data loss (exit code 9 does not apply to cache
   loss).
5. Both authoritative databases are opened in WAL mode via the DSN pragma per ADR-007, with
   foreign-key enforcement enabled per connection.

## Entity-to-table map

| Entity | Database | Table |
|---|---|---|
| Workspace | workspace (identity row) + global (registry) | `workspace` / `workspaces` |
| Project | workspace | `projects` |
| Configuration Profile | per `scope` | `configuration_profiles` |
| Session | workspace | `sessions` |
| Agent Profile | per `scope` (builtin: in binary) | `agent_profiles` |
| Run | workspace | `runs` |
| Agent | workspace | `agents` |
| Turn | workspace | `turns` |
| Message | workspace | `messages` |
| Plan | workspace | `plans` |
| Task | workspace | `tasks` |
| Tool | per provider scope (builtin: in binary) | `tools` |
| Tool Invocation | workspace | `tool_invocations` |
| Tool Result | workspace | `tool_results` |
| Approval | workspace (global for standing global requests) | `approvals` |
| Permission | per `scope` | `permissions` |
| Artifact | workspace (content under `.andromeda/artifacts/`) | `artifacts` |
| File Change | workspace | `file_changes` |
| Patch | workspace | `patches` |
| Command Execution | workspace | `command_executions` |
| Provider | global | `providers` |
| Model | global | `models` |
| Capability | not persisted (enum in code) | — |
| Credential | global only | `credentials` |
| Authentication Session | global only | `auth_sessions` |
| Workflow | per `scope` (builtin: in binary) | `workflows` |
| Workflow Run | workspace | `workflow_runs` |
| Skill | per `scope` (builtin: in binary) | `skills` |
| Plugin | per `scope` | `plugins` |
| MCP Server | per `scope` | `mcp_servers` |
| MCP Client Connection | workspace | `mcp_connections` |
| Package | per `scope` | `packages` |
| Extension | per `scope` | `extensions` |
| Release | global | `releases` |
| Memory Record | workspace (`session`/`workspace` layers) or global (`long_term`) | `memory_records` |
| Context Item | workspace | `context_items` |
| Index | workspace (metadata) | `content_indexes` |
| Embedding | index cache | `embeddings` |
| Event | workspace or global (per context) | `events` |
| Trace | workspace | `traces` (+ `trace_spans`) |
| Metric | definitions in code; optional local samples | `metric_points` (workspace) |
| Cost Record | workspace | `cost_records` |
| Audit Record | workspace or global (per context) | `audit_records` |

## SQLite schema conventions

These conventions bind every table in both authoritative databases. Volume 10 publishes the
full DDL; conformance to this section is checkable mechanically.

### Naming

1. Table names: `snake_case`, plural (`sessions`, `tool_invocations`); the single-row
   workspace identity table is deliberately singular (`workspace`).
2. Column names: `snake_case`, matching the attribute names in chapters 02–08 exactly.
3. Foreign-key columns: `<entity>_id` (e.g., `session_id`); self-references and role-named
   references keep their attribute names (`parent_agent_id`, `retry_of_id`).
4. Index names: `idx_<table>_<columns>`; unique constraints: `uq_<table>_<columns>`.

### Types and columns

| Model type (chapter 01) | SQLite storage | Convention |
|---|---|---|
| `ulid` | `TEXT` | 26-char uppercase canonical form; `CHECK (length(col) = 26)` on primary keys |
| `string`, `text`, `path`, `semver`, `hash` | `TEXT` | UTF-8; hashes lowercase hex |
| `integer`, `duration_ms` | `INTEGER` | 64-bit |
| `boolean` | `INTEGER` | 0/1 with `CHECK (col IN (0, 1))` |
| `timestamp` | `TEXT` | RFC 3339 UTC with milliseconds and `Z` suffix (e.g., `2026-07-11T08:30:00.000Z`); fixed width makes lexicographic order equal time order |
| `enum` | `TEXT` | `CHECK (col IN (...))` listing the frozen vocabulary (chapter 09 for states) |
| `json` | `TEXT` | Canonical JSON (below); `CHECK (json_valid(col))` |
| `blob` | `BLOB` | Opaque |

Column rules:

1. Every table's primary key is `id TEXT` (ULID) — no `AUTOINCREMENT` surrogate keys for
   entities. The implicit `rowid` remains and MAY serve as the intra-table insertion-order
   tiebreaker (chapter 01, identifier rule 3).
2. Timestamps are always UTC RFC 3339; local time never enters the database. Column names end
   in `_at`.
3. Mutable entities carry `revision INTEGER NOT NULL DEFAULT 1`; the Persistence Layer
   increments it on every committed update and uses it for optimistic concurrency (stale
   writes fail; retry semantics in Volume 10).
4. Foreign keys are declared with explicit actions mirroring the aggregate map: `ON DELETE
   CASCADE` inside an aggregate (e.g., `turns.run_id → runs.id`), `ON DELETE RESTRICT`
   across aggregates where invariants forbid orphaning (e.g., a Configuration Profile
   referenced as a default), and nullable `ON DELETE SET NULL` only where the model
   explicitly allows tombstoned references.
5. JSON columns are for flexible, schema-versioned payloads (`parts`, `payload`,
   `definition`, `step_states`, …) — never for data the schema queries relationally. A value
   that acquires a relational access pattern is promoted to columns by a migration.
6. Monetary amounts follow INV-COST-02: `INTEGER` micro-units plus an ISO 4217 `TEXT`
   currency column; `REAL` never stores money.

### Write discipline

1. All access goes through the Persistence Layer (ADR-007, decision rule 3); no component
   embeds SQL against the store directly.
2. One aggregate, one transaction: an aggregate-consistent change set commits atomically
   (chapter 01, rule 2 allows same-commit Record appends).
3. Execution progress persists incrementally — at minimum at every canonical state
   transition and every Record append (PRD-010); batching windows are Volume 12's budgets.

## Serialization

### Canonical JSON

Every entity export, event payload, and hash-input serialization uses **canonical JSON**:

1. UTF-8, no byte-order mark; object keys in `snake_case` exactly matching attribute names.
2. Object keys sorted lexicographically (byte order) at every nesting level — required so
   hashes over serializations (INV-AUD-02, `content_hash` fields) are reproducible.
3. Optional attributes that are absent are **omitted**, never `null`.
4. Timestamps serialize as their RFC 3339 text form; ULIDs as canonical 26-char strings;
   binary as lowercase hex for hashes and standard base64 for embedded content; integers
   never as floats (no exponent notation).
5. Every serialized document carries `schema_version` (integer ≥ 1), versioning that
   document kind's shape. Event payload versions are per event name (INV-EVT-01 context);
   entity export versions are per entity kind.

### Export forms

| Form | Format | Used for |
|---|---|---|
| Entity document | Single canonical JSON object | One entity (with `schema_version`, no embedded aggregates) |
| Record stream | JSON Lines (one canonical document per line) | Run records, session exports, event streams, audit exports |
| Human rendering | TOML (ADR-008) | Configuration Profile and Agent Profile editing views — presentation only; the persisted and exchanged form is JSON |

Exports never include: secret material or Secret Store references beyond opaque IDs
(INV-CRED-01, INV-AUTHS-02), un-redacted content past Volume 9 gates, or cache-database
contents (embeddings, index data). Import validation is owned by Volume 10.

## Schema versioning and migrations

Decided by [ADR-029](../annexes/adr/ADR-029.md): **forward-only migrations, no downgrade
support; recovery is via backup.**

### Versioning

1. Each authoritative database carries a schema version: a monotonically increasing integer,
   stored both in `PRAGMA user_version` and as rows in the `schema_migrations` table
   (`version INTEGER PRIMARY KEY`, `applied_at TEXT`, `checksum TEXT` — the SHA-256 of the
   migration script applied). The two MUST agree; disagreement is an integrity error.
2. The workspace and global databases version independently (they ship in one binary but
   migrate separately — a binary may touch many workspace databases of different ages).
3. Migration scripts are numbered, immutable, and embedded in the binary; editing a shipped
   migration is forbidden (the checksum makes it detectable).

### Applying migrations

On opening a database whose version is lower than the binary's target, the Persistence Layer
MUST:

1. Take a pre-migration backup copy of the database file (location and rotation owned by
   Volume 10) and verify the backup's integrity before proceeding.
2. Run `PRAGMA integrity_check` and `PRAGMA foreign_key_check`; refuse to migrate a database
   that fails them (integrity error, exit code 9 per ADR-016 and the Volume 0 exit-code
   table).
3. Apply each pending migration in order, each inside a transaction, recording its
   `schema_migrations` row in the same transaction. A DDL step that SQLite cannot roll back
   cleanly is structured per Volume 10's migration playbook (new-table-copy-swap), so failure
   never leaves a half-migrated schema visible.
4. Re-run both checks after the final migration; update `PRAGMA user_version` last.
5. On any failure: stop, leave the database file as-is alongside the untouched backup, report
   the stable error with the failing migration number, and exit with code 9. The recovery
   path is restore-from-backup (automatic where Volume 10 says so, otherwise user-invoked);
   there is no downgrade migration, ever.

### Compatibility rules

1. An older binary opening a **newer** database MUST refuse cleanly (integrity/compatibility
   error, exit code 9), naming both versions; it MUST NOT attempt partial reads
   (forward-only means data written by the future is out of contract).
2. A newer binary MUST open any database whose version its migration chain reaches, however
   old (no minimum-hop requirement within one major product line; Volume 14 governs
   cross-major paths via `min_upgrade_from`).
3. Serialization `schema_version` values evolve additively where possible (new optional
   fields); a breaking payload change increments the version, and readers MUST reject
   versions above what they know rather than guess.
4. Cache databases (`index.db`) do not migrate: on layout change
   (`index_schema_version` bump) they are dropped and rebuilt (INV-IDX-02).

## Retention pointers

This volume defines *what* persists and where; retention **policies** (windows, triggers,
quotas, purge procedures) are owned elsewhere and referenced here once:

| Data | Retention owner |
|---|---|
| Memory Records, Embeddings, Index data | Volume 7 |
| Audit Records, Approvals, Permissions (audit precedence, INV-AUD-04) | Volume 9 |
| Sessions, Runs and members, Events, Traces, Cost Records, Artifacts content, backups | Volume 10 |
| Releases, update history | Volume 14 |

One cross-cutting rule is fixed by the model itself: retention MUST respect the invariants of
chapters 02–08 — audit precedence (INV-AUD-04), tombstoning over deletion where records
demand attribution (INV-ART-03, INV-TOOL-04), and cascade completeness on purge (INV-MEM-04,
INV-EMB-03).
