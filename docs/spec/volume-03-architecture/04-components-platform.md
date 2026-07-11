# 04 — Components: Platform Services

This chapter covers the components that connect Andromeda to models, credentials, and the
extension ecosystem. Table conventions are those of chapter [03](03-components-core.md): the
universal dependencies (Core Domain, injected ports, Logging, TelemetryPort, EventBusPort,
SchedulerPort) are not restated, and each component's owning volume holds its behavioral
contract.

## Provider Layer

| Aspect | Specification |
|---|---|
| Responsibility | The provider contract and its adapters: implementing ProviderPort per provider, plus the routing/fallback composite that itself implements ProviderPort. Behavior owner: Volume 5 |
| Boundaries | All provider-specific code lives here, one package per adapter (Principle 1); the layer holds no agent logic, no context logic, and no credential material (references only, resolved via AuthPort/SecretStorePort) |
| Public API | Implements ProviderPort (adapters and router); consumes AuthPort, SecretStorePort, ConfigPort |
| Internal API | Adapter helper kit: HTTP client discipline per ADR-019, streaming decoders, error mapping to E-PROV, capability declaration helpers |
| Allowed dependencies | PAL (network via stdlib; proxy/CA config through ConfigPort); pinned external SDKs only where an adapter's ADR-019 PENDING VALIDATION resolves in favor |
| Prohibited dependencies | Engines (L2); other adapters; Tool Runtime; drivers |
| Inputs | ChatRequest/EmbedRequest/etc. contract types; provider responses and streams |
| Outputs | Contract-typed responses, capability sets, model catalogs, usage accounting data |
| Events emitted | `provider.*` family (Volume 5 mints names; e.g., the request-failed and fallback events Principle 7 requires) |
| Errors | E-PROV family (Volume 5) |
| States | Drives the Provider connection machine (`configured`, `verifying`, `available`, `degraded`, `unavailable`, `disabled`, `removed` — frozen names) |
| Persistence | Provider and Model catalog rows in the global DB (ADR-028) via the Persistence Layer |
| Concurrency | Concurrent requests per adapter within configured limits; router serializes state transitions per provider; streams follow port rules |
| Security | Official APIs only (Volume 1 constraint); credentials attach per-request from references and never persist in adapter state; TLS posture per Volume 9 |
| Observability | Every request spanned with provider/model attribution; token/cost usage emitted for Cost Records; provider switches announced (Principle 7) |
| Testing | Provider conformance suite (SM-01, SM-04) run against recorded fixtures and live seed providers per Volume 13; contract tests per adapter |
| Extensibility | Primary extension surface: new providers = new adapters (in-tree or via plugins exposing provider surfaces per ADR-009/Volume 6) |
| Phase | MVP (seed adapters per Volume 1: generic OpenAI-compatible, Anthropic, Ollama) |

The router deserves emphasis: because it implements ProviderPort itself, everything above it
is structurally incapable of knowing which provider served a request except through declared
attribution — which is exactly what Principle 7 requires be *announced*. Vendor agnosticism
stops being a code-review judgment and becomes a package-boundary fact (FR-ARCH-001).

## Authentication Layer

| Aspect | Specification |
|---|---|
| Responsibility | Implementing AuthPort: credential acquisition flows (API keys, OAuth device/browser where officially supported), refresh, rotation, revocation, and Authentication Session management. Behavior owner: Volume 5 (flows); Volume 9 (storage model) |
| Boundaries | Owns flows and session state; never stores material (SecretStorePort does), never decides storage policy (Volume 9), never talks to providers except their official auth endpoints |
| Public API | Implements AuthPort; consumed by Provider Layer, MCP Runtime (server auth), CLI/TUI auth commands |
| Internal API | Flow drivers per mechanism; token lifecycle scheduler (refresh-before-expiry) |
| Allowed dependencies | SecretStorePort, ConfigPort, PAL (browser/URL opening via Notifications/Shell surfaces where flows need it) |
| Prohibited dependencies | Engines; Tool Runtime; provider adapters (adapters depend on auth, never the reverse) |
| Inputs | Auth specs, provider auth endpoint responses, user interaction results relayed by drivers |
| Outputs | Authentication handles, rotation reports, profile listings (never material) |
| Events emitted | `auth.*` family (Volume 5 mints names) |
| Errors | E-AUTH family (Volume 5); exit code 4 mapping at the CLI boundary |
| States | Drives the Authentication Session machine (`unauthenticated`, `authenticating`, `active`, `refreshing`, `expired`, `failed`, `revoked` — frozen names); Credential `status` vocabulary |
| Persistence | Credential and Authentication Session rows in the **global DB only** (ADR-028 rule 2) |
| Concurrency | Single-flight refresh per session; flows requiring interaction serialize per credential |
| Security | The most security-sensitive component after Secret Store: official mechanisms only, no scraping/reverse engineering (Out of Scope list); material handled only as `SecretValue` in transit to SecretStorePort; all decisions audit-logged |
| Observability | Flow lifecycle events without secret content; expiry/refresh metrics; audit records per acquisition/rotation/revocation |
| Testing | Flow tests against recorded official endpoints; expiry/refresh clock tests; leak tests asserting no material in logs/events (Volume 13) |
| Extensibility | New auth mechanisms per provider arrive with their adapters (Volume 5); authentication is a named extension surface (Principle 6) |
| Phase | MVP |

