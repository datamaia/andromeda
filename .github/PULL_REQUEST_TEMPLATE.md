<!--
Pull request template (Volume 11 chapter 04). Keep PRs small and focused (SHOULD ≤ ~400
changed lines; MUST split large ones). A linked issue is required except for the enumerated
bot exemptions. Human review is mandatory; AI-generated changes MUST be labeled at the PR
level with `ai-assisted` or `ai-generated` — never inside commit messages.
-->

## Summary

<!-- What changes and why. One paragraph. -->

Closes #<!-- issue number (required) -->

## Requirements / traceability

<!-- Requirement IDs implemented or affected, e.g. FR-CFG-001, NFR-OBS-003. -->

- Requirements:
- Epic / milestone:

## Type of change

- [ ] feat
- [ ] fix
- [ ] docs
- [ ] refactor
- [ ] perf
- [ ] test
- [ ] build / ci
- [ ] chore
- [ ] spec

## Risk and compatibility

- [ ] No breaking change to a public contract (`internal/ports`, `sdk/`, `schemas/`)
- [ ] Breaking change — `!` / `BREAKING CHANGE` footer present and justified below
- [ ] Security-relevant (permissions, sandbox, secrets, auth) — describe below

## Verification

<!-- How this was verified. Paste the relevant `make ci` result or specific test output. -->

- [ ] `make ci` passes locally
- [ ] Tests added/updated for the change
- [ ] Docs / spec updated if behavior or contracts changed

## AI provenance

- [ ] This change was authored or substantially assisted by an AI agent (apply the
      `ai-assisted` / `ai-generated` label)
- [ ] Human-reviewed before merge (required)
