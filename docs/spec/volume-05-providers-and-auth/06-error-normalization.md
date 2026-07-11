# 06 — Error Normalization

Every failure crossing the provider boundary is normalized into the **E-PROV** family under
the ADR-016 envelope. Consumers — Agent Engine, Workflow Engine, CLI/TUI, scripts — branch
on stable codes and classes, never on provider-specific conditions (Principle 1). No raw
transport, driver, or provider error type escapes the Provider Layer (Volume 3 port rule 2).

## Normalization pipeline

Normalization is two-layered (ADR-061):

1. **Adapter layer.** Each adapter maps wire conditions — HTTP status plus the provider's
   documented error body schema — to a normalized code using its declared `ErrorMapping`
   rules (FR-PROV-002). The adapter preserves the provider's raw error identity (documented
   error type/code strings and a redacted body excerpt bounded at 2048 bytes) in the
   envelope's safe technical detail. Unmapped conditions normalize to E-PROV-008
   (malformed/undocumented response) or E-PROV-012 (documented-but-unclassified provider
   error) — never to an invented specific code.
2. **Router layer.** The Provider Router enriches the normalized error with routing context
   (attempt number, breaker state, candidates skipped, fallback disposition) and finalizes
   retryability per the matrix below. Retryability derives from the normalized code only —
   raw wire conditions never drive retry decisions.

Generic origin mapping (adapters specialize per documented error schemas; catalog chapter
09):

| Origin condition (typical) | Normalized code |
|---|---|
| DNS/TCP/TLS failure, connection refused | E-PROV-001 |
| HTTP 401 / 403 (authentication/authorization) | E-PROV-002 |
| HTTP 429 (rate limit) | E-PROV-003 |
| Documented quota/billing-exhausted error body | E-PROV-004 |
| HTTP 404 on a model resource / documented unknown-model error | E-PROV-005 |
| HTTP 400 with documented validation error body | E-PROV-007 |
| Documented context-length error body | E-PROV-014 |
| Documented content-policy refusal | E-PROV-013 |
| HTTP 5xx | E-PROV-012 |
| Undocumented shape, unparseable body, schema-mapping failure | E-PROV-008 |
| Transport cut mid-stream | E-PROV-009 |
| Deadline expiry (any timeout class) | E-PROV-010 |
| Context cancellation | E-PROV-011 |

## Retryability and class matrix

| Code | Retryable (router) | Breaker-eligible | Fallback trigger class |
|---|---|---|---|
| E-PROV-001 | yes | yes | `unreachable` |
| E-PROV-002 | after single-flight refresh only | no | `auth_failed` (opt-in, guard F7) |
| E-PROV-003 | yes, delay-signal aware | yes (ratio only) | `rate_limited` |
| E-PROV-004 | no | no | `quota_exhausted` |
| E-PROV-005 | no | no | — |
| E-PROV-006 | no | no | `capability_gap` (via chapter 02 reroute) |
| E-PROV-007 | no | no | — |
| E-PROV-008 | yes (once) | yes | `internal_error` |
| E-PROV-009 | no (post-delivery; chapter 03) | yes | — |
| E-PROV-010 | yes | yes | `timeout` |
| E-PROV-011 | no | no | — |
| E-PROV-012 | yes | yes | `internal_error` |
| E-PROV-013 | no | no | — |
| E-PROV-014 | no | no | — |
| E-PROV-015 | after open duration | not applicable | `breaker_open` |
| E-PROV-016 | no | no | — (terminal routing outcome) |
| E-PROV-017 | bounded re-request (chapter 03) | no | — |
| E-PROV-018 | no (agent-level decision) | no | — |
| E-PROV-019 | no | not applicable | — |

### FR-PROV-050 — Provider error normalization

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: adapters, Provider Router, all ProviderPort consumers, CLI/TUI
- Dependencies: FR-PROV-001, FR-PROV-002; ADR-016, ADR-061
- Related risks: RISK-PROV-001

#### Description

Every failure on any ProviderPort path MUST surface as exactly one E-PROV code from the
catalog below, carrying the full ADR-016 envelope, produced by the two-layer pipeline.
Retryability, breaker eligibility, and fallback trigger classes MUST follow the matrix.
Unrecognized conditions normalize to E-PROV-008/E-PROV-012 with raw identity preserved in
safe detail — the family is total over the failure space.

