# 04 — Pull Requests

Every change to the `andromeda` repository lands through a pull request into `main` (or a
`release/*` branch for backports). This chapter defines the PR contract: template,
size, description duties, review rules — human review is mandatory, self-approval is
impossible — merge mechanics, and the provenance rules for AI-generated changes.
Provenance is recorded **at the PR level through labels and template declarations;
commit messages carry change information only** (ADR-015). The validators referenced
here are specified as automation in chapter [07](07-traceability-automation.md).

## PR template

`.github/PULL_REQUEST_TEMPLATE.md` contains the following mandatory sections; the
traceability validator fails PRs with required sections missing or left as placeholders:

```text
## Summary
What changes and why, in reviewable terms.

## Linked issue
Closes #<issue>          <!-- closing keyword mandatory (FR-GH-001) -->

## Requirements
IDs realized or affected (from the linked issue), or "process-only".

## Changes
- Bullet list of user-visible and contract-visible changes.

## Risks and compatibility
Breaking changes (label + BREAKING CHANGE footer), config changes, security-relevant
changes, migration or deprecation notes; "none" is an acceptable, explicit answer.

## Tests and evidence
What was added/updated; how reviewers can verify. TUI/CLI changes include terminal
captures where rendering is affected.

## Documentation
Docs/spec updates in this PR, or why none are needed.

## AI provenance
[ ] ai/none — authored without AI assistance
[ ] ai/assisted — human-authored with AI assistance
[ ] ai/generated — substantially AI-generated under human direction
```

## PR rules

| Rule | Requirement |
|---|---|
| Linked issue | Mandatory closing-keyword link to exactly one primary issue (additional `Refs` links allowed) |
| Title | Conventional Commits format with ADR-015 scopes — it becomes the squash-commit subject |
| Size (recommended) | ≤ 400 changed lines (additions + deletions), excluding generated code, lockfiles, golden files, and vendored data |
| Size (maximum) | ≤ 1500 changed lines with the same exclusions; larger PRs fail E-GH-005 and MUST be split; the `size-exempt` label (maintainers only, justification required in the PR body) is the audited escape hatch for atomic changes that cannot be split (large refactors, generated schema syncs) |
| Draft state | PRs open as drafts by default; marking ready-for-review asserts template completeness |
| Reviews | ≥ 1 approving human review from a maintainer who is not the PR author; code-owner review additionally required on owned paths (chapter 03) |
| Self-approval | Impossible: platform rules prevent authors approving their own PRs; approvals from bot identities or agent-operated accounts do not count toward the requirement |
| Conversations | All review conversations resolved before merge (platform setting) |
| Checks | All required checks green (chapter 06 list); stale approvals dismissed on new pushes |
| Merge method | Squash merge only; merge commits and rebase-merge disabled (ADR-015 linear history) |
| Squash message | Subject = PR title; body = PR summary plus generated trailer `Refs: #<issue>`; validated by the commit-message check before merge |
| Commit signing | Commits SHOULD be signed; the platform's verified-signature display is informational. Requiring signatures repo-wide is deliberately not enforced pre-v1 to keep the contribution barrier low; revisit at the v1 phase gate |
| Sensitive code | PRs touching CODEOWNERS-mandatory paths (ports, SDK, security subtrees, workflows, release config) additionally require the owning group's review and the `security-review` label when Volume 9-relevant |
| Rebase updates | Contributors keep branches current by rebase; reviewers re-review after force pushes (stale-approval dismissal enforces this) |

## Human review and AI provenance

1. **Human review is mandatory for every PR, without exception** — including
   maintainers' own PRs, dependency bumps, and AI-generated changes. Automation can
   block a merge; only a human approval can enable one.
2. **AI-generated changes are labeled at the PR level.** Exactly one `ai/*` label —
   `ai/none`, `ai/assisted`, or `ai/generated` — MUST be present at merge time. The
   template checkbox drives initial labeling; the provenance check (E-GH-006) fails on
   missing, multiple, or checkbox-label mismatches.
3. **Commit messages carry change information only.** No `Co-Authored-By` trailers, no
   "Generated with", no tool or vendor attribution, links, or badges (ADR-015). The
   versioned `commit-msg` hook (`scripts/hooks/`) enforces this locally; the CI
   commit-message check enforces it authoritatively on every PR head commit and on the
   squash message (E-GH-001). Provenance lives in labels and the PR body — grep the
   history for *what changed*, query the platform for *how it was produced*.
4. **Agent-operated PRs** follow the identical process. An agent MAY open, update, and
   respond to review on a PR; it MUST NOT approve, dismiss reviews, or merge. Merging a
   PR whose head commits were pushed by an agent requires the same non-author human
   approval as any PR.

## Requirements

### FR-GH-004 — Pull request process

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Development process (Volume 11)
- Affected components: None at runtime; enforced by platform configuration and CI
- Dependencies: ADR-004, ADR-015, ADR-148; FR-GH-003, FR-GH-001
- Related risks: RISK-GH-002

#### Description

