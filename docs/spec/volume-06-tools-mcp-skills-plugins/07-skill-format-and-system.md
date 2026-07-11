# 07 — Skill Format and System

A **Skill** is a packaged, versioned unit of procedural knowledge — prompts plus declared
requirements — loaded by the **Skill Engine** (Volume 3, chapter 04) and applied to agent
behavior. Skills carry **no executable code** (Volume 2, INV-SKL-04): anything executable
enters as a Plugin or Tool, and a skill references it. This chapter defines the skill
format and manifest ([ADR-078](../annexes/adr/ADR-078.md)), identity and versioning,
inputs and outputs, prompts, requirement declarations, inheritance/composition/overrides
([ADR-082](../annexes/adr/ADR-082.md)), testing and fixtures, and
publication/installation/signing/trust/deprecation. Runtime execution semantics (how the
Skill Engine applies an active skill during a run) are owned by Volume 4, chapter 08; this
chapter owns everything about what a skill *is*.

## Package layout and manifest

A skill is a directory (installed under the skill install location per scope; layout
below) with a `skill.toml` manifest at its root:

```text
my-skill/
  skill.toml            # manifest (required)
  prompts/
    main.md             # primary prompt document (required unless inherited)
    review-checklist.md # additional documents referenced from the manifest
  templates/
    report.tmpl.md      # output templates (optional)
  tests/
    cases.toml          # test case definitions (optional; FR-SKILL-004)
    fixtures/           # fixture inputs and golden outputs
```

```toml
[skill]
name = "conventional-review"
version = "1.2.0"
description = "Reviews diffs against the project's conventions and drafts findings."
authors = ["Ada Contributor <ada@example.com>"]
license = "Apache-2.0"
format_version = "1.0"            # skill format contract version (SM-20 surface)

[skill.prompts]
entry = "prompts/main.md"          # contributed to the agent system prompt slot
documents = ["prompts/review-checklist.md"]

[skill.inputs]
schema = """
{ "type": "object",
  "properties": { "strictness": { "type": "string", "enum": ["advisory", "blocking"] } },
  "additionalProperties": false }
"""                                # JSON Schema (ADR-024); optional

[skill.outputs]
artifacts = ["review-report"]      # declared artifact names this skill guides runs to produce

[skill.requires]
tools = ["fs.read", "git.diff"]    # tool name selectors that must resolve at activation
capabilities = ["tool_calling"]    # Volume 5 capability enum values
skills = { "diff-reading" = ">=1.0.0 <2.0.0" }  # skill dependencies, semver ranges

[skill.providers]
include = ["*"]                    # provider selectors; default all
exclude = []

[skill.workflows]
suggested = ["spec-driven-dev"]    # workflow names this skill is designed to accompany

[skill.extends]
parent = ""                        # single-parent inheritance; empty = none

[skill.overrides]
prompts = false                    # whether this skill may replace (not append to) parent prompts

[skill.deprecation]
deprecated = false
replacement = ""
```

The manifest MUST parse as TOML (ADR-008) and validate against the skill manifest schema
(JSON Schema per ADR-024, shipped with the binary and mirrored in `sdk/`). The following
manifest is invalid — executable content declared, malformed version — and MUST be
rejected with E-SKILL-001 / E-SKILL-006:

```toml invalid
[skill]
name = "bad skill!"        # invalid characters in name
version = "latest"          # not semver

[skill.hooks]
on_activate = "./run.sh"    # executable content is prohibited in skills (INV-SKL-04)
```

Field rules:

- `name`: lowercase `[a-z0-9-]+`, 3–64 characters, unique per scope (INV-SKL-01).
- `version`: SemVer (ADR-015 discipline); content immutable per version (INV-SKL-02).
- `format_version`: the skill-format contract version this skill targets; the engine
  supports the current and previous minor within a major (SM-20 regime).
- `requires.capabilities`: values from the closed Volume 5 capability enum only.
- `requires.tools`: tool name selectors; MCP-origin tools are addressed by their
  namespaced form (`mcp:<server>/<tool>`), making cross-machine portability of such
  skills explicitly dependent on the same server registration.
- Every path in the manifest MUST resolve inside the skill directory; absolute paths and
  `..` traversal are validation errors (E-SKILL-001).

