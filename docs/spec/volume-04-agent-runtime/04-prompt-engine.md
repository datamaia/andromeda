# 04 — Prompt Engine

The Prompt Engine renders versioned prompt templates into the prompt structures the Agent
Engine, Planner, and Workflow Engine send through ProviderPort. It owns the template format,
the registry, and the rendering semantics; it does not select context (Context Manager),
call providers, or store skill packages (Skill Engine loads, this engine composes). Its two
load-bearing properties are versioning and determinism: given the same template version, the
same parameters, and the same material, rendering MUST produce byte-identical output — this
is what lets run records reference prompts instead of storing them, and still support exact
reconstruction (SM-12).

## Template model

| Property | Specification |
|---|---|
| Identity | `(namespace, name, version)`; version is SemVer per template (ADR-044) |
| Namespaces | `builtin` (shipped in the binary), `workspace` and `global` (overrides), `skill/<skill-name>` (contributed by skills via the Skill Engine) |
| Format | Go `text/template` syntax restricted to a declared function allowlist; no I/O-capable, time-reading, or environment-reading functions (ADR-044) |
| Slots | A template declares named slots (typed insertion points: `system_material`, `context_material`, `task_statement`, `skill_fragments`, and template-declared extras); contributions compose through slots, never by string concatenation |
| Parameters | A template declares its parameter schema (names, types, defaults); rendering validates supplied parameters against it |
| Provenance | Every rendered output records `(namespace, name, version, parameter_hash)` — the render provenance persisted on the turn |

Override precedence at resolution: `workspace` > `global` > `builtin` for the same
`(name, major version)`; an override MUST declare which builtin version range it overrides,
and overrides are trust-gated: loading a workspace override requires the workspace trust
decision Volume 9 defines for code-adjacent assets. Skill-contributed templates live in
their own namespace and never silently replace builtin templates — profiles opt in by
referencing them.

Version resolution mirrors profile resolution: an explicit version pins exactly; a name
without a version resolves to the highest non-deprecated version at resolution time and the
resolved version is snapshotted into the run's `config_snapshot` — mid-run template changes
never affect a running run.

## Rendering pipeline

1. **Resolve** the template reference against the registry (E-AGT-009 on failure).
2. **Validate** supplied parameters against the template's parameter schema.
3. **Compose** slot contents: caller-supplied material (Context Manager output, task
   statements), profile parameters, and skill fragments for declared slots. Unfilled
   optional slots render empty; unfilled required slots fail (E-AGT-010).
4. **Render** deterministically: no clock, environment, filesystem, or randomness access is
   available to templates (the allowlist excludes them by construction).
5. **Classify** the output through redaction classification (Volume 9 rules) before it
   leaves the engine.
6. **Record** render provenance for the consuming turn.

Rendered output size is bounded by the caller's token budget (Context Manager); the engine
enforces a hard byte ceiling per render as a defense in depth (`agent.prompts.max_render_bytes`).

## Requirements

### FR-AGT-013 — Versioned prompt templates and registry

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Prompt Engine (Volume 4)
- Affected components: Prompt Engine, Skill Engine, Configuration Manager, Agent Engine
- Dependencies: ADR-044; FR-AGT-001; ConfigPort contract (Volume 3)
- Related risks: RISK-AGT-004

#### Description

The Prompt Engine MUST maintain a registry of prompt templates keyed by
`(namespace, name, version)` with SemVer versions, the namespace set of this chapter, the
override precedence `workspace` > `global` > `builtin`, trust-gated loading of override
sources per Volume 9, and skill contributions confined to `skill/<name>` namespaces.
Registration validates template syntax, the function allowlist, slot declarations, and the
parameter schema at load time; invalid templates are rejected at registration, never at
render time mid-run. Version resolution snapshots into the run's `config_snapshot`.

#### Motivation

Prompts steer everything the product does; unversioned, mutable prompts would make run
records unreproducible (SM-12) and behavior changes undiagnosable (PRD-006).

#### Actors

Prompt Engine; Skill Engine contributing templates; users installing overrides;
Configuration Manager supplying override paths.

#### Preconditions

Builtin templates embedded in the binary; override directories resolvable via ConfigPort.

#### Main flow

1. At startup, builtin templates register; override directories load through trust gates;
   skills contribute through the Skill Engine.
2. Registration validates syntax, allowlist, slots, and parameter schemas.
3. Callers resolve references; resolution results snapshot per run.
4. Configuration watch reloads overrides at defined reconfiguration points — never mid-run.

#### Alternative flows

- Trust denied for a workspace override: the override is not loaded, the builtin version
  serves, and the refusal is recorded and surfaced.
- Deprecated template version: resolvable only by explicit pin, mirroring profile rules.

#### Edge cases

- Two overrides claim the same `(name, major)`: precedence decides; the shadowed one is
  reported in diagnostics.
- An override declares an incompatible overridden range (builtin moved majors): the override
  is rejected at load with the range mismatch named.
- Registry reload while a run is active: running runs keep their snapshot; only new runs see
  the change.

