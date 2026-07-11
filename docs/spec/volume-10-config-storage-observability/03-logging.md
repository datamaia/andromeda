# 03 — Logging

This chapter is the behavioral contract of the **Logging** component (Volume 3, chapter 06):
structured, redacted, local-first logs implemented with the standard library `log/slog` JSON
handler per ADR-011. Logs are the lowest-level diagnostic channel; they are **not** events
(the Event Bus and the event envelope are chapter [04](04-events-and-envelope.md)) and they
are **not** telemetry (remote export is chapter [06](06-telemetry-and-consent.md) and never
carries log records). Redaction *rules* are owned by Volume 9; this chapter owns the pipeline
that enforces them and the observability-side redaction categories. Configuration schema and
precedence are owned by this volume's configuration chapters (keystone FR-CFG-001); the
`[logging]` key content is minted here.

## Log record schema

Every log record is one JSON object on one line (JSON Lines). The canonical field set:

| Field | Type | Presence | Meaning |
|---|---|---|---|
| `time` | RFC 3339 timestamp (UTC, millisecond precision or finer) | always | Record creation instant |
| `level` | string | always | `DEBUG` \| `INFO` \| `WARN` \| `ERROR` |
| `msg` | string | always | Human-readable message; stable wording is not a contract (error identity lives in `error.code`) |
| `component` | string | always | Emitting component, Volume 3 component names (e.g., `agent_engine`, `tool_runtime`) in `snake_case` |
| `pid` | integer | always | Operating-system process ID |
| `version` | string | always | Andromeda product version (SemVer string) |
| `trace_id` | string | when inside an active span | Correlating trace identifier (chapter [05](05-traces-metrics-costs.md)); injected by the correlating handler, never passed manually |
| `span_id` | string | when inside an active span | Correlating span identifier |
| `workspace_id` | ULID | when workspace-scoped | Correlation to the Workspace |
| `session_id` | ULID | when session-scoped | Correlation to the Session |
| `run_id` | ULID | when run-scoped | Correlation to the Run |
| `tool_invocation_id` | ULID | when inside a tool invocation | Correlation to the Tool Invocation |
| `error.code` | string | when logging a defined error | Stable `E-<AREA>-NNN` code (ADR-016) |
| `error.category` | string | with `error.code` | Envelope category of the error |
| `error.detail` | string | with `error.code` | The envelope's technical message (safe-to-log by definition, ADR-016) |
| `source` | object | only when `logging.include_source = true` | File/line of the call site (development diagnostics) |
| *attributes* | any JSON | per call site | Structured attributes; keys `snake_case`; values pass the redaction handler |

Rules:

1. Records MUST be emitted through the shared slog logger assembled by the Logging component.
   Direct writes of diagnostic text to stdout/stderr outside the Volume 8 stream discipline
   are defects.
2. Correlation fields MUST be injected by handlers from the active context (span context,
   run-scoped logger values) — call sites MUST NOT construct correlation fields by hand, so a
   record can never carry a stale or mismatched ID (SM-13 joins depend on this).
3. Attribute keys listed in the fixed field set above are reserved; a call-site attribute
   using a reserved key is a defect surfaced by the development-build handler.
4. `msg` and all attribute values pass redaction (below) before encoding. There is no
   unredacted sink.

## Levels and level policy

slog's four standard levels are the closed level set. There is no `FATAL`: terminating
failures log at `ERROR` and exit through the ADR-016 error envelope and exit-code mapping.

| Level | Use | Default visibility |
|---|---|---|
| `DEBUG` | Development diagnostics: state-machine transitions in detail, handler decisions, buffer occupancy | File sink only when `logging.level = "debug"`; never stderr unless `logging.stderr_level` lowers |
| `INFO` | Normal operation milestones: component start/stop, workspace open, provider selection, rotation | File sink (default level) |
| `WARN` | Degradations and recoverable anomalies: retries, fallbacks, overflow drops, sink degradation | File + stderr (default `logging.stderr_level`) |
| `ERROR` | Defined errors (`E-<AREA>-NNN`) and unexpected failures | File + stderr |