#### Motivation

Uniform failure semantics are what make retries, breakers, fallback, agent error handling,
and scripting portable across nineteen adapters (PRD-002; ADR-016 rationale).

#### Actors

Adapters (map); router (enrich, decide); consumers (branch on codes).

#### Preconditions

Adapter `ErrorMapping` validated at registration (FR-PROV-002).

#### Main flow

Wire condition → adapter mapping → router enrichment → typed `PortError` to the consumer,
event emission, exit-code mapping at the CLI boundary.

#### Alternative flows

- Locally generated failures (pacing rejection, breaker open, routing exhaustion) skip the
  adapter layer and are minted by the router with origin marked `local`.

#### Edge cases

- Provider returns a documented error body with an unknown code value inside a known
  schema: E-PROV-012 with the unknown value preserved.
- Failure during error-body parsing itself: E-PROV-008; the parse failure is the cause.
- Multiple simultaneous classifications (timeout racing a 5xx): the first terminal outcome
  observed by the router wins; the loser is discarded (FR-PROV-040 race rule).

#### Inputs

Wire failures; local rejection conditions; mapping rules.

#### Outputs

E-PROV `PortError` values; `provider.request.failed` events; exit codes.

#### States

None; feeds breaker windows via the matrix.

#### Errors

The catalog below is this requirement's error surface in full.

#### Constraints

Codes are stable and never reused (Volume 0 chapter 03); the matrix is normative for the
router; raw excerpts in safe detail are bounded and redacted.

#### Security

Raw provider payloads may contain reflected request content: excerpts in safe detail MUST
pass Volume 9 redaction (no credentials, no message content beyond the bounded excerpt of
the *error* body, no environment values). User messages never include raw bodies.

#### Observability

Every normalized failure emits `provider.request.failed` (or `provider.stream.interrupted`
per chapter 03) carrying code, category, attempt, and correlation ULIDs; per-code frequency
metrics feed Volume 12 dashboards.

#### Performance

Mapping is table-driven and constant-time; no budget impact beyond parsing the error body.

#### Compatibility

New codes are additive (minor); meanings never change (SM-20; ADR-016 reversal rules).

#### Acceptance criteria

- Given the fault-injection corpus (every origin condition in the mapping table), when
  failures surface, then each consumer-visible error is an E-PROV code with a complete
  envelope and correct exit-code mapping.
- Given an undocumented provider response shape, when normalization runs, then E-PROV-008
  results with the raw identity in safe detail and nothing unredacted.
- Negative case: no test in the corpus observes a raw `net/http`, driver, or SDK error type
  crossing the port (leak check).
- Permission case: permission denials on the provider path surface in the E-SEC family
  (Volume 9), not as E-PROV codes — the suite asserts the separation.
- Observability case: per-code metrics and failure events reconcile 1:1 with injected
  faults.

#### Verification method

Fault-injection suite over the mapping table; leak checks in contract tests; envelope
completeness linting against the catalog (Volume 13 check); exit-code integration tests
(Volume 8).

#### Traceability

PRD-002, PRD-006; ADR-016, ADR-061; FR-PROV-040..043; NFR-PROV-004.

## E-PROV error catalog

Common facts, stated once: all codes map to telemetry event `provider.request.failed`
version 1 unless the entry states otherwise; all safe context data includes provider slug,
model name (when resolved), method, attempt number, and correlation ULIDs; user messages
are templates — angle-bracket fields interpolate identity, never content. "Origin HTTP"
records the typical documented origin statuses; Andromeda itself exposes no HTTP surface
for these errors (headless mode is JSON-RPC over the ADR-012 socket).

### E-PROV-001 — Provider unreachable

