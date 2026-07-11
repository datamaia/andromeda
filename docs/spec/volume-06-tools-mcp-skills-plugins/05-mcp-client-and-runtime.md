# 05 — MCP Client and Runtime

Andromeda supports the Model Context Protocol (MCP) as a client. The **MCP Runtime**
(Volume 3, chapter 04) manages **MCP Server** registrations and **MCP Client Connection**
instances (Volume 2, chapter 06), wraps the official `modelcontextprotocol/go-sdk` per
[ADR-010](../annexes/adr/ADR-010.md), and bridges discovered tools, resources, and prompts
into Andromeda's existing surfaces. This chapter specifies registration, installation,
configuration, transports, authorization, discovery, lifecycle operations, health, logs,
update, uninstallation, versioning, compatibility, errors, and timeouts. The trust model,
isolation, supply-chain rules, and conformance testing are in
[chapter 06](06-mcp-security-and-conformance.md); the full MCP Client Connection state
machine is in [chapter 10](10-state-machines.md).

Scope notes:

- Andromeda acts as an MCP **client/host** only. Exposing Andromeda's own tools as an MCP
  server to other hosts is classified **Future** and is not specified by this corpus beyond
  this note.
- The MCP protocol revisions Andromeda pins and certifies are PENDING VALIDATION per ADR-010
  (tracked in this volume's register); revision negotiation itself is delegated to the SDK
  across its supported set.

## Design constraints inherited from the architecture

1. **SDK containment.** MCP SDK types MUST NOT appear in any port signature or in the Core
   Domain (ADR-010 rule 1; Volume 3 chapter 02, "No leakage"). The MCP Runtime is the only
   component that imports the SDK.
2. **Tools are ordinary tools.** Every tool discovered on an MCP server enters the Tool
   Runtime registry as a Tool row with `origin = mcp` and `origin_ref` pointing at the MCP
   Server registration (Volume 2, INV-MCPS-03). The invocation pipeline — validate,
   permission, sandbox placement, execute, record — applies identically to MCP-origin tools
   (FR-TOOL-001; Principle 4). The MCP Runtime provides ToolPort bridge implementations; it
   never executes tool logic outside that contract.
3. **Sandboxed transports.** stdio MCP servers are child processes and MUST be launched
   exclusively through SandboxPort under the MCP-server sandbox tier (FR-SEC-101, ADR-021);
   remote servers fall under the Volume 9 network policy.
4. **Secrets by reference.** Server credentials are Credential entities resolved through
   SecretStorePort at connection time; secret material MUST NOT appear in `[mcp]`
   configuration, launch specifications, logs, or events (Volume 2, INV-MCPS-02; FR-SEC-102).

## Registration and configuration

MCP servers are registered declaratively in configuration or imperatively through the CLI
(`mcp` command group; grammar owned by Volume 8). Both paths produce an MCP Server row
(scope `global` or `workspace`) and an Extension record of kind `mcp_server` (Volume 2).

Configuration lives in the `[mcp]` table (content owned by this volume; schema, precedence,
and validation by Volume 10 per FR-CFG-001):

```toml
[mcp]
# Runtime-wide defaults; every value overridable per server.
connect_timeout_ms = 10000     # transport establishment budget
initialize_timeout_ms = 10000  # MCP initialize/negotiation budget
request_timeout_ms = 60000     # per-request budget (tool calls, listings)
reconnect_max_attempts = 5     # automatic reconnection attempts before giving up
reconnect_backoff_initial_ms = 1000  # doubles per attempt, capped at 30000
log_capture = true             # capture MCP logging notifications into local logs

[mcp.servers.docs]
transport = "stdio"
command = "docs-mcp-server"
args = ["--workspace-mode"]
enabled = true

[mcp.servers.tracker]
transport = "streamable_http"
url = "https://tracker.example.com/mcp"
credential = "tracker-token"   # Credential reference name; never the secret itself
enabled = false
```

The following registration is invalid — it embeds secret material and mixes transport
fields — and MUST be rejected with E-MCP-007 at validation time:

```toml invalid
[mcp.servers.bad]
transport = "stdio"
url = "https://example.com/mcp"        # url is not a stdio field
headers = { Authorization = "Bearer sk-live-1234" }  # secret material inline
```

Config keys minted in this chapter (`[mcp]` table): `connect_timeout_ms`,
`initialize_timeout_ms`, `request_timeout_ms`, `reconnect_max_attempts`,
`reconnect_backoff_initial_ms`, `log_capture`, and per-server subtables
`[mcp.servers.<name>]` with `transport`, `command`, `args`, `env_allowlist`, `url`,
`headers`, `credential`, `enabled`, `scope_hint`, plus per-server overrides of every
runtime-wide default above.

### Installation

"Installing an MCP server" means one of:

1. **Registration only** — the server binary or endpoint already exists; the user registers
   it (configuration edit or CLI). No package operation occurs.
2. **Packaged installation** — the server is delivered as an extension package (kind
   `plugin` or `bundle` containing the server executable, or a registration-only manifest);
   it flows through the Package Manager's frozen installation states
   ([chapter 09](09-package-manager-supply-chain.md)) and requires the
   `package_installation` permission.

In both paths, first connection is gated by the trust decisions of chapter 06.

### FR-MCP-001 — MCP client support

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: MCP Runtime (Volume 6)
- Affected components: MCP Runtime, Tool Runtime, Sandbox Engine, Secret Store, Configuration Manager
- Dependencies: ADR-010, ADR-021; FR-TOOL-001, FR-SEC-101, FR-SEC-102, FR-CFG-001
- Related risks: RISK-MCP-001, RISK-MCP-002

#### Description

Andromeda MUST operate as an MCP client using the official `modelcontextprotocol/go-sdk`
pinned within its stable v1 major line (ADR-010). The MCP Runtime MUST support: server
registration at `global` and `workspace` scopes; the `stdio` and `streamable_http`
transports; protocol revision negotiation delegated to the SDK across its supported
revision set; discovery of tools, resources, and prompts; bridging of discovered surfaces
into the Tool Runtime, Context Manager, and prompt-selection surfaces; managed connection
lifecycle with health monitoring and reconnection; and structured errors and timeouts for
every operation. The concrete protocol revision set Andromeda certifies is PENDING
VALIDATION (ADR-010 rule 3).

#### Motivation

MCP is the ecosystem-standard way to attach third-party capability to agent hosts;
first-class client support realizes PRD-007 without inventing a proprietary equivalent.

#### Actors

Users registering servers; agents invoking bridged tools; the MCP Runtime; MCP servers.

#### Preconditions

A server registration exists and is `enabled`; for stdio, the launch command resolves; for
streamable HTTP, the URL is well-formed.

#### Main flow

1. The MCP Runtime reads registrations from resolved configuration (ConfigPort).
2. On demand (first use or explicit connect), it establishes the transport and runs MCP
   initialization, negotiating protocol revision and capabilities.
3. It discovers offered tools/resources/prompts and bridges them per FR-MCP-003.
4. Requests from Andromeda surfaces are routed over the `ready` connection with per-request
   timeouts; results and errors map to Andromeda envelopes.

#### Alternative flows

- Connection loss: automatic reconnection per the retry policy of the MCP Client Connection
  machine (chapter 10); requests during reconnection fail with E-MCP-004.
- Server disabled mid-session: connections terminate; bridged tools are disabled
  (INV-MCPS-04).

#### Edge cases

- A server offering zero tools/resources/prompts is valid; discovery records empty
  surfaces.
- Duplicate registration names within one scope are rejected at validation (INV-MCPS-01).
- A server reporting a protocol revision outside the SDK's supported set fails
  initialization with E-MCP-002; the connection records `failed`.

#### Inputs

Server registrations (`[mcp]`), connect/disconnect requests, MCP protocol messages,
credential references.

#### Outputs

MCP Client Connection rows with negotiated protocol version and capabilities; bridged
ToolPort implementations; discovery summaries (`discovered_surfaces` on the MCP Server
row); `mcp.*` events.

#### States

MCP Client Connection canonical states (`configured`, `connecting`, `initializing`,
`ready`, `reconnecting`, `disconnected`, `failed`, `disabled`, `removed`); full machine in
chapter 10.

#### Errors

E-MCP-001 through E-MCP-007 (this chapter); E-MCP-008 (chapter 06). Sandbox and permission
refusals surface in the E-SEC family.

#### Constraints

SDK containment (ADR-010 rule 1); launches only through SandboxPort; no OAuth dependence
at Beta (FR-MCP-004); one managed connection per server per process (INV-MCPC-02).

#### Security

All chapter 06 requirements apply: trust gating before exposure, sandbox tiers per
transport, secrets by reference only, descriptor drift detection.

#### Observability

Connection lifecycle transitions emit `mcp.*` events with correlation IDs (envelope per
FR-OBS-001); per-server request counters and latency metrics; discovery inventories
queryable via CLI.

#### Performance

Timeout defaults per the `[mcp]` table; conformance and interop targets per NFR-MCP-001
and NFR-MCP-002; connection establishment does not block session startup (connections are
lazy or background tasks under SchedulerPort).

#### Compatibility

Protocol revision negotiation across the SDK's supported set; revision pin PENDING
VALIDATION (ADR-010). Registrations are forward-portable configuration; the `transport`
vocabulary (`stdio`, `streamable_http`) follows Volume 2 and extends only through the
change procedure.

#### Acceptance criteria

- Given a registered, enabled stdio server, when a session first requests one of its
  tools, then the runtime launches it through SandboxPort, completes initialization,
  bridges the tool, and the invocation succeeds with the result recorded as a Tool Result.
- Given a streamable HTTP server with a credential reference, when connecting, then the
  credential is resolved through SecretStorePort and never appears in logs, events, or
  errors (verified by log scan).
- Given a server whose process is killed mid-session, when the next request is issued,
  then the connection transitions `ready` → `reconnecting`, reconnection follows the
  backoff policy, and the request fails with E-MCP-004 rather than hanging.
- Negative case: given a registration with `transport = "stdio"` and a `url` field, when
  configuration validates, then E-MCP-007 is reported and no connection is attempted.
- Permission case: given a stdio server whose launch is denied by the Permission Manager,
  when connection is attempted, then the connection records `failed` with the E-SEC denial
  surfaced and an Approval record exists.
- Observability case: every state transition of the connection emits exactly one `mcp.*`
  event carrying the connection ULID and server name.

#### Verification method

MCP conformance suite (FR-MCP-007) and recorded-session tests per revision; interop job
against the public reference-server set (SM-15); integration tests for sandbox launch,
credential resolution, and reconnection; log-scan tests for secret absence.

#### Traceability

PRD-002, PRD-004, PRD-007; ADR-010, ADR-021; SM-15; FR-TOOL-001; chapter 06; chapter 10.

## Transports

### FR-MCP-002 — MCP transports and connection establishment

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Derived
- Owner: MCP Runtime (Volume 6)
- Affected components: MCP Runtime, Sandbox Engine, Terminal Engine (PTY excluded — pipes only), Platform Abstraction Layer
- Dependencies: FR-MCP-001; ADR-010, ADR-021, ADR-023
- Related risks: RISK-MCP-001

#### Description

The MCP Runtime MUST support exactly two transports at Beta, matching the Volume 2
vocabulary: `stdio` (server as a sandboxed child process; JSON-RPC messages on
stdin/stdout per the MCP stdio transport) and `streamable_http` (server at an HTTPS URL
per the MCP Streamable HTTP transport). Transport establishment MUST respect
`connect_timeout_ms`; MCP initialization MUST respect `initialize_timeout_ms`; both
timeouts MUST surface as E-MCP-001 with the phase recorded. stdio servers MUST be spawned
through SandboxPort with a deny-by-default environment (`env_allowlist` names the only
passed-through variables) and MUST have their process trees terminated on disconnect,
teardown, or shutdown. `streamable_http` connections MUST use TLS except for loopback
addresses; the `network` permission governs non-loopback connections per Volume 9 policy.

#### Motivation

Two transports cover the documented MCP ecosystem paths without inventing undocumented
mechanisms; sandbox-mediated spawning keeps third-party server processes inside the
containment model.

#### Actors

MCP Runtime; Sandbox Engine; remote and local MCP servers.

#### Preconditions

Registration validated; for stdio, permission to spawn (`process_spawn`) resolved; for
non-loopback HTTP, `network` permission resolved.

#### Main flow

1. `connecting`: the transport is established (child process spawn or HTTP session).
2. `initializing`: MCP initialize request/response; revision and capability negotiation.
3. `ready`: requests are served; health per FR-MCP-005.

#### Alternative flows

- Spawn failure (missing executable, sandbox refusal): E-MCP-001 with cause; state
  `failed`.
- TLS or HTTP-level failure: E-MCP-001 with the transport error class; state `failed`.

#### Edge cases

- A stdio server writing non-protocol output to stdout corrupts framing: the connection
  fails with E-MCP-006 and the raw bytes are captured (redacted) for diagnostics; stderr
  is treated as log output, never protocol.
- Loopback HTTP without TLS is permitted; a non-loopback `http://` URL is a validation
  error (E-MCP-007).
- Environment variables not in `env_allowlist` never reach a stdio server, including
  `ANDROMEDA_*` values.

#### Inputs

Validated registration; sandbox policy for the MCP-server tier; credential reference where
required.

#### Outputs

An established transport bound to one MCP Client Connection row; process handle (stdio)
under sandbox management.

#### States

`connecting` and `initializing` per the chapter 10 machine.

#### Errors

E-MCP-001 (establishment/timeout), E-MCP-002 (negotiation), E-MCP-006 (protocol
violation), E-MCP-007 (registration invalid); E-SEC family for permission/sandbox refusal.

#### Constraints

No transport other than the two named; no direct process spawning outside SandboxPort;
bounded buffers on message reading (ADR-023).

#### Security

Child environment filtering; TLS enforcement; header values from configuration are
non-secret by definition (secrets only via `credential`); the effective containment level
of stdio servers is recorded per execution (ADR-021).

#### Observability

Transport establishment and failure events; connect/initialize latency metrics per
server.

#### Performance

Defaults: `connect_timeout_ms` 10000, `initialize_timeout_ms` 10000. Establishment runs as
a supervised task and never blocks the TUI thread.

#### Compatibility

Transport vocabulary is closed at Beta; additional documented MCP transports enter only
through the change procedure with a new ADR.

#### Acceptance criteria

- Given a stdio registration, when connected, then the server process runs under a sandbox
  handle whose recorded environment contains only `env_allowlist` variables.
- Given a non-loopback `http://` URL, when configuration validates, then E-MCP-007 is
  reported before any network activity.
- Given a server that accepts the TCP connection but never answers initialize, when
  `initialize_timeout_ms` elapses, then the connection records `failed` with E-MCP-001
  (phase `initializing`) and the child process (stdio) is terminated.
- Error case: given a stdio command that does not exist, when connecting, then E-MCP-001
  reports the spawn failure and no retry occurs without an explicit reconnect request.
- Observability case: connect latency for each successful establishment is recorded as a
  metric sample tagged with the server name.

#### Verification method

Integration tests with fixture servers for both transports; fault-injection (hang at each
phase, garbage on stdout, TLS failures); sandbox-policy assertion tests; conformance suite
transport sections.

#### Traceability

FR-MCP-001; ADR-010, ADR-021; Volume 2 INV-MCPS-02; chapter 06 isolation rules.

## Authorization and credentials

### FR-MCP-004 — MCP server authorization

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: MCP Runtime (Volume 6)
- Affected components: MCP Runtime, Authentication Layer, Secret Store
- Dependencies: FR-MCP-001, FR-AUTH-001, FR-SEC-102; ADR-010, ADR-014
- Related risks: RISK-MCP-001

#### Description

For servers requiring authorization, Andromeda MUST support static token/header
authorization at Beta: the registration names a Credential (`credential` key); at
connection time the MCP Runtime resolves it through SecretStorePort and applies it to the
transport (HTTP `Authorization` or server-documented header). OAuth-based MCP server
authorization is PENDING VALIDATION per ADR-010 rule 4 (the SDK's client OAuth is
experimental) and MUST NOT be exposed as a stable feature until that validation resolves;
registrations demanding OAuth fail with E-MCP-005 carrying the pending status. Use of a
credential requires the `credential_access` permission under Volume 9 decision semantics.
Credential material MUST never be written to configuration, logs, events, errors, or the
`discovered_surfaces` summary.

#### Motivation

Official mechanisms only (FR-AUTH-001): token/header auth is documented and stable;
OAuth's client side is not yet stable in the official SDK, and pretending otherwise would
overstate guarantees.

#### Actors

Users configuring credentials; MCP Runtime; Secret Store; remote servers.

#### Preconditions

Credential exists with `status = active`; permission decision available.

#### Main flow

1. Connection establishment reaches the authorization point.
2. The runtime requests `credential_access` (PermissionPort), resolves the reference, and
   applies the material to the transport.
3. Material is released (zeroized wrapper semantics per SecretStorePort) after transport
   configuration.

#### Alternative flows

- Server rejects the credential (HTTP 401/403 or MCP-level auth error): E-MCP-005,
  connection `failed`; the Credential row is untouched (rotation is a user/AuthPort
  action).
- Credential reference missing or revoked: E-MCP-005 with the resolution failure cause; no
  connection attempt.

#### Edge cases

- stdio servers requiring secrets receive them only through `env_allowlist`-named
  variables whose values are credential references resolved at spawn — never literal
  secrets in `args`.
- A credential rotated mid-session takes effect on the next (re)connection; in-flight
  requests complete under the prior transport session.

#### Inputs

Credential reference; permission decision; server auth expectations.

#### Outputs

Authorized transport session; audit records for credential access (Volume 9).

#### States

Authorization occurs inside `connecting`/`initializing`; failure lands in `failed`.

#### Errors

E-MCP-005; E-SEC family for permission denial; E-AUTH family untouched (provider auth is
Volume 5's).

#### Constraints

No OAuth flows at Beta (PENDING VALIDATION); no plaintext secrets anywhere (ADR-014); a
subscription to a service never implies programmatic access.

#### Security

Every credential resolution is audit-logged; redaction rules of Volume 9 apply to all MCP
diagnostics; secret absence from all sinks is a tested property.

#### Observability

`mcp.connection.failed` events carry the failure class (`authorization`) without secret
context; credential-access audit records correlate to the connection ULID.

#### Performance

Credential resolution is local-only (SecretStorePort) and adds no network round trip.

#### Compatibility

When SDK OAuth graduates, a superseding decision under this volume's ADR block or a
Volume 5 flow extension activates it; registrations remain unchanged (auth mechanism is a
per-server declaration).

#### Acceptance criteria

- Given a registration naming a valid Credential, when connecting, then the request
  carries the material, the connection reaches `ready`, and no sink contains the secret
  (log/event/error scan).
- Given a revoked Credential, when connecting, then E-MCP-005 reports resolution failure
  and no transport connection is opened.
- Permission case: given `credential_access` denied for the session, when connecting, then
  the connection fails with the E-SEC denial recorded and no credential material is read.
- Negative case: given a server demanding OAuth, when connecting, then E-MCP-005 reports
  the mechanism as unavailable pending validation, with the register reference.

#### Verification method

Integration tests with authorized fixture servers; secret-scan assertions over logs,
events, and error payloads; permission-denial tests; register cross-check for the OAuth
PENDING VALIDATION entry.

#### Traceability

FR-AUTH-001, FR-SEC-102; ADR-010 rule 4, ADR-014; Volume 9 audit semantics.

## Discovery and bridging

### FR-MCP-003 — Discovery and bridging of tools, resources, and prompts

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: MCP Runtime (Volume 6)
- Affected components: MCP Runtime, Tool Runtime, Context Manager, Prompt Engine (consumer), TUI/CLI
- Dependencies: FR-MCP-001, FR-TOOL-001; ADR-024, ADR-077
- Related risks: RISK-MCP-001, RISK-MCP-002

#### Description

After a connection reaches `ready`, the MCP Runtime MUST discover the server's offered
surfaces and bridge them:

1. **Tools** — each discovered tool registers as a Tool row (`origin = mcp`) named
   `mcp:<server>/<tool>` per ADR-077, carrying the server-declared input schema (validated
   as JSON Schema per ADR-024), description, and the server's trust context. Exposure to
   agents is gated by chapter 06 trust rules. Tool list change notifications from the
   server MUST trigger re-discovery and re-application of the drift rules (E-MCP-008).
2. **Resources** — resources and resource templates are enumerated and offered to the
   Context Manager as candidate Context Items with provenance `mcp:<server>`; content is
   fetched on selection, subject to `request_timeout_ms` and size limits from context
   policy (Volume 7 budgeting authority).
3. **Prompts** — server prompts are listed as user-invocable prompt entries under the
   `mcp:<server>/` namespace in the CLI/TUI prompt-selection surfaces; prompt arguments
   are validated against the server's declared argument list before sending.

Discovery results are summarized into the MCP Server row's `discovered_surfaces`
attribute. Bridged surfaces MUST disappear (tools disabled, resources/prompts delisted)
when the connection leaves `ready` or the server is disabled/removed (INV-MCPS-04).

#### Motivation

Bridging into existing surfaces (Tool Runtime, Context Manager, prompt selection) keeps
MCP a source of capability rather than a second execution system — one contract, one
permission model, one observability regime (Principle 4).

#### Actors

MCP Runtime; Tool Runtime; Context Manager; users browsing surfaces.

#### Preconditions

Connection `ready`; trust decisions for exposure resolved per chapter 06.

#### Main flow

1. List tools/resources/prompts (paginated per protocol).
2. Validate declared schemas; reject invalid declarations per edge cases.
3. Register bridges; emit `mcp.surfaces.discovered`; update `discovered_surfaces`.

#### Alternative flows

- Server sends `listChanged` notifications: re-discover, diff against pinned descriptors,
  apply drift policy (chapter 06), update registrations atomically.
- Discovery request times out: connection remains `ready`; discovery retries once, then
  records E-MCP-003 and leaves prior surfaces in force.

#### Edge cases

- A tool whose input schema does not parse as JSON Schema is not registered; E-MCP-006 is
  recorded against that tool only, other surfaces proceed.
- Name collisions are impossible across servers by construction (`mcp:<server>/` prefix);
  a server offering two tools with the same name is a protocol violation → E-MCP-006.
- Resource content exceeding the context size limit is truncated per Volume 7 rules with
  the truncation marked, never silently.

#### Inputs

MCP list/read/get responses; trust policy verdicts; context budgets.

#### Outputs

Tool rows (`origin = mcp`), Context Item candidates, prompt listings, discovery events and
summaries.

#### States

Bridged surfaces exist only while the connection is `ready` (INV-MCPC-03 for routing).

#### Errors

E-MCP-003 (timeout), E-MCP-006 (invalid declarations), E-MCP-008 (descriptor drift,
chapter 06).

#### Constraints

No invocation before registration completes; schemas validated before first use; the
`mcp:` namespace is reserved for MCP-origin surfaces (ADR-077).

#### Security

Tool descriptions and resource content are untrusted input (indirect prompt-injection
vector); chapter 06 requires provenance labeling wherever they are shown to users or
included in model context.

#### Observability

`mcp.surfaces.discovered` carries per-kind counts; per-tool registration events flow
through the Tool Runtime's `tool.*` family (Volume 6 chapter 02 rules).

#### Performance

Discovery of a server offering 100 tools completes within `request_timeout_ms`; bridging
is O(surfaces) with no per-item network round trips beyond protocol listings.

#### Compatibility

Pagination and capability-dependent listings follow the negotiated protocol revision;
servers lacking a surface kind simply contribute none.

#### Acceptance criteria

- Given a server offering tools, resources, and prompts, when discovery completes, then
  each tool is invocable as `mcp:<server>/<tool>`, resources appear as context candidates
  with provenance, prompts are listed under the server namespace, and
  `discovered_surfaces` reflects exact counts.
- Given a connection that drops to `reconnecting`, when an agent requests a bridged tool,
  then the invocation fails with E-MCP-004 and the Tool row shows disabled status.
- Negative case: given a tool declaring a malformed input schema, when discovery runs,
  then that tool is absent from the registry, E-MCP-006 is recorded naming the tool, and
  remaining tools register normally.
- Permission case: given trust policy that blocks exposure for an unapproved server, when
  discovery completes, then zero tools are agent-visible and the block is recorded per
  chapter 06.
- Observability case: re-discovery after `listChanged` emits `mcp.surfaces.discovered`
  with a `revision` counter incremented.

#### Verification method

Conformance fixtures covering listings, pagination, and `listChanged`; schema-validation
unit tests; integration tests asserting bridge teardown on state exits; interop suite
(SM-15b).

#### Traceability

FR-TOOL-001; ADR-024, ADR-077; Volume 2 INV-MCPS-03/04; Volume 7 context budgeting;
chapter 06 trust rules.

## Lifecycle operations, health, and logs

### FR-MCP-005 — Connection health, server logs, and maintenance operations

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: MCP Runtime (Volume 6)
- Affected components: MCP Runtime, Logging, Observability, Package Manager
- Dependencies: FR-MCP-001, FR-MCP-002; ADR-011
- Related risks: RISK-MCP-002

#### Description

The MCP Runtime MUST provide, per server: **health** — liveness verification using
protocol pings where the negotiated revision supports them, otherwise a lightweight
listing probe, on a fixed interval (default 60 s, configurable per server); three
consecutive failures transition the connection per the chapter 10 machine. **Logs** — MCP
logging notifications are captured into Andromeda's structured logs (slog per ADR-011)
tagged `mcp.server=<name>`, subject to Volume 9 redaction; capture is enabled by
`log_capture` and the log level requested from the server follows the effective local log
level. **Update** — for packaged servers, updates flow through the Package Manager
(chapter 09) and MUST NOT occur automatically without consent; for registration-only
servers, Andromeda re-reads configuration at defined reconfiguration points and applies
changes by reconnecting. **Uninstall** — removing a registration terminates connections,
disables bridged tools, tombstones the MCP Server row and its Extension record; packaged
servers additionally flow through Package removal. **Versioning/compatibility** — the
negotiated protocol revision and the server-reported implementation version are recorded
on the connection row and surfaced in diagnostics; a server whose revision falls outside
the supported set after an update fails with E-MCP-002 and the prior state is reported.

#### Motivation

Third-party servers fail, log, and change versions; managed operations make those events
observable and recoverable instead of silent.

#### Actors

MCP Runtime; users running maintenance commands; Package Manager.

#### Preconditions

Server registered; for update/uninstall of packaged servers, `package_installation`
permission resolved.

#### Main flow

1. Health probes run as scheduled supervised tasks while `ready`.
2. Log notifications stream into local logs with redaction.
3. Update/uninstall commands drive Package Manager or configuration changes; connections
   restart or terminate accordingly.

#### Alternative flows

- Health probe failure below threshold: recorded, no transition.
- Uninstall while a bridged tool is executing: the invocation is cancelled through the
  Tool Runtime (`cancelled` outcome) before teardown completes.

#### Edge cases

- A server flooding log notifications is throttled by a bounded buffer; overflow drops
  oldest with a counter (ADR-023 backpressure), never blocking protocol traffic.
- Health probing a server that lacks ping support MUST NOT mark it failed for that lack
  alone (probe falls back to listing).

#### Inputs

Probe schedule, log notifications, update/uninstall requests, configuration changes.

#### Outputs

Health status on the connection row (`stats`), captured logs, updated/removed
registrations, events.

#### States

Full machine in chapter 10; maintenance operations enter through `disconnected`/`removed`
transitions.

#### Errors

E-MCP-003 (probe timeout contributes to failure count), E-MCP-002 (post-update revision
mismatch); package errors per chapter 09.

#### Constraints

No automatic server updates; log capture never includes credential material; probes are
lightweight (single request, no discovery).

#### Security

Captured logs are third-party content: redaction before persistence; log injection is
neutralized by structured encoding (no raw terminal control sequences re-emitted).

#### Observability

`mcp.health.checked` sampled metrics (not evented per probe); `mcp.connection.lost` /
`mcp.connection.failed` on transitions; update/uninstall audit records.

#### Performance

Probe cost ≤ 1 request per interval per server; log capture overhead bounded by the
buffer size (default 1000 entries).

#### Compatibility

Ping availability varies by negotiated revision — the fallback probe keeps health
monitoring revision-independent.

#### Acceptance criteria

- Given a `ready` connection whose server stops responding, when three consecutive probes
  fail, then the connection transitions per chapter 10 and `mcp.connection.lost` is
  emitted exactly once.
- Given `log_capture = true`, when the server emits a logging notification containing a
  token-shaped string, then the persisted log entry shows the redacted form.
- Given an uninstall command for a packaged server, when it completes, then the package is
  `removed`, the MCP Server row is tombstoned, connections are terminal, and previously
  bridged tools are disabled with provenance retained.
- Negative case: given an update that changes the server to an unsupported protocol
  revision, when reconnection runs, then E-MCP-002 is reported and diagnostics name both
  revisions.
- Permission case: uninstalling a packaged server without `package_installation` fails
  with the E-SEC denial and changes nothing.

#### Verification method

Fault-injection tests (probe failures, log floods); update/uninstall integration tests
against fixture packages; redaction assertion tests; conformance suite health sections.

#### Traceability

ADR-011; chapter 09 package operations; chapter 10 machine; Volume 9 redaction.

## Timeouts summary

| Operation | Budget (default) | On expiry |
|---|---|---|
| Transport establishment | `connect_timeout_ms` = 10000 | E-MCP-001, state `failed` |
| MCP initialization | `initialize_timeout_ms` = 10000 | E-MCP-001, state `failed` |
| Any request (tool call, listing, resource read) | `request_timeout_ms` = 60000, per-server override, per-invocation override from tool timeout declarations | E-MCP-003; invocation `timed_out` where applicable |
| Reconnection attempts | `reconnect_max_attempts` = 5, backoff initial 1000 ms, doubling, cap 30000 ms | state `disconnected` (chapter 10) |
| Health probe | interval 60 s; probe budget = `request_timeout_ms` | failure counter; 3 consecutive → transition |

All budgets honor context cancellation (FR-ARCH-004): user interrupts abort in-flight MCP
requests and, for stdio, terminate the server process tree via sandbox teardown when the
connection is being torn down.

## Events minted (mcp.*)

Envelope, ordering, delivery, persistence, retention, privacy, and failure behavior per
Volume 10 (FR-OBS-001). Payloads listed are the safe-context fields; all carry the
connection ULID and server name where applicable.

| Event | Version | Producer | Consumers | Payload summary |
|---|---|---|---|---|
| `mcp.server.registered` | 1 | MCP Runtime | TUI, Observability, Audit Log | server name, scope, transport |
| `mcp.server.updated` | 1 | MCP Runtime | TUI, Observability | server name, changed field names (values omitted) |
| `mcp.server.removed` | 1 | MCP Runtime | TUI, Observability, Audit Log | server name, scope |
| `mcp.connection.established` | 1 | MCP Runtime | TUI, Observability | connection ULID, negotiated protocol revision, server version string |
| `mcp.connection.lost` | 1 | MCP Runtime | TUI, Observability | connection ULID, failure class |
| `mcp.connection.failed` | 1 | MCP Runtime | TUI, Observability, Audit Log | connection ULID, error code, phase |
| `mcp.surfaces.discovered` | 1 | MCP Runtime | TUI, Observability | connection ULID, tool/resource/prompt counts, discovery revision |
| `mcp.request.failed` | 1 | MCP Runtime | Observability | connection ULID, method class, error code (no request content) |
| `mcp.log.received` | 1 | MCP Runtime | Logging | server name, level (message body goes to logs, not the event) |

## Error codes (E-MCP-001 – E-MCP-007)

### E-MCP-001 — MCP connection establishment failed

- Category: Connectivity
- Severity: Error
- User message: "Could not connect to MCP server '<name>': <cause summary>."
- Technical message: transport, phase (connect/initialize), underlying error class, timeout budget if expired
- Cause: spawn failure, network/TLS failure, initialize timeout, sandbox refusal surfaced from E-SEC
- Safe-to-log data: server name, transport, phase, error class, latency
- Recoverability: recoverable (fix registration, environment, or server; reconnect)
- Retry policy: automatic only within the reconnection policy after a prior `ready`; manual retry otherwise
- Recommended action: verify the command/URL and server health; consult captured server logs
- Exit-code mapping: 6 when it fails a tool invocation path; 1 for standalone connect commands
- HTTP mapping: not applicable
- Telemetry event: `mcp.connection.failed`
- Security implications: none beyond containment already applied; spawn refusals are audit-logged

### E-MCP-002 — MCP protocol negotiation failed

- Category: Compatibility
- Severity: Error
- User message: "MCP server '<name>' speaks an unsupported protocol revision."
- Technical message: server-proposed revision, SDK-supported set, negotiation error detail
- Cause: server revision outside the SDK-supported set, or malformed negotiation exchange
- Safe-to-log data: server name, both revision identifiers
- Recoverability: recoverable with a server or Andromeda update
- Retry policy: none (deterministic)
- Recommended action: update the server or Andromeda; check the pinned revision set note in this volume's register
- Exit-code mapping: 6 on invocation paths; 1 standalone
- HTTP mapping: not applicable
- Telemetry event: `mcp.connection.failed`
- Security implications: refusing negotiation is fail-closed; no partial protocol operation

### E-MCP-003 — MCP request timed out

- Category: Timeout
- Severity: Error
- User message: "MCP server '<name>' did not answer within <budget>."
- Technical message: method class, effective budget, connection state, correlation ID
- Cause: slow or hung server; network stall
- Safe-to-log data: server name, method class, budget, elapsed
- Recoverability: recoverable; connection health may transition per probe policy
- Retry policy: no automatic retry of the same request; idempotent listings retried once during discovery
- Recommended action: inspect server health; raise the per-server budget if the workload is legitimately slow
- Exit-code mapping: 8
- HTTP mapping: not applicable
- Telemetry event: `mcp.request.failed`
- Security implications: hung-request abandonment never leaks the request content into diagnostics

### E-MCP-004 — MCP connection not ready

- Category: State
- Severity: Error
- User message: "MCP server '<name>' is not connected; the request was not sent."
- Technical message: current connection state, requested method class
- Cause: request routed while the connection is outside `ready` (INV-MCPC-03)
- Safe-to-log data: server name, state, method class
- Recoverability: recoverable once the connection returns to `ready`
- Retry policy: caller-driven after reconnection; no queueing of requests
- Recommended action: reconnect or await automatic reconnection; check `mcp.connection.*` events
- Exit-code mapping: 6
- HTTP mapping: not applicable
- Telemetry event: `mcp.request.failed`
- Security implications: fail-closed routing guarantees no request rides a half-established transport

### E-MCP-005 — MCP authorization failed

- Category: Authorization
- Severity: Error
- User message: "Authorization to MCP server '<name>' failed."
- Technical message: mechanism (token/header, or unavailable-pending-validation for OAuth), resolution or server rejection detail (redacted)
- Cause: missing/revoked credential reference, server-side rejection, or an OAuth-only server at Beta
- Safe-to-log data: server name, mechanism, failure class; never material
- Recoverability: recoverable (fix credential, rotate, or await OAuth validation)
- Retry policy: none automatic
- Recommended action: verify the Credential reference and its status; rotate via auth commands
- Exit-code mapping: 4
- HTTP mapping: 401/403 recorded as cause class when the transport reported one
- Telemetry event: `mcp.connection.failed`
- Security implications: credential access is audit-logged; rejection details are redacted before persistence

### E-MCP-006 — MCP protocol violation by server

- Category: Protocol
- Severity: Error
- User message: "MCP server '<name>' returned an invalid response; the item was rejected."
- Technical message: violation class (framing corruption, malformed schema, duplicate names, envelope violation), offending item name where safe
- Cause: server bug or hostile server behavior
- Safe-to-log data: server name, violation class, item name
- Recoverability: item-level rejections recoverable by server fix; framing corruption requires reconnect
- Retry policy: none (deterministic against the same response)
- Recommended action: report to the server maintainer; consider disabling the server
- Exit-code mapping: 6
- HTTP mapping: not applicable
- Telemetry event: `mcp.request.failed`
- Security implications: invalid declarations never register; treated as untrusted input per chapter 06

### E-MCP-007 — MCP server registration invalid

- Category: Configuration
- Severity: Error
- User message: "MCP server registration '<name>' is invalid: <finding>."
- Technical message: field-level findings (transport/field mismatch, secret material detected inline, duplicate name, malformed URL)
- Cause: invalid `[mcp.servers.<name>]` content
- Safe-to-log data: server name, field names, finding classes (values redacted where flagged as secret-shaped)
- Recoverability: recoverable by configuration fix
- Retry policy: none
- Recommended action: correct the registration; move secrets to a Credential reference
- Exit-code mapping: 3
- HTTP mapping: not applicable
- Telemetry event: `mcp.server.updated` (with validation-failed marker in logs, not the event)
- Security implications: inline secret detection prevents plaintext credential persistence (INV-MCPS-02)
