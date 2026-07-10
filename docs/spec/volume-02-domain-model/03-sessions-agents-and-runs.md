# 03 — Sessions, Agents, and Runs

This chapter defines the Session aggregate (**Session**), the Agent Profile aggregate
(**Agent Profile**), and the Run aggregate members that carry execution: **Run** (root),
**Agent**, **Turn**, **Message**, **Plan**, and **Task**. Full state machines for all
stateful entities here are owned by Volume 4 (chapter 09 ownership table); this chapter
defines shapes, relations, and invariants.

## Session

Purpose: a bounded interactive or non-interactive engagement with Andromeda, holding runs,
context, and session memory; persistable and resumable (PRD-010).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `workspace_id` | `ulid` | yes | Hosting Workspace |
| `project_id` | `ulid` | no | Focused Project, when one is pinned |
| `title` | `string` | yes | Display title; user-set or generated from the first goal |
| `mode` | `enum` | yes | `interactive` \| `non_interactive` (PRD-009: both share this one entity) |
| `state` | `enum` | yes | Canonical Session state (chapter 09) |
| `default_agent_profile_id` | `ulid` | yes | Agent Profile used for new runs unless overridden |
| `config_profile_ids` | `json` | no | Ordered list of Configuration Profile IDs active for the session (resolution per Volume 10) |
| `run_count` | `integer` | yes | Cached count of Runs (derived; authoritative count is the `runs` table) |
| `started_at` | `timestamp` | yes | First activation |
| `last_active_at` | `timestamp` | yes | Last run activity or user interaction |
| `ended_at` | `timestamp` | no | Set when the session reaches a terminal state |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. No natural key: titles are not unique. CLI/TUI accept unambiguous ULID
  prefixes for session addressing (grammar owned by Volume 8).

### Relations

- Belongs to exactly one **Workspace**; optionally focuses one **Project**.
- Contains 0..n **Run** and 0..n **Workflow Run** (both reference `session_id`; both are their
  own aggregates).
- Scopes session-layer **Memory Record** rows and session-scoped **Permission** grants.

### Integrity invariants

1. **INV-SES-01** — A Session MUST belong to exactly one Workspace and never migrates between
   workspaces.
2. **INV-SES-02** — A Session in a terminal state MUST NOT accept new Runs; resuming a
   suspended session precedes starting a run in it.
3. **INV-SES-03** — Interactive and non-interactive sessions MUST use the same entity, tables,
   and invariants; `mode` changes presentation and permission resolution (Volume 9), never
   shape (PRD-009).
4. **INV-SES-04** — `ended_at` MUST be set if and only if `state` is terminal.
5. **INV-SES-05** — Session deletion MUST cascade to its session-scoped Memory Records and
   Permission grants, and is refused while any of its Runs is non-terminal.

### Lifecycle

Stateful — canonical states `created`, `active`, `suspended`, `ended`, `failed`
(chapter 09); full machine owned by Volume 4. Sessions persist incrementally so that recovery
after a crash finds a resumable `suspended` session rather than a corrupt one (PRD-010;
recovery semantics in Volume 4, durability in Volume 10).

### Persistence

Workspace database, table `sessions`. Retention: sessions are kept until explicitly deleted;
retention policies (age/size pruning of ended sessions) are owned by Volume 10.

### Versioning and serialization

Row versioning via `revision`. A session export (for bug reports or transfer) is a canonical
JSON document stream: the Session document followed by its Runs' documents; export redaction
rules are owned by Volume 9.

## Agent Profile

