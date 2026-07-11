# 08 — Plugin Runtime and the Andromeda Runtime Protocol

A **Plugin** is an external extension process integrated through the **Plugin Runtime**
over the **Andromeda Runtime Protocol (ARP)** — JSON-RPC 2.0 on stdio, per
[ADR-009](../annexes/adr/ADR-009.md). Plugins are the code-bearing extension surface:
tools, provider adapters, exporters, and command contributions that cannot be expressed as
data-only skills. This chapter defines the plugin manifest, the ARP wire contract
([ADR-079](../annexes/adr/ADR-079.md)): framing, handshake and version negotiation, the
method surface — plus permissions, sandboxing, and lifecycle supervision. The full Plugin
state machine is in [chapter 10](10-state-machines.md); plugin distribution is
[chapter 09](09-package-manager-supply-chain.md).

Design stance (Volume 3, Plugin Runtime): the runtime is a **host, not a framework**.
Everything a plugin provides is an implementation of an existing contract — a
plugin-provided tool is a ToolPort implementation registered with the Tool Runtime and is
indistinguishable from a built-in at every layer above registration (Principle 4;
FR-ARCH-011).

## Plugin manifest

A plugin package contains an entry-point executable and a `plugin.toml` manifest:

```toml
[plugin]
name = "jira-bridge"
version = "0.3.1"
description = "Tools for querying and updating Jira issues."
authors = ["Sam Author <sam@example.com>"]
license = "Apache-2.0"
entrypoint = "bin/jira-bridge"           # relative to the install location
protocol = ">=1.0 <2.0"                   # ARP version range the plugin implements
contract_version = "1.0"                  # Extension SDK contract targeted (SM-20)

[plugin.surfaces]
tools = ["jira.search", "jira.update"]    # every surface MUST be declared (INV-PLG-01)
commands = []
exporters = []

[plugin.permissions]
required = ["network", "external_service_access"]  # Volume 9 permission names only
optional = ["notifications"]

[plugin.credentials]
declared = ["jira-token"]                 # credential names the plugin may request

[plugin.runtime]
autostart = false                          # start on demand (first surface use)
idle_stop_minutes = 15                     # 0 disables idle stop
```

The manifest MUST parse as TOML and validate against the plugin manifest schema
(ADR-024, mirrored in `sdk/`). The following manifest is invalid — undeclared permission
name and an absolute entrypoint — and MUST be rejected with E-PLUG-006:

```toml invalid
[plugin]
name = "bad"
version = "1.0.0"
entrypoint = "/usr/local/bin/anything"   # absolute paths are prohibited
protocol = ">=1.0 <2.0"

[plugin.permissions]
required = ["root_access"]               # not a Volume 9 permission name
```

Rules: `name` follows the skill name grammar (chapter 07); `version` is SemVer;
`entrypoint` MUST resolve inside the install location (no traversal, no symlink escape);
`permissions.required`/`optional` use only the closed Volume 9 permission vocabulary;
surfaces not declared MUST NOT be registered at runtime (INV-PLG-01, enforced with
E-PLUG-007).

## The Andromeda Runtime Protocol

### Framing and transport

- JSON-RPC 2.0 messages, UTF-8, one message per line (`\n`-delimited, no embedded
  newlines), on the plugin process's stdin (host→plugin) and stdout (plugin→host) —
  the same stdio framing conventions as MCP (ADR-009 decision), keeping one mental model.
- stderr is never protocol: it is captured as plugin log output, line-buffered, redacted,
  and tagged `plugin=<name>` in local logs.
- Maximum frame size: 4 MiB. Larger payloads (artifacts, file contents) pass out-of-band
  as filesystem references inside sandbox-visible paths (ADR-009 negative-consequence
  mitigation); a frame exceeding the limit is a protocol violation (E-PLUG-003).
- Request concurrency: the host MAY pipeline requests; the plugin MAY answer out of
  order; `id` correlation is authoritative. Notifications are never answered.

### Versioning

ARP versions are `MAJOR.MINOR` strings; the initial version is `1.0`. The host supports
the current and previous minor within a major (Volume 3 compatibility strategy); breaking
protocol changes require a major bump and follow SM-20 discipline via ADR-015. The
negotiated version is recorded on the Plugin row (`protocol_version`, Volume 2).

### Handshake

1. The Plugin Runtime spawns the entrypoint through SandboxPort (`ExecuteIn`) under the
   plugin sandbox tier.
2. The host sends the `arp.initialize` request. The plugin MUST NOT write anything to
   stdout before this request arrives.
3. The plugin answers within `plugins.handshake_timeout_ms` (default 10000) choosing the
   highest mutually supported protocol version and declaring its identity and surfaces.
4. The host validates the response against the manifest (identity match, surfaces ⊆
   manifest declarations), registers the declared surfaces with their governing
   components, sends the `arp.initialized` notification, and the plugin is `running`.

