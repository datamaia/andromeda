# 03 — Installation, Uninstallation, and Data

This chapter specifies what installation leaves on a machine and how it is detected
(FR-REL-009), what uninstallation removes and — more importantly — preserves by default
(FR-REL-010), and the explicit data-removal procedure (FR-REL-011). The authoritative state
footprint (which files exist where, per ADR-022/ADR-028) is Volume 3 chapter 09's; this
chapter owns the install/uninstall/cleanup *procedures* over that footprint.

## Ownership model

Every installation has exactly one **installation owner**, detected through the PAL
Installer surface:

| Owner | Detection | Self-update behavior |
|---|---|---|
| `self` (installer script or manual placement) | Binary path outside manager-owned trees; no manager records the path | Full FR-REL-006 flow |
| `homebrew` | Binary resolves into the Homebrew prefix | Defer: E-REL-008 names `brew upgrade andromeda` |
| `package` (deb/rpm/apk) | The owning package manager's file-ownership query claims the binary path | Defer: E-REL-008 names the manager command |

Detection runs at Updater invocation time (no cached guess can go stale) and its evidence
is recorded in update history. The uninstall procedure is likewise owner-specific: the
channel that placed the binary is the channel that removes it.

## Data classes at uninstall time

| Class | Location (per FR-PORT-003 resolution) | Default at uninstall |
|---|---|---|
| Binary + retained versions | Install dir; Updater area of data dir | Removed |
| Global configuration (`andromeda.toml`) | Config dir | **Preserved** |
| Global database, update history | Data dir | **Preserved** |
| Logs | Log location | **Preserved** |
| Caches, staged downloads | Cache dir | Removed (disposable by definition) |
| Secret material | OS credential store / age fallback file (ADR-014) | **Preserved**; removal is a listed manual step |
| Workspace `.andromeda/` directories | Workspace roots | **Never touched** by any product-level procedure |

## Requirements

### FR-REL-009 — Installation layout and ownership detection

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Derived
- Owner: Updater (Volume 14)
- Affected components: Updater, PAL, CLI
- Dependencies: FR-REL-003; FR-PORT-003; ADR-022
- Related risks: RISK-REL-001

#### Description

Andromeda MUST operate from a single installed binary whose runtime footprint is exactly
the Volume 3 chapter 09 state map — installation MUST NOT require privileged directories,
daemons, or registrations beyond placing the binary on `PATH`. The Updater MUST detect the
installation owner (`self`, `homebrew`, `package`) via the PAL Installer surface at every
mutating invocation, record the evidence, and route self-update per the ownership model
above. First run initializes the global footprint lazily (directories and databases are
created on demand, never by the installer).

#### Motivation

Uninstallable-by-deletion software is the honest counterpart of Safe by Default; ownership
detection prevents two updaters (Andromeda's and a package manager's) fighting over one
binary path.

#### Actors

Installers (script, brew, packages); the Updater; `andromeda doctor` reporting the detected
owner.

#### Preconditions

Binary present and executable; PAL directory resolution functional (fallback per ADR-022).

#### Main flow

1. First run resolves directories via FR-PORT-003 and creates state lazily.
2. Any Updater mutating call detects ownership and records evidence.
3. Self-owned installs proceed; manager-owned installs defer with E-REL-008.

#### Alternative flows

- Fallback layout (`~/.andromeda`, ADR-022): all procedures operate on the resolved paths;
  documentation presents per-platform and fallback tables generated from the same
  resolution rules.

#### Edge cases

- Binary copied out of a Homebrew prefix into `~/.local/bin`: detection follows the
  *current* binary path — the copy is `self`-owned; the stale brew install is reported by
  `doctor` as a duplicate-installation diagnostic.
- Multiple binaries on `PATH`: the running binary's own path decides; `doctor` warns about
  shadowed duplicates.

#### Inputs

Binary path; manager query results; resolved directories.

#### Outputs

Ownership record in update history; lazily created footprint.

#### States

None beyond first-run initialization; no installation state machine exists (deliberately —
placement is not a process Andromeda controls).

#### Errors

E-REL-008 (deference); directory-resolution failures surface via the PAL's E-PORT family.

#### Constraints

No setuid, no privileged helpers, no login items/units (Volume 3 process model); the binary
is the only mandatory file.

