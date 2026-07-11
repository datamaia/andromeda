# 06 — MCP Security and Conformance

MCP servers are third-party code and third-party content by definition (Volume 2, MCP
Server `trust_level` note). This chapter defines the MCP trust model, isolation
requirements, supply-chain rules, and the conformance test program that backs SM-15. It
builds on [chapter 05](05-mcp-client-and-runtime.md) and binds to the permission model
(FR-SEC-100), sandbox specification (FR-SEC-101), and secret management (FR-SEC-102) owned
by Volume 9.

## Trust model

Principles:

1. **Third-party by default.** Every MCP server carries a trust classification (vocabulary
   owned by Volume 9) recorded at registration; no MCP server is ever classified at the
   built-in level.
2. **Consent before exposure.** A newly registered server contributes nothing to agents
   until the user has approved its exposure (an Approval under Volume 9 semantics,
   requestable at registration or first connection). The approval covers the discovered
   surface inventory shown to the user at decision time.
3. **Descriptor pinning.** At approval time the runtime pins a SHA-256 digest of each
   exposed tool's descriptor (name, description, input schema). Later discovery that
   changes a pinned descriptor suspends that tool's exposure and requires re-approval
   (E-MCP-008). This is the defense against post-approval behavior swaps ("rug pulls") and
   description-based tool poisoning.
4. **Untrusted content labeling.** Tool descriptions, tool results, resource content, and
   prompt text from MCP servers are untrusted input. Wherever they enter model context or
   user-facing surfaces, provenance `mcp:<server>` MUST be attached (Context Item
   provenance per Volume 7; TUI labeling per Volume 8) so that injection attempts are
   attributable and filterable.
5. **Least exposure.** Configuration MAY restrict exposure per server to named tools
   (`expose_tools` allowlist in the server registration); default is all-discovered,
   post-approval.

Denied or not-yet-approved servers still connect for *inventory* purposes when the user
runs discovery commands explicitly; their surfaces are never agent-visible.

### FR-MCP-006 — MCP trust gating and isolation

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Design
- Owner: MCP Runtime (Volume 6)
- Affected components: MCP Runtime, Permission Manager, Sandbox Engine, Policy Engine, Audit Log
- Dependencies: FR-MCP-001, FR-MCP-003, FR-SEC-100, FR-SEC-101; ADR-021, ADR-077
- Related risks: RISK-MCP-001

#### Description

The MCP Runtime MUST enforce: (a) no MCP-origin surface is agent-visible before a recorded
positive Approval for the server's exposure; (b) descriptor pinning with drift suspension
per the trust model above; (c) stdio servers execute under the MCP-server sandbox tier
(FR-SEC-101) with deny-by-default environment, filesystem policy from the workspace
sandbox profile, and resource limits; (d) remote servers are reachable only under the
`network` permission and Volume 9 network policy (host/domain scopes); (e) MCP tool
invocations pass the full Tool Runtime pipeline — an MCP tool obtains no permission its
declared permission set and the user's decisions do not grant; (f) every exposure
decision, drift suspension, and denial is written to the Audit Log.

#### Motivation

MCP is the largest third-party attack surface in the product (tool poisoning, injection,
exfiltration via plausible-looking tools); gating and pinning convert those attacks from
silent to visible-and-blocked.

#### Actors

Users making exposure decisions; MCP Runtime; Permission Manager; hostile or compromised
servers (threat actors).

#### Preconditions

Server registered; sandbox tier profiles available; policy documents loaded.

#### Main flow

1. First connection completes discovery; the user is shown the inventory and asked for
   exposure approval (interactive) or policy decides (headless per PRD-009).
2. On `granted`, descriptors are pinned and tools registered agent-visible.
3. Subsequent discoveries diff against pins; unchanged surfaces stay exposed.

#### Alternative flows

- Drift detected: affected tools suspend (disabled in the registry), E-MCP-008 recorded,
  user notified; re-approval re-pins.
- Denial: surfaces remain hidden; the denial is a recorded decision (`denied` Approval).

#### Edge cases