Purpose: a named, versioned configuration of an agent — model/provider selection, prompts,
tool set, permission defaults, and behavioral parameters.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key (one row per profile *version*) |
| `name` | `string` | yes | Profile name (e.g., `default`, `reviewer`) |
| `version` | `semver` | yes | Version of this profile definition |
| `scope` | `enum` | yes | `builtin` \| `global` \| `workspace` |
| `workspace_id` | `ulid` | conditional | Required when `scope = workspace` |
| `description` | `text` | no | Human description |
| `model_selector` | `json` | yes | Provider slug + model name selectors, with optional routing/fallback preferences (semantics owned by Volume 5) |
| `prompt_refs` | `json` | yes | References to versioned prompt templates of the Prompt Engine (Volume 4) |
| `tool_policy` | `json` | yes | Tool allow/deny selectors evaluated by the Tool Runtime (Volume 6) |
| `permission_defaults` | `json` | yes | Default permission stance; vocabulary owned by Volume 9 |
| `parameters` | `json` | no | Behavioral parameters (temperature-class knobs, iteration limits); keys owned by Volume 4 |
| `deprecated` | `boolean` | yes | Version is discouraged for new use (kept for reproducibility) |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change (metadata only; see INV-AGP-02) |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(scope, workspace_id, name, version)` unique.
- `name` without version refers to the highest non-deprecated version at resolution time
  (resolution rules owned by Volume 4).

### Relations

- Referenced by **Session** (`default_agent_profile_id`), **Agent** (instantiation snapshot),
  and **Workflow** step definitions (Volume 4).
- References **Provider**/**Model** by selector (not by FK — providers may be absent on this
  machine), prompt templates (Volume 4), and Tools by selector.

### Integrity invariants

1. **INV-AGP-01** — `(scope, workspace_id, name, version)` MUST be unique.
2. **INV-AGP-02** — A published profile version is immutable in its behavioral fields
   (`model_selector`, `prompt_refs`, `tool_policy`, `permission_defaults`, `parameters`);
   changing behavior mints a new `version`. Only `description` and `deprecated` may change
   in place.
3. **INV-AGP-03** — A profile version that any persisted Agent references MUST NOT be deleted;
   it is marked `deprecated` instead (reproducibility, PRD-006).
4. **INV-AGP-04** — `permission_defaults` MUST NOT widen beyond what Volume 9 allows a profile
   to grant; profiles can narrow autonomy, only users and policies can widen it.

### Lifecycle

Stateless (versioned catalog; `deprecated` is a recorded flag, not a machine).

### Persistence

`builtin` profiles ship in the binary and are materialized read-only at startup; `global`
scope: global database, table `agent_profiles`; `workspace` scope: workspace database, same
table. Retention: versions persist while referenced (INV-AGP-03).

### Versioning and serialization

Semantic `version` per definition; `revision` for row metadata. Profiles serialize as
canonical JSON documents; TOML rendering for human editing per ADR-008.

## Run

Purpose: one top-level execution of an agent or workflow step within a session, from intake to
terminal state. The Run aggregate is the complete, auditable causal record of that execution
(PRD-006, PRD-010).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `session_id` | `ulid` | yes | Owning Session |
| `sequence_no` | `integer` | yes | Position of this run within its session (1-based, dense) |
| `initiator` | `enum` | yes | `user` \| `workflow` \| `resume` \| `system`; who/what started the run |
| `workflow_run_id` | `ulid` | no | Set when the run was spawned by a Workflow Run |
| `goal` | `text` | yes | The user goal or step objective the run was started with |
| `state` | `enum` | yes | Canonical Run state (chapter 09) |
| `root_agent_id` | `ulid` | yes | The run's root Agent instance |
| `config_snapshot` | `json` | yes | Resolved configuration at start: Configuration Profile IDs + resolved values digest, Agent Profile version IDs — sufficient to reproduce resolution (PRD-006) |
| `budgets` | `json` | no | Declared limits: max tokens, max cost (micro-units + currency), max duration, max tool invocations (enforcement owned by Volume 4/12) |
| `error` | `json` | no | Terminal error summary (stable error code + safe context) when `state = failed` |
| `usage_totals` | `json` | yes | Cached aggregate of Cost Records (tokens, cost); authoritative data is the Cost Record table |
| `started_at` | `timestamp` | yes | Transition out of `pending` |
| `ended_at` | `timestamp` | no | Terminal transition instant |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(session_id, sequence_no)` unique.
- The Run `id` is the primary correlation identifier across Events, Traces, Cost Records, and
  Audit Records (chapter 08).

### Relations

- Belongs to exactly one **Session**; optionally spawned by one **Workflow Run**.
- Owns (aggregate members): 1..n **Agent**, 0..n **Turn**, 0..n **Plan**, 0..n **Task** (via
  plans), 0..n **Tool Invocation** (with their **Tool Result** rows), 0..n **File Change**,
  0..n **Patch**, 0..n **Command Execution**, 0..n **Artifact**, 0..n **Context Item** (via
  turns).
- Correlated with exactly one **Trace** and 0..n **Event**, **Cost Record**, **Audit Record**
  rows (observability aggregates, chapter 08).

### Integrity invariants

1. **INV-RUN-01** — A Run MUST belong to exactly one Session; `workflow_run_id`, when set,
   MUST reference a Workflow Run in the same Session.