#### Inputs

Template sources (binary, directories, skills); trust decisions; configuration.

#### Outputs

The validated registry; resolution results; `prompt.template.registered` /
`prompt.template.overridden` events; rejection diagnostics.

#### States

Templates are versioned assets, not state machines; registry content is process state
rebuilt at startup.

#### Errors

E-AGT-009 (resolution failure); registration rejections carry the validation detail
(syntax, allowlist, slots, schema).

#### Constraints

No template may access I/O, clock, environment, or randomness (allowlist enforcement);
skill namespaces cannot shadow builtin names; registry mutation during a run never affects
that run.

#### Security

Overrides are code-adjacent assets behind Volume 9 trust gates; template content passes the
same injection threat analysis as skill content (Volume 9); the allowlist prevents
templates from exfiltrating environment data.

#### Observability

Registration and override events; the effective registry is queryable with provenance
(which source supplied each template) per Principle 7.

#### Performance

Templates parse once at registration and render from the parsed form; registry lookups are
in-memory. Render latency budgets are Volume 12's.

#### Compatibility

Template major versions follow the public-contract regime once the format is published for
skills (ADR-015, SM-20).

#### Acceptance criteria

- Given a workspace override with trust granted, when resolving the overridden name, then
  the override version serves and the run snapshot records it.
- Negative case: given an override using a function outside the allowlist, when loaded,
  then registration rejects it with the function named, and the builtin serves.
- Permission case: given a workspace override without a trust decision, when loading, then
  it is not loaded and the refusal is recorded and user-visible.
- Observability case: given any resolved template, when the run record is inspected, then
  the exact `(namespace, name, version)` is present and re-resolvable.

#### Verification method

Registry unit tests (precedence, shadowing, rejection matrices); trust-gate integration
tests; snapshot stability tests across reloads; golden registration diagnostics (Volume 13).

#### Traceability

PRD-006, PRD-007; ADR-044; FR-AGT-014; SM-12, SM-20.

### FR-AGT-014 — Deterministic rendering with provenance

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Derived
- Owner: Prompt Engine (Volume 4)
- Affected components: Prompt Engine, Agent Engine, Planner, Workflow Engine
- Dependencies: FR-AGT-013; ADR-044; NFR-AGT-003
- Related risks: RISK-AGT-004

#### Description

Rendering MUST be deterministic: identical template version, parameters, and slot material
produce byte-identical output, with no access to clock, environment, filesystem, or
randomness from template code. Every render MUST validate parameters against the template's
schema, fail on unfilled required slots (E-AGT-010) before any provider request is issued,
apply redaction classification to the output, enforce the render byte ceiling, and record
render provenance `(namespace, name, version, parameter_hash)` on the consuming turn.

#### Motivation

Determinism plus provenance is the reproducibility contract: run records reference prompts
compactly, and exact model inputs are reconstructable on demand (SM-12) without logging
prompt content by default (privacy precedence).

#### Actors

Agent Engine, Planner, Workflow Engine as callers; Prompt Engine.

#### Preconditions

A resolved template; caller-supplied parameters and slot material.

#### Main flow

Steps 1–6 of the rendering pipeline.

#### Alternative flows

- Optional slot unfilled: renders empty by declaration; the provenance hash covers the
  absence (parameter/material digest includes slot fill map).

#### Edge cases

- Material exceeding the byte ceiling: the render fails with E-AGT-010 naming the ceiling —
  the engine never truncates silently (truncation decisions belong to the Context Manager's
  budgeting, upstream).
- Parameter type mismatch: validation failure before rendering; the turn never starts.
- Template rendering to zero bytes: legal only for templates declaring emptiness (an
  explicit flag); otherwise E-AGT-010 (a guard against wiring mistakes).

#### Inputs

Template reference, parameters, slot material.

#### Outputs

Rendered prompt structure; render provenance; redaction classification verdicts.

#### States

Stateless service; provenance persists on Turn rows.

#### Errors

E-AGT-009 (resolution), E-AGT-010 (render failure: required slot, schema, ceiling, empty
guard).

#### Constraints

Byte-identical determinism; no environment access; provenance mandatory on every consumed
render.

#### Security

Redaction classification runs before output leaves the engine; secret material can never be
a slot input (Secret Store references never resolve in engine space — Volume 9).

#### Observability

`prompt.rendered` event with provenance and output size class; render failure diagnostics
name the failing slot/parameter.

#### Performance

Rendering is pure computation from parsed templates; budgets in Volume 12.

#### Compatibility

Output shape is provider-independent; provider-specific message formatting happens in
adapters (Volume 5), never in templates.

#### Acceptance criteria

- Given a template version, parameters, and material, when rendered 1,000 times across
  processes and platforms, then all outputs are byte-identical (determinism property).
- Negative case: given a missing required slot, when rendered, then E-AGT-010 returns
  naming the slot, and no provider request is issued for that turn.
- Error case: given material exceeding `agent.prompts.max_render_bytes`, when rendered,
  then the render fails with the ceiling named — no silent truncation.
