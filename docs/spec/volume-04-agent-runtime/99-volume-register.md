# volume-04-agent-runtime — Volume Register

Merged from per-agent register fragments at the Phase B gate.

## Requirements index

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-AGT-001 | Agent loop | MVP | Loop tests over provider doubles; replay divergence (SM-12); crash injection (SM-11); UC-01 E2E; audit-chain tests (SM-13) |
| FR-AGT-002 | Turn handling and the message part vocabulary | MVP | Pipeline unit tests; message-part property tests; streaming/non-streaming contract tests; record validators |
| FR-AGT-003 | Run interruption, pause, and resume | MVP | SM-11 crash-injection at randomized points; pause/resume integration; double-resume races; approval-expiry fixtures |
| FR-AGT-004 | Sub-agent delegation | Beta | Delegation integration tests; permission-narrowing enforcement; audit-chain resolution; depth/budget property tests |
| FR-AGT-005 | Run budget enforcement | MVP | Budget property tests; late-usage fixtures; scripted-spend integration; accounting consistency validators |
| FR-AGT-006 | Workspace lifecycle over WorkspacePort | MVP | Filesystem fixtures (markers, nesting, permissions); concurrent-open tests; registry repair tests; Tier 1 matrix |
| FR-AGT-007 | Plan production | MVP | Golden decomposition tests; validation property tests; structured/text parity; failure-double integration |
| FR-AGT-008 | Plan revision | MVP | Revision protocol integration; transactional supersession (crash between steps); lineage validators; bound-breach tests |
| FR-AGT-009 | Plan inspection and approval interplay | MVP | Gate enforcement tests; mode × session-kind matrix; audit-chain validators; approval expiry fixtures |
| FR-AGT-010 | Task scheduling and dispatch | MVP | Deterministic DAG execution; dispatch-refusal tests; cancellation storms; audit-chain validators; capacity matrix |
| FR-AGT-011 | Task retry policy | MVP | Fault-injection matrices (retryable × side-effect); backoff property tests; gated-confirmation integration; exhaustion fixtures |
| FR-AGT-012 | Cancellation, skip, and error propagation | MVP | Cancellation storms; wedged-child fixtures; cancel-vs-complete races; reason validators; leak gates |
| FR-AGT-013 | Versioned prompt templates and registry | MVP | Registry unit tests (precedence, shadowing, rejection); trust-gate integration; snapshot stability; golden diagnostics |
| FR-AGT-014 | Deterministic rendering with provenance | MVP | Cross-platform determinism property tests; slot/schema failure matrices; provenance round-trips in replay suite |
| FR-AGT-015 | Canonical machine enforcement | Core | Property-based machine tests; illegal-write attempts; replay validation fixtures; transition/event parity validators |
| NFR-AGT-001 | State transition legality under load | MVP | Property/race/crash suites with record validation; zero-violation gate per mainline commit |
| NFR-AGT-002 | Resume fidelity | MVP | SM-11-method crash-injection plus scripted resume; record diffing; confirmation assertions |
| NFR-AGT-003 | Prompt render determinism | MVP | ≥ 100 repeated renders per template per Tier 1 platform; golden render audit per release |
| RISK-AGT-001 | Non-terminating or runaway agent loop | — | Mitigations tracked via limit/budget enforcement tests and soak runs |
| RISK-AGT-002 | Plan–execution divergence | — | Audit-chain resolution (SM-13); dispatch-refusal counters |
| RISK-AGT-003 | Resumption ambiguity across the crash boundary | — | SM-11 campaigns; pre/post record diffing; confirmation assertions |
| RISK-AGT-004 | Prompt template drift and override injection | — | Trust-gate tests; golden render diffs; override event audits |

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-WF-001 | Specification-driven development workflow | Beta | Workflow E2E suite over scripted doubles; gate-profile matrix; loop-budget and denial fixtures; audit-chain validation |
| FR-WF-002 | Workflow definition format and validation | Beta | Golden valid/invalid fixture suite; TOML→JSON round-trips; immutability and idempotence tests; trust-gating integration tests |
| FR-WF-003 | Workflow execution semantics | Beta | Deterministic engine tests over doubles; parallelism/serialization property tests; step-boundary crash injection; headless parity tests |
| FR-WF-004 | Workflow approval gates | Beta | Gate outcome matrix tests; headless prompt-absence tests; audit-chain checks; expiry timing tests |
| FR-WF-005 | Workflow Run state machine conformance | Beta | Transition-table property tests; event/transition matching audits; revision race tests |
| FR-WF-006 | Workflow interruption, recovery, and resume | Beta | Crash injection at randomized boundaries and mid-step; resume Approval fixtures; integrity-corruption fixtures; cross-machine resume tests |
| FR-WF-007 | Workflow cancellation and rollback | Beta | Cancellation storms; compensation-order property tests; failed-compensation fixtures; restore-point round-trips over GitPort doubles |
| FR-WF-008 | Skill application in runs and workflows | Beta | Snapshot determinism and replay tests; gating matrix tests; mid-run mutation tests; composition-order golden tests |
| FR-WF-009 | Workflow scheduling and supervision | Beta | Scheduler-integration suite with cancellation storms and saturation fixtures; leak gates; waiting-state footprint tests; panic injection |
| FR-WF-010 | Durable timers and timeout enforcement | Beta | Timer property tests under injected clocks; kill-restart fixtures across deadlines; grant-vs-expiry races; pause/resume budget golden tests |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-WF-001 | Workflow format public-contract stability | Beta | Contract-diff of the definition JSON Schema per release; Volume 14 release audit |
| NFR-WF-002 | Workflow resume fidelity | Beta | Crash-injection suite with automated post-resume `step_states` audit |
| NFR-WF-003 | Workflow orchestration overhead | Beta | Instrumented no-op workflow benchmark; parked-run heap accounting on reference hardware |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-WF-001 | Process overhead deters SDD adoption | Medium | Open |
| RISK-WF-002 | Irreversible external side effects defeat rollback | High | Open |
| RISK-WF-003 | Approval-gate fatigue erodes the safety model | High | Open |