2. **INV-RUN-02** — `(session_id, sequence_no)` MUST be unique, dense, and assigned in start
   order; it is the authoritative run ordering within a session.
3. **INV-RUN-03** — A Run in a terminal state is frozen: no member entity may be added to it,
   and no member's content may change (redaction and retention excepted).
4. **INV-RUN-04** — `config_snapshot` MUST be written before the first Turn starts and MUST
   NOT change afterward: what configuration a run executed under is a fact, not a lookup.
5. **INV-RUN-05** — Every side effect of a run on the environment MUST be attributable through
   member records: no File Change, Command Execution, or Artifact exists without its producing
   Tool Invocation (PRD-004, PRD-006).
6. **INV-RUN-06** — After a crash, a Run found in a non-terminal state MUST be surfaced as
   `interrupted` — never silently completed, never silently discarded (PRD-010; recovery
   procedure owned by Volume 4).

### Lifecycle

Stateful — canonical states `pending`, `planning`, `running`, `awaiting_approval`, `paused`,
`interrupted`, `completed`, `failed`, `cancelled` (chapter 09); full machine owned by
Volume 4.

### Persistence

Workspace database, table `runs`; members persist incrementally as they are created (PRD-010).
Retention: run records are kept until their session is deleted or Volume 10 retention policy
prunes ended runs; audit obligations (Volume 9) take precedence over pruning.

### Versioning and serialization

Row versioning via `revision`. A run export is the canonical JSON document stream of the run
and all member records in `sequence_no`/creation order — this is the "run record" that Volume 1
binds to reproducibility (SM-12) and that replay mode (Volumes 4/10) consumes.

## Agent

Purpose: an autonomous executor instance that plans and acts through tools under a profile,
permissions, and observability. Agent rows are *instances* bound to one run; the reusable
definition is the Agent Profile.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Owning Run |
| `parent_agent_id` | `ulid` | no | The delegating Agent for sub-agents; absent for the root agent |
| `agent_profile_id` | `ulid` | yes | Agent Profile *version row* this instance was instantiated from (snapshot reference) |
| `role` | `string` | yes | Instance role label within the run (e.g., `root`, `explorer`); vocabulary owned by Volume 4 |
| `state` | `enum` | yes | Canonical Agent state (chapter 09) |
| `provider_slug` | `string` | yes | Provider actually resolved for this instance (announced per Principle 7) |
| `model_name` | `string` | yes | Model actually resolved for this instance |
| `error` | `json` | no | Terminal error summary when `state = failed` |
| `started_at` | `timestamp` | yes | Instantiation |
| `ended_at` | `timestamp` | no | Terminal transition instant |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. No natural key; agents are addressed through their run.

### Relations

- Belongs to exactly one **Run** (aggregate member); optional parent **Agent** (delegation
  tree).
- Instantiated from exactly one **Agent Profile** version.
- Drives 0..n **Turn** and issues 0..n **Tool Invocation**.

### Integrity invariants

1. **INV-AG-01** — An Agent MUST belong to exactly one Run; the delegation tree
   (`parent_agent_id`) MUST be acyclic and rooted at the run's `root_agent_id`.
2. **INV-AG-02** — Exactly one Agent per Run has no parent, and it is the run's root agent.
3. **INV-AG-03** — `agent_profile_id` MUST reference an immutable profile version
   (INV-AGP-02); mid-run profile edits never affect running instances.
4. **INV-AG-04** — Provider/model changes for an instance (fallback, routing) MUST be recorded
   as Events and reflected in subsequent Turns' provider/model fields — never silent
   (Principle 7).

### Lifecycle

Stateful — canonical states `instantiated`, `idle`, `thinking`, `acting`, `waiting`,
`terminated`, `failed` (chapter 09); full machine owned by Volume 4.

### Persistence

Workspace database, table `agents`. Retention: with the owning run.

### Versioning and serialization

Row versioning via `revision`; serializes inside the run record stream.

## Turn

