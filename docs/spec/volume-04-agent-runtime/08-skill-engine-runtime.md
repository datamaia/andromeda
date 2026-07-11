# 08 — Skill Engine Runtime Semantics

This chapter specifies how skills are **applied during runs and workflow steps** — the
execution-time semantics the agent runtime and Workflow Engine rely on. The skill *format* —
manifest, identity, versioning, packaging, signing, trust, installation, deprecation — is
Volume 6's (keystone FR-SKILL-001), as are the Skill Engine's registry, validation, and
composition-conflict rules (Volume 3 chapter 04 names Volume 6 the Skill Engine's behavior
owner). What belongs here, because it is run-time behavior of this volume's engines: when a
skill set is resolved, how it becomes part of prompts and requirements, what is recorded,
and what happens when a skill cannot apply. A Skill is procedural knowledge — versioned
prompt and requirement bundles — never executable code (code-bearing extensions are
plugins).

## Resolution: the run-start skill snapshot

The **active skill set** of a Run is resolved once, at run start, and snapshotted (ADR-052):

1. **Sources**, in additive order: skills selected by the Agent Profile of the run's root
   agent; skills declared by the workflow step that spawned the run (`skills` selector,
   chapter 06); skills explicitly activated for the session by the user.
2. Selectors resolve against the Skill Engine registry to concrete skill versions (highest
   non-deprecated enabled version per scope precedence — workspace, then global, then
   builtin — mirroring workflow resolution). Resolution failures follow the gating rules
   below.
3. The resolved set — skill names, versions, `content_hash` values, and per-skill
   application order — is recorded in the Run's `config_snapshot` before the first Turn
   (INV-RUN-04 discipline), making the skill contribution reproducible (SM-12).
4. The snapshot is immutable for the run's lifetime: skill installation, upgrade, or
   deactivation during a run affects only subsequently started runs — never a running one
   (ADR-052). Sub-agents inherit the root agent's snapshot; a delegation MUST NOT widen it.

`workflow.skillset.resolved` is emitted with the resolved set; a resolution that degraded
(below) additionally emits `workflow.skillset.degraded`.

## Application: prompts, tools, capabilities

- **Prompt contribution.** Skill prompt fragments enter model requests exclusively through
  the Prompt Engine's template registry slots (Volume 3 chapter 03) — never by string
  concatenation into user or context content. Fragments compose in the snapshot's recorded
  application order; composition conflicts are resolved by the Skill Engine's Volume 6
  rules before the snapshot is taken, so application at run time is deterministic: identical
  snapshot plus identical inputs renders identical prompt structures.
- **Tool requirements.** A skill's `required_tools` selectors are verified against the Tool
  Runtime registry at resolution. A skill never grants tools: its requirements widen
  nothing — the run's tool policy and the permission model still mediate every invocation
  (Volume 9). A skill whose required tools are outside the run's tool policy is treated as
  requirement-unsatisfied.
- **Capability requirements.** `required_capabilities` (Volume 5 enum values) are checked
  against the declared capabilities of the run's resolved model at snapshot time — declared
  capabilities, never model-name assumptions (Principle 2). Absent capabilities are never
  silently simulated.
- **Workflow interaction.** For workflow-spawned runs, the step's `skills` selectors are
  requirements of the step: gating outcomes feed the step's entry criteria (chapter 06).
  Stage catalog agents of `spec-driven-dev` MAY carry skill selectors in their Agent
  Profiles; the workflow definition adds none by default.

## Gating and degradation

Every skill selector carries an implicit or explicit `required` flag (workflow step
selectors default to `required = true`; profile and session selectors default to
`required = false`):

| Condition | Required skill | Optional skill |
|---|---|---|
| Selector matches no installed/enabled version | E-WF-013; the run (or step) fails to start | Skill omitted; `workflow.skillset.degraded` emitted |
| `required_tools` unsatisfied or outside tool policy | E-WF-013 | Skill omitted; degradation event |
| `required_capabilities` absent from the resolved model | E-WF-013 | Skill omitted; degradation event |
| Trust policy refuses the skill's trust level (Volume 9) | E-WF-013 with the policy reference | Skill omitted; degradation event |
| Content hash mismatch at load (integrity) | E-WF-013 — never applied | Never applied; degradation event |

Degradation is always explicit: the omitted skill, the reason, and the decision are recorded
in the snapshot and emitted — a user can always answer "which skills actually shaped this
run, and which were dropped, and why" (Principle 7). Mid-run failures cannot occur by
construction: after snapshot, application is pure rendering of already-loaded, already-hashed
content; a fragment that fails to render is an engine defect surfaced through the run's
error family, not a skill-gating event.

## Requirements

### FR-WF-008 — Skill application in runs and workflows

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Design
- Owner: Workflow Engine (Volume 4) with Agent Engine
- Affected components: Agent Engine, Workflow Engine, Skill Engine, Prompt Engine, Tool Runtime
- Dependencies: FR-SKILL-001 (format keystone, Volume 6); FR-WF-003; ADR-052; Volume 2 Skill entity
- Related risks: RISK-WF-001