- Observability case: given any persisted turn, when its provenance is used to re-render
  with the recorded material references, then the reconstruction matches the original
  (SM-12 replay input check).

#### Verification method

Determinism property tests (cross-platform golden renders per template version); slot and
schema failure matrices; provenance round-trip tests in the replay suite (Volume 13).

#### Traceability

PRD-006; ADR-044; FR-AGT-013; NFR-AGT-003; SM-12.

## Non-functional requirements

### NFR-AGT-003 — Prompt render determinism

- Category: Maintainability
- Priority: P0
- Phase: MVP
- Metric: Fraction of renders producing byte-identical output for identical (template version, parameters, slot material) across repeated runs, processes, and Tier 1 platforms
- Target: 100%
- Minimum threshold: 100% (any divergence is a defect; there is no tolerance band)
- Measurement method: determinism property suite: for every builtin template version, ≥ 100 repeated renders with fixed fixtures per platform, byte-compared; golden render files verified in CI per release
- Test environment: CI on all Tier 1 platforms (Volume 13 harness)
- Measurement frequency: every mainline commit (suite); every release (golden audit)
- Owner: Prompt Engine (Volume 4)
- Dependencies: FR-AGT-013, FR-AGT-014
- Risks: RISK-AGT-004
- Acceptance criteria: The determinism suite passes with zero divergent renders on every Tier 1 platform; golden render files are stable across a release line except with an explicit template version bump.

## Configuration

Keys minted by this chapter, in the `[agent.prompts]` table (schema and precedence:
Volume 10).

```toml
[agent.prompts]
allow_workspace_overrides = false
override_dirs = []
max_render_bytes = 262144
```

| Key | Type | Default | Meaning |
|---|---|---|---|
| `agent.prompts.allow_workspace_overrides` | boolean | `false` | Enables loading workspace-scope template overrides (still trust-gated per Volume 9) |
| `agent.prompts.override_dirs` | array of strings | `[]` | Additional global override directories, resolved per ADR-022 |
| `agent.prompts.max_render_bytes` | integer | `262144` | Hard ceiling per rendered prompt, in bytes |

## Events

| Event | Producer | Emitted when | Payload highlights |
|---|---|---|---|
| `prompt.template.registered` | Prompt Engine | template accepted into the registry | namespace, name, version, source |
| `prompt.template.overridden` | Prompt Engine | an override shadows a lower-precedence version | name, winning and shadowed sources |
| `prompt.rendered` | Prompt Engine | render completed for a consumer | provenance tuple, output size class, consumer turn ID |

## Error codes

### E-AGT-009 — Prompt template resolution failure

- Category: Configuration
- Severity: Error
- User message: "The prompt template <name> could not be found or resolved."
- Technical message: reference (namespace/name/version), registry sources searched, deprecation or range-mismatch detail
- Cause: unknown template name, version absent, all versions deprecated without a pin, or an override rejected at load left no candidate
- Safe-to-log data: reference tuple, sources searched
- Recoverability: recoverable — fix the reference, install the template, or pin a version
- Retry policy: none
- Recommended action: inspect the registry (command per Volume 8); check profile `prompt_refs`
- Exit-code mapping: 3
- HTTP mapping: not applicable
- Telemetry event: `run.failed` (error code carried)
- Security implications: resolution never falls back to a different template silently

### E-AGT-010 — Prompt render failure

- Category: Internal defect / configuration
- Severity: Error
- User message: "A prompt could not be rendered; the request was not sent."
- Technical message: template provenance, failing element (required slot, parameter schema violation, byte ceiling, empty-output guard), caller identity
- Cause: unfilled required slot, parameter validation failure, output exceeding `agent.prompts.max_render_bytes`, or an undeclared empty render
- Safe-to-log data: provenance tuple, failing element class (never rendered content)
- Recoverability: recoverable for configuration causes; a defect report for wiring causes
- Retry policy: not retryable (deterministic failure)
- Recommended action: fix the template, parameters, or profile wiring named in the diagnostic
- Exit-code mapping: 1; 3 when caused by template/profile configuration
- HTTP mapping: not applicable
- Telemetry event: `prompt.rendered` (outcome: failed)
- Security implications: fail-closed — no partially rendered prompt is ever sent to a provider

## Risks

### RISK-AGT-004 — Prompt template drift and override injection

- Category: Security / product
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Trust-gated override loading (Volume 9); workspace overrides disabled by default; function allowlist denies I/O, clock, and environment access; version snapshots isolate running runs; provenance on every turn makes the serving template attributable; golden render files detect unintended builtin drift
- Detection: registration rejection and override events; golden render diffs in CI; audit of `prompt.template.overridden` occurrences in workspaces
- Owner: Prompt Engine (Volume 4)
- Status: Open

A malicious or careless repository could ship a workspace template override that redirects
agent behavior — a prompt-injection vector with persistence. The engine treats overrides as
untrusted code-adjacent assets: off by default, trust-gated, sandboxed to a pure rendering
function space, and always attributable through provenance.
