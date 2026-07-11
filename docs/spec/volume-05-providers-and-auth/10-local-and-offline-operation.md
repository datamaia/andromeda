# 10 — Local and Offline Operation

This chapter maps the product's Local First commitments (PRD-003; Volume 1 chapter 04,
Principle 3, offline guarantee list) onto provider-layer behavior: what "local provider"
means precisely, how the Provider Layer and Authentication Layer behave with no Internet
connectivity, and how the local-model quality bar (SM-04) is formalized. The governing
decision is ADR-066: **no reachability probing, passive fail-fast classification, and a
strict locality rule** — offline is a handled state of the provider layer, never a failure of
the product (Volume 1 non-goal 7).

## Locality rule

A configured provider endpoint is **local** exactly when its host resolves to a loopback
address (`127.0.0.0/8`, `::1`) or a Unix domain socket. Everything else — including LAN and
"my own server" hosts — is **remote**: it may be reachable without Internet, but it receives
no offline guarantees and full TLS requirements (FR-AUTH-006). Locality is computed from the
resolved address, never from hostname string matching (ADR-066), so `localhost`-lookalike DNS
names cannot acquire loopback privileges.

## Offline guarantee mapping

The eleven operations of the Volume 1 offline guarantee list, mapped to their provider-layer
involvement. Verification is the offline suite (SM-05; formalized by Volume 12): all network
interfaces disabled at the OS level except loopback, local provider running.

| # | Guaranteed operation | Provider-layer involvement |
|---|---|---|
| 1 | Opening a workspace | None — no provider call occurs |
| 2 | Querying local memory | None directly; semantic queries use locally stored embeddings (Volume 7); new embedding computation requires a local embeddings-capable provider or lexical fallback |
| 3 | Indexing files | Lexical indexing: none. Semantic indexing: local `Embed` only (FR-PROV-083, FR-PROV-081 against a local server), else the Indexing Engine's no-embeddings mode (Volume 7) |
| 4 | Executing local tools | None |
| 5 | Using local Git | None |
| 6 | Using the terminal | None |
| 7 | Executing offline-capable workflow steps | Model-using steps require a local provider; the workflow declares it (Volume 4) |
| 8 | Creating patches | Model-assisted creation requires a local provider; patch mechanics need none |
| 9 | Reviewing diffs | None |
| 10 | Running tests | None |
| 11 | Viewing logs | None |

Agent runs with a local provider are additionally covered by FR-PROV-084/085: the full
plan–act–observe loop, including streaming and tool calling, MUST work offline when every
element of the session is local.

## Requirements

### FR-PROV-084 — Local provider operation

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: Provider Layer, Authentication Layer, Agent Engine, Indexing Engine
- Dependencies: ADR-066; FR-PROV-081, FR-PROV-083; FR-AUTH-001
- Related risks: RISK-PROV-081

#### Description

Providers whose endpoints are local under the locality rule MUST operate with zero
authentication requirement when configured with `auth_kind = none` (no Credential row, no
Secret Store involvement, no `credential_access` evaluation on the request path), MUST be
usable over plain HTTP on loopback (TLS optional there, mandatory elsewhere), and MUST be
fully functional — chat, streaming, tool calling per detected capability, embeddings where
served, discovery — with all non-loopback networking disabled. Local providers participate in
routing, capability negotiation, cost accounting (zero-cost token records), and observability
identically to cloud providers: locality changes transport premises, never contract
semantics.

#### Motivation

PRD-003 makes local operation an identity property; uniform contract semantics keep
"local-first" from becoming a second, weaker product path (Principle 5's single-runtime logic
applied to providers).

#### Actors

Users running local servers; every ProviderPort consumer.

#### Preconditions

A local inference server is running (Ollama, or any OpenAI-compatible local server).

#### Main flow

1. Registration with a loopback endpoint and `auth_kind = none`.
2. Verification (chapter 11) probes locally; capabilities are detected and recorded.
3. Sessions run entirely against the local endpoint.

#### Alternative flows

- Local endpoint configured with a key (server-enforced): FR-AUTH-002 applies normally;
  locality and auth are orthogonal.

#### Edge cases