### FR-SKILL-001 — Skill format and manifest

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: Skill Engine (Volume 6)
- Affected components: Skill Engine, Prompt Engine (consumer), Package Manager, Extension SDK
- Dependencies: ADR-008, ADR-015, ADR-024, ADR-078; FR-CFG-001
- Related risks: RISK-SKILL-001

#### Description

Andromeda MUST define and validate the skill package format above: a directory with a
`skill.toml` manifest declaring identity (name, version, description, authors, license),
format contract version, prompt documents, optional input JSON Schema and output artifact
declarations, requirement declarations (tools, capabilities, skill dependencies),
provider selectors, workflow associations, single-parent inheritance, override
declarations, and deprecation metadata. Skills MUST NOT contain executable content or
declare execution hooks; validation MUST reject any manifest key or content file class
outside the declared format (fail-closed schema with `additionalProperties: false` at the
top level). Skill content is content-addressed: the `content_hash` (SHA-256 over the
sorted file manifest and file digests) is computed at install/registration and verified at
every load (INV-SKL-02).

#### Motivation

A precise, data-only format keeps the skill trust story reviewable (text, not code) and
makes skills portable across providers and machines to the exact extent their declared
requirements resolve (Principle 2 honesty).

#### Actors

Skill authors; Skill Engine; Package Manager; agents consuming activated skills.

#### Preconditions

Skill directory present at a registered location or delivered by a package.

#### Main flow

1. The engine parses `skill.toml`, validates against the manifest schema, and checks path
   containment and the no-executable rule.
2. It computes/verifies `content_hash` and registers the Skill row (scope, natural key
   per Volume 2).
3. The skill becomes loadable; activation follows FR-SKILL-002.

#### Alternative flows

- Manifest invalid: E-SKILL-001 with field-level findings; nothing registers.
- Hash mismatch on a later load: skill disabled with E-SKILL-002; user notified.

#### Edge cases

- Empty `prompts.entry` is valid only when `extends.parent` names a parent that provides
  one (checked at resolution, E-SKILL-004 otherwise).
- Two versions of the same skill MAY be registered simultaneously; activation selects per
  FR-SKILL-002 pinning rules.
- Symlinks inside skill directories are not followed (path containment); a symlinked file
  is a validation error naming the path.

#### Inputs

Skill directories; manifest schema; scope context.

#### Outputs

Validated Skill rows; validation reports; `skill.registered` events.

#### States

Skills are stateless catalog entities (Volume 2); `enabled` is a recorded flag.

#### Errors

E-SKILL-001, E-SKILL-002, E-SKILL-006.

#### Constraints

TOML manifest (ADR-008); JSON Schema validation (ADR-024); SemVer (ADR-015); no
executable content (INV-SKL-04); UTF-8 content files.

#### Security

Prompts alter agent behavior: skills are trust-classified (Volume 9 vocabulary) at
registration by origin (builtin, packaged-verified, packaged-unverified, local-path) and
the classification is visible wherever skills are listed or activated.

#### Observability

Registration/validation events; validation findings in structured logs; the active skill
set is recorded per run (SM-12 reproducibility).

#### Performance

Per NFR-SKILL-001 validation budgets.

#### Compatibility

`format_version` gates loading: current and previous minor within the major are accepted;
newer-than-supported is rejected with E-SKILL-005 semantics (compatibility class), never
best-effort parsed.

#### Acceptance criteria

- Given a well-formed skill directory, when registered, then the Skill row exists with
  correct natural key, `content_hash`, and trust classification, and `skill.registered`
  is emitted.
- Given a manifest declaring an execution hook, when validated, then E-SKILL-006 is
  reported and the skill does not register.
- Given a registered skill whose content file is modified on disk, when next loaded, then
  E-SKILL-002 disables it and the mismatch is logged with both digests.
- Negative case: a manifest with `version = "latest"` fails with E-SKILL-001 naming the
  field.
- Permission case: registration from a local path in a non-interactive run follows policy
  (trust-gated; denial recorded) rather than silently registering.
- Observability case: the validation report enumerates every finding, not only the first.

#### Verification method

Golden manifest tests (valid and invalid corpus); hash-verification tests with tampered
content; schema round-trip tests against the SDK-mirrored schema; lint of the shipped
manifest schema itself.

