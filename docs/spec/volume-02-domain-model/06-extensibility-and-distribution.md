# 06 — Extensibility and Distribution

This chapter defines the orchestration and extension aggregates: **Workflow** and **Workflow
Run** (behavior owned by Volume 4), **Skill**, **Plugin**, **MCP Server** with **MCP Client
Connection**, and the distribution chain **Package**, **Extension**, **Release** (behavior
owned by Volumes 6 and 14). These entities realize the open-platform objective (PRD-007):
third-party capability enters the system only through these shapes, under the same contracts
as built-in capability.

## Workflow

Purpose: a declared, stateful orchestration of agents, tools, and approvals — states,
transitions, gates, and artifacts written down as a versioned definition (PRD-012).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key (one row per definition *version*) |
| `name` | `string` | yes | Workflow name (e.g., `spec-driven-dev`) |
| `version` | `semver` | yes | Definition version |
| `scope` | `enum` | yes | `builtin` \| `global` \| `workspace` |
| `workspace_id` | `ulid` | conditional | Required when `scope = workspace` |
| `description` | `text` | yes | What the workflow does |
| `definition` | `json` | yes | Parsed definition: steps, transitions, agent profile selectors, tool requirements, approval gates, artifacts (definition language owned by Volume 4) |
| `source_ref` | `path` | no | Source document the definition was loaded from |
| `source_hash` | `hash` | yes | SHA-256 of the source/definition for integrity |
| `inputs_schema` | `json` | no | JSON Schema of workflow inputs |
| `required_capabilities` | `json` | no | Capability values the workflow needs (Principle 2 gating) |
| `deprecated` | `boolean` | yes | Version discouraged for new runs |
| `enabled` | `boolean` | yes | Version may be instantiated |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change (metadata only; INV-WFD-02) |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(scope, workspace_id, name, version)` unique.

### Relations

- Executed as 0..n **Workflow Run**.
- References **Agent Profile**, **Tool**, and **Skill** requirements by selector (not FK).
- May be delivered by a **Package** (as an Extension of kind `workflow`).

### Integrity invariants

1. **INV-WFD-01** — `(scope, workspace_id, name, version)` MUST be unique.
2. **INV-WFD-02** — A published definition version is immutable (`definition`,
   `inputs_schema`, `required_capabilities`, `source_hash`); behavior changes mint a new
   `version` (mirror of INV-AGP-02).
3. **INV-WFD-03** — Every step of `definition` that produces side effects MUST name its
   permission requirements; a definition that requests no permissions can gate nothing and is
   validated accordingly (validation rules owned by Volume 4/9).
4. **INV-WFD-04** — A version referenced by any persisted Workflow Run MUST NOT be deleted;
   deprecate instead (reproducibility).

### Lifecycle

Stateless versioned catalog (`enabled`/`deprecated` are recorded flags). The runtime machine
belongs to Workflow Run.

### Persistence

Per scope: global database or workspace database, table `workflows`; `builtin` definitions
ship in the binary and are materialized at startup. Retention: versions persist while
referenced (INV-WFD-04).

### Versioning and serialization

Semantic `version` for the definition; `revision` for row metadata. Definitions serialize as
canonical JSON; human-authored source format is owned by Volume 4.

## Workflow Run

Purpose: one execution instance of a workflow — resumable, step-tracked, with a complete
history (PRD-010, PRD-012).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `workflow_id` | `ulid` | yes | Definition version being executed (immutable snapshot reference) |
| `session_id` | `ulid` | yes | Owning Session |
| `state` | `enum` | yes | Canonical Workflow Run state (chapter 09) |
| `inputs` | `json` | no | Inputs the run was started with (validated against `inputs_schema`) |
| `current_step` | `string` | no | Definition step identifier currently active |
| `step_states` | `json` | yes | Per-step status map (step id → status + timestamps + spawned run/approval IDs); shape owned by Volume 4 |
| `outputs` | `json` | no | Declared workflow outputs on completion |
| `error` | `json` | no | Terminal error summary when `state = failed` |
| `started_at` | `timestamp` | yes | First activation |
| `ended_at` | `timestamp` | no | Terminal transition instant |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`.

