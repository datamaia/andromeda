# Annex — Consolidated Event Catalog

**Status:** Consolidated (Phase C). This annex is the corpus-wide index of every event name
minted in Volumes 3–14, aggregated from the volume registers' "Events minted" sections and
the defining chapters. It is a *reference view*: payload schemas, correlation requirements,
and emission rules live in the linked defining chapters, which are normative. This annex
mints nothing and renames nothing.

## The envelope, in one paragraph

Every event is one enveloped occurrence under the keystone FR-OBS-001 (Volume 10, chapter
[04](../volume-10-config-storage-observability/04-events-and-envelope.md); decision record
ADR-137). The envelope carries three persisted field groups — **identity** (`id` ULID,
registered `name`, `schema_version`, `occurred_at`, `producer`), **correlation**
(`workspace_id`, `session_id`, `run_id`, `turn_id`, `task_id`, `tool_invocation_id`,
`trace_id`, each present when that scope applies; `payload.provider_request_id` and
`payload.span_id` are reserved payload keys), and **content** (the redacted, canonical-JSON
`payload` whose schema the minting area owns; `payload.error_code` carries the `E-<AREA>-NNN`
identity on failure events) — plus a transit-only **delivery** group (`family`, bridge
`attempt`). Names follow the Volume 0 chapter 03 grammar `<area>.<noun>.<verb-past>`. Every
name is registered in a closed, compiled-in registry; free-form emission is rejected
(E-OBS-002). Each name belongs to exactly one delivery family — `lifecycle`, `action`,
`progress`, `security`, or `telemetry` — which fixes its overflow policy and default buffer;
lifecycle and action events are persisted transactionally with the writes they describe
before bus publication (persist-then-publish, ADR-137). Delivery, ordering, IPC bridging,
persistence, retention, and export semantics are all Volume 10 chapter 04's and are not
restated per row below.

## Reading the tables

One table per area family, in volume order. Columns: the event name; the producer component
(Volume 3 component names); a one-line meaning with a payload gloss where the defining
chapter publishes one; and the defining chapter. Producers and meanings are reproduced from
the registers and defining chapters; where a register lists only names, the meaning column
follows the defining chapter's event table.

## Runtime, PAL, scheduler, IPC (Volume 3)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `pal.platform.rejected` | PAL | Startup refused on an unsupported platform (E-PORT-001) | [Vol 3 ch 07](../volume-03-architecture/07-platform-abstraction-layer.md) |
| `pal.capability.degraded` | PAL | A PAL surface capability is absent or degraded (E-PORT-002/E-PORT-003) | [Vol 3 ch 07](../volume-03-architecture/07-platform-abstraction-layer.md) |
| `pal.fallback.engaged` | PAL | Directory-resolution fallback engaged (ADR-022) | [Vol 3 ch 07](../volume-03-architecture/07-platform-abstraction-layer.md) |
| `runtime.recovery.completed` | Runtime | Startup recovery finished, with reconciliation counts (FR-ARCH-009) | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| `runtime.shutdown.completed` | Runtime | Shutdown finished, orderly or forced, with step timings (FR-ARCH-010) | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| `scheduler.task.rejected` | Task Scheduler | Bounded-pool submission rejected (E-ARCH-005) | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| `ipc.client.connected` | IPC server | A verified same-user client completed the handshake | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| `ipc.request.rejected` | IPC server | A connection or request was rejected (identity, version, or endpoint state) | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |

## Sessions, workspaces, runs, turns, agents (Volume 4)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `session.created` | Runtime | Session row created (session ID, workspace ID, mode) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `session.activated` | Runtime | `created`/`suspended` → `active` (session ID) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `session.suspended` | Runtime | `active` → `suspended` (session ID, reason) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `session.resumed` | Runtime | Resume completed (session ID) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `session.ended` | Runtime | Terminal `ended` (session ID, run counts) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `session.failed` | Runtime | Terminal `failed` (session ID, error code) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `workspace.discovered` | Workspace Engine | Discovery resolved (root path class, marker kind) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `workspace.opened` | Workspace Engine | Open completed (workspace ID) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `workspace.closed` | Workspace Engine | Close completed (workspace ID) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.created` | Agent Engine | Run accepted, state `pending` (run ID, goal digest, initiator) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.started` | Agent Engine | `pending` → `planning` (run ID, root agent ID) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.planned` | Agent Engine | `planning` → `running` (run ID, plan ID/version) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.revision.started` | Agent Engine | `running` → `planning` (run ID, trigger) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.approval.requested` | Agent Engine | Transition into `awaiting_approval` (run ID, approval ID, subject) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.paused` | Agent Engine | Transition into `paused` (run ID) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.resumed` | Agent Engine | `paused`/`interrupted`/`awaiting_approval` → active states (run ID, resume report digest) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.interrupted` | Runtime (recovery) | Transition into `interrupted` (run ID, incarnation) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.completed` | Agent Engine | Terminal `completed` (run ID, usage totals) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.failed` | Agent Engine | Terminal `failed` (run ID, error code) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.cancelled` | Agent Engine | Terminal `cancelled` (run ID, reason) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `run.budget.exhausted` | Agent Engine | Budget breach detected (run ID, budget kind, limit, accumulated) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `turn.started` | Agent Engine | Turn request issued (turn ID, agent ID, provider, model) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `turn.completed` | Agent Engine | Turn closed successfully (turn ID, usage) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `turn.failed` | Agent Engine | Turn closed with error (turn ID, error code) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `turn.interrupted` | Agent Engine | Turn cut by stop (turn ID) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.instantiated` | Agent Engine | Agent row created (agent ID, profile version, role) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.state.changed` | Agent Engine | Any Agent machine transition (agent ID, from, to) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.delegation.started` | Agent Engine | Sub-agent spawned (parent ID, child ID, role) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.delegation.completed` | Agent Engine | Sub-agent result returned (parent ID, child ID, outcome) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |

## Plans and tasks (Volume 4)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `plan.drafted` | Planner | Plan row persisted, state `draft` (plan ID, version, task count) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.proposed` | Planner | `draft` → `proposed` (plan ID, version) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.approved` | Planner | `proposed` → `approved` (plan ID, approval ID, decider kind) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.execution.started` | Execution Engine | `approved` → `executing` (plan ID) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.revision.started` | Planner | `executing` → `revising` (plan ID, trigger) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.revision.abandoned` | Planner | `revising` → `executing` (plan ID, reason) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.superseded` | Planner | Predecessor terminal, successor drafted (plan ID, successor ID) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.completed` | Execution Engine | All tasks terminal-successful (plan ID) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `plan.abandoned` | Planner | Discarded without completion (plan ID, reason) | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `task.created` | Planner | Task rows persisted with their plan (task ID, plan ID, ordinal) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.scheduled` | Execution Engine | `pending` → `ready` (task ID) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.started` | Execution Engine | `ready` → `running` (task ID, agent ID, attempt) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.blocked` | Execution Engine | `running` → `blocked` (task ID, condition) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.unblocked` | Execution Engine | `blocked` → `ready` (task ID) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.approval.requested` | Execution Engine | Transition into `awaiting_approval` (task ID, approval ID) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.resumed` | Execution Engine | `awaiting_approval` → `running` (task ID, decision) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.retried` | Execution Engine | Retry scheduled (task ID, attempt, delay) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.interrupted` | Runtime (recovery) | Transition into `interrupted` (task ID) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.completed` | Execution Engine | Terminal success (task ID, result summary digest) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.failed` | Execution Engine | Terminal failure (task ID, error code, attempts) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.cancelled` | Execution Engine | Terminal cancellation (task ID, reason) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `task.skipped` | Execution Engine | Deliberate non-execution (task ID, reason) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `prompt.template.registered` | Prompt Engine | Template accepted into the registry (namespace, name, version, source) | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |
| `prompt.template.overridden` | Prompt Engine | An override shadows a lower-precedence version (name, winning and shadowed sources) | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |
| `prompt.rendered` | Prompt Engine | Render completed for a consumer (provenance tuple, output size class, consumer turn ID) | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |

## Workflows (Volume 4)

Producer for every row: Workflow Engine.