Level policy rules:

- Hot paths (per-token streaming, per-chunk terminal capture) MUST NOT log per item at
  `INFO` or above; per-item diagnostics are `DEBUG` and SHOULD be aggregated (counts,
  boundaries) rather than per-item.
- Every logged occurrence of a defined error MUST carry `error.code`; logging an error
  without its code is a defect detectable by the Volume 13 log-conformance suite.
- Raising verbosity (`--verbose`, `--debug`, Volume 8) changes *presentation and level
  thresholds only*; it MUST NOT change what is recorded in events, traces, or audit records
  (Volume 1, principles-in-tension rule 1).

## Pipeline and destinations

```mermaid
flowchart LR
    CS[Call sites via slog API] --> COR[Correlating handler]
    COR --> RED[Redaction handler]
    RED --> ENC[JSON encoder]
    ENC --> FILE[JSONL file sink]
    ENC --> STDERR[stderr console handler]
    FILE -- write failure --> FB[Fallback: stderr + drop counter]
```

The diagram shows the fixed handler chain. Components log through the standard slog API
(the sanctioned corpus-wide helper, Volume 3 chapter 06); the **correlating handler** injects
trace/span and entity correlation IDs from the context; the **redaction handler** applies the
rules below to every field; the **JSON encoder** produces the canonical JSON Lines form; two
sinks receive encoded records — the JSONL file sink (canonical, always structured) and the
stderr console handler (human-readable rendering, level-gated by `logging.stderr_level`,
formatted per Volume 8 stream discipline and ADR-026 theming tiers). The constraint arrows:
redaction sits strictly before encoding, so no sink can receive unredacted content; file-sink
failure degrades to stderr with a drop counter and never propagates to the logging call site.

Destinations:

- **File sink.** JSON Lines files under the platform log location resolved through the PAL
  (ADR-022): `<state_dir>/andromeda/logs/`. One file per process instance, named
  `andromeda-<start-instant>-<pid>.jsonl` (start instant in UTC `YYYYMMDDTHHMMSSZ` form), so
  concurrent processes (several workspaces, headless instances) never contend on one file and
  no cross-process file locking is needed on the hot path.
- **stderr.** Human-readable rendering for interactive and CI diagnosis; `WARN` and above by
  default. In `--json` mode (Volume 8), stderr diagnostics remain line-oriented JSON.
- **No network destination exists in the Logging component.** Log records are never exported
  remotely — not even under telemetry consent (chapter 06 collects no log records). This is
  structural: the component has no exporter (ADR-011, ADR-138).

Logging MUST be fully functional offline: no destination, handler, or configuration source
involves the network, and the offline suite (SM-05 method) asserts zero network access from
the logging path.

## Rotation and retention

Decided by ADR-138 (built-in rotation; no third-party logging or rotation dependency,
per ADR-011 rule 5):

1. **Size rotation.** When the active file reaches `logging.file.max_size_mb`, the sink
   closes it, opens a successor with a fresh start-instant name, and emits
   `log.file.rotated` (an `INFO` record and an event, chapter 04). Rotation is atomic from
   the writer's perspective: no record is split across files.
2. **Startup sweep.** At process start (and after each rotation), the Logging component
   deletes the oldest log files that exceed `logging.file.max_files` **or** are older than
   `logging.file.max_age_days`, oldest first. The sweep never deletes the active files of
   live processes (liveness check via PID and file age).
3. **Failure behavior.** If rotation or the sweep fails (permissions, disk full), the sink
   degrades: records continue to stderr, a drop counter increments per lost file record, and
   E-OBS-007 is reported once per degradation episode with the event `log.sink.degraded`.
   Logging failure MUST NOT fail or block the operation being logged.

Retention defaults are deliberately small (logs are diagnostics, not the durable record —
events, traces, and audit records are): 10 files × 32 MiB × 30 days bounds the footprint at
320 MiB per retention window.

## Redaction