Splitting authentication from the Provider Layer keeps a one-way dependency: adapters ask for
ready-to-use request credentials and cannot influence how they were obtained or stored.
It also gives OAuth-style interactive flows one home with driver integration — the same
Approval-like interaction discipline users already learn for permissions.

## Tool Runtime

| Aspect | Specification |
|---|---|
| Responsibility | The tool boundary (PRD-004): registration and discovery of tools from all origins, schema validation, permission mediation, sandbox placement, execution driving via ToolPort, result recording, and uniform observability. Behavior owner: Volume 6 |
| Boundaries | Mediates every tool invocation; implements no tool logic itself (built-in tools are ToolPort implementations registered like any other); never bypasses PermissionPort or SandboxPort |
| Public API | Tool invocation and registry API consumed by the Execution Engine, Agent Engine, and drivers (tool listings); consumes ToolPort (per tool), PermissionPort, SandboxPort, TerminalPort, SessionStorePort |
| Internal API | Registry with origin/trust bookkeeping; invocation pipeline (validate → permission → sandbox → execute → record) |
| Allowed dependencies | Plugin Runtime and MCP Runtime as tool *sources* via their registration APIs (they push ToolPort implementations in; the Tool Runtime never imports them — registration happens at composition) |
| Prohibited dependencies | Provider adapters; ProviderPort; direct PAL process spawning (SandboxPort only) |
| Inputs | Tool registrations (built-in, plugin, MCP), invocation requests with inputs, permission decisions, tool events |
| Outputs | Tool Results, invocation records, streamed tool activity, registry listings with origin and trust level always visible (Principle 4) |
| Events emitted | `tool.*` family (Volume 6 mints names) |
| Errors | E-TOOL family (Volume 6) |
| States | Drives the Tool Invocation machine (`requested`, `awaiting_approval`, `approved`, `executing`, `succeeded`, `failed`, `denied`, `timed_out`, `cancelled` — frozen names) |
| Persistence | Tool catalog rows per scope; invocations/results in the Run aggregate (Volume 2) |
| Concurrency | Per-tool concurrency limits from `Describe`; invocation pipeline stages are supervised tasks; cancellation propagates to sandbox teardown |
| Security | The enforcement point for SM-16(b): 100% of side-effecting invocations mediated; uniform citizenship — third-party tools get the same contract, permissions, sandboxing, and observability as built-ins (Principle 4) |
| Observability | Invocation spans with input summaries (redacted), outcome, timing; denial is a first-class recorded outcome (PRD-005) |
| Testing | Contract suite every tool must pass (Volume 13); fault injection (hangs, garbage output, limit breaches); SM-10 reliability measurement |
| Extensibility | The central extension surface: tools from SDK (in-process built-ins), plugins (ARP), and MCP servers are equal citizens |
| Phase | MVP |

The invocation pipeline's fixed order — validate, then permission, then sandbox, then execute
— is normative and not reorderable: validation before permission keeps malformed requests
from consuming user attention; permission before sandbox keeps denied work from allocating
resources; and nothing executes outside a sandbox handle (ADR-021). The pipeline is also
where every invocation acquires its correlation spine (run → task → invocation → permission
decision) that SM-13 audits end-to-end.

## Plugin Runtime