```json
{"jsonrpc": "2.0", "id": 1, "method": "arp.initialize", "params": {
  "host": {"name": "andromeda", "version": "0.4.0"},
  "protocol_versions": ["1.0"],
  "plugin_id": "01J9ZC3AC9V1B2C3D4E5F6G7H8",
  "workspace": {"root_hint": "/work/project"}}}
```

```json
{"jsonrpc": "2.0", "id": 1, "result": {
  "protocol_version": "1.0",
  "plugin": {"name": "jira-bridge", "version": "0.3.1"},
  "surfaces": {"tools": ["jira.search", "jira.update"], "commands": [], "exporters": []}}}
```

Failure modes: no mutual version → the plugin returns the JSON-RPC error `-32001`
(version unsupported) with its supported range, the host records E-PLUG-002 and the
plugin enters `failed`; malformed or manifest-inconsistent response → E-PLUG-003;
handshake timeout → E-PLUG-003 (timeout class) and process termination via sandbox
teardown.

### Method surface (ARP 1.0)

Host → plugin requests:

| Method | Params (summary) | Result (summary) | Budget |
|---|---|---|---|
| `arp.initialize` | host info, offered protocol versions, plugin ULID, workspace hint | chosen version, identity, surfaces | `handshake_timeout_ms` |
| `arp.shutdown` | reason | acknowledgment (then process exit expected) | `stop_timeout_ms` |
| `arp.ping` | nonce | nonce echo | 5000 ms fixed |
| `arp.tool.describe` | tool name | full tool declaration per FR-TOOL-001 | `request_timeout_ms` |
| `arp.tool.validate` | tool name, input JSON | validation result | `request_timeout_ms` |
| `arp.tool.execute` | invocation ULID, tool name, validated input, granted permission set, effective limits | terminal result frame (stream via notifications below) | tool-declared timeout |
| `arp.tool.cancel` | invocation ULID | acknowledgment | 5000 ms fixed |
| `arp.command.execute` | command surface invocation (declared commands only) | command result | `request_timeout_ms` |

Plugin → host notifications:

| Notification | Payload (summary) | Semantics |
|---|---|---|
| `arp.tool.event` | invocation ULID, ordered ToolEvent (progress, partial output, log line, artifact reference) | streamed tool activity; order per invocation is preserved |
| `arp.log` | level, message, fields | merged into local logs with redaction |

Plugin → host requests: none in ARP 1.0. A plugin never calls back into the host for
permissions, secrets, or execution — permissions are granted per invocation in
`arp.tool.execute` params, and credentials the manifest declares are delivered as
sandbox-filtered environment entries resolved at spawn (host-mediated; the plugin cannot
request more at runtime). This keeps the authority direction one-way and the protocol
surface auditable; a reverse-request surface, if ever justified, is a new protocol minor
with its own decision record.

JSON-RPC error codes reserved by ARP: `-32001` version unsupported, `-32002` unknown
surface, `-32003` invocation unknown, `-32004` plugin busy (declared concurrency
exceeded), `-32005` shutting down. Standard JSON-RPC codes retain their meanings. The
host maps plugin-reported errors into E-PLUG codes; tool-level failures inside
`arp.tool.execute` are Tool Results (data), not protocol errors — mirroring ToolPort
semantics.

### FR-PLUG-001 — Plugin runtime over the Andromeda Runtime Protocol

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: Plugin Runtime (Volume 6)
- Affected components: Plugin Runtime, Tool Runtime, Sandbox Engine, Permission Manager, Extension SDK
- Dependencies: ADR-009, ADR-079; FR-TOOL-001, FR-SDK-001, FR-SEC-101; FR-ARCH-011
- Related risks: RISK-PLUG-001, RISK-PLUG-002

#### Description

Andromeda MUST run plugins as supervised subprocesses speaking ARP as specified in this
chapter: manifest-validated registration; sandbox-mediated spawn; the `arp.initialize`
handshake with version negotiation; surface registration limited to manifest
declarations; request routing from governing components (Tool Runtime for tools) over the
plugin's connection; streamed tool events; cooperative cancellation; orderly shutdown via
`arp.shutdown` with kill-after-timeout; and crash containment such that no plugin failure
mode (crash, hang, garbage output, resource exhaustion) terminates or corrupts the
Andromeda process. The Extension SDK MUST ship a Go reference implementation of the
plugin side plus conformance fixtures; the protocol — not the SDK — is the contract
(any language reading stdin and writing stdout can implement it).

#### Motivation

The plugin surface is how third-party code extends Andromeda without forking it
(PRD-007); the subprocess boundary is what makes that safe to offer (ADR-009).

#### Actors

Plugin authors; the Plugin Runtime; governing components consuming registered surfaces;
users managing plugins.

#### Preconditions

Plugin registered (manifest validated); enabled; sandbox tier available.

#### Main flow

1. Start trigger (autostart, first surface use, or explicit command).
2. Spawn through SandboxPort; handshake; surface registration; `running`.
3. Invocations route through governing components to ARP methods; results and events
   stream back; records are written by the governing component's pipeline.

