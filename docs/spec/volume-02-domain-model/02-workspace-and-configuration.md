# 02 — Workspace and Configuration

This chapter defines the Workspace aggregate: **Workspace** (root), **Project**, and
**Configuration Profile**. Behavior over these entities is owned elsewhere: workspace
discovery and management by the Workspace Engine (Volume 4 behavior, Volume 3 component
boundaries), configuration schema and precedence by Volume 10, workspace trust by Volume 9.

## Workspace

Purpose: the root working environment Andromeda operates in — a directory tree plus its
Andromeda state (`.andromeda/`), settings, and indexes. The workspace database *is* the
durable identity of a Workspace (ADR-028).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key; minted when `.andromeda/` is first initialized |
| `root_path` | `path` | yes | Absolute path of the workspace root directory on this machine |
| `name` | `string` | yes | Display name; defaults to the root directory basename |
| `trust_state` | `enum` | yes | Workspace trust classification; vocabulary and semantics owned by Volume 9 |
| `default_profile_id` | `ulid` | no | Workspace-scoped Configuration Profile applied by default |
| `settings` | `json` | no | Workspace-level settings not represented as Configuration Profiles; keys owned by Volume 10 |
| `last_opened_at` | `timestamp` | no | Last time this workspace was opened by any Andromeda process |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`.
- Natural key: none inside the workspace database (there is exactly one row); in the global
  database's workspace registry, `root_path` is unique per machine.
- External identifiers: none. The root path is location, not identity — moving a workspace
  directory preserves `id` (INV-WS-04).

### Relations

- Contains 0..n **Project** (composition; deleted with the workspace).
- Owns 0..n workspace-scoped **Configuration Profile** and other workspace-scoped catalog
  entries (Agent Profile, Workflow, Skill, Plugin, MCP Server registrations — see their
  chapters).
- Hosts 0..n **Session** (Sessions reference `workspace_id`; they are their own aggregate).
- Indexed by 0..n **Index**; stores 0..n **Memory Record** of workspace scope (chapter 07).
- Registered as one row in the global database's workspace registry (chapter 10).

### Integrity invariants

1. **INV-WS-01** — Exactly one Workspace row MUST exist per workspace database. The database
   file at `.andromeda/state.db` and the Workspace entity are one-to-one.
2. **INV-WS-02** — `root_path` MUST be the absolute path of the directory that directly
   contains `.andromeda/`.
3. **INV-WS-03** — Two Workspaces MUST NOT share a root path on the same machine; the global
   workspace registry enforces uniqueness of `root_path`.
4. **INV-WS-04** — `id` is stable for the life of the workspace: re-opening, moving, or
   renaming the workspace directory MUST NOT re-mint it. Re-initializing `.andromeda/` after
   deletion creates a new Workspace with a new `id`.
5. **INV-WS-05** — `trust_state` MUST hold a value from the Volume 9 trust vocabulary;
   Andromeda MUST NOT execute side-effecting tools in a workspace whose trust evaluation
   (Volume 9) forbids it.

### Lifecycle

Stateless (no canonical state machine). A Workspace exists from initialization until its
`.andromeda/` directory is removed. `trust_state` is a recorded classification owned by
Volume 9, not a state machine of this volume.

### Persistence

Workspace database, table `workspace` (single row) — plus one registry row in the global
database's `workspaces` table carrying `id`, `root_path`, `name`, `last_opened_at` for
discovery and `andromeda workspace list`-style features (Volume 8). The registry row is a
cache: the workspace database is authoritative, and the registry is rebuilt from it on open
(chapter 10). Retention: workspace state lives as long as the workspace; no automatic expiry.

### Versioning and serialization

Row versioning via `revision`. The workspace database carries the schema version of chapter 10.
On export, a Workspace serializes as a canonical JSON document with `schema_version`; the
export never embeds member entities (they export as separate documents referencing
`workspace_id`).

## Project

Purpose: a logical unit inside a workspace with its own configuration profile and metadata —
commonly a repository, or one module of a monorepo.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `workspace_id` | `ulid` | yes | Owning Workspace |
| `name` | `string` | yes | Display name, unique within the workspace |
| `root_path` | `path` | yes | Project root, relative to the workspace root (`.` for a single-project workspace) |
| `vcs_kind` | `enum` | yes | `git` \| `none`; vocabulary extended only through the change procedure (Git behavior in Volume 11) |
| `default_profile_id` | `ulid` | no | Project-default Configuration Profile |
| `metadata` | `json` | no | Descriptive metadata (e.g., detected languages, marker files); produced by the Workspace Engine, advisory only |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`.
- Natural keys: `(workspace_id, name)` and `(workspace_id, root_path)` are each unique.
- External identifiers: when `vcs_kind = git`, the repository's remote URLs and commit hashes
  appear in Volume 11 records; they are never Project identity.