| Aspect | Specification |
|---|---|
| Responsibility | Managing plugin subprocesses over the Andromeda Runtime Protocol (ADR-009): spawn through SandboxPort, handshake and version negotiation, capability/surface registration, health, restart policy, orderly stop. Behavior owner: Volume 6 |
| Boundaries | Owns the ARP host side and plugin process lifecycles; the tools/providers/exporters a plugin registers are handed to their governing components (Tool Runtime, Provider Layer registry, Telemetry) as port implementations — the Plugin Runtime never executes tool logic |
| Public API | Plugin lifecycle and registry API consumed by CLI/TUI (plugin commands) and the composition root; consumes SandboxPort, PermissionPort, ConfigPort, SecretStorePort (plugin-declared credential needs per Volume 6/9 policy) |
| Internal API | ARP codec (JSON-RPC 2.0 framing shared conventions with ADR-010/ADR-012), supervision state machines, capability tables |
| Allowed dependencies | PAL (Processes, Process Trees, Local IPC surfaces via SandboxPort) |
| Prohibited dependencies | Engines; provider adapters; MCP Runtime |
| Inputs | Plugin manifests and registrations, ARP messages, health signals, policy decisions |
| Outputs | Registered extension surfaces (ToolPort implementations et al.), plugin status, ARP requests/responses |
| Events emitted | `plugin.*` family (Volume 6 mints names) |
| Errors | E-PLUG family (Volume 6) |
| States | Drives the Plugin machine (`registered`, `starting`, `running`, `stopping`, `stopped`, `failed`, `disabled`, `removed` — frozen names) |
| Persistence | Plugin registry rows per scope (ADR-028); runtime state in memory, reconstructed on start |
| Concurrency | One supervisor per plugin process; concurrent ARP calls multiplexed per protocol rules; restart with backoff per policy |
| Security | Plugins launch exclusively through SandboxPort (ADR-009 risk mitigation); trust level and origin recorded; capabilities mediated by Permission Manager; crash containment at the process boundary |
| Observability | Handshake/lifecycle events; per-plugin resource metrics via PAL Process Trees; ARP frames traceable in debug mode with redaction |
| Testing | ARP conformance fixtures (shared with SDK); chaos tests (kill, hang, garbage frames); restart-policy property tests |
| Extensibility | This *is* an extension mechanism; the SDK ships the Go reference implementation of the plugin side (ADR-009) |
| Phase | Beta |

The Plugin Runtime is deliberately a *host*, not a framework: everything a plugin can do is
expressed as an implementation of an existing port surface, so a plugin-provided tool is
indistinguishable from a built-in at every layer above registration (Principle 4). ARP's
JSON-RPC 2.0 grammar is shared with MCP and external IPC (ADR-009/010/012), keeping one
protocol mental model across all three process boundaries.

## Workflow Engine

| Aspect | Specification |
|---|---|
| Responsibility | Executing Workflow definitions as state machines: step sequencing, agent/tool step dispatch, approval gates, timers, resumability, and Workflow Run bookkeeping — including the specification-driven development workflow (PRD-012). Behavior owner: Volume 4 |
| Boundaries | Orchestrates; delegates agent work to the Agent Engine, task execution to the Execution Engine, approvals to PermissionPort. Owns the workflow definition format and its validation |
| Public API | Workflow execution API consumed by the Runtime (workflow runs enter like any run family) and drivers; consumes Agent Engine, Execution Engine, Planner (L2 peers); PermissionPort, SessionStorePort, ConfigPort |
| Internal API | Definition compiler (declared YAML/TOML form → executable machine per Volume 4), gate/timer machinery |
| Allowed dependencies | L2 peers named above; their ports |
| Prohibited dependencies | Adapters; ToolPort directly; provider adapters |
| Inputs | Workflow definitions (versioned catalog entities), run parameters, approval decisions, step results |
| Outputs | Workflow Runs with complete step histories, spawned Runs (referenced by ID per Volume 2), artifacts |
| Events emitted | `workflow.*` family (Volume 4 mints names) |
| Errors | E-WF family (Volume 4) |
| States | Drives the Workflow Run machine (`pending`, `running`, `awaiting_approval`, `paused`, `interrupted`, `completed`, `failed`, `cancelled` — frozen names) |
| Persistence | Workflow definitions per scope; Workflow Runs in the workspace DB; step-boundary persistence makes `interrupted` resume at the last persisted step (Volume 2 chapter 09) |
| Concurrency | Parallel steps where the definition declares them, under run-level supervision; gates block without holding threads |
| Security | Approval gates are PermissionPort decisions with full audit; workflows execute under the invoking session's permission scope — a workflow cannot escalate beyond it |
| Observability | Step transitions evented; workflow histories complete by construction (PRD-012); span tree per workflow run |
| Testing | Definition-validation golden tests; resumability crash tests at every step boundary; gate/timeout property tests |
| Extensibility | Workflows are a named extension surface (Principle 6): definitions are user- and package-distributable; step vocabulary extension per Volume 4 |
| Phase | Beta |

Workflows are the "process, not prompt" answer (PRD-012): declared, versioned, resumable
orchestrations whose histories are complete. Architecturally the engine adds no new
authority — it composes the same engines and permission decisions an interactive session
uses, which is why workflow runs inherit every guarantee (persistence, audit, cancellation)
without duplicate machinery.

## Skill Engine

