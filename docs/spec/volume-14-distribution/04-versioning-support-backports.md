# 04 — Versioning, Support Windows, and Backports

This chapter applies ADR-015 (SemVer + Conventional Commits) to the product and its public
contracts (FR-REL-012), defines the deprecation regime (FR-REL-013) and the support/backport
policy over release branches (FR-REL-014, ADR-193), and fixes what every changelog and
release note must contain (FR-REL-015). Commit-message rules, the closed scope list, and
their CI enforcement are ADR-015's and Volume 11's; they are cited, never restated.

## Versioned surfaces

The product version is one SemVer line; the public contracts carry their own explicit
versions, each owned by its volume. This table is the release-audit checklist: a release
MUST state the version of every row.

| Surface | Version carrier | Content owner |
|---|---|---|
| Product (binary, CLI grammar, exit codes, JSON output schemas) | Product SemVer | Volumes 8/0 |
| Public Go API (`sdk/` module) | Module SemVer | Volume 3 (layout), Volume 6 (SDK content) |
| Port contracts (`internal/ports`) | Product SemVer + contract-diff regime | Volume 3 (NFR-ARCH-002) |
| Andromeda Runtime Protocol (plugins) | Protocol version in handshake | Volume 6 (ADR-009) |
| Tool contract schema | `schema_version` | Volume 6 |
| Provider adapter contract | `schema_version` | Volume 5 |
| Skill format | Format version in manifest | Volume 6 |
| Workflow format | Format version in definition | Volume 4 |
| Configuration schema | Config `schema_version` | Volume 10 |
| Event envelope | Envelope version | Volume 10 |
| MCP compatibility | Supported protocol revision list | Volume 6 (ADR-010) |
| Database schemas | Monotonic migration integers (not SemVer) | Volume 10 procedures per ADR-029 |
| Distribution grammar (artifact names, installer flags, mirror `index.json`) | Product SemVer | Volume 14 (ADR-190) |

## Requirements

### FR-REL-012 — Semantic versioning of the product and public contracts

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Release engineering (Volume 14)
- Affected components: all public-contract-owning components; Updater; CI
- Dependencies: ADR-015; NFR-ARCH-002; SM-20
- Related risks: RISK-REL-004

#### Description

Product releases MUST carry SemVer versions computed from Conventional Commit history per
ADR-015, with pre-release identifiers per ADR-191. A **breaking change** is: removal or
rename of any element of a versioned surface above; a semantic change that invalidates a
previously valid input or changes the meaning of a previously produced output; tightening
of accepted inputs; or a database change outside ADR-029's regime. Additive,
backward-readable evolution (new optional fields, new commands, new enum values where the
surface declares open enums) is a MINOR change; fixes without contract effect are PATCH.
From v1, breaking changes ship only in a MAJOR release preceded by the FR-REL-013 window;
before v1, Volume 1 phase discipline applies (Core contracts move freely until MVP exit,
Beta breaks require migration notes).

#### Motivation

Extension authors, fleets, and scripts pin against these surfaces; a version number that
changes if and only if compatibility changes is the entire value of SemVer (ADR-015
rationale).

#### Actors

Maintainers; release automation deriving bumps; extension authors consuming contract
versions.

#### Preconditions

Conventional Commit history (ADR-015 CI checks, Volume 11).

#### Main flow

1. Release automation proposes the next version from commit types and `!`/`BREAKING
   CHANGE` markers.
2. The release audit verifies the versioned-surfaces table: every surface version stated,
   contract-diff results consistent with the bump class.
3. The release publishes with the surface-version table in its notes.

#### Alternative flows

- Undeclared breaking change detected by contract-diff after merge: the release is blocked
  until the commit history is corrected (revert or explicit re-marking) — automation never
  silently upgrades the bump class.

#### Edge cases

- A dependency upgrade that changes observable behavior of a surface is a change to that
  surface, classified by the definition above regardless of "it was the dependency".
- Database migration integers advance independently of SemVer; the release notes map the
  product version to the schema versions it migrates to.

#### Inputs

Commit history; contract-diff reports; surface version declarations.

#### Outputs

Version bump; per-surface version table in release notes.

#### States

Not applicable — governs release artifacts, not runtime state.

#### Errors

Specification/CI defects, not runtime E-codes.

#### Constraints

Versions are never reused (INV-REL-01, ADR-015); the breaking-change definition is closed
and extended only through the change procedure.

#### Security

Version honesty is a security property: fleets that pin "no majors" rely on it to avoid
unreviewed contract changes (NFR-REL-003 enforcement).