#### Security

Lazily created directories use restrictive permissions (0700-class per PAL Temporary
Files/Permissions surfaces); ownership evidence contains paths only, never package-manager
credentials.

#### Observability

Ownership and layout are visible in `andromeda doctor` (Volume 8) and recorded in update
history rows.

#### Performance

Ownership detection is two local queries; it MUST NOT add measurable latency to `update
check` (bounded by NFR-REL-001's non-transfer budget).

#### Compatibility

Detection heuristics are platform-specific inside the PAL; behavior contract identical on
all Tier 1 platforms.

#### Acceptance criteria

- Given a script-installed binary, when `update` runs, then ownership is `self` and the
  flow proceeds.
- Given a brew-installed binary, when `update` runs, then E-REL-008 names the brew command
  and nothing is modified.
- Given a fresh install, when the first command runs, then only the directories that
  command needed exist (lazy initialization verified by filesystem snapshot).
- Negative case: given a machine without resolvable home (container), when Andromeda runs,
  then the ADR-022 fallback engages with its diagnostic and procedures still resolve.

#### Verification method

Per-owner installation fixtures in the Volume 13 installation matrix; filesystem snapshot
tests for lazy initialization; duplicate/shadow fixtures for `doctor` diagnostics.

#### Traceability

MVP item 22; PRD-011; ADR-022; FR-PORT-003; Volume 3 chapter 09 state footprint; E-REL-008.

### FR-REL-010 — Uninstallation with data preservation by default

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Updater (Volume 14)
- Affected components: Updater, PAL
- Dependencies: FR-REL-009; ADR-022
- Related risks: RISK-REL-004

#### Description

Each installation channel MUST provide a documented uninstall procedure that removes the
binary, retained versions, and caches, and **preserves** configuration, databases, logs,
secret material, and every workspace `.andromeda/` directory. Concretely: `brew uninstall
andromeda`; package-manager removal for deb/rpm/apk; `install.sh --uninstall` for
script/manual installs (removes exactly what the installer placed plus retained versions
and caches). Every procedure MUST end by stating what was preserved and where, and how to
remove it (FR-REL-011). Uninstall MUST NOT require the product to be functional (a broken
installation is a primary uninstall scenario).

#### Motivation

Users uninstall for reasons that include reinstallation and migration; destroying state by
default converts a reversible action into data loss — the precedence order puts integrity
above tidiness.

#### Actors

Users; the channel's uninstall mechanism.

#### Preconditions

None (works on broken installs by design).

#### Main flow

1. The owner-specific procedure removes binary, retained versions, caches.
2. The procedure prints the preserved-data statement with resolved paths.

#### Alternative flows

- Package purge modes (`dpkg --purge` class): the packages install only the binary, so
  purge semantics equal removal; user data is user-owned files the package never claimed —
  stated in package metadata descriptions.

#### Edge cases

- Uninstall while an Andromeda process runs: removal of the binary file does not kill
  running processes (POSIX inode semantics); the procedure warns and lists running
  instances when detectable.
- Repeated uninstall: idempotent — absent files are skipped silently, the preserved-data
  statement still prints.

#### Inputs

Installed footprint; channel-specific uninstall invocation.

#### Outputs

Binary-free machine with intact user data; preserved-data statement.

#### States

None.

#### Errors

Script-level exit codes with remediation text; no runtime E-codes (the product may be
absent).

#### Constraints

No procedure may touch workspace directories or the OS credential store; cache removal is
unconditional (caches are rebuildable, ADR-028).

#### Security

Secret material is never deleted implicitly — keychain entries outlive uninstall and the
statement names them (removal instructions in FR-REL-011); no procedure elevates by
default.

#### Observability

Pre-runtime: the procedure transcript is the record; nothing product-side remains to log.

#### Performance

Not applicable (file removal).

#### Compatibility

Identical preservation semantics across channels and platforms; fallback-layout paths
included in the statement.

#### Acceptance criteria

- Given each installation channel on a Tier 1 platform, when its uninstall procedure runs,
  then the binary, retained versions, and caches are gone and config/databases/logs/
  secrets/workspaces are byte-identical (snapshot compare).
- Given a broken installation (corrupted binary), when uninstall runs, then it completes
  and preserves data identically.
- Given a second uninstall run, when it completes, then it exits successfully with the
  statement and no error.
- Negative case: given a workspace under the user's home, when uninstall runs, then its
  `.andromeda/` directory is untouched (explicit test).

#### Verification method

Uninstall matrix in Volume 13 (per channel × per platform) with before/after filesystem
snapshots; broken-install and idempotency fixtures.

#### Traceability

Brief: uninstallation and data preservation; ADR-022; Volume 3 chapter 09 uninstall note;
FR-REL-011.

### FR-REL-011 — Explicit data removal

- Type: Functional
- Status: Draft
- Priority: P2
- Phase: MVP
- Source: Provided
- Owner: Updater (Volume 14)
- Affected components: Updater, PAL, Secret Store
- Dependencies: FR-REL-010; ADR-014, ADR-022
- Related risks: RISK-REL-004

#### Description

A documented **purge procedure** MUST allow complete removal of Andromeda's global
footprint after (or with) uninstall: `install.sh --uninstall --purge` removes the resolved
config, data, state, log, and cache directories after an explicit confirmation that lists
every path it will delete; it then prints the two removals it deliberately does not
perform, with exact instructions: (a) secret material in the OS credential store (or the
age fallback file's path), removed via the platform's credential manager per the
documented service/item names; (b) per-workspace `.andromeda/` directories, removed by
deleting them inside each workspace. The documentation MUST also provide the manual path
list for users who never used the installer script.

#### Motivation

Complete removal is a legitimate user right and an enterprise off-boarding requirement;
doing it exactly once, with named paths and explicit consent, prevents both residue and
overreach.

#### Actors

Users; enclave/fleet administrators.

#### Preconditions

Resolved paths known (script resolves them the same way the product does, including the
ADR-022 fallback).

#### Main flow

1. `--purge` prints the full deletion list and requires explicit confirmation.
2. Listed directories are removed.
3. The credential and workspace statements print with exact instructions.

#### Alternative flows

- Purge without prior uninstall: performs uninstall first, then purge (one confirmation
  covering both).
- Manual purge: documentation table of per-platform paths (default and fallback layouts)
  with removal commands.

#### Edge cases

- Custom `XDG_*` overrides: the script honors the same resolution as the product (ADR-022)
  — purge lists the *actual* resolved paths, never hardcoded defaults.
- Partially removed footprint: idempotent; missing paths are reported as already absent.

#### Inputs

Confirmation; resolved path set.

#### Outputs

Empty global footprint; printed credential/workspace instructions.

#### States

None.

#### Errors

Script-level; refusal to proceed without confirmation is the default (non-interactive purge
requires the script's explicit `--yes`).

#### Constraints

Purge MUST NOT recurse into any path outside the resolved footprint; workspace and
credential deletion are never performed by the script.

#### Security

Deleting databases does not shred blocks (documented honestly); users with hostile-disk
threat models are pointed to full-disk encryption guidance rather than promised secure
erasure. Credential names are documented so keychain cleanup is verifiable.

#### Observability

Transcript lists every deleted path; nothing remains to emit product telemetry (and none is
sent — telemetry is consent-gated and local-first, Volume 10).

#### Performance

Not applicable.

#### Compatibility

Path tables cover macOS default, Linux XDG, explicit `XDG_*` overrides, and the ADR-022
fallback.

#### Acceptance criteria

- Given a used installation, when `--uninstall --purge` is confirmed, then the resolved
  config/data/state/log/cache directories are absent, and the transcript lists each removed
  path plus the credential and workspace instructions.
- Given declined confirmation, when purge runs, then nothing is deleted.
- Given `XDG_DATA_HOME` overridden, when purge runs, then the overridden path is the one
  listed and removed.
- Negative case: given a workspace directory and populated keychain entries, when purge
  completes, then both are untouched and their instructions were printed.

#### Verification method

Purge fixtures in the Volume 13 installation matrix with filesystem and (mocked) credential
store snapshots; override-layout and idempotency fixtures; documentation walkthrough in CI.

#### Traceability

Brief: data removal and cleanup; ADR-014, ADR-022; FR-REL-010; Volume 9 credential storage
model (referenced, not restated).
