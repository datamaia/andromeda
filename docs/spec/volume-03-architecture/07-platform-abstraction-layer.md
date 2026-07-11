# 07 — Platform Abstraction Layer

The Platform Abstraction Layer (PAL) is the single component that touches operating-system
specifics. Unix is the reference behavior (PRD-011): Andromeda uses processes, POSIX signals,
file descriptors, POSIX permissions, pseudoterminals, pipes, Unix domain sockets, symlinks,
environment variables, shells, process groups, and execute bits — and every one of those uses
is encapsulated here. Components above the PAL are platform-blind; the future Windows phase
(v2 candidate, Volume 1) is a PAL backend project, not a codebase audit.

## PAL component contract

| Aspect | Specification |
|---|---|
| Responsibility | One Go API per platform surface (19 surfaces below) with per-OS backends: macOS and Linux at MVP, Windows in the v2 candidate phase |
| Boundaries | Mechanics only — the PAL implements *how* on each OS and never decides *whether* (policy belongs to Permission Manager, Sandbox Engine, and configuration) |
| Public API | The 19 surface interfaces, consumed by L3 adapters (and by drivers only for terminal capability probing) |
| Internal API | Per-OS backend implementations selected at build time via build tags |
| Allowed dependencies | Standard library, `golang.org/x/sys`, adrg/xdg (ADR-022), go-keyring + age (ADR-014), PTY library per Volume 6 selection — all pinned |
| Prohibited dependencies | Every other Andromeda component except the Core Domain; no engine, adapter, or driver types |
| Inputs | Surface calls from L3 adapters |
| Outputs | Portable results and E-PORT errors; capability probe reports |
| Events emitted | `pal.platform.rejected`, `pal.capability.degraded`, `pal.fallback.engaged` (minted by this volume, below) |
| Errors | E-PORT family (this chapter) |
| States | Stateless services; capability probe results cached per process |
| Persistence | None (it resolves where others persist) |
| Concurrency | All surfaces concurrent-safe; blocking calls context-bounded per FR-ARCH-004 |
| Security | Correct mode/permission application on created files and sockets; environment filtering primitives for ADR-021; secret material passes through Credential Store calls only |
| Observability | Capability probes and fallbacks logged at startup diagnostics; per-surface conformance status queryable |
| Testing | The PAL conformance suite: one behavioral suite per surface, executed on every Tier 1 platform in CI (NFR-PORT-002) |
| Extensibility | New OS backends per surface; not third-party pluggable |
| Phase | Core |

Contract rules:

1. **No platform checks outside the PAL.** `runtime.GOOS`, build tags on OS, `syscall`/
   `golang.org/x/sys` usage, and OS-conditional behavior are permitted only inside
   `internal/pal` (FR-PORT-001). Everything else consumes surfaces.
2. **Portable signatures.** Surface interfaces use portable types — `SignalName` not raw
   signal integers, portable path values not separator-assuming strings, capability
   descriptors not mode bits — so that a Windows backend can implement the same interface
   without signature changes (FR-PORT-002).
3. **Probed capabilities, explicit degradation.** Each surface exposes `Probe(ctx)` reporting
   availability and feature level on the current host. Absence degrades per the surface's
   declared policy — never by silent no-op (E-PORT-002; `pal.capability.degraded`).
4. **Mechanics, not policy.** The PAL never prompts, never consults permissions, and never
   makes trust decisions; callers hold decisions before calling.

## The 19 platform surfaces

