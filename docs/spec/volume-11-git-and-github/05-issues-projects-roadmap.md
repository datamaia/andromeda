# 05 — Issues, Projects, and Roadmap

Issues are the unit of planned work and the anchor of the traceability chain between
requirements and branches (chapter 07). GitHub Projects provides the planning surface;
Milestones bind work to releases; Discussions absorbs everything that is not actionable
work so the issue tracker stays executable.

## Issue types

Every issue is created through an issue form (`.github/ISSUE_TEMPLATE/`, chapter 03) —
blank issues are disabled. Every form carries: a **Requirements** field (corpus IDs this
work realizes or affects; `process-only` is a valid value), an **Area** selector (the 24
ADR-015 scopes), a **Phase** selector, and type-specific fields listed below. The form
applies the matching `type/*` label automatically.

| Type | Form | Type-specific mandatory fields |
|---|---|---|
| Bug | `bug.yml` | Reproduction steps, expected/actual, version + platform, severity proposal, logs (redacted) |
| Feature | `feature.yml` | Problem statement, proposed behavior, acceptance sketch, affected contracts |
| Security | `security.yml` | Redirect-only: links private vulnerability reporting; public form collects only non-sensitive hardening work |
| Documentation | `documentation.yml` | Affected docs/spec paths, audience |
| Performance | `performance.yml` | Metric affected (Volume 12 NFR reference), measurement evidence, reference conditions |
| Refactor | `refactor.yml` | Motivation, behavior-preservation statement, risk class |
| Architecture | `architecture.yml` | Decision needed, ADR candidacy, affected volumes |
| Research | `research.yml` | Question, timebox, output artifact (report/decision) |
| Technical Debt | `tech-debt.yml` | Origin (PR/audit finding), interest being paid, remediation sketch |
| Release | `release.yml` | Target version, checklist reference (Volume 14), scope freeze date |
| Provider Integration | `provider-integration.yml` | Provider, official API documentation link, capability matrix sketch (Volume 5) |
| Tool Integration | `tool-integration.yml` | Tool, permission set, schema sketch (Volume 6) |
| MCP | `mcp.yml` | Server/feature, protocol revision, conformance impact (Volume 6) |
| Plugin | `plugin.yml` | Plugin/ARP surface, protocol version impact (Volume 6) |
| Platform Compatibility | `platform-compat.yml` | Platform + version, tier, divergence description (Volume 3 matrix) |

**Epics** are not a separate type: an epic is any issue carrying the `epic` label and a
task list of child issues; the traceability validator treats the epic as the
requirement→issue fan-out point. Security vulnerabilities are NEVER filed as public
issues; the security form's redirect and `SECURITY.md` route them to private reporting
(Volume 9 chapter 08 owns response).

## Label taxonomy

Labels are namespaced, defined as data in `.github/labels.yml`, and synchronized by the
labels workflow (chapter 06) — manual label edits outside the data file are reverted by
the sync. Selection rules: exactly one `type/*`; exactly one `priority/*` once triaged;
exactly one `phase/*`; `severity/*` mandatory for bugs and security; exactly one `ai/*`
on PRs (chapter 04); everything else as applicable.

| Namespace | Values | Meaning |
|---|---|---|
| `type/*` | the 15 types above | work classification |
| `priority/*` | `p0` `p1` `p2` `p3` | urgency; p0 = drop-everything |
| `severity/*` | `critical` `high` `medium` `low` | impact of bugs/security findings |
| `status/*` | `triage` `needs-info` `blocked` `wontfix` `duplicate` | tracker state beyond the Project board |
| `platform/*` | `macos` `linux` `windows` | platform specificity (windows = future phase work) |
| `area/*` | the 24 ADR-015 scopes | component/area routing; drives CODEOWNERS-aligned triage |
| `phase/*` | `core` `mvp` `beta` `v1` `v2` `future` | delivery phase binding |
| `risk/*` | `high` `medium` `low` | delivery/technical risk flag for planning |
| `breaking-change` | — | requires `!` marker + BREAKING CHANGE footer at merge |
| `security-review` | — | Volume 9-relevant change; owning-group review required |
| `backport` | — | targets or requires `release/*` cherry-picks |
| `epic` | — | fan-out tracking issue |
| `good-first-issue` | — | curated onboarding work with a mentor named in the issue |
| `help-wanted` | — | maintainers invite external contribution |
| `ai/none` `ai/assisted` `ai/generated` | — | PR provenance (chapter 04) |
| `size-exempt` | — | audited PR size exemption (chapter 04) |
| `policy-check` | — | audited maintainer override of a process validator (ADR-148) |

## GitHub Projects

One organization-level project — **Andromeda Roadmap** — tracks all work. Item statuses
are exactly: `Backlog`, `Ready`, `In Progress`, `In Review`, `Blocked`, `Validation`,
`Done`, `Released`. `Validation` holds merged work awaiting phase-gate or release
verification (Volume 13 gates); `Released` is set when the change ships in a published
release.