The redaction handler enforces, on every record and every field, the categories below.
Rule *content* for secrets and sensitive material is Volume 9's (keystone FR-SEC-102); the
observability pipeline enforces it at the handler layer (ADR-011 rule 4). Replacement
tokens follow Volume 9 FR-SEC-109 — `[REDACTED:<fingerprint>]` for redaction-registry
matches and `[REDACTED]` for every other (structural or pattern) match; the lowercase
`[redacted:<detector>]` renderings of chapter 02's configuration findings are that
surface's own finding grammar, distinct from the FR-SEC-109 sink token.

| Category | Rule |
|---|---|
| Secret material | Never present in any form. Secret values are typed (`SecretValue`, Volume 3) and unloggable by construction; the handler additionally rejects attribute values matching Volume 9 secret patterns and replaces them with the FR-SEC-109 token (`[REDACTED:<fingerprint>]` for registry matches, `[REDACTED]` otherwise) |
| Secret references | Only opaque `secret_ref` identifiers may appear; metadata such as provider association may appear, key material never |
| Environment variables | Values never logged; names only, and only for the `ANDROMEDA_*` namespace |
| File content | Not logged beyond Volume 9 excerpt limits; logs carry paths, sizes, hashes, and counts instead |
| Prompts and model output | Never logged at `INFO` or above; at `DEBUG`, only under Volume 9 content gates with size caps |
| Credentials in URLs | Userinfo and query-credential components of any URL are replaced with `"[REDACTED]"` (FR-SEC-109 pattern-match case) |
| Personal identifiers | OS username and hostname appear only in fields explicitly defined to carry them (none in the canonical set); free-text occurrences are the operator's own content and are not scanned |