| Surface | Abstraction (one per surface) | Unix mapping (reference) | Windows-future mapping | Primary consumers |
|---|---|---|---|---|
| Filesystem | Read/write/stat/watch with atomic replace, symlink-aware traversal, case-sensitivity metadata | POSIX fs, inotify/FSEvents | Win32 fs, ReadDirectoryChangesW; junctions; case-insensitivity metadata | Workspace Engine, Configuration Manager, Persistence Layer, Indexing Engine |
| Paths | Portable path values: join/split/normalize, workspace-relative forms, reserved-name and length validation | `/` paths | Drive/UNC forms, long-path handling | All L3 adapters |
| Permissions | File permission get/set with a portable model | POSIX mode bits, umask, execute bits | ACL mapping (Windows chapter of the v2 phase) | Secret Store, Persistence Layer, Updater |
| Processes | Spawn with argv arrays, filtered environment, working dir, resource limits; wait with outcome | fork/exec, rlimits | CreateProcess, Job Object limits | Sandbox Engine, Git Engine, Plugin Runtime |
| Signals | Deliver/handle portable signal names; interrupt forwarding | POSIX signals, process groups | Console control events + Job semantics | Terminal Engine, Sandbox Engine, chapter 08 shutdown |
| PTY | Allocate/resize/drive pseudoterminals | openpty/forkpty family | ConPTY | Terminal Engine |
| Shell | Locate and describe user/login shells; portable invocation specs (never string interpolation) | `$SHELL`, `/etc/shells`; bash/zsh/fish | PowerShell discovery | Terminal Engine, Tool Discovery |
| Credential Store | Store/retrieve/delete secrets by reference | Keychain (macOS), Secret Service (Linux) via go-keyring; age-encrypted fallback (ADR-014) | Credential Manager | Secret Store |
| Notifications | Post local user notifications; capability-probed | osascript/notify-send class mechanisms, terminal bell fallback | Windows notifications | TUI (long-run completion per Volume 8) |
| Clipboard | Read/write text clipboard; capability-probed | pbcopy/pbpaste, wl-clipboard/xclip; OSC 52 fallback over SSH | Win32 clipboard | TUI |
| Installer | Register/unregister the installed binary and integrations; ownership checks | Install paths per Volume 14 (Homebrew, tarball, script) | Installer integration (MSI/winget class, Volume 14 v2) | Updater, Volume 14 tooling |
| Updater | Atomic binary replace-or-restore with retained previous version | rename-swap on same filesystem, retained artifact | Locked-file-aware swap strategy | Updater |
| Sandbox | Apply isolation primitives to a process spec | Env filtering + rlimits (MVP); Seatbelt / Landlock / namespaces / bubblewrap (Beta/v1, each PENDING VALIDATION per ADR-021) | AppContainer-class mechanisms (evaluated in the Windows phase) | Sandbox Engine |
| Tool Discovery | Locate executables and runtime prerequisites with version probes | `$PATH` walk, `--version` probes (git ≥ 2.40 per ADR-025) | PATH/App Paths | Git Engine, Terminal Engine, diagnostics |
| Config Directories | Resolve config/data/state/cache/log/runtime directories | XDG via adrg/xdg; Apple-native mapping honoring `XDG_*` overrides (ADR-022) | Known Folders via the same library | Configuration Manager, Persistence Layer, Logging, Updater |
| File Locking | Advisory exclusive/shared locks with timeout | flock | LockFileEx | Persistence Layer, Updater, Secret Store fallback, Workspace Engine |
| Local IPC | Listen/dial local endpoints with peer identity and owner-only access | Unix domain sockets, `0600`, runtime dir (ADR-012, ADR-022) | Named pipes (ADR-012) | IPC server, future companion tooling |
| Temporary Files | Private temp files/dirs with correct modes and cleanup registration | `$TMPDIR`/`/tmp` with 0700 dirs | `%TEMP%` with owner ACLs | Package Manager, Updater, Git Engine, Sandbox Engine |
| Process Trees | Enumerate/terminate a process and all descendants; usage accounting | Process groups/sessions, kill-by-group | Job Objects | Sandbox Engine, Plugin Runtime, Terminal Engine |

Surface notes where the table needs elaboration:

- **Filesystem** exposes case-sensitivity and symlink-support metadata per path root, because
  macOS defaults (case-insensitive APFS) and Linux defaults differ and the Workspace
  Engine/Indexing Engine must not guess. Atomic replace is the primitive behind config
  writes, database backups, and update apply.
- **Local IPC** carries peer credential resolution (UID on Unix) so the IPC server can
  enforce same-user access before any request parsing (chapter 08).
- **Sandbox** is mechanism application only. The MVP level (environment filtering, rlimits,
  working-dir confinement) is portable; each OS-level isolation mechanism is a separately
  probed feature with its own PENDING VALIDATION gate (ADR-021) — the Sandbox Engine, not the
  PAL, decides which probed level a policy requires.
- **Updater/Installer** split: Installer knows where and how the product was installed
  (package-manager-owned installs MUST be detected and self-update refused in favor of the
  owning manager, per Volume 14 rules); Updater performs the atomic swap where self-update is
  permitted.

## Platform support matrix

Tier definitions are Volume 1's (chapter 05): Tier 1 platforms build, test, and gate every
release on the full acceptance suite; Tier 2 platforms build and smoke-test, and defects do
not gate releases.

