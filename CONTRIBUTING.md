# Contributing to Andromeda

Thank you for helping build Andromeda. This guide summarizes the conventions defined
normatively in **Volume 11** of the Andromeda specification. The specification is a private
companion document maintained by the project owner and is not published in this repository;
the volume, chapter, and `ADR-*` / `FR-*` / `NFR-*` references throughout these documents are
provided for traceability. When this guide and the specification disagree, the specification
wins — please open an issue.

## Ground rules

- **The specification is the source of truth.** Behavior, contracts, and requirements live in
  the specification, whose conventions are governed by **Volume 0**. Code implements
  requirements identified as `FR-*` / `NFR-*`; reference them in your PR.
- **Human review is mandatory.** No change merges without an approving human review. You may
  not approve your own PR.
- **AI-assisted changes are labeled at the PR level** with `ai-assisted` or `ai-generated` —
  never inside commit messages.

## Setup

```bash
git clone https://github.com/datamaia/andromeda
cd andromeda
git config core.hooksPath .githooks   # commit-message policy hook
make ci                               # the authoritative local quality gate
```

Requirements: Go (version per `go.mod`); `golangci-lint` (optional locally; `make lint`
degrades to gofmt+vet without it). Python 3.11+ is used by the spec linter only when the
private specification is present in the checkout; `make ci` skips it otherwise.

## Commit messages

Commit messages follow **Conventional Commits** and carry **change information only**. They
MUST NOT contain co-authorship, attribution, or advertising lines for AI tools, AI vendors,
or any company (no `Co-Authored-By`, no "Generated with", no emoji badges). The
`.githooks/commit-msg` hook and CI enforce this. See ADR-015.

Format: `<type>(<scope>): <imperative description>`

- **types**: `feat fix docs style refactor perf test build ci chore revert`
- **scopes**: the closed list in ADR-015 (e.g. `core`, `runtime`, `agent`, `provider`,
  `cli`, `tui`, `config`, `sandbox`, `perms`, `obs`, `spec`, `ci`, `release`, `deps`)
- Breaking changes use `!` and a `BREAKING CHANGE:` footer.

## Branches and pull requests

- Branch from `main`: `<type>/<issue-number>-<slug>` (e.g. `feat/141-provider-router`).
- Keep branches short-lived (≤ 3 working days; split anything over 10).
- Keep PRs small and focused; link the issue they close.
- Fill in the PR template, including the traceability and verification sections.
- `main` uses squash merges and linear history; head branches auto-delete on merge.

## Before you push

Run the full gate:

```bash
make ci
```

It runs formatting, lint, build, tests (with the race detector), the coverage gate, the
specification linter, and the repository structure check — the same set CI runs.

## Reporting security issues

Do not open a public issue for vulnerabilities. See [SECURITY.md](SECURITY.md).
