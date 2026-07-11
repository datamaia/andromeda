# 02 — Updater and Rollback

This chapter is the behavioral contract behind **UpdaterPort** (frozen signature, Volume 3
chapter 02): update checks and channel subscription (FR-REL-005), the download → verify →
apply pipeline (FR-REL-006), the automation policy (FR-REL-007), and rollback (FR-REL-008).
It also mints the `[update]` configuration keys, the `update.*`/`release.*` events, and the
E-REL error catalog. The Update process machine over the frozen states is chapter
[05](05-state-machines.md); the manual command surface (`andromeda update`, `update check`,
`update rollback`) is Volume 8's and is not restated here.

## UpdaterPort semantics

| Method | Contract elaborated here |
|---|---|
| `Check` | Queries release metadata for `[update].channel` from `[update].source`, applies the ADR-191 offer rule and the FR-REL-006 upgrade-path guard, and returns `up_to_date` or `update_available` (frozen state names). The only method that may use the network; MUST fail cleanly offline (E-REL-001) within `[update.timeouts].check_seconds`. Also reports when the *installed* version is yanked (INV-REL-03). |
| `Download` | Fetches the applicable artifact plus integrity metadata into the platform cache directory (staging), streaming resumable progress. Partial downloads resume by byte range where the source supports it; otherwise restart. Disk preflight guards staging (E-REL-010). |
| `Verify` | FR-REL-002 verification of staged files; produces the persisted verification report. `Apply` MUST refuse without a passed `Verify` for the same artifact set (port rule, Volume 3). |
| `Apply` | Consent-gated swap of the installed binary via the PAL Updater surface (atomic replace-or-restore), after retaining the current version per ADR-192. Runs the health probe (`andromeda version` of the new binary) before declaring success. Defers to the owning package manager for externally managed installs (E-REL-008, FR-REL-009). |
| `Rollback` | Offline restore of the retained version per ADR-192, including the schema-pairing dialogue. |

Machine-wide exclusivity: all mutating methods take the **update lock** (PAL File Locking
on a lock file in the platform runtime directory); a concurrent attempt fails with
E-REL-007. `Check` is concurrent-safe and lock-free. Cancellation follows FR-ARCH-004: any
step before the `applying` critical section aborts cleanly (E-REL-012); the swap itself is
the short, bounded, non-interruptible section whose outcome is always old-or-new, never
half-replaced.

Running instances: replacing the binary does not disturb already-running Andromeda
processes (they hold the previous file image until they exit); `Apply` reports this so
users know running sessions continue on the old version until restarted.

## Configuration

Key content minted here; schema, precedence, env-var mapping, and validation are Volume
10's (single-home matrix).

```toml
[update]
channel = "stable"           # stable | rc | beta | nightly (ADR-191)
source = "github"            # "github" or a mirror root path/URL (FR-REL-004)
auto_check = true            # scheduled + post-command checks (FR-REL-005)
check_interval_hours = 24
auto_download = false        # stage verified artifacts ahead of consent (FR-REL-007)
auto_apply = false           # opt-in unattended apply; never across majors (ADR-191)
notify = true                # surface update_available notices (Volume 8 renders)
keep_versions = 1            # retained previous versions (ADR-192)
signature_policy = "when_present"   # "when_present" | "required" (FR-REL-002)

[update.timeouts]
check_seconds = 30
download_seconds = 600
apply_seconds = 60
```

An invalid channel or malformed key fails configuration validation with Volume 10's E-CFG
semantics; the `--channel` flag overrides for one invocation only (Volume 8).

## Requirements

### FR-REL-005 — Update check, channel subscription, and notification

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Updater (Volume 14)
- Affected components: Updater, CLI, TUI, Event Bus
- Dependencies: ADR-191; FR-REL-001; `[update]` keys above
- Related risks: RISK-REL-001

#### Description

The Updater MUST evaluate release metadata for the configured channel with the ADR-191
offer rule and report exactly one of the frozen outcomes `up_to_date` or
`update_available`, including the offered version, channel, and whether the installed
version is yanked or out of support (FR-REL-014 guidance). With `auto_check` enabled,
checks run at most once per `check_interval_hours` — scheduled opportunistically when an
Andromeda process is running (never via self-installed daemons, Volume 3 process model) —
and results feed the Volume 8 post-command notice when `notify` is true. Checks MUST be
metadata-only: no artifact downloads, no side effects beyond persisting the check result
and refreshing local Release rows. Check requests MUST carry no identifying payload beyond
standard HTTP headers and the product version in the user agent.