| Event | Meaning / payload | Defined in |
|---|---|---|
| `workflow.definition.registered` | A definition version passed validation and was registered | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.definition.rejected` | A definition source failed validation (E-WF-001/E-WF-002) or trust policy | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.run.started` | Instantiated run began executing | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| `workflow.run.paused` | Explicit user pause took effect at a step boundary | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| `workflow.run.resumed` | Resume from `paused` or `interrupted` | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| `workflow.run.interrupted` | Crash/shutdown marking | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| `workflow.run.completed` | Terminal success with outputs recorded | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| `workflow.run.failed` | Terminal failure with persisted error | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| `workflow.run.cancelled` | Terminal cancellation with recorded reason | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| `workflow.step.entered` | A step began (dispatch or gate raise) | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.step.completed` | A step reached `completed` | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.step.failed` | A step reached `failed` (E-WF-008) | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.step.retried` | A failed attempt re-dispatched under its retry policy | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.step.skipped` | A step was deliberately not executed (routing/condition) | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.step.timed_out` | A step timer fired (E-WF-009) | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.step.deferred` | Work deferred on scheduler backpressure (E-ARCH-005 handled) | [Vol 4 ch 09](../volume-04-agent-runtime/09-task-scheduler.md) |
| `workflow.gate.requested` | An Approval with subject kind `workflow_gate` was raised | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.gate.granted` | Gate Approval terminal `granted` | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.gate.denied` | Gate Approval terminal `denied` (E-WF-006) | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.gate.expired` | Gate Approval terminal `expired` (E-WF-007) | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.artifact.recorded` | A declared step artifact was recorded | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.rollback.started` | Compensation sequence began | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.rollback.completed` | All applicable compensations completed | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.rollback.failed` | Rollback halted on a failed compensation (E-WF-011) | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflow.skillset.resolved` | A run's skill snapshot was resolved and recorded | [Vol 4 ch 08](../volume-04-agent-runtime/08-skill-engine-runtime.md) |
| `workflow.skillset.degraded` | An optional skill was omitted, or a required skill failed (E-WF-013) | [Vol 4 ch 08](../volume-04-agent-runtime/08-skill-engine-runtime.md) |
| `workflow.timer.armed` | A durable deadline was persisted | [Vol 4 ch 09](../volume-04-agent-runtime/09-task-scheduler.md) |
| `workflow.timer.fired` | A deadline fired and its transition applied | [Vol 4 ch 09](../volume-04-agent-runtime/09-task-scheduler.md) |
| `workflow.timer.restored` | Timers re-armed from persisted deadlines at recovery/resume | [Vol 4 ch 09](../volume-04-agent-runtime/09-task-scheduler.md) |

## Providers (Volume 5)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `provider.request.completed` | Provider Layer | Request finished (provider slug, model, method, latency ms, usage summary, finish reason) | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `provider.request.failed` | Provider Layer | Request failed (provider slug, model, method, E-PROV code, retryable flag, attempt count) | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `provider.discovery.completed` | Provider Layer | Model discovery reconciled (provider slug, counts offered/new/deprecated, duration ms) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `provider.model.deprecated` | Provider Layer | A model was marked deprecated (provider slug, model name, evidence) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `provider.capability.changed` | Provider Layer | A model's effective capability set changed (old set, new set, provenance) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `provider.capability.verified` | Provider Layer | A capability probe concluded (capability, outcome `verified`/`refuted`) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `provider.degradation.applied` | Provider Layer | A degradation strategy was applied (capability, strategy, provider, model, reason) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `provider.stream.interrupted` | Provider Layer | A stream ended abnormally (delivered event count, E-PROV code, delivered-usage snapshot) | [Vol 5 ch 03](../volume-05-providers-and-auth/03-streaming-toolcalling-structured-outputs.md) |
| `provider.cost.recorded` | Provider Layer / Runtime | Per-request accounting recorded (token fields, cost micros + currency when present, cost basis, attempt) | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `provider.request.retried` | Provider Router | A retry was scheduled (attempt number, delay ms, trigger E-PROV code, delay source) | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `provider.route.selected` | Provider Router | Routing chose a target (provider, model, strategy, reason chain) | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `provider.fallback.activated` | Provider Router | A fallback chain step activated (chain name, from/to, trigger class, guard summary, approval reference) | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `provider.breaker.opened` | Provider Router | Circuit breaker opened (provider slug, window statistics, consecutive-reopen count) | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `provider.breaker.closed` | Provider Router | Circuit breaker closed after a probe (provider slug, probe latency ms) | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `provider.adapter.registered` | Provider Layer | An adapter declaration passed validation (adapter ID, adapter version, contract version) | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| `provider.adapter.rejected` | Provider Layer | An adapter declaration was refused (adapter ID, adapter version, findings summary; E-PROV-019) | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| `provider.deprecation.announced` | Provider Layer | An adapter detected a provider-documented deprecation (subject kind, subject name, sunset date when available) | [Vol 5 ch 09](../volume-05-providers-and-auth/09-provider-adapters-catalog.md) |

