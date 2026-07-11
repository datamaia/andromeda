# volume-10-config-storage-observability — Volume Register

Merged from per-agent register fragments at the Phase B gate.

## Requirements index

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-CFG-001 | Configuration precedence | MVP | Layer-matrix unit tests (every layer pair, both orders); determinism property tests; offline-suite zero-network assertion |
| FR-CFG-002 | Configuration documents, locations, and loading | MVP | Per-platform golden path tests (ADR-022); fault injection (unreadable/oversize/malformed); stray-file non-consultation test |
| FR-CFG-003 | Configuration Profiles as scope-bound layers | MVP | Selection-matrix unit tests (selector × scope × existence); persisted-profile integration tests; golden attribution fixtures |
| FR-CFG-004 | Environment variable mapping algorithm | MVP | Round-trip property tests over all schema keys; name fuzzing; ambiguity/duplicate matrices; Volume 8 environment-table integration |
| FR-CFG-005 | Invocation and runtime overrides | MVP | Protected-table matrix tests; TUI/IPC integration; watch-delta tests under concurrent edits; audit-chain event assertions |
| FR-CFG-006 | Include mechanism | Beta | Merge-order/cycle/depth/count/containment matrices incl. symlink escapes; include-graph fuzzing; attribution fixtures |
| FR-CFG-007 | Typed validation with complete findings | MVP | Seeded-defect corpus (100% recall per NFR-CFG-003); golden diagnostics; cross-key rule matrices; ADR-008 parser-differential tests |
| FR-CFG-008 | Configuration schema versioning, migration, and deprecation | MVP | Transform-chain tests (per-version and 1→current); rewrite crash/I-O fault injection; deprecation matrix (old/new/both) |
| FR-CFG-009 | Database backups, retention, and workspace locking | MVP | Crash injection at every step boundary; concurrent-process lock tests; retention audit-preservation property tests; disk-full injection |
| FR-CFG-010 | Configuration error reporting | MVP | Golden fixtures for every catalog entry (human + JSON); human/machine parity property test; redaction assertions |
| FR-CFG-011 | Secret detection and redaction in configuration surfaces | MVP | Seeded-secret corpus across all sinks (NFR-CFG-004 method); detector true/false-positive suites; write-path refusal tests |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-CFG-001 | Configuration resolution latency | MVP | Isolated benchmark harness, 50 iterations, p95, both reference machines; SM-06 budget decomposition |
| NFR-CFG-002 | Resolution determinism and snapshot completeness | MVP | Double-resolution hash property tests; SM-12 record-completeness validator over all suite runs |
| NFR-CFG-003 | Validation finding completeness | MVP | Seeded-defect corpus (≥ 50 configurations) with exact-match report assertions per merge |
| NFR-CFG-004 | Redaction effectiveness | MVP | Seeded-secret corpus; automated fragment scan of every sink output per merge and release |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-CFG-001 | Drift between schema, defaults, and published reference | Medium | Open |
| RISK-CFG-002 | Users mispredict the effective value across ten layers | Medium | Open |
| RISK-CFG-003 | Secrets committed to version control through workspace configuration | High | Open |
| RISK-CFG-004 | Version-skew refusals in synced or downgraded environments | Medium | Open |

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-OBS-001 | Event envelope | Core | Registry conformance over all suite-produced events; name-grammar lint in CI; SM-20 contract-diff of envelope and registry; negative-emission unit tests |
| FR-OBS-002 | Structured logging pipeline | MVP | Log-conformance schema validation over suite output; correlation-join tests (SM-13 method); sink fault injection; Tier 1 matrix |
| FR-OBS-003 | Log rotation and retention | MVP | Rotation unit tests; retention property tests (age × count × liveness); crash injection between rotation steps; disk-full fault injection |
| FR-OBS-004 | Log redaction enforcement | MVP | Secret-leak canary suite (release gate); matcher property tests; handler fault injection; stderr capture audits |
| FR-OBS-005 | Event delivery and subscription semantics | Core | Burst/soak delivery suites with policy assertions; ordering property tests; replay boundary tests; shutdown drain tests; IPC bridge conformance |
| FR-OBS-006 | Event persistence, retention, and export | MVP | Crash injection around transactional appends (SM-11 method); retention property tests with audit-linked exclusions; export round-trips against payload schemas |
| FR-OBS-007 | Trace model and OpenTelemetry mapping | MVP | Trace-completeness validator (SM-12 component); tree-invariant property tests; crash injection with post-recovery audits; cross-boundary propagation tests; redaction leak tests |
| FR-OBS-008 | Metric registry and core catalog | MVP | Registry conformance over emitted samples; catalog presence test per release; cardinality scanners over snapshots; persistence fault injection |
| FR-OBS-009 | Cost observability: rollups, honesty, retention | MVP | Rollup property tests (idempotence, late arrivals, corrections, multi-currency); retention guards; split-basis honesty tests; divergence injection |
| FR-OBS-010 | Strict local/remote separation | MVP | SM-05 offline suite with network observation; no-exporter composition test; endpoint fault injection; policy-lock matrix |
| FR-OBS-011 | Telemetry consent lifecycle | MVP | Consent lifecycle integration tests (grant/revoke/version bump/endpoint change/lock/non-interactive refusal); backup-restore staleness tests; audit assertions |
| FR-OBS-012 | Collected-data catalog and prohibited data | MVP | Payload-enumeration conformance with prohibited-pattern scanners and canaries; catalog↔allowlist CI divergence check; extension-aggregation tests |
| FR-OBS-013 | Telemetry queue, export, and deletion | MVP | OTLP-double integration matrices (success/partial/outage/auth); queue-bound property tests; backoff timing with mock clocks; shutdown flush tests |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-OBS-001 | Logging hot-path overhead | Beta | Handler-chain benchmark and suppressed-call microbenchmark per release on reference hardware |
| NFR-OBS-002 | Event publication overhead and delivery latency | Beta | Instrumented bus benchmark (subscriber counts 1/4/16, 500 events/s) per release; SM-07 regression cross-check |
| NFR-OBS-003 | Side-effect traceability | MVP | SM-13 automated audit-chain test per release (gating from MVP exit; 0 orphans) |
| NFR-OBS-004 | Run record completeness for replay | v1 | SM-12 record-completeness validator over all suite runs; replay divergence test per release |
| NFR-OBS-005 | Instrumentation overhead | Beta | Span/sample microbenchmarks plus paired instrumented/no-op E2E runs per release |
| NFR-OBS-006 | Zero-egress default posture | MVP | OS-level network observation across suites under default configuration; no-exporter composition test (gating from MVP exit) |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-OBS-001 | Observability data volume degrades the host | Medium | Open |
| RISK-OBS-002 | Correlation discontinuity across process boundaries | High | Open |
| RISK-OBS-003 | Accidental telemetry egress | High | Open |