#### Traceability

PRD-007; ADR-078, ADR-082; Volume 2 Skill entity and INV-SKL-01..04; FR-SKILL-002;
SM-02-adjacent authoring economics (register cross-reference).

## Loading, validation, and activation

### FR-SKILL-002 — Skill loading, requirement resolution, and activation

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: Skill Engine (Volume 6)
- Affected components: Skill Engine, Tool Runtime (registry reads), Agent Engine (consumer), Prompt Engine
- Dependencies: FR-SKILL-001; FR-TOOL-001 (registry semantics)
- Related risks: RISK-SKILL-001

#### Description

Activation makes a registered skill effective for a session, agent profile, or run. The
Skill Engine MUST, at activation time: resolve the skill's dependency closure
(FR-SKILL-003 rules); verify every required tool selector resolves to an enabled
registered tool and every required capability is present in the active model's declared
CapabilitySet (Volume 5 vocabulary); verify provider selectors match the active provider;
and only then contribute the skill's prompts and requirement declarations to the agent
configuration. Unsatisfiable requirements MUST fail activation with E-SKILL-003 listing
every missing item (INV-SKL-03 — reported, never silently degraded). Version selection:
an activation request names a version or a semver range; unpinned requests select the
highest registered non-deprecated version. The active skill set (names, versions,
content hashes) MUST be recorded on the run record.

#### Motivation

Silent degradation of missing requirements would break Principle 2 (capability honesty)
and make runs unreproducible; explicit resolution keeps skill effects deterministic and
recorded.

#### Actors

Users activating skills; agent profiles referencing skills; Skill Engine.

#### Preconditions

Skill registered and enabled; target session/profile identified.

#### Main flow

1. Resolve version, dependencies, inheritance chain, and composition set.
2. Check tools, capabilities, providers; compute the effective prompt contribution.
3. Record activation; contribute to the agent configuration; emit `skill.activated`.

#### Alternative flows

- Deactivation removes contributions at the next turn boundary and emits
  `skill.deactivated`.
- Requirement lost mid-session (tool disabled, provider switched): affected skills are
  suspended with notification; the run record notes the change (Volume 5 change
  notification pattern).

#### Edge cases

- Activating two versions of the same skill simultaneously is rejected (E-SKILL-004,
  conflict class).
- A skill requiring an MCP-origin tool activates only if the server is exposed
  (chapter 06); the missing-item report names the server registration.
- Capability checks use declared capabilities, never model-name heuristics.

#### Inputs

Activation requests (skill selector, target scope), registries (skills, tools),
CapabilitySet of the active model.

#### Outputs

Effective agent configuration contributions; activation records; events.

#### States

Activation state is per-session configuration (Volume 3, Skill Engine row); skills remain
stateless catalog entities.

#### Errors

E-SKILL-003 (requirements), E-SKILL-004 (resolution/conflict), E-SKILL-005
(composition/compatibility).

#### Constraints

Resolution is deterministic: identical registries and requests yield identical outcomes;
the resolver has no network access.

#### Security

Activation of a skill below the trust threshold set by policy requires an Approval per
Volume 9; a skill cannot grant itself tools — its requirements still pass the permission
model at invocation time.

#### Observability

`skill.activated`/`skill.deactivated` with version and hash; the composition report
(FR-SKILL-003) is attached to the run record.

#### Performance

Activation resolution for a set of 20 skills completes within the NFR-SKILL-001 budget.

#### Compatibility

Version-range activation insulates profiles from patch churn while pinning stays
available for reproducibility.

#### Acceptance criteria

- Given a skill whose requirements all resolve, when activated, then its prompt
  contribution is present in the next turn's assembled prompt (verifiable via the
  run-record prompt references) and the run record lists name/version/hash.
- Given a skill requiring a capability the model does not declare, when activated, then
  E-SKILL-003 lists the capability and the activation changes nothing.
- Negative case: activating a disabled skill fails with a not-loadable finding.
- Permission case: activating a local-path skill classified below the policy threshold
  raises an Approval; denial leaves the skill inactive with a recorded decision.
- Observability case: suspension on requirement loss emits `skill.deactivated` with the
  cause class.

