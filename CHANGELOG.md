# Changelog

All notable changes to Andromeda are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows
[Semantic Versioning](https://semver.org/spec/v2.0.0.html) per
[ADR-015](docs/spec/annexes/adr/ADR-015.md). Release entries are derived from Conventional
Commit history by the release automation (ADR-013) and committed at release time.

## [Unreleased]

### Added

- **EP-01 — Repository, CI, and process foundations.** Go module and walking-skeleton
  binary (`andromeda version`); the `make ci` local quality gate (format, lint, build,
  race tests, coverage gate, spec lint, structure check); a lean Linux CI mirror; issue
  and pull-request templates; CODEOWNERS; Dependabot; label taxonomy; and the community,
  security, and governance files. Realizes the Volume 11 repository structure (FR-GH-002)
  and branching rules (FR-GH-003).
