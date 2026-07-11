# 03 — Repository Structure and Branching

From this chapter to chapter 07, Volume 11 specifies GitHub as the **development platform
of the `andromeda` project itself** — the second of the two roles the brief distinguishes.
The main repository is named `andromeda` (Volume 0 chapter 01 product identity); the
owning GitHub organization/namespace is PENDING VALIDATION (Volume 0 open-questions
register; referenced in this volume's register). GitHub provides code hosting, Issues,
Pull Requests, Discussions, Projects, Milestones, Releases, Actions, artifact storage,
Security Advisories, Dependabot, CodeQL, and secret scanning for the project; the
traceability chain (chapter 07) runs across these features.

## Repository layout

The tree below realizes ADR-003 (monorepo) and ADR-031 (two Go modules, `internal/`
component subtrees). It is normative at the level shown: top-level entries and the
`.github/` contents MUST exist as listed; deeper structure follows ADR-031 and the owning
volumes.

```text
andromeda/
├── .github/
│   ├── ISSUE_TEMPLATE/
│   │   ├── config.yml               # blank issues off; links to Discussions/security
│   │   ├── bug.yml                  # one form per issue type (chapter 05)
│   │   ├── feature.yml
│   │   ├── security.yml             # redirects to private vulnerability reporting
│   │   ├── documentation.yml
│   │   ├── performance.yml
│   │   ├── refactor.yml
│   │   ├── architecture.yml
│   │   ├── research.yml
│   │   ├── tech-debt.yml
│   │   ├── release.yml
│   │   ├── provider-integration.yml
│   │   ├── tool-integration.yml
│   │   ├── mcp.yml
│   │   ├── plugin.yml
│   │   └── platform-compat.yml
│   ├── PULL_REQUEST_TEMPLATE.md     # chapter 04
│   ├── CODEOWNERS                   # path-rule ownership (below)
│   ├── dependabot.yml               # gomod + github-actions, grouped weekly
│   ├── labels.yml                   # label taxonomy as data (chapter 05)
│   └── workflows/                   # chapter 06 pipelines
├── cmd/andromeda/                   # composition root (ADR-031)
├── internal/
│   ├── core/                        # L0 domain
│   ├── ports/                       # L1: 18 port interfaces + contract types
│   └── <component>/                 # one subtree per glossary component
├── sdk/                             # public Extension SDK; separate go.mod (ADR-003)
├── docs/
│   ├── spec/                        # this specification (Volumes 0–15 + annexes)
│   └── user/                        # user documentation source (docs pipeline)
├── examples/                        # runnable extension and configuration examples
├── benchmarks/                      # Volume 12 benchmark suite
├── schemas/                         # published JSON Schemas (config, tool, workflow, output)
├── scripts/
│   ├── spec_lint.py                 # specification linter (Volume 0)
│   ├── traceability/                # chapter 07 validators
│   └── hooks/                       # versioned git hooks (commit-msg, pre-push)
├── test/
│   ├── e2e/                         # end-to-end journeys (Volume 13)
│   └── fixtures/                    # shared cross-package fixtures
├── packaging/
│   ├── installer/install.sh         # shell installer source (Volume 14)
│   └── nfpm/                        # Linux package metadata inputs
├── .goreleaser.yaml                 # release pipeline config (ADR-013)
├── .golangci.yml                    # lint config (ADR-018)
├── CHANGELOG.md                     # generated per ADR-015; committed per release
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── GOVERNANCE.md                    # governance summary; full model in Volume 15
├── LICENSE                          # Apache-2.0 (ADR-002)
├── MAINTAINERS.md
├── README.md
├── SECURITY.md                      # reporting policy; Volume 9 chapter 08 pointer
├── go.mod
└── go.sum
```

Prose constraints on the tree:

- **Go-conventional test placement.** Unit tests live next to their packages (`_test.go`
  files, per-package `testdata/`); `test/` holds only cross-package suites and shared
  fixtures. Coverage gates are Volume 13's.
- **`docs/spec/` is code.** Specification changes flow through the same PR process, gated
  by `scripts/spec_lint.py` as a required check on paths under `docs/spec/`.
- **`schemas/` is public contract surface** under ADR-015 SemVer rules; the contract-diff
  tooling (NFR-ARCH-002, SM-20) watches it.
- **Release metadata** lives in `.goreleaser.yaml`, `packaging/`, and generated release
  notes; no hand-maintained release metadata exists elsewhere.
- **Ignore rules:** the repository `.gitignore` covers build outputs, coverage artifacts,
  and `.andromeda/` (the project-local directory of the product itself when developing
  Andromeda inside Andromeda).

### CODEOWNERS

`CODEOWNERS` expresses review ownership by path, aligned with component subtrees and the
Volume 0 area taxonomy. Mandatory rules: `internal/ports/` and `sdk/` (contract
surfaces), `docs/spec/` (specification), `.github/workflows/` and `scripts/traceability/`
(process integrity), `internal/security/`-relevant subtrees (Permission Manager, Sandbox
Engine, Secret Store), and `.goreleaser.yaml` + `packaging/` (release integrity) MUST
each name at least one maintainer group as required reviewers. CODEOWNERS gates review,
not write access (ADR-003 consequence); sensitive-path review is additionally enforced by
branch protection requiring code-owner review (chapter 04).

### Community and security files

- `SECURITY.md` states the private vulnerability-reporting path (GitHub Security
  Advisories / private reporting), the response-time commitment bound to SM-16(c) (≤ 3
  business days to first response), and points to Volume 9 chapter 08 for incident
  response and to Volume 15 for disclosure governance.
- `CONTRIBUTING.md` documents the branch/commit/PR conventions of this volume, the local
  hook installation (`scripts/hooks/`), and the AI-provenance labeling duty (chapter 04).
- `GOVERNANCE.md` and `MAINTAINERS.md` summarize the Volume 15 governance model;
  Discussions hosts design questions and Q&A so Issues stay actionable (chapter 05).

## Branching model

ADR-004 fixes trunk-based development; this section is its repository-level
implementation.

| Rule | Value |
|---|---|
| Default branch | `main` — the only permanent branch |
| Protection (main) | PRs only; required checks green; ≥ 1 approving review (code-owner review on owned paths); linear history (squash merge only, ADR-015); force pushes and deletion disabled; rules apply to administrators |
| Working branches | `<type>/<issue-number>-<slug>`; short-lived (SHOULD merge ≤ 3 working days; MUST split/close > 10, per ADR-004) |
| Release branches | `release/vX.Y`, cut from `main` for stabilization/backports only; protected like `main`; no feature work; deleted or frozen at end of support (Volume 14 support windows) |
| Hotfix path | Fix lands on `main` first, then cherry-picked to affected `release/vX.Y` (backport policy per Volume 14); no `hotfix/*` long-lived branches |
| Branch naming grammar | `type ∈ {feat, fix, docs, refactor, perf, test, build, ci, chore, spec, release-prep}`; issue number mandatory except for the enumerated bot exemptions |
| Updates | Rebase working branches on `main`; merge commits from `main` into working branches are prohibited (linear-history protection rejects them at merge anyway) |
| Deletion | Head branches auto-delete on merge |
| Forks | External contributions come from forks; fork PRs are untrusted (ADR-149); maintainers MUST NOT push to contributor forks without consent |
| Bots | `dependabot/**` branches are exempt from the naming grammar and issue linkage (they carry their own traceability); other bots require a recorded maintainer decision before write access |

The naming grammar's `type` list mirrors Conventional Commits types (plus `spec` and
`release-prep`), so branch, commit type, and changelog section align. The traceability
validator (chapter 07) enforces the grammar on every PR.