## Authentication and provider connections (Volume 5)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `auth.credential.created` | Authentication Layer | Intake stored a new Credential (credential ULID, label, kind, provider slug, fingerprint) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.credential.rotated` | Authentication Layer | Rotation completed with successor linkage (predecessor ULID, successor ULID, label) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.credential.revoked` | Authentication Layer | Revocation completed (payload marks provider-side outcome) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.credential.expired` | Authentication Layer | Temporary credential reached expiry (credential ULID, label, expiry instant) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.credential.deleted` | Authentication Layer | Slot and row deletion completed (credential ULID, label) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.credential.access_failed` | Authentication Layer | Secret Store resolution failed (backend kind, error code; E-AUTH-006) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.credential.rotation_failed` | Authentication Layer | Rotation aborted (credential ULID, failing step; E-AUTH-008) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.mechanism.refused` | Authentication Layer | FR-AUTH-001 gate refused a mechanism (mechanism name, provider slug, requester; E-AUTH-007) | [Vol 5 ch 07](../volume-05-providers-and-auth/07-authentication-layer.md) |
| `auth.profile.selected` | Authentication Layer | Profile resolution succeeded at establishment (profile name, provider slug) | [Vol 5 ch 07](../volume-05-providers-and-auth/07-authentication-layer.md) |
| `auth.profile.resolution_failed` | Authentication Layer | Profile resolution failed (provider slug, candidate names; E-AUTH-010) | [Vol 5 ch 07](../volume-05-providers-and-auth/07-authentication-layer.md) |
| `auth.session.established` | Authentication Layer | Authentication Session reached `active` | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `auth.session.refreshed` | Authentication Layer | Renewal replaced token material | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `auth.session.expired` | Authentication Layer | Session reached `expired` | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `auth.session.failed` | Authentication Layer | Establishment or renewal failed; also the cancellation class | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `auth.session.revoked` | Authentication Layer | Session terminated by credential revocation or rotation | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `provider.connection.verified` | Provider Layer | Verification succeeded; provider `available` | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `provider.connection.degraded` | Provider Layer | Degradation thresholds crossed | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `provider.connection.recovered` | Provider Layer | Returned to `available` after degradation | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `provider.connection.lost` | Provider Layer | Provider became `unavailable` | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `provider.connection.disabled` | Provider Layer | Administrative disable | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `provider.connection.enabled` | Provider Layer | Administrative enable (verification follows) | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |
| `provider.connection.removed` | Provider Layer | Deregistration tombstone | [Vol 5 ch 11](../volume-05-providers-and-auth/11-state-machines.md) |

## Tools and terminal (Volume 6)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `tool.registration.completed` | Tool Runtime | Registration accepted (tool name/version/origin, trust level, scope) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tool.registration.rejected` | Tool Runtime | Registration refused (name, origin, violated rule, colliding party; E-TOOL-002 detail) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tool.enablement.changed` | Tool Runtime | Enable/disable applied (tool identity, new flag, actor) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tool.invocation.requested` | Tool Runtime | Invocation row created (invocation ID, tool snapshot, run/turn/task/agent IDs) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.approved` | Tool Runtime | Approval resolved allow (invocation ID, approval ID or grant refs, decider kind) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.denied` | Tool Runtime | Invocation denied (invocation ID, deciding record, decider kind, requested permissions) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.started` | Tool Runtime | Execution began (invocation ID, sandbox containment level, effective limits) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.succeeded` | Tool Runtime | Terminal success (invocation ID, duration, payload size, truncated flag) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.failed` | Tool Runtime | Terminal failure (invocation ID, E-TOOL code, tool-local code where present) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.timed_out` | Tool Runtime | Terminal timeout (invocation ID, effective timeout, teardown timings) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.cancelled` | Tool Runtime | Terminal cancellation (invocation ID, reason, teardown timings where applicable) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.invocation.retried` | Tool Runtime | Automatic retry created (new and prior invocation IDs, attempt number; ADR-072) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `tool.output.truncated` | Tool Runtime | Spillover applied (invocation ID, untruncated size, artifact ID; ADR-071) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `terminal.execution.started` | Terminal Engine | Command process started (execution ID, invocation ID, pty flag, sandbox profile ref) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `terminal.execution.ended` | Terminal Engine | Command process ended (execution ID, outcome, exit code/signal, durations) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |
| `terminal.output.truncated` | Terminal Engine | Capture cap reached (execution ID, stream, captured/total bytes) | [Vol 6 ch 04](../volume-06-tools-mcp-skills-plugins/04-tool-invocation-state-machine.md) |