- Code: `E-PROV-001` (stable)
- Category: connectivity
- Severity: error
- User message: "Provider '<slug>' is unreachable."
- Technical message: "connection to <endpoint-host> failed: <transport detail>"
- Cause: DNS resolution, TCP connect, or TLS handshake failure; endpoint down or blocked.
- Safe context data: common set + endpoint host, transport stage.
- Recoverability: recoverable when connectivity returns.
- Retry policy: retryable with backoff (FR-PROV-040); breaker-eligible.
- Recommended action: check network/endpoint configuration; for local providers, verify the server is running.
- Exit code: 7
- HTTP mapping: origin — none (below HTTP).
- Telemetry event: `provider.request.failed`.
- Security implications: endpoint host appears in logs; no content or credentials involved.

### E-PROV-002 — Provider authentication rejected

- Code: `E-PROV-002` (stable)
- Category: authentication
- Severity: error
- User message: "Provider '<slug>' rejected the configured credentials."
- Technical message: "auth rejected (origin status <status>); credential '<label>', session state <state>"
- Cause: invalid, expired, revoked, or under-scoped credential; provider-side access change.
- Safe context data: common set + credential label, Authentication Session state; never material or fingerprints beyond Volume 9 display rules.
- Recoverability: recoverable via refresh, rotation, or re-authentication (chapters 07–08).
- Retry policy: one retry after a successful single-flight refresh; otherwise not retryable; never breaker-eligible.
- Recommended action: run the authentication flow for the provider; verify credential status.
- Exit code: 4
- HTTP mapping: origin 401/403.
- Telemetry event: `provider.request.failed`.
- Security implications: repeated occurrences may indicate credential compromise or revocation; surfaced to the audit trail via the Authentication Layer (chapter 08).

### E-PROV-003 — Provider rate limited

- Code: `E-PROV-003` (stable)
- Category: rate_limit
- Severity: warning
- User message: "Provider '<slug>' is rate limiting requests."
- Technical message: "rate limited (origin <local|provider>); delay signal <ms|none>"
- Cause: provider-side request/token rate ceiling, or local pacing bound reached.
- Safe context data: common set + origin (local/provider), honored delay ms.
- Recoverability: recoverable after the pacing window.
- Retry policy: retryable, delay-signal aware (capped); ratio-based breaker eligibility.
- Recommended action: reduce concurrency, configure local pacing, or raise provider limits.
- Exit code: 7
- HTTP mapping: origin 429.
- Telemetry event: `provider.request.failed`.
- Security implications: none beyond load-pattern visibility in logs.

### E-PROV-004 — Provider quota or billing exhausted

- Code: `E-PROV-004` (stable)
- Category: quota
- Severity: error
- User message: "Provider '<slug>' reports quota or billing limits exhausted."
- Technical message: "quota/billing exhausted per documented error '<provider error id>'"
- Cause: account-level spend or usage cap reached at the provider.
- Safe context data: common set + provider error identity.
- Recoverability: not recoverable by retry; requires account action.
- Retry policy: not retryable; fallback class `quota_exhausted`.
- Recommended action: review provider account limits/billing; route to another provider if configured.
- Exit code: 7
- HTTP mapping: origin per provider documentation (commonly 402/429-variant bodies).
- Telemetry event: `provider.request.failed`.
- Security implications: account state appears in logs at identity level only.

### E-PROV-005 — Model not available

- Code: `E-PROV-005` (stable)
- Category: request
- Severity: error
- User message: "Model '<model>' is not available from provider '<slug>'."
- Technical message: "unknown/retired model per provider response; registry row deprecated=<bool>"
- Cause: model name unknown to the provider, retired, or not enabled for the account.
- Safe context data: common set + registry deprecation flag.
- Recoverability: recoverable by selecting another model.
- Retry policy: not retryable; no silent substitution (FR-PROV-013).
- Recommended action: refresh discovery; select an offered model.
- Exit code: 7
- HTTP mapping: origin 404 or documented equivalent.
- Telemetry event: `provider.request.failed`.
- Security implications: none.

### E-PROV-006 — Capability unavailable

- Code: `E-PROV-006` (stable)
- Category: capability
- Severity: error
- User message: "'<capability>' is not available for model '<model>' on provider '<slug>'."
- Technical message: "negotiation failed: required <capability> absent from effective set (provenance snapshot attached)"
- Cause: required capability absent and degradation resolved to refuse (chapter 02).
- Safe context data: common set + capability name, effective-set provenance summary.
- Recoverability: recoverable via another model/provider or an enabled substitute.
- Retry policy: not retryable; may trigger `capability_gap` reroute (chapter 02).
- Recommended action: choose a capable model, enable a documented substitute, or adjust the workflow's requirements.
- Exit code: 7
- HTTP mapping: not applicable (local decision).
- Telemetry event: `provider.request.failed`.
- Security implications: none.

