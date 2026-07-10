# 04 — Tools and Actions

This chapter defines the action boundary of the model: **Tool** (registry aggregate), the Run
aggregate members that record acting — **Tool Invocation**, **Tool Result**, **File Change**,
**Patch**, **Command Execution**, **Artifact** — and the standalone decision aggregates
**Approval** and **Permission**. The tool contract is owned by Volume 6; permission and
approval semantics by Volume 9. Everything an agent does to the world passes through these
entities (PRD-004): if an action left no rows here, it did not happen through Andromeda.

## Tool

Purpose: a named, versioned, schema-typed capability an agent can invoke, with declared
permissions and limits. Built-in, plugin-provided, and MCP-provided tools are equal citizens
of one contract (Principle 4).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key (one row per registered tool version per origin) |
| `name` | `string` | yes | Canonical tool name; naming grammar owned by Volume 6 |
| `version` | `semver` | yes | Tool contract version as declared by its origin |
| `origin` | `enum` | yes | `builtin` \| `plugin` \| `mcp` (closed; extended only via the change procedure) |
| `origin_ref` | `ulid` | conditional | Providing Plugin or MCP Server; absent for `builtin` |
| `description` | `text` | yes | Model- and human-facing description |
| `input_schema` | `json` | yes | JSON Schema of the tool's input (contract rules in Volume 6) |
| `output_schema` | `json` | yes | JSON Schema of the tool's output |
| `permission_declaration` | `json` | yes | Permissions the tool may require, using the Volume 9 permission vocabulary |
| `limits` | `json` | yes | Declared timeouts and resource limits (defaulting rules in Volume 6) |
| `trust_level` | `enum` | yes | Trust classification; vocabulary owned by Volume 9 (built-ins are the highest level; third-party levels per origin and signature state) |
| `enabled` | `boolean` | yes | Registration is active |
| `registered_at` | `timestamp` | yes | First registration of this version |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(name, version, origin, origin_ref)` unique.
- Agents address tools by `name` (highest enabled version wins; resolution owned by
  Volume 6); records always pin the exact `id`.

### Relations

- Provided by at most one **Plugin** or **MCP Server** (via `origin_ref`).
- Invoked by 0..n **Tool Invocation**.
- Selected by **Agent Profile** `tool_policy` (by name selectors, not FK).

### Integrity invariants

1. **INV-TOOL-01** — `(name, version, origin, origin_ref)` MUST be unique; two origins MUST
   NOT silently shadow each other's names (collision handling owned by Volume 6, but the
   registry never stores an ambiguous row).
2. **INV-TOOL-02** — A Tool MUST declare `input_schema`, `output_schema`,
   `permission_declaration`, and `limits` before it is registered; a tool without a complete
   declaration is rejected, not defaulted into existence (Principle 4).
3. **INV-TOOL-03** — `origin` and `trust_level` MUST be recorded at registration and MUST be
   visible wherever the tool is offered to a model or user (Principle 4: origin and trust
   always visible).
4. **INV-TOOL-04** — A Tool row that any persisted Tool Invocation references MUST NOT be
   deleted; unregistration disables it (`enabled = false`) and it remains for attribution.

### Lifecycle

Stateless registry entry (`enabled` is a recorded flag). The *providing* entities have
machines (Plugin, MCP Client Connection, chapter 09); tool availability follows them
(Volume 6).

### Persistence

Registered tools persist in the database matching their provider's install scope
(workspace database table `tools` for workspace-scoped plugins/servers; global database table
`tools` for global-scoped ones). `builtin` tools are materialized from the binary at startup
and are not persisted rows; invocation records still pin their `name`/`version` snapshot.
Retention: rows persist while referenced (INV-TOOL-04).

### Versioning and serialization

`version` is the tool contract version (semver, compatibility rules in Volume 6); `revision`
is row concurrency. Schemas serialize verbatim as JSON.

## Tool Invocation

Purpose: one call of a tool with concrete inputs, under a granted permission set — the atomic,
auditable unit of agent action.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Owning Run |
| `turn_id` | `ulid` | yes | Turn whose model response requested the call |
| `task_id` | `ulid` | no | Task being fulfilled, when the run is executing a plan |
| `agent_id` | `ulid` | yes | Agent that issued the call |
| `tool_id` | `ulid` | no | Registered Tool row (absent for `builtin` origin) |
| `tool_name` | `string` | yes | Snapshot of the tool name at call time |
| `tool_version` | `semver` | yes | Snapshot of the tool version at call time |
| `tool_origin` | `enum` | yes | Snapshot of origin (`builtin` \| `plugin` \| `mcp`) |
| `arguments` | `json` | yes | Concrete input, validated against `input_schema`; stored redacted per Volume 9 rules |
| `state` | `enum` | yes | Canonical Tool Invocation state (chapter 09) |
| `approval_id` | `ulid` | no | Approval that decided the call, when interactive consent was required |
| `permission_ids` | `json` | yes | Permission grants under which the call executed (list of Permission IDs; possibly empty for permissionless read-only tools per Volume 9) |
| `sequence_no` | `integer` | yes | Position among the run's invocations (1-based, dense) |
| `timeout_ms` | `duration_ms` | yes | Effective timeout applied |
| `retry_of_id` | `ulid` | no | Prior Tool Invocation this one retries |
| `started_at` | `timestamp` | no | Transition to `executing` |
| `ended_at` | `timestamp` | no | Terminal transition instant |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(run_id, sequence_no)` unique.
- `id` is the correlation key joining Tool Result, Command Execution, File Changes, Events,
  and trace spans for this action.