| Aspect | Specification |
|---|---|
| Responsibility | Loading, validating, composing, and applying Skills: package format parsing, capability/tool requirement checks, prompt-fragment contribution through the Prompt Engine registry, activation rules per agent profile. Behavior owner: Volume 6 |
| Boundaries | Owns the skill format and lifecycle; contributes to prompts and tool sets but executes nothing — a skill has no runtime code path of its own (procedural knowledge, not a process; code-bearing extensions are plugins) |
| Public API | Skill registry/activation API consumed by Agent Engine (profile resolution) and drivers (skill commands); consumes ConfigPort, PackagePort (skill delivery) |
| Internal API | Format parser/validator, composition resolver (conflict rules per Volume 6) |
| Allowed dependencies | Prompt Engine registry (L2); Tool Runtime registry (to verify required tools exist) |
| Prohibited dependencies | Adapters; ProviderPort; SandboxPort (nothing to execute) |
| Inputs | Skill packages, activation requests, agent profile references |
| Outputs | Validated skill registrations, prompt contributions, tool requirement declarations, composition reports |
| Events emitted | `skill.*` family (Volume 6 mints names) |
| Errors | E-SKILL family (Volume 6) |
| States | Skills are stateless catalog entities (Volume 2); activation state is per-session configuration |
| Persistence | Skill rows per scope (builtin: in binary) per ADR-028 |
| Concurrency | Read-mostly registry; loading/validation on demand; concurrent-safe |
| Security | Skills alter agent behavior, so sources are trust-gated (Volume 9 policy via Package Manager); a skill cannot grant itself tools — it declares requirements that permission policy still mediates |
| Observability | Active skill set recorded per run (SM-12 reproducibility); composition decisions evented |
| Testing | Format golden tests; composition conflict matrix tests; activation snapshot tests |
| Extensibility | Skills are a primary extension surface (Principle 6), distributable via packages |
| Phase | Beta |

The engine's core design constraint is that skills stay *data*: versioned prompt and
requirement bundles whose effect on a run is fully recorded. That keeps the trust story
tractable (reviewable text, not executable code) and cleanly separates the skill surface from
the plugin surface — when procedural knowledge needs code, it graduates to a plugin tool and
the skill references it.

## MCP Runtime

| Aspect | Specification |
|---|---|
| Responsibility | Managing MCP client connections per ADR-010: transport establishment (stdio subprocess or remote), protocol revision negotiation, discovery of tools/resources/prompts, bridging them into Andromeda surfaces, connection lifecycle and reconnection. Behavior owner: Volume 6 |
| Boundaries | Wraps the official Go SDK behind its own boundary — SDK types MUST NOT leak into any port signature or the Core Domain (ADR-010 rule 1); discovered tools surface through the Tool Runtime as ToolPort implementations |
| Public API | MCP server registry and connection API consumed by drivers (MCP commands) and composition; provides ToolPort bridge implementations to the Tool Runtime registry; consumes SandboxPort (stdio server spawn), SecretStorePort + AuthPort (server credentials), PermissionPort, ConfigPort |
| Internal API | SDK wrapper, revision policy, discovery-to-registry mapping, reconnection state machines |
| Allowed dependencies | `modelcontextprotocol/go-sdk` (pinned v1 line, ADR-010); PAL via SandboxPort for subprocess transports |
| Prohibited dependencies | Engines; Provider Layer; Plugin Runtime |
| Inputs | MCP Server registrations, protocol messages, discovery results, trust policy decisions |
| Outputs | Bridged ToolPort implementations (origin `mcp`, trust level visible), resource/prompt surfaces per Volume 6, connection status |
| Events emitted | `mcp.*` family (Volume 6 mints names) |
| Errors | E-MCP family (Volume 6) |
| States | Drives the MCP Client Connection machine (`configured`, `connecting`, `initializing`, `ready`, `reconnecting`, `disconnected`, `failed`, `disabled`, `removed` — frozen names) |
| Persistence | MCP Server rows per scope; connections in the workspace DB (Volume 2) |
| Concurrency | One connection supervisor per server; concurrent requests multiplexed per MCP protocol; reconnection with backoff |
| Security | Stdio servers spawn through SandboxPort; remote servers under Volume 9 network policy; OAuth-based server authorization is PENDING VALIDATION per ADR-010 and not a stable MVP feature — token/header auth via ADR-014 handling is the supported path; per-server trust levels gate tool exposure |
| Observability | Connection lifecycle events; per-server tool inventories; protocol errors surfaced with revision context |
| Testing | Conformance suite per declared protocol revision (SM-15a); interop job against the public reference-server set (SM-15b); recorded-session tests independent of SDK internals (ADR-010) |
| Extensibility | MCP is itself the extension mechanism; Andromeda-side extension is configuration (server registrations), not code |
| Phase | Beta |

The bridge design keeps MCP a *source of tools*, not a second tool system: once a discovered
tool is registered, the Tool Runtime's pipeline (validation, permissions, sandbox-relevant
limits, observability) applies identically (Principle 4). The wrapper boundary around the
official SDK is load-bearing for ADR-010's reversal plan — conformance is testable from
recorded sessions, and an SDK replacement would not ripple past this component.