### Relations

- Executes exactly one **Workflow** version; belongs to exactly one **Session**.
- Spawns 0..n **Run** (each Run carries `workflow_run_id`; Runs remain their own aggregates).
- Gated by 0..n **Approval** (subject kind `workflow_gate`).

### Integrity invariants

1. **INV-WFR-01** — A Workflow Run MUST reference an immutable Workflow version; the version
   in force never changes mid-run (upgrades apply to new runs only).
2. **INV-WFR-02** — `step_states` MUST cover every step the run has entered; a step that
   spawned Runs or Approvals lists their IDs — the history is navigable in both directions.
3. **INV-WFR-03** — Step progression MUST be persisted before the next step starts, so that
   recovery resumes at a step boundary and never re-executes a completed gate (PRD-010;
   recovery owned by Volume 4).
4. **INV-WFR-04** — An approval gate step MUST NOT be marked passed without a terminal
   `granted` Approval recorded for it (PRD-005, PRD-012).

### Lifecycle

Stateful — canonical states `pending`, `running`, `awaiting_approval`, `paused`,
`interrupted`, `completed`, `failed`, `cancelled` (chapter 09; deliberately the same family
as Run); full machine owned by Volume 4.

### Persistence

Workspace database, table `workflow_runs`. Retention: with the owning session; audit
precedence applies.

### Versioning and serialization

Row versioning via `revision`. Exports as a canonical JSON document referencing spawned run
records by ID.

## Skill

Purpose: a packaged, versioned unit of procedural knowledge — prompts plus required tools and
capabilities — loadable by agents (Skill Engine, Volume 6).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key (one row per skill version per scope) |
| `name` | `string` | yes | Skill name |
| `version` | `semver` | yes | Skill version |
| `scope` | `enum` | yes | `builtin` \| `global` \| `workspace` |
| `workspace_id` | `ulid` | conditional | Required when `scope = workspace` |
| `description` | `text` | yes | What the skill teaches/does; used for selection |
| `content_ref` | `path` | yes | Root of the skill's content (documents, templates), relative to its install location |
| `content_hash` | `hash` | yes | SHA-256 over the skill content manifest (integrity; manifest rules in Volume 6) |
| `required_tools` | `json` | no | Tool name selectors the skill depends on |
| `required_capabilities` | `json` | no | Capability values the skill depends on |
| `package_id` | `ulid` | no | Delivering Package, when installed from one |
| `extension_id` | `ulid` | conditional | Extension registry record; required for non-`builtin` scope |
| `trust_level` | `enum` | yes | Trust classification (Volume 9 vocabulary) |
| `enabled` | `boolean` | yes | Loadable by agents |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(scope, workspace_id, name, version)` unique.

### Relations

- Optionally delivered by one **Package** and classified by one **Extension**.
- Declares requirements on **Tool** names and **Capability** values (selectors, not FKs).

### Integrity invariants

1. **INV-SKL-01** — `(scope, workspace_id, name, version)` MUST be unique.
2. **INV-SKL-02** — Skill content is immutable per version: `content_hash` MUST verify at
   load time; a mismatch disables the skill with an integrity error (Volume 6).
3. **INV-SKL-03** — A Skill MUST declare its tool and capability requirements; loading a
   skill whose requirements are unsatisfiable is reported, not silently degraded
   (Principle 2).
4. **INV-SKL-04** — Skills carry no executable code; anything executable enters as a Plugin
   or Tool instead (boundary rule; enforcement in Volume 6).

### Lifecycle

Stateless versioned catalog (`enabled` is a recorded flag).

### Persistence

Per scope: global or workspace database, table `skills`; content on disk under the install
location (Volume 6 layout). Retention: versions persist while referenced by runs' records.

### Versioning and serialization

Semantic `version`; `revision` for row metadata. Serializes as canonical JSON metadata;
content exports as files.

## Plugin

Purpose: an external extension process integrated through the plugin runtime and the
Andromeda Runtime Protocol (JSON-RPC 2.0 subprocess, ADR-009).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `name` | `string` | yes | Plugin name |
| `version` | `semver` | yes | Plugin version |
| `scope` | `enum` | yes | `global` \| `workspace` |
| `workspace_id` | `ulid` | conditional | Required when `scope = workspace` |
| `executable_ref` | `path` | yes | Entry-point executable, relative to the install location |
| `manifest` | `json` | yes | Declared manifest: protocol version required, surfaces provided (tools, commands, …), permission requirements (manifest schema owned by Volume 6) |
| `protocol_version` | `string` | no | Andromeda Runtime Protocol version negotiated at last handshake |
| `state` | `enum` | yes | Canonical Plugin state (chapter 09) |
| `package_id` | `ulid` | no | Delivering Package |
| `extension_id` | `ulid` | yes | Extension registry record |
| `trust_level` | `enum` | yes | Trust classification (Volume 9) |
| `last_started_at` | `timestamp` | no | Last process start |
| `last_error` | `json` | no | Last failure summary |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(scope, workspace_id, name)` unique — one registered
  version per name per scope; upgrading replaces the row (the old version remains available
  through Package history).
