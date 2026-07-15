# Changelog

All notable changes to Andromeda are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html) per
ADR-015. Release entries are derived from Conventional
Commit history by the release automation (ADR-013) and committed at release time.

## [Unreleased]

## [0.1.7] - 2026-07-15

Interactive TUI overhaul.

### Added

- **Navigable command menus.** `/skills`, `/mcp`, `/workflows`, and a new `/plugins` open an
  interactive menu that lists what exists (each drilling into a detail view with its file path and
  an "edit via chat" action), shows a friendly empty state when there is nothing, and offers to
  create a new one. `/ontology`, `/graph`, and `/memory` use the same drill-in/back framework
  (breadcrumb, per-row descriptions, ↑/↓ · enter/→ · esc/←).
- **File-based workspace memory.** `/memory` is now a CRUD menu over Markdown notes under
  `.andromeda/memory/` — each with frontmatter (a consecutive id, title, tags, date) and a generated
  `MEMORY.md` index — with tag/text search. The index is folded into the agent's system prompt,
  alongside `AGENTS.md`, so the model can recall durable facts.
- **Header badge and version.** The top banner shows a colored active-mode badge (agent / plan /
  shell) and the running version.
- **Remembered setup.** A bare launch defaults to the provider and model you last used and skips
  onboarding when it builds cleanly.

### Fixed

- **Command output on the start screen.** Slash commands run before the first conversation turn no
  longer vanish behind the brand splash — the splash yields to the transcript as soon as a command
  produces output.
- **`update` is descriptive.** `andromeda update` (and `/update`) now report the running version,
  channel, and build, check the real release feed, explain a development build, and show the exact
  upgrade command instead of a bare "up to date". The TUI runs the check off the UI thread.

## [0.1.6] - 2026-07-15

### Added

- **Ontology and graph workspace maps.** Two context-engineering surfaces that give an agent
  (or a person) a fast, deterministic map of a repository before touching it.
  `andromeda ontology build` scans the workspace (honoring `.gitignore`) and writes a Turtle
  (`.ttl`) description of how files and directories relate — byte-for-byte reproducible, with no
  timestamps. `andromeda graph serve` derives a node/edge graph, writes `graph.json` plus
  human-readable Markdown notes, and opens a self-contained, dependency-free force-directed viewer
  on localhost. Both are exposed in the TUI as the `/ontology` and `/graph` slash commands, each
  with a build / show / adjust-via-chat / delete menu. Artifacts live under the gitignored
  `.andromeda/` surface (ADR-028).
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