#### Motivation

MVP item 23 starts with "check"; ADR-191's offer rule is only real if one component owns
its evaluation, and notification without consent-free side effects is the Safe by Default
posture applied to self-update.

#### Actors

The Updater; the CLI/TUI rendering notices; scheduled in-process checks.

#### Preconditions

Resolvable `[update].source`; for `github`, network reachability.

#### Main flow

1. `Check` fetches channel metadata from the source.
2. Local Release rows are refreshed (`release.metadata.refreshed`).
3. The offer rule and upgrade-path guard select the offered release or none.
4. The outcome persists to update history and emits `update.check.completed`.

#### Alternative flows

- Offline or unreachable source: E-REL-001 with the offline distinction (no network vs
  source error); scheduled checks retry at the next interval, never in a tight loop.
- Yanked installed version: outcome includes the yank notice; `release.yank.detected` is
  emitted once per learned yank.

#### Edge cases

- Channel switched to a less mature channel: no downgrade is offered (ADR-191 rule 3); the
  check reports `up_to_date` with the explanation.
- Clock skew does not affect correctness: offer evaluation is version-ordered, not
  time-ordered.

#### Inputs

Channel, source, installed version, local Release rows.

#### Outputs

Check outcome with offered version, yank/support notices; refreshed Release rows.

#### States

Update machine `checking` → `up_to_date` | `update_available` (chapter 05).

#### Errors

E-REL-001, E-REL-002.

#### Constraints

At most one scheduled check per interval per machine (global-DB bookkeeping); checks are
lock-free and MUST NOT block product operations.

#### Security

Metadata is untrusted input: parsing is bounded and schema-validated (E-REL-002 on
violation); the check path holds the `network` permission and nothing else.

#### Observability

`update.check.completed` / `update.check.failed`; check outcomes persisted in update
history with source and channel.

#### Performance

Bounded by `[update.timeouts].check_seconds`; the scheduled check MUST NOT add measurable
latency to foreground commands (it runs post-command / in background pools per ADR-023).

#### Compatibility

Outcome vocabulary is frozen (Volume 2); JSON shapes are Volume 8's structured-output
contract.

#### Acceptance criteria

- Given a newer applicable release on the channel, when `Check` runs, then
  `update_available` names exactly the ADR-191-selected version.
- Given no newer applicable release, when `Check` runs, then `up_to_date` is returned and
  no artifact bytes were fetched (egress capture).
- Given OS-level network disablement and `source = "github"`, when `Check` runs, then
  E-REL-001 returns within the check timeout with the offline cause.
- Permission case: given a policy denying `network`, when a scheduled check would run, then
  it is skipped and recorded as policy-skipped — never prompting from background context
  (PRD-009).
- Observability case: every check emits exactly one terminal check event carrying the
  correlation ID persisted in history.

#### Verification method

Contract tests over UpdaterPort `Check` with metadata fixtures (offer matrix, yank, skew);
offline suite; egress capture proving metadata-only behavior.

#### Traceability

MVP item 23; PRD-005, PRD-009; ADR-191; Volume 8 `update check` and post-command notice;
chapter 05 T1–T3.

### FR-REL-006 — Download, verification, and consent-gated apply

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Updater (Volume 14)
- Affected components: Updater, Permission Manager, PAL, Persistence Layer
- Dependencies: FR-REL-002, FR-REL-005; ADR-190, ADR-191, ADR-192; PermissionPort
- Related risks: RISK-REL-001, RISK-REL-003

#### Description

The update pipeline MUST proceed strictly `update_available` → consent → `downloading` →
`verifying` → `applying`, with these guards: (1) consent per Volume 8 confirmation rules
binds to the `system_modification` permission through PermissionPort and is recorded;
(2) the upgrade-path guard refuses releases whose `min_upgrade_from` excludes the installed
version (E-REL-009, naming the required stepping-stone version); (3) disk preflight covers
staging plus retention (E-REL-010); (4) `applying` retains the current version per ADR-192,
performs the PAL atomic swap, and health-probes the new binary; a failed swap or probe
auto-restores and terminates in `rolled_back`. Failures in any earlier step terminate in
`failed` with the installed version byte-identical to before. Externally managed
installations refuse with E-REL-008 and name the owning manager's command.