### Relations

- Belongs to exactly one **Run**, one **Turn**, one **Agent**; optionally one **Task**.
- Calls one **Tool** (by snapshot, plus `tool_id` when a registry row exists).
- Yields at most one **Tool Result**; produces at most one **Command Execution** and 0..n
  **File Change** / **Artifact** rows.
- Decided by at most one **Approval**; executed under 0..n **Permission** grants.
- May retry exactly one earlier **Tool Invocation** (`retry_of_id` chain, acyclic).

### Integrity invariants

1. **INV-TINV-01** — A Tool Invocation MUST belong to exactly one Run, and its `turn_id`,
   `task_id`, and `agent_id` MUST reference members of that same Run.
2. **INV-TINV-02** — The name/version/origin snapshot MUST be written at creation and remain
   valid even if the Tool row is later disabled or its provider removed: the record stands on
   its own for audit (PRD-006).
3. **INV-TINV-03** — A Tool Invocation MUST NOT enter `executing` before its permission
   context is recorded: either `approval_id` is set, or every required permission resolves to
   grants listed in `permission_ids`, or the tool requires none (Volume 9 decides which;
   Principle 8 forbids a fourth path).
4. **INV-TINV-04** — `arguments` MUST validate against the tool's `input_schema` before
   execution; invalid arguments terminate the invocation as `failed` without side effects.
5. **INV-TINV-05** — Denial is first-class: a denied invocation reaches `denied` with the
   deciding Approval/Permission recorded, and the denial is delivered to the agent as
   structured input (PRD-005).
6. **INV-TINV-06** — The `retry_of_id` chain MUST be acyclic and MUST stay within one Run.

### Lifecycle

Stateful — canonical states `requested`, `awaiting_approval`, `approved`, `executing`,
`succeeded`, `failed`, `denied`, `timed_out`, `cancelled` (chapter 09); full machine owned by
Volume 6.

### Persistence

Workspace database, table `tool_invocations`. Retention: with the owning run; audit
obligations take precedence over pruning (Volume 9).

### Versioning and serialization

Row versioning via `revision`. Serializes inside the run record stream; `arguments` are
serialized post-redaction — the persisted form is the exportable form.

## Tool Result