## MCP, skills, plugins, packages (Volume 6)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `mcp.server.registered` | MCP Runtime | Server registration accepted (server name, scope, transport) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.server.updated` | MCP Runtime | Registration changed (server name, changed field names, values omitted) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.server.removed` | MCP Runtime | Registration tombstoned (server name, scope) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.connection.established` | MCP Runtime | Connection `ready` (connection ULID, negotiated protocol revision, server version string) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.connection.lost` | MCP Runtime | Connection dropped (connection ULID, failure class) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.connection.failed` | MCP Runtime | Establishment failed (connection ULID, error code, phase) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.surfaces.discovered` | MCP Runtime | Tools/resources/prompts enumerated (counts, discovery revision) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.request.failed` | MCP Runtime | A bridged request failed (connection ULID, method class, error code; no request content) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.log.received` | MCP Runtime | Server log captured (server name, level; message body goes to logs, not the event) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.exposure.changed` | MCP Runtime | Exposure approval state changed (change kind granted/suspended/revoked, affected tool count) | [Vol 6 ch 06](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) |
| `skill.registered` | Skill Engine | Skill accepted into the registry (name, version, scope, trust classification, content hash) | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skill.validation.failed` | Skill Engine | Manifest or content validation failed (finding count, first finding class) | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skill.activated` | Skill Engine | Skill made effective (name, version, target scope, content hash) | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skill.deactivated` | Skill Engine | Skill deactivated (name, version, cause class: user, requirement-lost, conflict) | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skill.composition.resolved` | Skill Engine | An activation set composed (set hash, member count, override count) | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skill.deprecated` | Skill Engine | Deprecation recorded (name, version, replacement if declared) | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `plugin.registered` | Plugin Runtime | Plugin registered (name, version, scope, declared surfaces summary) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.started` | Plugin Runtime | Process started (plugin ULID, containment level) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.handshake.completed` | Plugin Runtime | ARP handshake succeeded (plugin ULID, negotiated protocol version) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.handshake.failed` | Plugin Runtime | ARP handshake failed (plugin ULID, error code, offered/refused versions) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.stopped` | Plugin Runtime | Process stopped (plugin ULID, cause class: user, idle, shutdown) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.crashed` | Plugin Runtime | Process crashed (plugin ULID, exit class, invocations in flight) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.restarted` | Plugin Runtime | Supervised restart (plugin ULID, attempt number, cause class) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.disabled` | Plugin Runtime | Administrative disable (plugin ULID, actor: user/policy) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugin.removed` | Plugin Runtime | Plugin removed (plugin ULID, name) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `package.resolution.completed` | Package Manager | Resolution plan produced (plan hash, package count, sources consulted, cache ages, warnings) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `package.installation.started` | Package Manager | Installation began (plan hash, source kind, scope) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `package.installation.completed` | Package Manager | Installation succeeded (duration, files count, signature state, registered extension kinds) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `package.installation.failed` | Package Manager | Installation failed (failing state, error code, reason class) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `package.verification.failed` | Package Manager | Verification failed (failure class: checksum/signature/policy/content; E-PLUG-011) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `package.rollback.completed` | Package Manager | Upgrade rollback restored the prior version (restored version, failing state that triggered it) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `package.removal.completed` | Package Manager | Removal completed (cascade flag, deregistered extension kinds) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |

## Memory, context, indexing (Volume 7)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `memory.record.ingested` | Memory Manager | A new memory record persisted (FR-MEM-004) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.record.superseded` | Memory Manager | A supersession committed (FR-MEM-005) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.record.archived` | Memory Manager | Retention or consolidation archived a record (FR-MEM-006/FR-MEM-007) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.record.expired` | Memory Manager | Retention expired a record (FR-MEM-007) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.record.deleted` | Memory Manager | Deletion tombstoned a record (FR-MEM-008) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.completed` | Memory Manager | A retention pass finished (FR-MEM-007) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.consolidation.completed` | Memory Manager | A consolidation pass finished (FR-MEM-006) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.ingestion.refused` | Memory Manager | The redaction/validation gate refused content (FR-MEM-003; E-MEM-001/E-MEM-002) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.export.completed` | Memory Manager | An export stream closed complete (FR-MEM-009) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `context.assembly.completed` | Context Manager | A context snapshot persisted and the set was handed off (FR-CTX-001/FR-CTX-007) | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.budget.exceeded` | Context Manager | Assembly refused an infeasible mandatory set (FR-CTX-002; E-CTX-001) | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.conflict.detected` | Context Manager | Conflicting candidates were found (FR-CTX-004) | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.pin.changed` | Context Manager | A pin was added or removed (FR-CTX-005) | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.exclusion.changed` | Context Manager | An exclusion was added or removed (FR-CTX-005) | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `index.build.started` | Indexing Engine | A full build began (FR-IDX-001) | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.build.completed` | Indexing Engine | A build committed its generation (FR-IDX-001) | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.build.failed` | Indexing Engine | A build failed (E-IDX-001) | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.update.completed` | Indexing Engine | An incremental update committed (FR-IDX-004) | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.scope.invalidated` | Indexing Engine | A scope or whole index was invalidated (FR-IDX-004) | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.state.changed` | Indexing Engine | Any Index machine transition | [Vol 7 ch 05](../volume-07-memory-context-indexing/05-index-state-machine.md) |

