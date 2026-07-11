# volume-09-security — Volume Register

Merged from per-agent register fragments at the Phase B gate.

## Requirements index

Risks carry no Phase field (Volume 0 risk template); the Phase column is marked `—`. Every
`RISK-SEC-*` entry is defined in chapters 02–04 with the full risk metadata plus the ten mandated
subsections (Asset, Actor, Vector, Preconditions, Impact, Prevention, Response, Recovery, Residual
risk, Tests).

| ID | Title | Phase | Verification method |
|---|---|---|---|
| RISK-SEC-001 | Prompt injection (direct) | — | Injection-corpus tests (no unmediated side effect); permission-matrix and approval-flow tests; audit-chain tests (SM-13) |
| RISK-SEC-002 | Indirect prompt injection | — | Indirect-injection fixtures over files/issues/tool results; provenance tests; egress-decision tests |
| RISK-SEC-003 | Tool injection | — | Input-schema mutation/fuzzing; sink-control tests; no-shell-interpolation assertion |
| RISK-SEC-004 | Tool poisoning | — | Descriptor-pinning re-consent tests; grant-scoping tests; declaration-diff conformance |
| RISK-SEC-005 | MCP poisoning | — | MCP conformance and trust-gating tests (SM-15); served-injection fixtures; context-exposure tests |
| RISK-SEC-006 | Malicious model output | — | Structured-output conformance/mutation; terminal-escape fixtures; tool-argument validation |
| RISK-SEC-007 | Memory poisoning | — | Provenance-tracking tests; poisoning fixtures with delete; secret-in-memory refusal tests |
| RISK-SEC-008 | Index poisoning | — | Index-hit provenance tests; invalidate-and-rebuild tests; untrusted-labeled retrieval tests |
| RISK-SEC-009 | Malicious files | — | Decompression-bomb/oversized fixtures; in-file injection fixtures; by-reference binary handling |
| RISK-SEC-010 | Malicious plugins | — | Plugin isolation/teardown tests; permission-scoping tests; install-verification and ARP conformance |
| RISK-SEC-011 | Malicious skills | — | Skill injection fixtures; required-tools gating tests; update-re-consent tests |
| RISK-SEC-012 | Malicious repositories | — | Hostile-repo fixtures (no hook/script auto-exec); path/symlink containment; Git-mutation gating |
| RISK-SEC-013 | Dependency attacks | — | Module-sum verification; CI dependency-audit gate (SM-16); SBOM diff; reproducible-build checks |
| RISK-SEC-014 | CI compromise | — | Pipeline-permission audits; pinned-action verification; fork-PR isolation; provenance verification |
| RISK-SEC-015 | Release compromise | — | Checksum verification; signature/provenance verification where enabled; yank-and-refuse tests |
| RISK-SEC-016 | Update compromise | — | Verify-before-apply refusal tests; tampered-artifact fixtures; atomic-apply and offline-rollback tests |
| RISK-SEC-017 | Compromised providers | — | Egress least-exposure tests; provider-change notification; fallback guard-rule tests (Volume 5) |
| RISK-SEC-018 | Compromised local models | — | Local-provider conformance (SM-04); capability-honesty tests; output validation |
| RISK-SEC-019 | Command injection | — | No-shell-interpolation fixtures; permission-matrix `execute` gating; sandbox env/limit/teardown tests |
| RISK-SEC-020 | Path traversal | — | Traversal fixtures (rejection); canonicalization and per-path permission tests; working-dir confinement |
| RISK-SEC-021 | Symlink attacks | — | Out-of-root symlink fixtures; TOCTOU race tests; resolved-target confinement tests |
| RISK-SEC-022 | Secret exfiltration | — | Redaction-conformance; environment-filtering; egress-gating; secret-access audit tests |
| RISK-SEC-023 | Credential theft | — | No-plaintext-at-rest tests; fallback encryption tests; zeroize-on-release; `credential_access` gating |
| RISK-SEC-024 | Sandbox escape | — | SandboxPort-only launch audits; containment-level recording; policy enforcement; teardown process-tree tests |
| RISK-SEC-025 | Privilege escalation | — | Scope-isolation tests; confused-deputy tests; grant-revocation/audit; approval-on-widening tests |
| RISK-SEC-026 | Log leakage | — | Redaction-conformance over logs/errors/events/traces; secret-pattern scans; telemetry-content tests |
| RISK-SEC-027 | Social engineering | — | Trusted-prompt effect/scope rendering tests; default-deny; approval-audit linkage; no-dangerous-default UI tests |

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-SEC-100 | Permission model | Core | Permission-matrix unit tests; mediation probes (NFR-SEC-002); CLI/TUI parity fixtures; audit-chain resolution (SM-13 method) |
| FR-SEC-101 | Sandbox | MVP | Sandbox conformance suite: escape attempts, limit fault injection, teardown timing, orphan-sweep crash tests; containment honesty documentation audit |
| FR-SEC-102 | Secret storage | MVP | Per-platform backend integration tests; leak-hunt suite; INV-CRED-04 ordering crash tests; permission-denial probes; audit-chain resolution |
| FR-SEC-103 | Evaluation precedence and inheritance | Core | Property-based candidate-permutation tests; golden decision tables; config-validation negatives; cross-platform determinism runs |
| FR-SEC-104 | Grant persistence, expiry, and revocation | MVP | Persistence crash injection (decision→mint boundary); revocation races; ADR-029 backup/restore divergence tests; listing golden tests |
| FR-SEC-105 | Non-interactive and policy-only enforcement | MVP | Mode-parity fixture suite; bypass-hunt over documented env vars/flags; prompt-free CI matrix (Volume 8); audit inspection |
| FR-SEC-106 | Sandbox tiers | MVP | Tier matrix integration tests (five subjects); workflow ceiling tests; plugin/MCP environment probes; record audits |
| FR-SEC-107 | Environment and secret filtering | MVP | Environment probes across tiers/platforms; planted-secret scrubs; single-construction-path static check; resolution fault injection |
| FR-SEC-108 | Filesystem policy, symlinks, temp directories, and cleanup | MVP | Path-policy fuzzing (traversal, links, case, depth) per platform; crash-injection sweep tests; teardown verification |
| FR-SEC-109 | Secret redaction at every sink | MVP | Leak-hunt suite with planted markers across all sinks; chunk-boundary property tests; schema field tests; redaction fault injection |
| FR-SEC-110 | Fallback store consent and lifecycle | MVP | Consent-flow integration tests; permission/corruption fixtures; keychain↔fallback migration round-trips; wording audit |
| FR-SEC-111 | Audit Log semantics | MVP | Chain property tests (append/verify/tamper); record-before-effect crash injection; retention/archive round-trips with offline re-verification; SM-13 suite |
| FR-SEC-112 | Incident response and disclosure hooks | MVP | Trigger simulation per trigger row; dedup tests; incident-surface integration tests; evidence-export verification |
| FR-SEC-113 | Approval lifecycle enforcement | MVP | Machine property tests; decision/expiry race injection; crash injection at persistence boundaries (SM-11 method); prompt fidelity goldens |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-SEC-001 | Release vulnerability posture (SM-16 a) | v1 | CodeQL + dependency audit + secret scanning gating releases; release audit attachment |
| NFR-SEC-002 | Permission mediation coverage (SM-16 b) | MVP | Enforcement suite attempting unmediated side effects on every path; audit-chain resolution over instrumented runs |
| NFR-SEC-003 | Coordinated disclosure first response (SM-16 c) | v1 | Security-inbox tracking (Volume 15 process); quarterly and phase-gate review |
| NFR-SEC-004 | Secret leakage prevention across boundaries | MVP | Leak-hunt suite scanning all sinks for planted markers, all Tier 1 platforms, per merge and release |
| NFR-SEC-005 | Audit chain integrity and verification performance | MVP | Crash-injection chain continuity suite; 100k-record verification benchmark on reference hardware |
| NFR-SEC-006 | Permission decision latency | MVP | Micro-benchmark at reference grant population (1,000 grants / 200 rules) per release |