Purpose: the typed output (or error) of a tool invocation.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `tool_invocation_id` | `ulid` | yes | The invocation this result belongs to (unique) |
| `run_id` | `ulid` | yes | Owning Run (denormalized) |
| `status` | `enum` | yes | `success` \| `error` (recorded vocabulary, chapter 09) |
| `payload` | `json` | conditional | Output conforming to the tool's `output_schema`; required when `status = success` |
| `error` | `json` | conditional | Structured error (stable code, category, safe message, recoverability) per the Volume 0 error scheme; required when `status = error` |
| `truncated` | `boolean` | yes | Payload was truncated to size limits, with the full content spilled to an Artifact |
| `spillover_artifact_id` | `ulid` | no | Artifact holding the full output when `truncated` |
| `payload_size_bytes` | `integer` | yes | Size of the untruncated output |
| `duration_ms` | `duration_ms` | yes | Execution duration |
| `created_at` | `timestamp` | yes | Creation instant |

### Identifiers

- Primary key: `id`. Natural key: `tool_invocation_id` unique (1:1).

### Relations

- Belongs to exactly one **Tool Invocation** (1:1); optionally spills to one **Artifact**.

### Integrity invariants

1. **INV-TRES-01** — Exactly one Tool Result MUST exist for every Tool Invocation that reached
   `succeeded`, `failed`, or `timed_out`; none for `denied`, `cancelled` before execution, or
   non-terminal states.
2. **INV-TRES-02** — Tool Results are immutable records.
3. **INV-TRES-03** — `status` MUST be consistent with the invocation's terminal state
   (`succeeded` ⇔ `success`), and exactly one of `payload`/`error` MUST be present.
4. **INV-TRES-04** — A `success` payload MUST validate against the tool's `output_schema`;
   nonconforming output is recorded as an `error` result with the raw output spilled to an
   Artifact (never silently coerced).

### Lifecycle

Immutable record.

### Persistence

Workspace database, table `tool_results`. Retention: with the owning run.

### Versioning and serialization

No `revision` (immutable). Serializes inside the run record stream.

## Approval

Purpose: a recorded human decision granting or denying a requested action or permission —
including the policy-resolved case, so that *every* consent decision has exactly one record
(PRD-005).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `subject_kind` | `enum` | yes | `tool_invocation` \| `permission_request` \| `plan` \| `workflow_gate` (closed; extended via change procedure) |
| `subject_id` | `ulid` | yes | The entity awaiting decision |
| `run_id` | `ulid` | no | Run in whose context the request arose (absent for standing permission requests made outside runs) |
| `requested_scope` | `json` | yes | What was asked: permission names, resource selectors, action summary (vocabulary per Volume 9) |
| `state` | `enum` | yes | Canonical Approval state (chapter 09) |
| `decided_by_kind` | `enum` | conditional | `user` \| `policy`; required once decided |
| `policy_ref` | `string` | conditional | Identifier of the deciding policy rule (Volume 9), required when `decided_by_kind = policy` |
| `decision_note` | `text` | no | Optional human rationale |
| `granted_permission_ids` | `json` | no | Permission grants minted by this approval |
| `expires_at` | `timestamp` | no | Deadline after which an undecided request expires (timeout policy per Volume 9) |
| `requested_at` | `timestamp` | yes | Request instant |
| `decided_at` | `timestamp` | no | Decision instant |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change (state transitions only) |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. No natural key.

### Relations

- Decides exactly one subject (**Tool Invocation**, **Plan**, **Workflow Run** gate, or a
  standing permission request).
- May mint 0..n **Permission** grants; produces exactly one **Audit Record** on decision
  (chapter 08; Volume 9 owns the audit trigger list).

### Integrity invariants

1. **INV-APR-01** — An Approval decides exactly one subject; the subject MUST exist when the
   approval is created.
2. **INV-APR-02** — Once terminal, an Approval is immutable: decisions are never edited,
   re-decided, or deleted. A changed mind is a *new* request and a *new* Approval.
3. **INV-APR-03** — Every decision MUST record its decider: `decided_by_kind = user` for
   interactive decisions, `policy` plus `policy_ref` for policy-resolved ones. Unattended
   modes never skip the record (PRD-009; Volume 1 "Safety vs. automation").
4. **INV-APR-04** — `granted_permission_ids` MUST reference Permission rows whose `origin`
   points back to this Approval — the grant graph is bidirectionally consistent.
5. **INV-APR-05** — An expired or cancelled request MUST resolve the subject as denied-class
   (the subject does not proceed); expiry is a terminal state, not a pause.

### Lifecycle

