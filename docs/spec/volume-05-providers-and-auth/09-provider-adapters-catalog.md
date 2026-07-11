# 09 — Provider Adapters Catalog

This chapter catalogs the provider adapters Andromeda specifies, classifies each by delivery
phase (ADR-065), and fixes the requirements for the three MVP seed adapters mandated by
Volume 1 chapter 05 (MVP provider seed): the **generic OpenAI-compatible adapter**, the
**Anthropic adapter**, and the **Ollama adapter**. Every adapter implements the ProviderPort
behavioral contract and the declaration set of the provider-contract keystone (FR-PROV-001,
chapter 01); transport follows the ADR-019 baseline (stdlib HTTP against documented public
APIs; official SDK adoption per adapter PENDING VALIDATION per ADR-019). Presence in this
catalog does not imply MVP support (Volume 1 rule); it implies a committed phase and a stable
slug.

Two rules govern the catalog:

1. **Generic-adapter-first (ADR-065).** Any service exposing an OpenAI-compatible API surface
   is reachable from MVP through the generic adapter with a configured base URL and key. A
   *dedicated* adapter for such a service exists only to add declared value — model discovery,
   capability declarations, pricing metadata, provider-specific error normalization — and is
   phased by expected user value, not by reachability.
2. **No invented facts.** Capability, endpoint, pricing, and rate-limit details that this
   specification has not verified against the provider's public documentation are marked
   PENDING VALIDATION and resolved at the adapter's implementation (open question V5B-OQ-1);
   the catalog never asserts them. Authentication families listed below name the mechanism
   family the adapter declares under FR-AUTH-001; exact header names, endpoint paths, and
   token semantics are fixed at implementation from the provider's documentation.

## Catalog

| # | Adapter slug | Service | Kind | Dedicated-adapter phase | Auth family | Transport |
|---|---|---|---|---|---|---|
| 1 | `openai_compatible` | Any OpenAI-compatible endpoint | Cloud or local (endpoint-configured) | MVP | `none` or `api_key` | HTTP + SSE streaming |
| 2 | `anthropic` | Anthropic | Cloud | MVP | `api_key` | HTTP + SSE streaming |
| 3 | `ollama` | Ollama local server | Local | MVP | `none` | HTTP (localhost) |
| 4 | `openai` | OpenAI | Cloud | Beta | `api_key` | HTTP + SSE streaming |
| 5 | `gemini` | Google Gemini API | Cloud | Beta | `api_key`; service-account/Vertex surface PENDING VALIDATION | HTTP + streaming |
| 6 | `openrouter` | OpenRouter | Cloud aggregator | Beta | `api_key` | HTTP + SSE streaming |
| 7 | `mistral` | Mistral La Plateforme | Cloud | Beta | `api_key` | HTTP + SSE streaming |
| 8 | `azure_openai` | Azure OpenAI | Cloud | v1 | `api_key`; Entra ID / managed identity PENDING VALIDATION | HTTP + SSE streaming |
| 9 | `groq` | Groq | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 10 | `together` | Together AI | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 11 | `deepseek` | DeepSeek | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 12 | `xai` | xAI | Cloud | v1 | `api_key` | HTTP + SSE streaming |
| 13 | `vllm` | vLLM server | Local or self-hosted | v1 | `none` or `api_key` (server-configured) | HTTP + SSE streaming |
| 14 | `lm_studio` | LM Studio local server | Local | v1 | `none`; details PENDING VALIDATION | HTTP (localhost) |
| 15 | `llama_cpp` | llama.cpp server | Local | v1 | `none` or `api_key` (server-configured); details PENDING VALIDATION | HTTP (localhost) |
| 16 | `litellm` | LiteLLM proxy | Self-hosted aggregator | v2 | `api_key` (proxy-issued keys); details PENDING VALIDATION | HTTP + SSE streaming |
| 17 | `localai` | LocalAI | Local | v2 | `none` or `api_key`; details PENDING VALIDATION | HTTP (localhost) |
| 18 | `fastchat` | FastChat server | Local or self-hosted | Future | details PENDING VALIDATION | HTTP |
| 19 | `text_generation_webui` | Text Generation WebUI API | Local | Future | details PENDING VALIDATION | HTTP |