- Loopback endpoint proxied by `[auth.proxy]` settings: loopback traffic MUST bypass proxies
  regardless of `no_proxy` configuration (FR-AUTH-006 exemption uses the same locality rule).
- A local server exposing a remote tunnel hostname: remote by rule; its offline claims are
  simply false and verification fails offline — honest classification, no exception.
- Capability variance across local models: detection per chapter 02 governs; a local model
  without `tool_calling` runs plan-only or degrades per documented strategy with user
  notification, never silent simulation (Principle 2).

#### Inputs

Loopback endpoint configuration; local model inventory.

#### Outputs

Fully functional provider service without credentials or Internet.

#### States

Provider connection machine (chapter 11), identical for local providers.

#### Errors

E-PROV connectivity classes when the server is down; no E-AUTH involvement at
`auth_kind = none`.

#### Constraints

Locality by resolved address only; no plain HTTP off loopback; no reachability probing of
anything (ADR-066).

#### Security

Loopback plain-HTTP exemption cannot be extended by configuration; local providers still pass
through permission and observability layers identically (a local model's tool calls are as
governed as a cloud model's).

#### Observability

Runs record provider slug and model identically; cost records carry zero-cost token entries;
the offline suite asserts zero non-loopback socket activity.

#### Performance

Local inference latency is hardware-dependent and out of Andromeda's budgets; Andromeda-added
overhead budgets (SM-08) apply unchanged.

#### Compatibility

macOS and Linux Tier 1; any server meeting the FR-PROV-081 surface or the Ollama documented
API.

#### Acceptance criteria

- Given a loopback provider with `auth_kind = none`, when a full agent run executes with
  non-loopback networking disabled, then it completes with streaming and recorded accounting,
  and no Secret Store access occurs.
- Given the same provider, when permissions are inspected, then tool invocations were
  permission-evaluated exactly as with a cloud provider.
- Negative case: given a non-loopback endpoint claiming locality via a DNS name resolving
  externally, then it is classified remote and plain HTTP is refused.
- Observability case: the offline suite's socket monitor records loopback-only activity.

#### Verification method

Offline test suite (OS-level network disablement) running the UC-09 journey; socket-activity
assertion; locality-classification unit tests over address families; permission-parity
integration test.

#### Traceability

PRD-003; Volume 1 offline guarantee list; ADR-066; FR-PROV-083.

### FR-PROV-085 — Offline behavior of the provider layer

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: Provider Layer, Authentication Layer, Agent Engine, CLI, TUI
- Dependencies: ADR-066; FR-PROV-084; FR-AUTH-010; chapter 11 machines
- Related risks: RISK-PROV-081

#### Description

With no Internet connectivity, the provider layer MUST behave as follows (ADR-066):

1. **No probing.** Andromeda MUST NOT test reachability of anything to "detect offline";
   connectivity state is inferred passively from actual request failures.
2. **Fail fast, classify honestly.** A cloud provider request failing with a
   connectivity-class error fails within its normal timeout, normalizes per chapter 06, marks
   the Provider `unavailable` (chapter 11), and is NOT retried by the resilience layer's
   transient policy (connectivity-class failures suppress retry storms; policy details in
   chapter 05).
3. **Local paths unaffected.** Requests to local providers proceed normally; routing and
   fallback consider only providers not `unavailable`, so sessions configured with local
   providers never wait on cloud timeouts.
4. **Authentication degrades cleanly.** Credential resolution is local-only (SecretStorePort
   contract) and works offline; token refresh requiring network fails fast per FR-AUTH-010's
   offline case, leaving still-valid material in use; establishment flows requiring network
   (OAuth) fail with a connectivity-classified error, not a hang.
5. **Recovery is user-visible.** A later successful request (or explicit re-verification)
   returns the provider to service through the chapter 11 machine with events; nothing
   auto-polls.

#### Motivation

Connectivity loss is a handled state (Volume 1 non-goal 7); the failure mode this requirement
kills is the product freezing on cloud timeouts or spamming retries while the user works
locally.

#### Actors

Provider Layer; Authentication Layer; users offline by choice or circumstance.

#### Preconditions

None — the behavior is unconditional.

#### Main flow

1. A cloud request fails with a connectivity-class error.
2. Normalization, provider → `unavailable`, event emitted, error surfaced with a
   local-alternative hint when local providers are registered.
3. Local-provider work continues throughout.

#### Alternative flows

- User runs explicit re-verification (`provider` command family): the chapter 11 verifying
  path executes once, on demand.

#### Edge cases

- Captive portals and DNS blackholes (connection succeeds, garbage returns): normalize to
  malformed-response or connectivity classes per chapter 06; the no-probing rule means
  Andromeda never "detects" portals, it just fails the real request honestly.
- Partial connectivity (DNS works, TCP blocked): whatever class the real failure produces
  governs; no inference beyond the observed failure.
- Offline at startup with persisted `available` cloud providers: rows keep their persisted
  state until a real request fails (staleness handling per chapter 11 recovery).

#### Inputs

Real request outcomes.

#### Outputs

Classified failures; updated connection states; user-visible provider status.

#### States

Provider connection machine transitions (`available`/`degraded` → `unavailable`;
re-verification path) per chapter 11.

#### Errors

E-PROV connectivity classes (chapter 06); E-AUTH-003 for refresh exhaustion offline.

#### Constraints

Zero network access attempts from any provider-layer component while all configured
providers are local (SM-05's "0 network-access attempts observed" applied to this layer).

#### Security

No probing also means no beaconing: an offline machine generates zero provider-layer
network traffic.

#### Observability

`provider.connection.lost` / `provider.connection.recovered` events (chapter 11); provider
status surfaces in CLI/TUI; failure classes visible in run records.

#### Performance

Fail-fast bounds: a connectivity-class failure surfaces within the configured request
timeout with no added retry delay.

#### Compatibility

Platform-neutral; the offline suite runs on all Tier 1 platforms.

#### Acceptance criteria

- Given all interfaces disabled except loopback and a session on a local provider, when the
  UC-09 journey runs, then it completes and the socket monitor records zero non-loopback
  attempts.
- Given a cloud provider request while offline, then it fails within its timeout with a
  connectivity class, the provider goes `unavailable`, exactly zero automatic retries occur,
  and the error suggests registered local providers.
- Negative case: given connectivity restored, when no request is made, then the provider
  remains `unavailable` (no background polling) until the next real use or explicit
  re-verification.
- Observability case: the loss and recovery both emit their events with timestamps and
  correlation IDs.

#### Verification method

Offline suite with OS-level network disablement (SM-05 method); fault-injection for
captive-portal and partial-connectivity shapes; retry-count assertion on connectivity-class
failures; socket-activity monitor.

#### Traceability

PRD-003; SM-05; ADR-066; FR-PROV-084; chapter 11 Provider connection machine.

## Relation to the SM-04 formalization

The local-model quality bar (SM-04) is formalized as **NFR-PROV-002** in
[chapter 02](02-capabilities-model-discovery.md); this chapter supplies the two serving paths
that NFR measures — the `ollama` adapter (FR-PROV-083) and the `openai_compatible` adapter
against a local server (FR-PROV-081) under the FR-PROV-084 locality regime — and the offline
condition (FR-PROV-085) under which the SM-05 suite exercises them. No second formalization is
minted here.

## Risks

### RISK-PROV-081 — Local model and server capability variance

- Category: Technical / external dependency
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: Capability detection over declaration trust (chapter 02) applied uniformly to
  local paths; NFR-PROV-002 conformance gating with pinned models; documented degradation
  strategies with user notification instead of silent simulation (Principle 2); the honesty
  check makes over-declaration a measured defect
- Detection: Conformance-suite failures per release; capability-detection mismatch events;
  field error rates on local providers
- Owner: Provider Layer (Volume 5)
- Status: Open

Local serving stacks differ in tool-calling fidelity, streaming framing, and structured-output
support far more than cloud APIs do, and the same server yields different behavior per loaded
model. Detection-first capability handling turns this variance into recorded, per-model facts;
the residual risk is models that pass detection probes but fail under real workloads, bounded
by the conformance suite's agent-loop scenarios.
