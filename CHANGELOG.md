# Changelog

All notable changes to Andromeda are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html) per
ADR-015. Release entries are derived from Conventional
Commit history by the release automation (ADR-013) and committed at release time.

## [Unreleased]

### Added

- **Roadmap board automation (FR-GH-008).** `project.yml` now drives the full status lifecycle of
  the Andromeda Roadmap board, not just intake: a linked PR opening moves its issues to
  `In Review`, merging moves them to `Validation`, a published release moves every `Validation`
  item to `Released` and stamps the Target release field, and an issue closed as *not planned* is
  archived off the board. Linked issues are resolved via the PR's `closingIssuesReferences`, and
  transitions are forward-only (a merged PR never regresses a shipped item). The logic lives in
  `scripts/project_sync.py`; a maintainer runbook (`docs/maintainers/roadmap-board.md`) documents
  the board's statuses, fields, automations, and the five UI-only views.

## [0.1.3] - 2026-07-12

Hardening and distribution release. Binaries are rebuilt with a patched standard library and
Homebrew installation is now available. (Supersedes 0.1.1 and 0.1.2, whose binaries published
but whose cask-publishing steps failed on goreleaser template issues; use 0.1.3 or later for
`brew`.)

### Added

- **Homebrew tap.** `brew install datamaia/tap/andromeda` (the release publishes a cask to the
  `datamaia/homebrew-tap` repository, with a Gatekeeper-quarantine post-install step until
  notarization lands).
- **CI expansion.** The full Tier-1 platform matrix (Linux and macOS on amd64/arm64), an
  enforced `golangci-lint` gate, and `e2e`, `security` (CodeQL + govulncheck), and
  `traceability` workflows.

### Fixed

- **Terminal signal delivery.** `Signal` now targets the whole process group, so `Wait` can no
  longer hang when a shell orphans a child holding the output pipes.
- **Standard-library advisories.** The build toolchain is pinned to go1.25.12, clearing the
  crypto/tls and crypto/x509 advisories that govulncheck flagged in the go1.25.0 stdlib.
- Resolved all 374 `golangci-lint` findings across the codebase.

## [0.1.0] - 2026-07-12

First tagged release. Cross-platform signed binaries (Linux and macOS, amd64 and arm64)
with SBOMs and a cosign signature are published on the GitHub release; install via
`scripts/install.sh` or download an archive directly.

### Added

- **EP-01 — Repository, CI, and process foundations.** Go module and walking-skeleton
  binary (`andromeda version`); the `make ci` local quality gate (format, lint, build,
  race tests, coverage gate, spec lint, structure check); a lean Linux CI mirror; issue
  and pull-request templates; CODEOWNERS; Dependabot; label taxonomy; and the community,
  security, and governance files. Realizes the Volume 11 repository structure (FR-GH-002)
  and branching rules (FR-GH-003).
- **EP-02 — Architecture skeleton and PAL.** The L0 core domain (ULID, phases, closed
  capability/permission enumerations), the L1 `ports` package with all 18 frozen port
  interfaces (FR-ARCH-003), the Platform Abstraction Layer surfaces with Unix reference
  implementations (FR-PORT-001), the layer manifest with an import-graph dependency test
  (ADR-033), and the `sdk/` mirror module (ADR-031).
- **EP-03 — Persistence and configuration.** The monotonic ULID generator (ADR-027); the
  SQLite persistence layer with WAL, the workspace/global database split (ADR-028), and
  forward-only migrations with backups (ADR-029); a `SessionStore`; reusable stream
  implementations; and the Configuration Manager with layered precedence, env mapping, and
  source attribution (FR-CFG-001).
- **EP-04 — Observability foundation.** The in-process event bus (ADR-012), the event
  envelope and name grammar (FR-OBS-001), `slog` JSON logging with redaction (ADR-011),
  a local telemetry recorder, event persistence, and the `andromeda doctor` diagnostic.
  **Completes milestone MS-1 (Foundations).**
