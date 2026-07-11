# 99 — Volume 11 Register

Machine-parseable register of everything Volume 11 minted, per Volume 0 chapters 02 and
03. Merged into the Volume 0 registers at consolidation.

## Requirements index

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-GIT-001 | Git Engine | MVP | GitPort contract suite incl. cancellation; equivalence suite; permission-enforcement tests; Tier 1 matrix |
| FR-GIT-002 | Repository discovery and version gating | MVP | Fixture-tree unit tests (worktrees, submodules, bare, symlinks); version-gate tests |
| FR-GIT-003 | Read-query fidelity: status, diff, log, show, blame | MVP | Equivalence suite; streaming/cancellation contract tests; parser fuzzing |
| FR-GIT-004 | Staging and commit creation | MVP | Hook fixtures; message byte-equivalence; attribution-refusal tests; permission tests |
| FR-GIT-005 | Branches, tags, and worktrees | MVP | Dirty-tree/unmerged fixtures; safety-ref restore tests; permission tests |
| FR-GIT-006 | Remote operations: fetch, pull, push | Beta | Local fixture remotes; lease-race tests; permission matrix; timeout injection |
| FR-GIT-007 | History modification, conflicts, and recovery | Beta | Conflict-matrix tests; published-rewrite classification property tests; offline recovery round-trips |
| FR-GIT-008 | Repository feature passthrough (hooks, ignore, signing, submodules, sparse checkout, LFS) | MVP | Feature fixtures; signed-commit verification; audit assertions |
| FR-GIT-009 | Hosting integration layer | Beta | Contract tests over official-API fixtures; permission enforcement; enterprise-endpoint tests |
| FR-GIT-010 | Change-request preparation flow | Beta | E2E over hosting doubles; duplicate-prevention; SDD stage integration |
| FR-GH-001 | Traceability automation | MVP | Validator golden fixtures; fixture-repository chain E2E incl. release report; phase-gate sampling |
| FR-GH-002 | Repository structure | Core | Structure-check CI job; CODEOWNERS/branch-protection audit; release qualification audit |
| FR-GH-003 | Branching rules and branch hygiene | Core | Protection-configuration audit; validator tests; nightly audit assertions |
| FR-GH-004 | Pull request process | Core | Platform-configuration audit; validator tests; nightly audit; chain report |
| FR-GH-005 | AI provenance labeling and commit-message enforcement | Core | Golden message corpus (hook/CI parity); label-consistency tests; configuration audit |
| FR-GH-006 | Issue taxonomy and intake forms | Core | Form-schema validation; submission walkthrough at phase gates; audit assertions |
| FR-GH-007 | Label taxonomy as synchronized data | Core | Sync-tool tests against fixture state; workflow permission audit |
| FR-GH-008 | Projects, milestones, and roadmap operation | MVP | Automation integration tests on fixture project; nightly audit; phase-gate review |
| FR-GH-009 | Quality pipelines and required checks | Core | Workflow integration tests; protection audit; NFR-GH-002 measurement; policy self-test |
| FR-GH-010 | Security scanning pipelines | MVP | Fixture-based scan tests; release-gate integration test; weekly report audit |
| FR-GH-011 | Release, upgrade, and documentation pipelines | MVP | Snapshot dry runs; rc rehearsals; artifact verification commands; environment audit |
| FR-GH-012 | Workflow security posture enforcement | Core | Rule-set unit tests over fixture workflows; self-application; phase-gate ADR-149 comparison |
| NFR-GIT-001 | Git output fidelity | MVP | Equivalence suite, 0 divergences, per merge and per release across the git version matrix |
| NFR-GIT-002 | Destructive-operation recoverability | Beta | Instrumented destructive-operation campaign with offline restoration, 100% |
| NFR-GH-001 | Development traceability completeness | MVP | Nightly chain audit; release chain reports; phase-gate trend review |
| NFR-GH-002 | Pull-request feedback latency | MVP | Check-run timestamp rollups, p85 targets, monthly and phase-gate review |
| RISK-GIT-001 | Git version and output-format drift | — | Equivalence-suite failures; E-GIT-005 telemetry |
| RISK-GIT-002 | User data loss through destructive operations | — | Class D audit records; NFR-GIT-002 campaign |
| RISK-GIT-003 | Repository hooks as arbitrary code execution | — | Hook-execution audit trail; Volume 9 monitoring |
| RISK-GIT-004 | Hosting API drift, deprecation, or policy change | — | Contract-fixture failures; error telemetry trends |
| RISK-GH-001 | CI or release workflow compromise | — | Policy-check findings; provenance verification; platform audit log review |
| RISK-GH-002 | Traceability erosion through process bypass or decay | — | Nightly audit orphan/override reports; NFR-GH-001 trend |
| RISK-GH-003 | Fork pull-request secret or privilege exposure | — | Policy check on workflow changes; nightly fork-trigger assertion |