#### Verification method

Integration tests over registries with controlled gaps; run-record inspection tests;
property tests on resolver determinism (identical inputs → identical outputs).

#### Traceability

Principle 2 (Volume 1); INV-SKL-03; Volume 4 chapter 08 (execution semantics); Volume 5
capability enum.

## Inheritance, composition, and overrides

Model (full rationale in [ADR-082](../annexes/adr/ADR-082.md)):

- **Inheritance** — `extends.parent` names one parent skill (same registry, semver range
  allowed). The child inherits prompt documents and requirement declarations; child
  declarations append. Replacement of a parent prompt document requires
  `overrides.prompts = true` in the child *and* is limited to documents the child names
  explicitly. Inheritance chains are bounded at depth 5; cycles are errors.
- **Composition** — multiple active skills compose per deterministic precedence:
  workspace scope over global over builtin; within a scope, explicit activation order;
  ties broken by name. Prompt contributions concatenate in precedence order into their
  Prompt Engine slots; requirement sets union.
- **Conflicts** — two skills declaring the same *named artifact* output or replacing the
  same prompt document are a composition conflict (E-SKILL-005) resolved only by explicit
  user choice or deactivation; the engine never picks silently.

### FR-SKILL-003 — Skill inheritance, composition, and overrides

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: Skill Engine (Volume 6)
- Affected components: Skill Engine, Prompt Engine
- Dependencies: FR-SKILL-001, FR-SKILL-002; ADR-082
- Related risks: RISK-SKILL-001

#### Description

The Skill Engine MUST implement single-parent inheritance with append-by-default and
explicit, declared overrides; deterministic multi-skill composition with the precedence
order above; a composition report enumerating every contribution, its source skill, and
every override applied; and conflict detection that fails closed (E-SKILL-005) for
same-target replacements. Dependency and inheritance resolution MUST detect cycles and
depth violations (E-SKILL-004). The composition result for a given set of skill versions
MUST be reproducible and is recorded with the run.

#### Motivation

Composition is where skills interact; determinism and explicit conflicts keep combined
behavior auditable instead of order-of-load-dependent.

#### Actors

Skill authors (declaring extends/overrides); users activating sets; Skill Engine.

#### Preconditions

Participating skills registered; activation request defines the set.

#### Main flow

1. Resolve inheritance chains bottom-up; apply declared overrides.
2. Order the composition set by precedence; union requirements; concatenate prompts per
   slot.
3. Emit the composition report; attach to the activation record.

#### Alternative flows

- Conflict: E-SKILL-005 names both skills and the contested target; activation of the
  conflicting subset fails; previously active skills are unaffected.

#### Edge cases

- A parent version bump that removes a document a child overrides: resolution fails with
  E-SKILL-004 naming the missing target (no dangling overrides).
- Deactivating a skill mid-set recomputes composition at the next turn boundary; the
  report is re-emitted with a new revision.
- An inheritance chain crossing scopes (workspace child of a global parent) is permitted;
  precedence applies to composition, not inheritance.

#### Inputs

Manifests (extends/overrides), activation set, scope context.

#### Outputs

Effective composition, composition report, events.

#### States

Not applicable — pure resolution over catalog entities; no new machine.

#### Errors

E-SKILL-004, E-SKILL-005.

#### Constraints

Depth ≤ 5; single parent; no dynamic (runtime-conditional) composition — the set is fixed
per activation.

#### Security

Overrides cannot escalate: the composed requirement set is the union (never a reduction
that hides a tool a prompt still references); trust classification of the composed set is
the minimum of its members for approval purposes.

#### Observability

Composition reports are structured artifacts referenced from run records;
`skill.composition.resolved` event carries the set hash.

#### Performance

Resolution is in-memory over registered manifests; NFR-SKILL-001 budget applies.

#### Compatibility

Composition semantics version with `format_version`; report schema is additive-evolution
per Volume 2 chapter 10 serialization rules.

#### Acceptance criteria

- Given a child with `overrides.prompts = true` naming a parent document, when composed,
  then the child document replaces the parent's and the report records the override.
- Given two skills declaring the same output artifact name, when activated together, then
  E-SKILL-005 is raised and neither contributes until the user resolves.