Future providers beyond this list integrate through the Extension SDK adapter surface
(FR-SDK-001) or, where OpenAI-compatible, immediately through adapter 1.

## Requirements

### FR-PROV-080 — Adapter catalog and phase classification

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: Provider Layer, Configuration Manager, CLI, TUI
- Dependencies: ADR-019, ADR-065; FR-PROV-001
- Related risks: RISK-PROV-080

#### Description

Andromeda MUST maintain the adapter registry exactly as cataloged above: each adapter has a
stable slug (the `adapter` value of Provider rows, Volume 2), a committed dedicated-adapter
phase, a declared authentication family consistent with FR-AUTH-001, and a declaration set per
FR-PROV-001. The three MVP seed adapters (`openai_compatible`, `anthropic`, `ollama`) MUST be
functional at MVP exit. Adapters MUST NOT ship before their phase; services in the catalog are
reachable earlier only through the generic adapter (rule 1). Catalog changes — new adapters,
phase moves, slug retirements — occur only through the Volume 0 change procedure.

#### Motivation

The brief mandates the integration list with per-phase classification and explicitly states
list membership does not imply MVP support; a governed registry prevents both scope creep and
silent slug drift that would corrupt historical attribution (Volume 2 INV-PRV-01).

#### Actors

Provider Layer; adapter implementers; users configuring providers.

#### Preconditions

FR-PROV-001 declaration validation exists.

#### Main flow

1. A `[providers.<slug>]` configuration names an adapter slug.
2. The registry resolves the slug; the adapter's declaration set is validated.
3. The Provider row is created with `connection_state = configured`.

#### Alternative flows

- Unknown adapter slug: configuration validation fails (Volume 10 E-CFG semantics) listing
  registered slugs.
- Adapter slug known but phase-gated out of the current build: the same validation failure
  names the phase.

#### Edge cases

- A service's dedicated adapter arrives while users run it through `openai_compatible`:
  both registrations may coexist (different slugs, different Provider rows); migration is a
  user action, never automatic (Principle 7 — no silent provider change).
- An Extension-provided adapter claims a catalog slug: rejected; catalog slugs are reserved.

#### Inputs

Configuration; adapter declarations.

#### Outputs

Validated Provider registrations; registry listings (`provider` command family, Volume 8).

#### States

Provider connection machine (chapter 11) governs registered providers.

#### Errors

Configuration errors surface per Volume 10; runtime provider errors per the E-PROV family
(chapters 01–06).

#### Constraints

Slugs are permanent once released (INV-PRV-01); phases only move via change procedure.

#### Security

Adapter declarations are validated against FR-AUTH-001 at registration; extension adapters
additionally pass Volume 6 trust evaluation.

#### Observability

Registry listings expose slug, phase, auth family, and declaration summary;
`provider.deprecation.announced` (below) surfaces provider-announced deprecations.

#### Performance

Registry resolution is an in-memory lookup.

#### Compatibility

The catalog is platform-neutral; local adapters depend on the user running the server.

#### Acceptance criteria

- Given MVP exit, when the acceptance suite runs, then `openai_compatible`, `anthropic`, and
  `ollama` pass their adapter conformance suites.
- Given a configuration naming an unregistered slug, when validation runs, then it fails with
  the registered-slug list and exit code 3.
- Negative case: given an extension attempting to register slug `anthropic`, then registration
  is rejected.
- Observability case: given a registry listing, then every adapter shows slug, phase, and auth
  family without secret data.

#### Verification method

Adapter conformance suites (Volume 13) per shipped adapter; registry unit tests; configuration
validation tests; release audit against the catalog table.

#### Traceability

PRD-002; Volume 1 chapter 05 (MVP provider seed); ADR-065; FR-PROV-001.

### FR-PROV-081 — Generic OpenAI-compatible adapter

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: Provider Layer, Authentication Layer, Context Manager, Indexing Engine
- Dependencies: ADR-019; FR-PROV-001, FR-AUTH-002; FR-PROV-084
- Related risks: RISK-PROV-080, RISK-PROV-081

#### Description