## ADRs minted

| ADR | Title |
|---|---|
| ADR-040 | One agent loop for interactive and autonomous operation |
| ADR-041 | Plan representation: versioned task DAG with supersession |
| ADR-042 | Task retry policy: bounded backoff with side-effect gating |
| ADR-043 | Conservative resumption of interrupted work |
| ADR-044 | Prompt template format: restricted Go text/template with a versioned slot registry |
| ADR-045 | Sub-agent delegation: bounded depth, permission narrowing, budget slices |

ADR numbers 046–047 of this agent's block are unused; per Volume 0 chapter 03 they remain
permanent gaps.

Volume 4 block allocation is ADR-040–054; this fragment used 048–054 (040–047 are fragment
A's).

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-048](../annexes/adr/ADR-048.md) | Workflow definition format | Accepted | TOML-authored definitions compiled deterministically to a versioned canonical JSON document validated per ADR-024 |
| [ADR-049](../annexes/adr/ADR-049.md) | SDD as a fixed fourteen-stage pipeline | Accepted | Fixed stage order, exactly four bounded feedback loops into implementation, tailoring only via gate profiles and definition versions |
| [ADR-050](../annexes/adr/ADR-050.md) | Step-boundary resume with approval-gated re-execution | Accepted | Resume only at persisted step boundaries; interrupted side-effecting steps re-dispatch only through an explicit resume Approval |
| [ADR-051](../annexes/adr/ADR-051.md) | Durable workflow timers | Accepted | Deadlines persisted as UTC instants with the arming transition; ≤ 30 s wall-clock sweep bounds firing latency; guarded idempotent firing |
| [ADR-052](../annexes/adr/ADR-052.md) | Run-start skill snapshot | Accepted | Skill sets resolve once at run start, recorded in `config_snapshot`, immutable per run, inherited unwidened by sub-agents |
| [ADR-053](../annexes/adr/ADR-053.md) | Workflow orchestration as supervised scheduler tasks | Accepted | One scheduler group per workflow run; short control tasks; waiting states hold no pool slots or goroutines |
| [ADR-054](../annexes/adr/ADR-054.md) | Compensation-based rollback with git restore points | Accepted | Declared per-step compensations executed in reverse order, halting honestly on failure; SDD anchors on a git restore point |

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-AGT-001 | Not a workspace | 1 |
| E-AGT-002 | Workspace exclusivity conflict | 1 |
| E-AGT-003 | Session or run not resumable | 1 (9 on integrity cause) |
| E-AGT-004 | Run budget exhausted | 8 |
| E-AGT-005 | Iteration limit reached | 8 |
| E-AGT-006 | Agent profile resolution failure | 3 |
| E-AGT-007 | Plan validation failed | 1 |
| E-AGT-008 | Task dependencies unsatisfiable | 1 |
| E-AGT-009 | Prompt template resolution failure | 3 |
| E-AGT-010 | Prompt render failure | 1 (3 on configuration cause) |
| E-AGT-011 | Illegal state transition | 1 (9 via replay/integrity validation) |
| E-WF-001 | Workflow definition invalid | 3 |
| E-WF-002 | Workflow definition version unsupported | 3 |
| E-WF-003 | Workflow not found | 2 |
| E-WF-004 | Workflow inputs invalid | 2 |
| E-WF-005 | Workflow requirements unsatisfied | 1 |
| E-WF-006 | Workflow gate denied | 5 |
| E-WF-007 | Workflow gate expired | 8 |
| E-WF-008 | Workflow step failed | 1 |
| E-WF-009 | Workflow step timed out | 8 |
| E-WF-010 | Workflow run cancelled | 8 |
| E-WF-011 | Workflow compensation failed | 1 |
| E-WF-012 | Workflow run state integrity failure | 9 |
| E-WF-013 | Skill application failed | 1 |

