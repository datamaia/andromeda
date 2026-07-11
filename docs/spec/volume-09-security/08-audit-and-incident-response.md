# 08 — Audit and Incident Response

This chapter owns the behavioral semantics of the **Audit Log**: the closed audited-action
catalog, the hash-chain verification procedure and tamper response, retention, and export.
It then defines security event emission, the incident-response procedure for users and
operators, and recovery. The Audit Record entity — attributes, hash chaining, invariants
INV-AUD-01..05 — is Volume 2's (chapter 08); this chapter supplies everything Volume 2
delegates: the action catalog, verification, tamper handling, retention windows, and
safe-detail rules. Project-side vulnerability handling and the coordinated disclosure policy
live with open-source governance in Volume 15; this chapter binds the product-side hooks and
the response-time commitment (SM-16 c).

## Audited-action catalog

The closed catalog of actions that MUST each produce exactly one Audit Record
(INV-AUD-03), including denied and failed outcomes. Action names are stable identifiers used
in the `action` attribute:

| Action | Trigger | Producer |
|---|---|---|
| `permission.decided` | every permission evaluation that decides an action (chapter 05) | Permission Manager |
| `permission.grant.minted` | a Permission row created | Permission Manager |
| `permission.grant.revoked` | revocation or expiry of a grant | Permission Manager |
| `approval.resolved` | an Approval reaches a terminal state (chapter 09) | Permission Manager |
| `secret.stored` | Secret Store `Set` | Secret Store |
| `secret.accessed` | Secret Store `Get` (all outcomes) | Secret Store |
| `secret.deleted` | Secret Store `Delete` | Secret Store |
| `secret.fallback.enabled` | fallback consent granted (chapter 07) | Secret Store |
| `secret.orphan.swept` | orphan sweep repair action | Secret Store |
| `credential.lifecycle.changed` | Credential status change (rotation, revocation, expiry — flows per Volume 5) | Authentication Layer |
| `sandbox.violation` | blocked policy violation (chapter 06) | Sandbox Engine |
| `sandbox.degraded` | execution proceeded below requested isolation | Sandbox Engine |
| `sandbox.orphan.swept` | startup sweep of dead-incarnation sandbox state | Sandbox Engine |
| `tool.invocation.denied` | invocation denied by permission or trust policy | Tool Runtime |
| `git.mutation.performed` | any mutating Git operation (Volume 11 gates) | Git Engine |
| `command.executed` | terminal command execution concluded (any outcome) | Terminal Engine |
| `package.changed` | extension/package install, update, or removal (Volume 6) | Package Manager |
| `update.applied` | product update apply or rollback (Volume 14) | Updater |
| `config.security_key.changed` | effective change to `[permissions]`, `[sandbox]`, or `[security]` keys | Configuration Manager |
| `audit.verified` | chain verification run (any outcome) | Audit Log |
| `audit.retention.applied` | post-window archival/pruning pass | Audit Log |
| `incident.recorded` | user or system opened/closed an incident record | Audit Log |

Catalog rules: the list is closed — additions arrive through this volume's change procedure
(they are additive and expected as new subsystems land in later phases, e.g., Volume 11
hosting integrations). Every producer feeds records through one Audit Log writer per
database; the writer assigns the chain position. `detail` payloads follow safe-to-log rules
(chapter 07 redaction; INV-AUD-05): entity IDs, rule names, resolved paths
(workspace-relative where possible), outcome classifications — never secret material, never
raw file or model content.

## Write discipline and fail-closed rule

An audited action and its Audit Record are inseparable (ADR-125):

1. The record is appended in the same transactional boundary as the action's own state
   change where both live in one database; where they cannot share a transaction (global vs
   workspace databases), the record is appended *before* the action's effects are allowed to
   proceed.
2. If the audit append fails, the action fails with E-SEC-014 — no audited action executes
   unrecorded. Read-only operations that are not in the catalog are unaffected.