| Platform | Architecture | Minimum version | Tier | Phase |
|---|---|---|---|---|
| macOS on Apple Silicon | arm64 | macOS 13 | Tier 1 | MVP |
| macOS on Intel | x86_64 | macOS 13 | Tier 2 — PENDING VALIDATION (build/test capacity; open questions register) | MVP when viable |
| Ubuntu | x86_64, arm64 | 22.04 LTS | Tier 1 | MVP |
| Debian | x86_64, arm64 | 12 | Tier 1 | MVP |
| Fedora | x86_64, arm64 | 39 | Tier 1 | MVP |
| Other Linux distributions | x86_64, arm64 | Kernel floor below | Best effort (no gate) | MVP |
| Windows 11 native | x86_64; arm64 subject to viability | Windows 11 | — | v2 candidate (Volume 1) |
| Other Unix or Unix-like systems | — | — | — | Future |

Kernel and platform notes:

- **Linux kernel floor.** MVP functionality assumes no kernel features beyond the Go runtime's
  requirements as shipped by the reference distributions (Ubuntu 22.04 ships 5.15). Feature
  floors are per capability, probed at runtime: **Landlock requires kernel ≥ 5.13 and is
  PENDING VALIDATION per ADR-021** before any Beta/v1 isolation claim; namespaces and
  bubblewrap availability vary by distribution policy and are likewise PENDING VALIDATION
  (ADR-021). Absence degrades the sandbox level observably, never silently.
- **Linux binaries are static.** The pure-Go build (ADR-001, ADR-007 cgo-free posture)
  produces statically linked Linux binaries with no glibc/musl coupling; distribution
  packaging is Volume 14's.
- **macOS floor rationale.** macOS 13 keeps two major versions of headroom below current
  releases while bounding the Seatbelt/notarization validation matrix (ADR-013, ADR-021).
- **WSL is Linux.** Andromeda under WSL is supported *as Linux* and MUST NOT be documented or
  marketed as Windows support (Volume 1 platform scope rule).

### Reference shells

| Shell | Role |
|---|---|
| bash ≥ 5 | Reference POSIX-family shell; CI and script baseline |
| zsh | Reference interactive shell (macOS default); completion target |
| fish | Supported interactive shell; completion target |
| POSIX `sh` | Floor for generated scripts (install/uninstall per Volume 14) |

Shell integration features (completions, launchers) target exactly these; the Terminal Engine
executes commands via argv arrays by default and via a user-selected shell only where Volume 6
semantics say so.

### Terminal expectations

The TUI operates within these capability tiers (theming tiers per ADR-026, behavior per
Volume 8): truecolor, 256-color, 16-color, and monochrome — probed, never assumed. Minimum
supported geometry is 80×24 with UTF-8; narrower terminals degrade per Volume 8 layout rules.
Reference terminal emulators for the Volume 13 compatibility matrix: Terminal.app, iTerm2,
GNOME Terminal, Konsole, Alacritty, kitty, WezTerm; tmux and screen are first-class
multiplexer targets (with OSC 52 clipboard fallback over SSH).

### Constrained environments

| Environment | Contract |
|---|---|
| SSH sessions | Full CLI/TUI function over a remote TTY; no GUI dependency anywhere in core paths; Clipboard falls back to OSC 52, Notifications to the terminal bell, per surface probes |
| Headless (no TTY) | CLI non-interactive mode and the IPC surface operate fully (PRD-009); the TUI refuses cleanly with a usage error; permission resolution is policy-only — no prompts (exit code 5 on unresolved) |
| Containers | Supported for CLI/headless use; missing HOME or XDG roots engage the ADR-022 fallback (`~/.andromeda` or configured root) with the `pal.fallback.engaged` diagnostic |
| CI (GitHub Actions reference) | Non-interactive contract as headless; deterministic exit codes; no terminal capability assumptions; the offline suite (SM-05) runs here with network disabled |

## XDG compatibility and directory layout

Directory resolution follows ADR-022 exactly: adrg/xdg through the Config Directories
surface; XDG Base Directory layout on Linux; Apple-native mapping on macOS (`~/Library/
Application Support`, `~/Library/Caches`) honoring explicitly set `XDG_*` overrides;
project-local state in `.andromeda/`; global configuration at the platform config directory
plus `andromeda.toml`; the `~/.andromeda` fallback engaged only on resolution failure, always
with a diagnostic. Volume 10 owns the precedence semantics of what lives in these locations;
chapter [09](09-deployment-update-extensibility-compatibility.md) tabulates the deployment
footprint. No component other than the PAL computes any of these paths (FR-PORT-003).

## Windows-future encapsulation rules

The v2-candidate Windows phase (Volume 1) is prepared for now, structurally:

1. Every surface's Windows mapping is recorded in the surface table above and kept honest as
   surfaces evolve; a surface change that has no plausible Windows implementation is rejected
   at review (FR-PORT-002).