## Events minted

Session family: `session.created`, `session.activated`, `session.suspended`,
`session.resumed`, `session.ended`, `session.failed`.
Workspace family: `workspace.discovered`, `workspace.opened`, `workspace.closed`.
Run family: `run.created`, `run.started`, `run.planned`, `run.revision.started`,
`run.approval.requested`, `run.paused`, `run.resumed`, `run.interrupted`, `run.completed`,
`run.failed`, `run.cancelled`, `run.budget.exhausted`.
Turn family: `turn.started`, `turn.completed`, `turn.failed`, `turn.interrupted`.
Agent family: `agent.instantiated`, `agent.state.changed`, `agent.delegation.started`,
`agent.delegation.completed`.
Plan family: `plan.drafted`, `plan.proposed`, `plan.approved`, `plan.execution.started`,
`plan.revision.started`, `plan.revision.abandoned`, `plan.superseded`, `plan.completed`,
`plan.abandoned`.
Task family: `task.created`, `task.scheduled`, `task.started`, `task.blocked`,
`task.unblocked`, `task.approval.requested`, `task.resumed`, `task.retried`,
`task.interrupted`, `task.completed`, `task.failed`, `task.cancelled`, `task.skipped`.
Prompt family: `prompt.template.registered`, `prompt.template.overridden`,
`prompt.rendered`.

Names per the Volume 0 grammar; envelope, delivery, persistence, and retention semantics per
Volume 10. Producer for all rows: Workflow Engine.