**Fields:**

| Field | Kind | Source |
|---|---|---|
| Status | single select (the 8 above) | automation + humans |
| Area | single select (24 scopes) | mirrors `area/*` label |
| Phase | single select | mirrors `phase/*` label |
| Priority | single select (P0–P3) | mirrors `priority/*` label |
| Size | single select (XS, S, M, L, XL) | triage estimate |
| Iteration | iteration field (2-week cadence) | planning |
| Target release | text (`vX.Y.Z`) | mirrors milestone |
| Risk | single select | mirrors `risk/*` label |
| Requirements | text (corpus IDs) | from the issue form |

**Views:** Board by Status (execution); Roadmap by Iteration/Target release (planning);
Table grouped by Area (triage and load); filtered views for `priority/p0`+`severity/*`
(incident planning), `phase/mvp` burn-down, and `type/security`+`security-review`
(security review queue).

**Automations:** new issues with a `type/*` label auto-add to the project in `Backlog`;
linked-PR opened → `In Review`; PR merged → `Validation`; release-published workflow
(chapter 06) moves shipped items to `Released` and stamps Target release; items closed
as not-planned leave the board. Status regressions (e.g., `Done` → `In Progress`) are
manual by design.

**Milestones** are release-scoped (`v0.3.0`) plus phase gates (`MVP`, `Beta`, `v1`).
Every committed (non-backlog) issue carries a milestone; the release issue type's
checklist reconciles milestone content at scope-freeze. **Dependencies** between issues
are recorded with task-list references and the `blocked` status plus a `Blocked by
#<issue>` line the validator can parse; cross-epic dependencies surface in the Roadmap
view. The public roadmap is the Roadmap view plus Volume 15's sequencing narrative —
there is no second roadmap document to drift.

## Requirements

### FR-GH-006 — Issue taxonomy and intake forms

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Core
- Source: Provided
- Owner: Development process (Volume 11)
- Affected components: None at runtime; platform templates + validators
- Dependencies: FR-GH-002; ADR-148
- Related risks: RISK-GH-002

#### Description

The repository MUST provide the 15 issue forms of this chapter, each collecting the
mandatory common fields (Requirements, Area, Phase) and the type-specific fields
listed, with blank issues disabled and the security form redirecting vulnerabilities to
private reporting. Forms MUST apply their `type/*` label automatically.

#### Motivation

Structured intake is what makes the requirement→issue link machine-checkable (ADR-148)
and keeps triage cost linear as contribution volume grows.

#### Actors

Reporters (users, contributors, agents), triagers, validators.

#### Preconditions

FR-GH-002 structure; labels synchronized.

#### Main flow

1. Reporter selects a form; mandatory fields enforce structure at submission.
2. The `type/*` label lands automatically; triage adds priority/severity/milestone.
3. The project automation places the item in `Backlog`.

#### Alternative flows

- Non-actionable submissions → converted to Discussions with a linking comment.
- Vulnerability reported publicly despite the redirect → maintainers minimize content,
  file privately, and follow Volume 9 chapter 08.

#### Edge cases

- Agent-filed issues: identical forms; the validator applies the same field
  requirements regardless of author.
- Requirements field on exploratory work: `process-only` prevents fake ID citations.

#### Inputs

Form submissions.

#### Outputs

Labeled, structured issues; project items.

#### States

Tracker states via `status/*` and the project Status field.

#### Errors

Missing mandatory fields are blocked at submission by the form engine; validator
findings on later edits use E-GH-003 semantics (chapter 07).

#### Constraints

Forms are data files under review like code; new types require a PR to this chapter.

#### Security

Vulnerability details never in public issues; reproduction logs redaction guidance in
the bug form.

#### Observability

Nightly audit reports issues missing triage labels past 5 working days.

#### Performance

Not applicable beyond platform behavior.

#### Compatibility

GitHub issue forms; portable as YAML data.

#### Acceptance criteria

- Given the New Issue flow, when a reporter opens it, then only the 15 forms (plus the
  config-linked external routes) are offered and blank issues are unavailable.
- Given a submitted bug form, when it lands, then `type/bug` is present and the
  Requirements/Area/Phase fields are non-empty.
- Given a vulnerability, when the security form is opened, then the private-reporting
  path is presented before any public field.
- Negative: a form edit removing the Requirements field fails the structure check
  (FR-GH-002).

#### Verification method

Form-schema validation in CI; submission walkthrough at phase gates; nightly audit
assertions.

#### Traceability

ADR-148; FR-GH-001; Volume 9 chapter 08 (security intake).

### FR-GH-007 — Label taxonomy as synchronized data

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Core
- Source: Provided
- Owner: Development process (Volume 11)
- Affected components: None at runtime; labels workflow
- Dependencies: FR-GH-002, FR-GH-006
- Related risks: RISK-GH-002

#### Description

The label set of this chapter MUST be defined in `.github/labels.yml` and synchronized
by automation: labels are created, renamed, recolored, and pruned to match the data
file, and the selection rules (exactly-one constraints) are enforced by the
traceability validator on issues and PRs.

