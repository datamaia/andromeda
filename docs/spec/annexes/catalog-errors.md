# Annex — Consolidated Error Catalog

**Status:** Consolidated (Phase C). This annex is the corpus-wide index of every error code
minted in Volumes 3–14, aggregated from the volume registers and the defining chapters. It is
a *reference view*: the full fourteen-field envelope of each error — user message, technical
message, cause, safe context data, recoverability, recommended action, HTTP mapping,
telemetry event, and security implications — lives in the linked defining chapter, which is
normative. This annex mints nothing and renames nothing. Scheme owner: Volume 0, chapter 03;
decision record: [ADR-016](adr/ADR-016.md).

## The ADR-016 envelope, summarized

Error identity is `E-<AREA>-NNN`: `<AREA>` follows the Volume 0 chapter 03 ownership table
(the same area ownership as requirements) and `NNN` is a per-area sequence that is never
reused or renumbered. Every defined error declares a fixed envelope of fourteen fields:

1. **Stable code** — the `E-<AREA>-NNN` identity, stable across releases and rewording.
2. **Category** — the failure class within its area (vocabulary per defining chapter).
3. **Severity** — informational/warning/error/critical/fatal classification.
4. **User message** — human-facing text; never the error's identity key.
5. **Technical message** — diagnostic detail for logs and support.
6. **Cause** — what produces the condition.
7. **Safe context data** — the fields that may appear in logs and events (redaction is a
   definition-time property).
8. **Recoverability** — whether and how the condition can be recovered from.
9. **Retry policy** — whether automated retry is permitted, and under what bounds.
10. **Recommended action** — what the user or caller does next.
11. **Exit-code mapping** — exactly one code from the closed scheme below.
12. **HTTP mapping** — where an IPC/HTTP surface applies; otherwise "not applicable".
13. **Telemetry event** — the event name emitted when the error occurs.
14. **Security implications** — what the error means for the security posture.

### Exit-code scheme (closed; new codes require a superseding ADR)

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error |
| 2 | Usage error (invalid arguments or flags) |
| 3 | Configuration error |
| 4 | Authentication error |
| 5 | Permission denied (by Andromeda's permission model) |
| 6 | Tool execution failure |
| 7 | Provider failure |
| 8 | Timeout or cancellation |
| 9 | Integrity error (corrupted state, failed migration) |

## Reading the tables

One table per area, in volume order. Columns: the stable code; the error's name (its
one-line meaning); the envelope's category and severity exactly as the defining chapter
declares them (vocabularies are per-area and are not normalized here); the exit-code
mapping, including its qualifications; the retry policy, abridged where long (the defining
chapter's full text governs); and the defining chapter. Codes absent from a sequence (none
today) would be permanent gaps per Volume 0 chapter 03.


## E-ARCH — Architecture and runtime composition (Volume 3)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-ARCH-001 | Component wiring failure | Startup | Fatal | 3 when configuration-caused; otherwise 1 | none automatic | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| E-ARCH-002 | Port contract violation | Internal defect | Error | 1 | not retryable (deterministic defect) | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| E-ARCH-003 | IPC endpoint unavailable | Environment | Error | 1 | single retry after stale-socket cleanup; otherwise none | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| E-ARCH-004 | IPC protocol version unsupported | Compatibility | Error | not applicable (server-side rejection); clients map to 2 | none | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| E-ARCH-005 | Task submission rejected | Capacity | Warning (component-level); Error when surfaced to a … | 1 if it fails a foreground command; usually not surfaced | submitter-defined (background work typically defers and retries) | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| E-ARCH-006 | Forced shutdown | Lifecycle | Warning | 8 | not applicable | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |
| E-ARCH-007 | Recovery reconciliation failure | Integrity | Fatal | 9 | recovery re-runs on next start (idempotent) | [Vol 3 ch 08](../volume-03-architecture/08-processes-concurrency-ipc.md) |