#### Alternative flows

- Crash while `running`: containment + restart policy per FR-PLUG-004.
- Cancellation: `arp.tool.cancel` gives the plugin the chance to stop cooperatively; the
  invocation terminates through its event stream with the `cancelled` outcome; sandbox
  teardown backstops a non-cooperating plugin at the tool teardown budget.

#### Edge cases

- A plugin declaring a tool name already registered by another origin: registration
  fails for that tool with the conflict surfaced through the Tool Runtime registry rules
  (Volume 6 chapter 01 ownership); other declared surfaces proceed.
- A plugin writing to stdout before `arp.initialize`: protocol violation, E-PLUG-003,
  process terminated.
- Two registered versions of a plugin: prohibited by the natural key (one row per name
  per scope, Volume 2); upgrades replace via package operations.

#### Inputs

Manifests, ARP frames, invocation requests with granted permission sets, lifecycle
commands.

#### Outputs

Registered surfaces; Tool Results and streamed events; plugin status and diagnostics;
`plugin.*` events.

#### States

Plugin canonical states (`registered`, `starting`, `running`, `stopping`, `stopped`,
`failed`, `disabled`, `removed`); full machine in chapter 10.

#### Errors

E-PLUG-001 – E-PLUG-007; tool-level failures are Tool Results per ToolPort semantics.

#### Constraints

Spawn exclusively through SandboxPort (ADR-009 risk mitigation); frame limit 4 MiB;
declared-surfaces-only registration (INV-PLG-01); no plugin→host request surface in ARP
1.0.

#### Security

Per FR-PLUG-003; additionally, the host treats every plugin frame as untrusted input:
schema-validated before dispatch, size-bounded, and never evaluated.

#### Observability

Handshake and lifecycle events; per-plugin resource metrics via PAL process accounting;
ARP frames traceable in debug mode with redaction (Volume 3 Plugin Runtime row).

#### Performance

Per NFR-PLUG-002 (dispatch overhead) and the timeout table in this chapter.

#### Compatibility

ARP version negotiation with current+previous minor support; `contract_version` recorded
per Extension row (SM-20 tracking); plugin binaries are platform-specific artifacts
handled by chapter 09 packaging.

#### Acceptance criteria

- Given a registered plugin exposing one tool, when an agent invokes it, then the
  invocation flows through the Tool Runtime pipeline, `arp.tool.execute` is issued with
  the granted permission set, events stream to the TUI live, and the terminal event
  becomes the recorded Tool Result.
- Given a plugin process killed externally mid-invocation, when the host detects exit,
  then the invocation records `failed` with E-PLUG-005, the Plugin row transitions per
  chapter 10, and the Andromeda process continues unaffected.
- Given a plugin that answers `arp.initialize` claiming a surface absent from its
  manifest, when the host validates, then that surface is not registered, E-PLUG-007 is
  recorded, and the handshake fails.
- Negative case: a frame of 5 MiB is rejected with E-PLUG-003 and the offending request
  fails without affecting other in-flight requests.
- Permission case: a plugin tool invoked without its required permission granted is
  denied by the Tool Runtime pipeline before any ARP traffic occurs (denial recorded).
- Observability case: every lifecycle transition emits exactly one `plugin.*` event with
  the plugin ULID and correlation IDs.

#### Verification method

ARP conformance fixtures (shared with the SDK) covering handshake, streaming,
cancellation, and every reserved error code; chaos tests (kill, hang, garbage frames,
oversize frames); cross-language smoke plugin (script-based) proving
language-agnosticism; SM-03 timed exercise.

#### Traceability

PRD-004, PRD-007; ADR-009, ADR-079; FR-TOOL-001, FR-SDK-001; Volume 2 Plugin entity and
INV-PLG-01..04; chapter 10.

### FR-PLUG-002 — ARP handshake, version negotiation, and method conformance

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Derived
- Owner: Plugin Runtime (Volume 6)
- Affected components: Plugin Runtime, Extension SDK
- Dependencies: FR-PLUG-001; ADR-009, ADR-015, ADR-079
- Related risks: RISK-PLUG-002

#### Description

The Plugin Runtime MUST implement the handshake exactly as specified: host-initiated
`arp.initialize` carrying the offered protocol version list; plugin selection of the
highest mutual version; manifest-consistency validation of the response; and the
`arp.initialized` completion notification. The host MUST support ARP versions per the
current+previous-minor rule and MUST record the negotiated version persistently
(INV-PLG-02). Every ARP method and notification of the negotiated version MUST behave per
the method surface tables, including reserved error codes and budgets; unknown methods
received by the host are answered with JSON-RPC method-not-found and logged; unknown
notifications are ignored and counted. The SDK reference implementation and the
conformance fixtures MUST cover 100% of the negotiated version's method surface.

#### Motivation

Version negotiation from day one is the ADR-009 mitigation for protocol evolution debt;
mechanical conformance keeps N host implementations and M SDK languages honest against
one contract.

#### Actors