#### Motivation

This is MVP item 23's substance: an update path whose every failure mode ends in a defined,
recoverable state — the precedence of integrity over convenience applied to self-update.

#### Actors

Users (consent); the Updater; PermissionPort; the PAL Updater surface.

#### Preconditions

A passed check with `update_available`; update lock acquirable.

#### Main flow

1. Consent is obtained (interactive, `--yes`, or recorded policy) and persisted.
2. Artifacts stage into the cache directory with resumable progress events.
3. `Verify` passes per FR-REL-002.
4. Retention, swap, health probe; update history and events record `applied`.

#### Alternative flows

- Consent declined or unavailable non-interactively: the process terminates as chapter 05
  defines (reason `consent_declined`); nothing was downloaded unless `auto_download`
  staged it earlier.
- `auto_download` pre-staged artifacts: the flow resumes at `verifying` (staged bytes are
  re-verified — verification is never skipped because staging was earlier).

#### Edge cases

- Cancellation mid-download: staged partials persist for resume; the process instance
  terminates per chapter 05 cancellation rules (E-REL-012, exit code 8 at the CLI).
- Crash during `applying`: the PAL surface guarantees old-or-new; recovery on next start
  reconciles history (chapter 05 recovery) and the health record shows which binary
  survived.
- Target release yanked between check and apply: re-evaluation at `verifying` refuses with
  E-REL-011.

#### Inputs

Offered release, consent decision, staged artifacts, retention settings.

#### Outputs

New installed binary (or unchanged/restored one), verification report, update history row,
events.

#### States

`update_available` → `downloading` → `verifying` → `applying` → `applied` |
`rolled_back` | `failed` (chapter 05).

#### Errors

E-REL-003, E-REL-004, E-REL-005, E-REL-007..E-REL-012.

#### Constraints