Every repository change MUST land through a PR satisfying this chapter: mandatory
template sections, one linked issue via closing keyword, Conventional-Commits title,
size limits with the audited exemption, mandatory non-author human review plus
code-owner review on owned paths, resolved conversations, green required checks, and
squash-only merge with the validated message shape. Draft PRs are exempt from review
and provenance checks but not from functional checks.

#### Motivation

The PR is the join point of the traceability chain (ADR-148) and the human-control
point mandated for a project where a large share of changes is AI-produced.

#### Actors

Contributors (human and agent), maintainers, CODEOWNERS groups, CI.

#### Preconditions

FR-GH-002/FR-GH-003 in place; validators deployed.

#### Main flow

1. Contributor opens a draft PR from a grammar-compliant branch with the template
   filled.
2. Checks run; contributor iterates; PR marked ready.
3. Reviews complete (non-author human; code owners where owned).
4. Squash merge produces the validated mainline commit; head branch deletes.

#### Alternative flows

- Required section legitimately empty → explicit "none" with justification (validators
  accept explicit negatives, never blanks).
- Size over recommended but under maximum → warning annotation; over maximum →
  E-GH-005 failure or maintainer `size-exempt` with justification.

#### Edge cases

- Backport PRs to `release/*`: same rules; linked issue is the original issue;
  `backport` label required.
- Reverts: `revert:` typed title, linked to the incident/tracking issue; template
  "Changes" section states what is reverted and why.
- Co-authored human work: multiple humans MAY be credited via the platform's
  co-contributor mechanisms on the PR; commit messages still carry no `Co-Authored-By`
  trailers (uniform rule, ADR-015).

#### Inputs

PR metadata, template body, labels, reviews, check results.

#### Outputs

Squash commits on `main` with validated messages; platform records forming the
traceability chain.

#### States

Platform PR states (draft, ready, merged, closed); no product state machine.

#### Errors

E-GH-001 (message policy), E-GH-003 (missing linkage — chapter 07), E-GH-005 (size),
E-GH-006 (provenance).

#### Constraints

Squash-only; stale-approval dismissal on; conversations-resolved on; admin-inclusive
enforcement.

#### Security

Non-author human review is the control against both malicious and defective automation;
sensitive paths get owning-group review; fork PRs follow ADR-149 isolation.

#### Observability

Check annotations per violation; nightly audit counts merges, exemption-label usage,
and review-rule anomalies; chain report per release (chapter 07).

#### Performance

Validator latency within the FR-GH-009 PR-feedback budget.

#### Compatibility

GitHub primitives (required reviews, code owners, squash merge); portable to platforms
with equivalent primitives per ADR-148.

#### Acceptance criteria

- Given a PR authored by maintainer M, when M attempts to approve it, then the platform
  refuses; when only agent accounts approve, then the review requirement remains
  unmet.
- Given a ready PR with an unfilled "Requirements" section, when validators run, then
  the PR fails with the section named.
- Given a PR with 2000 changed lines of hand-written code and no exemption, when
  validators run, then E-GH-005 fails the check; with `size-exempt` applied by a
  maintainer and a justification present, then the check passes and the exemption is
  logged in the nightly audit.
- Given all rules satisfied, when merged, then `main`'s new commit subject equals the
  PR title, its body contains `Refs: #<issue>`, and the head branch no longer exists.
- Negative: a merge attempt with an unresolved conversation is rejected by the
  platform.

#### Verification method

Platform-configuration audit at phase gates; validator unit/integration tests (chapter
07); process assertions in the nightly audit; release-time chain report.

#### Traceability

PRD-013; ADR-004, ADR-015, ADR-148; FR-GH-001; Volume 0 chapter 10 change procedure
(spec PRs).

### FR-GH-005 — AI provenance labeling and commit-message enforcement

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Development process (Volume 11)
- Affected components: None at runtime; commit-msg hook + CI checks
- Dependencies: ADR-015, ADR-148; FR-GH-004
- Related risks: RISK-GH-002

#### Description

Every PR MUST carry exactly one `ai/*` provenance label consistent with its template
declaration at merge time. Every commit message on a PR head and every squash message
MUST satisfy Conventional Commits with the ADR-015 closed scope list and MUST contain
no AI/vendor attribution content (prohibited-pattern set: `Co-Authored-By:` trailers,
"Generated with", "Co-authored by <tool>", tool URLs and badge markup in message
bodies). Enforcement is two-layer: the versioned `commit-msg` hook for local feedback
and the authoritative CI check (E-GH-001, E-GH-006).

#### Motivation

Honest provenance without message pollution: history stays a clean record of change
intent that release automation parses (ADR-015), while consumers who need provenance
query PR labels via official APIs.

#### Actors

Contributors, agents, CI, release automation.

#### Preconditions

Hook installed locally (CONTRIBUTING.md path); CI checks required.

#### Main flow

1. Contributor commits; the hook validates format, scope, and the prohibition set.
2. PR opens with the provenance checkbox; labeling applies the matching `ai/*` label.
3. CI re-validates all head commits, the projected squash message, and label
   consistency.

#### Alternative flows