## CLI and TUI (Volume 8)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `cli.command.started` | CLI | Command accepted (command path, invocation-mode record, flag names present) | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.command.completed` | CLI | Command finished with exit 0 (command path, duration ms) | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.command.failed` | CLI | Command finished nonzero (command path, error code, exit code, duration ms) | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.confirmation.resolved` | CLI | A destructive confirmation resolved (subject class, outcome, source prompt/yes-flag) | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.tui.launched` | CLI | TUI hand-off occurred (terminal columns/rows, color decision, workspace presence) | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.update.notified` | CLI | Post-command update notice rendered (installed version, available version, channel) | [Vol 8 ch 06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) |
| `tui.shell.started` | TUI | Shell started (terminal size, tier, mode, mouse active, workspace presence) | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.shell.exited` | TUI | Shell exited (duration ms, exit reason, overflow counters, restoration-failure flag) | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.screen.changed` | TUI | Active screen changed (from, to, trigger) | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.resize.applied` | TUI | Resize applied (columns, rows, layout class, coalesced count) | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.render.failed` | TUI | Render pipeline failure recorded (error code, screen, phase) | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.theme.resolved` | TUI | Theme resolution completed (mode, tier, per-decision source, error flag) | [Vol 8 ch 08](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) |
| `tui.palette.opened` | TUI | Command palette opened (no payload beyond envelope; recency context is local) | [Vol 8 ch 10](../volume-08-cli-and-tui/10-wireframes-platform.md) |
| `tui.action.invoked` | TUI | A registered action ran (action identifier, source, disposition executed/cancelled/refused) | [Vol 8 ch 10](../volume-08-cli-and-tui/10-wireframes-platform.md) |
| `tui.clipboard.copied` | TUI | A copy completed (content kind and byte count; never content, ADR-113) | [Vol 8 ch 11](../volume-08-cli-and-tui/11-interaction-patterns.md) |
| `tui.render_profile.changed` | TUI | Resolved render profile changed (color tier, glyph set, accessible flag, deciding signals) | [Vol 8 ch 12](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) |

## Security (Volume 9)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `permission.decision.recorded` | Permission Manager | Every evaluation that decides an action (outcome, deciding tier and references, resolution mode) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `permission.grant.created` | Permission Manager | A Permission row minted from a decision or rule | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `permission.grant.revoked` | Permission Manager | A standing grant revoked (`revoked_by` recorded) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `permission.grant.expired` | Permission Manager | A grant expired (TTL or session end) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `approval.requested` | Permission Manager | An Approval was raised (`requested`) | [Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md) |
| `approval.granted` | Permission Manager | Approval terminal `granted` | [Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md) |
| `approval.denied` | Permission Manager | Approval terminal `denied` | [Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md) |
| `approval.expired` | Permission Manager | Approval terminal `expired` (E-SEC-003) | [Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md) |
| `approval.cancelled` | Permission Manager | Approval terminal `cancelled` (E-SEC-004) | [Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md) |
| `sandbox.prepared` | Sandbox Engine | A sandbox handle prepared for a subject | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.policy.applied` | Sandbox Engine | The effective policy applied to an execution | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.violation.blocked` | Sandbox Engine | A policy violation blocked (E-SEC-005) | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.containment.degraded` | Sandbox Engine | Execution proceeded below requested isolation (E-SEC-006 context) | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.teardown.completed` | Sandbox Engine | Teardown finished (timings; failures raise E-SEC-007) | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `secret.stored` | Secret Store | Secret Store `Set` completed | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| `secret.accessed` | Secret Store | Secret Store `Get` occurred (all outcomes) | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| `secret.deleted` | Secret Store | Secret Store `Delete` completed | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| `secret.fallback.enabled` | Secret Store | Fallback store consent granted and the age-encrypted backend activated | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| `secret.orphan.swept` | Secret Store | Orphan sweep repaired store/row divergence | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| `audit.chain.verified` | Audit Log | Chain verification ran (any outcome) | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |
| `audit.chain.broken` | Audit Log | Chain integrity violation detected (E-SEC-013) | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |
| `security.incident.opened` | Audit Log | An incident record opened (paired `incident.recorded` audit action) | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |
| `security.incident.closed` | Audit Log | An incident record closed | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |

## Configuration and storage (Volume 10)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `config.file.loaded` | Configuration Manager | A layer or included file was read and parsed | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.resolution.completed` | Configuration Manager | A resolution froze successfully (layer digest: paths, profile names, counts — never values) | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.profile.selected` | Configuration Manager | A profile became active at a scope | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.change.detected` | Configuration Manager | A persisted layer changed or an override applied | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.override.applied` | Configuration Manager | A runtime override was accepted | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.override.rejected` | Configuration Manager | A runtime override was refused (E-CFG-014) | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.migration.applied` | Configuration Manager | A file was rewritten to the current schema | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.migration.failed` | Configuration Manager | A file rewrite failed and was rolled back (E-CFG-011) | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.deprecation.detected` | Configuration Manager | A deprecated key is present in a document (E-CFG-013 context) | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.validation.completed` | Configuration Manager | A validation pass finished with no error-severity findings | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.validation.failed` | Configuration Manager | A validation pass finished with at least one error-severity finding | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `config.secret.detected` | Configuration Manager | A definite secret detector matched a configuration value (E-CFG-012) | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| `storage.backup.created` | Persistence Layer | A pre-migration backup was written and verified | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.backup.pruned` | Persistence Layer | A backup beyond the retain count was removed | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.migration.applied` | Persistence Layer | A database migration chain completed | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.migration.failed` | Persistence Layer | A database migration stopped on failure (E-CFG-016) | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.integrity.failed` | Persistence Layer | An integrity or version check failed (E-CFG-017/E-CFG-015) | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.lock.denied` | Persistence Layer | A workspace write-lock wait expired (E-CFG-019) | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.retention.applied` | Persistence Layer | A retention pass completed | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |

## Logging, event bus, traces, telemetry (Volume 10)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `log.file.rotated` | Logging | Size rotation opened a successor log file | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `log.sink.degraded` | Logging | File sink failed; stderr degradation active (E-OBS-007) | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `event.publish.rejected` | Event Bus | Registry validation rejected an emission (E-OBS-001/E-OBS-002) | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| `event.subscriber.overflowed` | Event Bus | A subscriber's buffer applied its overflow policy (rate-limited signal) | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| `event.retention.pruned` | Event Bus | Retention pass removed event rows (counts per class) | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| `observability.sink.failed` | Observability | An observability store write failed (E-OBS-005) | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| `trace.completed` | Observability | A run's trace closed (span count, status) | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `trace.retention.pruned` | Observability | Whole traces removed by retention | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `metric.registration.violated` | Observability | Unregistered metric emission detected (E-OBS-006) | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `cost.rollup.compacted` | Observability | Daily cost compaction produced/replaced rollup rows | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `telemetry.consent.granted` | Telemetry | Interactive consent recorded | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.consent.revoked` | Telemetry | Consent tombstoned; export stopped; queue purged | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.consent.violated` | Telemetry | Enablement/ship attempt failed the consent gate (E-OBS-009) | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.export.enabled` | Telemetry | Export pipeline constructed (gate satisfied) | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.export.disabled` | Telemetry | Export pipeline torn down | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.batch.exported` | Telemetry | An OTLP batch was accepted by the endpoint | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.export.failed` | Telemetry | An export outage episode began (E-OBS-008) | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.data.deleted` | Telemetry | Local telemetry data purged (queue and/or identifier) | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |

## Git and hosting (Volume 11)

Payloads carry operation, repository digest, refs touched, and correlation IDs; diff and
patch bodies never appear in event payloads (Volume 10 redaction).

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `git.repository.discovered` | Git Engine | Repository discovery resolved (repository count per workspace) | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.commit.created` | Git Engine | A commit was created | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.branch.created` | Git Engine | A branch was created | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.branch.switched` | Git Engine | The checked-out branch changed | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.tag.created` | Git Engine | A tag was created | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.patch.applied` | Git Engine | A patch applied to the working tree | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.worktree.added` | Git Engine | A worktree was added | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.worktree.removed` | Git Engine | A worktree was removed | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.remote.fetched` | Git Engine | A fetch completed | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.remote.pulled` | Git Engine | A pull completed | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.remote.pushed` | Git Engine | A push completed | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.merge.completed` | Git Engine | A merge concluded | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.rebase.completed` | Git Engine | A rebase concluded | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.history.rewritten` | Git Engine | A history-modifying operation concluded | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.conflict.detected` | Git Engine | An operation stopped on conflicts (E-GIT-006) | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.conflict.resolved` | Git Engine | Conflict resolution recorded | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.safety_ref.created` | Git Engine | A pre-operation safety ref recorded (ADR-146) | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.operation.refused` | Git Engine | An operation refused (version gate, confirmation, permission, protected branch) | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.pull_request.opened` | Hosting integration layer | A change request was opened (host, repository, change-request identity) | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `git.pull_request.updated` | Hosting integration layer | A change request was updated (host, repository, change-request identity) | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |

The GH area mints no runtime events: development-process automation reports through CI check
results and audit-filed issues, not through the product event bus (Volume 11 register).