## Requirements

### FR-GH-002 — Repository structure

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Development process (Volume 11)
- Affected components: None at runtime (repository artifact); enforced by CI
- Dependencies: ADR-002, ADR-003, ADR-031; FR-GH-009
- Related risks: RISK-GH-002

#### Description

The `andromeda` repository MUST contain the layout of this chapter: the listed top-level
entries, the complete `.github/` set (issue forms for every chapter 05 type, PR template,
CODEOWNERS with the mandatory path rules, dependabot configuration, labels data,
workflows), the community and security files with the stated contents, and the two-module
Go layout of ADR-031. A structure check in CI MUST fail when a mandatory entry is missing
or a CODEOWNERS mandatory rule is absent.

#### Motivation

The structure is where every other process rule attaches: templates make the traceability
chain fillable, CODEOWNERS makes sensitive review enforceable, and the fixed layout keeps
spec, code, and packaging co-versioned (ADR-003 rationale).

#### Actors

Maintainers, contributors (human and AI agents), CI.

#### Preconditions

Repository exists on the (PENDING VALIDATION) organization.

#### Main flow

1. The repository is initialized with the layout before implementation phases begin.
2. Every structural change lands via PR; the structure check validates presence rules.
3. CODEOWNERS routes review for owned paths automatically.

#### Alternative flows

- A new component subtree is added → `CODEOWNERS` and the ADR-033 layer manifest are
  updated in the same PR (the structure check cross-references them).

#### Edge cases

- Generated files (CHANGELOG.md, labels sync) are committed by automation through the
  same PR-or-release paths, never by direct push.
- The `sdk/` module MUST NOT gain a dependency without the ADR-002 license-policy
  allowlist check passing.

#### Inputs

Repository tree; CODEOWNERS; templates.

#### Outputs

Structure-check CI results; review routing.

#### States

Not applicable (static artifact validated per change).

#### Errors

