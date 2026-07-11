# Volume 10 — Configuration, Storage, and Observability

**Status:** Authored (draft) · **Owner:** Configuration Manager / Persistence Layer /
Observability (Volume 10)

Volume 10 specifies the surfaces every other volume defers to by reference: the
`andromeda.toml` configuration model with its schema, exact precedence, validation,
versioning, migration, and secret hygiene (keystone FR-CFG-001); the operational storage
rules over the ADR-028 database topology (backups, retention, locking, migration
execution per ADR-029); and the observability stack of ADR-011 — structured logging, the
event envelope and delivery semantics (keystone FR-OBS-001), traces, metrics, and cost
observability, and the strictly consent-gated telemetry split. Per Volume 0 chapter 03,
this volume mints all `CFG` and `OBS` identifiers and owns the behavioral contracts of
`ConfigPort`, `SessionStorePort` (storage mechanics), `EventBusPort`, and `TelemetryPort`
(Volume 3, chapter 02).

Foundations assumed: Volume 0 (conventions, identifier ownership, config-table ownership),
Volume 1 (objectives, principles, phases, SM-12/SM-13 metrics), Volume 2 (entities,
persistence conventions, chapter 10 model-facing rules), Volume 3 (ports, components,
PAL). Key *content* of foreign TOML tables belongs to each table's area owner; this volume
owns schema, precedence, and validation for all of them.

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Configuration Model](01-configuration-model.md) | Keystone FR-CFG-001; layers and exact precedence, complete `andromeda.toml` example, profiles, environment-variable mapping algorithm, CLI/runtime overrides, includes, defaults, validation per ADR-024, schema versioning and forward-only migration, deprecations; storage backups, retention, and workspace locking |
| [02 — Configuration Errors and Secret Redaction](02-config-errors-and-redaction.md) | Error-reporting contract with exact messages, minimal and invalid examples, the complete E-CFG catalog under the ADR-016 envelope, secret detection and universal sink redaction |
| [03 — Logging](03-logging.md) | slog JSON logging per ADR-011: record schema, levels, pipeline, rotation and retention, redaction enforcement, `[logging]` keys, E-OBS logging errors |
| [04 — Events and Envelope](04-events-and-envelope.md) | Keystone FR-OBS-001; the event envelope, registration, families and overflow policy, delivery semantics, IPC bridging, persistence and retention |
| [05 — Traces, Metrics, and Costs](05-traces-metrics-costs.md) | OTel trace model over runs, metric registry and naming, latency/token/cost/retry metrics, cost observability, SM-13 formalization |
| [06 — Telemetry and Consent](06-telemetry-and-consent.md) | Local/remote pipeline split, consent model, collected vs prohibited data, retention, export, deletion, enterprise lock, OTLP export |
| [99 — Volume Register](99-volume-register.md) | Everything this volume minted (merged from the per-agent authoring fragments at the Phase B gate) |

## Reading guide

1. Chapter 01 is the hub every key-minting volume points at: read the layer table and
   FR-CFG-001 before any per-table catalog elsewhere in the corpus.
2. Chapter 02 completes chapter 01 — the two are one contract split between model and
   failure behavior; the E-CFG catalog covers both configuration (exit code 3) and
   storage-integrity (exit code 9) failures.
3. Chapters 03–06 form the observability stack in dependency order: logs are the floor,
   the envelope of chapter 04 is what every `Events minted` table in the corpus binds to,
   chapter 05 correlates signals into traces/metrics/costs, and chapter 06 governs the
   only path any signal may leave the machine.
4. State names used here are Volume 2 chapter 09's; this volume owns no entity state
   machine — its contracts are ports, procedures, and pipelines.