### Risks

Fragment B mints no RISK identifiers; the volume's RISK-SEC-* threat catalog is fragment A's
(chapters 01–04).

## ADRs minted

None. Block 115–120 is reserved for Volume 9. Fragment A introduces no new architecture decision:
the threat model mitigates against controls already decided by accepted ADRs (notably ADR-009,
ADR-013, ADR-014, ADR-016, ADR-019, ADR-021, ADR-025, ADR-029, ADR-032, ADR-033) and against the
Volume 9 keystones FR-SEC-100/101/102 formalized by fragment B. Any ADRs required for the control
chapters are minted by fragment B.

Block 115–129 belongs to Volume 9; fragment A mints from 115 (115–120 available to A);
fragment B used 121–125 (126–129 remain unused by B).

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-121](../annexes/adr/ADR-121.md) | Permission evaluation: deny-overrides with an ask-forcing tier | Accepted | Four-tier resolution deny > ask > allow > default-ask; no specificity or recency ordering; layer precedence stays Volume 10's |
| [ADR-122](../annexes/adr/ADR-122.md) | Sandbox tiers: five subject profiles over one policy interface | Accepted | `process`/`tool`/`workflow`/`plugin`/`mcp_server` base profiles; effective policy = tier ∩ declaration ∩ grants; closed set |
| [ADR-123](../annexes/adr/ADR-123.md) | Audit chain verification: O(1) head check by default, full walk on demand | Accepted | Head check on open; full verify in doctor/schedule/export/restore; freeze-preserve-acknowledge tamper response |
| [ADR-124](../annexes/adr/ADR-124.md) | Redaction strategy: registry exact-match plus structural and heuristic layers | Accepted | In-process resolved-secret registry (≥ 8 chars, encodings) + schema-field redaction + additive pattern heuristics at every sink; fail-closed |
| [ADR-125](../annexes/adr/ADR-125.md) | Fail-closed security decision path | Accepted | Evaluation failure denies (E-SEC-002); unwritable audit blocks the action (E-SEC-014); unappliable sandbox policy refuses launch; redaction failure withholds payloads |

