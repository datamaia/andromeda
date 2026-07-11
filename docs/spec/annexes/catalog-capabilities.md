# Annex — Consolidated Capability Catalog

**Status:** Consolidated (Phase C). This annex is the corpus-wide index of the provider/
model capability enum — the frozen 14-value seed extended by exactly one value
(`token_counting`, [ADR-056](adr/ADR-056.md)) — with each value's level, testable meaning,
and degradation behavior, plus the committed capability posture of the three MVP adapters.
It is a *reference view*: the normative home of the enum, provenance resolution,
negotiation, and degradation is Volume 5 chapter
[02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) (FR-PROV-010,
FR-PROV-011); the adapter commitments are chapter
[09](../volume-05-providers-and-auth/09-provider-adapters-catalog.md)'s (FR-PROV-080..083).
This annex mints nothing and renames nothing. The enum is closed: values are added only by
ADR, absence of a value means *not available*, no component may infer a capability from a
provider or model name, and no component may simulate an absent capability silently.

## Resolution and negotiation, in one paragraph

A model's **effective capability set** resolves deterministically from four provenance
classes: start from **`declared`** (the Adapter Declaration baseline, sourced from official
documentation), apply **`discovered`** (metadata officially reported by the provider's
discovery mechanism — it narrows or confirms, never adds unmappable values), apply
**`configured`** (user overrides in `[providers.<slug>.capability_overrides]` — removals
always allowed, additions only where the declaration marks the capability *configurable*),
then mask **`refuted`** values (failed verification probes; a refuted value is hidden until
re-verified and the change is evented). Every value carries its provenance class, visible
in the CLI/TUI provider views. Every request declares its required capabilities; the router
negotiates them against the target's effective set **before dispatch**; a gap resolves
through the missing capability's degradation strategy — never through silent simulation —
and every degradation emits `provider.degradation.applied` plus a user notification
(FR-PROV-011, FR-PROV-013). Verification depth is `providers.<slug>.verify_capabilities`
(`off` | `basic` | `probe`, default `basic`); probes use fixed synthetic prompts, run only
during connection verification or on explicit user command, and record
`verified`/`refuted` provenance. Related configuration keys are cataloged in
[catalog-config.md](catalog-config.md).

## The capability enum (15 values)

Each value declares its level (`provider` | `model` | `both`) and a testable meaning
(Volume 2 Capability attributes). Degradation strategies are the closed set `refuse`,
`report_unavailable`, `substitute` (opt-in via configuration only), and `reroute`
(re-enters routing; exhaustion yields E-PROV-016); a mandatory capability gap is
E-PROV-006, raised before any wire request.

| Capability | Level | Declaring it promises (testable meaning) | Default strategy when missing | Documented substitute (opt-in) | Defined in |
|---|---|---|---|---|---|
| `chat` | model | The model completes `Chat`/`ChatStream` requests with role-structured messages | refuse | none | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `streaming` | model | `ChatStream` delivers incremental events; not a local simulation over a blocking call | report_unavailable | non-streaming `Chat` presented without incremental delivery — an absence, not a simulation | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `tool_calling` | model | The model emits structured tool calls the adapter maps losslessly to the normal form (Vol 5 ch 03) | refuse (agent runs require it) | none — prompted pseudo-tool-calls are prohibited as silent simulation | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `parallel_tool_calling` | model | A single response may carry multiple independent tool calls, each individually addressable by ID | report_unavailable | sequential tool calls | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `structured_outputs` | model | The provider enforces an output schema natively (Vol 5 ch 03 `native` mode) | reroute, then refuse | `tool_call` or `prompted` modes per Vol 5 ch 03, each explicit and validated (`prompted` requires `allow_prompted = true`) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `reasoning` | model | The provider officially exposes reasoning artifacts: summaries and/or reasoning token counts — never private chain-of-thought | report_unavailable | none | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `vision` | model | Image content parts are accepted as input | refuse when the request carries the modality | none | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `audio_input` | model | Audio content parts are accepted as input | refuse when the request carries the modality | none | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `audio_output` | model | The model can produce audio output parts | refuse when the request carries the modality | none | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `embeddings` | model | `Embed` returns vectors for this model | refuse | none (the Indexing Engine operates without embeddings per ADR-020 / Volume 7) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `token_usage_reporting` | both | Responses carry official token usage; declared fields per the adapter's `UsageReporting` declaration (Vol 5 ch 04) | report_unavailable | accounting records `cost_basis = unavailable` (Vol 5 ch 04) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `cost_reporting` | both | The provider officially reports monetary cost per request or account usage | report_unavailable | local pricing tables yield estimates (Vol 5 ch 04) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `model_discovery` | provider | A documented enumeration mechanism backs `DiscoverModels` | report_unavailable | models from explicit configuration only | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `cancellation` | provider | Aborting the transport observably stops generation billing-side per provider documentation; absent it, cancellation is client-side only | report_unavailable | client-side abort; possible provider-side token spend recorded as such | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |
| `token_counting` | both | A documented counting mechanism backs `CountTokens` (ADR-056) | report_unavailable | Context Manager estimation (Volume 7) | [Vol 5 ch 02](../volume-05-providers-and-auth/02-capabilities-model-discovery.md) |

