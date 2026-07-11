# Annex — Consolidated Configuration Key Catalog

**Status:** Consolidated (Phase C). This annex is the corpus-wide index of every
`andromeda.toml` configuration key minted in Volumes 4–14, grouped by TOML table per the
Volume 0 chapter 03 table-ownership map. It is a *reference view*: the schema, precedence,
environment-variable mapping algorithm, validation, versioning, and migration rules are
Volume 10 chapter [01](../volume-10-config-storage-observability/01-configuration-model.md)'s
(keystone FR-CFG-001); each key's semantics, ranges, and cross-key rules live in the linked
defining chapter, which is normative for that key's content. This annex mints nothing and
renames nothing.

## Precedence, in one paragraph (FR-CFG-001)

Effective configuration resolves by merging ten layers, highest first: **(1) CLI flag →
(2) environment variable (`ANDROMEDA_*`) → (3) runtime override → (4) project profile →
(5) project file (`<project_root>/andromeda.toml`) → (6) workspace profile → (7) workspace
file (`<workspace_root>/.andromeda/andromeda.toml`) → (8) global profile → (9) global file
(`<config_dir>/andromeda/andromeda.toml`, ADR-022) → (10) built-in defaults.** Merge
granularity is the key; tables merge key-wise; arrays and array-of-tables replace wholesale
across layers; absent layers are skipped; every resolved value carries source attribution.
A profile overrides the file of its own scope and is overridden by every nearer scope
([ADR-130](adr/ADR-130.md)). Runtime overrides targeting the protected tables
`[permissions]`, `[sandbox]`, `[security]`, `[auth]`, and `[telemetry]` are rejected with
E-CFG-014 (FR-CFG-005). Validation is typed and strict — unknown keys are E-CFG-003, and
every key has a compiled-in default so the schema is total (FR-CFG-007).

## Environment-variable forms

The environment column below shows the canonical documented form per FR-CFG-004
([ADR-131](adr/ADR-131.md)): the dotted key path uppercased with `.` → `_` (compact form),
or the explicit `__`-separated form for keys inside dynamic-name tables, where
`<SLUG>`/`<NAME>`/`<PATTERN>`/`<MODEL>` stand for the declared entry name. Rules that this
catalog does not restate per row:

- The reserved control variables `ANDROMEDA_CONFIG`, `ANDROMEDA_WORKSPACE`,
  `ANDROMEDA_PROFILE`, `ANDROMEDA_NO_INPUT`, and `ANDROMEDA_NO_COLOR` never map to
  configuration keys (FR-CFG-004 rule 1; execution controls per Volume 8).
- A compact form that becomes ambiguous at resolution time (e.g., against a declared
  dynamic-table entry) is rejected with E-CFG-009 and must use the explicit `__` form.
- Environment variables override existing dynamic-table entries; they never create them.
- Table-typed keys are not settable from the environment (FR-CFG-004 rule 5); their
  environment cell is "—".

## Reading the tables

One table per TOML table, in the Volume 0 chapter 03 ownership order, with the reserved
root keys first. Columns: the key (dotted path); the schema type (Volume 10 value-type
vocabulary); the built-in default (layer 10) — a "—" default means the defining chapter
declares the key per entry (required or optional fields of dynamic-name or array-of-tables
entries) without a compiled scalar default; the canonical environment-variable form; the
owning volume per the ownership map; and the defining chapter.

## Reserved root keys — owner: Volume 10

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `config_version` | integer | current schema version | — (per-document declaration) | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `include` | array of paths | `[]` | — (per-document declaration) | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `default_profile` | string | `""` | — (profile selection uses the reserved `ANDROMEDA_PROFILE` control variable, FR-CFG-003) | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |

## `[agent]` — owner: Volume 4

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `agent.default_profile` | string | `"default"` | `ANDROMEDA_AGENT_DEFAULT_PROFILE` | Volume 4 | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.session.idle_suspend_after` | duration | `"0s"` | `ANDROMEDA_AGENT_SESSION_IDLE_SUSPEND_AFTER` | Volume 4 | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.loop.max_iterations` | integer | `50` | `ANDROMEDA_AGENT_LOOP_MAX_ITERATIONS` | Volume 4 | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.loop.turn_timeout` | duration | `"5m"` | `ANDROMEDA_AGENT_LOOP_TURN_TIMEOUT` | Volume 4 | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.loop.max_repair_attempts` | integer | `2` | `ANDROMEDA_AGENT_LOOP_MAX_REPAIR_ATTEMPTS` | Volume 4 | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.loop.max_subagent_depth` | integer | `2` | `ANDROMEDA_AGENT_LOOP_MAX_SUBAGENT_DEPTH` | Volume 4 | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.loop.delegation_enabled` | boolean | `false` | `ANDROMEDA_AGENT_LOOP_DELEGATION_ENABLED` | Volume 4 | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| `agent.planner.approval_mode` | enum (`always` \| `policy` \| `never`) | `"policy"` | `ANDROMEDA_AGENT_PLANNER_APPROVAL_MODE` | Volume 4 | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `agent.planner.max_attempts` | integer | `3` | `ANDROMEDA_AGENT_PLANNER_MAX_ATTEMPTS` | Volume 4 | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `agent.planner.attempt_timeout` | duration | `"2m"` | `ANDROMEDA_AGENT_PLANNER_ATTEMPT_TIMEOUT` | Volume 4 | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `agent.planner.max_tasks_per_plan` | integer | `30` | `ANDROMEDA_AGENT_PLANNER_MAX_TASKS_PER_PLAN` | Volume 4 | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `agent.planner.max_revisions` | integer | `10` | `ANDROMEDA_AGENT_PLANNER_MAX_REVISIONS` | Volume 4 | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| `agent.execution.max_parallel_tasks` | integer | `1` | `ANDROMEDA_AGENT_EXECUTION_MAX_PARALLEL_TASKS` | Volume 4 | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `agent.execution.task_timeout` | duration | `"30m"` | `ANDROMEDA_AGENT_EXECUTION_TASK_TIMEOUT` | Volume 4 | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `agent.execution.retry.max_attempts` | integer | `3` | `ANDROMEDA_AGENT_EXECUTION_RETRY_MAX_ATTEMPTS` | Volume 4 | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `agent.execution.retry.base_delay` | duration | `"1s"` | `ANDROMEDA_AGENT_EXECUTION_RETRY_BASE_DELAY` | Volume 4 | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `agent.execution.retry.max_delay` | duration | `"60s"` | `ANDROMEDA_AGENT_EXECUTION_RETRY_MAX_DELAY` | Volume 4 | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `agent.execution.retry.multiplier` | float | `2.0` | `ANDROMEDA_AGENT_EXECUTION_RETRY_MULTIPLIER` | Volume 4 | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| `agent.prompts.allow_workspace_overrides` | boolean | `false` | `ANDROMEDA_AGENT_PROMPTS_ALLOW_WORKSPACE_OVERRIDES` | Volume 4 | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |
| `agent.prompts.override_dirs` | array of paths | `[]` | `ANDROMEDA_AGENT_PROMPTS_OVERRIDE_DIRS` | Volume 4 | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |
| `agent.prompts.max_render_bytes` | integer | `262144` | `ANDROMEDA_AGENT_PROMPTS_MAX_RENDER_BYTES` | Volume 4 | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |

## `[workflows]` — owner: Volume 4

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `workflows.paths` | array of paths | `[]` | `ANDROMEDA_WORKFLOWS_PATHS` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflows.default_step_timeout` | duration | `"30m"` | `ANDROMEDA_WORKFLOWS_DEFAULT_STEP_TIMEOUT` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflows.default_gate_expiry` | duration | `"24h"` | `ANDROMEDA_WORKFLOWS_DEFAULT_GATE_EXPIRY` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflows.default_max_attempts` | integer | `1` | `ANDROMEDA_WORKFLOWS_DEFAULT_MAX_ATTEMPTS` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflows.max_parallel_steps` | integer | `4` | `ANDROMEDA_WORKFLOWS_MAX_PARALLEL_STEPS` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflows.max_run_duration` | duration | `"168h"` | `ANDROMEDA_WORKFLOWS_MAX_RUN_DURATION` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflows.artifacts_dir` | path | `".andromeda/artifacts"` | `ANDROMEDA_WORKFLOWS_ARTIFACTS_DIR` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| `workflows.sdd.gate_profile` | enum (`strict` \| `standard` \| `minimal`) | `"standard"` | `ANDROMEDA_WORKFLOWS_SDD_GATE_PROFILE` | Volume 4 | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |

## `[providers]` and `[providers.*]` — owner: Volume 5

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `providers.default` | string (provider slug) | `""` (no built-in default provider) | `ANDROMEDA_PROVIDERS_DEFAULT` | Volume 5 | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `providers.discovery_ttl_hours` | integer | `24` | `ANDROMEDA_PROVIDERS_DISCOVERY_TTL_HOURS` | Volume 5 | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `providers.<slug>.adapter` | string (adapter slug) | — (required per entry) | `ANDROMEDA_PROVIDERS__<SLUG>__ADAPTER` | Volume 5 | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `providers.<slug>.enabled` | boolean | `true` | `ANDROMEDA_PROVIDERS__<SLUG>__ENABLED` | Volume 5 | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `providers.<slug>.base_url` | string (URL) | adapter's documented default endpoint (required for `openai_compatible`) | `ANDROMEDA_PROVIDERS__<SLUG>__BASE_URL` | Volume 5 | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `providers.<slug>.default_model` | string | `""` (models are user-selected, never assumed) | `ANDROMEDA_PROVIDERS__<SLUG>__DEFAULT_MODEL` | Volume 5 | [Vol 5 ch 01](../volume-05-providers-and-auth/01-provider-contract.md) |
| `providers.<slug>.auth_profile` | string (profile name) | `""` (falls back to `auth.default_profile`) | `ANDROMEDA_PROVIDERS__<SLUG>__AUTH_PROFILE` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `providers.<slug>.api_key_env` | string (variable name) | `""` | `ANDROMEDA_PROVIDERS__<SLUG>__API_KEY_ENV` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `providers.<slug>.verify_capabilities` | enum (`off` \| `basic` \| `probe`) | `"basic"` | `ANDROMEDA_PROVIDERS__<SLUG>__VERIFY_CAPABILITIES` | Volume 5 | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `providers.<slug>.reverify_s` | integer (seconds) | `300` | `ANDROMEDA_PROVIDERS__<SLUG>__REVERIFY_S` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.capability_overrides."<pattern>".add` | array of capability names | `[]` (additions only where the declaration marks the capability configurable) | `ANDROMEDA_PROVIDERS__<SLUG>__CAPABILITY_OVERRIDES__<PATTERN>__ADD` | Volume 5 | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `providers.<slug>.capability_overrides."<pattern>".remove` | array of capability names | `[]` | `ANDROMEDA_PROVIDERS__<SLUG>__CAPABILITY_OVERRIDES__<PATTERN>__REMOVE` | Volume 5 | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `providers.<slug>.structured_outputs.allow_prompted` | boolean | `false` | `ANDROMEDA_PROVIDERS__<SLUG>__STRUCTURED_OUTPUTS__ALLOW_PROMPTED` | Volume 5 | [Vol 5 ch 03](../volume-05-providers-and-auth/03-streaming-toolcalling-structured-outputs.md) |
| `providers.<slug>.structured_outputs.validation_retries` | integer | `1` | `ANDROMEDA_PROVIDERS__<SLUG>__STRUCTURED_OUTPUTS__VALIDATION_RETRIES` | Volume 5 | [Vol 5 ch 03](../volume-05-providers-and-auth/03-streaming-toolcalling-structured-outputs.md) |
| `providers.<slug>.pricing."<model>".input_per_million_micros` | integer (micro-units) | — (per entry) | `ANDROMEDA_PROVIDERS__<SLUG>__PRICING__<MODEL>__INPUT_PER_MILLION_MICROS` | Volume 5 | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `providers.<slug>.pricing."<model>".output_per_million_micros` | integer (micro-units) | — (per entry) | `ANDROMEDA_PROVIDERS__<SLUG>__PRICING__<MODEL>__OUTPUT_PER_MILLION_MICROS` | Volume 5 | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `providers.<slug>.pricing."<model>".cached_input_per_million_micros` | integer (micro-units) | — (per entry) | `ANDROMEDA_PROVIDERS__<SLUG>__PRICING__<MODEL>__CACHED_INPUT_PER_MILLION_MICROS` | Volume 5 | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `providers.<slug>.pricing."<model>".reasoning_per_million_micros` | integer (micro-units) | — (per entry) | `ANDROMEDA_PROVIDERS__<SLUG>__PRICING__<MODEL>__REASONING_PER_MILLION_MICROS` | Volume 5 | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `providers.<slug>.pricing."<model>".currency` | string (ISO 4217) | — (per entry) | `ANDROMEDA_PROVIDERS__<SLUG>__PRICING__<MODEL>__CURRENCY` | Volume 5 | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `providers.<slug>.pricing."<model>".source` | string | — (required per entry, INV-MDL-04) | `ANDROMEDA_PROVIDERS__<SLUG>__PRICING__<MODEL>__SOURCE` | Volume 5 | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `providers.<slug>.pricing."<model>".effective_date` | string (date) | — (required per entry, INV-MDL-04) | `ANDROMEDA_PROVIDERS__<SLUG>__PRICING__<MODEL>__EFFECTIVE_DATE` | Volume 5 | [Vol 5 ch 04](../volume-05-providers-and-auth/04-token-and-cost-accounting.md) |
| `providers.<slug>.limits.max_concurrent_requests` | integer | `4` | `ANDROMEDA_PROVIDERS__<SLUG>__LIMITS__MAX_CONCURRENT_REQUESTS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.limits.requests_per_minute` | integer | unset (no local pacing) | `ANDROMEDA_PROVIDERS__<SLUG>__LIMITS__REQUESTS_PER_MINUTE` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.timeouts.connect_ms` | integer (ms) | `10000` | `ANDROMEDA_PROVIDERS__<SLUG>__TIMEOUTS__CONNECT_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.timeouts.request_ms` | integer (ms) | `120000` | `ANDROMEDA_PROVIDERS__<SLUG>__TIMEOUTS__REQUEST_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.timeouts.first_token_ms` | integer (ms) | `60000` | `ANDROMEDA_PROVIDERS__<SLUG>__TIMEOUTS__FIRST_TOKEN_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.timeouts.stream_idle_ms` | integer (ms) | `60000` | `ANDROMEDA_PROVIDERS__<SLUG>__TIMEOUTS__STREAM_IDLE_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.timeouts.stream_total_ms` | integer (ms) | `600000` | `ANDROMEDA_PROVIDERS__<SLUG>__TIMEOUTS__STREAM_TOTAL_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.timeouts.discovery_ms` | integer (ms) | `30000` | `ANDROMEDA_PROVIDERS__<SLUG>__TIMEOUTS__DISCOVERY_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.timeouts.embed_ms` | integer (ms) | `60000` | `ANDROMEDA_PROVIDERS__<SLUG>__TIMEOUTS__EMBED_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.retry.max_attempts` | integer | `3` | `ANDROMEDA_PROVIDERS__<SLUG>__RETRY__MAX_ATTEMPTS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.retry.base_delay_ms` | integer (ms) | `500` | `ANDROMEDA_PROVIDERS__<SLUG>__RETRY__BASE_DELAY_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.retry.backoff_multiplier` | float | `2.0` | `ANDROMEDA_PROVIDERS__<SLUG>__RETRY__BACKOFF_MULTIPLIER` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.retry.max_delay_ms` | integer (ms) | `10000` | `ANDROMEDA_PROVIDERS__<SLUG>__RETRY__MAX_DELAY_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.retry.retry_after_cap_ms` | integer (ms) | `60000` | `ANDROMEDA_PROVIDERS__<SLUG>__RETRY__RETRY_AFTER_CAP_MS` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.breaker.enabled` | boolean | `true` | `ANDROMEDA_PROVIDERS__<SLUG>__BREAKER__ENABLED` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.breaker.failure_threshold` | integer | `5` | `ANDROMEDA_PROVIDERS__<SLUG>__BREAKER__FAILURE_THRESHOLD` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.breaker.failure_ratio` | float | `0.5` | `ANDROMEDA_PROVIDERS__<SLUG>__BREAKER__FAILURE_RATIO` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.breaker.min_samples` | integer | `10` | `ANDROMEDA_PROVIDERS__<SLUG>__BREAKER__MIN_SAMPLES` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.breaker.window_s` | integer (seconds) | `60` | `ANDROMEDA_PROVIDERS__<SLUG>__BREAKER__WINDOW_S` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.breaker.open_base_s` | integer (seconds) | `30` | `ANDROMEDA_PROVIDERS__<SLUG>__BREAKER__OPEN_BASE_S` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.<slug>.breaker.open_max_s` | integer (seconds) | `600` | `ANDROMEDA_PROVIDERS__<SLUG>__BREAKER__OPEN_MAX_S` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.routing.strategy` | enum (`explicit` \| `preference_list`) | `"explicit"` | `ANDROMEDA_PROVIDERS_ROUTING_STRATEGY` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.routing.preference` | array of provider slugs | `[]` | `ANDROMEDA_PROVIDERS_ROUTING_PREFERENCE` | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains` (array of tables) | array of tables | `[]` (explicit chains only — no implicit fallback) | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains[].name` | string | — (per entry) | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains[].from` | string (source provider slug) | — (per entry) | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains[].targets` | array of provider slugs | — (per entry) | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains[].triggers` | array of normalized failure classes | — (per entry; vocabulary: `unreachable`, `auth_failed`, `rate_limited`, `quota_exhausted`, `timeout`, `internal_error`, `breaker_open`, `capability_gap`) | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains[].allow_local_to_cloud` | boolean | `false` | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains[].max_price_multiplier` | float | `1.0` | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |
| `providers.fallback.chains[].require_approval` | boolean | `false` | — | Volume 5 | [Vol 5 ch 05](../volume-05-providers-and-auth/05-resilience-routing-fallback.md) |

Adapter-owned keys beyond this closed set validate against the adapter's declared
`ConfigSchema` (Vol 5 ch 01).

## `[auth]` — owner: Volume 5

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `auth.default_profile` | string (profile name) | `""` | `ANDROMEDA_AUTH_DEFAULT_PROFILE` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.refresh_lead_time_seconds` | integer (seconds) | `300` | `ANDROMEDA_AUTH_REFRESH_LEAD_TIME_SECONDS` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.flow_timeout_seconds` | integer (seconds) | `300` | `ANDROMEDA_AUTH_FLOW_TIMEOUT_SECONDS` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.proxy.url` | string (URL) | `""` (empty honors `HTTPS_PROXY`/`HTTP_PROXY`) | `ANDROMEDA_AUTH_PROXY_URL` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.proxy.no_proxy` | string | `""` (empty honors `NO_PROXY`) | `ANDROMEDA_AUTH_PROXY_NO_PROXY` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.proxy.credential` | string (credential label — never secret material) | `""` | `ANDROMEDA_AUTH_PROXY_CREDENTIAL` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.proxy.ca_bundle` | path (PEM file) | `""` | `ANDROMEDA_AUTH_PROXY_CA_BUNDLE` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.profiles.<name>.provider` | string (provider slug) | — (required per profile) | `ANDROMEDA_AUTH__PROFILES__<NAME>__PROVIDER` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| `auth.profiles.<name>.credential` | string (credential label) | `""` (empty for `auth_kind = none`) | `ANDROMEDA_AUTH__PROFILES__<NAME>__CREDENTIAL` | Volume 5 | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |

## `[tools]` — owner: Volume 6

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `tools.default_timeout_ms` | integer (ms) | `60000` | `ANDROMEDA_TOOLS_DEFAULT_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.max_timeout_ms` | integer (ms) | `600000` | `ANDROMEDA_TOOLS_MAX_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.max_output_bytes` | integer | `1048576` | `ANDROMEDA_TOOLS_MAX_OUTPUT_BYTES` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.max_concurrent_invocations` | integer | `8` | `ANDROMEDA_TOOLS_MAX_CONCURRENT_INVOCATIONS` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.teardown_grace_ms` | integer (ms) | `2000` | `ANDROMEDA_TOOLS_TEARDOWN_GRACE_MS` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.teardown_kill_ms` | integer (ms) | `3000` | `ANDROMEDA_TOOLS_TEARDOWN_KILL_MS` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.max_auto_retries` | integer | `2` | `ANDROMEDA_TOOLS_MAX_AUTO_RETRIES` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.disabled` | array of strings (tool name selectors) | `[]` | `ANDROMEDA_TOOLS_DISABLED` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.allowed_origins` | array of strings | `["builtin", "plugin", "mcp"]` | `ANDROMEDA_TOOLS_ALLOWED_ORIGINS` | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| `tools.aliases` | table (alias → canonical name) | empty table | — | Volume 6 | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |

## `[plugins]` — owner: Volume 6

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `plugins.enabled` | boolean | `true` | `ANDROMEDA_PLUGINS_ENABLED` | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.handshake_timeout_ms` | integer (ms) | `10000` | `ANDROMEDA_PLUGINS_HANDSHAKE_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.request_timeout_ms` | integer (ms) | `60000` | `ANDROMEDA_PLUGINS_REQUEST_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.stop_timeout_ms` | integer (ms) | `5000` | `ANDROMEDA_PLUGINS_STOP_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.health_interval_ms` | integer (ms) | `30000` | `ANDROMEDA_PLUGINS_HEALTH_INTERVAL_MS` | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.restart_max_attempts` | integer | `5` (per 10-minute window) | `ANDROMEDA_PLUGINS_RESTART_MAX_ATTEMPTS` | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.restart_backoff_initial_ms` | integer (ms) | `500` (doubles, capped at 30000) | `ANDROMEDA_PLUGINS_RESTART_BACKOFF_INITIAL_MS` | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.sources` | array of tables (source entry schema below) | `[]` | — | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| `plugins.overrides.<name>` | table (per-plugin overrides: `autostart`, `idle_stop_minutes`, timeout keys) | — (per entry) | — | Volume 6 | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |

## `[mcp]` — owner: Volume 6

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `mcp.connect_timeout_ms` | integer (ms) | `10000` | `ANDROMEDA_MCP_CONNECT_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.initialize_timeout_ms` | integer (ms) | `10000` | `ANDROMEDA_MCP_INITIALIZE_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.request_timeout_ms` | integer (ms) | `60000` | `ANDROMEDA_MCP_REQUEST_TIMEOUT_MS` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.reconnect_max_attempts` | integer | `5` | `ANDROMEDA_MCP_RECONNECT_MAX_ATTEMPTS` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.reconnect_backoff_initial_ms` | integer (ms) | `1000` (doubles per attempt, capped at 30000) | `ANDROMEDA_MCP_RECONNECT_BACKOFF_INITIAL_MS` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.log_capture` | boolean | `true` | `ANDROMEDA_MCP_LOG_CAPTURE` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.transport` | enum (`stdio` \| `streamable_http`) | — (required per registration) | `ANDROMEDA_MCP__SERVERS__<NAME>__TRANSPORT` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.command` | string | — (stdio transport only) | `ANDROMEDA_MCP__SERVERS__<NAME>__COMMAND` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.args` | array of strings | `[]` | `ANDROMEDA_MCP__SERVERS__<NAME>__ARGS` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.env_allowlist` | array of strings | `[]` | `ANDROMEDA_MCP__SERVERS__<NAME>__ENV_ALLOWLIST` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.url` | string (URL) | — (HTTP transport only) | `ANDROMEDA_MCP__SERVERS__<NAME>__URL` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.headers` | table | — (never secret material inline — E-MCP-007) | — | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.credential` | string (credential reference name) | `""` | `ANDROMEDA_MCP__SERVERS__<NAME>__CREDENTIAL` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.enabled` | boolean | — (per registration) | `ANDROMEDA_MCP__SERVERS__<NAME>__ENABLED` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.scope_hint` | string | — (per registration) | `ANDROMEDA_MCP__SERVERS__<NAME>__SCOPE_HINT` | Volume 6 | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| `mcp.servers.<name>.expose_tools` | array of strings (least-exposure allowlist) | all-discovered when absent | `ANDROMEDA_MCP__SERVERS__<NAME>__EXPOSE_TOOLS` | Volume 6 | [Vol 6 ch 06](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) |

Every `[mcp]` runtime-wide default above is additionally overridable per server inside its
`[mcp.servers.<name>]` table (Vol 6 ch 05).