Stateful — canonical states `requested`, `granted`, `denied`, `expired`, `cancelled`
(chapter 09); full machine owned by Volume 9.

### Persistence

Workspace database, table `approvals` (approvals arise in workspace context; standing global
permission requests are recorded in the global database's `approvals` table). Retention:
approvals are audit-relevant and follow Volume 9 audit retention, which takes precedence over
run pruning.

### Versioning and serialization

Row versioning via `revision` (state transitions only). Serializes in the run record stream
when run-bound, and in audit exports.

## Permission

Purpose: a grant to perform a class of side-effecting action within a scope. The permission
model — names, scopes, evaluation, precedence — is owned by Volume 9; this entity is the
persisted grant record.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `permission_name` | `enum` | yes | Value from the closed permission enum minted by Volume 9 |
| `effect` | `enum` | yes | `allow` \| `deny` |
| `scope` | `enum` | yes | `invocation` \| `run` \| `session` \| `workspace` \| `global` |
| `scope_ref` | `ulid` | conditional | The scoping entity (Run, Session, Workspace); absent for `global` and `invocation` (invocation grants bind via `permission_ids` on the invocation) |
| `resource_selector` | `json` | no | Constrains the grant to resources (path globs, command patterns, hosts); selector grammar owned by Volume 9 |
| `origin_kind` | `enum` | yes | `approval` \| `policy` \| `default`; how the grant came to exist |
| `origin_ref` | `ulid` | conditional | The minting Approval (`origin_kind = approval`) |
| `valid_from` | `timestamp` | yes | Start of validity |
| `valid_until` | `timestamp` | no | Expiry; absent = until revoked |
| `revoked_at` | `timestamp` | no | Revocation instant |
| `revoked_by` | `enum` | no | `user` \| `policy` \| `system`; required when revoked |
| `created_at` | `timestamp` | yes | Creation instant |

### Identifiers

- Primary key: `id`. No natural key (overlapping grants are legal; Volume 9 defines
  evaluation precedence).

### Relations

- Minted by at most one **Approval**; referenced by **Tool Invocation** `permission_ids`.
- Scoped to at most one **Run**, **Session**, or **Workspace** via `scope_ref`.

### Integrity invariants

1. **INV-PERM-01** — `permission_name` MUST be a value of the Volume 9 closed enum; unknown
   names are invalid rows, not forward-compatible ones.
2. **INV-PERM-02** — Grants are append-only: revocation sets `revoked_at`/`revoked_by` and
   never deletes or edits the grant; history of what was permitted when is preserved
   (PRD-006).
3. **INV-PERM-03** — A revoked or expired grant MUST NOT satisfy any new permission check;
   in-flight actions on revocation are handled per Volume 9.
4. **INV-PERM-04** — `scope` and `scope_ref` MUST be consistent (INV-CFGP-03 pattern);
   `origin_kind = approval` requires `origin_ref`.

### Lifecycle

Record with recorded validity (`valid_from`/`valid_until`/`revoked_at`); not a canonical
state machine — evaluation semantics live in Volume 9.

### Persistence

Workspace database, table `permissions` for workspace/session/run/invocation scopes; global
database, same table, for `global` scope. Retention: per Volume 9 audit retention.

### Versioning and serialization

Immutable except revocation fields; no `revision`. Serializes in audit exports and the run
record stream (grants referenced by invocations).

## Artifact

Purpose: a durable output produced by a run — file exports, reports, generated patches, spilled
tool output — content-addressed and attributable.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Producing Run |
| `task_id` | `ulid` | no | Producing Task, when plan-driven |
| `tool_invocation_id` | `ulid` | no | Producing invocation, when tool-produced |
| `kind` | `enum` | yes | `file` \| `patch` \| `report` \| `export` \| `tool_output` \| `log_bundle` (closed; extended via change procedure) |
| `name` | `string` | yes | Display name (e.g., file name) |
| `media_type` | `string` | yes | MIME type |
| `content_ref` | `path` | conditional | Location of the content under `.andromeda/artifacts/`, relative to the workspace root; absent when pruned (tombstone) |
| `content_hash` | `hash` | yes | SHA-256 of the content at creation |
| `size_bytes` | `integer` | yes | Content size |
| `retention_class` | `enum` | yes | Retention category; vocabulary and policies owned by Volume 10 |
| `created_at` | `timestamp` | yes | Creation instant |

### Identifiers

- Primary key: `id`. `content_hash` deduplicates storage (two artifacts may share content).

### Relations

- Produced by exactly one **Run**; optionally by one **Task** / **Tool Invocation**.
- Referenced by **Tool Result** spillover, **Patch** bodies, and Message `file_ref`/
  `image_ref` parts.

### Integrity invariants

1. **INV-ART-01** — An Artifact MUST be attributable: `run_id` always; `tool_invocation_id`
   whenever a tool produced it (INV-RUN-05).
2. **INV-ART-02** — `content_hash` MUST match the stored content; a mismatch is an integrity
   error (exit code 9 class; verification procedure owned by Volume 10).
3. **INV-ART-03** — Pruning content keeps the metadata row as a tombstone (`content_ref`
   cleared, hash and size retained) so the record of *what was produced* survives the content.
4. **INV-ART-04** — Artifact metadata is immutable; content files are write-once (no in-place
   mutation after commit).

### Lifecycle

Immutable record (content prunable per retention).

### Persistence

Workspace database, table `artifacts`; content files under `.andromeda/artifacts/` (layout
owned by Volume 10). Retention: `retention_class` policies in Volume 10.

### Versioning and serialization

No `revision`. Exports include metadata always and content by reference (hash + optional
inline base64 for small artifacts; threshold owned by Volume 10).

## File Change

Purpose: a recorded modification to a file (create, edit, delete, rename) attributable to a
run — the ground truth for "what did Andromeda change?" (PRD-006).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Attributed Run |
| `tool_invocation_id` | `ulid` | yes | Invocation that performed the change |
| `patch_id` | `ulid` | no | Patch this change is grouped into, if any |
| `op` | `enum` | yes | `create` \| `edit` \| `delete` \| `rename` |
| `path` | `path` | yes | Affected file, relative to the workspace root |
| `old_path` | `path` | conditional | Prior path; required when `op = rename` |
| `before_hash` | `hash` | conditional | SHA-256 of file content before; required for `edit`, `delete`, `rename` |
| `after_hash` | `hash` | conditional | SHA-256 of file content after; required for `create`, `edit`, `rename` |
| `diff_artifact_id` | `ulid` | no | Artifact holding this change's diff hunks |
| `applied_at` | `timestamp` | yes | When the change hit the filesystem |
| `reverted_by_id` | `ulid` | no | Later File Change that reverted this one |
| `created_at` | `timestamp` | yes | Creation instant |

### Identifiers

- Primary key: `id`.

### Relations

- Attributed to exactly one **Run** and one **Tool Invocation**.
- Grouped into at most one **Patch**; optionally reverted by one later **File Change**.

### Integrity invariants

1. **INV-FCH-01** — Every File Change MUST reference the Tool Invocation that performed it;
   filesystem effects without an invocation record are a defect of the tool boundary
   (PRD-004).
2. **INV-FCH-02** — Hash fields MUST be present per `op` as specified above; they are what
   makes revert and conflict detection decidable (revert semantics owned by Volumes 6/11).
3. **INV-FCH-03** — `path` (and `old_path`) MUST lie within scopes the invocation's permission
   context allowed; the record stores the paths actually touched (enforcement is Volume 9's,
   the record is this volume's).
4. **INV-FCH-04** — File Changes are immutable records; a revert is a new File Change
   referencing the original via `reverted_by_id` back-link.

### Lifecycle

Immutable record.

### Persistence

Workspace database, table `file_changes`. Retention: with the owning run; audit precedence
applies.

### Versioning and serialization

No `revision`. Serializes in the run record stream.

## Patch

Purpose: a reviewable diff representing one or more file changes — the unit users review,
apply, or reject (inline diff review, PRD-008).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Producing Run |
| `title` | `string` | yes | Review title |
| `format` | `enum` | yes | `unified_diff` (closed vocabulary; extended via change procedure) |
| `body_artifact_id` | `ulid` | yes | Artifact holding the full diff body |
| `base_vcs_ref` | `string` | no | VCS commit hash the patch was computed against (external identifier, Volume 11) |
| `status` | `enum` | yes | Recorded status: `proposed` \| `applied` \| `rejected` \| `reverted` (chapter 09 recorded vocabulary) |
| `decided_at` | `timestamp` | no | When it was applied or rejected |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change (status only) |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`.

### Relations

- Produced by exactly one **Run**; groups 1..n **File Change** (when applied) and one body
  **Artifact**.

### Integrity invariants

1. **INV-PATCH-01** — A Patch body is immutable; revising a proposed patch creates a new
   Patch.
2. **INV-PATCH-02** — `status = applied` requires at least one File Change referencing this
   Patch; a `proposed` or `rejected` Patch has none (proposal precedes effect).
3. **INV-PATCH-03** — Applying a Patch happens through a Tool Invocation like any other write
   (PRD-004); the resulting File Changes carry both references.

### Lifecycle

Recorded status only (`proposed`, `applied`, `rejected`, `reverted`); application semantics
owned by Volume 6 (filesystem tools) and Volume 11 (VCS-aware application).

### Persistence

Workspace database, table `patches`; body content as Artifact. Retention: with the owning run.

### Versioning and serialization

Row versioning via `revision` (status only). Exports embed the body by artifact reference.

## Command Execution

Purpose: a recorded execution of a terminal command, with its inputs, environment policy, and
outcome — the terminal counterpart of File Change.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Attributed Run |
| `tool_invocation_id` | `ulid` | yes | The invocation that executed the command (1:1) |
| `argv` | `json` | yes | Command and arguments as executed, redacted per Volume 9 |
| `cwd` | `path` | yes | Working directory at execution |
| `env_policy_ref` | `string` | yes | Identifier of the environment passthrough policy applied (Volume 9); variable *names* allowed through MAY be recorded, values never |
| `pty` | `boolean` | yes | Executed under a PTY (Terminal Engine) vs. pipe capture |
| `sandbox_profile_ref` | `string` | no | Sandbox profile applied (Volume 9) |
| `exit_code` | `integer` | no | Process exit code; absent if killed before exit |
| `signal` | `string` | no | Terminating signal name, when signaled |
| `outcome` | `enum` | yes | Recorded outcome: `succeeded` \| `failed` \| `timed_out` \| `killed` (chapter 09 recorded vocabulary) |
| `stdout_artifact_id` | `ulid` | no | Captured stdout (truncation rules in Volume 6) |
| `stderr_artifact_id` | `ulid` | no | Captured stderr |
| `truncated` | `boolean` | yes | Captures were truncated |
| `started_at` | `timestamp` | yes | Process start |
| `ended_at` | `timestamp` | no | Process end |
| `created_at` | `timestamp` | yes | Creation instant |

### Identifiers

- Primary key: `id`. Natural key: `tool_invocation_id` unique (1:1). OS process IDs are
  external identifiers, recorded only inside Events for diagnostics.

### Relations

- Attributed to exactly one **Run**; records exactly one **Tool Invocation**'s process.
- Captures spill to **Artifact** rows.

### Integrity invariants

1. **INV-CMD-01** — Exactly one Command Execution exists per terminal-command Tool Invocation
   that started a process; commands cannot run outside an invocation (PRD-004).
2. **INV-CMD-02** — Environment variable *values* MUST NOT be persisted anywhere in this
   record; only the policy reference and, at most, variable names (Volume 9 redaction).
3. **INV-CMD-03** — `outcome` MUST be consistent with `exit_code`/`signal` (`succeeded` ⇔
   exit code 0; mapping owned by Volume 6).
4. **INV-CMD-04** — Command Executions are immutable records.

### Lifecycle

Immutable record with recorded `outcome`; process management (timeouts, kill escalation) is
owned by Volume 6 (Terminal Engine contract) and Volume 9 (sandbox).

### Persistence

Workspace database, table `command_executions`; captures as Artifacts. Retention: with the
owning run.

### Versioning and serialization

No `revision`. Serializes in the run record stream with captures by reference.