- A server renaming a tool is treated as removal plus addition: the removed name suspends,
  the new name requires approval (additions after the initial approval are not
  auto-exposed).
- Registry-scope interplay: a workspace-scoped registration never gains exposure from a
  global approval of a same-named server (approvals bind to the registration row ULID).
- Non-interactive runs encountering an unapproved server: treat as denial (PRD-009), fail
  the dependent invocation with E-MCP-004 semantics plus the audit record.

#### Inputs

Discovery inventories, Approval decisions, policy documents, sandbox profiles.

#### Outputs

Exposure state per tool, pinned digests (persisted with the MCP Server row), audit
records, notifications.

#### States

Exposure state is derived (Tool rows enabled/disabled); connection states unchanged
(chapter 10).

#### Errors

E-MCP-008 (drift); E-SEC family for permission/sandbox refusals.

#### Constraints

Pinning uses SHA-256 over the canonical JSON encoding of the descriptor; approvals expire
per Volume 9 approval semantics only by revocation, not time, unless policy says
otherwise.

#### Security

This requirement is itself a security control; its bypass is a defect of the highest
severity (Tool Runtime mediation is the SM-16(b) enforcement point).

#### Observability

Audit records for every decision; `mcp.exposure.changed` event on grant/suspend/revoke
with tool counts.

#### Performance

Digest computation is O(descriptor size) at discovery time only; no per-invocation
overhead beyond the standard pipeline.

#### Compatibility

Pinning is revision-independent (digests over Andromeda's canonical descriptor form, not
raw protocol frames).

#### Acceptance criteria

- Given a newly registered server, when an agent requests one of its tools before
  approval, then the invocation is denied, an Audit Record exists, and the tool is absent
  from the agent-visible registry.
- Given an approved server whose tool description changes upstream, when re-discovery
  runs, then that tool is suspended, E-MCP-008 is recorded, and other tools remain
  exposed.
- Given a stdio server, when it attempts to read an environment variable outside the
  allowlist, then the variable is absent (sandbox assertion test).
- Permission case: given a remote server and a session without the `network` grant for its
  domain, when connecting, then the connection fails with the E-SEC denial and no TCP
  connection is made.
- Observability case: grant, suspension, and revocation each emit `mcp.exposure.changed`
  exactly once with correlation IDs.

#### Verification method

Enforcement tests attempting pre-approval invocation and post-drift invocation; sandbox
environment assertion tests; audit-chain tests (SM-16 pattern); conformance suite security
sections.

#### Traceability

PRD-005, PRD-006; FR-SEC-100, FR-SEC-101; ADR-021; RISK-MCP-001; chapter 05.

## Supply chain

Rules for how MCP servers arrive on the machine:

1. **Packaged servers** flow through the Package Manager: checksum verification is
   unconditional, signature policy per [chapter 09](09-package-manager-supply-chain.md).
   A packaged server's executable is recorded in the package `files_manifest`; integrity
   re-verification (`PackagePort.Verify`) detects post-install tampering.
2. **Registration-only servers** (user-supplied binaries or remote endpoints) are outside
   Andromeda's integrity envelope; their trust classification reflects that, and the
   registration records the executable path or URL for audit. Andromeda MUST NOT download
   server binaries implicitly as a side effect of registration.
3. **No implicit updates.** Andromeda never self-updates an MCP server; packaged updates
   are explicit package operations, and post-update reconnection re-runs negotiation and
   drift checks.
4. **Provenance in diagnostics.** Every MCP-origin Tool row carries `origin_ref` to its
   server; every server row carries its Extension record; the chain
   tool → server → package → source locator is navigable offline (SM-13 pattern).

### RISK-MCP-001 — Malicious or compromised MCP server

- Category: Security / supply chain
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: FR-MCP-006 gating, descriptor pinning, sandbox tiers, least exposure, untrusted-content labeling; chapter 09 verification for packaged servers; Volume 9 threat responses
- Detection: drift suspension (E-MCP-008), audit-chain review, health/probe anomalies, conformance-suite behavioral checks
- Owner: MCP Runtime (Volume 6) with Volume 9 threat model
- Status: Open