## MVP adapter capability posture (Volume 5, chapter 09)

The three MVP seed adapters are `openai_compatible`, `anthropic`, and `ollama`
(FR-PROV-080; auth families `none`/`api_key`, `api_key`, and `none` respectively). The
catalog chapter deliberately commits *postures*, not a boolean truth table: per-model
values are resolved at runtime through the provenance procedure, and unverified provider
facts are never asserted (chapter 09 rule 2). Cell vocabulary:

- **declared** — the chapter commits the adapter to declaring the capability from the
  documented API surface.
- **detected** — established per model by explicit detection (plus configuration override
  where marked configurable), never assumed.
- **not asserted** — the chapter makes no commitment; the value resolves at implementation
  from official documentation and the chapter 02 provenance procedure.

| Capability | `openai_compatible` (FR-PROV-081) | `anthropic` (FR-PROV-082) | `ollama` (FR-PROV-083) |
|---|---|---|---|
| `chat` | declared — the chat completions endpoint is the defining surface | declared — documented Messages API | declared — documented `/api/chat` |
| `streaming` | declared — SSE streaming responses are part of the defined surface | declared — documented SSE event families | declared — chat/stream requests execute locally |
| `tool_calling` | detected; configurable override permitted | declared — documented tool-use request/response shapes, verified by detection at first use | detected — local models vary widely in tool-calling fidelity |
| `parallel_tool_calling` | detected | not asserted | not asserted |
| `structured_outputs` | detected; configurable override permitted | not asserted | detected — local models vary widely in structured-output fidelity |
| `reasoning` | detected; configurable override permitted | declared where documented — officially provided summaries map to `reasoning` | not asserted |
| `vision` | detected; configurable override permitted | not asserted | not asserted |
| `audio_input` | not asserted | not asserted | not asserted |
| `audio_output` | not asserted | not asserted | not asserted |
| `embeddings` | detected — optional embeddings endpoint, treated as absent until detected; configurable override permitted | not asserted | declared — documented `/api/embed`; anchors local embeddings for the Indexing Engine |
| `token_usage_reporting` | detected | declared — documented response fields map to the capability | declared where documented fields expose counts; local inference recorded as zero-cost with token counts when reported |
| `cost_reporting` | not asserted | not asserted | not asserted (local inference has no provider price; Cost Records carry zero-cost markings per Vol 5 ch 04) |
| `model_discovery` | detected — optional models listing endpoint; when absent, configured static models serve and the absence is recorded, not erased | declared — documented models listing | declared — enumerates locally installed models; Andromeda never pulls model weights |
| `cancellation` | not asserted | not asserted | not asserted |
| `token_counting` | not asserted | not asserted | not asserted |

Chapter 02 fixes the configurable set for the generic adapter: the `openai_compatible`
declaration marks `tool_calling`, `structured_outputs`, `vision`, `embeddings`, and
`reasoning` configurable (an arbitrary endpoint's support cannot be known statically), so
`capability_overrides` may add those values; overrides may remove any value on any adapter.
The `ollama` adapter anchors the offline guarantee: all its operations function with no
Internet connectivity when the server and models are local (FR-PROV-083). Adapters 4–19 of
the catalog carry no committed capability posture — their specifics resolve at each
adapter's implementation from official documentation (register entry V5B-OQ-1), and
OpenAI-compatible services are reachable earlier through adapter 1
([ADR-065](adr/ADR-065.md)).

## Capability observability

Capability changes are never silent (FR-PROV-013): `provider.capability.changed` (old set,
new set, provenance), `provider.capability.verified` (capability, `verified`/`refuted`
outcome), `provider.degradation.applied` (capability, strategy, reason, run correlation),
and `provider.discovery.completed` are the change events, cataloged with producers in
[catalog-events.md](catalog-events.md); effective sets are snapshotted into run records for
reproducibility.

## Consolidation notes

- **Coverage.** 15 enum values (14 frozen seed values plus the single ADR-056 extension)
  round-trip against Volume 5 chapter 02's enum table, degradation table, and the register;
  the three MVP adapter columns reproduce chapter 09's committed postures (FR-PROV-081,
  FR-PROV-082, FR-PROV-083). No register conflicts with a defining chapter.
- **Degradation column merge.** Chapter 02's degradation table keys three modality values
  (`vision`, `audio_input`, `audio_output`) in one row ("refuse when the request carries
  the modality"); this catalog splits them into one row each with identical content.
- **"Not asserted" is a posture, not an absence claim.** Chapter 09 rule 2 forbids
  asserting unverified provider facts; a "not asserted" cell means the runtime resolves the
  value through the provenance procedure at implementation and configuration time — it does
  not mean the capability is known to be missing.
- **Truth-table stance.** The [compatibility matrix](compatibility-matrix.md) section 3
  records the same decision from the platform side: no adapter × capability truth table is
  published anywhere in the corpus, by design.