Plugin Runtime; plugin implementations; SDK maintainers.

#### Preconditions

Spawn succeeded; stdio channels open.

#### Main flow

1. Offer versions; receive selection and declarations.
2. Validate; register; notify `arp.initialized`.

#### Alternative flows

- No mutual version: `-32001` → E-PLUG-002, `failed`, diagnostics show both ranges.
- Response inconsistent with manifest: E-PLUG-003/E-PLUG-007 per finding class.

#### Edge cases

- A plugin supporting a newer minor than the host (plugin `1.1`, host `1.0`): negotiation
  selects `1.0`; plugin features beyond `1.0` stay dormant.
- Re-handshake is prohibited within one process lifetime; a second `arp.initialize`
  response or request is a protocol violation.

#### Inputs

Version tables, manifest, handshake frames.

#### Outputs

Negotiated `protocol_version` on the Plugin row; registered surfaces.

#### States

`starting` → `running` on success; `starting` → `failed` on all failure modes.

#### Errors

E-PLUG-002, E-PLUG-003, E-PLUG-007.

#### Constraints

Handshake budget `plugins.handshake_timeout_ms`; silence-before-initialize rule; single
handshake per process.

#### Security

The handshake carries no secrets; the workspace hint is a path string subject to the
sandbox filesystem policy the plugin already lives under.

#### Observability

`plugin.handshake.completed` / `plugin.handshake.failed` with negotiated or refused
versions.

#### Performance

Handshake completes within its budget; measured in conformance runs.

#### Compatibility

The negotiation rule *is* the compatibility mechanism; version support windows follow
ADR-015 SemVer discipline for the protocol artifact.

#### Acceptance criteria

- Given a plugin implementing exactly ARP 1.0, when the host offers `["1.0"]`, then
  negotiation selects `1.0` and `protocol_version` records it.
- Given a plugin whose only supported range is `>=2.0`, when handshaking, then the host
  records E-PLUG-002 with both ranges and the plugin is `failed` without restart (the
  failure is deterministic; restart policy excludes it).
- Negative case: a plugin emitting a valid JSON-RPC frame before `arp.initialize` is
  terminated with E-PLUG-003 (early-output class).
- Observability case: handshake events carry offered, selected, and (on failure) refused
  version data.

#### Verification method

Conformance fixtures parameterized by version tables; SDK cross-tests; property tests on
the negotiation function (any two version sets → deterministic outcome or refusal).

#### Traceability

ADR-009 risks; ADR-079; INV-PLG-02; SM-20 (protocol in the public contract set).

## Permissions and sandboxing

### FR-PLUG-003 — Plugin permission mediation and sandbox containment

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: Plugin Runtime (Volume 6)
- Affected components: Plugin Runtime, Permission Manager, Sandbox Engine, Secret Store, Audit Log
- Dependencies: FR-PLUG-001, FR-SEC-100, FR-SEC-101, FR-SEC-102; ADR-014, ADR-021
- Related risks: RISK-PLUG-001

#### Description