A server the user trusted can turn hostile (compromise, sale, maintainer change) and
attempt tool poisoning, prompt injection through descriptions or results, or data
exfiltration through plausible tool arguments. The mitigations bound the blast radius
(sandbox, permissions, exposure gating) and make behavior changes visible (pinning);
residual risk — a consistently hostile server approved by the user — is addressed by
Volume 9's threat entries and user-facing provenance labeling.

### RISK-MCP-002 — MCP specification churn and revision skew

- Category: Compatibility / external dependency
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: official SDK delegation (ADR-010), deferred revision pin (PENDING VALIDATION in the register), conformance suite per declared revision, fallback health probing
- Detection: conformance-suite failures on SDK updates; E-MCP-002 rates in interop runs
- Owner: MCP Runtime (Volume 6)
- Status: Open

The protocol revises on dated revisions and the ecosystem lags unevenly. Delegating
negotiation to the official SDK and testing per revision keeps skew a detected condition
rather than silent misbehavior.

## Conformance testing

### FR-MCP-007 — MCP conformance test program

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: MCP Runtime (Volume 6)
- Affected components: MCP Runtime, testing infrastructure (Volume 13 stack)
- Dependencies: FR-MCP-001; ADR-010, ADR-017
- Related risks: RISK-MCP-002

#### Description

Andromeda MUST maintain an MCP conformance test program with three layers: (1)
**recorded-session conformance** — fixture exchanges per declared protocol revision
(initialize, listings, pagination, tool call, resource read, prompt get, notifications,
error frames), replayed against the MCP Runtime without the SDK's network stack, verifying
Andromeda-side behavior independent of SDK internals (ADR-010 reversal-plan support); (2)
**behavioral fixtures** — scripted fixture servers exercising failure modes (hangs at each
phase, malformed frames, duplicate names, drift, log floods, oversized results); (3)
**live interop** — a scheduled job connecting to a maintained reference set of at least 10
public MCP servers, verifying connect, discover, and one tool invocation each (SM-15b).
The suite MUST run in CI per release for layers 1–2 and on schedule for layer 3; layer 3
results never gate merges (external availability) but gate releases per NFR-MCP-002.

#### Motivation

SM-15 makes MCP interoperability a measured product commitment; without a conformance
program, protocol correctness would be an assumption.

#### Actors

CI; test authors; the MCP Runtime under test.

#### Preconditions

Declared revision set recorded (until the pin resolves, the SDK-supported set stands in);
reference-server list maintained in-repo.

#### Main flow

1. CI executes recorded-session and behavioral layers per revision.
2. The scheduled interop job runs against the reference set and publishes a scorecard.
3. Failures open defects tagged to the affected revision or server.

#### Alternative flows

- A reference server goes offline permanently: it is replaced via PR to the list; the
  scorecard notes the substitution.

#### Edge cases

- Revision-specific behavior (e.g., ping availability) is asserted only for revisions
  that document it; the suite parameterizes by revision rather than duplicating fixtures.
- Fixtures MUST NOT require network access (offline CI viability); only layer 3 touches
  the network.

#### Inputs

Fixture recordings, fixture server scripts, reference-server list, declared revision set.

#### Outputs

Pass/fail per check per revision; interop scorecard; defect reports.

#### States

Not applicable — test program, no runtime entity states.

#### Errors

Test failures reference the E-MCP codes they expected versus observed; the suite itself
mints no runtime errors.

#### Constraints

Layers 1–2 run offline; fixtures live in the monorepo (ADR-003); testing stack per
ADR-017.

#### Security

Behavioral fixtures include hostile-server cases (drift, injection-shaped descriptions)
asserting the FR-MCP-006 controls fire.

#### Observability

Suite results exported as CI artifacts; interop scorecard retained per release for trend
review (metric governance, Volume 1).

#### Performance

Layers 1–2 complete within the CI integration-test budget (Volume 13 gates); no
per-release manual steps.

#### Compatibility

The suite parameterizes by declared protocol revision: adding or dropping a revision from
the declared set adds or retires that revision's applicable checks without touching the
others; fixture recordings are per-revision artifacts kept while their revision stays
declared.