## E-PORT — Portability and the platform abstraction layer (Volume 3)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-PORT-001 | Unsupported platform | Environment | Fatal | 3 | none | [Vol 3 ch 07](../volume-03-architecture/07-platform-abstraction-layer.md) |
| E-PORT-002 | Platform capability unavailable | Environment | Error (or Warning when the surface's degradation … | 3 when fatal to the invoked command; otherwise reported and degraded | none automatic; re-probe on restart | [Vol 3 ch 07](../volume-03-architecture/07-platform-abstraction-layer.md) |
| E-PORT-003 | Credential store backend unavailable | Environment | Error | 3 (environment); authentication flows blocked by it report through … | none automatic | [Vol 3 ch 07](../volume-03-architecture/07-platform-abstraction-layer.md) |

## E-AGT — Agent engine, planner, execution engine, prompt engine (Volume 4)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-AGT-001 | Not a workspace | Environment | Error | 1 | none automatic | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| E-AGT-002 | Workspace exclusivity conflict | Environment | Error | 1 | single retry after stale-lock verification; otherwise none | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| E-AGT-003 | Session or run not resumable | State | Error | 1; 9 when caused by integrity validation failure | none for terminal targets; single retry after conflict for revision mismatches | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| E-AGT-004 | Run budget exhausted | Policy | Warning (expected policy outcome) | 8 (policy cancellation class) | none automatic (deliberate policy stop) | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| E-AGT-005 | Iteration limit reached | Policy | Warning | 8 | none automatic | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| E-AGT-006 | Agent profile resolution failure | Configuration | Error | 3 | none | [Vol 4 ch 01](../volume-04-agent-runtime/01-agent-engine.md) |
| E-AGT-007 | Plan validation failed | Model output | Error | 1 | attempts are the retry mechanism; no outer automatic retry | [Vol 4 ch 02](../volume-04-agent-runtime/02-planner.md) |
| E-AGT-008 | Task dependencies unsatisfiable | State | Error | 1 | none (structural condition, not transient) | [Vol 4 ch 03](../volume-04-agent-runtime/03-execution-engine.md) |
| E-AGT-009 | Prompt template resolution failure | Configuration | Error | 3 | none | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |
| E-AGT-010 | Prompt render failure | Internal defect / configuration | Error | 1; 3 when caused by template/profile configuration | not retryable (deterministic failure) | [Vol 4 ch 04](../volume-04-agent-runtime/04-prompt-engine.md) |
| E-AGT-011 | Illegal state transition | Internal defect | Error | 1; 9 when raised by replay/integrity validation of stored records | not retryable (deterministic) | [Vol 4 ch 05](../volume-04-agent-runtime/05-core-state-machines.md) |

## E-WF — Workflows and specification-driven development (Volume 4)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-WF-001 | Workflow definition invalid | Validation | Error | 3 | none — deterministic until the source changes | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| E-WF-002 | Workflow definition version unsupported | Compatibility | Error | 3 | none | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| E-WF-003 | Workflow not found | Usage | Error | 2 | none | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| E-WF-004 | Workflow inputs invalid | Validation | Error | 2 | none | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| E-WF-005 | Workflow requirements unsatisfied | Environment | Error | 1 | none automatic; re-instantiation re-evaluates | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| E-WF-006 | Workflow gate denied | Permission | Warning (run-level outcome, not a defect) | 5 | none — denials are decisions, not transient failures | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| E-WF-007 | Workflow gate expired | Timeout | Warning | 8 | none automatic; `on_expired` routing governs | [Vol 4 ch 06](../volume-04-agent-runtime/06-workflow-engine-and-sdd.md) |
| E-WF-008 | Workflow step failed | Execution | Error | 1 | per-step `retry` declaration, then `on_failed` routing; no engine-level retry beyond … | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| E-WF-009 | Workflow step timed out | Timeout | Error | 8 | per-step `retry` applies to step timeouts; run-deadline expiry is not retried | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| E-WF-010 | Workflow run cancelled | Cancellation | Info | 8 | none | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| E-WF-011 | Workflow compensation failed | Execution | Error | 1 | rollback halts at first failure; re-invocation resumes from the failed compensation | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| E-WF-012 | Workflow run state integrity failure | Integrity | Fatal (for the affected run) | 9 | none | [Vol 4 ch 07](../volume-04-agent-runtime/07-workflow-run-state-machine.md) |
| E-WF-013 | Skill application failed | Environment | Error | 1 | none automatic; the next run re-resolves | [Vol 4 ch 08](../volume-04-agent-runtime/08-skill-engine-runtime.md) |

## E-PROV — Providers, models, capabilities (Volume 5)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-PROV-001 | Provider unreachable | connectivity | error | 7 | retryable with backoff (FR-PROV-040); breaker-eligible | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-002 | Provider authentication rejected | authentication | error | 4 | one retry after a successful single-flight refresh; otherwise not retryable; never … | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-003 | Provider rate limited | rate_limit | warning | 7 | retryable, delay-signal aware (capped); ratio-based breaker eligibility | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-004 | Provider quota or billing exhausted | quota | error | 7 | not retryable; fallback class `quota_exhausted | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-005 | Model not available | request | error | 7 | not retryable; no silent substitution (FR-PROV-013) | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-006 | Capability unavailable | capability | error | 7 | not retryable; may trigger `capability_gap` reroute (chapter 02) | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-007 | Request rejected as invalid | request | error | 7 | not retryable | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-008 | Provider response malformed | response | error | 7 | one retry; breaker-eligible | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-009 | Stream interrupted | stream | error | 7 | not retryable at the layer (post-delivery, chapter 03 rule 5); breaker-eligible | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-010 | Provider request timeout | timeout | error | 8 | retryable with backoff; breaker-eligible | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-011 | Request cancelled | cancellation | info-level outcome recorded as error envelope for … | 8 | never retried | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-012 | Provider internal error | provider_internal | error | 7 | retryable with backoff; breaker-eligible; fallback class `internal_error | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-013 | Content refused by provider policy | content_policy | error | 7 | not retryable; automatic fallback prohibited by default (a policy refusal on one … | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-014 | Context window exceeded | request | error | 7 | not retryable unchanged | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-015 | Circuit breaker open | connectivity | warning | 7 | rejected locally until half-open; fallback class `breaker_open | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-016 | No eligible provider or model | routing | error | 7 | not retryable | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-017 | Structured output validation failed | validation | error | 7 | caller-driven re-request within the router-enforced ceiling; not transport-retryable | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-018 | Tool call unparseable | validation | error | 7 | not transport-retryable; never dispatched to the Tool Runtime | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |
| E-PROV-019 | Adapter declaration invalid | configuration | critical | 3 | not retryable | [Vol 5 ch 06](../volume-05-providers-and-auth/06-error-normalization.md) |

## E-AUTH — Authentication and credentials (Volume 5)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-AUTH-001 | Credential not found | authentication | error | 4 | not retryable automatically | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-002 | Authentication rejected by provider | authentication | error | 4 | not retryable with the same material; rotation or re-intake required | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-003 | Credential expired and renewal failed | authentication | error | 4 | renewal already retried per FR-AUTH-010; the surfacing error is not auto-retried | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-004 | Authorization flow denied | authentication | error | 4 | not auto-retried (human decision) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-005 | Authorization flow timed out | timeout | error | 8 | not auto-retried; a rerun mints fresh codes | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-006 | Secret store unavailable | environment | error | 3 | retryable after the environment is fixed; no automatic retry | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-007 | Prohibited authentication mechanism requested | validation | error | 3 | never retried | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-008 | Credential rotation failed | authentication | error | 4 | not auto-retried; rotation is operator-driven | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-009 | Provider-side revocation failed | authentication | warning | 4 (command exits nonzero so the partial outcome is visible) | one automatic retry on transient classes; then surfaced | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-010 | Authentication profile not found or ambiguous | configuration | error | 3 | never retried | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |
| E-AUTH-011 | Proxy authentication failed | environment | error | 4 | not auto-retried (repeated failures can lock enterprise accounts) | [Vol 5 ch 08](../volume-05-providers-and-auth/08-credential-lifecycle.md) |

## E-TOOL — Tools and the Tool Runtime (Volume 6)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-TOOL-001 | Tool not found | validation | error | 6 (Volume 8 maps user-typed unknown names on the CLI surface to 2 as … | not retryable (deterministic until registry changes) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-002 | Tool registration rejected | validation | error | 6 | not retryable without change | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-003 | Input validation failed | validation | error | 6 | not retryable unchanged | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-004 | Output validation failed | execution | error | 6 | retryable only per ADR-072 conditions | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-005 | Invocation denied | permission | warning | 5 | never retried automatically (ADR-072 exclusion) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-006 | Tool execution failed | execution | error | 6 | automatic only per ADR-072; otherwise caller-decided | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-007 | Invocation timed out | timeout | error | 8 | retryable automatically only for idempotent tools (ADR-072) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-008 | Invocation cancelled | cancellation | info | 8 | never automatic | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-009 | Resource limit exceeded | resource | error | 6 | not retryable unchanged (deterministic against the same input and caps) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-010 | Tool origin unavailable | availability | error | 6 | retryable per ADR-072 for idempotent tools; fails fast, never queues on a dead … | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-011 | Concurrency capacity exhausted | resource | error | 6 | retryable with backoff (idempotency irrelevant — nothing executed) | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |
| E-TOOL-012 | Execution interrupted | integrity | error | 6 | never automatic (side-effect state unknown, even for declared-idempotent tools — … | [Vol 6 ch 02](../volume-06-tools-mcp-skills-plugins/02-tool-lifecycle-permissions-trust.md) |

## E-MCP — Model Context Protocol support (Volume 6)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-MCP-001 | MCP connection establishment failed | Connectivity | Error | 6 when it fails a tool invocation path; 1 for standalone connect … | automatic only within the reconnection policy after a prior `ready`; manual retry … | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| E-MCP-002 | MCP protocol negotiation failed | Compatibility | Error | 6 on invocation paths; 1 standalone | none (deterministic) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| E-MCP-003 | MCP request timed out | Timeout | Error | 8 | no automatic retry of the same request; idempotent listings retried once during … | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| E-MCP-004 | MCP connection not ready | State | Error | 6 | caller-driven after reconnection; no queueing of requests | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| E-MCP-005 | MCP authorization failed | Authorization | Error | 4 | none automatic | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| E-MCP-006 | MCP protocol violation by server | Protocol | Error | 6 | none (deterministic against the same response) | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| E-MCP-007 | MCP server registration invalid | Configuration | Error | 3 | none | [Vol 6 ch 05](../volume-06-tools-mcp-skills-plugins/05-mcp-client-and-runtime.md) |
| E-MCP-008 | MCP tool descriptor drift detected | Security | Warning (exposure suspended; nothing invoked) | 5 when it blocks a requested invocation; otherwise not surfaced as … | none — deterministic until re-approval | [Vol 6 ch 06](../volume-06-tools-mcp-skills-plugins/06-mcp-security-and-conformance.md) |

## E-SKILL — Skills (Volume 6)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-SKILL-001 | Skill manifest invalid | Validation | Error | 3 | none | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| E-SKILL-002 | Skill content integrity mismatch | Integrity | Error | 9 | none | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| E-SKILL-003 | Skill requirements unsatisfiable | Dependency | Error | 1 | none automatic; re-activation after remediation | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| E-SKILL-004 | Skill resolution failed | Dependency | Error | 1 | none (deterministic) | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| E-SKILL-005 | Skill composition conflict | Conflict | Error | 1 | none | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |
| E-SKILL-006 | Executable content rejected in skill | Validation | Error | 3 | none | [Vol 6 ch 07](../volume-06-tools-mcp-skills-plugins/07-skill-format-and-system.md) |

## E-PLUG — Plugins and extension packages (Volume 6)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-PLUG-001 | Plugin spawn failed | Execution | Error | 6 when failing an invocation path; 1 for lifecycle commands | restart policy applies only for transient classes; deterministic classes rest in … | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| E-PLUG-002 | ARP version negotiation failed | Compatibility | Error | 1 | none (deterministic; excluded from restart policy) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| E-PLUG-003 | ARP protocol violation | Protocol | Error | 6 | restart policy applies except for deterministic handshake violations | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| E-PLUG-004 | Plugin request timed out | Timeout | Error | 8 | no automatic retry of the timed-out request | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| E-PLUG-005 | Plugin process crashed | Execution | Error | 6 | restart policy (FR-PLUG-004); invocations are not auto-retried | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| E-PLUG-006 | Plugin manifest invalid | Validation | Error | 3 | none (deterministic; excluded from restart policy) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| E-PLUG-007 | Undeclared surface registration attempted | Security | Error | 6 | none (deterministic) | [Vol 6 ch 08](../volume-06-tools-mcp-skills-plugins/08-plugin-runtime-and-arp.md) |
| E-PLUG-008 | Package source unavailable | Connectivity | Error | 1 | none automatic; discovery/resolution proceed over remaining sources where possible | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| E-PLUG-009 | Package resolution failed | Dependency | Error | 1 | none (deterministic against the same inputs) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| E-PLUG-010 | Package download failed | Connectivity | Error | 1 | one automatic re-attempt per archive within the same operation; further retries are … | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| E-PLUG-011 | Package verification failed | Integrity | Critical | 9 | none — deterministic against the same artifact | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| E-PLUG-012 | Package installation conflict | Conflict | Error | 1 | none (deterministic until state changes) | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |
| E-PLUG-013 | Package operation interrupted or cancelled | Cancellation | Warning | 8 | user-driven re-run; no automatic resumption | [Vol 6 ch 09](../volume-06-tools-mcp-skills-plugins/09-package-manager-supply-chain.md) |

## E-MEM — Memory (Volume 7)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-MEM-001 | Memory record validation failure | Validation | Error | 1 (2 when caused by invalid CLI arguments) | no automatic retry (deterministic failure) | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| E-MEM-002 | Redaction gate refusal | Security | Error | 5 when a confirmation was denied; otherwise 1 | no automatic retry | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| E-MEM-003 | Memory store unavailable | Storage | Error | 9 for integrity/migration failures; otherwise 1 | one automatic retry after lock-wait for transient locks; none for corruption | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| E-MEM-004 | Memory record not found | Not found | Warning | 1 | no automatic retry | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| E-MEM-005 | Memory maintenance or transfer failure | Execution | Error | 1 | automatic retry only for the pass scheduler (next scheduled pass); manual retry … | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |
| E-MEM-006 | Invalid retention policy | Configuration | Error | 3 | no automatic retry | [Vol 7 ch 02](../volume-07-memory-context-indexing/02-memory-lifecycle.md) |

## E-CTX — Context management (Volume 7)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-CTX-001 | Context budget infeasible | Validation | Error | 1 | no automatic retry (deterministic) | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| E-CTX-002 | Context source read failure | Execution | Warning (degradation) / Error (mandatory item) | 1 when a turn fails; otherwise none (recorded degradation) | one re-read attempt for transient filesystem errors; none for missing sources | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| E-CTX-003 | Context snapshot persistence failure | Storage | Error | 9 for integrity failures; otherwise 1 | one automatic retry on transient lock contention; then fail the turn | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |
| E-CTX-004 | Invalid pin or exclusion | Validation | Error | 2 (usage error at the CLI boundary) | no automatic retry | [Vol 7 ch 03](../volume-07-memory-context-indexing/03-context-manager.md) |

## E-IDX — Indexing (Volume 7)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-IDX-001 | Index build failure | Execution | Error | 1 | no automatic retry of full builds; the next workspace open or a manual rebuild … | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| E-IDX-002 | Embedding acquisition unavailable | External dependency | Error | 1 (7 when a CLI command's sole purpose was the provider-backed … | batch-level 3 attempts with exponential backoff (FR-IDX-003); update-path failures … | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| E-IDX-003 | Embedding space violation | Internal defect / contract | Error | 1 | none automatic (deterministic contract violation) | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| E-IDX-004 | Index cache corruption | Storage | Warning | none (never fails a command by itself; rebuild proceeds in background) | automatic — affected indexes transition per chapter 05 recovery and rebuild | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| E-IDX-005 | Index not queryable | Not found / state | Warning | 1 | none automatic at query time; consumers degrade (FR-CTX-001) | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| E-IDX-006 | Index scale limit reached | Resource limit | Warning (degrade) / Error (refuse) | none when degrading; 1 when configured to refuse | none automatic | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |
| E-IDX-007 | Concurrent index mutation rejected | Concurrency | Warning | 1 when a CLI command was rejected | automatic re-queue of update batches; manual retry for explicit builds | [Vol 7 ch 04](../volume-07-memory-context-indexing/04-indexing-engine.md) |

## E-CLI — CLI grammar and commands (Volume 8)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-CLI-001 | Unknown command or flag | Usage | Error | 2 | not retryable unchanged | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-002 | Invalid argument or flag value | Usage | Error | 2 | not retryable unchanged | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-003 | Confirmation required but unavailable | Usage | Error | 2 | not retryable unchanged | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-004 | Extension command unavailable | Environment | Error | 1 | single re-assembly retry, then fail | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-005 | Interactive terminal required | Environment | Error | 1 | not retryable unchanged | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-006 | Conflicting flags | Usage | Error | 2 | not retryable unchanged | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-007 | Input read failure | I/O | Error | 1 | not retryable automatically | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-008 | Output write failure | I/O | Error | 1 | none (writing more output to a broken stream is the failure) | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |
| E-CLI-009 | Invocation deadline exceeded | Timeout | Error | 8 | caller-driven; the CLI performs no automatic retry | [Vol 8 ch 02](../volume-08-cli-and-tui/02-cli-conventions.md) |

## E-TUI — Terminal UI (Volume 8)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-TUI-001 | TUI initialization failure | Environment | Error | 1 | not retryable automatically; user may re-invoke | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| E-TUI-002 | Terminal capabilities insufficient | Environment | Error | 1 | not retryable in the same environment | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| E-TUI-003 | Terminal state restoration failure | Environment | Warning | does not override the exit code of the terminating path; standalone … | restore sequence is attempted exactly once per step; failures do not block exit | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| E-TUI-004 | Invalid theme configuration | Configuration | Error | 3 (validation/CLI surfaces); in-TUI occurrence degrades to defaults … | re-validated on configuration change | [Vol 8 ch 08](../volume-08-cli-and-tui/08-theming-and-design-tokens.md) |
| E-TUI-005 | Invalid keymap configuration | Configuration | Error | 3 (when surfaced by validation commands); in-TUI occurrence degrades … | re-validated on configuration change | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| E-TUI-006 | Render pipeline failure | Internal | Critical | 1 | not retryable in-process; relaunch is the recovery | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| E-TUI-007 | Runtime unavailable to the interface | Internal | Error | 1 (only if the shell must terminate without recovery) | automatic reconnect with exponential backoff, 1 s initial, 30 s cap, indefinite … | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |
| E-TUI-008 | Event subscription overflow | Internal | Warning | none in-TUI; never terminates the shell (standalone mapping 1 is … | not applicable; drops are presentation-only by design | [Vol 8 ch 07](../volume-08-cli-and-tui/07-tui-architecture.md) |

## E-SEC — Security, permissions, sandbox, secrets, audit (Volume 9)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-SEC-001 | Action denied by permission model | Permission | Warning (expected policy outcome) | 5 | none automatic; gated retries re-evaluate (Volume 4) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| E-SEC-002 | Permission evaluation failure | Integrity | Error | 5 (the action is denied); 9 when caused by store integrity failure | single automatic retry for transient store errors; then fail closed | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| E-SEC-003 | Approval expired | Permission | Warning (expected policy outcome) | 5 | none automatic (consent is never retried automatically) | [Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md) |
| E-SEC-004 | Approval cancelled | Permission | Info (bookkeeping outcome) | 8 (cancellation class) when it terminates a command; otherwise not … | none | [Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md) |
| E-SEC-005 | Sandbox policy violation | Security | Error | 5 when surfaced as a denial before launch; 6 when a running tool is … | none automatic | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| E-SEC-006 | Isolation mechanism unavailable | Environment | Error | 6 | none automatic | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| E-SEC-007 | Sandbox teardown failure | Integrity | Critical | 6 | teardown retried once after kill escalation; then recorded for sweep | [Vol 9 ch 06](../volume-09-security/06-sandbox-specification.md) |
| E-SEC-008 | Secret backend unavailable | Environment | Error | 4 | none automatic | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| E-SEC-009 | Secret reference not found | State | Error | 4 | none automatic | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| E-SEC-010 | Secret access denied by OS store | Permission | Error | 4 | single retry after interactive unlock; none non-interactively | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| E-SEC-011 | Fallback passphrase unavailable | Environment | Error | 4 | none automatic | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| E-SEC-012 | Fallback store unreadable | Integrity | Error | 4; 9 when corruption is confirmed | passphrase re-prompt up to 3 per session; lock retry once; none for corruption | [Vol 9 ch 07](../volume-09-security/07-credential-and-secret-management.md) |
| E-SEC-013 | Audit chain integrity violation | Integrity | Critical | 9 | none — verification is deterministic | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |
| E-SEC-014 | Audit write failure | Integrity | Critical | 9 | one immediate writer-level retry; then fail the action | [Vol 9 ch 08](../volume-09-security/08-audit-and-incident-response.md) |

## E-CFG — Configuration and storage (Volume 10)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-CFG-001 | Configuration file unreadable | configuration | error | 3 | none (deterministic until the environment changes) | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-002 | Configuration file is not valid TOML | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-003 | Unknown configuration key | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-004 | Invalid type for configuration key | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-005 | Invalid value for configuration key | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-006 | Configuration keys conflict | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-007 | Include cannot be loaded | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-008 | Configuration profile cannot be resolved | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-009 | Environment variable mapping failure | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-010 | Configuration schema version unsupported | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-011 | Configuration migration failed | configuration | error | 3 | manual retry after addressing the cause | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-012 | Secret material detected in configuration | configuration | error | 3 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-013 | Deprecated configuration key in use | configuration | warning | 3 — applied only when a validation surface is invoked with warnings … | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-014 | Runtime override refused | configuration | error | 3 (surfaces only on explicit override requests; never at startup) | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-015 | Database schema is newer than this build | storage integrity | critical | 9 | none | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-016 | Database migration failed | storage integrity | critical | 9 | manual after cause resolution; the chain re-runs idempotently from the recorded … | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-017 | Database integrity check failed | storage integrity | critical | 9 | none (deterministic until repaired) | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-018 | Database backup could not be created | storage integrity | critical | 9 | manual after freeing space or fixing the backup location | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |
| E-CFG-019 | Workspace database is locked | concurrency | error | 1 | automatic bounded wait already applied; further retries are manual | [Vol 10 ch 02](../volume-10-config-storage-observability/02-config-errors-and-redaction.md) |

## E-OBS — Observability, logging, telemetry (Volume 10)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-OBS-001 | Malformed event envelope | internal defect | error (development), degraded-to-counter (release … | 1 when a CLI operation's primary purpose was the emission (not the … | never (the same emission fails identically) | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| E-OBS-002 | Unregistered event name | internal defect | error (development), degraded-to-counter (release) | 1 (same qualification as E-OBS-001) | never | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| E-OBS-003 | Event bus closed | lifecycle race | info | none (never terminal by itself) | never — producers MUST NOT retry against a closed bus | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| E-OBS-004 | Subscriber buffer overflow | backpressure | warning | none | not applicable (no operation to retry) | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| E-OBS-005 | Observability persistence failure | storage | error | 1; 9 when the cause is integrity/migration (ADR-016 scheme) | one immediate retry for transient `SQLITE_BUSY`-class causes; no retry for integrity … | [Vol 10 ch 04](../volume-10-config-storage-observability/04-events-and-envelope.md) |
| E-OBS-006 | Unregistered metric emission | internal defect | error (development builds), degraded-to-counter … | 1 in development-build test contexts only; none in release operation | never | [Vol 10 ch 05](../volume-10-config-storage-observability/05-traces-metrics-costs.md) |
| E-OBS-007 | Log sink failure | storage/degradation | warning | none (logging failure never terminates or fails an operation); 1 only … | automatic re-probe (60 s interval); the failing append itself is not retried | [Vol 10 ch 03](../volume-10-config-storage-observability/03-logging.md) |
| E-OBS-008 | Telemetry export failure | network/export | warning | none in normal operation (asynchronous); 1 when a direct CLI-invoked … | capped exponential backoff (30 s base, ×2, 1 h cap); partial rejections not retried | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |
| E-OBS-009 | Telemetry consent violation | configuration/policy | error | 3 (configuration error) when surfaced by a CLI operation | never automatic — resolution requires a human decision by design | [Vol 10 ch 06](../volume-10-config-storage-observability/06-telemetry-and-consent.md) |

## E-GIT — Git engine and hosting integration (Volume 11)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-GIT-001 | Git binary not found | Environment | Error | 3 | none automatic; re-resolved on configuration change | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-002 | Git version below minimum | Environment | Error | 3 | none automatic; re-checked on binary change | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-003 | Not a Git repository | Validation | Error | 1 | none | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-004 | Git subprocess failed | Execution | Error | 1 | not retryable automatically (repository state may have changed) | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-005 | Unparseable git output | Integrity | Error | 1 | none (deterministic) | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-006 | Operation stopped on conflicts | Conflict | Warning | 1 | not applicable (interactive resolution state) | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-007 | Destructive operation refused without confirmation | Policy | Warning | 5 | none automatic — automation MUST NOT auto-confirm | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-008 | Git mutation permission denied | Permission | Warning | 5 | none automatic | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-009 | Remote operation failed | Network | Error | 1 (4 when the failure class is authentication) | retryable for connectivity classes with caller backoff; not retryable for auth … | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-010 | Git operation timed out or was cancelled | Timeout | Warning | 8 | caller MAY retry read operations; mutations require state inspection first | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-011 | Repository state conflict | Validation | Warning | 1 | none automatic | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-012 | Protected branch refusal | Policy | Warning | 5 | none | [Vol 11 ch 01](../volume-11-git-and-github/01-git-engine.md) |
| E-GIT-013 | Hosting service request failed | External service | Error | 1 | retry with backoff for 5xx and network failures; never for 4xx other than 429 (see … | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| E-GIT-014 | Hosting authentication failed | Authentication | Error | 4 | none automatic | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |
| E-GIT-015 | Hosting rate limited | External service | Warning | 1 | single retry after the reported reset; otherwise surface to the caller with the … | [Vol 11 ch 02](../volume-11-git-and-github/02-github-gitlab-product-integrations.md) |

## E-GH — Development-process validators (CI-side) (Volume 11)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-GH-001 | Commit message policy violation | Process validation | Error | 1 (validator process) | not applicable (deterministic validation) | [Vol 11 ch 04](../volume-11-git-and-github/04-pull-requests.md) |
| E-GH-002 | Branch naming violation | Process validation | Error | 1 (validator process) | not applicable | [Vol 11 ch 07](../volume-11-git-and-github/07-traceability-automation.md) |
| E-GH-003 | Missing or unresolvable linkage | Process validation | Error | 1 (validator process) | not applicable | [Vol 11 ch 07](../volume-11-git-and-github/07-traceability-automation.md) |
| E-GH-004 | Traceability chain validation failed | Process validation | Fatal (blocks release publication) | 1 (validator process; release pipeline fails) | re-run after remediation; the generator is deterministic | [Vol 11 ch 07](../volume-11-git-and-github/07-traceability-automation.md) |
| E-GH-005 | Pull request size limit exceeded | Process validation | Error | 1 (validator process) | not applicable | [Vol 11 ch 04](../volume-11-git-and-github/04-pull-requests.md) |
| E-GH-006 | AI provenance label missing or inconsistent | Process validation | Error | 1 (validator process) | not applicable | [Vol 11 ch 04](../volume-11-git-and-github/04-pull-requests.md) |

## E-PERF — Performance and reliability (Volume 12)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-PERF-001 | Performance budget exceeded | Diagnostic | Warning | none (never fails an operation) | not applicable | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |
| E-PERF-002 | Operational limit exceeded | Capacity | Error | 1 when it fails a foreground command; 6 when surfaced as a tool … | not automatically retried; deterministic until inputs or configuration change | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |
| E-PERF-003 | Resource exhaustion refusal | Capacity | Error | 1 when it refuses a foreground command | not automatically retried; the refusal repeats until the mode exits | [Vol 12 ch 02](../volume-12-performance-and-reliability/02-reliability-and-degradation.md) |
| E-PERF-004 | Benchmark regression gate failure | Quality gate | Error | 1 (benchmark harness process); not applicable at the product CLI | re-run permitted only to confirm reproducibility; two consecutive failing runs … | [Vol 12 ch 03](../volume-12-performance-and-reliability/03-benchmarks-and-operational-limits.md) |

## E-TEST — Testing and quality (Volume 13)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-TEST-001 | Fixture integrity failure | Test infrastructure | Error | 9 (integrity error) when surfaced by harness tooling | No retry; regenerate from the generator or restore from version control | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| E-TEST-002 | Recorded interaction replay divergence | Test infrastructure | Error | 1 (general error) when surfaced by harness tooling | No retry; re-record via the ADR-176 scheduled path after confirming intent | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| E-TEST-003 | Scenario script rejected | Test infrastructure | Error | 3 (configuration error) when surfaced by harness tooling | No retry; fix the script | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| E-TEST-004 | Hermeticity violation | Test infrastructure | Critical | 1 (general error) when surfaced by harness tooling | No retry — the violation is the finding (SM-05 binds at MVP exit) | [Vol 13 ch 03](../volume-13-testing-and-quality/03-fixtures-fakes-determinism.md) |
| E-TEST-005 | Qualification evidence incomplete | Release qualification | Critical | 9 (integrity error) | Re-run qualification stages; no automatic retry of publication | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |
| E-TEST-006 | Gate evaluation failure | Test infrastructure | Error | 1 (general error) | One automatic re-run permitted (evaluation is read-only over recorded results); … | [Vol 13 ch 04](../volume-13-testing-and-quality/04-release-qualification-and-gates.md) |

## E-REL — Releases, distribution, updates (Volume 14)

| Code | Name (one-line meaning) | Category | Severity | Exit code | Retry | Defined in |
|---|---|---|---|---|---|---|
| E-REL-001 | Update check failed | Network | Warning (scheduled) / Error (explicit command) | 1 | scheduled checks retry next interval; explicit commands do not auto-retry | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-002 | Release metadata invalid | Validation | Error | 1 | none automatic (deterministic until source changes) | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-003 | Artifact download failed | Network | Error | 1 | up to 3 automatic resume attempts with exponential backoff within … | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-004 | Artifact verification failed | Security | Critical | 9 | none automatic — verification failures never auto-retry | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-005 | Update apply failed | Integrity | Error | 1 | none automatic; manual re-run permitted | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-006 | Rollback failed or unavailable | Integrity | Error | 1 when unavailable; 9 when a restore attempt left verification-failed … | none automatic | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-007 | Update already in progress | Concurrency | Error | 1 | none automatic; stale locks (holder dead) are reclaimed by the recovery pass | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-008 | Externally managed installation | Configuration | Error | 1 | none | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-009 | Unsupported upgrade path | Validation | Error | 1 | none | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-010 | Insufficient disk space for update | Resource | Error | 1 | none automatic | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-011 | Release yanked | Validation | Error | 1 | none; the next check offers the replacement | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |
| E-REL-012 | Update step timed out or cancelled | Timeout | Error | 8 | none automatic for explicit commands; scheduled automation retries next interval | [Vol 14 ch 02](../volume-14-distribution/02-updater-and-rollback.md) |

## Consolidation notes

- **Coverage.** 222 error codes across 23 areas round-trip against the volume registers'
  "Error codes minted" sections: every register row appears above and no code appears that a
  register does not list. Per-area counts: ARCH 7, PORT 3, AGT 11, WF 13, PROV 19, AUTH 11, TOOL 12, MCP 8, SKILL 6, PLUG 13, MEM 6, CTX 4, IDX 7, CLI 9, TUI 8, SEC 14, CFG 19, OBS 9, GIT 15, GH 6, PERF 4, TEST 6, REL 12.
- **Envelope field labels.** Chapters vary between the label forms "Exit code:" and
  "Exit-code mapping:" and between plain and bold bullet labels; both carry the same
  ADR-016 field. This annex reads them as one field.
- **Severity/category vocabularies** are per-area (e.g., Volume 13 uses "Test
  infrastructure"; Volume 5 uses lower-case classes). They are reproduced verbatim, not
  normalized, because the defining chapters are normative.
- **Register/chapter agreement.** Exit-code mappings in the registers agree with the
  defining chapters for all areas. Two register rows abbreviate what their chapters
  qualify: the Volume 6 register lists E-TOOL-001 as "6 (CLI usage surface: 2)" and the
  chapter states the same rule in prose (user-typed unknown names map to 2 at the CLI
  boundary); the Volume 7 register's conditional exits (e.g., E-MEM-002, E-IDX-002) match
  their chapters' qualified mappings. No conflicting rows were found.
- **E-GH codes** are raised by repository validators and CI checks (Volume 11), not by the
  product binary; their "exit code 1 (validator process)" mappings refer to the validator
  process, not to `andromeda`.