## `[skills]` — owner: Volume 6

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `skills.enabled` | boolean | `true` | `ANDROMEDA_SKILLS_ENABLED` | Volume 6 | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skills.paths` | array of paths | `[]` | `ANDROMEDA_SKILLS_PATHS` | Volume 6 | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skills.autoload` | boolean | `true` | `ANDROMEDA_SKILLS_AUTOLOAD` | Volume 6 | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skills.sources` | array of tables (source entry schema below) | `[]` | — | Volume 6 | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| `skills.activation_policy` | enum (`prompt` \| `allow` \| `deny`) | `"prompt"` | `ANDROMEDA_SKILLS_ACTIVATION_POLICY` | Volume 6 | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |

### Source entry fields (shared schema of `plugins.sources` and `skills.sources` entries)

| Field | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `name` | string | — (required; unique per scope) | — | Volume 6 | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `kind` | enum (`registry` \| `git` \| `archive` \| `path`) | — (required) | — | Volume 6 | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `location` | string | — (required) | — | Volume 6 | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `priority` | integer | `100` (lower consults first) | — | Volume 6 | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `enabled` | boolean | `true` | — | Volume 6 | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `signature_required` | boolean | `false` | — | Volume 6 | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| `timeout_ms` | integer (ms) | `300000` | — | Volume 6 | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |

## `[memory]` — owner: Volume 7

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `memory.enabled` | boolean | `true` | `ANDROMEDA_MEMORY_ENABLED` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.max_content_bytes` | integer | `16384` | `ANDROMEDA_MEMORY_MAX_CONTENT_BYTES` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.ingestion.mode` | enum (`explicit` \| `assisted` \| `off`) | `"assisted"` | `ANDROMEDA_MEMORY_INGESTION_MODE` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.session_days` | integer | `90` | `ANDROMEDA_MEMORY_RETENTION_SESSION_DAYS` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.workspace_days` | integer | `365` | `ANDROMEDA_MEMORY_RETENTION_WORKSPACE_DAYS` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.long_term_days` | integer | `0` (keep forever) | `ANDROMEDA_MEMORY_RETENTION_LONG_TERM_DAYS` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.archive_before_expire` | boolean | `true` | `ANDROMEDA_MEMORY_RETENTION_ARCHIVE_BEFORE_EXPIRE` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.archive_grace_days` | integer | `30` | `ANDROMEDA_MEMORY_RETENTION_ARCHIVE_GRACE_DAYS` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.purge_after_days` | integer | `30` | `ANDROMEDA_MEMORY_RETENTION_PURGE_AFTER_DAYS` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.retention.importance_protect_threshold` | integer | `8` | `ANDROMEDA_MEMORY_RETENTION_IMPORTANCE_PROTECT_THRESHOLD` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.ranking.weight_relevance` | float | `0.4` | `ANDROMEDA_MEMORY_RANKING_WEIGHT_RELEVANCE` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.ranking.weight_recency` | float | `0.3` | `ANDROMEDA_MEMORY_RANKING_WEIGHT_RECENCY` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.ranking.weight_importance` | float | `0.15` | `ANDROMEDA_MEMORY_RANKING_WEIGHT_IMPORTANCE` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.ranking.weight_trust` | float | `0.15` | `ANDROMEDA_MEMORY_RANKING_WEIGHT_TRUST` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.consolidation.enabled` | boolean | `false` | `ANDROMEDA_MEMORY_CONSOLIDATION_ENABLED` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.consolidation.max_records_per_pass` | integer | `200` | `ANDROMEDA_MEMORY_CONSOLIDATION_MAX_RECORDS_PER_PASS` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| `memory.consolidation.max_tokens_per_pass` | integer | `50000` | `ANDROMEDA_MEMORY_CONSOLIDATION_MAX_TOKENS_PER_PASS` | Volume 7 | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |

## `[context]` — owner: Volume 7

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `context.budget.reserve_output_tokens` | integer | `2048` | `ANDROMEDA_CONTEXT_BUDGET_RESERVE_OUTPUT_TOKENS` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.budget.safety_margin_ratio` | float | `0.05` | `ANDROMEDA_CONTEXT_BUDGET_SAFETY_MARGIN_RATIO` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.history.max_turns` | integer | `50` | `ANDROMEDA_CONTEXT_HISTORY_MAX_TURNS` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.history.max_tokens_ratio` | float | `0.5` | `ANDROMEDA_CONTEXT_HISTORY_MAX_TOKENS_RATIO` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.tool_results.max_tokens_per_result` | integer | `4000` | `ANDROMEDA_CONTEXT_TOOL_RESULTS_MAX_TOKENS_PER_RESULT` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.files.max_file_bytes` | integer | `262144` | `ANDROMEDA_CONTEXT_FILES_MAX_FILE_BYTES` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.files.excerpt_tokens` | integer | `800` | `ANDROMEDA_CONTEXT_FILES_EXCERPT_TOKENS` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.compaction.min_tokens` | integer | `64` | `ANDROMEDA_CONTEXT_COMPACTION_MIN_TOKENS` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.pinning.max_ratio` | float | `0.5` | `ANDROMEDA_CONTEXT_PINNING_MAX_RATIO` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.summarization.use_memory_summaries` | boolean | `false` | `ANDROMEDA_CONTEXT_SUMMARIZATION_USE_MEMORY_SUMMARIES` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| `context.models.<name>.max_input_tokens` | integer | — (per entry; tighten-only per-model override) | `ANDROMEDA_CONTEXT__MODELS__<NAME>__MAX_INPUT_TOKENS` | Volume 7 | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |

## `[index]` — owner: Volume 7

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `index.enabled` | boolean | `true` | `ANDROMEDA_INDEX_ENABLED` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.include` | array of strings (globs) | `[]` | `ANDROMEDA_INDEX_INCLUDE` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.exclude` | array of strings (globs) | `[]` | `ANDROMEDA_INDEX_EXCLUDE` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.max_file_bytes` | integer | `1048576` | `ANDROMEDA_INDEX_MAX_FILE_BYTES` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.max_chunks` | integer | `100000` | `ANDROMEDA_INDEX_MAX_CHUNKS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.on_scale_exceeded` | enum (`degrade` \| `refuse`) | `"degrade"` | `ANDROMEDA_INDEX_ON_SCALE_EXCEEDED` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.chunk.target_tokens` | integer | `400` | `ANDROMEDA_INDEX_CHUNK_TARGET_TOKENS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.chunk.max_tokens` | integer | `512` | `ANDROMEDA_INDEX_CHUNK_MAX_TOKENS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.chunk.overlap_tokens` | integer | `40` | `ANDROMEDA_INDEX_CHUNK_OVERLAP_TOKENS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.semantic.provider` | string (provider slug) | `""` (empty = semantic indexing off) | `ANDROMEDA_INDEX_SEMANTIC_PROVIDER` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.semantic.model` | string | `""` | `ANDROMEDA_INDEX_SEMANTIC_MODEL` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.semantic.batch_size` | integer | `64` | `ANDROMEDA_INDEX_SEMANTIC_BATCH_SIZE` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.watch.enabled` | boolean | `true` | `ANDROMEDA_INDEX_WATCH_ENABLED` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.watch.debounce_ms` | integer (ms) | `500` | `ANDROMEDA_INDEX_WATCH_DEBOUNCE_MS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.stale.max_pending_changes` | integer | `500` | `ANDROMEDA_INDEX_STALE_MAX_PENDING_CHANGES` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.stale.max_age_seconds` | integer (seconds) | `3600` | `ANDROMEDA_INDEX_STALE_MAX_AGE_SECONDS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.timeouts.build_seconds` | integer (seconds) | `1800` | `ANDROMEDA_INDEX_TIMEOUTS_BUILD_SECONDS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| `index.timeouts.update_seconds` | integer (seconds) | `300` | `ANDROMEDA_INDEX_TIMEOUTS_UPDATE_SECONDS` | Volume 7 | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |

## `[cli]` — owner: Volume 8

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `cli.color` | enum (`auto` \| `always` \| `never`) | `"auto"` | `ANDROMEDA_CLI_COLOR` | Volume 8 | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.pager` | enum (`auto` \| `always` \| `never`) | `"auto"` | `ANDROMEDA_CLI_PAGER` | Volume 8 | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.pager_command` | string | `""` (empty = use `PAGER`, else none) | `ANDROMEDA_CLI_PAGER_COMMAND` | Volume 8 | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.editor` | string | `""` (empty = use `VISUAL`/`EDITOR`) | `ANDROMEDA_CLI_EDITOR` | Volume 8 | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.update_notice` | boolean | `true` | `ANDROMEDA_CLI_UPDATE_NOTICE` | Volume 8 | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| `cli.default_timeout` | duration | `"0s"` (no deadline) | `ANDROMEDA_CLI_DEFAULT_TIMEOUT` | Volume 8 | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |

## `[tui]` — owner: Volume 8

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `tui.mouse` | boolean | `true` | `ANDROMEDA_TUI_MOUSE` | Volume 8 | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.splash` | enum (`auto` \| `always` \| `never`) | `"auto"` | `ANDROMEDA_TUI_SPLASH` | Volume 8 | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.default_screen` | enum (core-ring screen names) | `"session"` | `ANDROMEDA_TUI_DEFAULT_SCREEN` | Volume 8 | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| `tui.clipboard` | enum (`auto` \| `osc52` \| `off`) | `"auto"` | `ANDROMEDA_TUI_CLIPBOARD` | Volume 8 | [Vol 8 ch 11](../volume-08-cli-and-tui/11-interaction-patterns.md) |
| `tui.clipboard_max_bytes` | integer | `1048576` | `ANDROMEDA_TUI_CLIPBOARD_MAX_BYTES` | Volume 8 | [Vol 8 ch 11](../volume-08-cli-and-tui/11-interaction-patterns.md) |
| `tui.toast_duration_ms` | integer (ms) | `4000` | `ANDROMEDA_TUI_TOAST_DURATION_MS` | Volume 8 | [Vol 8 ch 11](../volume-08-cli-and-tui/11-interaction-patterns.md) |
| `tui.search_debounce_ms` | integer (ms) | `150` | `ANDROMEDA_TUI_SEARCH_DEBOUNCE_MS` | Volume 8 | [Vol 8 ch 11](../volume-08-cli-and-tui/11-interaction-patterns.md) |
| `tui.list_page_size` | integer | `100` | `ANDROMEDA_TUI_LIST_PAGE_SIZE` | Volume 8 | [Vol 8 ch 11](../volume-08-cli-and-tui/11-interaction-patterns.md) |
| `tui.glyphs` | enum (`auto` \| `unicode` \| `ascii`) | `"auto"` | `ANDROMEDA_TUI_GLYPHS` | Volume 8 | [Vol 8 ch 12](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) |
| `tui.reduce_motion` | boolean | `false` | `ANDROMEDA_TUI_REDUCE_MOTION` | Volume 8 | [Vol 8 ch 12](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) |
| `tui.accessible_output` | boolean | `false` | `ANDROMEDA_TUI_ACCESSIBLE_OUTPUT` | Volume 8 | [Vol 8 ch 12](../volume-08-cli-and-tui/12-accessibility-and-compatibility.md) |

`[tui.keymap]` is reserved for Beta key remapping ([ADR-108](adr/ADR-108.md)); its keys are
minted when that feature is specified, and E-TUI-005 already defines its validation failure
envelope.

## `[tui.theme]` — owner: Volume 8

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `tui.theme.mode` | enum (`auto` \| `dark` \| `light`) | `"auto"` (explicit value, else background detection, dark fallback) | `ANDROMEDA_TUI_THEME_MODE` | Volume 8 | [Vol 8 ch 08](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) |
| `tui.theme.tier` | enum (`auto` \| tier names incl. `ansi16`, `none`) | `"auto"` | `ANDROMEDA_TUI_THEME_TIER` | Volume 8 | [Vol 8 ch 08](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) |

## `[permissions]` — owner: Volume 9 (protected table — runtime overrides rejected, E-CFG-014)

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `permissions.approval_timeout` | duration | `"10m"` (`"0s"` disables expiry) | `ANDROMEDA_PERMISSIONS_APPROVAL_TIMEOUT` | Volume 9 | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `permissions.workspace_grant_ttl` | duration | `"0s"` (no auto-expiry) | `ANDROMEDA_PERMISSIONS_WORKSPACE_GRANT_TTL` | Volume 9 | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `permissions.rules` (array of tables) | array of tables (`name`, `permission`, `effect` `allow` \| `deny` \| `ask`, plus resource-qualifier keys per the selector grammar) | `[]` | — | Volume 9 | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |

The rule-field and selector vocabulary is cataloged in
[catalog-permissions.md](catalog-permissions.md).