## Performance and reliability (Volume 12)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `perf.budget.exceeded` | Runtime (instrumentation) | Instrumented operation exceeded its declared budget by more than 2× (operation class, budget, measured value) | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |
| `perf.limit.enforced` | Runtime (instrumentation) | Any operational-limit enforcement (limit ID, class, effective value, source layer) | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |
| `perf.overload.shed` | Runtime (instrumentation) | Load-shedding action under sustained saturation (pool, policy, counts; FR-PERF-002) | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |
| `perf.degradation.entered` | Runtime (resource watchdog) | Degraded-mode entry (mode, cause, trigger sample; FR-PERF-003, ADR-162) | [Vol 12 ch 02](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) |
| `perf.degradation.exited` | Runtime (resource watchdog) | Degraded-mode exit (mode, duration) | [Vol 12 ch 02](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) |
| `perf.benchmark.completed` | Benchmark harness | Benchmark run finished with stored results (suite subset, result reference, fingerprint) | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |
| `perf.benchmark.regressed` | Benchmark harness | Relative regression beyond the warning or failure band (benchmark, baseline, measured value, band) | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |

## Testing and quality (Volume 13)

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `test.suite.completed` | Test harness | A classified suite finished in any tier (suite ID, type, tier, platform, counts, duration, report refs) | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |
| `test.gate.evaluated` | Gate evaluator | A gate computed a result (gate ID, subject, result, evidence references) | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |
| `test.qualification.completed` | Qualification pipeline | Qualification stage S6 finished for a candidate (candidate identity, decision, bundle digest, waiver count) | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |
| `test.flake.quarantined` | Test harness | A quarantine change landed (test identifier, issue link, quarantined date; ADR-177) | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |
| `test.fixture.failed` | Test harness | E-TEST-001 raised (fixture path, digests) | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| `test.replay.diverged` | Test harness | E-TEST-002 raised (cassette name, frame index) | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| `test.scenario.rejected` | Test harness | E-TEST-003 raised (script name, offending directive) | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| `test.hermeticity.violated` | Network sentinel | E-TEST-004 raised (dialer seam observation) | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| `test.evidence.rejected` | Qualification pipeline | E-TEST-005 raised (bundle schema or completeness failure) | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |
| `test.gate.errored` | Gate evaluator | E-TEST-006 raised (gate evaluation failure) | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |

## Distribution and updates (Volume 14)

Payloads are content-free: versions, channels, digests, counts, ULIDs.

| Event | Producer | Meaning / payload | Defined in |
|---|---|---|---|
| `update.check.completed` | Updater | Check finished with `up_to_date` or `update_available` (FR-REL-005) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.check.failed` | Updater | Check failed (E-REL-001/E-REL-002) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.artifact.downloaded` | Updater | Artifact finished staging (FR-REL-006) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.artifact.verified` | Updater | Verification passed for the staged set (FR-REL-002) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.verification.failed` | Updater | Verification failed (E-REL-004) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.version.applied` | Updater | New version active (`applied`) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.version.rolled_back` | Updater | Restore completed, automatic or manual (`rolled_back`) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.process.failed` | Updater | Update process terminated in `failed` | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.state.changed` | Updater | Any Update machine transition | [Vol 14 ch 05](../volume-14-distribution/05-state-machines.md) |
| `release.metadata.refreshed` | Updater | Local Release rows refreshed from the source | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `release.yank.detected` | Updater | Installed or targeted version learned to be yanked | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |

## Consolidation notes

- **Coverage.** 322 event names round-trip against the registers' "Events minted" sections:
  Volume 3: 8; Volume 4: 83 (54 agent-runtime + 29 workflow); Volume 5: 39 (17 provider +
  22 auth/connection); Volume 6: 48 (16 tool/terminal + 32 MCP/skill/plugin/package);
  Volume 7: 20; Volume 8: 16; Volume 9: 23; Volume 10: 37; Volume 11: 20; Volume 12: 7;
  Volume 13: 10; Volume 14: 11. No register row is omitted and no name appears here that a
  register does not list.
- **Volume 4 producers.** The merged Volume 4 register places the line "Producer for all
  rows: Workflow Engine" above its workflow table; it applies to the `workflow.*` family
  only. Producers for the session/workspace/run/turn/agent/plan/task/prompt families follow
  the chapter event tables (Runtime, Workspace Engine, Agent Engine, Planner, Execution
  Engine, Prompt Engine), which are reproduced here.
- **Volume 12 and Volume 13 producers.** Those registers name no producer component; the
  producer column above follows the defining chapters' emission rules (runtime
  instrumentation and watchdog for `perf.*`; harness, gate evaluator, and qualification
  pipeline for `test.*`).
- **Audit actions are not events.** The Volume 9 audited-action catalog (`permission.decided`,
  `git.mutation.performed`, `command.executed`, …) names Audit Record `action` attributes,
  not event names; per its register they are deliberately excluded from this catalog.
- **`provider.deprecation.announced`** is listed once; the Volume 5 register mentions it in
  its fragment B table while its defining section is chapter 09 (fragment B's file), which
  this catalog links.