## Error codes minted

None. Security error codes (`E-SEC-*`) are minted by fragment B (permission, sandbox, and secret
control chapters). Threat entries reference the E-SEC family and existing E-TOOL codes by name.
| Code | Name | Exit code |
|---|---|---|
| E-SEC-001 | Action denied by permission model | 5 |
| E-SEC-002 | Permission evaluation failure | 5 (9 on store integrity failure) |
| E-SEC-003 | Approval expired | 5 |
| E-SEC-004 | Approval cancelled | 8 (when it terminates a command) |
| E-SEC-005 | Sandbox policy violation | 5 (pre-launch denial) / 6 (running tool terminated) |
| E-SEC-006 | Isolation mechanism unavailable | 6 |
| E-SEC-007 | Sandbox teardown failure | 6 |
| E-SEC-008 | Secret backend unavailable | 4 |
| E-SEC-009 | Secret reference not found | 4 |
| E-SEC-010 | Secret access denied by OS store | 4 |
| E-SEC-011 | Fallback passphrase unavailable | 4 |
| E-SEC-012 | Fallback store unreadable | 4 (9 on confirmed corruption) |
| E-SEC-013 | Audit chain integrity violation | 9 |
| E-SEC-014 | Audit write failure | 9 |

## Events minted

None. Security events (permission decisions, sandbox refusals, secret access, audit) are minted by
fragment B (chapter 08) under the Volume 10 envelope. Threat entries reference existing events
(`tool.invocation.denied`, `terminal.execution.*`, `permission.decision.recorded`) by name.

Envelope, delivery, persistence, and retention semantics per Volume 10 (keystone
FR-OBS-001); grammar per Volume 0 chapter 03.

| Event | Emitted by |
|---|---|
| `permission.decision.recorded` | Permission Manager |
| `permission.grant.created` | Permission Manager |
| `permission.grant.revoked` | Permission Manager |
| `permission.grant.expired` | Permission Manager |
| `approval.requested` | Permission Manager |
| `approval.granted` | Permission Manager |
| `approval.denied` | Permission Manager |
| `approval.expired` | Permission Manager |
| `approval.cancelled` | Permission Manager |
| `sandbox.prepared` | Sandbox Engine |
| `sandbox.policy.applied` | Sandbox Engine |
| `sandbox.violation.blocked` | Sandbox Engine |
| `sandbox.containment.degraded` | Sandbox Engine |
| `sandbox.teardown.completed` | Sandbox Engine |
| `secret.stored` | Secret Store |
| `secret.accessed` | Secret Store |
| `secret.deleted` | Secret Store |
| `secret.fallback.enabled` | Secret Store |
| `secret.orphan.swept` | Secret Store |
| `audit.chain.verified` | Audit Log |
| `audit.chain.broken` | Audit Log |
| `security.incident.opened` | Audit Log |
| `security.incident.closed` | Audit Log |

The audited-action catalog (chapter 08) additionally names the audit `action` identifiers;
those are Audit Record attributes, not event names, and are listed in chapter 08.

## Config keys minted

