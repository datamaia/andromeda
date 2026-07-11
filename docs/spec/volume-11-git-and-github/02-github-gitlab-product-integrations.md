# 02 — GitHub and GitLab as Product Integrations

This chapter specifies Andromeda-the-product's integrations with Git hosting services:
GitHub first, GitLab later. It is strictly distinct from GitHub as the platform on which
the `andromeda` repository is developed (chapters 03–07). Per ADR-147, hosting
integrations are delivered as the built-in `github` and `gitlab` tools of the Volume 6
catalog — this chapter owns their behavioral contract — over official, documented HTTP
APIs only (ADR-019 baseline). Unofficial mechanisms of any kind are OUT OF SCOPE
(Volume 1 provided constraints): no web automation, no undocumented endpoints, no session
or cookie reuse, and no programmatic access inferred from a user subscription.

## The change-request vocabulary

Pull requests (GitHub) and merge requests (GitLab) are the same engineering object. The
provider-neutral vocabulary below defines the shared operation surface both tools
implement; CLI/TUI flows and workflows use it so that switching hosting providers does
not change the user's flow. Provider-specific capabilities beyond this vocabulary are
declared per tool and never silently emulated (mirroring the Volume 5 capability-honesty
principle).

| Operation | Meaning | Side effect class |
|---|---|---|
| `change_request.create` | Open a PR/MR from a pushed branch, draft by default | Mutation |
| `change_request.update` | Edit title/body/labels/reviewers/draft state | Mutation |
| `change_request.list` | Enumerate PRs/MRs with filters | Read |
| `change_request.get` | One PR/MR with review and check state | Read |
| `change_request.comment` | Add a review comment | Mutation |
| `issue.get` / `issue.list` | Read issues for traceability context | Read |
| `release.list` | Read releases/tags for context | Read |

Operations that approve, merge, or close review objects are deliberately absent from the
product surface at every phase: **review conclusion is human work on the hosting
platform**. Andromeda prepares and describes change requests; it MUST NOT approve or
merge them (consistent with the human-review mandate of chapter 04 and PRD-005 autonomy
boundaries).

## Behavior common to both tools

1. **Permissions.** Every mutation binds `external_service_access` and `network`
   (Volume 9 names) at `domain` scope for the configured host; reads bind
   `external_service_access` at the same scope. Tool-level mediation (FR-TOOL-001)
   applies: schema-validated inputs, permission evaluation, audit, events.
2. **Authentication.** Tokens are acquired through AuthPort flows and stored behind
   SecretStorePort references (ADR-014); configuration references a named auth profile,
   never a literal token. The concrete grant types each provider officially documents
   for third-party applications are PENDING VALIDATION at implementation time against
   current official documentation (register entry); personal/fine-grained access tokens
   supplied by the user are the baseline that requires no provider-side application
   registration.
3. **Endpoints.** Hosts are configurable for self-managed installations
   (`api_base_url`); the tools MUST function against GitHub Enterprise Server and
   self-managed GitLab endpoints that expose the same documented APIs, and MUST surface
   version-related capability differences explicitly.
4. **Rate limits.** Documented rate-limit responses map to E-GIT-015 with the reset
   time when the provider supplies one; the tools respect reset times before retrying
   and never spin on 4xx.
5. **Data handling.** Fetched issue/PR content enters context assembly as provenance-
   tagged external content (Volume 7 trust rules; hosting content is untrusted input per
   the Volume 9 injection threat classes).
6. **Traceability.** Change requests created by Andromeda record the originating run and
   task in the PR/MR body's structured section (correlation IDs), satisfying SM-13
   attribution for hosting side effects.

## GitHub (`github` tool — Beta)

The `github` tool implements the vocabulary over the documented GitHub REST API, using
the documented GraphQL API where the REST surface lacks an operation. Defaults:
`api_base_url = "https://api.github.com"`, remote `origin`, draft-by-default creation.
PR creation flow: verify the branch is pushed (via the Git Engine — chapter 01 remains
the only git path), resolve the upstream repository and base branch, render the PR body
from the repository's PR template when present (never bypassing required template
sections), apply requested labels, and return the PR URL and number. The tool reads
check-run and review state so the TUI can present PR status; it MUST NOT create or
re-run checks.