Purpose: one request/response exchange inside a run — one model request by one agent and
everything returned for it.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Owning Run |
| `agent_id` | `ulid` | yes | Agent that drove the turn |
| `sequence_no` | `integer` | yes | Position within the run (1-based, dense across all agents of the run) |
| `status` | `enum` | yes | Recorded status: `in_progress` \| `completed` \| `failed` \| `interrupted` (recorded vocabulary, chapter 09) |
| `provider_slug` | `string` | yes | Provider used for this turn |
| `model_name` | `string` | yes | Model used for this turn |
| `usage` | `json` | no | Token usage reported by the provider (official usage data only) |
| `error` | `json` | no | Error summary when `status = failed` |
| `started_at` | `timestamp` | yes | Request assembly start |
| `ended_at` | `timestamp` | no | Response fully received / terminal |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(run_id, sequence_no)` unique; authoritative turn order.

### Relations

- Belongs to exactly one **Run**; driven by exactly one **Agent**.
- Carries 1..n **Message** and 0..n **Context Item** (the assembled request context,
  chapter 07).
- Produces 0..n **Tool Invocation** (tool calls requested in the model response are issued
  under this turn).

### Integrity invariants

1. **INV-TRN-01** — A Turn MUST belong to exactly one Run and exactly one Agent of that run.
2. **INV-TRN-02** — At most one Turn per Agent may have `status = in_progress` at any time.
3. **INV-TRN-03** — `provider_slug`/`model_name` MUST record what was actually used, per turn;
   a mid-run fallback shows up as differing values between consecutive turns plus the Event of
   INV-AG-04.
4. **INV-TRN-04** — Token usage in `usage` MUST come from official provider accounting;
   estimated values live only in Cost Records marked as estimates (chapter 08).

### Lifecycle

Recorded status only (`in_progress`, `completed`, `failed`, `interrupted`); not a canonical
state machine — the run-loop semantics that move turns are owned by Volume 4.

### Persistence

Workspace database, table `turns`. Retention: with the owning run.

### Versioning and serialization

Row versioning via `revision`; serializes inside the run record stream.

## Message

Purpose: a single unit of conversation content (user, agent, system, or tool) inside a turn.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `turn_id` | `ulid` | yes | Owning Turn |
| `run_id` | `ulid` | yes | Owning Run (denormalized for query paths; MUST match the turn's run) |
| `sequence_no` | `integer` | yes | Position within the turn (1-based, dense) |
| `role` | `enum` | yes | `user` \| `agent` \| `system` \| `tool` (closed; matches the glossary roles) |
| `parts` | `json` | yes | Ordered content parts; part kinds form a closed vocabulary owned by Volume 4 (baseline: `text`, `file_ref`, `image_ref`, `tool_call`, `tool_result_ref`, `reasoning_summary`) |
| `origin` | `enum` | yes | `human` \| `model` \| `runtime` \| `tool_runtime`; producer of the message |
| `redaction_state` | `enum` | yes | `none` \| `partial` \| `full`; redaction semantics owned by Volume 9 |
| `created_at` | `timestamp` | yes | Creation instant |

### Identifiers

- Primary key: `id`. Natural key: `(turn_id, sequence_no)` unique.

### Relations

- Belongs to exactly one **Turn** (and transitively one Run).
- `tool_call` / `tool_result_ref` parts reference **Tool Invocation** / **Tool Result** rows
  by ID.

### Integrity invariants

1. **INV-MSG-01** — Messages are immutable records: after commit, only redaction (Volume 9)
   may alter content, and redaction MUST set `redaction_state` and preserve part structure.
2. **INV-MSG-02** — A Message MUST belong to exactly one Turn; `run_id` MUST equal the turn's
   `run_id`.
3. **INV-MSG-03** — Every `tool_call` part MUST correspond to exactly one Tool Invocation row
   of the same run; every `tool_result_ref` part MUST reference an existing Tool Result.
4. **INV-MSG-04** — `parts` kinds MUST come from the Volume 4 closed vocabulary; unknown kinds
   fail validation rather than passing through opaquely.
5. **INV-MSG-05** — A Message MUST NOT claim to contain a model's private internal reasoning;
   `reasoning_summary` parts carry only officially provided summaries (Principle 7).

### Lifecycle

Immutable record.

### Persistence

Workspace database, table `messages`. Large binary content is not inlined: `file_ref` /
`image_ref` parts point to Artifact content (chapter 04). Retention: with the owning run.

### Versioning and serialization

Immutable; no `revision`. Part payloads carry their own `schema_version` inside the run record
stream (chapter 10).

## Plan

Purpose: a structured, inspectable set of intended steps produced by the Planner for a run.
Plans are versioned within the run: revising a plan supersedes the previous version instead of
rewriting it (inspectability, PRD-006).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `run_id` | `ulid` | yes | Owning Run |
| `version` | `integer` | yes | Plan version within the run (1-based, dense) |
| `state` | `enum` | yes | Canonical Plan state (chapter 09) |
| `produced_by_agent_id` | `ulid` | yes | Agent that produced this plan version |
| `objective` | `text` | yes | What this plan intends to achieve |
| `rationale` | `text` | no | Planner-provided rationale summary |
| `supersedes_id` | `ulid` | no | The Plan version this one replaced |
| `approved_by` | `ulid` | no | Approval that accepted the plan, when plan approval was required (Volume 9/4 policy) |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(run_id, version)` unique.