- Negative case: a cycle A extends B extends A fails with E-SKILL-004 listing the cycle.
- Observability case: identical activation sets produce identical set hashes across
  processes (determinism check).
- Permission case: composed sets containing any below-threshold skill require the same
  Approval as FR-SKILL-002 single activation.

#### Verification method

Composition conflict matrix tests (Volume 3 Skill Engine testing row); property tests for
determinism and cycle detection; golden composition reports.

#### Traceability

ADR-082; FR-SKILL-002; Volume 4 prompt assembly (Prompt Engine slots).

## Testing and fixtures

### FR-SKILL-004 — Skill testing and fixtures

- Type: Functional
- Status: Approved
- Priority: P2
- Phase: Beta
- Source: Provided
- Owner: Skill Engine (Volume 6)
- Affected components: Skill Engine, Extension SDK, CLI (test command surface per Volume 8)
- Dependencies: FR-SKILL-001; ADR-017
- Related risks: RISK-SKILL-001

#### Description

The skill format includes an optional `tests/` directory: `cases.toml` defines named test
cases (inputs conforming to the input schema, activation context stubs, expected
composition outcomes, and golden rendered-prompt files under `tests/fixtures/`). Andromeda
MUST provide a skill test runner (surfaced via the CLI skill command group, grammar per
Volume 8) that executes cases without any provider call: it validates the manifest,
resolves composition in an isolated registry containing the skill and declared stub
dependencies, renders prompt contributions with the case inputs, and compares against
golden fixtures. The Extension SDK skill template MUST ship with a passing example test.
Test execution requires no permissions beyond `read` on the skill directory.

#### Motivation

Skills change agent behavior through text; golden testing makes that text diffable and
regression-checked like code, and provider-free execution keeps tests deterministic and
offline.

#### Actors

Skill authors; CI of skill repositories; the test runner.

#### Preconditions

Skill directory with `tests/`; runner available.

#### Main flow

1. Runner validates and resolves the skill in isolation.
2. Each case renders contributions and diffs against goldens.
3. Results report per case; non-zero process exit on failure.

#### Alternative flows

- Missing goldens: the runner can record mode-write them on explicit flag (never by
  default), reporting created files.

#### Edge cases

- Cases exercising unsatisfied requirements assert the E-SKILL-003 report content rather
  than rendering.
- Golden comparison normalizes line endings; content is otherwise byte-exact.

#### Inputs

`cases.toml`, fixtures, the skill and stub dependencies.

#### Outputs

Case results, diffs, optional recorded goldens.

#### States

Not applicable — test tooling, no runtime entity states.

#### Errors

Runner reports reuse E-SKILL-001..005 where the failure is a format/resolution failure;
case failures are test outcomes, not error-catalog entries.

#### Constraints

No network; no provider calls; deterministic rendering (fixed clock and ULID stubs in the
runner context).

#### Security

The runner never executes skill-referenced tools; requirement checks are registry
lookups against stubs.

#### Observability