## GitLab (`gitlab` tool — v1)

The `gitlab` tool implements the same vocabulary over the documented GitLab REST API for
gitlab.com and self-managed instances. MR-specific semantics (approval rules, merge
trains) are read-only context at v1; feature parity for each vocabulary operation against
the officially documented API of supported GitLab versions is PENDING VALIDATION at v1
implementation (register entry). Divergences are declared in the tool's capability
notes, never papered over.

## Requirements

### FR-GIT-009 — Hosting integration layer

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: Tool Runtime (tools) / Volume 11 (contract)
- Affected components: Tool Runtime, Authentication Layer, Secret Store, CLI, TUI
- Dependencies: ADR-147, ADR-019, ADR-014; FR-TOOL-001, FR-GIT-001
- Related risks: RISK-GIT-004

#### Description

Andromeda MUST provide Git hosting integrations exclusively as the built-in `github`
(Beta) and `gitlab` (v1) tools implementing the change-request vocabulary of this chapter
over official, documented APIs, with permissions, authentication, endpoint
configurability, and rate-limit behavior as specified above. The tools MUST NOT approve
or merge change requests at any phase.

#### Motivation

PR/MR preparation is the last mile of the engineering loop the product automates; doing
it through mediated tools keeps hosting side effects inside the permission and audit
model, and the neutral vocabulary prevents hosting lock-in in user flows.

#### Actors

Agents, workflows (SDD release preparation), CLI/TUI users.

#### Preconditions

Configured hosting profile with a stored credential reference; network permission
available; repository with a configured remote on the target host.

#### Main flow

1. A consumer invokes a vocabulary operation through the Tool Runtime.
2. Input schema validation; permission evaluation (`external_service_access`,
   `network`).
3. The tool resolves the credential reference, calls the documented endpoint, and maps
   the response to the neutral result shape.
4. Records, events, and audit entries are written with correlation IDs.

#### Alternative flows

- Missing credential → E-GIT-014 with the auth-profile remediation path (no interactive
  token capture inside the tool call).
- Provider capability absent for a requested option → declared capability error, no
  emulation.

#### Edge cases

- Forked-repository flows: the tool resolves head/base across forks explicitly and
  refuses ambiguous resolution rather than guessing.
- Repository without the remote configured → E-GIT-011 semantics surfaced from the Git
  Engine before any network call.
- Enterprise hosts with reduced API surfaces: capability differences reported per
  operation.

#### Inputs

Vocabulary operation parameters (JSON per tool schema); `[github]` /
`[git.hosting.gitlab]` configuration; auth profile references.

#### Outputs

Neutral result shapes (change-request identity, URL, state); events; audit records.

#### States

Change requests live on the hosting platform; Andromeda records references, not a
mirrored state machine.

#### Errors

E-GIT-013, E-GIT-014, E-GIT-015; input validation errors per the tool contract.

#### Constraints

Official APIs only; no approval/merge operations; drafts by default; provider-specific
logic confined to the respective tool.

#### Security

Tokens behind Secret Store references; minimal documented scopes per operation class
listed in tool documentation; hosting content treated as untrusted input; every mutation
audited.

#### Observability

`git.pull_request.opened`, `git.pull_request.updated` events with host, repository, and
change-request number; request spans with rate-limit headroom where the API reports it.

#### Performance

Network-bound; per-call timeout from the tool declaration; no polling loops — status is
read on demand.

#### Compatibility

github.com, GitHub Enterprise Server, gitlab.com, self-managed GitLab per their
documented APIs; capability differences declared. Provider API version pinning is
recorded in each tool's declaration.

#### Acceptance criteria

- Given a pushed branch and a granted permission set, when `change_request.create`
  runs, then a draft PR exists with the rendered template body, the correlation IDs in
  its structured section, and the returned URL resolves.