Plugins MUST be contained on two planes. **Process plane:** every plugin process runs
under the plugin sandbox tier (FR-SEC-101): filtered environment (deny-by-default; only
sandbox-approved entries plus manifest-declared credential deliveries), filesystem policy
derived from workspace scope, resource limits (CPU, memory, output size, wall clock per
sandbox policy), and process-tree termination on stop. Spawning a plugin requires a
`process_spawn` permission decision; first start of a plugin additionally follows trust
policy (an Approval for below-threshold trust, mirroring skill activation).
**Invocation plane:** plugin-provided tools declare permissions in their tool
declarations (bounded by the manifest's `permissions.required ∪ optional`); each
invocation passes the Tool Runtime pipeline, and the granted set travels in
`arp.tool.execute` — the plugin never holds standing grants. Manifest-declared
credentials are resolved through SecretStorePort at spawn and delivered as filtered
environment entries; every delivery is audit-logged. A plugin MUST NOT be able to obtain
any permission, credential, or filesystem/network reach beyond these two planes.

#### Motivation

Third-party code with ambient authority is the classic extension failure; two-plane
mediation keeps the process boundary (ADR-009) meaningful and every grant recorded
(PRD-005, PRD-006).

#### Actors

Plugin Runtime; Permission Manager; Sandbox Engine; plugin processes (contained
subjects).

#### Preconditions

Sandbox tier profile available; policy documents loaded; manifest validated.

#### Main flow

1. Start: permission decision (`process_spawn` + trust policy) → sandbox prepare →
   credential delivery per manifest → spawn.
2. Invoke: pipeline decision per tool invocation → granted set in `arp.tool.execute`.

#### Alternative flows

- Trust denial at first start: plugin stays `registered`/`disabled` per user choice;
  decision recorded.
- Credential declared but absent: start proceeds without it; the plugin's dependent
  tools fail at invocation with the missing-credential cause (no implicit prompting at
  spawn).

#### Edge cases

- A tool declaration requesting a permission outside the manifest bound fails
  registration (E-PLUG-006 consistency class) — the manifest is the ceiling.
- Manifest `optional` permissions denied by policy simply never appear in granted sets;
  the plugin must degrade per its declarations.
- Child processes spawned by a plugin live inside the same sandbox process tree and die
  with it (ADR-021 child-process control).

#### Inputs

Manifest permission/credential declarations, policy, sandbox profiles, invocation grant
sets.

#### Outputs

Contained processes; per-invocation grants; audit records for spawn, credential
delivery, denials.

#### States

Interacts with `starting` (containment applied before handshake) and every invocation's
Tool Invocation machine.

#### Errors

E-SEC family for denials and sandbox refusals (surfaced through E-PLUG-001 at spawn);
E-PLUG-006 for manifest/declaration inconsistencies.

#### Constraints

No direct spawn path outside SandboxPort; environment deny-by-default; credentials by
reference until the delivery point; effective containment level recorded per execution
(ADR-021).

#### Security

This requirement is the plugin-surface instantiation of SM-16(b): 100% of plugin side
effects mediated; the OS-level isolation strengthening at Beta/v1 remains PENDING
VALIDATION per ADR-021 and is tracked in the register.

#### Observability

Audit records for every decision and delivery; sandbox containment level visible in
plugin diagnostics; denial events flow through the Tool Runtime's recorded outcomes.

#### Performance

Containment setup cost is paid at spawn, not per invocation; per-invocation overhead
budget per NFR-PLUG-002.

#### Compatibility

Uniform across platforms through the PAL; per-OS mechanism differences live behind
SandboxPort per ADR-021.

#### Acceptance criteria

- Given a plugin whose manifest requires `network`, when its tool executes with the grant,
  then network egress succeeds; when the grant is withheld, then egress attempts fail
  inside the sandbox and the denial is attributable in the audit chain.
- Given a plugin attempting to read an environment variable not delivered, when it
  executes, then the variable is absent (assertion fixture).
- Given a declared credential, when the plugin starts, then delivery is audit-logged and
  the secret never appears in logs, events, or ARP frames captured in debug mode.
- Negative case: a tool declaring `system_modification` while the manifest omits it fails
  registration with the inconsistency named.
- Observability case: spawn denial produces an Approval record and a `plugin.*` event
  with the denial class.

#### Verification method

Sandbox assertion fixtures (environment, filesystem, network, process-tree kill);
audit-chain tests; manifest-bound consistency tests; secret-scan over debug frame
captures.

#### Traceability

PRD-005, PRD-006; FR-SEC-100/101/102; ADR-009, ADR-014, ADR-021; SM-16.

## Supervision and lifecycle operations

Operational policy (full machine in chapter 10):

- **Start** — autostart at workspace open (`autostart = true`), on-demand at first
  surface use, or explicit command.
- **Health** — `arp.ping` every `plugins.health_interval_ms` (default 30000) while
  `running`; 3 consecutive misses → the plugin is deemed unresponsive → orderly stop
  escalation (below) → `failed` with restart policy.
- **Restart policy** — on unexpected exit or unresponsiveness: exponential backoff
  starting at `plugins.restart_backoff_initial_ms` (default 500), doubling, capped at
  30000 ms, at most `plugins.restart_max_attempts` (default 5) within a 10-minute window;
  exhaustion leaves `failed` (resting) until user action. Deterministic failures
  (E-PLUG-002 version refusal, E-PLUG-006 manifest invalidity) never restart.
- **Stop** — `arp.shutdown`, wait `plugins.stop_timeout_ms` (default 5000) for clean
  exit, then sandbox teardown (kill process tree). In-flight invocations receive
  cancellation first.
- **Idle stop** — `idle_stop_minutes` after the last surface use, on-demand plugins stop
  to reclaim resources; next use restarts them.

### FR-PLUG-004 — Plugin supervision, health, and restart

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Design
- Owner: Plugin Runtime (Volume 6)
- Affected components: Plugin Runtime, Task Scheduler, Sandbox Engine
- Dependencies: FR-PLUG-001; ADR-023
- Related risks: RISK-PLUG-001

#### Description

Each plugin process MUST run under one supervisor (a SchedulerPort-managed task per
ADR-023) implementing the operational policy above: health probing, unresponsiveness
detection, bounded restart with backoff and windowed attempt counting, orderly stop with
kill escalation, idle stop, and startup reconciliation (running-family rows found at
startup are reconciled to `stopped` before restart policy applies — INV-PLG-03).
Supervision events and restart counters MUST be observable per plugin. A restart MUST
re-run the full containment and handshake path (no state carryover into the new process);
surface registrations from the dead process are replaced atomically when the new
handshake completes, and remain absent in between (invocations during the gap fail with
E-PLUG-005 semantics, never queue).

#### Motivation

Plugins are long-lived third-party processes; without supervision their failure modes
(hangs, leaks, crash loops) become user-facing mysteries. Bounded restart converts crash
loops into a diagnosable resting `failed` state.

#### Actors

Plugin Runtime supervisors; Task Scheduler; users inspecting status.

#### Preconditions

Plugin `running` (health) or terminated unexpectedly (restart).

#### Main flow

1. Probe on interval; track misses; escalate at threshold.
2. On unexpected exit: compute backoff, schedule restart attempt, re-contain,
   re-handshake.
3. On success: counters reset after 10 minutes of stable `running`.

#### Alternative flows

- Attempt exhaustion: `failed` resting; `plugin.restarted` not emitted; user notified
  once (not per attempt).
- Stop requested during backoff wait: pending restart cancelled; state `stopped`.

#### Edge cases

- Crash during `stopping` counts as stop completion (`stopped`), not a restartable
  failure.
- A plugin that exits cleanly with code 0 outside a stop request is still an unexpected
  exit (the host, not the plugin, owns lifecycle) — restart policy applies with the
  clean-exit class recorded.
- Clock adjustments do not extend backoff (monotonic timers).

#### Inputs

Process exit notifications (PAL), probe results, stop requests, configuration.

#### Outputs

Restart attempts, state transitions, counters, notifications.

#### States

`running` ↔ `failed`/`stopping`/`stopped` per chapter 10; this requirement owns the
policy parameters.

#### Errors

E-PLUG-004 (probe timeout class), E-PLUG-005 (crash), E-PLUG-001 (respawn failure).

#### Constraints

Monotonic backoff timing; single supervisor per plugin; reconciliation before any
startup restart (INV-PLG-03).

#### Security

Restart never skips containment or trust checks; a plugin disabled by policy mid-backoff
does not restart.

#### Observability

Restart counters and last-error on the Plugin row (`last_error`, Volume 2); supervision
transitions evented; per-plugin resource metrics sampled.

#### Performance

Probe cost one round trip per interval; restart storm bounded by policy (≤ 5 spawns per
10 minutes per plugin).

#### Compatibility

Policy parameters configurable per install; defaults uniform across platforms.

#### Acceptance criteria

- Given a plugin that crashes immediately on every start, when started, then exactly
  `restart_max_attempts` respawns occur with doubling backoff, the plugin rests in
  `failed`, and one user notification exists.
- Given a plugin that stops answering `arp.ping`, when 3 probes miss, then the host
  escalates stop → kill and applies restart policy, and in-flight invocations record
  `failed` with E-PLUG-004/E-PLUG-005 causes.
- Given Andromeda restarting after a crash, when the registry loads, then plugins
  persisted as `running` are reconciled to `stopped` before any restart.
- Negative case: a version-refused plugin (E-PLUG-002) is never auto-restarted.
- Observability case: `plugin.restarted` events carry the attempt number and cause
  class.

#### Verification method

Chaos tests (kill, hang, crash-loop, clean-exit-loop); restart-policy property tests
(attempt/backoff sequences); recovery tests over persisted state; leak checks after
repeated restarts (NFR-ARCH-004 pattern).

#### Traceability

ADR-023; INV-PLG-03; chapter 10 machine; Volume 12 resource budgets.

## Configuration keys (`[plugins]`)

| Key | Type | Default | Meaning |
|---|---|---|---|
| `plugins.enabled` | bool | `true` | Master switch for the Plugin Runtime |
| `plugins.handshake_timeout_ms` | int | `10000` | `arp.initialize` budget |
| `plugins.request_timeout_ms` | int | `60000` | Default ARP request budget (tool declarations override per invocation) |
| `plugins.stop_timeout_ms` | int | `5000` | Clean-exit wait before kill escalation |
| `plugins.health_interval_ms` | int | `30000` | `arp.ping` interval |
| `plugins.restart_max_attempts` | int | `5` | Restart attempts per 10-minute window |
| `plugins.restart_backoff_initial_ms` | int | `500` | First backoff; doubles, capped at 30000 |
| `plugins.sources` | array of tables | `[]` | Package sources for plugin discovery/installation (chapter 09 source schema) |
| `[plugins.overrides.<name>]` | table | — | Per-plugin overrides: `autostart`, `idle_stop_minutes`, timeout keys |

## Events minted (plugin.*)

Envelope per Volume 10 (FR-OBS-001).

| Event | Version | Producer | Consumers | Payload summary |
|---|---|---|---|---|
| `plugin.registered` | 1 | Plugin Runtime | TUI, Observability, Audit Log | name, version, scope, declared surfaces summary |
| `plugin.started` | 1 | Plugin Runtime | TUI, Observability | plugin ULID, containment level |
| `plugin.handshake.completed` | 1 | Plugin Runtime | Observability | plugin ULID, negotiated protocol version |
| `plugin.handshake.failed` | 1 | Plugin Runtime | TUI, Observability, Audit Log | plugin ULID, error code, offered/refused versions |
| `plugin.stopped` | 1 | Plugin Runtime | TUI, Observability | plugin ULID, cause class (user, idle, shutdown) |
| `plugin.crashed` | 1 | Plugin Runtime | TUI, Observability | plugin ULID, exit class, invocation count in flight |
| `plugin.restarted` | 1 | Plugin Runtime | Observability | plugin ULID, attempt number, cause class |
| `plugin.disabled` | 1 | Plugin Runtime | TUI, Observability, Audit Log | plugin ULID, actor (user/policy) |
| `plugin.removed` | 1 | Plugin Runtime | TUI, Observability, Audit Log | plugin ULID, name |

## Requirements — non-functional and risk

### NFR-PLUG-001 — Plugin creation time

- Category: Usability
- Priority: P1
- Phase: v1
- Metric: Person-hours to build, install, and invoke a plugin exposing one tool over ARP, from SDK scaffold to first successful invocation (SM-03 definition)
- Target: ≤ 8 person-hours (1 person-day)
- Minimum threshold: ≤ 12 person-hours; breach triggers SDK/documentation remediation before the next phase gate
- Measurement method: timed reference exercise against the SDK plugin template at each phase gate (Beta, v1), executed by a contributor familiar with the language and new to the codebase; SDK tutorial walkthrough verified in CI
- Test environment: Volume 1 reference hardware; documented SDK toolchain
- Measurement frequency: each phase gate; sampled from real contribution records where available
- Owner: Extension SDK / Plugin Runtime (Volume 6)
- Dependencies: FR-PLUG-001, FR-SDK-001
- Risks: RISK-PLUG-002
- Acceptance criteria: The phase-gate report records the timed exercise meeting the target; the CI tutorial walkthrough passes on every release.

### NFR-PLUG-002 — ARP invocation dispatch overhead

- Category: Performance
- Priority: P1
- Phase: Beta
- Metric: Andromeda-added latency per plugin tool invocation round trip — pipeline entry to `arp.tool.execute` write, plus terminal-event read to recorded Tool Result — excluding plugin execution time
- Target: ≤ 10 ms p95
- Minimum threshold: ≤ 25 ms p95; sustained breach triggers the ADR-009 protocol-performance review condition
- Measurement method: instrumented benchmark with a no-op fixture plugin, 1000 invocations, p95, per release (isolates host overhead from plugin work)
- Test environment: Volume 1 reference hardware, both reference machines per Volume 12 formalization
- Measurement frequency: per release in the Volume 12 benchmark suite
- Owner: Plugin Runtime (Volume 6)
- Dependencies: FR-PLUG-001; ADR-009
- Risks: RISK-PLUG-002
- Acceptance criteria: Benchmark report shows p95 within target on both reference machines; breaches of the minimum threshold are release blockers per Volume 12 gating.

### RISK-PLUG-001 — Malicious or compromised plugin

- Category: Security / supply chain
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: two-plane mediation (FR-PLUG-003), manifest permission ceilings, sandbox tiers with recorded containment level, chapter 09 verification for delivery, trust-gated first start, one-way protocol authority (no plugin→host requests)
- Detection: audit-chain review, sandbox violation refusals, resource-limit breaches, drift between manifest and runtime declarations (E-PLUG-007)
- Owner: Plugin Runtime (Volume 6) with Volume 9 threat model
- Status: Open

A plugin is arbitrary code the user chose to run; the design goal is that its reach is
exactly its granted permission set inside its sandbox, and every side effect is
attributable. Residual risk concentrates in what granted permissions legitimately allow —
addressed by least-privilege defaults and the Volume 9 threat entries.

### RISK-PLUG-002 — Protocol evolution breaking the plugin ecosystem

- Category: Compatibility / process
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: version negotiation from day one (FR-PLUG-002), current+previous minor support window, SM-20 discipline for the ARP artifact, conformance fixtures exercised across supported versions, SDK deprecation warnings
- Detection: conformance matrix failures; E-PLUG-002 rates after releases; SDK issue tracker signals
- Owner: Plugin Runtime (Volume 6)
- Status: Open

ARP evolves with the product; the negotiation window plus contract-diff discipline turns
ecosystem breakage into a scheduled, announced event rather than an accident (ADR-009
versioning-debt risk, operationalized).

## Error codes (E-PLUG-001 – E-PLUG-007)

### E-PLUG-001 — Plugin spawn failed

- Category: Execution
- Severity: Error
- User message: "Plugin '<name>' could not be started: <cause summary>."
- Technical message: entrypoint path, sandbox preparation or exec error, permission denial reference where applicable
- Cause: missing/non-executable entrypoint, sandbox refusal, permission denial, resource exhaustion
- Safe-to-log data: plugin name, entrypoint (path), error class, containment level attempted
- Recoverability: recoverable after remediation (reinstall, grant, resources)
- Retry policy: restart policy applies only for transient classes; deterministic classes rest in `failed`
- Recommended action: check the named cause; verify installation integrity (`PackagePort.Verify`)
- Exit-code mapping: 6 when failing an invocation path; 1 for lifecycle commands
- HTTP mapping: not applicable
- Telemetry event: `plugin.handshake.failed`
- Security implications: spawn denials are audit-logged; no partial process survives a failed spawn

### E-PLUG-002 — ARP version negotiation failed

- Category: Compatibility
- Severity: Error
- User message: "Plugin '<name>' requires an ARP version this Andromeda does not support."
- Technical message: host-offered versions, plugin-supported range
- Cause: protocol major/minor mismatch beyond the support window
- Safe-to-log data: plugin name, both version sets
- Recoverability: recoverable with a plugin or Andromeda update
- Retry policy: none (deterministic; excluded from restart policy)
- Recommended action: update the plugin (or Andromeda); consult the plugin's declared `protocol` range
- Exit-code mapping: 1
- HTTP mapping: not applicable
- Telemetry event: `plugin.handshake.failed`
- Security implications: fail-closed negotiation; no degraded protocol operation

### E-PLUG-003 — ARP protocol violation

- Category: Protocol
- Severity: Error
- User message: "Plugin '<name>' violated the runtime protocol and was stopped."
- Technical message: violation class (early output, framing corruption, oversize frame, manifest-inconsistent handshake, duplicate handshake, handshake timeout), offending frame summary (redacted, size-capped)
- Cause: plugin bug or hostile behavior
- Safe-to-log data: plugin name, violation class, frame size
- Recoverability: recoverable by plugin fix; the process is terminated on violation
- Retry policy: restart policy applies except for deterministic handshake violations
- Recommended action: report to the plugin author; run the SDK conformance fixtures against the plugin
- Exit-code mapping: 6
- HTTP mapping: not applicable
- Telemetry event: `plugin.handshake.failed` (handshake phase) or `plugin.crashed` (post-handshake)
- Security implications: violations terminate the process — no best-effort parsing of malformed frames

### E-PLUG-004 — Plugin request timed out

- Category: Timeout
- Severity: Error
- User message: "Plugin '<name>' did not answer within <budget>."
- Technical message: method, effective budget, invocation ULID where applicable, probe-miss counter state
- Cause: hung or overloaded plugin
- Safe-to-log data: plugin name, method, budget, elapsed
- Recoverability: recoverable; unresponsiveness escalation may restart the plugin
- Retry policy: no automatic retry of the timed-out request
- Recommended action: inspect plugin logs (stderr capture); adjust per-plugin budgets if the workload is legitimately slow
- Exit-code mapping: 8
- HTTP mapping: not applicable
- Telemetry event: `plugin.crashed` only on escalation; otherwise within `tool.*` invocation events
- Security implications: abandoned requests never block the pipeline; cancellation is delivered before escalation

### E-PLUG-005 — Plugin process crashed

- Category: Execution
- Severity: Error
- User message: "Plugin '<name>' stopped unexpectedly; affected operations failed."
- Technical message: exit code/signal, in-flight invocation ULIDs, restart-policy state
- Cause: plugin crash, OS kill (resource limits), or unexpected clean exit
- Safe-to-log data: plugin name, exit class, in-flight count, attempt counter
- Recoverability: recoverable via restart policy; in-flight work is failed, never assumed complete
- Retry policy: restart policy (FR-PLUG-004); invocations are not auto-retried
- Recommended action: inspect `last_error` and captured stderr; report crash loops to the author
- Exit-code mapping: 6
- HTTP mapping: not applicable
- Telemetry event: `plugin.crashed`
- Security implications: crash containment at the process boundary; no host state is trusted from a crashed process's partial frames

### E-PLUG-006 — Plugin manifest invalid

- Category: Validation
- Severity: Error
- User message: "Plugin '<name>' has an invalid manifest: <first finding>."
- Technical message: full finding list (schema violations, absolute/traversing entrypoint, unknown permission names, tool-declaration exceeding manifest permission ceiling)
- Cause: malformed or inconsistent `plugin.toml`
- Safe-to-log data: plugin path/name, finding classes and field names
- Recoverability: recoverable by fixing the manifest and re-registering
- Retry policy: none (deterministic; excluded from restart policy)
- Recommended action: validate against the SDK schema; fix listed findings
- Exit-code mapping: 3
- HTTP mapping: not applicable
- Telemetry event: `plugin.handshake.failed`
- Security implications: unknown permission names fail closed — never ignored

### E-PLUG-007 — Undeclared surface registration attempted

- Category: Security
- Severity: Error
- User message: "Plugin '<name>' tried to register a surface it did not declare; registration was refused."
- Technical message: surface kind and name attempted, manifest declaration set
- Cause: plugin declaring surfaces at handshake beyond its manifest (INV-PLG-01)
- Safe-to-log data: plugin name, attempted surface, declared set summary
- Recoverability: recoverable by correcting manifest or plugin; handshake fails
- Retry policy: none (deterministic)
- Recommended action: align the manifest and the plugin's initialize response; re-register
- Exit-code mapping: 6
- HTTP mapping: not applicable
- Telemetry event: `plugin.handshake.failed`
- Security implications: the manifest is the reviewable contract — runtime expansion beyond it is treated as hostile and audit-logged