3. Audit writes are never buffered across process boundaries and never dropped under
   backpressure; they are the one write class exempt from shed policies (Volume 10 event
   overflow rules explicitly do not apply — Audit Records are rows first, events second).

## Chain verification and tamper response

Each database (workspace, global) maintains its own chain (Volume 2): `record_hash` =
SHA-256 over the canonical serialization including `prev_hash`; first record uses the
zero-hash sentinel.

**Verification procedure** (ADR-123):

1. **Head check** — on database open: recompute the newest record's hash and its link to its
   predecessor. Cost O(1); default per `security.audit_verify_on_open = "head"`.
2. **Full verification** — walk the chain from the sentinel, recomputing every link; run by
   `andromeda doctor`, by the scheduled verification task, before audit export, and after
   any restore from backup. Reports records verified, elapsed time, and the first divergence
   if any.
3. **Divergence handling** — a broken link is an integrity error (E-SEC-013, exit code 9
   class): the log is marked tampered at the first divergent position; verification
   continues past the break to report every subsequent consistent segment (evidence
   preservation), and an `incident.recorded` audit action opens an incident on the *other*
   (intact) chain where possible.

**Tamper response** is user-facing and honest: Andromeda cannot prove *who* modified a local
SQLite file — it proves *that* the chain no longer matches. The response procedure: freeze
the affected database for writes except incident records, export the surviving evidence
(JSONL with re-verifiable hashes), guide the user through the [incident
procedure](#incident-response), and require an explicit acknowledgment to re-anchor: a new
chain segment starting with a sentinel record referencing the break, the export digest, and
the acknowledging actor. Silent chain restarts are prohibited.

## Retention and export

- Retention window: `security.audit_retention` (default `"400d"`). Within the window,
  records are immutable and undeletable (INV-AUD-01); audit retention takes precedence over
  operational pruning of the entities records reference (INV-AUD-04).
- After the window, records are archived to a JSONL export (chain-verifiable offline) before
  removal from the live database; `audit.retention.applied` records the pass, and archives
  land under the XDG data directory (ADR-022) unless exported elsewhere.
- Export: `andromeda export audit` (Volume 8) streams JSONL whose chain can be re-verified
  offline (Volume 2 serialization rule); exports are redacted content already — no
  additional material can appear by construction (chapter 07).

## Configuration

`[security]` keys owned by this chapter (schema/precedence per Volume 10; the table's other
keys are chapter 07's):

```toml
[security]
audit_retention = "400d"        # immutable window; archives after, never silent deletion
audit_verify_on_open = "head"   # head | full | off — off still verifies before export and in doctor
```

## Security events and incident response

### Security events

Events minted by this chapter (envelope per Volume 10, keystone FR-OBS-001):
`audit.chain.verified`, `audit.chain.broken`, `security.incident.opened`,
`security.incident.closed`. Audited actions themselves already emit their producers' events
(chapters 05–07, 09); this chapter's events cover the audit subsystem and incident
lifecycle. All are persisted event families (Volume 10 persistence classes) so the TUI can
show an incident banner after restart.

### Incident triggers

The following are **incident-response triggers** — conditions under which Andromeda actively
surfaces a security incident to the user (TUI banner, CLI warning on next invocation,
`security.incident.opened`):

| Trigger | Source |
|---|---|
| Audit chain divergence | verification (E-SEC-013) |
| Repeated sandbox violations from one subject (≥ 3 in one run) | Sandbox Engine records |
| Repeated permission evaluation failures (E-SEC-002 recurring) | Permission Manager |
| Fallback store permission widening or corruption (E-SEC-012 integrity class) | Secret Store |
| Unaccounted surviving process after teardown (E-SEC-007) | Sandbox Engine |
| Signature/verification failure during update or package install (Volumes 14/6) | Updater / Package Manager |

### Incident response procedure

User-facing, local-first — Andromeda is an installed product, not a service; there is no
phone-home and no automatic remote reporting (telemetry consent rules per Volume 10):

1. **Contain.** Stop affected runs (cancellation per Volume 4); disable implicated
   extensions (plugin/MCP `disabled` states, Volume 6); revoke implicated standing grants
   (chapter 05).
2. **Preserve.** Export audit evidence (chain-verifiable JSONL) and the relevant run
   records before any cleanup.
3. **Rotate.** For any credential plausibly exposed: rotate or revoke through Volume 5
   flows; the Credential status vocabulary records the outcome; confirm no dependent
   Authentication Session survives (INV-CRED-03).
4. **Verify.** Full audit verification on both databases; workspace integrity checks
   (ADR-029 procedures); index/memory rebuilds where poisoning is suspected (caches are
   rebuildable, Volume 7).
5. **Recover.** Restore from pre-incident backups where state is untrusted (ADR-029
   backup-restore path); re-anchor a tampered chain only with recorded acknowledgment.
6. **Record.** Close the incident (`incident.recorded`, `security.incident.closed`) with a
   disposition note.

Steps 1–3 are reachable directly from the incident surface (Volume 8 recovery wireframes);
`andromeda doctor` drives 4–5.

### Product vulnerabilities and coordinated disclosure

Suspected vulnerabilities in Andromeda itself are reported through the project's security
policy (`SECURITY.md` and private vulnerability reporting on the hosting platform — repo
structure per Volume 11; policy ownership and maintainer process per Volume 15 governance).
The product-side commitments bound here: the response-time target is NFR-SEC-003
(first response ≤ 3 business days, SM-16 c), and release gating on known vulnerabilities is
NFR-SEC-001 (SM-16 a).

## Requirements

### FR-SEC-111 — Audit Log semantics

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Audit Log (Volume 9)
- Affected components: Audit Log, Persistence Layer, Permission Manager, Secret Store, Sandbox Engine, Tool Runtime, Git Engine, Terminal Engine, Package Manager, Updater, Configuration Manager, CLI, TUI
- Dependencies: Volume 2 Audit Record entity (INV-AUD-01..05); ADR-123, ADR-125, ADR-028, ADR-029
- Related risks: Threat model chapter 04 (log leakage, tamper concealment); chapters 01–03 (evidence for every threat's Response section)

#### Description

Every action in the [audited-action catalog](#audited-action-catalog) produces exactly one
Audit Record through the single per-database Audit Log writer, hash-chained per Volume 2,
under the write discipline of this chapter: record-before-effect, fail-closed on append
failure (E-SEC-014), exempt from backpressure shedding. Verification follows the
head/full procedure with the tamper response defined here; retention applies the
`security.audit_retention` window with archive-before-removal; exports are offline
re-verifiable.

#### Motivation

PRD-006 makes auditability an identity property; SM-13 requires 100% of side effects to
resolve to their decision chain. A hash-chained, closed-catalog audit log is what turns
"trust me" into "verify it" for an agent that edits files and runs commands on a developer's
machine.

#### Actors

All producer components; the Audit Log writer; users (verification, export, incidents);
`andromeda doctor`.

#### Preconditions

Databases open; catalog producers registered.

#### Main flow

1. A producer performs a catalogued action; the writer composes the record (actor, action,
   subject, permission context, outcome, safe detail).
2. The writer assigns `prev_hash`, computes `record_hash`, appends.
3. The action's effects proceed; events are emitted.

#### Alternative flows

- Append failure: the action fails with E-SEC-014 (fail-closed); the failure itself is
  reported through error channels and retried once at the writer level.
- Verification (head on open; full on demand/schedule/export/restore) per the
  [procedure](#chain-verification-and-tamper-response).
- Retention pass: archive JSONL, verify archive digest, then prune; the pass is itself
  audited.

#### Edge cases

- Cross-database actions (e.g., a global credential used in a workspace run): each database
  records its own side (`secret.accessed` globally; `permission.decided` in the workspace),
  linked by correlation IDs — chains never span databases.
- Clock regression between records: chain order is the authority (ADR-027 sequence
  discipline); `occurred_at` may be non-monotonic and verification does not treat that as
  divergence.
- Restore from backup: full verification runs; records after the backup point are expected
  missing — the recovery report distinguishes truncation-by-restore (chain intact, shorter)
  from tampering (chain broken).
- Very large `detail` payloads: bounded (8 KiB per record); overflow is summarized with a
  reference to the operational record carrying the full content.

#### Inputs

Catalogued actions with context; verification requests; retention configuration.

#### Outputs

Chained Audit Records; verification reports; archives and exports; audit events.

#### States

Not applicable — append-only records; the log itself has no machine beyond frozen/normal
write modes during incidents.

#### Errors

E-SEC-013 (chain divergence), E-SEC-014 (append failure).

#### Constraints

One writer per database; closed catalog; record-before-effect; no deletion inside the
window; archives before pruning; safe-detail rules absolute (INV-AUD-05).

#### Security

The chain converts local tampering from silent to evident; freeze-and-acknowledge
re-anchoring preserves evidence while keeping the product usable after an incident.

#### Observability

`audit.chain.verified`/`audit.chain.broken` events; verification stats in doctor output;
SM-13 chain-resolution measurements run over these records.

#### Performance

Appends are single-row inserts with one SHA-256; verification budgets per NFR-SEC-005;
audit overhead on the tool hot path is inside Volume 12's dispatch budgets.

#### Compatibility

Identical across platforms; JSONL exports re-verify with any SHA-256 implementation
(documented canonical form, Volume 2 chapter 10).

#### Acceptance criteria

- Given any catalogued action in an instrumented E2E run, when the chain is walked, then
  exactly one record exists for it with intact links (INV-AUD-03; SM-13 method).
- Given an in-place modification of any historical record, when full verification runs,
  then divergence is reported at that position, the database freezes for normal writes, and
  an incident opens.
- Given an audit append failure injected before a tool invocation proceeds, when observed,
  then the invocation fails with E-SEC-014 and no side effect occurred.
- Negative case: pruning attempts inside the retention window are refused; operational
  pruning never removes referenced Audit Records (INV-AUD-04 test).
- Permission case: audit records for denied actions exist with outcome `denied`.
- Observability case: verification runs are themselves audited (`audit.verified`).

#### Verification method

Chain property tests (append/verify/tamper matrices); crash injection between record and
effect (record-before-effect assertion); retention and archive round-trip tests with offline
re-verification; SM-13 audit-chain suite (Volume 13).

#### Traceability

PRD-005, PRD-006; SM-13, SM-16; ADR-123, ADR-125, ADR-028, ADR-029; Volume 2
INV-AUD-01..05; FR-SEC-100, FR-SEC-101, FR-SEC-102.

### FR-SEC-112 — Incident response and disclosure hooks

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Audit Log (Volume 9)
- Affected components: Audit Log, CLI, TUI, Permission Manager, Sandbox Engine, Secret Store, Updater, Package Manager
- Dependencies: FR-SEC-111; Volume 5 rotation flows; Volume 14 verification failures; Volume 15 governance (disclosure policy)
- Related risks: Threat model chapters 01–04 (every threat's Response and Recovery subsections execute through this procedure)

#### Description

Andromeda detects the [incident triggers](#incident-triggers), opens a persisted incident
record with a `security.incident.opened` event, surfaces it on the next interactive surface
(TUI banner; CLI warning line on stderr), and provides the six-step response procedure
(contain, preserve, rotate, verify, recover, record) with direct actions: stop runs, disable
extensions, revoke grants, export evidence, run verification, and close with disposition.
Product-vulnerability reporting hooks point to the project security policy (Volume 15); the
product never reports incidents remotely on its own (telemetry consent rules, Volume 10).

#### Motivation

A threat model without a bound response procedure decays into documentation. The triggers
are exactly the conditions this volume's own mechanisms can detect locally with high signal;
the procedure turns each threat's Response/Recovery rows (chapters 01–04) into one
executable path.

#### Actors

Users (respond, acknowledge); detection sources (audit, sandbox, secret store, updater);
CLI/TUI (surfaces).

#### Preconditions

Audit subsystem operational (incident records are audit records).

#### Main flow

1. A trigger condition is met; an incident record opens (audited, evented).
2. The next interactive surface shows the incident with severity and affected subjects.
3. The user executes response actions from the incident surface; each is itself audited.
4. The incident closes with a disposition note.

#### Alternative flows

- Non-interactive environments: the incident is recorded and the CLI exits with its normal
  outcome plus a stderr warning; nothing blocks CI beyond the failing action itself.
- Chain-tamper incidents: write freeze and re-anchor acknowledgment per FR-SEC-111.

#### Edge cases

- Trigger storms (many violations at once): incidents deduplicate per subject and condition
  within one run — one incident, many linked records.
- Incident on the global database while workspaces are healthy: surfaced in every workspace
  session until closed (credentials affect all workspaces).
- User ignores the incident: the banner persists across sessions until explicitly closed;
  no auto-close.

#### Inputs

Trigger conditions; user response actions; disposition notes.

#### Outputs

Incident records; response-action audit records; evidence exports; closure events.

#### States

Incident records carry `open`/`closed` as recorded status (not a canonical machine —
recorded via paired `incident.recorded` audit actions and events).

#### Errors

Failures during response actions surface their own codes (E-SEC-013/014, Volume 5 rotation
errors); the incident record persists regardless.

#### Constraints

Local-first: no automatic remote reporting; incident surfaces MUST NOT display secret
material (redaction applies); response actions require the same permissions as their normal
paths (no privileged bypass).

#### Security

Bounds mean-time-to-containment for local compromise classes; evidence-first ordering
(preserve before recover) keeps forensics possible.

#### Observability

`security.incident.opened`/`security.incident.closed`; incident list queryable via CLI
(Volume 8); all response actions correlated to the incident ID.

#### Performance

Trigger evaluation is record-driven (no polling beyond existing writes); banner rendering
per Volume 8 budgets.

#### Compatibility

Identical across platforms and modes; headless mode records and reports on the IPC surface.

#### Acceptance criteria

- Given three sandbox violations from one plugin in one run, when the third records, then
  one incident opens naming the plugin, and the TUI shows it on next render.
- Given an incident, when the user disables the implicated extension and revokes its
  grants from the incident surface, then those actions are audited with the incident
  correlation.
- Given a tampered chain, when the incident flow completes with acknowledgment, then the
  re-anchor sentinel references the incident and the evidence export digest.
- Negative case: no incident triggers on isolated, expected denials (single E-SEC-001).
- Observability case: `security.incident.*` events and records resolve bidirectionally.

#### Verification method

Trigger simulation suite (each trigger row); dedup tests; incident-surface integration
tests (TUI/CLI); evidence-export verification; audit correlation checks (Volume 13).

#### Traceability

PRD-005, PRD-006; SM-16; FR-SEC-111; threat model Response/Recovery rows (chapters 01–04);
Volume 15 disclosure governance; Volume 8 recovery surfaces.

## Non-functional requirements

### NFR-SEC-001 — Release vulnerability posture

- Category: Security
- Priority: P0
- Phase: v1
- Metric: Count of known vulnerabilities of critical or high severity (code scanning, dependency audit, secret scanning findings) open in a release at publication time (SM-16 a)
- Target: 0 at publication
- Minimum threshold: 0 at publication — a finding either is fixed, or the release does not publish
- Measurement method: CodeQL, dependency audit, and secret-scanning results gating the release pipeline (Volume 11 pipelines; Volume 14 release gates); measured from MVP onward, gating at v1 per Volume 1 metric governance
- Test environment: CI release pipeline
- Measurement frequency: every release
- Owner: Volume 9 (posture) / Volume 11 (pipeline automation)
- Dependencies: Volume 11 CI definitions; Volume 14 release qualification
- Risks: scanner coverage gaps understate the count — mitigated by the SM governance measurement-honesty rule
- Acceptance criteria: Release audit attachment shows zero open critical/high findings at publication; a suppressed finding requires a recorded justification reviewed at the next release.

### NFR-SEC-003 — Coordinated disclosure first response

- Category: Compliance
- Priority: P1
- Phase: v1
- Metric: Elapsed business days from receipt of a security report (project security policy channel) to first substantive maintainer response (SM-16 c)
- Target: ≤ 3 business days
- Minimum threshold: ≤ 3 business days for 90% of reports in any rolling 12-month window, no report unanswered ≥ 10 business days
- Measurement method: security-inbox tracking with timestamps (process per Volume 15 governance); reviewed at phase gates
- Test environment: not applicable (process metric); tracking tooling verified in Volume 15 process audits
- Measurement frequency: continuous tracking; quarterly review
- Owner: Volume 9 (commitment) / Volume 15 (process)
- Dependencies: Volume 11 SECURITY.md and private reporting configuration; Volume 15 maintainer process
- Risks: single-maintainer availability gaps — mitigated by the Volume 15 escalation roster
- Acceptance criteria: Tracking shows the rolling-window targets met; the phase-gate review includes the disclosure-response report.

### NFR-SEC-005 — Audit chain integrity and verification performance

- Category: Security
- Priority: P0
- Phase: MVP
- Metric: (a) Chain continuity across the crash-injection suite (SM-11 method applied to audit writes); (b) full-verification wall-clock time for a 100,000-record chain
- Target: (a) 100% continuity — zero broken or missing links after any injected crash; (b) ≤ 10 s
- Minimum threshold: (a) 100%; (b) ≤ 30 s
- Measurement method: crash-injection suite verifying chains after kill −9 at randomized write points; benchmark harness over a generated 100k-record database on reference hardware
- Test environment: CI on Tier 1 platforms; reference hardware per Volume 1 chapter 06
- Measurement frequency: every release
- Owner: Audit Log (Volume 9)
- Dependencies: FR-SEC-111; ADR-123; Volume 10 storage write discipline
- Risks: verification cost pressuring users to disable checks — mitigated by the O(1) head check default
- Acceptance criteria: Crash suite reports zero divergences attributable to Andromeda writes; verification benchmark within threshold on both reference machines; `head` check adds ≤ 5 ms to database open.

## Error codes

### E-SEC-013 — Audit chain integrity violation

- Category: Integrity
- Severity: Critical
- User message: "The audit log failed integrity verification. Evidence may have been modified."
- Technical message: database (workspace/global), first divergent position, expected/actual hash prefixes, verified segment map
- Cause: in-place modification, deletion, or reordering of audit records; storage corruption is distinguished where SQLite-level checks also fail (ADR-029 path)
- Safe-to-log data: database kind, chain position, hash prefixes (8 chars)
- Recoverability: evidence is not restorable; the product recovers via the re-anchor procedure with acknowledgment
- Retry policy: none — verification is deterministic
- Recommended action: follow the incident procedure: preserve exports, investigate, acknowledge and re-anchor
- Exit-code mapping: 9
- HTTP mapping: not applicable
- Telemetry event: `audit.chain.broken`
- Security implications: this is the tamper-evidence mechanism firing; the write freeze protects the remaining evidence

### E-SEC-014 — Audit write failure

- Category: Integrity
- Severity: Critical
- User message: "A security-relevant action was blocked because its audit record could not be written."
- Technical message: action name, producer, storage error detail, retry outcome
- Cause: audit database unwritable (disk full, I/O error, lock corruption) at append time
- Safe-to-log data: action name, producer, error class
- Recoverability: recoverable once storage recovers
- Retry policy: one immediate writer-level retry; then fail the action
- Recommended action: resolve the storage condition (disk space, permissions); rerun the action
- Exit-code mapping: 9
- HTTP mapping: not applicable
- Telemetry event: `audit.chain.verified` (outcome: write_failed)
- Security implications: fail-closed by ADR-125 — no audited action executes unrecorded; repeated occurrences trigger an incident