#### Observability

Contract-diff reports are release artifacts; `andromeda version` exposes the product
version and channel (Volume 8).

#### Performance

Not applicable.

#### Compatibility

This requirement *is* the compatibility regime's anchor at the release level; NFR-ARCH-002
applies it to ports; NFR-REL-003 measures it across the SM-20 list.

#### Acceptance criteria

- Given a commit history containing a `BREAKING CHANGE` footer, when the bump is computed,
  then it is MAJOR (or the release is pre-1.0 and the Beta migration-note rule applies).
- Given a release candidate whose contract-diff shows a removed public field with a
  MINOR-class bump, when the release audit runs, then publication is blocked.
- Given any published release, when its notes are inspected, then every versioned-surface
  row states a version.
- Negative case: given an attempt to republish an existing version with different
  artifacts, when the pipeline runs, then it refuses (INV-REL-01).

#### Verification method

Release-audit CI job (bump-class vs contract-diff consistency; surface-table completeness);
Volume 13 upgrade-test matrix across bump classes; SM-20 measurement tooling.

#### Traceability

ADR-015; SM-20; NFR-ARCH-002, NFR-REL-003; Volume 1 phase discipline; FR-REL-013/014.

### FR-REL-013 — Deprecation policy

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: Release engineering (Volume 14)
- Affected components: all public-contract-owning components; CLI; Logging
- Dependencies: FR-REL-012; ADR-015
- Related risks: RISK-REL-004

#### Description

Removing or breaking any versioned-surface element MUST be preceded by a deprecation
period of at least one MINOR release. A deprecation MUST: be marked DEPRECATED in the
specification and reference documentation with its replacement and target removal major;
be announced in the deprecating release's notes; and produce a runtime warning wherever
technically visible (CLI/log warnings on deprecated flags, config keys via Volume 10
validation, contract fields via SDK warnings) that names the replacement. Removal happens
only in a MAJOR release whose notes list every removal with its migration note. Silent
removal — in any release class — is a defect.

#### Motivation

SM-20 commits to a deprecation window; a window users cannot *see* (no warnings, no notes)
is not a window.

#### Actors

Surface owners deprecating; users and extension authors migrating.

#### Preconditions

The element has a replacement or an explicit no-replacement rationale.

#### Main flow

1. A MINOR release marks, announces, and starts warning.
2. At least one MINOR of window elapses.
3. The next MAJOR removes, with migration notes.

#### Alternative flows

- Security-forced immediate removal: permitted only when the element itself is the
  vulnerability; requires an ADR, a patch-class advisory, and prominent notes — the
  exception is recorded, never quiet.

#### Edge cases

- Deprecations spanning multiple majors (announced in 1.x, removed in 3.0) are permitted;
  the warning persists across the entire window.
- Pre-1.0: Beta-phase breaks follow Volume 1 discipline (migration notes, recorded
  decision) without the MAJOR vehicle.

#### Inputs

Deprecation declarations; replacement mappings.

#### Outputs

Marked specs, warnings, release-note sections, removal lists.

#### States

Not applicable.

#### Errors

Deprecated-usage warnings are warnings, never errors, until removal; after removal, normal
unknown-element errors of the owning surface apply.

#### Constraints

Warnings MUST be suppressible per Volume 8 verbosity conventions but on by default; the
window is measured in released MINORs, not calendar time.

#### Security

The security-forced path prevents the window from becoming an obligation to ship known
vulnerabilities.

#### Observability

Deprecated-usage warnings are structured (surface, element, replacement, removal target) so
fleets can inventory exposure before a major upgrade.

#### Performance

Warning emission MUST be once-per-process per element (no log flooding).

#### Compatibility

This requirement is the mechanism by which compatibility changes stay announced and
survivable.

#### Acceptance criteria

- Given a deprecated CLI flag, when it is used, then exactly one structured warning per
  process names flag, replacement, and removal target.
- Given a MAJOR release, when its notes are compared to contract-diff removals, then every
  removal appears with a migration note and had a ≥ 1-MINOR window (audit query over
  release history).
- Negative case: given a removal whose element was never marked, when the release audit
  runs, then publication is blocked.
- Given the security-forced path, when exercised, then an ADR reference exists in the
  advisory and notes.

#### Verification method

Release-audit deprecation ledger (mark → window → removal reconciliation); warning
emission tests per surface in Volume 13; documentation lint for DEPRECATED markers.

#### Traceability