Structure-check failures surface as failed required checks with itemized findings
(tooling errors use E-GH-004 semantics from chapter 07's validator family).

#### Constraints

Top-level layout changes require a PR touching this chapter (specification and tree
evolve together); ADR-031 governs module boundaries.

#### Security

CODEOWNERS mandatory rules cover contract, security, process, and release paths;
SECURITY.md keeps the private reporting path discoverable (SM-16(c)).

#### Observability

Structure-check results as check annotations; nightly audit (chapter 07) re-validates.

#### Performance

Structure check runtime is trivial; no budget beyond FR-GH-009's overall check latency.

#### Compatibility

Layout is platform-neutral; paths are POSIX-style as Git stores them.

#### Acceptance criteria

- Given a fresh clone, when the structure check runs, then every mandatory entry of this
  chapter is present and the check passes.
- Given a PR deleting `SECURITY.md` or removing a mandatory CODEOWNERS rule, when checks
  run, then the structure check fails with the missing entry named.
- Given a PR touching `internal/ports/`, when review is requested, then code-owner
  review is required and non-owner approval alone cannot merge.
- Negative: a PR adding a third Go module without an accompanying ADR fails the
  structure check (module count is pinned to ADR-031).

#### Verification method

Structure-check job in CI (required); CODEOWNERS behavior verified by branch-protection
configuration audit at phase gates; layout audit in the release qualification (Volume
13).

#### Traceability

PRD-013; ADR-002, ADR-003, ADR-031; chapter 04 (review rules), chapter 06 (workflows),
chapter 07 (validators).

### FR-GH-003 — Branching rules and branch hygiene

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Development process (Volume 11)
- Affected components: None at runtime; enforced by branch protection and CI
- Dependencies: ADR-004, ADR-015, ADR-148; FR-GH-001
- Related risks: RISK-GH-002

#### Description

The repository MUST implement the branching table of this chapter: protected `main` with
squash-only linear history and admin-inclusive rules, the branch naming grammar validated
on every PR, release-branch discipline per ADR-004, auto-deletion of merged heads, fork
handling per ADR-149, and the enumerated bot exemptions. Direct pushes to protected
branches MUST be impossible for all actors.

#### Motivation

Branch mechanics are where trunk-based development either exists or silently erodes;
encoding the rules in protection settings and validators makes ADR-004 a property, not a
convention.

#### Actors

Contributors, maintainers, bots, CI.

#### Preconditions

FR-GH-002 structure in place; branch protection configured.

#### Main flow

1. Contributor branches from `main` using the grammar.
2. Work merges via PR under chapter 04 rules; the head branch auto-deletes.
3. Release stabilization cuts `release/vX.Y`; fixes cherry-pick from `main`.

#### Alternative flows

- Emergency fix: same path — the process defines no bypass; revert-first (ADR-004) is
  the incident response for a broken `main`.
- Backport: maintainer cherry-picks merged commits to `release/vX.Y` via PR with the
  `backport` label.

#### Edge cases

- Grammar collisions (two PRs, one issue) are legal: multiple branches MAY reference one
  issue; the slug differentiates.
- Long-running spec work uses the same grammar (`spec/…` type) and the same 10-day
  split rule.
- Renamed default branch is prohibited without a superseding decision (tooling pins
  `main`).

#### Inputs

Branch names, PR metadata; protection configuration.

#### Outputs

Validator results; protection enforcement.

#### States

Branch lifecycle (created → merged/closed → deleted) is platform state; no product
state machine.

#### Errors

Branch-name violations fail the traceability validator with E-GH-002.

#### Constraints

No long-lived branches other than `main` and active `release/*`; no merge commits on
`main`; grammar types fixed to the listed set.

#### Security

Admin-inclusive protection prevents privileged bypass; fork rules per ADR-149 keep
untrusted code out of privileged contexts.

#### Observability

Nightly audit reports branch-age outliers (> 10 working days) and protection-setting
drift against this chapter.

#### Performance

Branch-name validation is O(1) per PR within the FR-GH-009 latency budget.

#### Compatibility

GitHub branch-protection primitives; the rules are expressible on any platform with
protected branches + required checks (ADR-148 reversal note).

#### Acceptance criteria

- Given any actor including an administrator, when a direct push to `main` is attempted,
  then the platform rejects it.
- Given a PR from branch `feat/141-provider-router`, when the validator runs, then the
  grammar check passes; given `my-fix`, then it fails with E-GH-002 naming the grammar.
- Given a merged PR, when merge completes, then the head branch is deleted and `main`'s
  new commit is a squash commit whose subject is the PR title.
- Negative: a PR targeting `release/v0.4` with type `feat` fails validation
  (release branches accept only fix/backport types).

#### Verification method

Branch-protection configuration audit at phase gates (scripted against the platform
API); traceability validator tests (chapter 07); nightly audit assertions.

#### Traceability

ADR-004, ADR-015, ADR-148, ADR-149; FR-GH-001; Volume 14 backport policy.

## Cross-references

Merge mechanics, review requirements, and PR content rules: chapter
[04](04-pull-requests.md). Label taxonomy and issue forms: chapter
[05](05-issues-projects-roadmap.md). Workflow definitions: chapter
[06](06-github-actions.md). Enforcement automation: chapter
[07](07-traceability-automation.md).