## ADRs minted

Volume 11's block allocation is ADR-145–159; 145–149 are used. ADR numbers 150–159 are
unused and remain permanent gaps per Volume 0 chapter 03.

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-145](../annexes/adr/ADR-145.md) | Git operation catalog with phased, additive GitPort exposure | Accepted | Full operation catalog specified now; frozen 15 methods at MVP; Beta and v1 additive amendment batches via the change procedure; no passthrough or terminal side channel |
| [ADR-146](../annexes/adr/ADR-146.md) | Destructive Git operation policy: dual gate plus safety refs | Accepted | `git_mutation` permission plus operation-specific confirmation for destructive class; pre-operation safety refs under `refs/andromeda/safety/` with 30-day retention; lease-only force |
| [ADR-147](../annexes/adr/ADR-147.md) | Git hosting integrations as built-in tools over official APIs | Accepted | `github` (Beta) and `gitlab` (v1) built-in tools implement a provider-neutral change-request vocabulary over documented REST/GraphQL APIs; no approval/merge operations |
| [ADR-148](../annexes/adr/ADR-148.md) | Traceability enforcement: platform-native chain with required validators | Accepted | Each chain link recorded in one platform-native field and checked by one required validator; squash commit joins commit↔PR; nightly audit and per-release chain report |
| [ADR-149](../annexes/adr/ADR-149.md) | CI workflow security posture: least privilege, SHA pinning, untrusted forks | Accepted | Explicit minimal permissions per workflow, full-SHA action pins, fork PRs isolated from secrets/privileged contexts, environment-gated release jobs, hosted ephemeral runners only |

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-GIT-001 | Git binary not found | 3 |
| E-GIT-002 | Git version below minimum | 3 |
| E-GIT-003 | Not a Git repository | 1 |
| E-GIT-004 | Git subprocess failed | 1 |
| E-GIT-005 | Unparseable git output | 1 |
| E-GIT-006 | Operation stopped on conflicts | 1 |
| E-GIT-007 | Destructive operation refused without confirmation | 5 |
| E-GIT-008 | Git mutation permission denied | 5 |
| E-GIT-009 | Remote operation failed | 1 (4 on authentication cause) |
| E-GIT-010 | Git operation timed out or was cancelled | 8 |
| E-GIT-011 | Repository state conflict | 1 |
| E-GIT-012 | Protected branch refusal | 5 |
| E-GIT-013 | Hosting service request failed | 1 |
| E-GIT-014 | Hosting authentication failed | 4 |
| E-GIT-015 | Hosting rate limited | 1 |
| E-GH-001 | Commit message policy violation | 1 (validator process) |
| E-GH-002 | Branch naming violation | 1 (validator process) |
| E-GH-003 | Missing or unresolvable linkage | 1 (validator process) |
| E-GH-004 | Traceability chain validation failed | 1 (validator process; blocks release) |
| E-GH-005 | Pull request size limit exceeded | 1 (validator process) |
| E-GH-006 | AI provenance label missing or inconsistent | 1 (validator process) |

## Events minted

Git Engine family (envelope per Volume 10): `git.repository.discovered`,
`git.commit.created`, `git.branch.created`, `git.branch.switched`, `git.tag.created`,
`git.patch.applied`, `git.worktree.added`, `git.worktree.removed`, `git.remote.fetched`,
`git.remote.pulled`, `git.remote.pushed`, `git.merge.completed`, `git.rebase.completed`,
`git.history.rewritten`, `git.conflict.detected`, `git.conflict.resolved`,
`git.safety_ref.created`, `git.operation.refused`.

Hosting integration family: `git.pull_request.opened`, `git.pull_request.updated`.

The GH area mints no runtime events: development-process automation reports through CI
check results and audit-filed issues, not through the product event bus.

## Config keys minted

`[git]` table: `git.binary_path`, `git.operation_timeout_seconds`,
`git.remote_timeout_seconds`, `git.allow_force_push`, `git.protected_branches`,
`git.sign_commits`, `git.hooks.run`, `git.safety_refs.enabled`,
`git.safety_refs.retention_days`, `git.submodules.recurse`, `git.blame.ignore_revs`.

`[github]` table: `github.enabled`, `github.api_base_url`, `github.default_remote`,
`github.auth_profile`, `github.draft_by_default`.

`[git.hosting.gitlab]` sub-table: `enabled`, `api_base_url`, `default_remote`,
`auth_profile`, `draft_by_default`.

Schema, precedence, and validation are Volume 10's; key content ownership per Volume 0
chapter 03.

## Glossary additions