- OS process IDs are external identifiers, recorded in Events only.

### Relations

- Registers 0..n **Tool** (origin `plugin`); may provide other surfaces per its manifest
  (commands, exporters — registries owned by their area volumes).
- Delivered by at most one **Package**; classified by exactly one **Extension**.

### Integrity invariants

1. **INV-PLG-01** — A Plugin MUST declare its manifest before first start; surfaces not
   declared in the manifest MUST NOT be registered at runtime (Volume 6 enforces; the model
   forbids rows that bypass it — every plugin-origin Tool's `origin_ref` points to a Plugin
   whose manifest declares tools).
2. **INV-PLG-02** — `protocol_version` records the *negotiated* version; a plugin whose
   required protocol version cannot be negotiated stays in `failed` with `last_error` set
   (negotiation rules in Volume 6 per ADR-009).
3. **INV-PLG-03** — Plugin process state is runtime truth, the persisted `state` is last
   known: on startup after a crash, plugins found in running-family states are reconciled to
   `stopped` before restart policy applies (recovery owned by Volume 6).
4. **INV-PLG-04** — Removing a Plugin disables its Tools (INV-TOOL-04 keeps the rows) and
   reaches `removed` only after its process is confirmed terminated.

### Lifecycle

Stateful — canonical states `registered`, `starting`, `running`, `stopping`, `stopped`,
`failed`, `disabled`, `removed` (chapter 09); full machine owned by Volume 6.

### Persistence

Per scope: global or workspace database, table `plugins`. Retention: rows persist while
referenced by Tool provenance; `removed` rows are tombstones.

### Versioning and serialization

Semantic `version` (plugin), plus negotiated `protocol_version` (protocol). `revision` for row
concurrency. Manifest serializes verbatim.

## MCP Server