### E-PROV-007 — Request rejected as invalid

- Code: `E-PROV-007` (stable)
- Category: request
- Severity: error
- User message: "Provider '<slug>' rejected the request as invalid."
- Technical message: "origin 400-class rejection: <documented validation detail>; or pre-dispatch: unsupported construct <name>"
- Cause: request violates the provider's documented constraints, or pre-dispatch validation found constructs outside the adapter's declared support (chapter 03).
- Safe context data: common set + offending construct/parameter name (never values).
- Recoverability: recoverable by changing the request.
- Retry policy: not retryable.
- Recommended action: inspect the named construct; report an adapter defect if the request used only documented surface.
- Exit code: 7
- HTTP mapping: origin 400/422.
- Telemetry event: `provider.request.failed`.
- Security implications: parameter names in logs; values are never included.

### E-PROV-008 — Provider response malformed

- Code: `E-PROV-008` (stable)
- Category: response
- Severity: error
- User message: "Provider '<slug>' returned a response Andromeda could not interpret."
- Technical message: "response failed schema mapping at <stage>; raw identity <type/code>; excerpt bounded 2048B"
- Cause: undocumented response shape, API drift, or intermediary corruption.
- Safe context data: common set + mapping stage, bounded redacted excerpt.
- Recoverability: possibly transient; often adapter-drift (RISK-PROV-001).
- Retry policy: one retry; breaker-eligible.
- Recommended action: retry; if persistent, update Andromeda or report the adapter drift.
- Exit code: 7
- HTTP mapping: origin 200 with unparseable body, or any status with undocumented body.
- Telemetry event: `provider.request.failed`.
- Security implications: excerpts pass Volume 9 redaction before logging.

### E-PROV-009 — Stream interrupted

- Code: `E-PROV-009` (stable)
- Category: stream
- Severity: error
- User message: "The response stream from provider '<slug>' was interrupted."
- Technical message: "stream cut after <n> events; delivered-usage snapshot <fields>; transport detail <detail>"
- Cause: transport failure mid-stream after delivery began.
- Safe context data: common set + delivered event count, delivered-usage snapshot.
- Recoverability: turn-level decision belongs to the Agent Engine (Volume 4).
- Retry policy: not retryable at the layer (post-delivery, chapter 03 rule 5); breaker-eligible.
- Recommended action: resume or re-issue the turn; check provider status if recurrent.
- Exit code: 7
- HTTP mapping: origin — connection terminated mid-body.
- Telemetry event: `provider.stream.interrupted`.
- Security implications: none beyond timing metadata.

### E-PROV-010 — Provider request timeout

- Code: `E-PROV-010` (stable)
- Category: timeout
- Severity: error
- User message: "The request to provider '<slug>' timed out (<class>)."
- Technical message: "timeout class <class> at <ms> ms; attempt <n>"
- Cause: expiry of a chapter 05 timeout class.
- Safe context data: common set + class, configured ms.
- Recoverability: recoverable; often load-related.
- Retry policy: retryable with backoff; breaker-eligible.
- Recommended action: retry; raise class limits for slow local models; check provider status.
- Exit code: 8
- HTTP mapping: origin — no response within deadline.
- Telemetry event: `provider.request.failed`.
- Security implications: none.

### E-PROV-011 — Request cancelled

- Code: `E-PROV-011` (stable)
- Category: cancellation
- Severity: info-level outcome recorded as error envelope for uniformity
- User message: "The request to provider '<slug>' was cancelled."
- Technical message: "context cancelled at <stage>; provider-side cancellation <declared|client-only>"
- Cause: user interrupt, run cancellation, or shutdown (FR-ARCH-004).
- Safe context data: common set + stage, cancellation reason.
- Recoverability: not an error to recover; the run's cancellation semantics apply (Volume 4).
- Retry policy: never retried.
- Recommended action: none.
- Exit code: 8
- HTTP mapping: origin — request aborted by client.
- Telemetry event: `provider.request.failed`.
- Security implications: where `cancellation` capability is absent, provider-side token spend may continue briefly; recorded per chapter 02 degradation note.