## ADRs minted

Block 130–144 belongs to Volume 10; fragment A used 130–134 (135–136 remain unused by A;
fragment B mints from 137).

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-130](../annexes/adr/ADR-130.md) | Configuration precedence: invocation over environment over runtime overrides over scope layers, with scope-bound profiles | Accepted | Ten-layer order, key-granular merge, arrays replace wholesale; profiles bind above their own scope's file |
| [ADR-131](../annexes/adr/ADR-131.md) | Environment variable mapping: schema-matched compact names with double-underscore disambiguation | Accepted | Compact form resolves only when unique; `__` is the explicit form; ambiguity/unknown are loud errors; env overrides, never creates, dynamic entries |
| [ADR-132](../annexes/adr/ADR-132.md) | Configuration includes: explicit, bounded, containment-checked, local-only | Accepted | Ordered relative includes inside a layer; depth ≤ 8, ≤ 64 files; containment roots with symlink resolution; no URLs |
| [ADR-133](../annexes/adr/ADR-133.md) | Configuration schema versioning: integer `config_version`, forward-only transforms, in-memory upgrade with explicit backed-up rewrite | Accepted | Old files upgrade in-memory each load; rewrites only via explicit surface with verified backup; future versions refused |
| [ADR-134](../annexes/adr/ADR-134.md) | Configuration secret hygiene: fail-closed definite detectors plus universal sink redaction | Accepted | Definite detections fail the configuration (E-CFG-012, rotate remedy); heuristics warn and redact; every sink renders through redaction |
| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-137](../annexes/adr/ADR-137.md) | Event envelope: persist-then-publish, closed compiled registry, per-family overflow policy | Accepted | Lifecycle/action events append transactionally with their writes, bus delivery is a lossy live view, names live in a closed code registry, overflow policy is fixed per family |
| [ADR-138](../annexes/adr/ADR-138.md) | Logging sinks: per-process JSONL files with built-in size/age rotation | Accepted | One JSONL file per process instance under the ADR-022 log dir; standard-library rotation (32 MiB/10 files/30 days defaults); stderr degradation |
| [ADR-139](../annexes/adr/ADR-139.md) | Local traces unsampled with bounded retention; metric snapshots in SQLite | Accepted | Every run records a complete trace; metrics persist as 60 s snapshots; retention keys bound both; sampling held in reserve |
| [ADR-140](../annexes/adr/ADR-140.md) | Remote telemetry opt-in, disabled by default, interactive versioned revocable consent | Accepted | Consent = config flag + interactive confirmation of the versioned catalog and endpoint, recorded in the global DB; catalog broadening invalidates consent; enterprise lock; constructive gating |
| [ADR-141](../annexes/adr/ADR-141.md) | Telemetry transport: OTLP only, allowlist-built payloads, no log export | Accepted | One export path (OTLP http/protobuf or grpc); metrics+traces only; payloads built field-by-field from the catalog; entity IDs stripped from exported traces |