2. The Windows work already scoped by existing decisions: named pipes for Local IPC
   (ADR-012), Credential Manager for Credential Store (ADR-014 direction), known-folder
   mapping via adrg/xdg (ADR-022 review condition), ConPTY for PTY, Job Objects for Process
   Trees and limits. The full Windows platform chapter (paths, ACLs, long paths, junctions,
   case sensitivity, PowerShell, Windows Terminal, installers, signing, system security
   integration) is authored in the v2 phase as an amendment to this volume.
3. Until that phase, Windows code MUST NOT be merged outside `internal/pal` backends and MUST
   NOT weaken any surface signature (no `interface{}`/"platform-specific" escape hatches).

## Requirements

### FR-PORT-001 — Platform encapsulation

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Architecture (Volume 3)
- Affected components: all components; PAL
- Dependencies: FR-ARCH-001; ADR-022
- Related risks: RISK-PORT-003

#### Description

All operating-system-specific behavior MUST be encapsulated in the PAL. Outside
`internal/pal`, production code MUST NOT: branch on `runtime.GOOS` or OS build tags, import
`syscall` or `golang.org/x/sys`, invoke OS-specific binaries for platform services, or encode
platform-specific paths, separators, signal numbers, or permission bits. Platform-conditional
behavior above the PAL exists only as responses to portable capability probes.

#### Motivation

The brief's mandate is explicit: encapsulate every OS dependency; do not scatter platform
checks across the codebase. Scattered checks are how the Windows phase becomes a rewrite
(PRD-011) and how Tier 1 platforms drift apart behaviorally (SM-17).

#### Actors

Implementers; CI enforcement; PAL maintainers.

#### Preconditions

PAL surfaces cover the needed capability (FR-PORT-002 guarantees coverage evolution).

#### Main flow

1. A component needs an OS capability.
2. It consumes the corresponding PAL surface through its L3 position or port.
3. CI's platform-encapsulation check (NFR-PORT-004) verifies no direct OS coupling exists
   outside the PAL.

#### Alternative flows

- The capability has no surface: the contributor proposes a surface addition (or extension)
  to this chapter through the change procedure; interim direct usage is not permitted.

#### Edge cases

- Third-party libraries with internal platform handling (Bubble Tea's terminal handling,
  adrg/xdg) are permitted at their sanctioned layer; the rule governs Andromeda's own code.
- Test files MAY use platform conditionals to express platform-specific expectations.
- The Go standard library's portable APIs (`os`, `path/filepath`) are permitted everywhere;
  the rule targets platform-*conditional* behavior, not portable stdlib use — with the
  specific prohibitions listed in the Description.

#### Inputs

Source tree; encapsulation check configuration.

#### Outputs

CI verdicts; violation reports naming file and construct.

#### States

Not applicable — static structural requirement.

#### Errors

Violations are CI failures; at runtime, capability gaps surface as E-PORT-002 through probes,
never as scattered conditionals.

#### Constraints

The prohibited-construct list is maintained with the encapsulation check (ADR-033 mechanism).

#### Security

Centralized OS interaction concentrates the code able to change file modes, spawn processes,
and touch keychains into one audited component.

#### Observability

Check results per commit; PAL probe diagnostics at startup show exactly what platform
behavior is in effect.

#### Performance

Surface indirection is interface dispatch; hot-path surfaces (Filesystem, Processes) are
designed for zero per-call allocation beyond the operation's own.

#### Compatibility

This rule is the load-bearing wall for the Windows phase and for Tier parity (SM-17).

#### Acceptance criteria

- Given the production source outside `internal/pal`, when the encapsulation check runs, then
  zero occurrences of the prohibited constructs exist.
- Given a PR introducing `runtime.GOOS` into an engine, when CI runs, then the check fails
  naming the file and line.
- Given startup on a host lacking an optional capability (e.g., no Secret Service), when
  diagnostics run, then the gap is reported via probe results (`pal.capability.degraded`),
  and no code path outside the PAL branches on the OS to handle it.

#### Verification method

Automated CI check per NFR-PORT-004 (lint rules + grep-class scanner per ADR-033); release
audit; PAL conformance suite verifying surface behavior parity.

#### Traceability

PRD-011; ADR-022, ADR-033; FR-ARCH-001; NFR-PORT-004; RISK-PORT-003.

### FR-PORT-002 — PAL surface completeness and portability of signatures

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Architecture (Volume 3)
- Affected components: PAL; all L3 consumers
- Dependencies: FR-PORT-001
- Related risks: RISK-PORT-003