#### Acceptance criteria

- Given the declared revision set, when the conformance suite runs, then every applicable
  check executes and reports, with 100% pass required at the NFR-MCP-001 gate.
- Given the reference-server list, when the interop job runs, then each server yields a
  connect/discover/invoke verdict and the scorecard aggregates ≥ 95% success at the
  NFR-MCP-002 gate.
- Negative case: given a fixture server that swaps a tool description post-approval, when
  the behavioral layer runs, then the suite passes only if exposure was suspended with
  E-MCP-008.
- Observability case: suite runs record which revision each check targeted, so a failing
  SDK update is attributable to a revision.

#### Verification method

The suite is itself the method; its presence, coverage mapping to this chapter's
requirements, and gating wiring are audited at phase gates (Volume 13 release
qualification).

#### Traceability

SM-15; NFR-MCP-001, NFR-MCP-002; ADR-010, ADR-017; Volume 13 test-type catalog
(MCP-conformance row).

### NFR-MCP-001 — MCP conformance pass rate

- Category: Compatibility
- Priority: P1
- Phase: v1
- Metric: Percentage of applicable conformance checks (FR-MCP-007 layers 1–2) passing per declared protocol revision
- Target: 100% of applicable checks per declared revision (SM-15a)
- Minimum threshold: 100% — a failing applicable check is a release blocker at v1; measured without gating from Beta
- Measurement method: conformance suite in CI per release; per-revision reporting; "applicable" is determined by the revision's documented feature set
- Test environment: CI runners per Volume 13; offline for layers 1–2
- Measurement frequency: every release; every SDK version bump
- Owner: MCP Runtime (Volume 6)
- Dependencies: FR-MCP-007; ADR-010
- Risks: RISK-MCP-002
- Acceptance criteria: Release qualification report shows 100% applicable-check pass for every declared revision; any waiver is a recorded change-procedure decision.

### NFR-MCP-002 — MCP reference-server interoperation

- Category: Compatibility
- Priority: P1
- Phase: v1
- Metric: Fraction of the maintained reference set (≥ 10 public MCP servers) for which connect, discover, and one tool invocation succeed
- Target: ≥ 95% (SM-15b)
- Minimum threshold: ≥ 95% at v1 release qualification; measured without gating from Beta
- Measurement method: scheduled interop job (FR-MCP-007 layer 3) with published scorecard; transient outages retried once within the job before counting as failure
- Test environment: CI with network access, reference network conditions per Volume 1 chapter 06
- Measurement frequency: weekly scheduled runs; mandatory run at each release qualification
- Owner: MCP Runtime (Volume 6)
- Dependencies: FR-MCP-007
- Risks: RISK-MCP-001, RISK-MCP-002
- Acceptance criteria: The release-qualification scorecard shows ≥ 95% success over the current reference list, with each failure triaged (server-side versus Andromeda-side) in the report.

## Error codes (chapter 06)

### E-MCP-008 — MCP tool descriptor drift detected

- Category: Security
- Severity: Warning (exposure suspended; nothing invoked)
- User message: "MCP server '<name>' changed tool '<tool>' since you approved it; the tool is suspended until you re-approve."
- Technical message: pinned digest, observed digest, changed descriptor fields
- Cause: upstream descriptor change after approval (benign update or hostile swap)
- Safe-to-log data: server name, tool name, digests, changed field names
- Recoverability: recoverable by explicit re-approval (re-pins) or by disabling the server
- Retry policy: none — deterministic until re-approval
- Recommended action: review the diff shown by the drift notification, then re-approve or disable
- Exit-code mapping: 5 when it blocks a requested invocation; otherwise not surfaced as process exit
- HTTP mapping: not applicable
- Telemetry event: `mcp.exposure.changed`
- Security implications: primary anti-rug-pull control; suspension is fail-closed and audit-logged

## Events minted (chapter 06)

| Event | Version | Producer | Consumers | Payload summary |
|---|---|---|---|---|
| `mcp.exposure.changed` | 1 | MCP Runtime | TUI, Observability, Audit Log | server name, change kind (granted/suspended/revoked), affected tool count |