- Given a revoked token, when any operation runs, then E-GIT-014 is returned, no retry
  storm occurs, and the audit record captures the denial.
- Given a rate-limited response with a reset time, when the tool retries, then the retry
  waits for the reset and the second attempt is recorded against the same correlation
  ID.
- Negative: a request to merge or approve a change request fails schema validation —
  the operations do not exist in the tool surface.
- Permission: without `external_service_access` for the host domain, no network call is
  made.

#### Verification method

Contract tests against recorded official-API fixtures; permission-enforcement tests;
enterprise-endpoint configuration tests; audit-chain resolution tests (SM-13).

#### Traceability

PRD-004, PRD-005, PRD-006; ADR-147; FR-TOOL-001; chapter 04 human-review mandate.

### FR-GIT-010 — Change-request preparation flow

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: Beta
- Source: Derived
- Owner: Tool Runtime (tools) / Volume 11 (contract)
- Affected components: Tool Runtime, Git Engine, Workflow Engine, CLI, TUI
- Dependencies: FR-GIT-009, FR-GIT-006, FR-GIT-004
- Related risks: RISK-GIT-004

#### Description

The end-to-end preparation flow — verify branch state, push via the Git Engine, create
the draft change request, populate template and metadata, link the tracking issue —
MUST be available as a composed flow usable by agents and workflows, with each
side-effecting step individually permissioned and the whole flow stopping at draft
creation. Publishing a draft for review is a human action; Andromeda MAY do it only on
explicit per-invocation instruction.

#### Motivation

This is the concrete product moment where the traceability chain (chapter 07) begins on
the hosting side; making the safe path the easy path is how template and linkage
discipline actually happens.

#### Actors

Agents completing implementation tasks; the SDD workflow's release-preparation stage;
users.

#### Preconditions

FR-GIT-009 configured; branch exists locally with commits.

#### Main flow

1. Verify the working branch is committed and pushed (push through FR-GIT-006 with its
   gates).
2. Resolve base branch and repository (fork-aware).
3. Create the draft change request with template body, issue linkage, labels, and
   correlation IDs.
4. Report the URL; emit events.

#### Alternative flows

- Unpushed commits and no push permission → the flow stops before any hosting call with
  the combined remediation (grant or push manually).
- Existing open change request for the branch → the flow switches to
  `change_request.update` semantics instead of duplicating.

#### Edge cases

- Base branch protected or nonexistent on the host → provider error surfaced as
  E-GIT-013 with the provider's reason.
- Template with required sections the agent cannot fill → the flow leaves explicit
  section markers and keeps the request in draft (never deletes required sections).

#### Inputs

Branch, base, title (Conventional Commits format when targeting the `andromeda`
repository — chapter 04), body inputs, issue reference, labels.

#### Outputs

Draft change-request identity and URL; events; audit records.

#### States

Hosting-side state; locally only records.

#### Errors

E-GIT-009 (push leg), E-GIT-013, E-GIT-014, E-GIT-015.

#### Constraints

Stops at draft; no approval/merge; one change request per branch (update, not
duplicate).

#### Security

Same as FR-GIT-009; the push leg carries its own `git_mutation`+`network` gates.

#### Observability

`git.pull_request.opened` / `git.pull_request.updated`; flow span linking git and
hosting legs under one correlation ID.

#### Performance

Two network legs (push, API call) plus reads; no polling.

#### Compatibility

Both providers via the vocabulary; fork-aware resolution on both.

#### Acceptance criteria

- Given a completed task on a feature branch, when the preparation flow runs with
  grants, then a draft PR exists whose body links the tracking issue and whose head is
  the pushed branch tip.
- Given an existing open PR for the branch, when the flow re-runs, then the PR is
  updated and no duplicate is created.
- Negative: with the branch unpushed and push denied, no hosting API call is observed.
- Observability: the flow's git and hosting spans share the run's correlation ID.

#### Verification method