| Event | Meaning |
|---|---|
| `workflow.definition.registered` | A definition version passed validation and was registered |
| `workflow.definition.rejected` | A definition source failed validation (E-WF-001/002) or trust policy |
| `workflow.run.started` | Transition T1: instantiated run began executing |
| `workflow.run.paused` | Explicit user pause took effect at a step boundary |
| `workflow.run.resumed` | Resume from `paused` or `interrupted` |
| `workflow.run.interrupted` | Crash/shutdown marking (T14) |
| `workflow.run.completed` | Terminal success with outputs recorded |
| `workflow.run.failed` | Terminal failure with persisted error |
| `workflow.run.cancelled` | Terminal cancellation with recorded reason |
| `workflow.step.entered` | A step began (dispatch or gate raise) |
| `workflow.step.completed` | A step reached `completed` |
| `workflow.step.failed` | A step reached `failed` (E-WF-008) |
| `workflow.step.retried` | A failed attempt re-dispatched under its retry policy |
| `workflow.step.skipped` | A step was deliberately not executed (routing/condition) |
| `workflow.step.timed_out` | A step timer fired (E-WF-009) |
| `workflow.step.deferred` | Work deferred on scheduler backpressure (E-ARCH-005 handled) |
| `workflow.gate.requested` | An Approval with subject kind `workflow_gate` was raised |
| `workflow.gate.granted` | Gate Approval terminal `granted` |
| `workflow.gate.denied` | Gate Approval terminal `denied` (E-WF-006) |
| `workflow.gate.expired` | Gate Approval terminal `expired` (E-WF-007) |
| `workflow.artifact.recorded` | A declared step artifact was recorded |
| `workflow.rollback.started` | Compensation sequence began |
| `workflow.rollback.completed` | All applicable compensations completed |
| `workflow.rollback.failed` | Rollback halted on a failed compensation (E-WF-011) |
| `workflow.skillset.resolved` | A run's skill snapshot was resolved and recorded |
| `workflow.skillset.degraded` | An optional skill was omitted, or a required skill failed (E-WF-013) |
| `workflow.timer.armed` | A durable deadline was persisted |
| `workflow.timer.fired` | A deadline fired and its transition applied |
| `workflow.timer.restored` | Timers re-armed from persisted deadlines at recovery/resume |

## Config keys minted

| Key | Chapter |
|---|---|
| `agent.default_profile` | 01 |
| `agent.session.idle_suspend_after` | 01 |
| `agent.loop.max_iterations` | 01 |
| `agent.loop.turn_timeout` | 01 |
| `agent.loop.max_repair_attempts` | 01 |
| `agent.loop.max_subagent_depth` | 01 |
| `agent.loop.delegation_enabled` | 01 |
| `agent.planner.approval_mode` | 02 |
| `agent.planner.max_attempts` | 02 |
| `agent.planner.attempt_timeout` | 02 |
| `agent.planner.max_tasks_per_plan` | 02 |
| `agent.planner.max_revisions` | 02 |
| `agent.execution.max_parallel_tasks` | 03 |
| `agent.execution.task_timeout` | 03 |
| `agent.execution.retry.max_attempts` | 03 |
| `agent.execution.retry.base_delay` | 03 |
| `agent.execution.retry.max_delay` | 03 |
| `agent.execution.retry.multiplier` | 03 |
| `agent.prompts.allow_workspace_overrides` | 04 |
| `agent.prompts.override_dirs` | 04 |
| `agent.prompts.max_render_bytes` | 04 |

Key content of the `[workflows]` table (chapter 06); schema, precedence, and validation are
Volume 10's.

| Key | Type | Default |
|---|---|---|
| `workflows.paths` | array of paths | `[]` |
| `workflows.default_step_timeout` | duration | `"30m"` |
| `workflows.default_gate_expiry` | duration | `"24h"` |
| `workflows.default_max_attempts` | integer | `1` |
| `workflows.max_parallel_steps` | integer | `4` |
| `workflows.max_run_duration` | duration | `"168h"` |
| `workflows.artifacts_dir` | path | `".andromeda/artifacts"` |
| `workflows.sdd.gate_profile` | enum (`strict` \| `standard` \| `minimal`) | `"standard"` |

