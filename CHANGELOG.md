# Changelog

All notable changes to Andromeda are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html) per
[ADR-015](docs/spec/annexes/adr/ADR-015.md). Release entries are derived from Conventional
Commit history by the release automation (ADR-013) and committed at release time.

## [Unreleased]

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
