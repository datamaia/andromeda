# volume-05-providers-and-auth — Volume Register

Merged from per-agent register fragments at the Phase B gate.

## Requirements index

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-PROV-001 | Provider contract | Core | Provider conformance suite; ADR-033 dependency checks; contract tests incl. cancellation |
| FR-PROV-002 | Adapter declaration and registration | Core | Registry validation tests; declaration-vs-behavior conformance checks |
| FR-PROV-010 | Capability declaration and the capability matrix | Core | Conformance honesty checks (SM-04); ADR-033 name-branch scan; resolution unit tests |
| FR-PROV-011 | Capability negotiation, verification, and degradation | MVP | Conformance degradation scenarios; refutation fault injection |
| FR-PROV-012 | Model discovery and the model registry | MVP | Reconciliation unit tests; recorded-fixture discovery tests; offline suite |
| FR-PROV-013 | Provider and model change notification | MVP | Event-before-output integration tests; SM-13 audit-chain test |
| FR-PROV-020 | Streaming contract | MVP | Paced-fixture conformance; fault injection; SM-08 benchmarks |
| FR-PROV-021 | Tool-calling normalization | MVP | Round-trip corpus per adapter; malformed-output fault injection |
| FR-PROV-022 | Structured outputs | MVP | Schema corpus across modes; violation fixtures; retry-ceiling tests |
| FR-PROV-030 | Token usage accounting | MVP | Accounting-honesty conformance; SM-13 chain; CountTokens contract tests |
| FR-PROV-031 | Cost accounting and pricing tables | MVP | Resolution/arithmetic unit tests; config validation tests; labeling tests |
| FR-PROV-040 | Timeouts, rate limits, and retries | MVP | Fault injection with fake clocks; abort-on-timeout contract tests |
| FR-PROV-041 | Circuit breaker and health verification | MVP | Scripted failure sequences; single-probe concurrency tests |
| FR-PROV-042 | Routing and selection | MVP | Selection unit tests; SM-12 replay determinism; permission-path tests |
| FR-PROV-043 | Fallback and its guard rules | MVP | Per-guard integration scenarios; headless policy tests; SM-13 verification |
| FR-PROV-050 | Provider error normalization | Core | Fault-injection corpus; leak checks; envelope completeness check |
| NFR-PROV-001 | Provider integration effort | Beta | Timed reference-integration exercise at phase gates (SM-01) |
| NFR-PROV-002 | Local-model conformance | v1 | Conformance suite per release on ≥ 2 local serving paths (SM-04) |
| NFR-PROV-003 | Accounting completeness | MVP | Record-completeness validator over suite runs |
| NFR-PROV-004 | Error normalization coverage | MVP | Leak-detection assertions; per-code fault reconciliation |
| RISK-PROV-001 | Provider API drift breaks adapters | — | Risk register review at phase gates; live conformance runs |
| RISK-PROV-002 | Capability misdeclaration | — | Risk register review at phase gates; honesty checks |
| RISK-PROV-003 | Stale or wrong pricing data misleads users | — | Risk register review at phase gates; basis metrics |
| RISK-PROV-004 | Fallback amplifies cost or data exposure | — | Risk register review at phase gates; SM-13 audit trail |

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-AUTH-001 | Official-mechanisms-only authentication | Core | Registry validation tests; conformance prohibited-mechanism cases; adapter review checklist; CI static checks |
| FR-AUTH-002 | API key authentication | MVP | Intake-path integration tests; canary leak scan; CI env-indirection test; permission tests |
| FR-AUTH-003 | OAuth 2.0 authorization code flow with PKCE | Beta | Mock authorization-server integration tests; grant-type static check; leak scan |
| FR-AUTH-004 | OAuth 2.0 device authorization grant | Beta | Mock device-flow integration tests; SSH manual test at Beta gate; leak scan |
| FR-AUTH-005 | Service accounts and managed identity | v1 | Mock token-exchange tests; platform-conditional fakes; per-provider validation record check |
| FR-AUTH-006 | Enterprise proxies and trust anchors | Beta | Local authenticating-proxy integration tests; no-bypass static check |
| FR-AUTH-007 | Temporary credentials | Beta | Fake-clock unit tests; expiry sweep test |
| FR-AUTH-008 | Multiple authentication profiles | MVP | Precedence-matrix unit tests; ambiguity cases; SM-12 record check |
| FR-AUTH-009 | Credential storage and resolution through the Secret Store | MVP | SecretStorePort ordering contract tests with crash injection; canary scans; audit completeness check |
| FR-AUTH-010 | Token refresh | Beta | Mock token-endpoint tests (single-flight, rotation-on-use, offline); fake-clock tests; leak scan |
| FR-AUTH-011 | Credential rotation and revocation | MVP | Atomicity and cascade integration tests; idempotence tests; endpoint failure injection; audit chain |
| FR-PROV-080 | Adapter catalog and phase classification | MVP | Adapter conformance suites; registry and validation tests; release audit against catalog |
| FR-PROV-081 | Generic OpenAI-compatible adapter | MVP | Conformance suite (fixtures + live local server); framing fault injection; offline suite |
| FR-PROV-082 | Anthropic adapter | MVP | Conformance suite with recorded fixtures; scheduled live smoke; vendor-type leak check |
| FR-PROV-083 | Ollama adapter | MVP | Conformance suite against pinned local Ollama; offline suite; dependency audit |
| FR-PROV-084 | Local provider operation | MVP | Offline suite (UC-09 journey); socket-activity assertion; locality unit tests; permission parity test |
| FR-PROV-085 | Offline behavior of the provider layer | MVP | Offline suite per SM-05 method; fault injection (portals, partial connectivity); retry-count assertions |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-AUTH-001 | No plaintext credential material at rest | MVP | Canary-secret scan over all written artifacts, every gated CI run |
| NFR-AUTH-002 | Credential redaction in logs, events, errors, and memory records | MVP | Canary scan over observability channels with debug verbosity, every gated CI run |
| NFR-AUTH-003 | Credential resolution latency | Beta | Per-backend benchmark harness, p95, per release |