### Relations

- Belongs to exactly one **Run**; produced by one **Agent**.
- Decomposes into 0..n **Task** (aggregate members under the same Run).
- Optionally accepted by one **Approval**; optionally supersedes one earlier **Plan**.

### Integrity invariants

1. **INV-PLAN-01** — `(run_id, version)` MUST be unique and dense; `supersedes_id`, when set,
   MUST reference version `version - 1` of the same run.
2. **INV-PLAN-02** — At most one Plan per Run may be in a non-terminal state at any time; a
   new version may be created only after its predecessor reached `superseded` or another
   terminal state.
3. **INV-PLAN-03** — A Plan MUST reach `approved` before any of its Tasks may start when the
   active policy requires plan approval; auto-approval by policy is recorded via an Approval
   with `decided_by_kind = policy` (chapter 04).
4. **INV-PLAN-04** — Terminal plans and their tasks are frozen (INV-RUN-03 applies).

### Lifecycle

Stateful — canonical states `draft`, `proposed`, `approved`, `executing`, `revising`,
`completed`, `superseded`, `abandoned` (chapter 09); full machine owned by Volume 4.

### Persistence

Workspace database, table `plans`. Retention: with the owning run.

### Versioning and serialization

Plan `version` is the in-run revision lineage; `revision` is row concurrency. Serializes
inside the run record stream with its tasks.

## Task

Purpose: a unit of executable work derived from a plan, with its own state machine — the
granularity at which the Execution Engine schedules, retries, cancels, and reports.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `plan_id` | `ulid` | yes | Owning Plan |
| `run_id` | `ulid` | yes | Owning Run (denormalized; MUST match the plan's run) |
| `ordinal` | `integer` | yes | Position in the plan's declared order (1-based, dense) |
| `title` | `string` | yes | Short imperative description |
| `description` | `text` | no | Full task statement given to the executing agent |
| `state` | `enum` | yes | Canonical Task state (chapter 09) |
| `depends_on` | `json` | no | List of Task IDs in the same plan that must be terminal-successful first |
| `assigned_agent_id` | `ulid` | no | Agent executing the task, once assigned |
| `attempt` | `integer` | yes | Execution attempt counter (1-based; retry policy owned by Volume 4) |
| `result_summary` | `text` | no | Outcome summary on completion |
| `error` | `json` | no | Error summary when `state = failed` |
| `started_at` | `timestamp` | no | First transition to `running` |
| `ended_at` | `timestamp` | no | Terminal transition instant |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(plan_id, ordinal)` unique.

### Relations

- Belongs to exactly one **Plan** (and transitively one Run).
- Depends on 0..n sibling **Task** rows; fulfilled by 0..n **Tool Invocation**.
- Optionally assigned to one **Agent**.

### Integrity invariants

1. **INV-TASK-01** — A Task MUST belong to exactly one Plan; `run_id` MUST equal the plan's
   `run_id`; `depends_on` MUST reference only Tasks of the same Plan.
2. **INV-TASK-02** — The dependency graph over a plan's Tasks MUST be acyclic.
3. **INV-TASK-03** — A Task MUST NOT enter `running` while any dependency is not in a
   successful terminal state (`completed` or `skipped`); the guard details are Volume 4's.
4. **INV-TASK-04** — A Task interrupted by a crash MUST be reported as `interrupted`, never
   assumed complete (PRD-010).
5. **INV-TASK-05** — Task state changes MUST be persisted before their side effects are
   reported to the user (write-ahead of presentation; durability rules in Volume 10).

### Lifecycle

Stateful — canonical states `pending`, `ready`, `running`, `blocked`, `awaiting_approval`,
`interrupted`, `completed`, `failed`, `cancelled`, `skipped` (chapter 09); full machine owned
by Volume 4.

### Persistence

Workspace database, table `tasks`. Retention: with the owning run.

### Versioning and serialization

Row versioning via `revision`; serializes inside the run record stream under its plan.