ADR numbers 142–144 of this fragment's block are unused permanent gaps.

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-CFG-001 | Configuration file unreadable | 3 |
| E-CFG-002 | Configuration file is not valid TOML | 3 |
| E-CFG-003 | Unknown configuration key | 3 |
| E-CFG-004 | Invalid type for configuration key | 3 |
| E-CFG-005 | Invalid value for configuration key | 3 |
| E-CFG-006 | Configuration keys conflict | 3 |
| E-CFG-007 | Include cannot be loaded | 3 |
| E-CFG-008 | Configuration profile cannot be resolved | 3 |
| E-CFG-009 | Environment variable mapping failure | 3 |
| E-CFG-010 | Configuration schema version unsupported | 3 |
| E-CFG-011 | Configuration migration failed | 3 |
| E-CFG-012 | Secret material detected in configuration | 3 |
| E-CFG-013 | Deprecated configuration key in use | 3 (only when warnings are promoted on a validation surface; diagnostic otherwise) |
| E-CFG-014 | Runtime override refused | 3 |
| E-CFG-015 | Database schema is newer than this build | 9 |
| E-CFG-016 | Database migration failed | 9 |
| E-CFG-017 | Database integrity check failed | 9 |
| E-CFG-018 | Database backup could not be created | 9 |
| E-CFG-019 | Workspace database is locked | 1 |
| E-OBS-001 | Malformed event envelope | 1 (qualified; normally non-fatal) |
| E-OBS-002 | Unregistered event name | 1 (qualified; normally non-fatal) |
| E-OBS-003 | Event bus closed | none (shutdown race) |
| E-OBS-004 | Subscriber buffer overflow | none |
| E-OBS-005 | Observability persistence failure | 1 (9 on integrity cause) |
| E-OBS-006 | Unregistered metric emission | 1 (development builds only) |
| E-OBS-007 | Log sink failure | none (1 for direct log-operation commands) |
| E-OBS-008 | Telemetry export failure | none (1 for direct export operations) |
| E-OBS-009 | Telemetry consent violation | 3 |