SM-04 is formalized by fragment A as NFR-PROV-002 (chapter 02); chapter 10 supplies its
serving paths and offline condition and mints no second formalization.

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-AUTH-001 | Secret Store backend unavailability and fallback overuse | Medium | Open |
| RISK-AUTH-002 | Credential leakage through logs, errors, or crash artifacts | High | Open |
| RISK-AUTH-003 | Provider authentication mechanism drift | Medium | Open |
| RISK-PROV-080 | Adapter catalog drift against provider realities | High | Open |
| RISK-PROV-081 | Local model and server capability variance | High | Open |

## ADRs minted

| ADR | Title | One-line decision |
|---|---|---|
| ADR-055 | Adapter Declaration manifest | Static, versioned, registration-validated per-adapter declaration set; invalid declarations refused (E-PROV-019) |
| ADR-056 | Capability enum extension and provenance resolution | Add `token_counting`; resolve effective sets from declared/discovered/configured/verified provenance with refutation masking |
| ADR-057 | Unified stream event model | Adapters normalize all wire streams to the four-member `ChatEvent` union; closed finish-reason set; no whole-response buffering |
| ADR-058 | No shipped price data | Cost estimates only from user-maintained pricing tables with `source`/`effective_date`; basis honesty everywhere |
| ADR-059 | Class-keyed retries and per-provider circuit breaker | Router-owned jittered retries and breaker keyed to normalized error classes; no hedging; SDK-internal retries disabled |
| ADR-060 | Explicit-first routing, guarded fallback chains | MVP strategies `explicit`/`preference_list`; fallback only via configured chains under ordered guards F1–F8 |
| ADR-061 | Two-layer error normalization | Adapter wire-mapping + router enrichment; raw identity preserved redacted; retryability from normalized codes only |