SM-20; ADR-015; FR-REL-012; NFR-REL-003; Volume 10 config deprecation mechanics
(referenced).

### FR-REL-014 — Support windows, release branches, and backports

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Release engineering (Volume 14)
- Affected components: Updater, CI, release branches
- Dependencies: ADR-193, ADR-004, ADR-015; FR-REL-005
- Related risks: RISK-REL-004

#### Description

Support status MUST follow ADR-193: pre-v1, only the latest stable release is supported;
from v1, the latest minor of the current major receives full fixes and the last minor of
the previous major receives security and data-integrity fixes for 12 months from the
current major's first stable release. Maintenance work happens on `release/vX.Y` branches
(cherry-pick-only, cut at first rc) and ships as patch releases through the normal
FR-REL-001 pipeline. Backport qualification is ADR-193's closed list. Every release
declares `min_upgrade_from` such that direct updates are supported from any version of the
same major and from the previous major's last minor; older installations are routed
through stepping-stones (E-REL-009). The update check MUST surface end-of-support status
for the installed version as guidance, never as blocking.

#### Motivation

A support promise users can compute (version + date → status) is the difference between a
policy and a hope; bounded windows keep the promise honest at open-source maintainer scale.

#### Actors

Maintainers (backporting); users planning upgrades; the Updater surfacing status.

#### Preconditions

ADR-193 windows published; release branches exist per policy.

#### Main flow

1. A qualifying defect lands on trunk.
2. Cherry-pick PR labeled `backport` targets the supported release branch(es).
3. Patch releases tag from the branches; notes cite the advisory/defect.

#### Alternative flows

- Trunk-inapplicable fix (code no longer exists on trunk): a minimal branch-only patch is
  permitted with the trunk-equivalence rationale documented in the PR (ADR-193 risk note).

#### Edge cases

- A backport that cannot avoid a contract change does not ship — the affected line's
  remediation is an upgrade, stated in the advisory.
- Support windows never extend implicitly: a new patch on the previous major does not
  reset its 12-month clock.

#### Inputs

Defect classification; branch inventory; window dates.

#### Outputs

Patch releases on supported lines; support-status data consumed by FR-REL-005.

#### States

Release machine states for each patch release (chapter 05).

#### Errors

E-REL-009 (upgrade path); support status itself is informational.

#### Constraints

Cherry-pick-only branches (ADR-004); the backport label and branch protections are Volume
11 automation.

#### Security

Security backports follow the Volume 9 disclosure process; advisories name fixed versions
per supported line.

#### Observability

Update-check outcomes carry support status; release notes state each line's end-of-support
date.

#### Performance

Not applicable.

#### Compatibility

Patch releases never change contracts (by the backport qualification list).

#### Acceptance criteria

- Given an installed version and today's date, when `update check` runs, then the reported
  support status matches the published policy table (computable-status test).
- Given a security fix on trunk affecting a supported previous major, when the backport
  process completes, then a patch release exists on that line whose diff contains only the
  qualifying fix.
- Given an installation older than the target's `min_upgrade_from`, when update runs, then
  E-REL-009 names the stepping-stone chain and applying it succeeds end-to-end (upgrade
  matrix test).
- Negative case: given a feature PR labeled `backport`, when Volume 11 automation
  evaluates it, then it is rejected by the qualification check.

#### Verification method

Upgrade-path matrix in Volume 13 (same-major, previous-major-last-minor, stepping-stone
chains); release-audit checks on branch content vs qualification list; support-status
computation tests against fixture dates.

#### Traceability

ADR-193, ADR-004, ADR-015; E-REL-009; FR-REL-005; Volume 9 disclosure process; Volume 11
branch automation.

### FR-REL-015 — Changelog and release notes

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Release engineering (Volume 14)
- Affected components: CI, documentation
- Dependencies: ADR-013, ADR-015; FR-REL-002, FR-REL-012
- Related risks: RISK-REL-002

#### Description

Every release MUST publish notes generated from Conventional Commit history (ADR-015
sections by type/scope) and completed with the sections automation cannot derive. Mandatory
content: version, date, channel; highlights; **breaking changes with migration notes**
(MAJOR only, by FR-REL-012); deprecations announced and removals executed (FR-REL-013);
the versioned-surfaces table (FR-REL-012); database schema versions migrated to;
`min_upgrade_from` and support-window changes; signing status (Volume 1 signing viability
statement); verification instructions (FR-REL-002); and yank notices issued since the
previous release. The repository `CHANGELOG.md` (file conventions Volume 11) MUST be
updated by the same pipeline run, so notes and changelog never diverge.