## Glossary additions

| Term | Meaning |
|---|---|
| Agent loop | The single plan–act–observe iteration cycle the Agent Engine drives for every run (FR-AGT-001) |
| Iteration boundary | The loop checkpoint between turns where budgets, limits, and revision triggers are evaluated |
| Direct-execution plan | A one-task Plan produced for goals classified single-step, preserving the every-run-has-a-plan invariant |
| Render provenance | The `(namespace, name, version, parameter_hash)` tuple recorded per rendered prompt on its consuming turn |
| Budget slice | The portion of run budgets explicitly allocated to a delegated sub-agent |
| Ready computation | The Execution Engine's derivation of dispatchable tasks from dependency and plan state |
| Term | One-line meaning |
|---|---|
| Workflow step | One declared unit of a workflow definition: kind `agent`, `gate`, or `transform`, with dependencies, criteria, timeout, retry, and routing. |
| Stage | A named step of the builtin `spec-driven-dev` workflow (fourteen stages, ADR-049). |
| Gate profile | The configured selection (`strict`/`standard`/`minimal`) of which SDD stage gates require human approval. |
| Approval gate | A `gate` step raising an Approval with subject kind `workflow_gate`; passed only by a terminal `granted` decision (INV-WFR-04). |
| Effects classification | A step's declared side-effect class (`none`/`workspace`/`external`) driving resume and rollback rules (ADR-050). |
| Resume Approval | The explicit human decision required before an interrupted side-effecting step re-dispatches. |
| Compensation | A step's declared undo action, executed during rollback in reverse completion order (ADR-054). |
| Restore point | The git branch/commit anchor SDD records before its first side-effecting stage, used by implementation compensation. |
| Skill snapshot | The per-run record of resolved skill names, versions, hashes, and application order, immutable for the run (ADR-052). |
| Durable timer | A workflow deadline persisted as an absolute UTC instant, enforced by guarded idempotent firing with a ≤ 30 s sweep bound (ADR-051). |

## Assumptions

| Statement | Kind | Validation path |
|---|---|---|
| Models declaring `structured_outputs` produce plan candidates that pass deterministic validation within `agent.planner.max_attempts` often enough for planning to be practical | Product hypothesis | SM-04-class conformance runs against pinned local and cloud models; planning-failure rate metrics from MVP onward |
| Side-effect attribution records (Volume 2/6) are complete enough for the ADR-042/ADR-043 gating to be sound in practice | Technical assumption | SM-13 orphan-side-effect audits; fault-injection with effect-recording tools |
| Default limits chosen in this fragment (iterations 50, attempts 3, tasks/plan 30, revisions 10) are workable for the UC-01-class workloads | Product hypothesis | MVP field metrics on limit-triggered cancellations; revisited at Beta gate |

Local list per Volume 0, chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Product hypothesis | The fourteen-stage granularity with four bounded loops matches real engineering iteration patterns | Beta adoption metrics (completion/abandonment per stage); Volume 15 feedback channels | Re-version `spec-driven-dev` per the ADR-049 reversal plan; the format supports lighter user-authored processes today |
| Technical assumption | Step-boundary persistence granularity yields acceptable lost-work windows for typical stage durations | Crash-injection lost-work measurements; Beta telemetry on stage durations | Decompose long stages into more steps; ADR-050 review conditions govern finer-grained resume |
| Technical assumption | A closed, machine-checkable predicate set suffices for most entry/exit criteria, with free-text criteria carried to gate approvers | Workflow-author feedback during Beta | Extend the predicate set via a `schema_version` increment under NFR-WF-001 rules |
| Technical assumption | The 30-second sweep bound is imperceptible for minutes-scale workflow timeouts and 24-hour gate expiries | Volume 12 firing-latency metrics | Tighten the sweep interval (internal change; the persisted-deadline contract is unaffected) |