Block 055–069 belongs to Volume 5; this fragment used 062–067 (068–069 remain permanent
gaps unless fragment A's orchestrated merge records otherwise; fragment A uses 055–061).

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-062](../annexes/adr/ADR-062.md) | Named authentication profiles bind credentials to providers | Accepted | `[auth.profiles.<name>]` units with strict selection precedence; ambiguity is an error; profile recorded per run |
| [ADR-063](../annexes/adr/ADR-063.md) | OAuth for native clients: PKCE + loopback redirect, plus device grant | Accepted | Exactly two flows (auth code with PKCE over 127.0.0.1; device authorization grant); implicit and password grants prohibited |
| [ADR-064](../annexes/adr/ADR-064.md) | Proactive single-flight token refresh with on-demand fallback | Accepted | Use-triggered lead-time refresh + provider-signal refresh; single-flight per session; transient/definitive failure split |
| [ADR-065](../annexes/adr/ADR-065.md) | Adapter catalog phasing: generic-adapter-first | Accepted | OpenAI-compatible services reachable from MVP via one adapter; dedicated adapters only for declared value, phased per catalog |
| [ADR-066](../annexes/adr/ADR-066.md) | Offline behavior: passive detection, fail-fast, resolved-address locality | Accepted | No probing ever; connectivity failures exempt from retry and mark providers unavailable; local = loopback/UDS by resolved address |
| [ADR-067](../annexes/adr/ADR-067.md) | Enterprise proxies via standard env vars and configured trust anchors | Accepted | Proxy env honored, `[auth.proxy]` overrides, PEM `ca_bundle`; no TLS verification bypass exists for non-loopback endpoints |

## Error codes minted

E-PROV-001 (provider unreachable), E-PROV-002 (authentication rejected), E-PROV-003 (rate
limited), E-PROV-004 (quota/billing exhausted), E-PROV-005 (model not available),
E-PROV-006 (capability unavailable), E-PROV-007 (request rejected as invalid), E-PROV-008
(response malformed), E-PROV-009 (stream interrupted), E-PROV-010 (request timeout),
E-PROV-011 (request cancelled), E-PROV-012 (provider internal error), E-PROV-013 (content
refused by provider policy), E-PROV-014 (context window exceeded), E-PROV-015 (circuit
breaker open), E-PROV-016 (no eligible provider or model), E-PROV-017 (structured output
validation failed), E-PROV-018 (tool call unparseable), E-PROV-019 (adapter declaration
invalid). Full ADR-016 envelopes in chapter 06.
| Code | Name | Exit code |
|---|---|---|
| E-AUTH-001 | Credential not found | 4 |
| E-AUTH-002 | Authentication rejected by provider | 4 |
| E-AUTH-003 | Credential expired and renewal failed | 4 |
| E-AUTH-004 | Authorization flow denied | 4 |
| E-AUTH-005 | Authorization flow timed out | 8 |
| E-AUTH-006 | Secret store unavailable | 3 |
| E-AUTH-007 | Prohibited authentication mechanism requested | 3 |
| E-AUTH-008 | Credential rotation failed | 4 |
| E-AUTH-009 | Provider-side revocation failed | 4 |
| E-AUTH-010 | Authentication profile not found or ambiguous | 3 |
| E-AUTH-011 | Proxy authentication failed | 4 |

## Events minted

| Event | Defined in |
|---|---|
| `provider.request.completed`, `provider.request.failed` | chapter 01 |
| `provider.discovery.completed`, `provider.model.deprecated`, `provider.capability.changed`, `provider.capability.verified`, `provider.degradation.applied` | chapter 02 |
| `provider.stream.interrupted` | chapter 03 |
| `provider.cost.recorded` | chapter 04 |
| `provider.request.retried`, `provider.route.selected`, `provider.fallback.activated`, `provider.breaker.opened`, `provider.breaker.closed` | chapter 05 |
| `provider.adapter.registered`, `provider.adapter.rejected` | chapter 06 |