Purpose: an external Model Context Protocol server offering tools, resources, or prompts,
registered in configuration and reached through managed client connections (MCP support per
ADR-010).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `name` | `string` | yes | Registration name (configuration key under `[mcp]`, Volume 10) |
| `scope` | `enum` | yes | `global` \| `workspace` |
| `workspace_id` | `ulid` | conditional | Required when `scope = workspace` |
| `transport` | `enum` | yes | `stdio` \| `streamable_http` (per the MCP specification; vocabulary tracked by Volume 6) |
| `launch` | `json` | conditional | For `stdio`: command + args to launch the server (values redacted per Volume 9; secrets by Credential reference only) |
| `endpoint` | `json` | conditional | For `streamable_http`: URL and non-secret headers |
| `credential_id` | `ulid` | no | Credential used to authenticate to the server, when required |
| `trust_level` | `enum` | yes | Trust classification (Volume 9); MCP servers are third-party by definition |
| `discovered_surfaces` | `json` | no | Last discovery summary: counts and names of tools/resources/prompts offered |
| `extension_id` | `ulid` | yes | Extension registry record |
| `enabled` | `boolean` | yes | Connections may be established |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(scope, workspace_id, name)` unique.

### Relations

- Connected via 0..n **MCP Client Connection** (aggregate members).
- Offers 0..n **Tool** (origin `mcp`); resources and prompts are runtime surfaces specified in
  Volume 6 (not separate catalog entities in this model — they surface as Context Items and
  prompt inputs).
- Optionally authenticates via one **Credential**; classified by exactly one **Extension**.

### Integrity invariants

1. **INV-MCPS-01** — `(scope, workspace_id, name)` MUST be unique.
2. **INV-MCPS-02** — `launch`/`endpoint` MUST be consistent with `transport` and MUST NOT
   contain secret material (secrets via `credential_id` only).
3. **INV-MCPS-03** — Tools offered by an MCP Server enter the registry as ordinary Tool rows
   with `origin = mcp` and `origin_ref = id` — same contract, same permissions, same
   observability as every other tool (Principle 4).
4. **INV-MCPS-04** — Disabling or removing an MCP Server MUST terminate its connections and
   disable its Tools.

### Lifecycle

Stateless registry entry (`enabled` is a recorded flag); the runtime machine belongs to MCP
Client Connection.

### Persistence

Per scope: global or workspace database, table `mcp_servers`. Retention: while referenced by
tool provenance; tombstoned on removal.

### Versioning and serialization

Row versioning via `revision`; the server's own version string, when reported during MCP
initialization, is recorded in `discovered_surfaces` as external information.

## MCP Client Connection

Purpose: Andromeda's managed connection to one MCP server — handshake, capability negotiation,
health, and teardown, as a recorded runtime instance.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `mcp_server_id` | `ulid` | yes | Owning MCP Server registration |
| `state` | `enum` | yes | Canonical MCP Client Connection state (chapter 09) |
| `protocol_version` | `string` | no | MCP protocol version negotiated |
| `negotiated_capabilities` | `json` | no | Capability set agreed during initialization (MCP capability semantics, Volume 6) |
| `server_session_id` | `string` | no | Server-issued session identifier, when the transport provides one (external identifier) |
| `connected_at` | `timestamp` | no | When the connection reached `ready` |
| `disconnected_at` | `timestamp` | no | When it left service |
| `last_error` | `json` | no | Last failure summary |
| `stats` | `json` | no | Counters (requests, notifications, reconnects) for diagnostics |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. `server_session_id` is external and never identity.

### Relations

- Belongs to exactly one **MCP Server** (aggregate member).

### Integrity invariants

1. **INV-MCPC-01** — A connection MUST belong to exactly one MCP Server and cannot be
   re-pointed at another.
2. **INV-MCPC-02** — At most one MCP Client Connection per MCP Server may be in a
   non-terminal, non-failed state per Andromeda process (single managed connection;
   reconnection creates a new row linked by the server, preserving connection history).
3. **INV-MCPC-03** — Tool invocations routed over a connection MUST occur only in `ready`
   state; requests against any other state fail with a stable error (codes minted by
   Volume 6).
4. **INV-MCPC-04** — Persisted connection state is last-known runtime truth; recovery
   reconciles running-family states to `disconnected` on startup (mirror of INV-PLG-03).

### Lifecycle

Stateful — canonical states `configured`, `connecting`, `initializing`, `ready`,
`reconnecting`, `disconnected`, `failed`, `disabled`, `removed` (chapter 09); full machine
owned by Volume 6.

### Persistence

Workspace database, table `mcp_connections` (connections are established in workspace
context, including for global-scoped servers). Retention: connection history pruned per
Volume 10 policy.

### Versioning and serialization

Row versioning via `revision`. Serializes for diagnostics exports; `server_session_id`
included (it is not a secret per MCP; if a server treats it as one, Volume 9 redaction
applies).

## Package

Purpose: a distributable unit — plugin, skill, or bundle — with version, checksum, and
signature metadata, moving through a verified installation lifecycle.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `name` | `string` | yes | Package name (official packages use the `andromeda-` prefix) |
| `version` | `semver` | yes | Package version |
| `kind` | `enum` | yes | `plugin` \| `skill` \| `bundle` (closed; extended via change procedure) |
| `scope` | `enum` | yes | Install scope: `global` \| `workspace` |
| `workspace_id` | `ulid` | conditional | Required when `scope = workspace` |
| `source` | `json` | yes | Where it came from: source kind (`registry` \| `archive` \| `path` \| `git`) plus locator (URL/path/ref); acquisition rules owned by Volume 6/14 |
| `checksum` | `hash` | yes | SHA-256 of the package archive |
| `signature_state` | `enum` | yes | `verified` \| `unverified` \| `invalid`; signing policy owned by Volume 14 (signing viability per Volume 1) |
| `state` | `enum` | yes | Canonical Package installation state (chapter 09) |
| `installed_at` | `timestamp` | no | When installation completed |
| `files_manifest` | `json` | no | Installed files with per-file hashes (for verification and clean removal) |
| `last_error` | `json` | no | Last installation failure summary |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(scope, workspace_id, name, version)` unique.