Runner output is structured (`--json` output shape owned by Volume 8's CLI conventions);
results carry the skill content hash for traceability.

#### Performance

A 20-case suite completes in ≤ 5 s on the Volume 1 reference hardware (runner overhead,
no model latency by construction).

#### Compatibility

Case schema versions with `format_version`; goldens are plain text.

#### Acceptance criteria

- Given the SDK skill template, when its tests run, then all cases pass offline with all
  network interfaces disabled.
- Given a prompt document edit that changes rendered output, when tests run, then the
  affected case fails showing a unified diff.
- Negative case: a `cases.toml` referencing a missing fixture fails that case with the
  path named, other cases unaffected.
- Observability case: `--json` results include per-case status, duration, and the skill
  content hash.

#### Verification method

Runner self-tests in the monorepo; SDK template test executed in CI (SM-02-adjacent
walkthrough pattern); offline-suite inclusion (Volume 13).

#### Traceability

ADR-017; FR-SDK-001 (template delivery); Volume 13 fixtures/determinism chapter.

## Publication, installation, signing, trust, and deprecation

Skills are distributed as extension packages of kind `skill` through the Package Manager
([chapter 09](09-package-manager-supply-chain.md)); local development uses registered
paths (`skills.paths`) without packaging.

- **Publication** — a skill package is an archive of the skill directory plus package
  metadata; checksums are mandatory, signatures per ADR-081 policy. Publication targets
  are the chapter 09 sources (registry index, git, archive URL); a curated marketplace is
  Future (ADR-080).
- **Installation** — `package_installation` permission; frozen Package installation
  states; post-install the skill registers per FR-SKILL-001. Workspace-scope installs
  never shadow global installs silently: shadowing is reported at registration.
- **Signing and trust** — `signature_state` from package verification feeds the trust
  classification (Volume 9 vocabulary): unsigned local-path skills sit below packaged
  verified skills; policy thresholds gate activation approvals (FR-SKILL-002).
- **Deprecation** — `deprecation.deprecated = true` with optional `replacement` excludes
  the version from unpinned selection, warns at activation, and marks listings.
  Deprecated versions remain loadable while referenced (INV-SKL retention rules;
  reproducibility).

### FR-SKILL-005 — Skill distribution and deprecation

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: Skill Engine (Volume 6)
- Affected components: Skill Engine, Package Manager, Policy Engine
- Dependencies: FR-SKILL-001, FR-PLUG-005 (package operations), FR-PLUG-007 (verification); ADR-080, ADR-081
- Related risks: RISK-SKILL-001, RISK-PLUG-003

#### Description

Skill distribution MUST flow through the extension package pipeline: packaging (archive +
metadata), source acquisition, checksum verification (unconditional), signature
verification per policy, staged installation, and registration. The Skill Engine MUST
map `signature_state` and source kind into the skill's trust classification and enforce
activation policy accordingly. Deprecation metadata MUST affect version selection
(excluded from unpinned resolution), listings (marked), and activation (warning event)
without breaking pinned reproducibility. Uninstallation MUST deregister the skill, leave
tombstoned provenance, and never delete versions still referenced by persisted run
records.

#### Motivation

Skills steer agents; their acquisition path must carry the same integrity guarantees as
executable extensions, proportionate to their text-only content.

#### Actors

Skill publishers; users installing; Package Manager; Policy Engine.

#### Preconditions

Package source configured; permission resolved.

#### Main flow

1. Install: resolve → download → verify → stage → install → register (frozen states).
2. Trust classification computed and recorded; listings show it.
3. Deprecation metadata honored at selection/activation time.

#### Alternative flows

- Verification failure: installation `failed` with the chapter 09 error; nothing
  registers.
- Uninstall with active references: skill deactivates for future runs; historical records
  keep name/version/hash.

#### Edge cases

- Installing a version equal to a registered local-path skill: both exist; scope and
  precedence rules decide selection, and listings disambiguate origins.
- A deprecated version that is the only satisfier of another skill's dependency range
  still resolves, with the warning attached to the composition report.

#### Inputs

Skill packages, source configuration, policy thresholds, deprecation metadata.

#### Outputs

Installed and registered skills with trust classification; deprecation warnings;
tombstones on removal.

#### States

Package installation machine (chapter 10); skills themselves stateless.

#### Errors

Chapter 09 E-PLUG package errors during acquisition; E-SKILL-001/002 at registration.

#### Constraints

No skill executes anything at install time (data-only guarantee holds through the
pipeline); no automatic skill updates without consent.

#### Security

Signature policy per ADR-081; trust thresholds per Volume 9; the install path is the
supply-chain control point for skills.

#### Observability

Package events (chapter 09) plus `skill.registered`/`skill.deprecated`; provenance chain
skill → package → source navigable offline.

#### Performance

Install of a 1 MB skill package from a local archive completes in ≤ 2 s on reference
hardware (dominated by hashing).

#### Compatibility

`format_version` gating at registration; package `contract_version` checks at resolve
time (chapter 09).

#### Acceptance criteria

- Given a packaged skill with a valid checksum and signature per policy, when installed,
  then it registers with the packaged-verified trust classification and is listed with
  provenance.
- Given `signature_state = invalid`, when installing, then installation fails regardless
  of configuration (INV-PKG-02) and no skill registers.
- Given a deprecated version, when an unpinned activation resolves, then a newer
  non-deprecated version is selected; when pinned to the deprecated version, activation
  proceeds with `skill.deprecated` emitted.
- Permission case: installation without `package_installation` fails with the E-SEC
  denial and a recorded decision.
- Observability case: uninstall leaves a tombstoned Extension record and the historical
  run records still resolve name/version/hash.

#### Verification method

End-to-end package tests (fixture registry and archives); signature policy matrix tests;
deprecation selection tests; audit-chain inspection.

#### Traceability

ADR-080, ADR-081; chapter 09; INV-PKG-01..03; Volume 9 trust policy.

## Configuration keys (`[skills]`)

| Key | Type | Default | Meaning |
|---|---|---|---|
| `skills.enabled` | bool | `true` | Master switch for the Skill Engine |
| `skills.paths` | array of paths | `[]` | Additional local skill directories (development; lowest trust) |
| `skills.autoload` | bool | `true` | Register skills found at configured locations at startup |
| `skills.sources` | array of tables | `[]` | Package sources for skill discovery/installation (chapter 09 source schema) |
| `skills.activation_policy` | string | `"prompt"` | `prompt` \| `allow` \| `deny` for below-threshold trust activations (decision semantics per Volume 9) |

## Events minted (skill.*)

Envelope per Volume 10 (FR-OBS-001).

| Event | Version | Producer | Consumers | Payload summary |
|---|---|---|---|---|
| `skill.registered` | 1 | Skill Engine | TUI, Observability | name, version, scope, trust classification, content hash |
| `skill.validation.failed` | 1 | Skill Engine | TUI, Observability | name (if parseable), finding count, first finding class |
| `skill.activated` | 1 | Skill Engine | TUI, Observability, Audit Log | name, version, target scope, content hash |
| `skill.deactivated` | 1 | Skill Engine | TUI, Observability | name, version, cause class (user, requirement-lost, conflict) |
| `skill.composition.resolved` | 1 | Skill Engine | Observability | set hash, member count, override count |
| `skill.deprecated` | 1 | Skill Engine | TUI, Observability | name, version, replacement (if declared) |

## Requirements — non-functional and risk

### NFR-SKILL-001 — Skill validation and composition latency

- Category: Performance
- Priority: P2
- Phase: Beta
- Metric: (a) Wall-clock time to validate one skill (manifest parse, schema validation, hash verification, ≤ 50 files / 1 MB content); (b) time to resolve activation composition for 20 registered skills
- Target: (a) ≤ 50 ms p95; (b) ≤ 100 ms p95
- Minimum threshold: (a) ≤ 100 ms p95; (b) ≤ 250 ms p95
- Measurement method: benchmark harness over the fixture skill corpus, 50 iterations, p95, per release
- Test environment: Volume 1 reference hardware (both reference machines per Volume 12 formalization)
- Measurement frequency: per release; regression-tracked in the Volume 12 benchmark suite
- Owner: Skill Engine (Volume 6)
- Dependencies: FR-SKILL-001, FR-SKILL-003
- Risks: RISK-SKILL-001
- Acceptance criteria: Benchmark report shows both percentiles within target on both reference machines; breach of minimum threshold blocks release per Volume 12 gating.

### RISK-SKILL-001 — Malicious or manipulative skill content

- Category: Security
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: data-only format (INV-SKL-04) with executable-content rejection (E-SKILL-006); trust classification and activation policy thresholds; signature verification for packaged skills; composition reports making every contribution attributable; provenance recorded per run (SM-12)
- Detection: validation rejections; activation approval flow; run-record audits attributing behavior changes to specific skill versions; Volume 9 injection threat monitoring
- Owner: Skill Engine (Volume 6) with Volume 9 threat model
- Status: Open

A skill is a prompt-injection vector with distribution: hostile instructions embedded in
plausible procedural text can steer agents toward unsafe tool use. The permission model
remains the hard boundary (a skill cannot grant tools), and attribution (every prompt
contribution recorded with its source skill and hash) makes post-incident analysis
tractable; residual risk is a user approving a hostile skill, addressed by trust UX and
Volume 9 threat guidance.

## Error codes (E-SKILL-001 – E-SKILL-006)

### E-SKILL-001 — Skill manifest invalid

- Category: Validation
- Severity: Error
- User message: "Skill '<name>' has an invalid manifest: <first finding>."
- Technical message: full field-level finding list (schema violations, path containment, name/version rules)
- Cause: malformed `skill.toml` or non-conforming content layout
- Safe-to-log data: skill path, finding classes and field names
- Recoverability: recoverable by fixing the manifest
- Retry policy: none
- Recommended action: run the skill test runner locally; fix listed findings
- Exit-code mapping: 3
- HTTP mapping: not applicable
- Telemetry event: `skill.validation.failed`
- Security implications: fail-closed — invalid skills never register

### E-SKILL-002 — Skill content integrity mismatch

- Category: Integrity
- Severity: Error
- User message: "Skill '<name>' was modified since installation and has been disabled."
- Technical message: expected and observed content hashes, first divergent path
- Cause: on-disk modification, corruption, or tampering (INV-SKL-02)
- Safe-to-log data: name, version, both digests, divergent path
- Recoverability: recoverable by reinstalling or re-registering the skill
- Retry policy: none
- Recommended action: reinstall from a trusted source; investigate unexpected modification
- Exit-code mapping: 9
- HTTP mapping: not applicable
- Telemetry event: `skill.validation.failed`
- Security implications: tamper detection; the skill contributes nothing while disabled

### E-SKILL-003 — Skill requirements unsatisfiable

- Category: Dependency
- Severity: Error
- User message: "Skill '<name>' cannot activate: missing <n> requirement(s)."
- Technical message: complete list of missing tools, capabilities, providers, and skill dependencies with the selector each failed
- Cause: required tool absent/disabled, capability not declared by the model, provider excluded, dependency unresolved
- Safe-to-log data: name, version, missing-item list
- Recoverability: recoverable by installing/enabling requirements or changing model/provider
- Retry policy: none automatic; re-activation after remediation
- Recommended action: address the listed items; the report names each
- Exit-code mapping: 1
- HTTP mapping: not applicable
- Telemetry event: `skill.deactivated`
- Security implications: no silent degradation (INV-SKL-03) — behavior never partially applies

### E-SKILL-004 — Skill resolution failed

- Category: Dependency
- Severity: Error
- User message: "Skill '<name>' could not be resolved: <cause class>."
- Technical message: cycle path, depth violation, missing parent/target, or duplicate-version conflict detail
- Cause: inheritance cycle, chain depth > 5, dangling override target, missing dependency version, or conflicting simultaneous versions
- Safe-to-log data: name, version, resolution graph summary
- Recoverability: recoverable by fixing declarations or pinning versions
- Retry policy: none (deterministic)
- Recommended action: inspect the resolution report; correct extends/overrides/dependencies
- Exit-code mapping: 1
- HTTP mapping: not applicable
- Telemetry event: `skill.validation.failed`
- Security implications: none beyond fail-closed resolution

### E-SKILL-005 — Skill composition conflict

- Category: Conflict
- Severity: Error
- User message: "Skills '<a>' and '<b>' conflict over <target>; choose one or adjust overrides."
- Technical message: contested target (artifact name or prompt document), both contributors with versions
- Cause: same-target replacement or duplicate artifact declaration in one activation set
- Safe-to-log data: both skill identities, target name
- Recoverability: recoverable by deactivating one contributor or changing declarations
- Retry policy: none
- Recommended action: resolve explicitly; the engine never auto-picks
- Exit-code mapping: 1
- HTTP mapping: not applicable
- Telemetry event: `skill.deactivated`
- Security implications: fail-closed composition prevents silent behavior override between skills

### E-SKILL-006 — Executable content rejected in skill

- Category: Validation
- Severity: Error
- User message: "Skill '<name>' declares executable content, which skills cannot contain."
- Technical message: offending manifest keys or file classes detected
- Cause: hooks, scripts, or binaries inside a skill package (INV-SKL-04 boundary rule)
- Safe-to-log data: name, offending key/path names
- Recoverability: recoverable by moving executable behavior to a plugin or tool and referencing it
- Retry policy: none
- Recommended action: repackage — procedural knowledge stays in the skill; code graduates to a plugin
- Exit-code mapping: 3
- HTTP mapping: not applicable
- Telemetry event: `skill.validation.failed`
- Security implications: enforces the data-only trust boundary that keeps skills reviewable as text