Envelope, delivery, persistence, and retention per Volume 10; payload summaries in the
minting chapters.

Per the Volume 0 grammar; envelope, delivery, persistence, and retention per Volume 10.

| Event | Emitted by | Meaning |
|---|---|---|
| `auth.credential.created` | Authentication Layer | Intake stored a new Credential |
| `auth.credential.rotated` | Authentication Layer | Rotation completed with successor linkage |
| `auth.credential.revoked` | Authentication Layer | Revocation completed (payload marks provider-side outcome) |
| `auth.credential.expired` | Authentication Layer | Temporary credential reached expiry |
| `auth.credential.deleted` | Authentication Layer | Slot and row deletion completed |
| `auth.credential.access_failed` | Authentication Layer | Secret Store resolution failed (E-AUTH-006) |
| `auth.credential.rotation_failed` | Authentication Layer | Rotation aborted (E-AUTH-008) |
| `auth.mechanism.refused` | Authentication Layer | FR-AUTH-001 gate refused a mechanism (E-AUTH-007) |
| `auth.profile.selected` | Authentication Layer | Profile resolution succeeded at establishment |
| `auth.profile.resolution_failed` | Authentication Layer | Profile resolution failed (E-AUTH-010) |
| `auth.session.established` | Authentication Layer | Authentication Session reached `active` |
| `auth.session.refreshed` | Authentication Layer | Renewal replaced token material |
| `auth.session.expired` | Authentication Layer | Session reached `expired` |
| `auth.session.failed` | Authentication Layer | Establishment or renewal failed; also cancellation class |
| `auth.session.revoked` | Authentication Layer | Session terminated by credential revocation/rotation |
| `provider.connection.verified` | Provider Layer | Verification succeeded; provider `available` |
| `provider.connection.degraded` | Provider Layer | Degradation thresholds crossed |
| `provider.connection.recovered` | Provider Layer | Returned to `available` after degradation |
| `provider.connection.lost` | Provider Layer | Provider became `unavailable` |
| `provider.connection.disabled` | Provider Layer | Administrative disable |
| `provider.connection.enabled` | Provider Layer | Administrative enable (verification follows) |
| `provider.connection.removed` | Provider Layer | Deregistration tombstone |
| `provider.deprecation.announced` | Provider Layer | Adapter detected a provider-documented deprecation |

## Config keys minted

| Table | Keys |
|---|---|
| `[providers]` | `default`, `discovery_ttl_hours` |
| `[providers.<slug>]` | `adapter`, `enabled`, `base_url`, `default_model`, `auth_profile`, `verify_capabilities`, `reverify_s` |
| `[providers.<slug>.capability_overrides."<pattern>"]` | `add`, `remove` |
| `[providers.<slug>.structured_outputs]` | `allow_prompted`, `validation_retries` |
| `[providers.<slug>.pricing."<model>"]` | `input_per_million_micros`, `output_per_million_micros`, `cached_input_per_million_micros`, `reasoning_per_million_micros`, `currency`, `source`, `effective_date` |
| `[providers.<slug>.limits]` | `max_concurrent_requests`, `requests_per_minute` |
| `[providers.<slug>.timeouts]` | `connect_ms`, `request_ms`, `first_token_ms`, `stream_idle_ms`, `stream_total_ms`, `discovery_ms`, `embed_ms` |
| `[providers.<slug>.retry]` | `max_attempts`, `base_delay_ms`, `backoff_multiplier`, `max_delay_ms`, `retry_after_cap_ms` |
| `[providers.<slug>.breaker]` | `enabled`, `failure_threshold`, `failure_ratio`, `min_samples`, `window_s`, `open_base_s`, `open_max_s` |
| `[providers.routing]` | `strategy`, `preference` |
| `[[providers.fallback.chains]]` | `name`, `from`, `targets`, `triggers`, `allow_local_to_cloud`, `max_price_multiplier`, `require_approval` |