- Hook not installed → CI is authoritative; the failure message includes hook
  installation instructions.
- Provenance changes during a PR's life (human rewrite of agent output) → label edited;
  the check re-validates consistency with the updated checkbox.

#### Edge cases

- Revert and merge-queue-generated messages follow the same grammar (`revert:` type;
  no exemption from the prohibition set).
- Messages quoting prohibited strings in code blocks (e.g., docs about this rule):
  the check applies to trailer/footer positions and literal trailer syntax, not to
  quoted prose — documented pattern semantics keep false positives near zero, and the
  `policy-check` override path (maintainer label + audit) covers residual cases.
- Dependabot commits: tool-authored but vendor-attribution-free by configuration;
  they carry `ai/none` (they are template automation, not AI generation).

#### Inputs

Commit messages, PR labels, template checkbox state.

#### Outputs

Pass/fail check results with per-commit annotations; consistent provenance labels at
merge.

#### States

Not applicable.

#### Errors

E-GH-001, E-GH-006.

#### Constraints

Exactly one `ai/*` label; closed scope list; prohibition set versioned with the
validator and changed only via PR to this volume.

#### Security

Prevents attribution-laundering of automated changes past review (provenance is
tamper-evident in the audit trail); no secrets involved.

#### Observability

Provenance distribution (share of `ai/*` labels over merged PRs) reported by the
nightly audit — an input to Volume 15's sustainability metrics.

#### Performance

Regex/grammar checks; negligible against the FR-GH-009 budget.

#### Compatibility

Hook is a portable script (POSIX sh + the repo's pinned tooling); CI check is the
authoritative layer so unhookable environments lose nothing.

#### Acceptance criteria

- Given a commit message `feat(provider): add retry budget` with a body containing
  `Co-Authored-By: SomeBot <bot@vendor.example>`, when the check runs, then E-GH-001
  fails naming the prohibited trailer and the commit.
- Given a PR whose checkbox says `ai/generated` but whose label is `ai/none`, when the
  provenance check runs, then E-GH-006 fails with the mismatch.
- Given a compliant PR, when merged, then exactly one `ai/*` label exists on the merged
  PR and the squash message contains no prohibited pattern.
- Negative: a scope outside the ADR-015 closed list (`feat(gui): …`) fails with the
  allowed-scope list in the annotation.

#### Verification method

Validator test suite with a golden corpus of valid/invalid messages (including quoted-
prose false-positive cases); hook/CI parity tests (same corpus, same verdicts);
platform-configuration audit confirming the checks are required.

#### Traceability

ADR-015 (provided constraint), ADR-148; FR-GH-004; chapter 07 validator family.

## Errors

### E-GH-001 — Commit message policy violation

- Category: Process validation
- Severity: Error
- User message: "Commit <short-hash> violates the commit-message policy: <finding>."
- Technical message: commit hash, rule violated (format, scope, breaking-change marker,
  prohibited attribution pattern), offending line reference
- Cause: message not in Conventional Commits form, scope outside the closed list, or
  prohibited attribution content present
- Safe-to-log data: commit hash, rule ID, line number (message content only in check
  annotations, which are public in the repository context)
- Recoverability: recoverable by rewording (`git commit --amend` / rebase) before merge
- Retry policy: not applicable (deterministic validation)
- Recommended action: reword the message per CONTRIBUTING.md; install the local hook
- Exit-code mapping: 1 (validator process)
- HTTP mapping: not applicable
- Telemetry event: none (CI tooling; results are check annotations)
- Security implications: blocks attribution-content injection into permanent history

### E-GH-005 — Pull request size limit exceeded

- Category: Process validation
- Severity: Error
- User message: "This PR changes <n> lines (limit 1500 after exclusions). Split it, or
  a maintainer may apply size-exempt with justification."
- Technical message: counted lines, exclusion classes applied, per-file breakdown
- Cause: diff size beyond the hard limit without an exemption
- Safe-to-log data: counts and file classes
- Recoverability: recoverable by splitting or exemption
- Retry policy: not applicable
- Recommended action: split along reviewable seams; keep refactors separate from
  behavior changes
- Exit-code mapping: 1 (validator process)
- HTTP mapping: not applicable
- Telemetry event: none (CI tooling)
- Security implications: bounds review-attention dilution — oversized PRs are where
  defects and injected changes hide

### E-GH-006 — AI provenance label missing or inconsistent

- Category: Process validation
- Severity: Error
- User message: "Exactly one ai/* label consistent with the PR's provenance declaration
  is required before merge."
- Technical message: labels found, checkbox state, mismatch kind
- Cause: no `ai/*` label, multiple labels, or label/declaration mismatch
- Safe-to-log data: label names, checkbox value
- Recoverability: recoverable by correcting label or declaration
- Retry policy: not applicable
- Recommended action: set the checkbox truthfully; maintainers adjust labels on
  evidence
- Exit-code mapping: 1 (validator process)
- HTTP mapping: not applicable
- Telemetry event: none (CI tooling)
- Security implications: preserves the integrity of the provenance record that ADR-015
  moves out of commit messages