One update operation machine-wide (update lock); verification precedes apply for the same
artifact set (UpdaterPort rule); migrations never run during apply (they run at the new
version's first start, ADR-029).

#### Security

Consent and every terminal outcome produce Audit Records; staged artifacts live in
disposable cache with restrictive permissions; the apply path never executes downloaded
code before verification passes (the health probe runs the swapped, verified binary).

#### Observability

`update.artifact.downloaded`, `update.artifact.verified`, `update.verification.failed`,
`update.version.applied`, `update.process.failed`, plus `update.state.changed` per
transition — all correlated to one update process ULID.

#### Performance

NFR-REL-001 (SM-18): ≤ 60 s p95 end-to-end on the reference network; ≤ 10 s p95 excluding
transfer.

#### Compatibility

Applies to every channel identically; artifact grammar per ADR-190; port signature frozen
(FR-ARCH-003).

#### Acceptance criteria

- Given consent and a pristine release, when the flow runs, then the binary is replaced,
  the previous version is retained, history shows `applied`, and the health probe output
  is recorded.
- Given a verification failure, when the flow runs, then it terminates `failed`, the
  installed binary is byte-identical to before (hash compare), and staged files are
  preserved for inspection.
- Given a simulated swap failure, when apply runs, then the prior binary is restored, the
  outcome is `rolled_back`, and `update.version.rolled_back` is emitted.
- Permission case: given no recorded consent path in non-interactive mode, when the flow
  reaches the gate, then it stops after reporting `update_available` (Volume 8 semantics)
  and no download occurs (unless pre-staged by policy).
- Negative case: given a concurrent update holding the lock, when a second apply starts,
  then E-REL-007 is returned without touching state.

#### Verification method

Volume 13 update suite: N−1 → N automated update per release (SM-18 harness),
fault-injection (network cut, kill −9 during each state, swap failure, tamper), lock
contention tests, egress capture for consent-before-download.

#### Traceability

MVP item 23; PRD-005, PRD-006, PRD-010; ADR-190/191/192, ADR-029; FR-REL-002; NFR-REL-001;
chapter 05 T4–T11.

### FR-REL-007 — Update automation policy

- Type: Functional
- Status: Draft
- Priority: P2
- Phase: Beta
- Source: Provided
- Owner: Updater (Volume 14)
- Affected components: Updater, Permission Manager, Configuration Manager
- Dependencies: FR-REL-005, FR-REL-006; ADR-191; Volume 9 `always_allow_policy` decision
- Related risks: RISK-REL-001

#### Description

Unattended automation MUST follow the ADR-191 ladder: `auto_check` (default on) may only
check; `auto_download` (default off) may additionally stage and verify artifacts;
`auto_apply` (default off) may complete the flow only when a recorded
`always_allow_policy` decision for `system_modification` exists, and MUST NOT cross a
major-version boundary — majors are always offered, never self-applied. Automated applies
run only at process start or idle (never mid-run), emit the same events and history as
manual updates, and honor the same guards (verification, upgrade path, lock, disk).

#### Motivation

Fleets need convergence without a human at every machine; the brief mandates automatic
updates. Binding automation to the recorded-policy mechanism keeps unattended self-
modification consented, auditable, and bounded.

#### Actors

Fleet administrators (policy); the Updater's scheduled paths.

#### Preconditions

Automation switches configured; for `auto_apply`, the recorded policy decision exists.

#### Main flow

1. A scheduled check finds `update_available`.
2. With `auto_download`, artifacts stage and verify in background pools.
3. With `auto_apply` and a same-major target and recorded policy, apply runs at the next
   process start/idle window and reports the outcome.

#### Alternative flows

- `auto_apply` set but no recorded policy decision: the apply is skipped and recorded as
  policy-blocked with remediation guidance — configuration alone MUST NOT self-apply.
- Major available under `auto_apply`: notification only, with the manual command named.

#### Edge cases

- A staged release superseded before apply: staging re-runs for the newer offer; stale
  staging is garbage-collected with the cache.
- Repeated automated failures (three consecutive `failed` outcomes for the same target)
  suspend auto-apply for that target and surface a persistent notice — no retry storms.

#### Inputs

Automation keys, recorded policy decisions, check outcomes.

#### Outputs

Staged artifacts; applied updates; policy-skip records.

#### States

Same machine as FR-REL-006; automated instances are flagged `automated` in history.

#### Errors

Same catalog; automation adds no new codes.

#### Constraints

Background work runs in ADR-023 supervised pools; automation never prompts (PRD-009
non-interactive discipline).

#### Security

The recorded policy decision is the consent artifact; audit records distinguish automated
from manual applies; disabling automation is always a plain configuration change.

#### Observability

History rows carry `automated: true` and the policy decision reference; the suspension
notice is an event (`update.process.failed` with suspension context).

#### Performance

Background staging respects Volume 12 pool budgets; automated applies must not delay
process start beyond the Volume 12 startup budget — apply defers rather than blocks.

#### Compatibility

Ladder semantics identical across platforms; package-manager-owned installs never
auto-apply (E-REL-008 deference).

#### Acceptance criteria

- Given `auto_download = true` only, when a check finds an update, then artifacts are
  staged and verified but the binary is untouched.
- Given `auto_apply = true` with the recorded policy and a same-major offer, when the
  window arrives, then the update applies with full history and events.
- Given `auto_apply = true` without the recorded policy, when the window arrives, then
  nothing applies and the policy-skip record names the missing decision.
- Negative case: given a major-version offer under full automation, when evaluated, then
  only a notification results.

#### Verification method

Policy-matrix integration tests (all 2³ switch combinations × policy present/absent ×
major/minor); idle-window scheduling tests; suspension behavior under injected repeated
failures.

#### Traceability

ADR-191; PRD-005, PRD-009; Volume 9 permission decisions vocabulary; FR-REL-005/006.

### FR-REL-008 — Rollback of the installed version

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Updater (Volume 14)
- Affected components: Updater, PAL, Persistence Layer
- Dependencies: ADR-192, ADR-029; FR-REL-006
- Related risks: RISK-REL-003

#### Description

`Rollback` MUST restore the retained previous version fully offline: re-verify the
retained artifact's recorded checksum, warn per ADR-192 when any database schema is newer
than the retained binary supports (proceeding only on explicit confirmation of the named
consequence), then atomically swap via the PAL surface and record the outcome. When no
retained version exists, rollback MUST refuse with E-REL-006 naming the air-gapped
reinstall procedure. Rollback MUST NOT touch databases: the paired database restore
(ADR-029 rule 6) is a separate, documented, explicitly-invoked procedure.

#### Motivation

SM-19 makes rollback a timed, offline product guarantee; a bad update must be reversible
without network and without silent data decisions.

#### Actors

Users via `andromeda update rollback` (Volume 8); the automatic apply-failure path
(FR-REL-006) reusing the mechanism.

#### Preconditions

A retained version exists; update lock acquirable.

#### Main flow

1. Retained artifact re-verified against its recorded checksum.
2. Schema comparison; confirmation dialogue when boundaries are crossed.
3. Atomic swap; health probe of the restored binary; history + events.

#### Alternative flows

- Schema-boundary confirmation declined: rollback aborts with no changes; the dialogue
  output names both the migration boundary and the paired-restore procedure.

#### Edge cases

- Retained artifact corrupted: E-REL-006 with reinstall guidance; the swap is never
  attempted with an unverified artifact.
- Rollback after rollback (flip-flop): the previously-current version becomes the retained
  one, so a second rollback returns forward; history keeps the full sequence.

#### Inputs

Retained artifact + metadata; user confirmation(s).

#### Outputs

Restored binary; update history row (`rolled_back`); schema-pairing outcome record.

#### States

Rollback instances traverse `checking` (local inventory only) → `applying` →
`rolled_back` | `failed` (chapter 05).

#### Errors

E-REL-006, E-REL-007, E-REL-012.

#### Constraints

Zero network I/O on the entire path (SM-19); binds `system_modification` with destructive
confirmation (Volume 8).

#### Security

Checksum re-verification before swap; audit record with both versions; the confirmation
text names data consequences explicitly — no silent data loss path.

#### Observability

`update.version.rolled_back`; history rows pair the rollback with the update it reverts.

#### Performance

NFR-REL-002 (SM-19): ≤ 30 s p95, offline.

#### Compatibility

Restored binary refuses newer schemas per ADR-029 rule 5 — defined refusal, not
corruption; the error names the remedy.

#### Acceptance criteria

- Given an applied update and OS-level network disablement, when rollback runs and is
  confirmed, then the previous binary is restored, verified, and history records
  `rolled_back` — within the NFR-REL-002 budget.
- Given a schema migration occurred after the update, when rollback is invoked, then the
  confirmation names the affected database(s) and the consequence, and declining changes
  nothing.
- Given no retained version, when rollback is invoked, then E-REL-006 returns with the
  reinstall procedure and nothing is modified.
- Observability case: the rollback history row references the reverted update's process
  ULID.

#### Verification method

SM-19 automated rollback test per release (offline harness); corruption and
schema-boundary fixtures; flip-flop sequence test; crash injection during the swap.

#### Traceability

SM-19; PRD-010; ADR-192, ADR-029; Volume 8 `update rollback`; chapter 05 T12–T14.

## Non-functional requirements

### NFR-REL-001 — Update time

- Category: Performance
- Priority: P1
- Phase: v1
- Metric: Wall-clock time of the full update path (check → download → verify → apply → report), N−1 to N, p95; and the same excluding download transfer time
- Target: ≤ 60 s p95 end-to-end on the Volume 1 reference network; ≤ 10 s p95 excluding transfer
- Minimum threshold: ≤ 60 s / ≤ 10 s p95 (SM-18; weakening requires the Volume 0 change procedure; measured non-gating from MVP, gating at v1)
- Measurement method: automated update test from release N−1 to N in the Volume 13 release suite, instrumented per state, ≥ 20 iterations per platform
- Test environment: Volume 12 reference machines; Volume 1 reference network (50 Mbit/s, 40 ms RTT)
- Measurement frequency: every release; trend dashboards from MVP
- Owner: Updater (Volume 14) with Volume 12 harness
- Dependencies: FR-REL-006; SM-18
- Risks: RISK-REL-001
- Acceptance criteria: Release qualification (Volume 13) shows p95 within target on all Tier 1 platforms; per-state timings persisted in the harness output so regressions localize to a state.

### NFR-REL-002 — Rollback time

- Category: Reliability
- Priority: P1
- Phase: v1
- Metric: Wall-clock time to restore the previously installed version from retained artifacts, p95, with all network interfaces disabled
- Target: ≤ 30 s p95, executable fully offline
- Minimum threshold: ≤ 30 s p95 offline (SM-19; weakening requires the Volume 0 change procedure; measured non-gating from MVP, gating at v1)
- Measurement method: automated rollback test per release under OS-level network disablement, ≥ 20 iterations per platform, including verification and health probe
- Test environment: Volume 12 reference machines, offline condition (Volume 1)
- Measurement frequency: every release
- Owner: Updater (Volume 14) with Volume 12 harness
- Dependencies: FR-REL-008; ADR-192; SM-19
- Risks: RISK-REL-003
- Acceptance criteria: p95 within target on all Tier 1 platforms with zero network-access attempts observed (egress capture); failed iterations count as misses, not exclusions.

## Risks

### RISK-REL-001 — Failed or interrupted update leaves an unusable installation

- Category: Reliability
- Probability: Low
- Impact: High
- Severity: High
- Mitigation: verify-before-apply (FR-REL-002/006); PAL atomic replace-or-restore with old-or-new guarantee; retained version + automatic restore on swap/probe failure (ADR-192); migrations deferred to first start with ADR-029 backups; crash-recovery reconciliation (chapter 05)
- Detection: health probe after swap; recovery pass on next start marking interrupted processes; SM-18/SM-19 suites with kill −9 injection at every state
- Owner: Updater (Volume 14)
- Status: Open

The one outcome distribution must never produce is a machine without a working `andromeda`.
Every layer (verification, atomic swap, retention, probe, recovery) exists to bound this;
the residual risk concentrates in platform filesystem semantics, covered by PAL golden
tests per platform.

### RISK-REL-003 — Binary–database version skew after rollback or workspace sync

- Category: Data integrity
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: ADR-029 clean refusal of future schemas (never partial reads); ADR-192 schema-pairing dialogue at rollback time with recorded outcome; documentation of the paired-restore procedure; version-alignment guidance for synced workspaces (ADR-029 negative consequences)
- Detection: refusal errors carry both versions and are counted in local metrics; support-channel monitoring
- Owner: Updater (Volume 14) / Persistence Layer (Volume 10 procedures)
- Status: Open

Skew is a certainty (rollback, multi-machine sync), so the posture is defined refusal plus
an explicit, named recovery path — the risk is user frustration, not corruption.

## Error catalog (E-REL)

Every code follows the ADR-016 envelope. Exit codes per Volume 0 chapter 03; the update
surface is CLI-driven, so HTTP mappings apply only where the ADR-012 IPC surface exposes
update operations (headless instances).

### E-REL-001 — Update check failed

- Category: Network
- Severity: Warning (scheduled) / Error (explicit command)
- User message: "Could not check for updates: <cause: offline | source unreachable | source error>."
- Technical message: source kind and root, HTTP status or dial error class, timeout used
- Cause: no network, unreachable source, source-side failure, mirror path missing
- Safe-to-log data: source kind, sanitized root, error class, latency — never response bodies
- Recoverability: recoverable — connectivity or source repair
- Retry policy: scheduled checks retry next interval; explicit commands do not auto-retry
- Recommended action: verify connectivity or `[update].source`; air-gapped sites confirm mirror mount
- Exit-code mapping: 1
- HTTP mapping: 502 over IPC
- Telemetry event: `update.check.failed`
- Security implications: none; check carries no sensitive payload

### E-REL-002 — Release metadata invalid

- Category: Validation
- Severity: Error
- User message: "The update source returned release information Andromeda cannot use."
- Technical message: schema violation or unknown newer `schema_version` (both versions named), offending field, source
- Cause: malformed or truncated metadata; mirror index built for a newer product
- Safe-to-log data: source kind, schema versions, field path — never full payloads
- Recoverability: recoverable after source repair or product upgrade (schema-skew case)
- Retry policy: none automatic (deterministic until source changes)
- Recommended action: rebuild the mirror with the documented procedure, or upgrade via a supported path
- Exit-code mapping: 1
- HTTP mapping: 502 over IPC
- Telemetry event: `update.check.failed`
- Security implications: metadata is untrusted input; parsing is bounded and rejects rather than guesses

### E-REL-003 — Artifact download failed

- Category: Network
- Severity: Error
- User message: "Downloading the update failed: <cause>."
- Technical message: artifact name, bytes transferred/expected, resume attempts, error class
- Cause: connection loss, source error, out-of-range resume
- Safe-to-log data: artifact name, byte counts, attempt count
- Recoverability: recoverable — staged partials support resume
- Retry policy: up to 3 automatic resume attempts with exponential backoff within `download_seconds`; then fail
- Recommended action: re-run the update; persistent failures: check source or use a mirror
- Exit-code mapping: 1
- HTTP mapping: 502 over IPC
- Telemetry event: `update.process.failed`
- Security implications: partial files stay in cache staging and are never executed or applied

### E-REL-004 — Artifact verification failed

- Category: Security
- Severity: Critical
- User message: "The downloaded update failed integrity verification and was not applied."
- Technical message: artifact, expected vs computed digest, signature status and policy, source
- Cause: corruption in transit, stale mirror, tampering, unsigned release under `required` policy
- Safe-to-log data: artifact name, digests, signature status, policy, source kind
- Recoverability: recoverable by re-download from a trusted source; policy-caused refusals by policy change or a signed release
- Retry policy: none automatic — verification failures never auto-retry
- Recommended action: re-run against the canonical source; if it recurs, report through the security disclosure channel (Volume 9)
- Exit-code mapping: 9
- HTTP mapping: 502 over IPC
- Telemetry event: `update.verification.failed`
- Security implications: possible supply-chain attack; offending files preserved for inspection; Audit Record mandatory

### E-REL-005 — Update apply failed

- Category: Integrity
- Severity: Error
- User message: "Applying the update failed; your previous version was restored."
- Technical message: failing step (retention, swap, health probe), PAL error chain, restore outcome
- Cause: filesystem errors during swap, probe failure of the new binary
- Safe-to-log data: step, paths (sanitized), probe output hash
- Recoverability: recovered automatically (restore); the target release should be treated as suspect
- Retry policy: none automatic; manual re-run permitted
- Recommended action: re-run once; if the probe failed, wait for a fixed release
- Exit-code mapping: 1
- HTTP mapping: 500 over IPC
- Telemetry event: `update.version.rolled_back`
- Security implications: none beyond RISK-REL-001 handling; restore path is audit-logged

### E-REL-006 — Rollback failed or unavailable

- Category: Integrity
- Severity: Error
- User message: "Rollback is not possible: <no previous version is retained | the retained version failed verification | restore failed>."
- Technical message: retention inventory state, checksum result, PAL error chain where applicable
- Cause: fresh install (nothing retained), pruned retention, corrupted retained artifact, swap failure during restore
- Safe-to-log data: retained version id, checksum status, failing step
- Recoverability: recoverable via air-gapped reinstall of the desired version (documented procedure)
- Retry policy: none automatic
- Recommended action: follow the reinstall procedure; for restore failures run `andromeda doctor`
- Exit-code mapping: 1 when unavailable; 9 when a restore attempt left verification-failed state
- HTTP mapping: 500 over IPC
- Telemetry event: `update.process.failed`
- Security implications: corrupted retained artifacts are preserved and reported, never swapped in

### E-REL-007 — Update already in progress

- Category: Concurrency
- Severity: Error
- User message: "Another update operation is already running on this machine."
- Technical message: lock path, holder PID and start time where readable
- Cause: concurrent update/rollback (second CLI invocation, automation overlap)
- Safe-to-log data: lock path, holder PID, holder age
- Recoverability: recoverable — wait for completion
- Retry policy: none automatic; stale locks (holder dead) are reclaimed by the recovery pass
- Recommended action: wait; if the holder is dead, re-run (recovery reclaims the lock)
- Exit-code mapping: 1
- HTTP mapping: 409 over IPC
- Telemetry event: `update.process.failed`
- Security implications: none

### E-REL-008 — Externally managed installation

- Category: Configuration
- Severity: Error
- User message: "This installation is managed by <manager>; update it with: <command>."
- Technical message: detected manager and evidence (path/query per FR-REL-009)
- Cause: self-update invoked on a Homebrew- or package-managed install
- Safe-to-log data: manager name, detection evidence class
- Recoverability: recoverable via the owning manager
- Retry policy: none
- Recommended action: run the named manager command (e.g., `brew upgrade andromeda`)
- Exit-code mapping: 1
- HTTP mapping: 409 over IPC
- Telemetry event: `update.process.failed`
- Security implications: prevents split-brain ownership of the binary path

### E-REL-009 — Unsupported upgrade path

- Category: Validation
- Severity: Error
- User message: "Direct update from <installed> to <target> is not supported; update to <stepping-stone> first."
- Technical message: installed version, target, `min_upgrade_from`, computed stepping-stone chain
- Cause: installation older than the target's `min_upgrade_from` (FR-REL-014 path rule)
- Safe-to-log data: all three versions
- Recoverability: recoverable via stepped updates
- Retry policy: none
- Recommended action: run the named stepping-stone update(s) in order
- Exit-code mapping: 1
- HTTP mapping: 409 over IPC
- Telemetry event: `update.check.completed` (path-refusal outcome)
- Security implications: none; prevents untested migration jumps (ADR-029 discipline)

### E-REL-010 — Insufficient disk space for update

- Category: Resource
- Severity: Error
- User message: "Not enough disk space to stage the update safely (<needed> required, <available> free)."
- Technical message: staging path, needed vs available bytes (staging + retention + migration-backup headroom)
- Cause: full cache or data volume
- Safe-to-log data: byte figures, volume identifier
- Recoverability: recoverable after freeing space
- Retry policy: none automatic
- Recommended action: free space (cache locations are documented and disposable) and re-run
- Exit-code mapping: 1
- HTTP mapping: 507 over IPC
- Telemetry event: `update.process.failed`
- Security implications: none; preflight prevents mid-apply exhaustion

### E-REL-011 — Release yanked

- Category: Validation
- Severity: Error
- User message: "Version <v> was withdrawn (<reason>) and cannot be installed."
- Technical message: release version, yank reason, detection point (check re-evaluation or verify-time)
- Cause: target release yanked between offer and apply, or a stale mirror serving a yanked release
- Safe-to-log data: version, reason, source kind
- Recoverability: recoverable — a newer release supersedes
- Retry policy: none; the next check offers the replacement
- Recommended action: re-run `andromeda update` to receive the superseding release
- Exit-code mapping: 1
- HTTP mapping: 410 over IPC
- Telemetry event: `release.yank.detected`
- Security implications: enforces INV-REL-03 at the last possible gate

### E-REL-012 — Update step timed out or cancelled

- Category: Timeout
- Severity: Error
- User message: "The update was <cancelled | timed out> during <step>; your installed version is unchanged."
- Technical message: step, elapsed vs configured timeout, cancellation origin (user, deadline, shutdown)
- Cause: `[update.timeouts]` exceeded, user interrupt, process shutdown
- Safe-to-log data: step, timings, origin
- Recoverability: recoverable — re-run resumes staged downloads
- Retry policy: none automatic for explicit commands; scheduled automation retries next interval
- Recommended action: re-run; raise the relevant timeout for slow links or use a mirror
- Exit-code mapping: 8
- HTTP mapping: 504 over IPC
- Telemetry event: `update.process.failed`
- Security implications: cancellation never interrupts the swap critical section (FR-ARCH-004 edge rule)

## Events

Envelope, ordering, delivery, persistence, retention, and redaction per Volume 10; names
follow the Volume 0 grammar. Payloads are content-free: versions, channels, digests,
counts, ULIDs.

| Event | Emitted by | Meaning |
|---|---|---|
| `update.check.completed` | Updater | A check finished with `up_to_date` or `update_available` (FR-REL-005) |
| `update.check.failed` | Updater | A check failed (E-REL-001/002) |
| `update.artifact.downloaded` | Updater | An artifact finished staging (FR-REL-006) |
| `update.artifact.verified` | Updater | Verification passed for the staged set (FR-REL-002) |
| `update.verification.failed` | Updater | Verification failed (E-REL-004) |
| `update.version.applied` | Updater | The new version is active (`applied`) |
| `update.version.rolled_back` | Updater | A restore completed — automatic or manual (`rolled_back`) |
| `update.process.failed` | Updater | An update process terminated in `failed` |
| `update.state.changed` | Updater | Any Update machine transition (chapter 05) |
| `release.metadata.refreshed` | Updater | Local Release rows were refreshed from the source |
| `release.yank.detected` | Updater | The installed or targeted version was learned to be yanked |