## Open questions

None from this fragment. Fragment A introduces no PENDING VALIDATION markers: its scope is
internal design with no dependence on unconfirmed external facts.
| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V4B-OQ-1 | Behavior of Go monotonic timers across system sleep on Tier 1 platforms (whether sleep time counts toward timer expiry) — PENDING VALIDATION; affects only the in-memory optimization path's latency, since the ≤ 30 s wall-clock sweep bounds worst-case firing regardless | Chapter 09 (durable timers); ADR-051 | No — the wall-clock sweep is the contractual bound | Beta implementation spike measuring timer behavior across sleep on macOS and Linux reference systems; record findings against ADR-051 review conditions | Open |

## Cross-volume references

- Volume 2 chapters 03, 09, 10: entity shapes, frozen state names, write discipline — cited
  throughout; INV-* invariants enforced by chapters 01–05.
- Volume 3 chapters 02, 03, 08: WorkspacePort/SessionStorePort contracts elaborated here;
  FR-ARCH-004/006/009/010 consumed; E-ARCH-005 referenced in dispatch semantics.
- Volume 5: ProviderPort behavior, capability enum (`streaming`, `tool_calling`,
  `structured_outputs`, `reasoning`), provider retry/timeout policy, cost reporting.
- Volume 6: Tool Runtime mediation, Tool Invocation machine, tool timeout/teardown budgets,
  attribution obligations of the tool contract.
- Volume 7: Context Manager assembly (keystone FR-CTX-001), memory ingestion rules, token
  estimation fallback.
- Volume 8: presentation of plans, runs, approvals; command grammar for
  sessions/runs/profiles.
- Volume 9: permission model (keystone FR-SEC-100), Approval machine and expiry, trust gates
  for template overrides, redaction rules.
- Volume 10: configuration schema/precedence (keystone FR-CFG-001), event envelope (keystone
  FR-OBS-001), storage durability, replay mode, exclusivity rules.
- Volume 12: budgets for loop latency, dispatch capacity, cancellation quiescence, recovery
  time.
- Volume 13: test harnesses cited in every Verification method.

- **Volume 2** — Workflow and Workflow Run entities and invariants (INV-WFD-01..04,
  INV-WFR-01..04); frozen Workflow Run states (chapter 09); write discipline (chapter 10).
- **Volume 3** — SchedulerPort structural contract and supervision tree (chapter 08,
  FR-ARCH-006); Workflow Engine and Skill Engine component boundaries (chapter 04);
  cancellation contract (FR-ARCH-004); recovery procedure (FR-ARCH-009).
- **Volume 5** — capability enum values consumed by `required_capabilities` checks.
- **Volume 6** — skill format and Skill Engine registry/composition rules (FR-SKILL-001);
  tool contract and Tool Runtime mediation (FR-TOOL-001 by name); E-SKILL family for
  registry-side errors.
- **Volume 9** — permission model and enum (FR-SEC-100); Approval full machine; trust
  policy for skills and package-delivered definitions; audit and redaction semantics.
- **Volume 10** — event envelope and delivery semantics; configuration schema and
  precedence for `[workflows]`; SessionStorePort storage mechanics.
- **Volume 11** — Git Engine semantics behind SDD restore points; mandatory human review of
  pull requests (the SDD review stage precedes, never replaces, it).
- **Volume 12** — pool sizes, saturation budgets, reference hardware, and the latency
  budgets NFR-WF-003 binds to.
- **Volume 13** — crash-injection, property, and conformance suite definitions used by
  every verification method above.
- **Volume 14** — release publication (explicitly out of the release-preparation stage's
  scope); release audit for NFR-WF-001.
- **Volume 4, fragment A** — agent loop (FR-AGT-001), Run/Plan/Task machines, Prompt
  Engine rendering, Execution Engine retry mechanics consumed by workflow steps.