The `openai_compatible` adapter MUST implement ProviderPort against a **user-configured base
URL** exposing the OpenAI-compatible API surface, defined for Andromeda's purposes as: a chat
completions endpoint accepting the messages/tools/stream request families, SSE streaming
responses, an optional models listing endpoint for `DiscoverModels`, and an optional
embeddings endpoint for `Embed`. Authentication is `none` or `api_key` (bearer header) per
configuration. Because "OpenAI-compatible" is a convention, not a conformance mark, the
adapter MUST treat every optional surface as absent until detected (capability negotiation,
chapter 02) and MUST degrade per the declared strategies — never by silent simulation
(Principle 2). Capability declarations for any concrete service configured through this
adapter are established by explicit detection plus user configuration override, and recorded
per model.

#### Motivation

One adapter covers the majority of cloud and local services in the catalog at MVP (SM-01
leverage); it is also the vendor-agnosticism proof required by Volume 1's seed justification.

#### Actors

Users configuring endpoints; Provider Layer consumers.

#### Preconditions

Reachable base URL; credential per FR-AUTH-002 when the service requires one.

#### Main flow

1. `[providers.<slug>]` sets `adapter = "openai_compatible"` and the base URL.
2. Verification (chapter 11 machine) probes the declared surfaces and records capabilities.
3. Chat/stream/embed requests execute per the ProviderPort semantics with uniform retry,
   timeout, and cancellation behavior (ADR-019 baseline; policies per chapter 05).

#### Alternative flows

- Models endpoint absent: `DiscoverModels` returns the configured static model list from
  `[providers.<slug>]`; discovery absence is recorded, not erased.
- Embeddings endpoint absent: `Embed` returns the capability-unavailable E-PROV class so the
  Indexing Engine falls back per Volume 7.

#### Edge cases

- Divergent SSE framing or nonstandard finish signals: normalized per chapter 06; unparseable
  stream events terminate the stream with a malformed-response error, never a hang.
- Server answering the chat endpoint but violating tool-calling semantics: tool calling is
  declared absent for that model after detection fails; the run proceeds per degradation
  strategy with user notification.
- Base URL behind an enterprise proxy: FR-AUTH-006 applies unchanged.

#### Inputs

Base URL, optional key, optional static model/capability configuration.

#### Outputs

ProviderPort results; detected capability records; normalized errors.

#### States

Provider connection machine (chapter 11).

#### Errors

E-PROV family (chapters 01, 06); auth failures per E-AUTH.

#### Constraints

No service-specific branches inside the adapter: behavior differences enter only through
detection results and configuration (Principle 1 applied inside the generic adapter).

#### Security

TLS verification per FR-AUTH-006 rules; loopback endpoints follow FR-PROV-084 locality
semantics.

#### Observability

Standard provider request/stream events (chapters 01, 03–06); detection outcomes are recorded
per Volume 2 INV-MDL-03.

#### Performance

Streaming overhead within SM-08 budget; the adapter adds no buffering beyond bounded SSE
frame assembly.

#### Compatibility

Any endpoint honoring the defined surface, cloud or local, including catalog services 9–19
before their dedicated adapters ship.

#### Acceptance criteria

- Given a conforming endpoint with a key, when the conformance suite runs, then chat,
  streaming, and (where offered) discovery and embeddings pass.
- Given an endpoint without a models listing, when discovery runs, then configured static
  models are returned and the absence is visible in provider status.
- Negative case: given a stream with malformed SSE data, then the stream ends with a
  normalized malformed-response error and no partial result is presented as complete.
- Error case: given a 401 from the endpoint, then E-AUTH-002 semantics surface through the
  session layer.
- Observability case: every request carries correlation IDs linking run, provider request,
  and (where applicable) tool invocation (SM-13).

#### Verification method

Adapter conformance suite against recorded fixtures plus a live local OpenAI-compatible
server in CI (SM-04 second serving path); fault-injection for framing violations; offline
suite participation.

#### Traceability

PRD-002, PRD-003; Volume 1 MVP provider seed; ADR-019, ADR-065; FR-PROV-001.