Schema, precedence, and validation per Volume 10 (`[auth]` keys are fragment B's).

Content owner: Volume 5 (`[auth]`, `[providers.*]` per Volume 0 chapter 03); schema,
precedence, and validation: Volume 10.

| Key | Table | Meaning |
|---|---|---|
| `default_profile` | `[auth]` | Profile used when no selector applies |
| `refresh_lead_time_seconds` | `[auth]` | Proactive refresh window and expiry margin (default 300) |
| `flow_timeout_seconds` | `[auth]` | Browser/device flow lifetime ceiling (default 300) |
| `url` | `[auth.proxy]` | Explicit proxy override (empty honors env) |
| `no_proxy` | `[auth.proxy]` | Explicit bypass list override (empty honors env) |
| `credential` | `[auth.proxy]` | Credential label for authenticating proxies |
| `ca_bundle` | `[auth.proxy]` | PEM file appended to system trust anchors |
| `provider` | `[auth.profiles.<name>]` | Provider slug bound by the profile |
| `credential` | `[auth.profiles.<name>]` | Credential label bound by the profile |
| `auth_profile` | `[providers.<slug>]` | Per-provider profile binding override |
| `api_key_env` | `[providers.<slug>]` | Environment variable indirection for API keys (FR-AUTH-002) |

## Glossary additions

| Term | One-line meaning |
|---|---|
| Adapter Declaration | The static, versioned, registration-validated manifest of every provider fact an adapter claims (Volume 5, chapter 01). |
| Provider Router | The Provider Layer composite that implements ProviderPort, owning routing, fallback, retries, pacing, and breakers (Volume 5, chapters 01/05). |
| Effective capability set | The per-model capability set resolved from declared/discovered/configured provenance with refutation masking (Volume 5, chapter 02). |
| Capability provenance | The class recording why a capability value is present: `declared`, `discovered`, `configured`, `verified`, or masked by `refuted` (Volume 5, chapter 02). |
| Degradation strategy | The per-capability rule applied when a required capability is absent: `refuse`, `report_unavailable`, `substitute` (opt-in), or `reroute` (Volume 5, chapter 02). |
| Usage report | The per-request token/cost accounting structure populated exclusively from official provider accounting (Volume 5, chapter 04). |
| Pricing table | User-maintained per-(provider, model) price configuration with mandatory `source` and `effective_date`, the only source of cost estimates (Volume 5, chapter 04; ADR-058). |
| Fallback chain | A configured, ordered set of fallback targets with trigger classes and guard parameters, the only mechanism by which fallback occurs (Volume 5, chapter 05). |
| Guard rules F1–F8 | The ordered normative conditions every fallback step must pass: explicit chains, egress, capability, policy, cost, stream boundary, no auth masking, announcement (Volume 5, chapter 05). |
| Authentication Profile | A named configuration unit binding a provider slug to a credential label with selection options (ADR-062). |
| Ephemeral credential | Environment-sourced credential material used for the process lifetime and never persisted (FR-AUTH-002). |
| Locality rule | The ADR-066 classification: an endpoint is local exactly when it resolves to loopback or a Unix domain socket. |
| Generic-adapter-first | The ADR-065 rule that OpenAI-compatible services are reachable via the generic adapter before any dedicated adapter ships. |
| Single-flight refresh | The FR-AUTH-010 discipline: at most one renewal exchange in flight per Authentication Session, all consumers awaiting its outcome. |

## Assumptions

Local list per Volume 0, chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | Seed providers' documented error-body schemas are stable enough for table-driven `ErrorMapping` rules | E-PROV-008 rates versus fault-injection baselines per release | Mapping moves from tables to per-adapter code paths; NFR-PROV-004 method unchanged |
| Technical assumption | All seed providers' official streaming mechanisms normalize to the four-member `ChatEvent` union without losing consumer-relevant information | Conformance suite round-trip checks per adapter (SM-04, SM-08) | Additive union member proposed per ADR-057 review conditions |
| Product hypothesis | Cost-sensitive users will maintain pricing tables given honest `unavailable` defaults | Basis-distribution telemetry from consenting installs; feedback channels | Prioritize an official-source pricing importer per ADR-058 reversal plan |

Local list per Volume 0 chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | The OpenAI-compatible surface defined in FR-PROV-081 (chat completions + SSE, optional models/embeddings endpoints) is a sufficient common denominator for the catalog services classified to the generic path | Conformance suite runs against each service as adapters are implemented (V5B-OQ-1) | Narrow the generic surface or promote affected services' dedicated adapters |
| Technical assumption | Provider expired-credential signals are distinguishable from other authentication rejections via documented error responses (needed for FR-AUTH-010 on-demand refresh) | Per-adapter contract tests at implementation | On-demand refresh degrades to re-authentication prompts for affected providers |
| Technical assumption | Loopback listeners on ephemeral 127.0.0.1 ports are available in target desktop environments for FR-AUTH-003 | Beta-phase testing on Tier 1 platforms | Device authorization grant (FR-AUTH-004) is the fallback flow |
| Product hypothesis | Three MVP adapters (generic, Anthropic, Ollama) cover the practical provider needs of MVP users given the generic path's reach | MVP feedback channels (Volume 15) | Promote Beta-phase adapters earlier via change procedure |

## Open questions

Entries for every PENDING VALIDATION used in chapters 01–06 (Volume 0, chapter 08 format;
global `OQ-NNN` minted at consolidation).

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V5A-OQ-1 | Per-adapter official SDK adoption (ADR-019: PENDING VALIDATION at each adapter's implementation) | Chapter 01 (contract constraints) | No — stdlib baseline defined | Evaluation checklist outcome recorded per adapter in the decision register at implementation time | Open |
| V5A-OQ-2 | Which providers document native structured-output enforcement, per model family (PENDING VALIDATION per adapter) | Chapter 03 (modes table) | No — negotiation degrades honestly | Adapter catalog (chapter 09) records the per-provider outcome from official documentation | Open |
| V5A-OQ-3 | Which usage fields (cached/reasoning tokens, monetary cost) each provider officially reports (PENDING VALIDATION per adapter) | Chapter 04 (usage report) | No — absent fields stay absent with honest basis | Adapter catalog records per-provider `UsageReporting`/`CostReporting` facts from official documentation | Open |
| V5A-OQ-4 | Which rate-limit response metadata each provider documents for delay signaling (PENDING VALIDATION per adapter) | Chapter 05 (rate limits) | No — capped backoff applies absent signals | Adapter catalog records documented signal formats per provider | Open |

Every PENDING VALIDATION occurrence in chapters 07–11 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V5B-OQ-1 | Per-adapter capability, endpoint, auth-detail, and pricing facts for all non-MVP catalog adapters (including LM Studio, vLLM, LiteLLM, llama.cpp, LocalAI, FastChat, Text Generation WebUI specifics) — PENDING VALIDATION | Chapter 09 catalog and per-adapter notes | No — generic adapter covers reachability; markers resolved per adapter | Verify against official documentation at each adapter's implementation; record outcomes in the decision register | Open |
| V5B-OQ-2 | Which providers officially document OAuth (authorization code and/or device grant) for third-party clients — PENDING VALIDATION per provider | Chapter 07 (FR-AUTH-003/004); ADR-063 | No — API keys are the MVP mechanism | Per-provider documentation verification at adapter implementation | Open |
| V5B-OQ-3 | Service-account and managed-identity concrete flows (Google/Vertex surface, Azure Entra ID tokens) — PENDING VALIDATION | Chapter 07 (FR-AUTH-005); chapter 09 (`gemini`, `azure_openai`) | No — v1 phase; abstraction fixed | Validation spike per provider before its v1 adapter ships | Open |
| V5B-OQ-4 | Account/subscription-based authentication: no provider currently validated as offering an official, documented third-party mechanism; the FR-AUTH-001 gate stays closed per provider — PENDING VALIDATION | Chapter 07 method-family table; FR-AUTH-001 | No — gate closed is the safe default | Re-verify per provider at each phase gate; open the gate only via change procedure with a recorded validation | Open |
| V5B-OQ-5 | Official SDK adoption per adapter (openai-go, anthropic-sdk-go) remains PENDING VALIDATION per ADR-019 (referenced by chapters 09 FR-PROV-082 and per-adapter notes) | Chapter 09 | No — stdlib HTTP is the baseline | Tracked by ADR-019 review conditions; resolved at each adapter's implementation | Open |
| V5B-OQ-6 | Provider-side credential revocation API availability per provider — PENDING VALIDATION | Chapter 08 (FR-AUTH-011) | No — local invalidation always applies | Per-provider documentation verification at adapter implementation | Open |

## Cross-volume references

- Volume 2, chapter 05 (Provider/Model/Capability/Credential entities; INV-PRV, INV-MDL,
  INV-CAP, INV-CRED, INV-AUTHS invariants) and chapter 08 (Cost Record, INV-COST).
- Volume 3, chapter 02 (frozen ProviderPort/AuthPort signatures; FR-ARCH-003/FR-ARCH-004)
  and chapter 04 (Provider Layer / Authentication Layer component tables).
- Volume 4: turn-level retry/repair decisions, run budgets, replay (referenced by name).
- Volume 6: tool contract and JSON Schema dialect (ADR-024) transported by chapter 03.
- Volume 7: token estimation fallback on E-PROV-006 from `CountTokens`.
- Volume 8: presentation of announcements, cost labeling, provider views, exit codes.
- Volume 9: permission enum (`network`, scope `provider`), Approval records, redaction.
- Volume 10: configuration schema/precedence, event envelope, storage, rollups.
- Volume 12: latency/overhead budgets (SM-06..SM-09 formalization) consumed by chapters
  01/03/05.
- Volume 13: provider conformance suite, fault-injection corpus, offline suite.
- Fragment B (chapters 07–11): authentication flows, credential lifecycle, adapter catalog,
  offline mapping, Authentication Session and Provider connection machines.

- **Volume 3**: AuthPort and SecretStorePort frozen signatures (chapter 02) elaborated by
  chapters 07–08; ProviderPort elaborated by chapters 09–10; E-PORT-003 surfaced beneath
  E-AUTH-006.
- **Volume 2**: Credential and Authentication Session entities and invariants (INV-CRED-01..04,
  INV-AUTHS-01..04); frozen state enums (chapter 09) used exactly by chapter 11; recorded
  Credential status vocabulary.
- **Volume 9**: credential storage model (keystone FR-SEC-102, ADR-014); permission names
  `credential_access`, `network`; Audit Log semantics; redaction display rules.
- **Volume 10**: event envelope and delivery semantics for all events minted here;
  configuration schema, precedence, and validation for the keys minted here; logging
  redaction mechanics behind NFR-AUTH-002.
- **Volume 8**: `auth` and `provider` CLI command families (grammar and exit-code surfaces);
  TUI presentation of provider status and auth flows.
- **Volume 12**: SM-05 offline-operation NFR formalization (FR-PROV-085 provides the
  provider-layer behavior it tests); performance budget consolidation for NFR-AUTH-003.
- **Volume 13**: conformance suites (adapter, offline, canary leak scans) that verify this
  fragment's requirements.
- **Volume 1**: PRD-002, PRD-003, PRD-005, PRD-006, PRD-009, PRD-010; MVP provider seed and
  offline guarantee list; SM-04, SM-05, SM-12, SM-13 bindings.
- **Fragment A (this volume)**: provider contract keystone FR-PROV-001; capability
  negotiation (chapter 02); resilience thresholds and routing policy (chapter 05); error
  normalization classes (chapter 06) used by chapters 08–11.
