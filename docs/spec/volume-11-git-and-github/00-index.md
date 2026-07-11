# Volume 11 — Git, GitHub, and Development Platforms

**Status:** Complete · **Owner:** Git and development platform (Volume 11)

Volume 11 covers both roles Git and GitHub play for Andromeda. Chapters 01–02 specify the
**product**: the Git Engine (GitPort elaboration per ADR-025, the full operation catalog
with permissions, confirmations, and the no-silent-destruction policy) and the GitHub/
GitLab hosting integrations delivered as built-in tools over official APIs. Chapters
03–07 specify the **project**: the `andromeda` repository's structure, trunk-based
branching per ADR-004, the pull-request process with mandatory human review and PR-level
AI-provenance labels (commit messages carry change information only, per ADR-015), the
issue/label/Projects taxonomy, the GitHub Actions pipelines with their security posture,
and the traceability automation enforcing the Volume 0 chain from objective to shipped
artifact. Volume 11 mints identifiers in the `GIT` and `GH` areas and ADRs 145–159 (145–149
used).

## Chapters

| Chapter | Contents | Status |
|---|---|---|
| [01 — Git Engine](01-git-engine.md) | Keystone FR-GIT-001: system-git adapter behavior, repository discovery, operation catalog (status, diff, staging, commits, branches, tags, remotes, rebase, merge, cherry-pick, revert, reset, blame, log, worktrees, hooks, conflicts, submodules, sparse checkout, LFS, ignore rules, signing, protected branches), destructive-operation gate, E-GIT errors, `[git]` configuration | Complete |
| [02 — GitHub and GitLab as Product Integrations](02-github-gitlab-product-integrations.md) | Change-request vocabulary; `github`/`gitlab` built-in tools over official APIs; PR/MR preparation flows; hosting errors and configuration | Complete |
| [03 — Repository Structure and Branching](03-repository-structure-and-branching.md) | The `andromeda` monorepo tree per ADR-003/031, CODEOWNERS, community and security files, trunk-based branching rules and branch grammar | Complete |
| [04 — Pull Requests](04-pull-requests.md) | PR template, size limits, mandatory human review, no self-approval, squash-only merge, AI provenance labels, commit-message enforcement | Complete |
| [05 — Issues, Projects, and Roadmap](05-issues-projects-roadmap.md) | 15 issue types with forms, namespaced label taxonomy as synchronized data, the Andromeda Roadmap project (fields, views, automations), milestones | Complete |
| [06 — GitHub Actions Pipelines](06-github-actions.md) | Workflow inventory (quality, security, release, docs, audit), platform matrix, quality gates, YAML patterns, least-privilege/pinning/fork-isolation posture | Complete |
| [07 — Traceability Automation](07-traceability-automation.md) | Keystone FR-GH-001: the objective→artifact chain, validator suite, nightly audit, per-release chain report | Complete |
| [99 — Volume Register](99-volume-register.md) | Everything Volume 11 minted | Complete |

## Reading order and dependencies

Chapter 01 stands on Volume 3 chapter 02 (GitPort) and ADR-025; chapter 02 stands on
chapter 01 and the Volume 6 tool contract. Chapters 03–07 form the development-process
stack and are best read in order: structure → branching → PRs → issues/projects →
pipelines → traceability. Prerequisites from other volumes: Volume 0 chapters 02–03
(templates, identifier ownership, the traceability chain definition), Volume 1 chapter 05
(phases, MVP minimum), Volume 9 (permission names used by every mutation). Downstream
consumers: Volume 4 (workflow restore points over GitPort), Volume 6 (git/github/gitlab
built-in tools), Volume 8 (`andromeda git` command surface and confirmation flags),
Volume 12 (git operation latency budgets), Volume 13 (gates the pipelines enforce),
Volume 14 (release pipeline semantics realized by chapter 06), Volume 15 (governance
files referenced in chapter 03).