### E-PROV-012 — Provider internal error

- Code: `E-PROV-012` (stable)
- Category: provider_internal
- Severity: error
- User message: "Provider '<slug>' reported an internal error."
- Technical message: "origin 5xx <status> / documented error '<provider error id>'"
- Cause: provider-side failure, or documented error condition with no more specific class.
- Safe context data: common set + origin status, provider error identity.
- Recoverability: usually transient.
- Retry policy: retryable with backoff; breaker-eligible; fallback class `internal_error`.
- Recommended action: retry; consult provider status; use fallback chains where configured.
- Exit code: 7
- HTTP mapping: origin 5xx.
- Telemetry event: `provider.request.failed`.
- Security implications: none.

### E-PROV-013 — Content refused by provider policy

- Code: `E-PROV-013` (stable)
- Category: content_policy
- Severity: error
- User message: "Provider '<slug>' declined to process this content under its usage policies."
- Technical message: "documented policy refusal '<provider refusal id>'; finish reason content_filter where streamed"
- Cause: provider-side content policy applied to input or output.
- Safe context data: common set + refusal identity; never the content itself.
- Recoverability: content- or provider-level decision; not transient.
- Retry policy: not retryable; automatic fallback prohibited by default (a policy refusal on one vendor is not a routing signal; users may act manually).
- Recommended action: revise the content or choose an appropriate provider manually.
- Exit code: 7
- HTTP mapping: origin per provider documentation (400-class bodies or stream finish reason).
- Telemetry event: `provider.request.failed`.
- Security implications: refusal identity is logged; content is not.

### E-PROV-014 — Context window exceeded

- Code: `E-PROV-014` (stable)
- Category: request
- Severity: error
- User message: "The request exceeds model '<model>''s context window."
- Technical message: "documented context-length error; declared window <n> tokens; request estimate <n>"
- Cause: assembled request larger than the model's context window.
- Safe context data: common set + declared window, size figures.
- Recoverability: recoverable via context re-budgeting (Context Manager, Volume 7).
- Retry policy: not retryable unchanged.
- Recommended action: reduce context; the Context Manager's budgeting handles this automatically on the next assembly.
- Exit code: 7
- HTTP mapping: origin 400-class documented error.
- Telemetry event: `provider.request.failed`.
- Security implications: none (sizes only).

### E-PROV-015 — Circuit breaker open

- Code: `E-PROV-015` (stable)
- Category: connectivity
- Severity: warning
- User message: "Provider '<slug>' is temporarily excluded after repeated failures."
- Technical message: "breaker open (reopen count <n>, remaining <s>s); trip statistics attached"
- Cause: local breaker trip per chapter 05 guards.
- Safe context data: common set + trip statistics, remaining open duration.
- Recoverability: automatic (half-open probe) or manual verification.
- Retry policy: rejected locally until half-open; fallback class `breaker_open`.
- Recommended action: wait, verify the provider explicitly, or rely on configured fallback.
- Exit code: 7
- HTTP mapping: not applicable (locally generated; no wire traffic).
- Telemetry event: `provider.request.failed`.
- Security implications: none.

### E-PROV-016 — No eligible provider or model

- Code: `E-PROV-016` (stable)
- Category: routing
- Severity: error
- User message: "No configured provider can serve this request."
- Technical message: "routing/fallback exhausted; per-candidate reasons: <candidate:reason list>"
- Cause: all candidates excluded (capability, state, breaker, policy, permission, guards) or all fallback steps failed/skipped.
- Safe context data: common set + full per-candidate reason chain.
- Recoverability: recoverable via configuration (enable providers, adjust chains, grant permissions).
- Retry policy: not retryable.
- Recommended action: inspect the reason chain in the error output; adjust provider configuration or grants.
- Exit code: 7
- HTTP mapping: not applicable (locally generated).
- Telemetry event: `provider.request.failed`.
- Security implications: reason chains may reveal policy structure in logs; policy identifiers only, per Volume 9 rules.