## Events minted

Envelope semantics per chapter 04 (fragment B, keystone FR-OBS-001); grammar per Volume 0
chapter 03.

| Event | Emitted by |
|---|---|
| `config.file.loaded` | Configuration Manager |
| `config.resolution.completed` | Configuration Manager |
| `config.profile.selected` | Configuration Manager |
| `config.change.detected` | Configuration Manager |
| `config.override.applied` | Configuration Manager |
| `config.override.rejected` | Configuration Manager |
| `config.migration.applied` | Configuration Manager |
| `config.migration.failed` | Configuration Manager |
| `config.deprecation.detected` | Configuration Manager |
| `config.validation.completed` | Configuration Manager |
| `config.validation.failed` | Configuration Manager |
| `config.secret.detected` | Configuration Manager |
| `storage.backup.created` | Persistence Layer |
| `storage.backup.pruned` | Persistence Layer |
| `storage.migration.applied` | Persistence Layer |
| `storage.migration.failed` | Persistence Layer |
| `storage.integrity.failed` | Persistence Layer |
| `storage.lock.denied` | Persistence Layer |
| `storage.retention.applied` | Persistence Layer |

All per the Volume 0 grammar; envelope, delivery, persistence, and retention per this
volume's chapter 04. Producers: Logging, Event Bus, Telemetry, Observability.

| Event | Defined in | Meaning |
|---|---|---|
| `log.file.rotated` | 03 | Size rotation opened a successor log file |
| `log.sink.degraded` | 03 | File sink failed; stderr degradation active (E-OBS-007) |
| `event.publish.rejected` | 04 | Registry validation rejected an emission (E-OBS-001/E-OBS-002) |
| `event.subscriber.overflowed` | 04 | A subscriber's buffer applied its overflow policy (rate-limited signal) |
| `event.retention.pruned` | 04 | Retention pass removed event rows (counts per class) |
| `observability.sink.failed` | 04 | An observability store write failed (E-OBS-005) |
| `trace.completed` | 05 | A run's trace closed (span count, status) |
| `trace.retention.pruned` | 05 | Whole traces removed by retention |
| `metric.registration.violated` | 05 | Unregistered metric emission detected (E-OBS-006) |
| `cost.rollup.compacted` | 05 | Daily cost compaction produced/replaced rollup rows |
| `telemetry.consent.granted` | 06 | Interactive consent recorded |
| `telemetry.consent.revoked` | 06 | Consent tombstoned; export stopped; queue purged |
| `telemetry.consent.violated` | 06 | Enablement/ship attempt failed the consent gate (E-OBS-009) |
| `telemetry.export.enabled` | 06 | Export pipeline constructed (gate satisfied) |
| `telemetry.export.disabled` | 06 | Export pipeline torn down |
| `telemetry.batch.exported` | 06 | An OTLP batch was accepted by the endpoint |
| `telemetry.export.failed` | 06 | An export outage episode began (E-OBS-008) |
| `telemetry.data.deleted` | 06 | Local telemetry data purged (queue and/or identifier) |

## Config keys minted

Fragment A mints the reserved root keys and the `[storage]` table; `[logging]` and
`[telemetry]` key content is fragment B's (chapters 03 and 06).

| Key | Type | Default |
|---|---|---|
| `config_version` | integer | current schema version |
| `include` | array of paths | `[]` |
| `default_profile` | string | `""` |
| `storage.lock_wait_ms` | integer | `5000` |
| `storage.backups.dir` | path | `""` (ADR-022 defaults per database) |
| `storage.backups.retain_count` | integer | `3` |
| `storage.retention.sessions_days` | integer | `0` (keep forever) |
| `storage.retention.runs_days` | integer | `0` |
| `storage.retention.artifacts_days` | integer | `0` |