End-to-end tests against fixture hosting doubles; duplicate-prevention tests;
permission matrix tests; workflow integration test in the SDD release-preparation
stage.

#### Traceability

FR-GIT-009; Volume 4 SDD release-preparation stage; chapter 04 PR process; SM-13.

## Risks

### RISK-GIT-004 — Hosting API drift, deprecation, or policy change

- Category: External dependency
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: official APIs only with per-release contract fixtures; declared per-tool
  capability notes; API version pinning in tool declarations; ADR-147 review conditions;
  degradation is explicit and user-visible, never silent
- Detection: contract-fixture failures in CI; E-GIT-013/E-GIT-015 telemetry trends;
  provider deprecation announcements tracked at phase gates
- Owner: Tool Runtime (tools) / Volume 11 (contract)
- Status: Open

## Errors

### E-GIT-013 — Hosting service request failed

- Category: External service
- Severity: Error
- User message: "<host> rejected or failed the request: <provider reason>."
- Technical message: operation, endpoint class (not full URL), HTTP status, provider
  error code/message, request ID header when present
- Cause: provider-side rejection (validation, permissions, state) or 5xx failure
- Safe-to-log data: host, operation, status, provider error code, provider request ID
- Recoverability: recoverable for state/validation causes; retryable for 5xx
- Retry policy: retry with backoff for 5xx and network failures; never for 4xx other
  than 429 (see E-GIT-015)
- Recommended action: read the provider reason; verify repository permissions of the
  token
- Exit-code mapping: 1
- HTTP mapping: pass-through of the provider status in structured output
- Telemetry event: `git.operation.refused`
- Security implications: response bodies pass redaction before logging

### E-GIT-014 — Hosting authentication failed

- Category: Authentication
- Severity: Error
- User message: "Authentication to <host> failed. The stored credential is missing,
  expired, or lacks required scopes."
- Technical message: auth profile name, credential reference presence, HTTP status,
  provider auth error class, scopes required by the operation class
- Cause: absent/expired/revoked token or insufficient scopes
- Safe-to-log data: host, profile name, failure class (never token material)
- Recoverability: recoverable via credential update/rotation (AuthPort flows)
- Retry policy: none automatic
- Recommended action: re-authenticate the profile or issue a token with the documented
  scopes
- Exit-code mapping: 4
- HTTP mapping: 401/403 pass-through in structured output
- Telemetry event: `git.operation.refused`
- Security implications: token material never appears in messages, logs, or events
  (Volume 9 redaction)

### E-GIT-015 — Hosting rate limited

- Category: External service
- Severity: Warning
- User message: "<host> rate limit reached; retrying after the reported reset."
- Technical message: operation, limit class, reset time when reported, remaining quota
  headers
- Cause: documented provider rate limiting
- Safe-to-log data: host, limit class, reset time
- Recoverability: recoverable by waiting
- Retry policy: single retry after the reported reset; otherwise surface to the caller
  with the reset time
- Recommended action: reduce request frequency; for CI use, stagger operations
- Exit-code mapping: 1
- HTTP mapping: 429 pass-through in structured output
- Telemetry event: `git.operation.refused`
- Security implications: none

## Configuration

Keys minted in the `[github]` table and the `[git.hosting.gitlab]` sub-table (the
primary integration gets the short table named in the Volume 0 ownership map; additional
providers nest under `[git.hosting.*]`):

```toml
[github]
enabled = false
api_base_url = "https://api.github.com"
default_remote = "origin"
auth_profile = ""            # named AuthPort profile; never a literal token
draft_by_default = true

[git.hosting.gitlab]
enabled = false
api_base_url = "https://gitlab.com/api/v4"
default_remote = "origin"
auth_profile = ""
draft_by_default = true
```

## Events

Minted here (envelope per Volume 10): `git.pull_request.opened`,
`git.pull_request.updated`. Payloads carry host, repository identifier, change-request
number/URL, draft state, and correlation IDs; bodies and diffs never appear in payloads.