#### Description

The runtime MUST resolve each Run's active skill set once at run start from the three
declared sources, snapshot it (names, versions, content hashes, application order) in
`config_snapshot` before the first Turn, apply skill prompt fragments only through Prompt
Engine registry slots in snapshot order, verify tool and capability requirements at
resolution time against declared facts, and enforce the gating table of this chapter —
failing required skills with E-WF-013 and explicitly recording every optional-skill
omission. The snapshot MUST be immutable for the run's lifetime and inherited unwidened by
sub-agents.

#### Motivation

Skills alter agent behavior; unrecorded or mutable skill influence would break run
reproducibility (SM-12) and make "why did the agent do that" unanswerable (PRD-006).

#### Actors

Agent Engine (run start); Workflow Engine (step-declared skills); Skill Engine (registry);
users activating session skills.

#### Preconditions

Skill registry loaded; run's model resolved (for capability checks).

#### Main flow

1. Collect selectors from profile, workflow step, and session activation.
2. Resolve, verify requirements and trust, take the snapshot, persist it, emit
   `workflow.skillset.resolved`.
3. Render prompts per turn using the snapshot's fragments and order.

#### Alternative flows

- Optional skill fails a check: omitted with `workflow.skillset.degraded`; the run proceeds.

#### Edge cases

- Two sources select different versions of one skill: the highest-precedence source wins
  (session over step over profile); the decision is recorded in the snapshot.
- A skill is uninstalled mid-run: the run is unaffected (content was loaded and hashed at
  snapshot); the next run re-resolves.
- Empty skill set: valid; the snapshot records an empty list.

#### Inputs

Skill selectors; registry contents; model capability declarations; tool policy; trust
policy.

#### Outputs

Persisted skill snapshot; prompt contributions; resolution/degradation events.

#### States

None of its own — outcomes land in run records; skills are stateless catalog entities
(Volume 2).

#### Errors

E-WF-013 for required-skill failures; Volume 6's E-SKILL family governs registry-side
validation errors surfaced through resolution.

#### Constraints

No mid-run snapshot mutation; no fragment entry outside Prompt Engine slots; no widening of
tool policy or permissions by any skill.

#### Security

Trust-gated application per Volume 9 policy; content-hash verification before any fragment
is rendered; skill content passes redaction classification like all prompt material.

#### Observability

Snapshot recorded per run (SM-12); `workflow.skillset.resolved` / `workflow.skillset.degraded`
events; per-skill application order inspectable.

#### Performance

Resolution and hashing occur once per run start; per-turn cost is template rendering only
(Prompt Engine determinism rules).

#### Compatibility

Skill format versioning and compatibility are Volume 6's (SM-20 surface); this requirement
consumes resolved versions only.

#### Acceptance criteria

- Given a run with profile and session skills, when started, then `config_snapshot` lists
  every applied skill with name, version, hash, and order, before the first Turn exists.
- Given a required skill whose `required_capabilities` include a value the resolved model
  does not declare, when the run starts, then it fails with E-WF-013 naming the capability
  and no Turn executes.
- Negative case: given an optional skill with a content-hash mismatch, when resolved, then
  it is never applied, the omission and reason are recorded, and
  `workflow.skillset.degraded` is emitted.
- Permission case: given a skill requiring a tool outside the run's tool policy, when
  resolved as optional, then it is omitted — the tool policy is not widened.
- Observability case: given any completed run, when replayed (SM-12), then the identical
  skill snapshot reproduces identical prompt structures.

#### Verification method

Snapshot determinism and replay tests (SM-12 suite); gating matrix tests over the five
conditions × required/optional; mid-run mutation tests (install/upgrade during a run);
composition-order golden tests with the Prompt Engine.

#### Traceability

PRD-006, PRD-007; FR-SKILL-001; SM-12; ADR-052; Principle 2 and Principle 7 (Volume 1).

## Error codes

### E-WF-013 — Skill application failed

- Category: Environment
- Severity: Error
- User message: "Required skill '<name>' cannot be applied: <reason>."
- Technical message: skill name and version constraint, failing condition (missing version, unsatisfied tools, absent capability, trust refusal, hash mismatch), source of the requirement (profile, workflow step, session)
- Cause: a `required` skill selector failing any gating-table condition at resolution
- Safe-to-log data: skill name, version constraint, condition class, requirement source
- Recoverability: recoverable after installing/enabling the skill, adjusting the tool policy or model selection, or removing the requirement
- Retry policy: none automatic; the next run re-resolves
- Recommended action: named per condition in the diagnostic
- Exit-code mapping: 1
- HTTP mapping: not applicable
- Telemetry event: `workflow.skillset.degraded`
- Security implications: hash-mismatched or trust-refused skill content is never rendered into any prompt — refusal precedes application
