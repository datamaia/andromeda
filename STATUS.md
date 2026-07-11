# Andromeda — Implementation Status

Living tracker of the build. The **specification** (`docs/spec/`, v1.0.0) is complete; this
file tracks the **implementation** against Volume 15's epics and milestones. Updated and
pushed on every advance.

**Last updated:** 2026-07-12 · **Current milestone:** MS-1 (Foundations) · **Phase:** Core

## How work is organized

Implementation follows the Volume 15 epic sequence. Milestone **MS-1 "Foundations"** =
EP-01 → EP-04, whose exit is: *a binary that starts on macOS and Linux, resolves
configuration with attribution, opens both databases with migrations and backups, and emits
enveloped events to persisted storage.*

The authoritative quality gate is **`make ci`** (runs locally, no CI-minute dependency).

## Legend

✅ done · 🔄 in progress · ⬜ pending

## Milestones

| Milestone | Epics | Status |
|---|---|---|
| MS-1 Foundations | EP-01 ✅, EP-02 ✅, EP-03, EP-04 | 🔄 |
| MS-2 Runtime core | EP-05, EP-06, EP-07, … | ⬜ |
| MS-3+ | per Volume 15 ch 02 | ⬜ |

## Epics

### EP-01 — Repository, CI, and process foundations · 🔄 (near complete)

Realizes FR-GH-002 (repository structure) and FR-GH-003 (branching rules).

- ✅ Go module (`github.com/datamaia/andromeda`, `go 1.24`; toolchain 1.26 installed)
- ✅ Walking-skeleton binary: `andromeda version` (cobra, ADR-005) + `internal/buildinfo`
- ✅ `Makefile` — the authoritative local gate (`make ci`): fmt-check, lint, build, race
  tests, coverage gate, spec lint, structure check, tidy check
- ✅ `.golangci.yml` (ADR-018) with a depguard seed for the layer rules (ADR-033)
- ✅ Lean CI mirror `.github/workflows/ci.yml` (single Linux job — see deviations)
- ✅ Issue forms (15 types + `config.yml`), PR template, CODEOWNERS, `dependabot.yml`,
  `labels.yml`
- ✅ Community/governance files: LICENSE (Apache-2.0, ADR-002), CONTRIBUTING, CODE_OF_CONDUCT,
  SECURITY, GOVERNANCE, MAINTAINERS, CHANGELOG
- ✅ `scripts/structure_check.sh` (FR-GH-002 structure check)
- ✅ Commit-message hook already active (`.githooks/commit-msg`, ADR-015)
- ⬜ Full traceability validators (`scripts/traceability/`, branch-grammar/commit/linkage
  checks) — chapter 07; deferred to a dedicated EP-01 follow-up issue
- ⬜ Branch protection settings on `main` (configured on GitHub, not in-repo)

**Gate status:** `make ci` passes — build OK, tests green with `-race`, coverage 80.8%
(≥ 70% SM-14 floor), spec lint 0/0, structure-check OK.

### EP-02 — Architecture skeleton and PAL · ✅

Realizes FR-ARCH-003 (port freeze), FR-ARCH-001/004 (layering, context propagation),
FR-PORT-001 (platform encapsulation).

- ✅ L0 core (`internal/core`): ULID, `Phase`, and the closed capability (15), permission
  (13), scope (10), and decision (7) enumerations with wire-value guard tests
- ✅ L1 ports (`internal/ports`): **all 18 frozen port interfaces** with faithful signatures
  (`Provider`, `Auth`, `Tool`, `Terminal`, `MemoryStore`, `Indexer`, `EventBus`, `Permission`,
  `SecretStore`, `Sandbox`, `Config`, `SessionStore`, `Git`, `Workspace`, `Scheduler`,
  `Updater`, `Package`, `Telemetry`), shared `Stream[T]`/`PortError`/error-family primitives,
  and minimal contract types (grow additively per owning volume)
- ✅ Platform Abstraction Layer (`internal/pal`): the 19 surface interfaces; Unix reference
  implementations for Paths, ConfigDirs (XDG-honoring), TempFiles, and FileLocking (flock)
  with tests
- ✅ Dependency-rule enforcement (`internal/arch`): layer manifest + an import-graph test
  (ADR-033) that runs in `make test` with no external tooling; depguard seed in
  `.golangci.yml`
- ✅ `sdk/` mirror module (second Go module per ADR-031; builds independently, does not import
  `internal/`)
- ⬜ Concrete Unix implementations of the remaining PAL surfaces (Processes, Signals, PTY,
  Shell, CredentialStore, …) — delivered by their owning epics
- ⬜ SDK mirror content (port contract mirror) — Extension SDK epics

**Gate status:** `make ci` passes — coverage 85.5%; 18/18 ports; import-graph rule enforced.

### EP-03 — Persistence and configuration · ⬜
SQLite (modernc, WAL) with the workspace + global database split (ADR-028), forward-only
migrations with backups (ADR-029), and `andromeda.toml` loading with precedence, validation,
and env mapping (FR-CFG-001).

### EP-04 — Observability foundation · ⬜
Event bus, event envelope (FR-OBS-001), `slog` JSON logging, OpenTelemetry traces/metrics,
persisted event storage. Completes MS-1.

## Deliberate deviations from the specification (free-tier accommodations)

Recorded honestly so they can be reverted when the constraint lifts.

1. **CI platform matrix.** Volume 11 ch 06 mandates a macOS + Linux (×arch) test matrix. The
   repository is **private on GitHub's free plan**, where macOS runners cost 10× Actions
   minutes. The CI workflow therefore runs a **single Linux job** that invokes `make ci`;
   macOS coverage is provided by running `make ci` locally on the maintainer's Mac. The full
   matrix and the separate `policy.yml` / `traceability.yml` / `security.yml` / `e2e.yml`
   workflows activate when the repo is public or on a paid plan. **Reversible.**
2. **Action pinning.** CI actions use major-version tags, not full commit SHAs (ADR-149 asks
   for SHA pins). SHA pinning + Dependabot enforcement is a hardening step to schedule.
3. **CODEOWNERS is a single maintainer** (`@datamaia`) rather than maintainer teams; teams
   activate once the GitHub org exists (namespace PENDING VALIDATION, OQ-001).
4. **Structure check enforces the EP-01 subset** of the mandated tree; it extends as later
   epics add `sdk/`, `schemas/`, `packaging/`, `.goreleaser.yaml`, etc.

## Environment notes

- Go was not installed at the start; installed via Homebrew (`go1.26`, satisfies `go 1.24`).
- `golangci-lint` is optional locally (`make lint` degrades to gofmt+vet); CI installs it
  pinned. Install locally to run the full linter set.