### FR-PROV-082 — Anthropic adapter

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: Provider Layer, Authentication Layer
- Dependencies: ADR-019; FR-PROV-001, FR-AUTH-002
- Related risks: RISK-PROV-080

#### Description

The `anthropic` adapter MUST implement ProviderPort against Anthropic's documented public
Messages API using `api_key` authentication (FR-AUTH-002), including: streaming via the
documented SSE event families, tool calling per the documented tool-use request/response
shapes, declared context-window and output-token metadata per model, and token-usage
reporting from documented response fields mapped to the `token_usage_reporting` capability.
Model discovery uses the documented models listing; capabilities per model are declared from
documentation and verified by detection at first use (chapter 02). SDK adoption
(anthropic-sdk-go) remains PENDING VALIDATION at implementation per ADR-019. Structural API
details beyond those verified in ADR-019 are fixed at implementation from current official
documentation, not from this chapter.

#### Motivation

Volume 1's seed requires a second, structurally different cloud API family to prove the
provider contract is not shaped around one vendor; Anthropic's Messages API differs
materially from the OpenAI-compatible surface in request shape, streaming events, and tool
semantics.

#### Actors

Provider Layer consumers; Authentication Layer.

#### Preconditions

Anthropic API key stored per FR-AUTH-002.

#### Main flow

1. Registration with `adapter = "anthropic"`; verification per chapter 11.
2. Requests translate contract types to the documented Messages API shapes at the adapter
   boundary (no vendor types escape, ADR-019).
3. Streaming and tool-call deltas map to `ChatEvent` per the ProviderPort semantics.

#### Alternative flows

- Documented API version header requirements change: the adapter pins the documented version
  it was implemented against and surfaces deprecation notices via
  `provider.deprecation.announced`.

#### Edge cases

- Responses reporting reasoning/thinking content: exposed only as officially provided
  summaries per Principle 7 (no private-reasoning promises); mapped to the `reasoning`
  capability where documented.
- Rate-limit responses: normalized to the E-PROV rate-limit class; retry/backoff policy per
  chapter 05.

#### Inputs

Contract requests; API key.

#### Outputs

Contract responses with usage accounting; capability records.

#### States

Provider connection machine (chapter 11).

#### Errors

E-PROV family via chapter 06 normalization; E-AUTH-002 for auth rejections.

#### Constraints

Vendor types confined to the adapter package; documented public API only (FR-AUTH-001
analogue for transport).

#### Security

Key injected per documented header scheme; never logged (NFR-AUTH-002).

#### Observability

Standard provider events; usage fields feed Cost Records (chapter 04).

#### Performance

Streaming within SM-08 overhead budget.

#### Compatibility

Cloud-only; macOS/Linux neutral.

#### Acceptance criteria

- Given a valid key, when the conformance suite runs against recorded fixtures, then chat,
  streaming, tool calling, and usage reporting pass.
- Given a documented-version deprecation notice, when detected, then
  `provider.deprecation.announced` emits and provider status shows it.
- Negative case: given an invalid key, then E-AUTH-002 surfaces with exit code 4.
- Observability case: token usage from responses appears in the run's Cost Records.

#### Verification method

Adapter conformance suite with recorded fixtures in CI; scheduled live smoke against the
documented API in a network-permitted job; contract-type leak check (no vendor types outside
the package).

#### Traceability

PRD-002; Volume 1 MVP provider seed; ADR-019; FR-PROV-001.

### FR-PROV-083 — Ollama adapter

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: Provider Layer, Indexing Engine, Context Manager
- Dependencies: ADR-019 (thin client decision); FR-PROV-001, FR-PROV-084
- Related risks: RISK-PROV-081

#### Description