#### Motivation

Labels drive routing, review requirements, provenance, and project mirroring; taxonomy
drift (ad-hoc labels, duplicates) breaks every automation built on them.

#### Actors

Maintainers (taxonomy changes via PR), sync workflow, validators.

#### Preconditions

Labels workflow deployed (chapter 06).

#### Main flow

1. Taxonomy changes land as PRs editing `labels.yml`.
2. The sync workflow applies the data on merge.
3. Validators enforce selection rules continuously.

#### Alternative flows

- Manual label created in the UI → next sync run removes it and the nightly audit
  reports the attempt.

#### Edge cases

- Label renames preserve issue history (platform rename, not delete+create) — the sync
  tool MUST rename by matching a stable `id` key in the data file.
- Removing a label in use requires the PR to state the migration (bulk re-label step).

#### Inputs

`labels.yml`; platform label state.

#### Outputs

Converged label set; validator findings.

#### States

Not applicable.

#### Errors

Sync failures surface as failed workflow runs; selection-rule violations use E-GH-003
semantics on the affected item.

#### Constraints

Namespaces and values exactly as specified; additions via PR to this chapter only.

#### Security

Label-driven review requirements (`security-review`) depend on sync integrity; the
sync job runs with minimal write scope on issues only (ADR-149 permissions).

#### Observability

Sync run summaries; nightly audit divergence report.

#### Performance

Not applicable.

#### Compatibility

GitHub labels API via pinned actions.

#### Acceptance criteria

- Given a merged edit to `labels.yml` renaming `risk/med` to `risk/medium`, when sync
  runs, then existing issues carry the renamed label and no duplicate exists.
- Given a manually created `misc` label, when sync runs, then it is removed and the
  audit records it.
- Given a triaged issue with two `priority/*` labels, when the validator runs, then the
  exactly-one violation is reported.
- Negative: the sync workflow token cannot push code (permission audit).

#### Verification method

Sync-tool unit tests against a fixture label state; permission audit of the workflow;
nightly audit assertions.

#### Traceability

FR-GH-006, FR-GH-001; ADR-149.

### FR-GH-008 — Projects, milestones, and roadmap operation

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Development process (Volume 11)
- Affected components: None at runtime; project automation
- Dependencies: FR-GH-006, FR-GH-007; ADR-148
- Related risks: RISK-GH-002

#### Description

The Andromeda Roadmap project MUST exist with the 8 statuses, 9 fields, and listed
views; the listed automations MUST run (intake to Backlog, PR-linked transitions,
release stamping); every committed issue MUST carry a milestone; and dependencies MUST
be recorded in the parseable `Blocked by` form. The Roadmap view is the public roadmap
surface.

#### Motivation

One planning surface with automated state keeps plan-versus-reality divergence visible
and gives Volume 15's sequencing a live instrument instead of a stale document.

#### Actors

Maintainers, contributors, automation, community (read).

#### Preconditions

Project created; automations deployed (chapter 06 project workflow).

#### Main flow

1. Intake automation adds labeled issues to Backlog.
2. Humans plan: Ready, iteration, milestone, size.
3. PR linkage moves execution states; merge → Validation; release → Released.

#### Alternative flows

- Work rejected at Validation (gate failure) → back to In Progress with a comment; the
  regression is manual and visible.

#### Edge cases

- Multi-PR issues: the issue reaches Validation only when all linked PRs merge (task
  list completion), otherwise stays In Progress.
- Issues without PRs (process/research): manual Done transition with the output
  artifact linked.

#### Inputs

Issue/PR events; release events; human planning edits.

#### Outputs

Live board, roadmap views, release-stamped history.

#### States

The 8 project statuses (planning vocabulary, not a product state machine).

#### Errors

Automation failures surface as workflow-run failures; stale items surface in the
nightly audit.

#### Constraints

One project; no per-team side boards for tracked work; status list changes require a
PR to this chapter.

#### Security

Project automation runs with project + issues scopes only (ADR-149).

#### Observability

Nightly audit: items In Progress with no linked open PR > 10 working days; Blocked
items with no `Blocked by` reference; milestone-less committed items.

#### Performance

Not applicable.

#### Compatibility

GitHub Projects (v2 GraphQL API) via pinned automation.

#### Acceptance criteria

- Given a merged PR closing issue #200, when merge completes, then #200's project item
  is in Validation with the PR linked.
- Given release v0.4.0 publishing with #200 shipped, when the release workflow runs,
  then #200 is Released with Target release v0.4.0.
- Given an issue labeled `status/blocked` without a parseable `Blocked by` line, when
  the nightly audit runs, then it is reported.
- Negative: an issue in Ready with no milestone fails the audit's committed-work rule.

#### Verification method

Automation integration tests against a fixture project; nightly audit assertions;
phase-gate review of board-versus-register consistency.

#### Traceability

FR-GH-001; Volume 15 sequencing; Volume 13 Validation gates.