Key content of the `[logging]` and `[telemetry]` tables and the observability sub-tables of
`[storage]`; schema, precedence, env-var mapping, and validation are fragment A's
(configuration chapters).

| Key | Type | Default | Chapter |
|---|---|---|---|
| `logging.level` | enum (`debug` \| `info` \| `warn` \| `error`) | `"info"` | 03 |
| `logging.stderr_level` | enum (same set) | `"warn"` | 03 |
| `logging.include_source` | bool | `false` | 03 |
| `logging.file.enabled` | bool | `true` | 03 |
| `logging.file.max_size_mb` | integer | `32` | 03 |
| `logging.file.max_files` | integer | `10` | 03 |
| `logging.file.max_age_days` | integer | `30` | 03 |
| `storage.events.retention_days` | integer (≥ 1) | `90` | 04 |
| `storage.events.max_size_mb` | integer | `512` | 04 |
| `storage.traces.retention_days` | integer (≥ 1) | `30` | 05 |
| `storage.traces.max_size_mb` | integer | `512` | 05 |
| `storage.metrics.local_persistence` | bool | `true` | 05 |
| `storage.metrics.retention_days` | integer (≥ 1) | `30` | 05 |
| `storage.cost_records.retention_days` | integer (≥ 1) | `365` | 05 |
| `telemetry.enabled` | bool | `false` | 06 |
| `telemetry.endpoint` | string (URL) | `""` | 06 |
| `telemetry.protocol` | enum (`http/protobuf` \| `grpc`) | `"http/protobuf"` | 06 |
| `telemetry.auth_secret_ref` | string (secret reference) | `""` | 06 |
| `telemetry.export_interval` | duration | `"60s"` | 06 |
| `telemetry.queue_max_size_mb` | integer | `64` | 06 |
| `telemetry.locked` | bool (system layer) | `false` | 06 |

## Glossary additions

| Term | One-line meaning |
|---|---|
| Configuration layer | One of the ten precedence-ordered value sources of FR-CFG-001, from CLI flags down to built-in defaults. |
| Source attribution | The per-key record of which layer and concrete source (file:line, profile, variable, flag) supplied every resolved value. |
| Reserved root key | A root-level `andromeda.toml` key owned by the schema itself: `config_version`, `include`, `default_profile`. |
| Runtime override | A session-scoped, in-memory, never-persisted key assignment issued through TUI or IPC surfaces (FR-CFG-005). |
| Protected tables | The tables runtime overrides may never touch: `[permissions]`, `[sandbox]`, `[security]`, `[auth]`, `[telemetry]`. |
| Sensitivity class | The per-key schema marker (`public`/`sensitive`) governing whether a value may ever render in any sink (FR-CFG-011). |
| Detector registry | The versioned, compiled-in set of named secret detectors with `definite` (blocking) and `heuristic` (warning) classes. |
| Redaction token | The ASCII placeholder (`[redacted:<class>]`, `[redacted:sha256:<8 hex>]`) that replaces a value in every redacted rendering. |
| Compact / explicit form | The two `ANDROMEDA_*` addressing forms of FR-CFG-004: schema-matched single-underscore names, and `__`-separated path segments. |
| Term | Meaning |
|---|---|
| Event family | The delivery classification of an event name (`lifecycle`, `action`, `progress`, `security`, `telemetry`) that fixes its overflow policy and default buffer (chapter 04) |
| Persist-then-publish | The delivery discipline whereby lifecycle/action events append transactionally with the writes they describe before bus publication (ADR-137) |
| Collected-data catalog | The versioned, user-facing document enumerating every field remote telemetry may carry; its version is the consent version (chapter 06) |
| Consent record | The global-database row created by the interactive telemetry confirmation (consent version, product version, timestamp, endpoint, scope) |
| Cost rollup | A daily aggregate of Cost Records keyed by day, provider, model, workspace, and cost basis; bases and currencies never merge (chapter 05) |
| Allowlist construction | Building telemetry payloads field-by-field from the catalog so prohibited data is unrepresentable rather than filtered (ADR-141) |