#### Description

The PAL MUST provide exactly one abstraction per platform surface for all 19 surfaces of this
chapter, each with: a portable interface signature implementable on POSIX and Windows
primitives, a documented Unix mapping and Windows-future mapping, a `Probe` capability report,
and a declared degradation policy for absence. Adding, splitting, or removing a surface is a
change to this chapter through the Volume 0 change procedure.

#### Motivation

A fixed surface list keeps the PAL a deliberate contract instead of an accretion zone; the
per-surface Windows mapping keeps the v2 phase implementable without signature breaks
(PRD-011).

#### Actors

PAL maintainers; L3 consumers; Windows-phase implementers (future).

#### Preconditions

None beyond Volume 0 conventions.

#### Main flow

1. A consumer requests a platform operation via its surface.
2. The surface's active backend executes the OS mechanics.
3. Results return in portable types; failures map to E-PORT codes.

#### Alternative flows

- Capability absent: `Probe` reports it; operations fail with E-PORT-002 or degrade per the
  surface's declared policy (e.g., Clipboard falls back to OSC 52).

#### Edge cases

- Partially available capabilities (Landlock present but old ABI) report feature levels, not
  booleans, so the Sandbox Engine can select mechanisms per ADR-021.
- Fallback engagement (Config Directories) is itself an observable event
  (`pal.fallback.engaged`), never a silent substitution.

#### Inputs

Surface calls; host capabilities.

#### Outputs

Portable results; probe reports; E-PORT errors.

#### States

Not applicable — surfaces are stateless services with cached probe results.

#### Errors

E-PORT-001, E-PORT-002, E-PORT-003 (below), plus per-surface mappings into consumer families.

#### Constraints

One abstraction per surface — no parallel APIs for the same capability; no surface exposes
OS-specific types.

#### Security

Surfaces that create files, sockets, or processes apply owner-restrictive defaults (0600
sockets, 0700 temp dirs, filtered environments) as part of their contract, not as caller
options.

#### Observability

Probe results and degradations logged at startup and queryable in diagnostics.

#### Performance

Surface contracts state complexity/blocking behavior; blocking calls honor contexts
(FR-ARCH-004).

#### Compatibility

The Windows mapping column is reviewed at each phase gate; ADR-022's review condition
(validate known-folder mapping before the Windows phase) applies.

#### Acceptance criteria

- Given the PAL package, when its exported surfaces are enumerated, then exactly the 19
  surfaces of this chapter exist, each with `Probe`.
- Given any surface interface, when inspected, then no parameter or result type is
  OS-specific (no raw signal integers, mode bits as the only form, or platform path
  assumptions).