## `[sandbox]` — owner: Volume 9 (protected table — runtime overrides rejected, E-CFG-014)

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `sandbox.isolation` | enum (`auto` \| `process` \| `os`) | `"auto"` | `ANDROMEDA_SANDBOX_ISOLATION` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.degradation` | enum (`refuse` \| `ask` \| `allow`) | `"ask"` | `ANDROMEDA_SANDBOX_DEGRADATION` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.env_allowlist` | array of strings | `["PATH", "HOME", "USER", "LOGNAME", "LANG", "LC_ALL", "LC_CTYPE", "TERM", "TMPDIR", "SHELL", "TZ"]` | `ANDROMEDA_SANDBOX_ENV_ALLOWLIST` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.writable_roots` | array of paths | `[]` | `ANDROMEDA_SANDBOX_WRITABLE_ROOTS` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.readonly_roots` | array of paths | `[]` | `ANDROMEDA_SANDBOX_READONLY_ROOTS` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.command_denylist` | array of strings (command patterns) | `["sudo *", "su *", "doas *", "shutdown *", "reboot *", "mkfs*", "dd *"]` | `ANDROMEDA_SANDBOX_COMMAND_DENYLIST` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.command_allowlist` | array of strings (command patterns) | `[]` (empty = not additionally restricted beyond the permission model) | `ANDROMEDA_SANDBOX_COMMAND_ALLOWLIST` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.max_cpu_seconds` | integer (seconds) | `300` | `ANDROMEDA_SANDBOX_MAX_CPU_SECONDS` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.max_memory_mb` | integer (MB) | `2048` | `ANDROMEDA_SANDBOX_MAX_MEMORY_MB` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.max_processes` | integer | `64` | `ANDROMEDA_SANDBOX_MAX_PROCESSES` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| `sandbox.max_open_files` | integer | `512` | `ANDROMEDA_SANDBOX_MAX_OPEN_FILES` | Volume 9 | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |

## `[security]` — owner: Volume 9 (protected table — runtime overrides rejected, E-CFG-014)

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `security.fallback_store` | boolean | `false` | `ANDROMEDA_SECURITY_FALLBACK_STORE` | Volume 9 | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| `security.redaction_patterns` | array of strings (patterns) | `[]` (additive-only heuristics) | `ANDROMEDA_SECURITY_REDACTION_PATTERNS` | Volume 9 | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| `security.audit_retention` | duration | `"400d"` | `ANDROMEDA_SECURITY_AUDIT_RETENTION` | Volume 9 | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |
| `security.audit_verify_on_open` | enum (`head` \| full-verification values per chapter) | `"head"` | `ANDROMEDA_SECURITY_AUDIT_VERIFY_ON_OPEN` | Volume 9 | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |

## `[logging]` — owner: Volume 10

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `logging.level` | enum (`debug` \| `info` \| `warn` \| `error`) | `"info"` | `ANDROMEDA_LOGGING_LEVEL` | Volume 10 | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `logging.stderr_level` | enum (same set) | `"warn"` | `ANDROMEDA_LOGGING_STDERR_LEVEL` | Volume 10 | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `logging.include_source` | boolean | `false` | `ANDROMEDA_LOGGING_INCLUDE_SOURCE` | Volume 10 | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `logging.file.enabled` | boolean | `true` | `ANDROMEDA_LOGGING_FILE_ENABLED` | Volume 10 | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `logging.file.max_size_mb` | integer (MB) | `32` | `ANDROMEDA_LOGGING_FILE_MAX_SIZE_MB` | Volume 10 | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `logging.file.max_files` | integer | `10` | `ANDROMEDA_LOGGING_FILE_MAX_FILES` | Volume 10 | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| `logging.file.max_age_days` | integer | `30` | `ANDROMEDA_LOGGING_FILE_MAX_AGE_DAYS` | Volume 10 | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |

## `[telemetry]` — owner: Volume 10 (protected table — runtime overrides rejected, E-CFG-014)

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `telemetry.enabled` | boolean | `false` (remote export also requires recorded consent) | `ANDROMEDA_TELEMETRY_ENABLED` | Volume 10 | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.endpoint` | string (URL) | `""` | `ANDROMEDA_TELEMETRY_ENDPOINT` | Volume 10 | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.protocol` | enum (`http/protobuf` \| `grpc`) | `"http/protobuf"` | `ANDROMEDA_TELEMETRY_PROTOCOL` | Volume 10 | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.auth_secret_ref` | string (secret reference) | `""` | `ANDROMEDA_TELEMETRY_AUTH_SECRET_REF` | Volume 10 | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.export_interval` | duration | `"60s"` | `ANDROMEDA_TELEMETRY_EXPORT_INTERVAL` | Volume 10 | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.queue_max_size_mb` | integer (MB) | `64` | `ANDROMEDA_TELEMETRY_QUEUE_MAX_SIZE_MB` | Volume 10 | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| `telemetry.locked` | boolean (system layer) | `false` | `ANDROMEDA_TELEMETRY_LOCKED` | Volume 10 | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |

## `[storage]` — owner: Volume 10

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `storage.lock_wait_ms` | integer (ms) | `5000` | `ANDROMEDA_STORAGE_LOCK_WAIT_MS` | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.backups.dir` | path | `""` (ADR-022 defaults per database) | `ANDROMEDA_STORAGE_BACKUPS_DIR` | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.backups.retain_count` | integer | `3` | `ANDROMEDA_STORAGE_BACKUPS_RETAIN_COUNT` | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.retention.sessions_days` | integer | `0` (keep forever) | `ANDROMEDA_STORAGE_RETENTION_SESSIONS_DAYS` | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.retention.runs_days` | integer | `0` (keep forever) | `ANDROMEDA_STORAGE_RETENTION_RUNS_DAYS` | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.retention.artifacts_days` | integer | `0` (keep forever) | `ANDROMEDA_STORAGE_RETENTION_ARTIFACTS_DAYS` | Volume 10 | [Vol 10 ch 01](../volume-10-config-storage-observability/01-configuration-model.md) |
| `storage.events.retention_days` | integer (≥ 1) | `90` | `ANDROMEDA_STORAGE_EVENTS_RETENTION_DAYS` | Volume 10 | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| `storage.events.max_size_mb` | integer (MB) | `512` | `ANDROMEDA_STORAGE_EVENTS_MAX_SIZE_MB` | Volume 10 | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| `storage.traces.retention_days` | integer (≥ 1) | `30` | `ANDROMEDA_STORAGE_TRACES_RETENTION_DAYS` | Volume 10 | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `storage.traces.max_size_mb` | integer (MB) | `512` | `ANDROMEDA_STORAGE_TRACES_MAX_SIZE_MB` | Volume 10 | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `storage.metrics.local_persistence` | boolean | `true` | `ANDROMEDA_STORAGE_METRICS_LOCAL_PERSISTENCE` | Volume 10 | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `storage.metrics.retention_days` | integer (≥ 1) | `30` | `ANDROMEDA_STORAGE_METRICS_RETENTION_DAYS` | Volume 10 | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| `storage.cost_records.retention_days` | integer (≥ 1) | `365` | `ANDROMEDA_STORAGE_COST_RECORDS_RETENTION_DAYS` | Volume 10 | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |

## `[git]` — owner: Volume 11

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `git.binary_path` | path | `""` (resolve from `PATH`) | `ANDROMEDA_GIT_BINARY_PATH` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.operation_timeout_seconds` | integer (seconds) | `120` | `ANDROMEDA_GIT_OPERATION_TIMEOUT_SECONDS` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.remote_timeout_seconds` | integer (seconds) | `300` | `ANDROMEDA_GIT_REMOTE_TIMEOUT_SECONDS` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.allow_force_push` | boolean | `false` | `ANDROMEDA_GIT_ALLOW_FORCE_PUSH` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.protected_branches` | array of strings (glob patterns) | `["main", "master", "release/*"]` | `ANDROMEDA_GIT_PROTECTED_BRANCHES` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.sign_commits` | enum (`auto` \| `always` \| `never`) | `"auto"` | `ANDROMEDA_GIT_SIGN_COMMITS` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.hooks.run` | boolean | `true` | `ANDROMEDA_GIT_HOOKS_RUN` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.safety_refs.enabled` | boolean | `true` | `ANDROMEDA_GIT_SAFETY_REFS_ENABLED` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.safety_refs.retention_days` | integer | `30` | `ANDROMEDA_GIT_SAFETY_REFS_RETENTION_DAYS` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.submodules.recurse` | boolean | `false` | `ANDROMEDA_GIT_SUBMODULES_RECURSE` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.blame.ignore_revs` | boolean | `true` | `ANDROMEDA_GIT_BLAME_IGNORE_REVS` | Volume 11 | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| `git.hosting.gitlab.enabled` | boolean | `false` | `ANDROMEDA_GIT_HOSTING_GITLAB_ENABLED` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `git.hosting.gitlab.api_base_url` | string (URL) | `"https://gitlab.com/api/v4"` | `ANDROMEDA_GIT_HOSTING_GITLAB_API_BASE_URL` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `git.hosting.gitlab.default_remote` | string | `"origin"` | `ANDROMEDA_GIT_HOSTING_GITLAB_DEFAULT_REMOTE` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `git.hosting.gitlab.auth_profile` | string (AuthPort profile name — never a literal token) | `""` | `ANDROMEDA_GIT_HOSTING_GITLAB_AUTH_PROFILE` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `git.hosting.gitlab.draft_by_default` | boolean | `true` | `ANDROMEDA_GIT_HOSTING_GITLAB_DRAFT_BY_DEFAULT` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |

## `[github]` — owner: Volume 11

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `github.enabled` | boolean | `false` | `ANDROMEDA_GITHUB_ENABLED` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `github.api_base_url` | string (URL) | `"https://api.github.com"` | `ANDROMEDA_GITHUB_API_BASE_URL` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `github.default_remote` | string | `"origin"` | `ANDROMEDA_GITHUB_DEFAULT_REMOTE` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `github.auth_profile` | string (AuthPort profile name — never a literal token) | `""` | `ANDROMEDA_GITHUB_AUTH_PROFILE` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| `github.draft_by_default` | boolean | `true` | `ANDROMEDA_GITHUB_DRAFT_BY_DEFAULT` | Volume 11 | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |

## `[update]` — owner: Volume 14

| Key | Type | Default | Environment variable | Owner | Defined in |
|---|---|---|---|---|---|
| `update.channel` | enum (`stable` \| `rc` \| `beta` \| `nightly`) | `"stable"` | `ANDROMEDA_UPDATE_CHANNEL` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.source` | string (`"github"` or a mirror root path/URL) | `"github"` | `ANDROMEDA_UPDATE_SOURCE` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.auto_check` | boolean | `true` | `ANDROMEDA_UPDATE_AUTO_CHECK` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.check_interval_hours` | integer (hours) | `24` | `ANDROMEDA_UPDATE_CHECK_INTERVAL_HOURS` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.auto_download` | boolean | `false` | `ANDROMEDA_UPDATE_AUTO_DOWNLOAD` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.auto_apply` | boolean | `false` (never across majors, ADR-191) | `ANDROMEDA_UPDATE_AUTO_APPLY` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.notify` | boolean | `true` | `ANDROMEDA_UPDATE_NOTIFY` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.keep_versions` | integer | `1` | `ANDROMEDA_UPDATE_KEEP_VERSIONS` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.signature_policy` | enum (`when_present` \| `required`) | `"when_present"` | `ANDROMEDA_UPDATE_SIGNATURE_POLICY` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.timeouts.check_seconds` | integer (seconds) | `30` | `ANDROMEDA_UPDATE_TIMEOUTS_CHECK_SECONDS` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.timeouts.download_seconds` | integer (seconds) | `600` | `ANDROMEDA_UPDATE_TIMEOUTS_DOWNLOAD_SECONDS` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| `update.timeouts.apply_seconds` | integer (seconds) | `60` | `ANDROMEDA_UPDATE_TIMEOUTS_APPLY_SECONDS` | Volume 14 | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |

## Consolidation notes

- **Coverage.** 283 rows across the reserved root keys and the 24 cataloged tables and
  sub-tables round-trip against the volume registers' "Config keys minted" sections: every
  register-listed key appears above, and no key appears that a register does not list.
  Volumes 1–3, 12, 13, and 15 mint no configuration keys (their registers say so
  explicitly; Volume 12's operational limits become keys only in the owning areas' tables).
- **Illustrative rows in the Volume 10 complete example.** Volume 10 chapter 01's
  end-to-end `andromeda.toml` example marks its `[tui]`, `[permissions]`, `[sandbox]`,
  `[security]`, `[git]`, `[github]`, and `[update]` entries "illustrative", and several of
  those placeholder keys (`permissions.default_decision`, `sandbox.level`,
  `security.workspace_trust_default`, `git.executable`, `tui.theme.mode = "dark"`) differ
  from the keys the owning chapters later minted. Per that chapter's own boundary rule
  ("the owning volume's key catalog is normative and the example is illustrative for those
  tables"), this catalog follows the owning chapters; no register conflicts with its
  defining chapter.
- **`[git.hosting.gitlab]`.** The Volume 0 chapter 03 ownership map names `[git]` and
  `[github]` for Volume 11; chapter 02 nests additional hosting providers under
  `[git.hosting.*]` (the primary integration keeps the short `[github]` table). The
  sub-table is cataloged under `[git]` accordingly.
- **`expose_tools`.** Chapter 05's `[mcp.servers.<name>]` key list omits `expose_tools`;
  the Volume 6 register attributes it to chapter 06 (least-exposure allowlist), which this
  catalog links as its defining chapter. Register and chapters agree.
- **`api_key_env` and `auth_profile`.** The Volume 5 register's provider-table row lists
  the core `[providers.<slug>]` keys without these two; its authentication section lists
  both as minted inside `[providers.<slug>]` by chapter 08 (FR-AUTH-002 indirection and
  per-provider profile binding). Both sections of the same register; no conflict.
- **File name.** The annex index ([00-index.md](00-index.md)) refers to this catalog as
  `catalog-configuration.md`; the file ships as `catalog-config.md`. The index row is
  descriptive (not a link); reconciliation of the name is deferred to the Volume 0 change
  procedure.