## Assumptions

Local list per Volume 0 chapter 05 (global numbers at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | The reference configuration (4 files ≤ 64 KiB, 2 profiles, 20 env vars) represents real installations for NFR-CFG-001 | Volume 12 benchmark corpus review at Beta | Re-derive the reference shape; budget stays, corpus changes |
| Technical assumption | A 1 MiB per-file cap and include bounds (depth 8, 64 files) exceed real configuration sizes by a wide margin | E-CFG-001/E-CFG-007 frequency in field diagnostics | Raise bounds additively (non-breaking) |
| Technical assumption | Shannon entropy ≥ 3.5 bits/char at length ≥ 20 separates secrets from identifiers well enough for a non-blocking heuristic | Detector true/false-positive suites; corpus review per release | Tune threshold or demote/promote detector class via ADR-134 review |
| Technical assumption | ADR-008 round-trip encoding cannot guarantee comment preservation, so rewrites must be backup-first with comment loss documented | Implementation spike on go-toml/v2 encode fidelity | If preservation is reliable, strengthen FR-CFG-008 rule 3 wording additively |

1. **Technical assumption** — the pinned OTel Go SDK provides OTLP exporters for
   `http/protobuf` and `grpc` and view-based metric aggregation, per its official
   documentation at pin time (ADR-011 already commits to the SDK; only the pin's exact
   surface is deferred to implementation).
2. **Technical assumption** — ULIDs' 128-bit binary form maps bijectively onto OTel
   128-bit trace IDs (ADR-027 fixes the ULID form; the mapping uses its 16 bytes directly).
3. **Design decision recorded** — SM-12 (run record completeness) is formalized here as
   NFR-OBS-004 because its measurement is the observability record chain; if fragment A
   also formalizes a storage-side NFR for SM-12, consolidation keeps both (Volume 1 allows
   multiple covering NFRs) or merges by agreement.
4. **Product hypothesis** — opt-in telemetry volume will be low and self-selected
   (ADR-140 accepts this consciously).

## Open questions

Every PENDING VALIDATION in chapters 00–02 and ADR-130..134 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V10A-OQ-1 | Concrete `known_prefix` secret-detector pattern set: which third-party token prefixes are documented by their vendors, and in what format — PENDING VALIDATION per vendor | Chapter 02 (FR-CFG-011); ADR-134 | No — the detector mechanism ships regardless; the pattern set is versioned registry data | Verify against each vendor's published token-format documentation before MVP freeze; record the verified set as registry data with sources | Open |
| # | Question | Where marked | Blocks progress? | Resolution path | Status |
|---|---|---|---|---|---|
| V10B-OQ-1 | Which OpenTelemetry semantic-conventions version to pin span/resource attribute names against (PENDING VALIDATION in chapter 05, span-semantics rule 3) | Chapter 05 | No — `andromeda.*` attributes are self-owned; standard attributes adopt the pinned set at implementation | Verify against the pinned OTel Go SDK release notes at implementation start; record the pin in the implementation manifest | Open |
| V10B-OQ-2 | Whether an upstream slog↔OTel logging bridge has stabilized enough to replace the custom correlating handler (ADR-011 review condition; the custom handler is the committed design) | Chapter 03 (pipeline design; ADR-011) | No — the custom handler is fully specified | Re-check the OTel Go contrib ecosystem at implementation; adopt via ADR-011 amendment only if it simplifies rule 4 | Open |

Both rows exist so every PENDING VALIDATION occurrence in this fragment's chapters has a
register entry (chapter 05 carries the marker for V10B-OQ-1; V10B-OQ-2 tracks an ADR-011
review condition and carries no in-chapter marker).

## Cross-volume references

- Volume 0 chapter 03: config-table ownership map (this volume owns schema/precedence/
  validation for all tables; key content per area owner); environment-variable namespace;
  exit codes.
- Volume 2 chapter 02: Workspace/Project/Configuration Profile entities and INV-CFGP
  invariants enforced here; chapter 10: persistence conventions and the migration model
  FR-CFG-009 executes.
- Volume 3 chapter 02: `ConfigPort` and `SessionStorePort` contracts elaborated here under
  frozen names; `WorkspacePort.Open` exclusivity backed by FR-CFG-009 locking.
- Volume 4/5/6/7: key content of `[agent]`, `[workflows]`, `[providers]`, `[auth]`,
  `[tools]`, `[plugins]`, `[mcp]`, `[skills]`, `[memory]`, `[context]`, `[index]` —
  reproduced in the chapter 01 complete example from their published catalogs.
- Volume 8: CLI grammar for `config` surfaces, flag↔key aliases, the reserved control
  variables of its environment-variable requirement, and `[cli]` key content; `[tui]` keys
  pending 8B are marked illustrative in the example.
- Volume 9: permission model and trust vocabulary referenced by resolution policy;
  redaction rule ownership; keystone FR-SEC-102 for the Secret Store the E-CFG-012 remedy
  points at; `[permissions]`/`[sandbox]`/`[security]` example keys marked illustrative.
- Volume 11 and Volume 14: `[git]`/`[github]` and `[update]` example keys marked
  illustrative pending their catalogs; Volume 14 update rollback pairs with database
  restore per ADR-029 rule 6.
- Volume 12: performance budgets consuming NFR-CFG-001; SM-06 decomposition; reference
  machines.
- Volume 13: verification methods named per requirement (corpora, fault injection,
  property tests).
- Fragment B (chapters 03–06): event envelope (keystone FR-OBS-001) for every event named
  here; logging redaction pipeline shared with FR-CFG-011; `[logging]`/`[telemetry]` keys.

- **Volume 2**: Event, Trace, Metric, Cost Record entities and invariants (chapter 08);
  canonical JSON, export forms, write discipline (chapter 10); recorded status vocabularies
  (chapter 09).
- **Volume 3**: EventBusPort, TelemetryPort, ConfigPort, SessionStorePort shapes (frozen,
  FR-ARCH-003); component contracts for Logging, Telemetry, Observability, Event Bus;
  shutdown budgets (its chapter 08).
- **Volume 4**: run/turn/task semantics whose transitions produce lifecycle events; run
  record stream for NFR-OBS-004.
- **Volume 5**: token/cost acquisition and `provider.cost.recorded`; provider metric
  semantics (owner area PROV in the chapter 05 catalog).
- **Volume 6**: tool/command/file-change metric semantics; ARP and MCP correlation
  metadata propagation (RISK-OBS-002 mitigation).
- **Volume 8**: `logs`, `trace`, `export`, `doctor` command surfaces; presentation of cost
  honesty rules and telemetry confirmation UI.
- **Volume 9**: redaction rule content (FR-SEC-102 keystone), permission mediation of IPC
  subscriptions (FR-SEC-100 keystone), audited-action catalog (telemetry state changes),
  Audit Log independence from bus delivery.
- **Volume 10 fragment A**: configuration schema/precedence/validation (FR-CFG-001
  keystone) for every key minted here; `[storage]` table cohabitation (this fragment mints
  only the `storage.events.*`, `storage.traces.*`, `storage.metrics.*`,
  `storage.cost_records.*` sub-namespaces).
- **Volume 12**: benchmark environments for NFR-OBS-001/002/005; hot-path write budgets.
- **Volume 13**: canary, conformance, crash-injection, leak, and no-exporter suites named
  in verification methods.
- **Volume 14**: SM-20 contract-diff regime for the envelope and structured surfaces.