Redaction is structural first (typed safe-to-log classification on Core Domain types),
pattern-based second (the handler's matchers), and verified third: the Volume 13 secret-leak
canary suite plants known secret material in configuration, environment, and provider
fixtures, then asserts zero occurrences across log files — failure gates release.

## `[logging]` configuration keys

Key content minted here; schema, precedence, env-var mapping, and validation per this
volume's configuration chapters (FR-CFG-001).

```toml
[logging]
level = "info"            # debug | info | warn | error — file-sink threshold
stderr_level = "warn"     # debug | info | warn | error — stderr threshold
include_source = false    # attach file:line call-site info (development)

[logging.file]
enabled = true            # disable to log to stderr only (CI containers)
max_size_mb = 32          # size-rotation threshold per file
max_files = 10            # retention: maximum files kept
max_age_days = 30         # retention: maximum file age
```

Invalid values are configuration errors (E-CFG family, exit code 3) surfaced by the
configuration chapters' validation; the Logging component never guesses a level.

## Requirements

### FR-OBS-002 — Structured logging pipeline

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Logging
- Affected components: Logging; every component as a logging call site; CLI/TUI (stderr rendering)
- Dependencies: ADR-011, ADR-016, ADR-022; FR-OBS-004; keystone FR-CFG-001 (configuration resolution)
- Related risks: RISK-OBS-001

#### Description

Andromeda MUST emit all diagnostic logging through one shared `log/slog` pipeline whose
canonical output is the JSON Lines record schema of this chapter: fixed fields (`time`,
`level`, `msg`, `component`, `pid`, `version`), handler-injected correlation fields
(`trace_id`, `span_id`, `workspace_id`, `session_id`, `run_id`, `tool_invocation_id`), error
fields carrying ADR-016 identities, and structured call-site attributes. The closed level set
is `DEBUG`/`INFO`/`WARN`/`ERROR`. Sinks are the per-process JSONL file under the ADR-022 log
location and the level-gated stderr console handler. Third-party logging frameworks MUST NOT
be introduced (ADR-011 rule 5).

#### Motivation

Logs are the first diagnostic surface users and contributors reach for. One pipeline with
handler-injected correlation makes every log line joinable with events, traces, and cost
records (SM-13); structured JSON makes logs scriptable without parsing prose (MVP minimum
item 18).

#### Actors

All components (producers); users, contributors, and support tooling (consumers); the CLI
`logs` command family (Volume 8) as the query surface.

#### Preconditions

Configuration resolved (levels, sink settings); PAL log directory resolvable (ADR-022
fallback rules apply otherwise).

#### Main flow

1. A component calls the slog API with a message and structured attributes.
2. The correlating handler injects trace/span and entity IDs from the context.
3. The redaction handler applies FR-OBS-004 rules to all fields.
4. The JSON encoder produces the canonical record; the file sink appends it; the stderr
   handler renders it when at or above `logging.stderr_level`.

#### Alternative flows

- File sink disabled (`logging.file.enabled = false`): records flow to stderr only; the
  record schema is unchanged.
- Record below both thresholds: the record is discarded after level check, before handler
  work (levels gate early; no redaction cost for suppressed records).

#### Edge cases

- Logging before configuration is resolved (early startup): a bootstrap logger with defaults
  (`info`/`warn`, stderr only) applies until resolution, then the assembled pipeline replaces
  it; bootstrap records are re-emitted to the file sink is NOT attempted (they remain
  stderr-only), keeping startup simple and honest.
- Log call with an already-cancelled context: the record is still emitted (logging is not a
  cancellable operation); correlation fields reflect the cancelled context's IDs.
- Two processes on one workspace: distinct per-process files by construction; no interleaving.

#### Inputs

slog calls with attributes; context correlation values; `[logging]` configuration.

#### Outputs

JSONL log files; stderr diagnostics; log volume/drop self-metrics (chapter 05).

#### States

Not applicable — the pipeline is stateless (Volume 3 component contract); no state machine.

#### Errors

E-OBS-007 on sink failure (degradation path, never propagated to the call site). Invalid
`[logging]` configuration is an E-CFG-family error at resolution time (exit code 3).

#### Constraints

No network destination; no third-party logging framework; reserved-key protection; hot-path
level discipline (per-item logging at `DEBUG` only).

#### Security

All records pass FR-OBS-004 redaction before any sink; `error.detail` carries only the
ADR-016 safe technical message; log files inherit the user-only permissions of the ADR-022
state directory.

#### Observability

The pipeline meters itself: `obs.log.records_total{level}` and `obs.log.dropped_total`
(chapter 05 registry); rotation emits `log.file.rotated`; degradation emits
`log.sink.degraded`.

#### Performance

Suppressed-level calls cost no allocation beyond the level check; emission overhead budget is
NFR-OBS-001; Volume 12 owns the benchmark environment.

#### Compatibility

The record schema is a public diagnostic surface: field additions are permitted in minor
releases; renaming or removing canonical fields follows SM-20 deprecation discipline.
Identical schema on all Tier 1 platforms.

#### Acceptance criteria

- Given any run, when its log records are inspected, then every record parses as JSON, carries
  the fixed fields, and every run-scoped record carries `run_id` matching the run.
- Given a log call inside an active span, when the record is emitted, then `trace_id` and
  `span_id` equal the active span context's values.
- Given a defined error is logged, when the record is inspected, then `error.code` and
  `error.category` match the error's ADR-016 envelope.
- Negative case: given a call-site attribute named `run_id` (reserved), when emitted in a
  development build, then the pipeline surfaces a defect diagnostic and the handler-injected
  value wins in the record.
- Permission case: given the log directory is unwritable, when logging proceeds, then records
  flow to stderr, E-OBS-007 is reported once, and the observed operation completes unaffected.
- Observability case: given 1,000 emitted records, when self-metrics are read, then
  `obs.log.records_total` increased by exactly 1,000 across levels.

#### Verification method

Log-conformance suite (schema validation of every record produced by integration and E2E
runs); correlation-join tests against traces and events (SM-13 method); fault-injection on
sink failure; Tier 1 platform matrix (Volume 13).

#### Traceability

PRD-006; Principle 9; ADR-011; MVP minimum item 18; SM-13; FR-OBS-004; NFR-OBS-001.

### FR-OBS-003 — Log rotation and retention

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Design
- Owner: Logging
- Affected components: Logging; PAL (Filesystem, Config Directories)
- Dependencies: FR-OBS-002; ADR-022, ADR-138
- Related risks: RISK-OBS-001

#### Description

The file sink MUST rotate the active log file when it reaches `logging.file.max_size_mb` and
MUST enforce retention at startup and after each rotation by deleting oldest-first the files
exceeding `logging.file.max_files` or older than `logging.file.max_age_days`. Rotation and
retention operate per log directory, never touch files of live processes, and emit
`log.file.rotated` and retention diagnostics. Defaults: 32 MiB, 10 files, 30 days.

#### Motivation

Unbounded logs exhaust disk on long-lived installations (RISK-OBS-001); rotation with
deterministic retention makes the logging footprint a configured, verifiable bound instead of
an accident.

#### Actors

Logging component (automatic); users adjusting `[logging.file]` keys.

#### Preconditions

File sink enabled; log directory resolvable and writable.

#### Main flow

1. The sink tracks the active file's size on each append.
2. On crossing the threshold, it closes the file, opens a successor, emits
   `log.file.rotated`, and triggers the retention sweep.
3. The sweep lists log files, excludes live-process actives, and deletes violations
   oldest-first.

#### Alternative flows

- Startup sweep: identical retention pass at process start before the first append.
- `max_age_days` reached without size rotation: the startup/rotation sweep removes the aged
  file; there is no timer-based mid-run deletion (a run's own active file is never removed
  mid-run).

#### Edge cases

- Clock skew making a file appear future-dated: retention treats files newer than "now" as
  age zero (never deleted by age) and logs a `WARN`.
- A single record larger than the rotation threshold: the record is written whole (records
  are never split), then rotation triggers immediately after.
- Crash between close and successor-open: the next append re-runs successor-open
  idempotently; at most one empty file can result and it is retention-eligible.

#### Inputs

Append sizes; file metadata (age, PID liveness); `[logging.file]` keys.

#### Outputs

Rotated file set within configured bounds; `log.file.rotated` events.

#### States

Not applicable — rotation is a sink-internal procedure, not an entity machine.

#### Errors

E-OBS-007 when rotation or deletion fails; degradation per FR-OBS-002 (stderr fallback,
one report per episode).

#### Constraints

Oldest-first deletion; never delete live actives; no cross-process locks on the append path.

#### Security

Deleting log files never bypasses audit retention: audit records live in the Audit Log
(Volume 9), not in log files (INV-AUD-04 unaffected). Rotated files keep user-only
permissions.

#### Observability

`log.file.rotated` event with old/new file names and sizes in payload;
`obs.log.dropped_total` counts records lost during degradation episodes.

#### Performance

Size tracking is O(1) per append; the sweep is O(files) and runs off the hot path.

#### Compatibility

Behavior identical on Tier 1 platforms; file naming is part of the documented layout, not a
programmatic contract.

#### Acceptance criteria

- Given `max_size_mb = 1` and sustained logging, when 5 MiB of records are emitted, then no
  file exceeds ~1 MiB plus one record and `log.file.rotated` was emitted at each boundary.
- Given 15 accumulated files with `max_files = 10`, when the sweep runs, then the 5 oldest
  are removed and the newest 10 remain.
- Negative case: given a second live process's active file among the oldest, when the sweep
  runs, then that file is not deleted and the sweep reports the exclusion at `DEBUG`.
- Error case: given deletion fails on a read-only directory, when the sweep runs, then
  E-OBS-007 is reported once, logging continues on stderr, and no operation fails.
- Observability case: retention decisions (kept/removed counts) appear as a `DEBUG` record
  after each sweep.

#### Verification method

Rotation unit tests with size fixtures; retention property tests over synthetic file sets
(age × count × liveness); crash-injection between rotation steps; disk-full fault injection
(Volume 13).

#### Traceability

RISK-OBS-001; ADR-138; ADR-022; FR-OBS-002.

### FR-OBS-004 — Log redaction enforcement

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Logging
- Affected components: Logging; all producers; Secret Store (typed material)
- Dependencies: ADR-011 (handler-layer redaction); Volume 9 redaction rules (keystone FR-SEC-102) and replacement-token grammar (FR-SEC-109); FR-OBS-002
- Related risks: RISK-OBS-001

#### Description

Every log record MUST pass the redaction handler before encoding, enforcing the category
table of this chapter: secret material unrepresentable and pattern-rejected, secret
references opaque, environment values never logged, file content bounded by Volume 9 excerpt
gates, prompts and model output excluded at `INFO`+ and content-gated at `DEBUG`, URL
credentials stripped. Redaction applies uniformly to `msg`, attributes, and error fields, on
every sink including stderr. There MUST be no configuration that disables redaction.

#### Motivation

Logs are the leak channel with the most call sites. Central handler enforcement converts
"every author remembers" into "the pipeline guarantees", which is the only version of the
guarantee that survives a large contributor base (Volume 9 threat: log leakage).

#### Actors

Redaction handler (enforcement); Volume 9 (rule content); Volume 13 canary suite
(verification).

#### Preconditions

Handler chain assembled per FR-OBS-002 (redaction strictly before encoding).

#### Main flow

1. A record reaches the redaction handler.
2. Typed safe-to-log classification is honored (secret-typed values render as
   `"[REDACTED]"` unconditionally, the FR-SEC-109 structural case).
3. Pattern matchers scan string fields; matches are replaced with the FR-SEC-109 tokens
   (`[REDACTED:<fingerprint>]` for registry matches, `[REDACTED]` otherwise).
4. The redacted record proceeds to encoding.

#### Alternative flows

- `DEBUG`-level content-bearing record: Volume 9 content gates apply (size caps, category
  allowances) before the generic matchers.

#### Edge cases

- Extremely large attribute values: values above the record size cap (64 KiB per record) are
  truncated with an explicit `"[TRUNCATED <n> bytes]"` marker before pattern scanning of the
  retained prefix — truncation is marked, never silent.
- Binary attribute values: rendered as size + SHA-256, never raw bytes.
- A matcher false-positive redacting benign text: acceptable by policy — redaction errs
  toward removal; the per-category redaction counter (Observability below) records which
  matcher fired so diagnosis remains possible.

#### Inputs

Records from the correlating handler; Volume 9 rule set; typed classifications.

#### Outputs

Redacted records only; redaction counters.

#### States

Not applicable — stateless handler.

#### Errors

A handler-internal failure fails closed: the record is dropped, the drop counter increments,
and a self-record at `WARN` (containing no user data) reports the drop. E-OBS-007 applies if
the failure is sink-related.

#### Constraints

Redaction is not configurable off; placeholders are stable strings; scanning cost is bounded
by the record size cap.

#### Security

This requirement is the logging half of the corpus secret-hygiene guarantee (with Volume 9's
FR-SEC-102): no secret material in any log sink, verified by canary tests planting known
material and asserting zero occurrences across all produced files.

#### Observability

Redaction counters per category (`obs.log.redactions_total{category}`, chapter 05 registry)
make silent-leak regressions visible as counter anomalies.

#### Performance

Pattern scanning bounded by the 64 KiB record cap; NFR-OBS-001 covers total pipeline
overhead.

#### Compatibility

Identical rules on all platforms; rule content evolves via Volume 9 without changing this
pipeline contract.

#### Acceptance criteria

- Given a canary secret planted in provider configuration, when the full integration suite
  runs, then zero occurrences of the canary appear in any log file or stderr capture.
- Given an attribute containing `https://user:pass@host/path`, when logged, then the record
  contains the host and path with `"[REDACTED]"` replacing the userinfo.
- Given a `SecretValue`-typed attribute, when logged at any level, then the rendered value is
  `"[REDACTED]"`.
- Negative case: given redaction handler failure injected, when a record passes, then the
  record is dropped, the counter increments, and no unredacted content reaches any sink.
- Permission case: redaction applies identically in headless mode (ADR-032) — no mode
  weakens it.
- Observability case: each redaction increments its category counter.

#### Verification method

Volume 13 secret-leak canary suite (release gate); property tests over matcher corpus;
fault-injection on the handler; stderr capture audits in CI.

#### Traceability

PRD-005, PRD-006; ADR-011 rule 4; Volume 9 FR-SEC-102 (keystone); SM-16; FR-OBS-002.

## Non-functional requirements

### NFR-OBS-001 — Logging hot-path overhead

- Category: Performance
- Priority: P1
- Phase: Beta
- Metric: (a) Added wall-clock latency per emitted log record through the full handler chain (correlate → redact → encode → file append), p95; (b) added latency per suppressed-level call, p95
- Target: (a) ≤ 250 µs p95; (b) ≤ 1 µs p95
- Minimum threshold: (a) ≤ 500 µs p95; (b) ≤ 5 µs p95
- Measurement method: Benchmark harness emitting representative record mixes (with/without correlation, attribute sizes to the 64 KiB cap) on reference hardware; suppressed-call microbenchmark; measured per release
- Test environment: Volume 12 reference machines (Volume 1 chapter 06 reference conditions)
- Measurement frequency: Every release; regression-gated at Beta exit
- Owner: Logging / Volume 12 (benchmark environment)
- Dependencies: FR-OBS-002, FR-OBS-004
- Risks: RISK-OBS-001
- Acceptance criteria: Benchmark report per release shows both percentiles within target (or above minimum threshold with a recorded justification); no release ships with the suppressed-call path allocating on the heap.

## Errors

### E-OBS-007 — Log sink failure

- **Code**: E-OBS-007. **Category**: storage/degradation. **Severity**: warning.
- **User message**: "Log files cannot be written (<short cause>); diagnostics continue on
  stderr. The operation was not affected."
- **Technical message**: "log file sink failed: <cause: directory unwritable | disk full |
  rotation failure | retention sweep failure> at <path>; degraded to stderr, <n> records
  dropped this episode".
- **Cause**: unwritable or missing log directory, disk exhaustion, or a failed rotation or
  retention operation.
- **Safe context data**: log directory path, cause class, dropped-record count. Never
  record contents.
- **Recoverability**: self-healing — the sink re-probes the file destination once per
  minute and resumes on success; user-recoverable by freeing disk or fixing permissions.
  **Retry policy**: automatic re-probe (60 s interval); the failing append itself is not
  retried.
- **Recommended action**: run `andromeda doctor` (Volume 8) to see the log location and its
  state; free disk space or correct directory permissions.
- **Exit code**: none (logging failure never terminates or fails an operation); 1 only when
  a CLI operation whose primary purpose is log manipulation (export/inspection) fails on
  this cause. **HTTP mapping**: not applicable (no HTTP surface).
- **Telemetry event**: `log.sink.degraded`.
- **Security implications**: degradation never bypasses redaction (stderr receives the same
  redacted records); dropped records are diagnostics only — events, traces, and audit
  records are unaffected, so the durable record's completeness does not depend on this sink.

## Risks

### RISK-OBS-001 — Observability data volume degrades the host

- Category: Technical / operational
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Bounded retention everywhere (FR-OBS-003 for logs; FR-OBS-006 and chapter 05 retention keys for events, traces, metrics, cost rollups); hot-path level discipline; record size caps; ADR-011 sampling knobs held in reserve; Volume 12 benchmarks watch footprint
- Detection: Self-metrics (`obs.log.records_total`, drop and prune counters); `andromeda doctor` (Volume 8) reports observability store sizes against configured bounds; Volume 12 long-session soak tests
- Owner: Volume 10 (Logging, Observability)
- Status: Open

Long agent sessions generate logs, events, spans, and metric points continuously; ADR-011
accepts local storage as the default sink, which converts an egress problem into a disk and
latency problem. The mitigation posture is bounds everywhere plus visible counters, so
exhaustion is a configuration decision, never a surprise.