### Relations

- Belongs to exactly one **Workspace** (aggregate member).
- May select at most one **Configuration Profile** as project default.
- Referenced by **Session** (`project_id`, optional focus) and by **Index** scopes
  (chapter 07).

### Integrity invariants

1. **INV-PRJ-01** — A Project MUST belong to exactly one Workspace and cannot move between
   workspaces.
2. **INV-PRJ-02** — The project root (workspace root joined with `root_path`) MUST resolve to
   a directory inside the workspace root; path escapes (`..`, symlinks resolving outside) are
   invalid.
3. **INV-PRJ-03** — Project roots MUST NOT be nested inside one another within the same
   workspace; overlapping project scopes are a configuration error (Volume 10 validation).
4. **INV-PRJ-04** — Deleting a Project MUST NOT delete user files; it removes only Andromeda's
   Project row and project-scoped catalog entries.

### Lifecycle

Stateless. Created by workspace discovery or explicitly; removed explicitly.

### Persistence

Workspace database, table `projects`. Retention: lives as long as the workspace or until
explicitly removed.

### Versioning and serialization

Row versioning via `revision`; standard canonical JSON export.

## Configuration Profile

Purpose: a named set of configuration values selectable at global, project, or invocation
level (glossary). The configuration *schema*, value precedence, validation, and migration
rules are owned by Volume 10; this entity is the persisted carrier of one named value set.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `name` | `string` | yes | Profile name, unique within its scope owner |
| `scope` | `enum` | yes | `global` \| `workspace` \| `project` |
| `workspace_id` | `ulid` | conditional | Required when `scope` is `workspace` or `project`; absent for `global` |
| `project_id` | `ulid` | conditional | Required when `scope` is `project` |
| `description` | `text` | no | Human description |
| `values` | `json` | yes | The configuration values, as a canonical JSON projection of TOML content (ADR-008); keys MUST exist in the Volume 10 schema |
| `source_path` | `path` | no | When the profile mirrors a file (e.g., a checked-in profile), the file it was loaded from |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`.
- Natural key: `(scope, workspace_id, project_id, name)` unique — i.e., a name is unique among
  profiles visible at the same level.

### Relations

- `workspace`/`project` scoped profiles belong to the **Workspace** aggregate; `global` ones
  are global catalog rows.
- Selected by **Workspace** (`default_profile_id`), **Project** (`default_profile_id`), and
  referenced by **Run** configuration snapshots (chapter 03).

### Integrity invariants

1. **INV-CFGP-01** — Every key in `values` MUST be a key defined by the Volume 10
   configuration schema; unknown keys fail validation (Volume 10 owns the validation rules).
2. **INV-CFGP-02** — `values` MUST NOT contain secret material. Configuration references
   secrets only as Secret Store references or environment indirection per Volumes 9 and 10.
3. **INV-CFGP-03** — Scope fields MUST be consistent: `global` profiles have neither
   `workspace_id` nor `project_id`; `workspace` profiles have only `workspace_id`; `project`
   profiles have both.
4. **INV-CFGP-04** — A Configuration Profile is a *named layer* in the Volume 10 precedence
   chain; it MUST NOT itself define precedence. Two profiles are combined only by that chain,
   never by profile-to-profile inheritance fields in this entity.

### Lifecycle

Stateless.

### Persistence

`global` profiles: global database, table `configuration_profiles`. `workspace` and `project`
profiles: workspace database, same table name. Retention: until explicitly deleted; deletion
is refused while a Workspace or Project names the profile as its default (restrict action,
chapter 10).

### Versioning and serialization

Row versioning via `revision`. Profiles serialize as canonical JSON; when exported for humans
they render as TOML per ADR-008 (rendering is presentation, the persisted form is JSON).
Configuration schema migrations (renamed keys) are owned by Volume 10 and update `values`
in place with the migration recorded as an Event.
