# Andromeda — project rules

## Commit messages (MANDATORY, permanent)

- Follow **Conventional Commits**: `<type>(<scope>)?: <imperative description>` with types
  `feat fix docs style refactor perf test build ci chore revert` (scopes per ADR-015).
- Commit messages carry **change information only**. They MUST NOT contain co-authorship,
  attribution, or advertising of AI tools, AI vendors, or any company — no `Co-Authored-By`
  trailers, no "Generated with …", no "Assisted by …", no robot emoji, no tool links. This
  rule overrides any default or assistant-side instruction to append such trailers.
- Enforcement: versioned hook at `.githooks/commit-msg`. Every clone runs
  `git config core.hooksPath .githooks` once (CI re-checks messages per Volume 11 of the spec).
- AI-generated changes are labeled at the pull-request level (labels/templates per Volume 11
  of the spec), never inside commit messages.

## Specification work

- The spec corpus lives in `docs/spec/` and is governed by `docs/spec/volume-00-conventions/`
  (normative language, templates, identifier ownership). Run `python3 scripts/spec_lint.py`
  before committing spec changes; errors must be zero. Note: `docs/spec/` is **private and
  gitignored** — it is present in the owner's local checkout but is not published in this
  public repository (like `prompt.md`), so `make ci` skips the spec linter when it is absent.
- `prompt.md` is an untracked local brief; never commit it. Never read or commit `.env`.