### E-PROV-017 — Structured output validation failed

- Code: `E-PROV-017` (stable)
- Category: validation
- Severity: error
- User message: "The model's output did not match the required schema."
- Technical message: "local validation failed (mode <native|tool_call|prompted>): <first N findings>; raw output preserved as data"
- Cause: provider output violating the declared response format (chapter 03).
- Safe context data: common set + mode, finding paths (JSON Pointers), retry count.
- Recoverability: bounded re-request per `validation_retries`.
- Retry policy: caller-driven re-request within the router-enforced ceiling; not transport-retryable.
- Recommended action: simplify the schema, use a capable model, or enable a stronger mode.
- Exit code: 7
- HTTP mapping: not applicable (local validation).
- Telemetry event: `provider.request.failed`.
- Security implications: finding paths only in logs; raw output stays in run records under Volume 9 rules.

### E-PROV-018 — Tool call unparseable

- Code: `E-PROV-018` (stable)
- Category: validation
- Severity: error
- User message: "The model produced a tool call Andromeda could not interpret."
- Technical message: "malformed tool call (reason <undeclared name|invalid JSON|structure>); returned to agent as data"
- Cause: model output presenting as a tool call but failing the chapter 03 normal form.
- Safe context data: common set + malformation reason, tool name when parseable.
- Recoverability: agent-level repair/retry (Volume 4).
- Retry policy: not transport-retryable; never dispatched to the Tool Runtime.
- Recommended action: none for users in the common case; recurring occurrences suggest a capability misdeclaration (RISK-PROV-002).
- Exit code: 7
- HTTP mapping: not applicable.
- Telemetry event: `provider.request.failed`.
- Security implications: malformed arguments are untrusted model output and are stored only in run records under redaction rules — never executed.

### E-PROV-019 — Adapter declaration invalid

- Code: `E-PROV-019` (stable)
- Category: configuration
- Severity: critical
- User message: "Adapter '<adapter-id>' cannot be registered: its declaration is invalid."
- Technical message: "declaration validation failed: <all findings>; contract version <declared> vs running <running>"
- Cause: schema violation, unknown capability values, invalid error-map targets, or contract-major mismatch (FR-PROV-002).
- Safe context data: adapter ID, adapter version, full findings list.
- Recoverability: requires a corrected adapter version or an Andromeda update.
- Retry policy: not retryable.
- Recommended action: update the adapter or Andromeda; report to the adapter maintainer.
- Exit code: 3
- HTTP mapping: not applicable.
- Telemetry event: `provider.adapter.rejected`.
- Security implications: refused adapters never receive requests or credentials.

## Registry observability

| Event | Version | Producer | Payload (summary) | Correlation |
|---|---|---|---|---|
| `provider.adapter.registered` | 1 | Provider Layer | adapter ID, adapter version, contract version | — |
| `provider.adapter.rejected` | 1 | Provider Layer | adapter ID, adapter version, findings summary | — |

Envelope per Volume 10.

### NFR-PROV-004 — Error normalization coverage

- Category: Reliability
- Priority: P0
- Phase: MVP
- Metric: (a) Fraction of provider-path failures in fault-injection and live suite runs surfaced as E-PROV codes with complete ADR-016 envelopes; (b) count of raw transport/driver/SDK error types observed crossing ProviderPort
- Target: (a) 100%; (b) 0
- Minimum threshold: same as target — a single leak is a contract defect (Volume 3 port rule 2)
- Measurement method: Leak-detection assertions in every contract and fault-injection test; envelope completeness check against the catalog; per-code reconciliation of injected faults versus surfaced errors
- Test environment: CI on Tier 1 platforms; fault-injection corpus covering every mapping-table origin condition
- Measurement frequency: Every merge (contract tests) and every release (full corpus)
- Owner: Provider Layer (Volume 5)
- Dependencies: FR-PROV-050; ADR-061
- Risks: RISK-PROV-001
- Acceptance criteria: The release fault-injection report shows every injected condition surfacing as its matrix-specified code, zero envelope fields empty, and zero leak-check failures.