None. The `[permissions]`, `[sandbox]`, and `[security]` table content is minted by fragment B;
schema and precedence are Volume 10's.

Key content owned by Volume 9; schema, precedence, and validation are Volume 10's
(keystone FR-CFG-001).

`[permissions]`: `permissions.approval_timeout` ("10m"), `permissions.workspace_grant_ttl`
("0s" = no auto-expiry), `permissions.rules` ([] — array of tables: `name`, `permission`,
`effect` allow|deny|ask, plus resource-qualifier keys per the chapter 05 selector grammar).

`[sandbox]`: `sandbox.isolation` ("auto"), `sandbox.degradation` ("ask"),
`sandbox.env_allowlist` (["PATH","HOME","USER","LOGNAME","LANG","LC_ALL","LC_CTYPE","TERM",
"TMPDIR","SHELL","TZ"]), `sandbox.writable_roots` ([]), `sandbox.readonly_roots` ([]),
`sandbox.command_denylist` (["sudo *","su *","doas *","shutdown *","reboot *","mkfs*",
"dd *"]), `sandbox.command_allowlist` ([]), `sandbox.max_cpu_seconds` (300),
`sandbox.max_memory_mb` (2048), `sandbox.max_processes` (64), `sandbox.max_open_files`
(512).

`[security]`: `security.fallback_store` (false), `security.redaction_patterns` ([]),
`security.audit_retention` ("400d"), `security.audit_verify_on_open` ("head").

## Glossary additions

Merged into the corpus glossary at consolidation.

| Term | One-line meaning |
|---|---|
| Trust boundary | A separation between domains of differing trust where a security control MUST sit. |
| Threat actor | A party (human, process, or the confused-deputy model) capable of realizing a threat. |
| Attack vector | The channel through which a threat reaches an asset (model context, tool surface, command/path, credential handling, supply chain, inference channel, human channel). |
| Untrusted-content labeling | Marking ingested content (files, tool results, memory, index hits, provider output) as data, never as privileged instruction. |
| Confused deputy | A component (including the model) that misuses its own granted authority when misled by untrusted input. |
| Least exposure | Sending or exposing the minimum context and surface necessary to a lower-trust domain. |
| Effective containment level | The observable, per-execution record of which sandbox layer actually applied (ADR-021). |
| Residual risk | The exposure that remains after a threat's named controls are applied. |
| TOCTOU | Time-of-check-to-time-of-use: a race where a checked resource changes before it is used (relevant to symlink handling). |
| Term | One-line meaning |
|---|---|
| Scope qualifier | One of the ten frozen dimensions (`session`, `workspace`, `command`, `tool`, `provider`, `host`, `path`, `domain`, `repository`, `organization`) bounding what a grant covers. |
| Grant scope | The lifetime attachment of a Permission row: `invocation`, `run`, `session`, `workspace`, or `global` (Volume 2 enum). |
| Ask-forcing rule | An `effect = "ask"` policy rule that pins interactive review even where a standing allow matches (ADR-121 tier 2). |
| Redaction registry | The in-process set of resolved secret material (and encodings) that every sink scrubs before emission (ADR-124). |
| Sandbox tier | One of five base isolation profiles keyed by subject kind: `process`, `tool`, `workflow`, `plugin`, `mcp_server` (ADR-122). |
| Effective containment level | The isolation layer actually achieved for one execution (`process` or `os:<mechanism>`), recorded per ADR-021. |
| Audited action | An entry of the closed chapter 08 catalog that MUST yield exactly one Audit Record, including denied/failed outcomes. |
| Re-anchor | The acknowledged restart of a tampered audit chain: a new sentinel segment referencing the break and the evidence-export digest. |
| Fallback store | The age-encrypted credential file backend, active only under explicit consent (ADR-014; FR-SEC-110). |
| Incident record | The paired `incident.recorded` audit actions (with `security.incident.*` events) marking an open or closed security incident. |

## Assumptions