### Relations

- Delivers 1..n **Extension** (and through them Plugins, Skills, MCP Server registrations,
  adapters).

### Integrity invariants

1. **INV-PKG-01** — `checksum` MUST be verified before any package content executes or
   registers; a checksum mismatch terminates installation as `failed` with nothing installed
   (verification order owned by Volume 6/14).
2. **INV-PKG-02** — `signature_state = invalid` MUST block installation regardless of user
   configuration; `unverified` proceeds only per the explicit trust policy of Volume 9/14.
3. **INV-PKG-03** — `files_manifest` MUST cover every file the installation wrote, so removal
   and integrity verification are complete.
4. **INV-PKG-04** — Two versions of the same package MAY be installed only if Volume 6's
   coexistence rules allow their kinds; the registry never holds two rows in `installed`
   state for the same `(scope, workspace_id, name)` unless those rules say so.

### Lifecycle

Stateful — canonical installation states `requested`, `resolving`, `downloading`,
`verifying`, `staged`, `installing`, `installed`, `failed`, `rolled_back`, `removed`
(chapter 09); full machine owned by Volume 6 (extension packages; product updates are
Volume 14's Update process).

### Persistence

Per scope: global or workspace database, table `packages`. Retention: rows persist for
provenance; `removed` rows are tombstones.

### Versioning and serialization

Semantic `version`; `revision` for row concurrency. Serializes as canonical JSON.

## Extension

Purpose: the umbrella registry record for any third-party addition — tool, plugin, skill,
provider adapter, exporter, or command — giving every extension surface one uniform row for
identity, origin, trust, and enablement (PRD-007).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `kind` | `enum` | yes | `plugin` \| `skill` \| `mcp_server` \| `provider_adapter` \| `exporter` \| `command` \| `workflow` \| `policy` \| `indexer` (closed vocabulary owned by Volume 6, aligned with the Principle 6 extension surfaces) |
| `name` | `string` | yes | Extension name as registered |
| `version` | `semver` | yes | Version of the concrete unit |
| `scope` | `enum` | yes | `global` \| `workspace` |
| `workspace_id` | `ulid` | conditional | Required when `scope = workspace` |
| `package_id` | `ulid` | no | Delivering Package |
| `subject_id` | `ulid` | conditional | The concrete row (Plugin, Skill, MCP Server, Workflow); absent for kinds whose registry lives outside this volume (e.g., `provider_adapter` registration, Volume 5) |
| `contract_version` | `string` | yes | Version of the public contract the extension targets (Extension SDK / adapter / protocol version; SM-20 stability tracking) |
| `trust_level` | `enum` | yes | Trust classification (Volume 9) |
| `enabled` | `boolean` | yes | Extension participates in its surface |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(scope, workspace_id, kind, name, version)` unique.

### Relations

- Delivered by at most one **Package**; classifies at most one concrete row (**Plugin**,
  **Skill**, **MCP Server**, **Workflow**, or a registration owned by another volume).

### Integrity invariants

1. **INV-EXT-01** — Every installed non-builtin Plugin, Skill, and MCP Server MUST have
   exactly one Extension record, and `kind`/`subject_id` MUST agree with the concrete row.
2. **INV-EXT-02** — `kind` values come from the Volume 6 closed vocabulary; adding a kind is
   an ADR-gated change (Principle 6 surfaces).
3. **INV-EXT-03** — Disabling an Extension MUST disable its concrete unit's participation
   (cascade semantics per kind owned by Volume 6).
4. **INV-EXT-04** — `contract_version` MUST be recorded at install/registration time; the
   compatibility decision it feeds is owned by Volume 6/14.

### Lifecycle

Stateless registry entry (`enabled` is a recorded flag; installation state lives on Package,
process state on Plugin/MCP Client Connection).

### Persistence

Per scope: global or workspace database, table `extensions`. Retention: for provenance, like
Package.

### Versioning and serialization

`version` mirrors the concrete unit; `contract_version` tracks the public contract targeted.
Serializes as canonical JSON.

## Release

Purpose: a published, versioned distribution of Andromeda itself, with artifacts, notes, and
provenance — the record the Updater reasons over (Volume 14).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `version` | `semver` | yes | Product version; unique |
| `channel` | `enum` | yes | Release channel; vocabulary owned by Volume 14 (e.g., `stable`, `beta`) |
| `state` | `enum` | yes | Canonical Release state (chapter 09) |
| `published_at` | `timestamp` | no | Publication instant (from release provenance) |
| `artifacts` | `json` | yes | Per-platform artifact descriptors: platform, file name, size, SHA-256 checksum, signature status |
| `notes_ref` | `string` | no | Locator of the release notes |
| `provenance` | `json` | no | Build provenance metadata (workflow, commit, attestations; scheme owned by Volume 14) |
| `min_upgrade_from` | `semver` | no | Oldest version that may update directly to this release (upgrade-path rule, Volume 14) |
| `yanked_reason` | `text` | conditional | Required when `state = yanked` |
| `created_at` | `timestamp` | yes | Creation instant (when this installation learned of the release) |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `version` unique.

### Relations

- Consulted by the Update process (chapter 09, Volume 14); the currently installed version is
  recorded in the global database's update history (Volume 14 tables), referencing Release
  rows.

### Integrity invariants

1. **INV-REL-01** — `version` MUST be unique; a published version's `artifacts` set is
   immutable (a fixed release is a *new* version).
2. **INV-REL-02** — Every artifact descriptor MUST carry a SHA-256 checksum; the Updater
   MUST NOT apply an artifact whose digest does not verify (Volume 14; exit and error
   semantics there).
3. **INV-REL-03** — `yanked` releases MUST NOT be offered by update checks; an installation
   already running a yanked version is notified per Volume 14.
4. **INV-REL-04** — Release rows record *published facts*; this installation's opinion
   (downloaded, applied, rolled back) lives in the Update process records (Volume 14), never
   on the Release row.

### Lifecycle

Stateful — canonical states `drafted`, `candidate`, `published`, `superseded`, `yanked`
(chapter 09); full machine owned by Volume 14. (Locally, most rows are first seen already
`published`; the early states exist for the release pipeline's own records.)

### Persistence

Global database, table `releases`. Retention: kept indefinitely (small rows, provenance
value); Volume 14 may prune pre-release channels.

### Versioning and serialization

`version` is the product semver; `revision` is row concurrency. Serializes as canonical JSON.