#### Motivation

Release notes are the only channel through which versioning, deprecation, and support
policies reach users; unpopulated policy is unenforceable policy.

#### Actors

Release automation; maintainers writing migration notes; users and packagers reading.

#### Preconditions

FR-REL-001 pipeline run; commit history conventions held.

#### Main flow

1. Automation drafts sections from history.
2. Maintainers complete migration notes and highlights (required for MAJOR; optional
   otherwise).
3. Publication attaches notes to the GitHub Release and updates `CHANGELOG.md`.

#### Alternative flows

- Nightly releases: notes are fully automated (commit list + surface table); the manual
  sections are exempt.

#### Edge cases

- A release with zero user-visible changes (dependency-only patch) still publishes the
  mandatory sections; "no changes" is stated, not omitted.
- Yanked releases get a superseding note in the *next* release plus an edit to the yanked
  release's notes stating reason and remedy.

#### Inputs

Commit history; surface versions; policy data; maintainer-authored sections.

#### Outputs

Release notes; updated changelog.

#### States

Not applicable.

#### Errors

Missing mandatory sections block publication in the release audit (CI defect class).

#### Constraints

Notes are immutable after publication except yank annotations; corrections ship as new
releases (INV-REL-01 spirit).

#### Security

Advisory references use the Volume 9 disclosure identifiers; notes never contain
reproduction details ahead of coordinated disclosure timelines.

#### Observability

Notes are release artifacts referenced by Release rows (`notes_ref`, Volume 2).

#### Performance

Not applicable.

#### Compatibility

Note structure is stable so downstream packagers can parse sections mechanically.

#### Acceptance criteria

- Given any published release, when the audit checks its notes, then every mandatory
  section exists and the surface table matches the release's contract-diff data.
- Given a MAJOR release, when notes are checked, then every breaking change carries a
  migration note (blocking otherwise).
- Given a release publication, when `CHANGELOG.md` is compared, then it contains the same
  version section (divergence test).
- Negative case: given a draft missing the signing-status statement, when the audit runs,
  then publication is blocked.

#### Verification method

Release-audit note checks (section presence, surface-table cross-check, changelog
consistency); documentation lint; MAJOR-release checklist in Volume 11 automation.

#### Traceability

ADR-013, ADR-015; FR-REL-002 (verification instructions), FR-REL-012/013/014; Volume 1
signing viability; Volume 2 `notes_ref`.

## Non-functional requirements

### NFR-REL-003 — Public-contract stability

- Category: Compatibility
- Priority: P0
- Phase: v1
- Metric: (a) Breaking changes to the SM-20 public contract set (provider contract, tool contract, ARP, skill format, workflow format, configuration schema, CLI structured-output schema, event envelope, plus the Volume 14 distribution grammar) shipped outside a MAJOR release; (b) fraction of breaking changes preceded by a deprecation window ≥ 1 MINOR release
- Target: (a) 0; (b) 100%
- Minimum threshold: identical to target — SM-20 tolerates no deviation
- Measurement method: contract-diff tooling in CI comparing released contract schemas per surface; release audit reconciling diffs against bump class and the deprecation ledger (FR-REL-013)
- Test environment: CI on every release candidate and release
- Measurement frequency: every release; audited at phase gates
- Owner: Release engineering (Volume 14); NFR-ARCH-002 applies the same regime to the port layer
- Dependencies: FR-REL-012, FR-REL-013; ADR-015; SM-20
- Risks: RISK-REL-004
- Acceptance criteria: Every release's contract-diff report shows zero unannounced removals/re-signings across the measured set within a major line; any detected violation blocks publication (release audit is a required check), and phase-gate audits show 100% ledger compliance historically.

## Risks

### RISK-REL-004 — Support and compatibility obligations exceeding maintainer capacity

- Category: Process / sustainability
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: bounded windows and a two-line maximum (ADR-193); closed backport qualification enforced by automation; contract-diff and audits automated rather than manual (NFR-REL-003); windows extendable but never implicitly extended
- Detection: backport queue depth and patch-release latency tracked per Volume 15 project health metrics; audit failures trending in CI
- Owner: Release engineering (Volume 14) / Volume 15 governance
- Status: Open

Every promise in this chapter is recurring labor. The design bounds the obligation set
(two lines, closed backport list, automated audits) so the policy stays honest at
volunteer scale; the review conditions in ADR-193 name the signals for resizing it.