- Given the PAL conformance suite on each Tier 1 platform, when it runs, then every surface's
  behavioral contract passes identically or the divergence is documented in the platform
  matrix (SM-17's documentation rule).
- Negative case: given a host without a required capability for an invoked operation, when
  the operation runs, then it fails with E-PORT-002 (or degrades per declared policy) and
  emits `pal.capability.degraded` — never a silent no-op.

#### Verification method

PAL conformance suite per surface per Tier 1 platform (NFR-PORT-002); interface review
against the surface table; Windows-mapping review at phase gates.

#### Traceability

PRD-011; ADR-012, ADR-014, ADR-021, ADR-022, ADR-025; FR-PORT-001; NFR-PORT-002.

### FR-PORT-003 — Directory resolution through the PAL with XDG semantics

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Derived
- Owner: Architecture (Volume 3)
- Affected components: PAL (Config Directories), Configuration Manager, Persistence Layer, Logging, Updater, Workspace Engine
- Dependencies: ADR-022; FR-PORT-001
- Related risks: RISK-PORT-003

#### Description

All configuration, data, state, cache, log, and runtime directory locations MUST be resolved
exclusively through the PAL Config Directories surface, implementing ADR-022: adrg/xdg
resolution, XDG Base Directory semantics on Linux, Apple-native mapping on macOS honoring
explicitly set `XDG_*` overrides, project-local `.andromeda/`, and the documented
`~/.andromeda` fallback (or explicitly configured root) with a mandatory diagnostic on
engagement.

#### Motivation

One resolution authority produces one answer to "where are my files" per platform, enables
lifecycle-correct placement (cache vs data vs config), and keeps the Windows phase a mapping
exercise (ADR-022 rationale).

#### Actors

Storage-touching components; users setting `XDG_*`; support/diagnostics.

#### Preconditions

Process environment readable; ADR-022 library available.

#### Main flow

1. A component requests a directory role (config/data/state/cache/log/runtime) and scope.
2. The surface resolves per ADR-022 and returns the path with its resolution source.
3. The component uses the path; creation applies restrictive modes per surface contract.

#### Alternative flows

- Resolution failure (no home): the fallback root engages, `pal.fallback.engaged` is
  emitted, and resolution continues against the fallback layout.

#### Edge cases

- Partially set `XDG_*` variables produce mixed layouts by design (each variable honored
  independently, exactly as adrg/xdg resolves); diagnostics display every resolved path and
  its source (ADR-022 risk mitigation).
- The per-instance runtime directory (IPC sockets, chapter 08) requires user-private modes;
  on hosts without a suitable runtime dir the surface falls back to a 0700 directory under
  the state root.

#### Inputs

Directory role + scope requests; environment.

#### Outputs

Absolute paths with source attribution; fallback diagnostics.

#### States

Not applicable — stateless resolution with per-process caching.

#### Errors

E-PORT-002 when a required role cannot be resolved even via fallback (unwritable roots).

#### Constraints

Resolution results are stable within a process lifetime; changes require restart (watched
configuration does not relocate stores mid-run).

#### Security

Created roots get owner-only modes where they hold state or sockets; resolution never
follows attacker-controlled environment into privileged locations (paths are validated
against the surface's rules).

#### Observability

Diagnostics command output (Volume 8) lists every role's resolved path and source; fallback
engagement always evented.

#### Performance

Resolution is startup-time and cached; no per-operation cost.

#### Compatibility

Identical override behavior on macOS and Linux (`XDG_*` honored on both, per ADR-022);
mapping changes between library versions are treated as migrations (ADR-022 risk rule).

#### Acceptance criteria

- Given default environments on each Tier 1 platform, when paths are resolved, then they
  match the ADR-022 layout for that platform (golden tests).
- Given `XDG_CONFIG_HOME` explicitly set on macOS, when config paths resolve, then the
  override is honored.
- Given a container with no resolvable home, when Andromeda starts, then the fallback engages
  with the diagnostic and all roles resolve under the fallback root.
- Negative case: given production code outside the PAL, when scanned, then no hardcoded
  platform directory literals exist.

#### Verification method

Golden resolution tests per platform in the PAL conformance suite; container-environment CI
job; source scan for path literals (NFR-PORT-004 tooling).

#### Traceability

PRD-011; ADR-022, ADR-028; FR-PORT-001; chapter 09 deployment footprint.

### FR-PORT-004 — Platform support matrix conformance

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Architecture (Volume 3)
- Affected components: all; release pipeline (Volume 14); test strategy (Volume 13)
- Dependencies: FR-PORT-001, FR-PORT-002
- Related risks: RISK-PORT-001, RISK-PORT-002

#### Description

Andromeda MUST build, pass its acceptance suite, and be released for every Tier 1 platform of
the [platform support matrix](#platform-support-matrix), with behavior identical across
Tier 1 platforms except where the platform matrix documents a divergence. Tier 2 platforms
MUST build and pass smoke tests. Feature floors (kernel capabilities, minimum git) MUST be
probed at runtime with defined degradation or refusal (E-PORT-001/E-PORT-002) — never
undefined behavior on unsupported hosts.

#### Motivation

The matrix is the product's portability promise (PRD-011, MVP items 20–21); unprobed floors
turn "supported platform" into a lottery.

#### Actors

Release pipeline; test infrastructure; users on supported and unsupported hosts.

#### Preconditions

CI runners exist for every Tier 1 platform (macOS Intel capacity is the PENDING VALIDATION
gate for its Tier 2 commitment).

#### Main flow

1. Every release candidate builds for all Tier 1 targets.
2. The full acceptance suite runs per Tier 1 platform; smoke suite per Tier 2.
3. Divergences either fail the gate or are documented in the platform matrix with
   justification.

#### Alternative flows

- A Tier 1 platform gate fails: the release does not ship (Tier 1 definition, Volume 1).
- A Tier 2 failure: recorded as a defect; release proceeds.

#### Edge cases

- Runtime start on an unsupported OS/version/architecture: refuse with E-PORT-001 and
  remediation guidance, before touching any state.
- Supported OS lacking an optional capability: degrade per FR-PORT-002 policy.

#### Inputs

Build targets; suite results; host probes.

#### Outputs

Released artifacts per Tier 1 target; the divergence log in the platform matrix.

#### States

Not applicable — release-gating requirement.

#### Errors

E-PORT-001 (unsupported platform), E-PORT-002 (capability gaps).

#### Constraints

Matrix changes (adding/removing platforms, floors) go through the change procedure.

#### Security

Refusal-before-state-access on unsupported hosts prevents undefined security behavior
(untested keychain/permission semantics).

#### Observability

Per-platform suite results published per release; probe outcomes in startup diagnostics.

#### Performance

Volume 12 budgets apply per Tier 1 platform, on the reference hardware Volume 12 defines.

#### Compatibility

This requirement operationalizes SM-17 (owned by this volume) as a release gate.

#### Acceptance criteria

- Given a release candidate, when Tier 1 gates run, then 100% of the acceptance suite passes
  on macOS arm64, Ubuntu 22.04 x86_64/arm64, Debian 12, and Fedora 39 targets, or the release
  is blocked.
- Given `andromeda` started on macOS 12, when it starts, then it exits with E-PORT-001 and
  its mapped exit code, naming the minimum version.
- Given behavior that differs between Tier 1 platforms, when the divergence audit runs, then
  the divergence appears in the platform matrix or the audit fails (SM-17: 0 undocumented
  differences).
- Observability case: given any release, when its published artifacts are inspected, then
  per-platform suite results are attached.

#### Verification method

CI release gates per platform (Volume 13 operates them); SM-17 measurement; startup-refusal
tests in the PAL conformance suite.

#### Traceability

PRD-011; SM-17; Volume 1 platform scope; NFR-PORT-001; RISK-PORT-001, RISK-PORT-002.

### NFR-PORT-001 — Tier 1 behavioral parity

- Category: Portability
- Priority: P0
- Phase: MVP
- Metric: SM-17 — fraction of the acceptance suite passing identically on all Tier 1 platforms; count of undocumented behavioral differences
- Target: 100% identical pass; 0 undocumented differences
- Minimum threshold: 100% / 0 (Tier 1 definition admits no tolerance)
- Measurement method: full acceptance suite in CI on every Tier 1 platform per release; divergence audit against the platform matrix
- Test environment: Tier 1 CI runners (macOS arm64; Ubuntu 22.04, Debian 12, Fedora 39 on x86_64 and arm64)
- Measurement frequency: every release; tracked per mainline commit for Tier 1 build+test
- Owner: Architecture (Volume 3) / Volume 13 (suite operation)
- Dependencies: FR-PORT-004
- Risks: RISK-PORT-001
- Acceptance criteria: SM-17 targets met for every published release; any platform-specific behavior carries a platform-matrix entry.

### NFR-PORT-002 — PAL conformance coverage

- Category: Portability
- Priority: P0
- Phase: Core
- Metric: PAL surfaces with a behavioral conformance suite executed on all Tier 1 platforms, out of 19
- Target: 19/19 surfaces
- Minimum threshold: 19/19 for surfaces consumed by shipped features; a surface may ship suite-less only while no shipped feature consumes it
- Measurement method: conformance suite inventory check in CI (each surface maps to its suite; suites run per platform)
- Test environment: Tier 1 CI runners
- Measurement frequency: every mainline commit (suites), every release (inventory audit)
- Owner: Architecture (Volume 3)
- Dependencies: FR-PORT-002
- Risks: RISK-PORT-003
- Acceptance criteria: The inventory audit lists 19 surfaces with green suites per Tier 1 platform at each release; a new surface cannot merge without its suite.

### NFR-PORT-003 — Single-binary deliverable with bounded prerequisites

- Category: Portability
- Priority: P0
- Phase: MVP
- Metric: Count of runtime prerequisites beyond the documented list; count of dynamic library dependencies of the Linux binary
- Target: Prerequisites exactly: system git ≥ 2.40 for Git features (ADR-025) and optional platform services probed per FR-PORT-002; 0 dynamic library dependencies on Linux
- Minimum threshold: same as target (a new hard prerequisite requires a change-procedure amendment)
- Measurement method: binary inspection in the release pipeline (linkage check); clean-machine install-and-run test per Tier 1 platform without developer tooling
- Test environment: pristine VMs/containers per Tier 1 platform
- Measurement frequency: every release
- Owner: Architecture (Volume 3) / Volume 14 (pipeline)
- Dependencies: ADR-001, ADR-007, ADR-013, ADR-025
- Risks: RISK-PORT-001
- Acceptance criteria: Clean-machine runs succeed for non-Git features with no prerequisites; Git features fail with the defined configuration error when git is absent or below 2.40; the Linux release binary is statically linked.

### NFR-PORT-004 — Platform-conditional code containment

- Category: Maintainability
- Priority: P1
- Phase: Core
- Metric: Occurrences of prohibited platform constructs (FR-PORT-001 list) outside `internal/pal` in production code
- Target: 0
- Minimum threshold: 0
- Measurement method: automated scanner (lint rules + pattern scan per ADR-033) in CI on every PR; release audit
- Test environment: CI
- Measurement frequency: every PR and mainline commit
- Owner: Architecture (Volume 3)
- Dependencies: FR-PORT-001, ADR-033
- Risks: RISK-PORT-003
- Acceptance criteria: The scanner reports 0 violations on mainline; violating PRs cannot merge.

## Error codes

### E-PORT-001 — Unsupported platform

- Category: Environment
- Severity: Fatal
- User message: "This operating system, version, or architecture is not supported by this build."
- Technical message: detected OS/version/architecture versus the support matrix entry that failed
- Cause: process started on a host below the platform support matrix floors
- Safe-to-log data: OS name, OS version, architecture, binary target triple
- Recoverability: not recoverable on this host
- Retry policy: none
- Recommended action: use a supported platform per the matrix, or a build targeting this host if one exists
- Exit-code mapping: 3
- HTTP mapping: not applicable
- Telemetry event: `pal.platform.rejected`
- Security implications: refusal occurs before any state or credential access

### E-PORT-002 — Platform capability unavailable

- Category: Environment
- Severity: Error (or Warning when the surface's degradation policy applies)
- User message: "A required platform capability is unavailable on this system: <capability>."
- Technical message: surface, capability, probe result, required feature level
- Cause: missing OS facility (no PTY availability, no Secret Service, kernel below a feature floor, missing prerequisite binary)
- Safe-to-log data: surface name, capability name, probe details, feature level found/required
- Recoverability: recoverable by installing/enabling the facility or using a degraded mode where declared
- Retry policy: none automatic; re-probe on restart
- Recommended action: the surface-specific remediation named in the message (e.g., install git ≥ 2.40)
- Exit-code mapping: 3 when fatal to the invoked command; otherwise reported and degraded
- HTTP mapping: not applicable
- Telemetry event: `pal.capability.degraded`
- Security implications: degradations affecting containment are additionally surfaced through Sandbox Engine observability (ADR-021)

### E-PORT-003 — Credential store backend unavailable

- Category: Environment
- Severity: Error
- User message: "No usable credential store is available, and the encrypted file fallback is not enabled."
- Technical message: backend probe results (Keychain/Secret Service state), fallback enablement state
- Cause: OS keychain absent, locked, or unreachable; ADR-014 fallback not opted into
- Safe-to-log data: backend identifiers and probe outcomes; never key names or material
- Recoverability: recoverable by unlocking/installing the OS store or enabling the fallback
- Retry policy: none automatic
- Recommended action: unlock or configure the OS credential store, or explicitly enable the age-encrypted fallback (ADR-014)
- Exit-code mapping: 3 (environment); authentication flows blocked by it report through E-AUTH at exit code 4
- HTTP mapping: not applicable
- Telemetry event: `pal.capability.degraded`
- Security implications: the fallback is opt-in by design; this error MUST NOT auto-enable it

## Risks

### RISK-PORT-001 — macOS Intel viability gap

- Category: Platform / delivery
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Tier 2 classification (build + smoke only) pending validation of build/test capacity; identical PAL surfaces keep the port cost bounded; decision recorded via the open-questions register before MVP exit
- Detection: CI capacity monitoring; smoke-suite results on Intel builds
- Owner: Architecture (Volume 3) / Volume 14
- Status: Open — PENDING VALIDATION (register entry V3-OQ-1)

### RISK-PORT-002 — Linux isolation primitive fragmentation

- Category: Platform / security
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: ADR-021 layering — process-level controls are the guaranteed floor on every kernel; Landlock (≥ 5.13), namespaces, and bubblewrap are probed feature levels validated per distribution before any isolation claim; effective containment always observable
- Detection: PAL Sandbox surface probes across the reference-distribution matrix in CI; field diagnostics reporting containment levels
- Owner: Architecture (Volume 3) / Volume 9
- Status: Open — PENDING VALIDATION (register entry V3-OQ-2, aligned with ADR-021)

### RISK-PORT-003 — PAL abstraction leaks blocking the Windows phase

- Category: Technical
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: FR-PORT-001/FR-PORT-002 signature rules; per-surface Windows mapping maintained from day one; NFR-PORT-004 containment scanning; Windows-mapping review at every phase gate
- Detection: containment scanner; surface-signature review; a Windows-phase spike validating the mappings before v2 commitment (register entry V3-OQ-3)
- Owner: Architecture (Volume 3)
- Status: Open
