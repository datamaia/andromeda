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
| MS-1 Foundations | EP-01 ✅, EP-02 ✅, EP-03 ✅, EP-04 ✅ | ✅ |
| MS-2 Runtime core | EP-05 ✅, EP-06 ✅, EP-07 | 🔄 |
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

### EP-03 — Persistence and configuration · ✅

Realizes FR-CFG-001 (configuration precedence) and the ADR-007/028/029 persistence decisions.

- ✅ ULID generator (`internal/core`, ADR-027): monotonic, Crockford base32, with uniqueness
  and sort-order tests
- ✅ SQLite persistence (`internal/storage`, ADR-007): pure-Go modernc driver, WAL mode,
  `libc` pinned; workspace `state.db` + machine `global.db` split (ADR-028)
- ✅ Forward-only migrations (ADR-029): `user_version` tracking, pre-migration file backup,
  `integrity_check` + `foreign_key_check`, and clean refusal of future schemas
- ✅ `SessionStore` implementing `SessionStorePort`: sessions with optimistic-concurrency
  revisions, run records with per-run sequencing, and crash-recovery `MarkInterrupted`
- ✅ Reusable `Stream[T]` implementations (`internal/streams`): slice- and channel-backed
- ✅ Configuration Manager (`internal/config`) implementing `ConfigPort`: layered precedence
  (defaults → global → profile → workspace → project → runtime → env → cli), `ANDROMEDA_*`
  env mapping, TOML parsing (go-toml/v2), validation with E-CFG findings, per-value source
  attribution, and a file loader assembling layers from disk via the PAL
- ⬜ Live config file watching (`Watch` currently returns an empty stream) — later epic
- ⬜ Typed schema validation (ADR-024) beyond TOML syntax — later epic

**Gate status:** `make ci` passes — coverage 82.0%.

**Decision recorded (to reconcile with Volume 10 FR-CFG-004):** the `ANDROMEDA_*` env-var
mapping uses `__` to separate config-table levels and treats a single `_` as literal within a
key segment (so `ANDROMEDA_AGENT__MAX_ITERATIONS` → `agent.max_iterations`), with a
single-underscore fallback when no `__` is present (so the spec's `ANDROMEDA_TUI_THEME_MODE →
tui.theme.mode` example still works). This resolves the underscore ambiguity the spec itself
flags; Volume 10's text should adopt the same rule.

### EP-04 — Observability foundation · ✅

Realizes FR-OBS-001 (event envelope) and the ADR-011/012 observability decisions. **Completes
milestone MS-1.**

- ✅ Event bus (`internal/eventbus`) implementing `EventBusPort` (ADR-012): in-process typed
  pub/sub, exact-name and prefix topic selectors, bounded per-subscriber buffers with a
  drop-oldest overflow policy that never blocks publishers, context-cancel auto-close
- ✅ Event envelope (FR-OBS-001): `NewEvent` builder (version, UTC timestamp, correlation ID)
  and event-name grammar validation (`<area>[.<noun>].<verb-past>`)
- ✅ Structured logging (`internal/logging`, ADR-011): `slog` JSON handler with level control
  and secret redaction at the handler (Volume 9 redaction)
- ✅ Telemetry (`internal/telemetry`) implementing `TelemetryPort`: local-first metric registry
  and span tree; never fails the observed operation
- ✅ Event persistence (`internal/storage`): workspace-DB `events` table (migration v2) and an
  `EventStore` writing/reading enveloped Event records
- ✅ **`andromeda doctor`** composition (`internal/app`) exercising the MS-1 exit end to end:
  resolves config with attribution, opens both databases with migrations, emits and persists
  an enveloped event — verified live and by hermetic tests
- ⬜ OpenTelemetry Go SDK + OTLP export wiring (consent-gated) — local sinks now; SDK later
- ⬜ Full metric/trace/cost catalogs — owning volumes' later epics

**Gate status:** `make ci` passes — coverage 77.0%. `andromeda doctor` exits 0 with all checks
green.

## Milestone MS-1 — Foundations · ✅ COMPLETE

Exit criterion met: the `andromeda` binary starts on macOS (verified) and Linux (CI), resolves
configuration with source attribution, opens both the workspace and global databases with
migrations and backups, and emits enveloped events to persisted storage — all demonstrated by
`andromeda doctor` and covered by tests.

### EP-05 — Security kernel · ✅

- ✅ Permission Manager (`internal/permission`) implementing `PermissionPort` (**FR-SEC-100**):
  the closed 13/10/7 enums, the deny > ask > allow > else-ask evaluation algorithm (ADR-121),
  scope enclosure (session/workspace) and selector matching (exact, `/prefix/**`, `/prefix/*`),
  standing grants with expiry and revocation, policy rules, interactive approvals minting
  scoped grants, decision persistence and per-decision audit (fail-closed on audit failure,
  E-SEC-014), and unknown-permission → E-SEC-002 deny. Backed by workspace-DB migration v3.
- ✅ Secret Store (`internal/secret`) implementing `SecretStorePort` (**FR-SEC-102**, ADR-014):
  OS keychain backend via zalando/go-keyring (PAL `CredentialStore`, macOS/Linux) with an
  opt-in age-encrypted file fallback (passphrase/scrypt, 0600, verified encrypted-at-rest and
  wrong-passphrase-rejected); a per-namespace reference index so `List` works on keychains
  that cannot enumerate; secrets never surface in errors (E-SEC PortErrors carry no material)
- ✅ Sandbox Engine (`internal/sandbox`) implementing `SandboxPort` (**FR-SEC-101**, ADR-021):
  process-level MVP controls — deny-by-default env filtering (with sensitive-name stripping),
  command allow/deny lists, working-directory path policy, wall-clock time limit, and
  process-group teardown that kills the whole tree; effective containment level (`process`) is
  observable and never silently weakened. OS-level isolation (Seatbelt/Landlock) is the
  Beta/v1 layer (PENDING VALIDATION).
- ⬜ Approval state machine wiring to a real driver (CLI/TUI) — with EP-13

**Gate status:** `make ci` passes — coverage ~74%.

### EP-06 — Workspace and Git engines · ✅

- ✅ Git Engine (`internal/git`) implementing `GitPort` (**FR-GIT-001**, ADR-025): shells out
  to system git (≥ 2.40, gated at `Version`); `Status` (porcelain), `Diff`/`Log` (streamed),
  `Show`, `Stage`/`Unstage`, `Commit`, `ListBranches`/`CreateBranch`/`SwitchBranch`,
  `ApplyPatch` (check-then-apply, atomic), and worktree add/list/remove; failures map to E-GIT
  with git's stderr as safe detail. Verified against real temporary repositories.
- ✅ Workspace Engine (`internal/workspace`) implementing `WorkspacePort`: upward `Discover`
  (`.andromeda/` marker or `.git`), `Open` (creates the marker, opens the workspace database,
  registers in the global registry with a stable ID across reopens), `Snapshot` (root,
  projects, VCS summary via the Git Engine, timestamp), and clean `Close`.

**Gate status:** `make ci` passes. Ports implemented: 10 / 18.

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