The `ollama` adapter MUST implement ProviderPort against the documented Ollama REST API on
its default local endpoint `localhost:11434` (endpoints `/api/chat`, `/api/generate`,
`/api/embed` per ADR-019's verification), using the thin hand-rolled HTTP client ADR-019
fixes, with `auth_kind = none`. It anchors the offline guarantee (FR-PROV-085): all its
operations MUST function with no Internet connectivity when the server and models are local.
`DiscoverModels` enumerates locally installed models via the documented listing surface;
capability declaration per model follows chapter 02 detection with the honesty rules of
Principle 2 (local models vary widely in tool-calling and structured-output fidelity —
detection, not assumption). Model pulling/management is the user's action through Ollama
itself: Andromeda is not a model manager (Volume 1 non-goal 10) and MUST NOT download model
weights.

#### Motivation

Satisfies the "≥ 1 local provider" MVP minimum and the Local First principle with the most
common local serving path; keeps the no-network core path lean (ADR-019 dependency-weight
rationale).

#### Actors

Users running Ollama; Provider Layer consumers; Indexing Engine (local embeddings).

#### Preconditions

Ollama server running and reachable on the configured endpoint (default
`http://localhost:11434`, overridable in `[providers.<slug>]`).

#### Main flow

1. Registration; verification probes the documented endpoints.
2. Chat/stream/embed requests execute locally; usage accounting reports what the documented
   response fields expose.
3. Discovery lists installed models; capability detection records per-model results.

#### Alternative flows

- Server not running: connection-refused normalizes to the E-PROV connectivity class with a
  user message naming the endpoint and suggesting the server be started; the Provider row
  goes `unavailable` per chapter 11.

#### Edge cases

- Model absent for a request: the documented error normalizes to a model-not-found E-PROV
  class; Andromeda suggests `DiscoverModels`-based correction and never pulls the model
  itself.
- Cost accounting: local inference has no provider price; Cost Records carry token counts
  with a zero-cost marking (chapter 04 semantics), when the documented fields expose counts.
- Non-default endpoint on a remote host: locality is then determined per FR-PROV-084 rules;
  offline guarantees apply only to local endpoints.

#### Inputs

Endpoint configuration; local model inventory.

#### Outputs

Contract responses; local capability records.

#### States

Provider connection machine (chapter 11).

#### Errors

E-PROV family; no E-AUTH involvement (`auth_kind = none`).

#### Constraints

Thin client per ADR-019 (no ollama module import); no model weight management.

#### Security

Loopback plain HTTP permitted per FR-PROV-084 locality rules; non-loopback endpoints require
TLS or explicit locality classification.

#### Observability

Standard provider events; offline suite asserts zero network-interface access beyond
loopback.

#### Performance

Streaming within SM-08 budget; no adapter-added model load management.

#### Compatibility

macOS and Linux Tier 1; server versions per Ollama's documented API stability.

#### Acceptance criteria

- Given a running local server with an installed model, when the offline suite disables all
  non-loopback networking, then chat, streaming, discovery, and embeddings all pass (SM-05
  participation).
- Given the server stopped, then requests fail with the connectivity class, exit code 7, and
  the provider goes `unavailable`.
- Negative case: given a request naming an uninstalled model, then the model-not-found class
  surfaces and no download occurs.
- Observability case: usage records mark local inference as zero-cost with token counts when
  reported.

#### Verification method

Conformance suite against a pinned local Ollama in CI (SM-04 first serving path); offline
test suite with OS-level network disablement; dependency audit asserting no ollama module
import.

#### Traceability

PRD-003; Volume 1 MVP provider seed and offline guarantees; ADR-019; FR-PROV-084,
FR-PROV-085.

## Per-adapter notes

Notes for adapters beyond the MVP seed. Every unmarked detail below names only the mechanism
family; all quantitative and structural specifics are fixed at implementation from official
documentation (V5B-OQ-1).

#### OpenAI (`openai`, Beta)

Dedicated adapter adds documented model discovery, capability and pricing metadata, and
provider-specific error normalization over what the generic adapter offers. Auth: `api_key`.
SDK adoption (openai-go, major-pinned) PENDING VALIDATION per ADR-019. Capability details per
model PENDING VALIDATION at implementation.

#### Google Gemini (`gemini`, Beta)

Targets the documented Gemini API with `api_key` auth. The separate Vertex AI surface
(service accounts, managed identity per FR-AUTH-005) is a distinct declaration set within the
same adapter and is PENDING VALIDATION at implementation. Capability details PENDING
VALIDATION.

#### OpenRouter (`openrouter`, Beta)

Aggregator exposing an OpenAI-compatible surface with `api_key` auth and a documented models
listing spanning many upstream vendors. Dedicated adapter exists chiefly for discovery
metadata (upstream attribution, pricing fields) — details PENDING VALIDATION. Routing *inside*
OpenRouter is the service's own behavior; Andromeda's routing policies (chapter 05) treat it
as one provider.

#### Mistral (`mistral`, Beta)

Documented public API with `api_key` auth; capability details (tool calling, embeddings)
PENDING VALIDATION at implementation.

#### Azure OpenAI (`azure_openai`, v1)

Deployment-scoped endpoints under the customer's Azure resource; auth families: `api_key` and
Microsoft Entra ID tokens including managed identity (FR-AUTH-005) — both PENDING VALIDATION
in their concrete flows. Endpoint shape (resource, deployment, API version) enters through
`[providers.<slug>]` endpoint configuration; details PENDING VALIDATION.

#### Groq (`groq`, v1) · Together (`together`, v1) · DeepSeek (`deepseek`, v1) · xAI (`xai`, v1)

Cloud services documenting OpenAI-compatible (or closely analogous) surfaces with `api_key`
auth; reachable at MVP through the generic adapter. Dedicated adapters add discovery,
capability, and pricing metadata plus provider-specific error normalization — all specifics
PENDING VALIDATION at each implementation.

#### vLLM (`vllm`, v1)

Self-hosted serving engine exposing a documented OpenAI-compatible server; auth is
server-configured (`none` or a static `api_key`). Dedicated adapter adds served-model
discovery and local-deployment ergonomics; specifics PENDING VALIDATION.

#### LM Studio (`lm_studio`, v1)

Local desktop server exposing an OpenAI-compatible endpoint; default port and any local SDK
surfaces PENDING VALIDATION. Auth family `none` on loopback.

#### llama.cpp (`llama_cpp`, v1)

The bundled server exposes an OpenAI-compatible HTTP surface with optional server-configured
key; endpoint and flag specifics PENDING VALIDATION. Loopback locality per FR-PROV-084.

#### LiteLLM (`litellm`, v2)

Self-hosted proxy aggregating upstream providers behind an OpenAI-compatible surface with
proxy-issued keys; virtual-key semantics and admin surfaces PENDING VALIDATION. Classified v2
because the generic adapter covers the core path and the added value is administrative.

#### LocalAI (`localai`, v2)

Local OpenAI-compatible server; optional key; endpoint specifics PENDING VALIDATION.

#### FastChat (`fastchat`, Future) · Text Generation WebUI (`text_generation_webui`, Future)

Local/self-hosted servers with OpenAI-compatible API surfaces of varying maintenance status;
both reachable via the generic adapter today. Dedicated adapters are uncommitted (`Future`);
all specifics PENDING VALIDATION, including whether the projects' API surfaces remain
maintained enough to justify dedicated adapters at all.

## Events minted in this chapter

| Event | Emitted when | Payload summary |
|---|---|---|
| `provider.deprecation.announced` | An adapter detects a provider-documented deprecation (model, API version, endpoint) affecting a registered Provider | provider slug, subject kind, subject name, documented sunset date when available |

Envelope and delivery semantics per Volume 10. The event backs Principle 7's requirement that
provider-side change is announced, never silent.

## Risks

### RISK-PROV-080 — Adapter catalog drift against provider realities

- Category: External dependency
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: Per-adapter contract tests with recorded fixtures run in CI (ADR-019); PENDING
  VALIDATION markers resolved and recorded at each adapter's implementation; scheduled live
  smoke jobs for shipped cloud adapters; catalog review at every phase gate; deprecation
  detection via `provider.deprecation.announced`
- Detection: Contract-test failures; smoke-job failures; provider changelog review at phase
  gates; per-provider error-rate metrics
- Owner: Provider Layer (Volume 5)
- Status: Open

Nineteen external services evolve independently of this specification: endpoints move, auth
requirements tighten, capabilities appear and disappear, projects go dormant. The catalog
bounds the blast radius (one adapter, one leaf package) and the fixtures-plus-smoke regime
converts drift into failing CI rather than field incidents; the residual risk is the window
between a provider change and the next scheduled verification.