Local list per Volume 0 chapter 05 (global numbering assigned at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | Threat probability ratings in the risk matrix are prior design estimates, to be recalibrated from the security telemetry of consenting installations (SM-16) | Field metrics from consenting installations; phase-gate security review | Reclassify affected threats' Probability/Severity through the register; controls are unaffected |
| Technical assumption | Process-level sandbox controls (env filtering, path policy, resource limits) eliminate the common failure classes at MVP without OS-level isolation (ADR-021) | Sandbox conformance tests per platform (RISK-SEC-024 tests) | OS-level isolation is brought forward or the MVP containment claim is narrowed further in documentation |
| Provided constraint | Andromeda cannot defend against an attacker already holding the user's OS account; that defense is the operating system's | Stated boundary; no test claims protection there | No change — the exclusion is explicit so no threat overclaims |
| Product hypothesis | Effect-naming, default-deny approval prompts meaningfully reduce social-engineering success versus generic confirmations | Beta/v1 usage telemetry; approval-audit review | Revisit prompt design (Volume 8) through the change procedure |

Local list per Volume 0 chapter 05 (global numbers at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | Official authentication flows do not issue secret material shorter than 8 characters, so the registry redaction floor loses nothing real | Review per provider adapter at its phase start (Volume 5 catalog) | Register short material with an exact-length entry class; review the ADR-124 floor |
| Technical assumption | SHA-256 append cost and the O(1) head check are negligible on the hot path and open path respectively | NFR-SEC-005 benchmarks; Volume 12 startup budgets | Batch-tolerant chain writer within the record-before-effect rule; revisit ADR-123 defaults |
| Technical assumption | The process-level floor's resolve-then-act path policy, with per-sandbox temp dirs, eliminates the practical symlink/traversal classes short of a concurrently mutating attacker | Path-policy fuzzing (FR-SEC-108); threat model chapter 04 residual-risk review | Accelerate the OS-level layer for affected platforms (ADR-021 review) |
| Product hypothesis | A 10-minute default `approval_timeout` matches interactive rhythm without routine expiries | Beta usage feedback; expiry-rate metrics | Adjust the default — configuration change, not contract change |
| Product hypothesis | Five sandbox tiers cover all execution subjects through v1 without a sixth | Volume 6/4 integration during Beta | Add a tier via ADR (ADR-122 closed-set rule) |

## Open questions

Every `PENDING VALIDATION` used in chapters 01–04 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V9A-OQ-1 | OS-level sandbox mechanism availability and behavior per platform (macOS `sandbox-exec`; Linux Landlock/namespaces/bubblewrap) — PENDING VALIDATION (ADR-021) | Chapters 01, 04 (RISK-SEC-019/021/024) | No — MVP ships process-level controls; OS-level isolation is Beta/v1 | Per-platform mechanism validation during Beta/v1 implementation; recorded via ADR-021 review conditions | Open |
| V9A-OQ-2 | Cryptographic signature and macOS notarization availability for this project's releases — PENDING VALIDATION (Volume 1 signing viability note; ADR-013) | Chapter 03 (RISK-SEC-015/016) | No — checksums and provenance are the MVP floor; signing is a configuration change | Confirm signing identities and notarization credentials; enable in the release pipeline | Open |
| V9A-OQ-3 | Coverage of relevant advisories by the chosen dependency/vulnerability scanners — PENDING VALIDATION | Chapter 03 (RISK-SEC-013) | No — pinning, minimal deps, and SBOM apply regardless | Validate scanner coverage against known advisory sets at CI setup; supplement scanners if gaps found | Open |
| V9A-OQ-4 | Per-provider retention of submitted content — PENDING VALIDATION per provider (governed by each provider's terms) | Chapter 03 (RISK-SEC-017) | No — least exposure and local-first operation apply regardless | Record each provider's documented retention stance in Volume 5 at that adapter's phase | Open |

Every PENDING VALIDATION in chapters 00, 05–09 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V9B-OQ-1 | macOS OS-level isolation via Seatbelt (`sandbox-exec`): availability, profile language stability, and behavior per supported macOS version — PENDING VALIDATION per ADR-021 | Chapter 06 (layered enforcement; `sandbox.isolation`) | No — MVP ships the process-level floor | Implementation-time mechanism tests per macOS version before any isolation claim (ADR-021 review condition) | Open |
| V9B-OQ-2 | Linux OS-level isolation via Landlock (kernel-version dependent), namespaces (distribution/privilege dependent), and bubblewrap (external binary) — PENDING VALIDATION per ADR-021, each mechanism separately | Chapter 06 (layered enforcement; `sandbox.isolation`) | No — MVP ships the process-level floor | Per-mechanism validation on reference distributions before any isolation claim; record outcomes via ADR-021 amendment | Open |

## Cross-volume references

- **Volume 0:** risk template (chapter 02); identifier taxonomy and single-home matrix placing the
  permission and sandbox models in Volume 9 (chapter 03); precedence order (safety first).
- **Volume 1:** phase definitions; product principles Safe by Default and Local-first; the signing
  viability note (RISK-SEC-015/016); success metrics SM-05, SM-13, SM-16.
- **Volume 2:** frozen states referenced by threat responses — Approval (`denied`/`granted`/…),
  Tool Invocation (`denied`), Provider connection (`degraded`/`unavailable`), Memory Record and
  Index statuses, Update/Release states, Credential `status`.
- **Volume 3:** frozen ports the controls sit on — `PermissionPort`, `SandboxPort`,
  `SecretStorePort`, `ToolPort`, `TerminalPort`, `GitPort`, `IndexerPort`, `MemoryStorePort`,
  `ProviderPort`, `AuthPort`, `PackagePort`, `UpdaterPort`; the SandboxPort-only launch rule; the
  dependency matrix (ADR-030) and CI enforcement (ADR-033).
- **Volume 4:** Agent Engine loop and run cancellation (interrupt → sandbox teardown); Context
  Manager exclusion controls referenced by injection responses.
- **Volume 5:** provider auth flows, fallback guard rules (RISK-SEC-017), capability-honesty
  conformance (SM-04, RISK-SEC-018); per-provider retention (V9A-OQ-4).
- **Volume 6:** tool/MCP/skill/plugin trust vocabulary, descriptor pinning, sandbox tiers,
  conformance (SM-15) referenced by RISK-SEC-003/004/005/010/011; ARP isolation (ADR-009).
- **Volume 7:** memory provenance/trust and index rebuildability (INV-IDX-02) for RISK-SEC-007/008;
  least-exposure context assembly.
- **Volume 9 fragment B:** keystones FR-SEC-100 (permission model), FR-SEC-101 (sandbox), FR-SEC-102
  (secret management); E-SEC-* codes; security events and Audit Log (chapter 08); Approval machine
  (chapter 09) — the normative homes of every control this fragment references.
- **Volume 10:** event envelope and redaction (RISK-SEC-026); logging (ADR-011); telemetry consent;
  configuration schema for `[permissions]`/`[sandbox]`/`[security]`.
- **Volume 11:** Git engine and destructive-operation gating (RISK-SEC-012); CI/GitHub Actions,
  branch protection, fork-PR isolation, provenance (RISK-SEC-014); commit/release integrity.
- **Volume 14:** release, update, and rollback machines; checksums/signatures/provenance and yank
  (RISK-SEC-015/016).

Volume 2: Permission, Approval, Credential, Audit Record entity shapes and invariants
(INV-PERM, INV-APR, INV-CRED, INV-AUD); frozen Approval states; write discipline and
canonical serialization (chapters 04/05/08/09/10). Volume 3: frozen PermissionPort,
SandboxPort, SecretStorePort signatures elaborated here; SandboxPort-only launch rule;
recovery procedure consumed by Approval recovery and orphan sweeps. Volume 1: SM-16
formalized as NFR-SEC-001/002/003; phase definitions; precedence order. Volume 4: run/task
`awaiting_approval` blocking, cancellation cascades, gated-retry re-evaluation; workflow
gates (FR-WF-001). Volume 5: authentication flows over FR-SEC-102 (keystone FR-AUTH-001);
credential rotation in incident response. Volume 6: tool/plugin/MCP permission declarations
and sandbox binding (FR-TOOL-001, FR-PLUG-001, FR-MCP-001); Tool Invocation
`awaiting_approval`/`denied` states; teardown budgets. Volume 7: memory ingestion redaction
constraint (mechanism minted here). Volume 8: permission prompts, non-interactive denial
(FR-CLI-001 surfaces), incident and recovery surfaces, `andromeda doctor`. Volume 10:
`[permissions]`/`[sandbox]`/`[security]` schema and precedence (FR-CFG-001); event envelope
(FR-OBS-001); logging redaction binding; storage write discipline. Volume 11: Git mutation
gates (FR-GIT-001); CI scanning for NFR-SEC-001; SECURITY.md placement. Volume 12:
latency/startup budgets consumed by NFR-SEC-005/006. Volume 13: security, sandbox,
permission, and leak-hunt suites. Volume 14: update verification failures as incident
triggers; release gating. Volume 15: disclosure governance and maintainer process behind
NFR-SEC-003. Within Volume 9 (fragment A): threat catalog chapters 01–04 whose Prevention/
Response rows bind to FR-SEC-100..113.