| Term | One-line meaning |
|---|---|
| Safety ref | A local-only ref under `refs/andromeda/safety/<ulid>` recorded before every destructive Git operation, restorable offline within its retention window (ADR-146). |
| Operation class | The Git operation catalog's classification — read-only, additive, history, destructive, remote — driving permission and confirmation requirements (Volume 11, chapter 01). |
| Change-request vocabulary | The provider-neutral operation set over pull requests and merge requests implemented by the `github` and `gitlab` tools (Volume 11, chapter 02; ADR-147). |
| Traceability chain report | The per-release artifact resolving every commit in the release range to its PR, issue, and requirement IDs (Volume 11, chapter 07). |
| Workflow policy check | The required CI check mechanically enforcing the ADR-149 workflow security posture (Volume 11, chapter 06). |
| Provenance label | The mandatory PR-level `ai/none` / `ai/assisted` / `ai/generated` label recording how a change was produced (Volume 11, chapter 04; ADR-015). |

## Assumptions

Local list per Volume 0 chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | GitHub's platform features this volume builds on (Actions with environments and OIDC, issue forms, Projects v2, squash-merge message control, required checks, code-owner review) remain available to open-source projects on workable terms | Phase-gate reviews; mirrors the Volume 1 assumption on GitHub availability | ADR-148's reversal plan (platform-portable chain) and ADR-149 successor for the new CI; migration is a MAJOR document change |
| Technical assumption | git's porcelain v2 and NUL-terminated output formats remain stable across versions ≥ 2.40 within the parsing contract | NFR-GIT-001 equivalence suite across the pinned git version matrix, per release | Adapter parsing updates per ADR-025 maintenance duty; version floor review |
| Technical assumption | The squash-merge commit message can be composed as PR title + body with the generated `Refs` trailer via platform configuration and merge automation | Verified during FR-GH-004 implementation; configuration audit | The commit-message check composes and validates the message via the merge automation path instead of platform defaults |

## Open questions

Entries per Volume 0 chapter 08; every PENDING VALIDATION in this volume is listed.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V11-OQ-1 | GitHub organization/namespace for the `andromeda` repository (PENDING VALIDATION, shared with Volume 0 chapter 01) | Chapter 03 | No — structure and process are namespace-independent | Project owner decision; Volume 0 open-questions register entry resolves | Open |
| V11-OQ-2 | Officially documented authentication grant types for third-party applications, per hosting provider (PENDING VALIDATION) | Chapter 02; ADR-147 | No — user-supplied tokens are the baseline | Verify against current official GitHub/GitLab documentation at tool implementation; record per-provider outcome | Open |
| V11-OQ-3 | GitLab feature parity for each change-request vocabulary operation on supported GitLab versions (PENDING VALIDATION) | Chapter 02 | No — GitLab is v1 phase | Capability verification against documented GitLab REST API at v1 implementation; divergences declared in tool capability notes | Open |
| V11-OQ-4 | Hosted-runner labels and coverage for the Tier 1 platform matrix, including macOS arm64 (PENDING VALIDATION) | Chapter 06 | No — required platform coverage is normative; labels are implementation detail | Check GitHub's documented hosted-runner offering at CI implementation; record labels in workflow config | Open |
| V11-OQ-5 | macOS Developer ID signing and notarization in the release pipeline (PENDING VALIDATION, owned by ADR-013 / Volume 1 register) | Chapter 06 (FR-GH-011 edge case) | No — checksums and cosign signatures ship regardless | Apple Developer account decision per ADR-013 review conditions | Open |

## Cross-volume references

Expectations this volume places on other volumes; consolidation verifies each.

| This volume defines | Volume expected to formalize / consume |
|---|---|
| Git operation catalog and the Beta/v1 additive GitPort amendment batches (ADR-145) | Volume 3 (port amendments via the change procedure at consolidation/phase boundaries) |
| Destructive-operation gate semantics bound to `git_mutation` and Approvals | Volume 9 (permission model, Approval machine); Volume 8 (confirmation flag grammar and prompts) |
| `github`/`gitlab` tool behavioral contracts | Volume 6 (built-in tool catalog declarations conform to this contract) |
| Restore-point and worktree semantics used by workflow rollback | Volume 4 (FR-WF-007 consumes GitPort worktree/safety-ref behavior) |
| `[git]`, `[github]`, `[git.hosting.gitlab]` key content | Volume 10 (schema, precedence, validation) |
| `git.*` event names and payload constraints | Volume 10 (envelope, delivery, retention, redaction) |
| Git operation latency expectations (status, diff on reference repositories) | Volume 12 (formal NFR-PERF budgets) |
| Pipeline quality gates (coverage SM-14, flaky-test quarantine policy, release qualification) | Volume 13 (gate definitions the chapter 06 workflows enforce) |
| Release pipeline CI realization (FR-GH-011) | Volume 14 (release semantics, channels, SM-18/SM-19, Release/Update machines) |
| Governance, maintainers, disclosure, and contribution files in the repository tree | Volume 15 (governance model); Volume 9 chapter 08 (security response) |
| PR provenance distribution reporting | Volume 15 (sustainability metrics input) |
