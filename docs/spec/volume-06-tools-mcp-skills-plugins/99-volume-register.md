# volume-06-tools-mcp-skills-plugins — Volume Register

Merged from per-agent register fragments at the Phase B gate.

## Requirements index

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-TOOL-001 | Tool contract | Core | Tool contract conformance suite over all built-ins and SDK fixtures; registration fuzzing; mediation probes |
| FR-TOOL-002 | Declaration and payload schema validation | Core | Official JSON Schema vectors per draft; mutation and nonconformance fixtures; network-isolation assertion |
| FR-TOOL-003 | Tool naming, namespaces, and resolution | Core | Grammar fuzzing; collision matrices; alias fixtures; canonical-identity audit checks |
| FR-TOOL-004 | Tool registration, availability, and enablement | MVP | Provider-lifecycle integration tests; per-scope persistence tests; audit-chain resolution |
| FR-TOOL-005 | Permission mediation for every invocation | MVP | Per-mode permission matrix tests; unmediated-execution probes; expiry/revocation race tests |
| FR-TOOL-006 | Execution limits, sandbox placement, and teardown | MVP | Fault-injection suite (hang, fork-bomb, flood); teardown timing per platform; containment-level record assertions |
| FR-TOOL-007 | Built-in tool catalog and phasing | MVP | Per-tool conformance and golden fixtures; offline suite; Tier 1 matrix; recorded-API contract tests |
| FR-TOOL-008 | Tool Invocation machine conformance | MVP | Transition-matrix property tests; crash injection at every boundary; race tests; audit-chain resolution |
| FR-SDK-001 | Extension SDK | Beta | CI tutorial walkthrough (SM-02 method); mirror-equivalence check; fixture self-tests; timed gate exercises |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-SDK-001 | Tool creation time (SM-02) | Beta | Timed phase-gate exercise; CI tutorial walkthrough; contribution-record sampling |
| NFR-TOOL-001 | Built-in tool contract conformance | MVP | Conformance suite in CI per release across Tier 1 matrix; release audit attachment |
| NFR-TOOL-002 | Invocation record and event completeness | MVP | Audit-chain test over integration/E2E and crash-injection runs per release |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-TOOL-001 | Dishonest or over-broad tool declarations | Critical | Open |
| RISK-TOOL-002 | Name shadowing and typosquatting across origins | High | Open |
| RISK-TOOL-003 | External service API drift breaking integration tools | High | Open — PENDING VALIDATION per service |
| RISK-TOOL-004 | Output flooding and resource exhaustion through tools | High | Open |

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-MCP-001 | MCP client support (keystone) | Beta | MCP conformance suite and interop job (SM-15); sandbox-launch, credential, and reconnection integration tests; secret-scan over sinks |
| FR-MCP-002 | MCP transports and connection establishment | Beta | Fixture-server integration for both transports; fault injection per phase; sandbox-policy assertions |
| FR-MCP-003 | Discovery and bridging of tools, resources, and prompts | Beta | Conformance fixtures (listings, pagination, listChanged); schema-validation units; bridge-teardown integration; interop suite (SM-15b) |
| FR-MCP-004 | MCP server authorization | Beta | Authorized fixture servers; secret-scan assertions; permission-denial tests; OAuth PENDING VALIDATION register cross-check |
| FR-MCP-005 | Connection health, server logs, and maintenance operations | Beta | Fault injection (probe failures, log floods); update/uninstall integration over fixture packages; redaction assertions |
| FR-MCP-006 | MCP trust gating and isolation | Beta | Pre-approval and post-drift enforcement probes; sandbox environment assertions; audit-chain tests (SM-16 pattern) |
| FR-MCP-007 | MCP conformance test program | Beta | Suite presence, coverage mapping, and gating wiring audited at phase gates (Volume 13 release qualification) |
| FR-MCP-008 | MCP Client Connection machine conformance | Beta | Transition-matrix property tests; per-phase fault injection; crash-injection recovery; routing-refusal probes in every non-ready state |
| FR-SKILL-001 | Skill format and manifest (keystone) | Beta | Golden manifest corpus (valid + invalid); hash tamper tests; schema round-trip against the SDK mirror |
| FR-SKILL-002 | Skill loading, requirement resolution, and activation | Beta | Registry-gap integration tests; run-record inspection; resolver-determinism property tests |
| FR-SKILL-003 | Skill inheritance, composition, and overrides | Beta | Composition conflict matrix; determinism and cycle-detection property tests; golden composition reports |
| FR-SKILL-004 | Skill testing and fixtures | Beta | Runner self-tests; SDK template test in CI; offline-suite inclusion |
| FR-SKILL-005 | Skill distribution and deprecation | Beta | End-to-end package tests; signature policy matrix; deprecation selection tests; audit-chain inspection |
| FR-PLUG-001 | Plugin runtime over the Andromeda Runtime Protocol (keystone) | Beta | ARP conformance fixtures (handshake, streaming, cancellation, reserved codes); chaos tests; cross-language smoke plugin; SM-03 timed exercise |
| FR-PLUG-002 | ARP handshake, version negotiation, and method conformance | Beta | Version-parameterized conformance fixtures; SDK cross-tests; negotiation property tests |
| FR-PLUG-003 | Plugin permission mediation and sandbox containment | Beta | Sandbox assertion fixtures (env, filesystem, network, process tree); audit-chain tests; manifest-bound consistency tests; secret-scan on debug captures |
| FR-PLUG-004 | Plugin supervision, health, and restart | Beta | Chaos tests (kill, hang, crash-loop); restart-policy property tests; persisted-state recovery tests; leak checks |
| FR-PLUG-005 | Extension package operations | Beta | End-to-end operation tests over all four source kinds; crash injection per state; upgrade/rollback tests; cascade-removal tests; audit-chain inspection |
| FR-PLUG-006 | Package sources, discovery, and dependency resolution | Beta | Resolver property tests (determinism, bounds, intersection); multi-source integration incl. divergent checksums; offline-cache tests; permission-denial tests |
| FR-PLUG-007 | Package verification and integrity | Beta | Verification matrix (checksum × signature × policy × source flags); tamper fixtures; offline verification tests; audit-chain assertions |
| FR-PLUG-008 | Plugin lifecycle machine conformance | Beta | Transition-matrix property tests; chaos suite; crash-injection reconciliation tests; event-sequence reconstruction |
| FR-PLUG-009 | Package installation machine conformance | Beta | Transition-matrix property tests; crash injection at every boundary (fresh + upgrade); bundle-atomicity tests; staging-cleanup assertions |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-MCP-001 | MCP conformance pass rate (SM-15a) | v1 | Conformance suite in CI per release and per SDK bump; per-revision reporting |
| NFR-MCP-002 | MCP reference-server interoperation (SM-15b) | v1 | Weekly scheduled interop job; mandatory run with scorecard at release qualification |
| NFR-SKILL-001 | Skill validation and composition latency | Beta | Benchmark harness over the fixture skill corpus, p95, per release |
| NFR-PLUG-001 | Plugin creation time (SM-03) | v1 | Timed phase-gate exercise against the SDK plugin template; CI tutorial walkthrough |
| NFR-PLUG-002 | ARP invocation dispatch overhead | Beta | Instrumented no-op fixture plugin benchmark, 1000 invocations, p95, per release |
| NFR-PLUG-003 | Extension package operation latency | Beta | Benchmark harness over fixture packages and indexes, p95, per release |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-MCP-001 | Malicious or compromised MCP server | High | Open |
| RISK-MCP-002 | MCP specification churn and revision skew | Medium | Open |
| RISK-SKILL-001 | Malicious or manipulative skill content | High | Open |
| RISK-PLUG-001 | Malicious or compromised plugin | High | Open |
| RISK-PLUG-002 | Protocol evolution breaking the plugin ecosystem | Medium | Open |
| RISK-PLUG-003 | Extension supply-chain compromise | High | Open |

## ADRs minted

Block 070–084 belongs to Volume 6; fragment A used 070–074 (075–076 remain unused by A;
fragment B mints from 077).

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-070](../annexes/adr/ADR-070.md) | Tool identity: dotted names, reserved namespaces, and collision rejection | Accepted | `namespace.action` grammar; built-in namespaces reserved; cross-origin collisions rejected, user aliasing the only coexistence path |
| [ADR-071](../annexes/adr/ADR-071.md) | Tool output handling: bounded inline results with artifact spillover | Accepted | Inline output capped (default 1 MiB); overflow spills to content-addressed Artifacts; truncation always marked; binary always by reference |
| [ADR-072](../annexes/adr/ADR-072.md) | Automatic tool retries bound to declared idempotency | Accepted | Runtime auto-retries only idempotent tools on retryable errors under attempt caps; each retry a new chained invocation row |
| [ADR-073](../annexes/adr/ADR-073.md) | Third-party tool delivery: plugins and MCP servers only | Accepted | No in-process third-party tool loading; SDK tool kit authors built-ins and plugin tool surfaces; WASM stays a v2 candidate per ADR-009 |
| [ADR-074](../annexes/adr/ADR-074.md) | Built-in integration tools: official public APIs only, phased beyond MVP | Accepted | Integrations use official documented APIs with Secret Store credentials; Beta → v1 → v2 phasing; per-service facts PENDING VALIDATION |

Fragment B used 077–083 of the Volume 6 block (070–084); 084 remains an unused permanent
gap per Volume 0 chapter 03.

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-077](../annexes/adr/ADR-077.md) | MCP surfaces bridge into existing runtimes under the reserved `mcp:` scheme | Accepted | MCP tools/resources/prompts bridge into Tool Runtime, Context Manager, and prompt selection as `mcp:<server>/<name>`; no second execution path; `mcp:` reserved |
| [ADR-078](../annexes/adr/ADR-078.md) | Skill package format: content-addressed directory with a TOML manifest, data only | Accepted | Directory + `skill.toml` (schema-validated), separate prompt files, SHA-256 content addressing, executable content rejected fail-closed |
| [ADR-079](../annexes/adr/ADR-079.md) | Andromeda Runtime Protocol 1.0: line-delimited JSON-RPC 2.0 with one-way authority | Accepted | NDJSON framing on stdio, 4 MiB frame cap with out-of-band artifacts, MAJOR.MINOR negotiation, no plugin→host requests, five reserved error codes |
| [ADR-080](../annexes/adr/ADR-080.md) | Extension distribution: declared package sources now, curated marketplace Future | Accepted | Four source kinds (`registry`/`git`/`archive`/`path`) with a JSON registry-index format; official index PENDING VALIDATION; marketplace = a future registry-kind source |
| [ADR-081](../annexes/adr/ADR-081.md) | Extension signature policy: checksums unconditional, cosign signatures policy-gated, invalid always blocks | Accepted | cosign (v3 pinned) over archive digests; `verified`/`unverified`/`invalid` mapping; `invalid` blocks unconditionally; per-source `signature_required` downgrade defense |
| [ADR-082](../annexes/adr/ADR-082.md) | Skill inheritance and composition: single parent, append-by-default, deterministic precedence, fail-closed conflicts | Accepted | Depth-5 single-parent chains, declared overrides only, scope/order/name precedence, conflicts never auto-resolved, composition reports with set hashes |
| [ADR-083](../annexes/adr/ADR-083.md) | Extension dependency resolution: constraint intersection with highest-satisfying version, no backtracking search | Accepted | Bounded closure (depth 5, 64 nodes), range intersection, highest satisfier, first-source-wins by priority, recorded resolution plans as the only Install input |

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-TOOL-001 | Tool not found | 6 (CLI usage surface: 2) |
| E-TOOL-002 | Tool registration rejected | 6 |
| E-TOOL-003 | Input validation failed | 6 |
| E-TOOL-004 | Output validation failed | 6 |
| E-TOOL-005 | Invocation denied | 5 |
| E-TOOL-006 | Tool execution failed | 6 |
| E-TOOL-007 | Invocation timed out | 8 |
| E-TOOL-008 | Invocation cancelled | 8 |
| E-TOOL-009 | Resource limit exceeded | 6 |
| E-TOOL-010 | Tool origin unavailable | 6 |
| E-TOOL-011 | Concurrency capacity exhausted | 6 |
| E-TOOL-012 | Execution interrupted | 6 |
| E-MCP-001 | MCP connection establishment failed | 6 (invocation path) / 1 (standalone) |
| E-MCP-002 | MCP protocol negotiation failed | 6 (invocation path) / 1 (standalone) |
| E-MCP-003 | MCP request timed out | 8 |
| E-MCP-004 | MCP connection not ready | 6 |
| E-MCP-005 | MCP authorization failed | 4 |
| E-MCP-006 | MCP protocol violation by server | 6 |
| E-MCP-007 | MCP server registration invalid | 3 |
| E-MCP-008 | MCP tool descriptor drift detected | 5 (when blocking an invocation) |
| E-SKILL-001 | Skill manifest invalid | 3 |
| E-SKILL-002 | Skill content integrity mismatch | 9 |
| E-SKILL-003 | Skill requirements unsatisfiable | 1 |
| E-SKILL-004 | Skill resolution failed | 1 |
| E-SKILL-005 | Skill composition conflict | 1 |
| E-SKILL-006 | Executable content rejected in skill | 3 |
| E-PLUG-001 | Plugin spawn failed | 6 (invocation path) / 1 (lifecycle commands) |
| E-PLUG-002 | ARP version negotiation failed | 1 |
| E-PLUG-003 | ARP protocol violation | 6 |
| E-PLUG-004 | Plugin request timed out | 8 |
| E-PLUG-005 | Plugin process crashed | 6 |
| E-PLUG-006 | Plugin manifest invalid | 3 |
| E-PLUG-007 | Undeclared surface registration attempted | 6 |
| E-PLUG-008 | Package source unavailable | 1 |
| E-PLUG-009 | Package resolution failed | 1 |
| E-PLUG-010 | Package download failed | 1 |
| E-PLUG-011 | Package verification failed | 9 |
| E-PLUG-012 | Package installation conflict | 1 |
| E-PLUG-013 | Package operation interrupted or cancelled | 8 |

## Events minted

Envelope semantics per Volume 10; grammar per Volume 0 chapter 03.

| Event | Emitted by |
|---|---|
| `tool.registration.completed` | Tool Runtime |
| `tool.registration.rejected` | Tool Runtime |
| `tool.enablement.changed` | Tool Runtime |
| `tool.invocation.requested` | Tool Runtime |
| `tool.invocation.approved` | Tool Runtime |
| `tool.invocation.denied` | Tool Runtime |
| `tool.invocation.started` | Tool Runtime |
| `tool.invocation.succeeded` | Tool Runtime |
| `tool.invocation.failed` | Tool Runtime |
| `tool.invocation.timed_out` | Tool Runtime |
| `tool.invocation.cancelled` | Tool Runtime |
| `tool.invocation.retried` | Tool Runtime |
| `tool.output.truncated` | Tool Runtime |
| `terminal.execution.started` | Terminal Engine |
| `terminal.execution.ended` | Terminal Engine |
| `terminal.output.truncated` | Terminal Engine |

Envelope semantics per Volume 10 (FR-OBS-001); grammar per Volume 0 chapter 03.

| Event | Emitted by |
|---|---|
| `mcp.server.registered` | MCP Runtime |
| `mcp.server.updated` | MCP Runtime |
| `mcp.server.removed` | MCP Runtime |
| `mcp.connection.established` | MCP Runtime |
| `mcp.connection.lost` | MCP Runtime |
| `mcp.connection.failed` | MCP Runtime |
| `mcp.surfaces.discovered` | MCP Runtime |
| `mcp.request.failed` | MCP Runtime |
| `mcp.log.received` | MCP Runtime |
| `mcp.exposure.changed` | MCP Runtime |
| `skill.registered` | Skill Engine |
| `skill.validation.failed` | Skill Engine |
| `skill.activated` | Skill Engine |
| `skill.deactivated` | Skill Engine |
| `skill.composition.resolved` | Skill Engine |
| `skill.deprecated` | Skill Engine |
| `plugin.registered` | Plugin Runtime |
| `plugin.started` | Plugin Runtime |
| `plugin.handshake.completed` | Plugin Runtime |
| `plugin.handshake.failed` | Plugin Runtime |
| `plugin.stopped` | Plugin Runtime |
| `plugin.crashed` | Plugin Runtime |
| `plugin.restarted` | Plugin Runtime |
| `plugin.disabled` | Plugin Runtime |
| `plugin.removed` | Plugin Runtime |
| `package.resolution.completed` | Package Manager |
| `package.installation.started` | Package Manager |
| `package.installation.completed` | Package Manager |
| `package.installation.failed` | Package Manager |
| `package.verification.failed` | Package Manager |
| `package.rollback.completed` | Package Manager |
| `package.removal.completed` | Package Manager |

## Config keys minted

`[tools]` table content (schema/precedence Volume 10's): `tools.default_timeout_ms` (60000),
`tools.max_timeout_ms` (600000), `tools.max_output_bytes` (1048576),
`tools.max_concurrent_invocations` (8), `tools.teardown_grace_ms` (2000),
`tools.teardown_kill_ms` (3000), `tools.max_auto_retries` (2), `tools.disabled` ([]),
`tools.allowed_origins` (["builtin", "plugin", "mcp"]), `tools.aliases` (empty table).

Schema, precedence, and validation are Volume 10's; key content per Volume 0 chapter 03
table ownership.

- `[mcp]` (chapter 05): `mcp.connect_timeout_ms` (10000), `mcp.initialize_timeout_ms`
  (10000), `mcp.request_timeout_ms` (60000), `mcp.reconnect_max_attempts` (5),
  `mcp.reconnect_backoff_initial_ms` (1000), `mcp.log_capture` (true); per-server subtables
  `[mcp.servers.<name>]` with `transport`, `command`, `args`, `env_allowlist`, `url`,
  `headers`, `credential`, `enabled`, `scope_hint`, `expose_tools` (chapter 06 least-exposure
  allowlist), plus per-server overrides of every runtime-wide default.
- `[skills]` (chapter 07): `skills.enabled` (true), `skills.paths` ([]), `skills.autoload`
  (true), `skills.sources` ([]), `skills.activation_policy` ("prompt").
- `[plugins]` (chapter 08): `plugins.enabled` (true), `plugins.handshake_timeout_ms`
  (10000), `plugins.request_timeout_ms` (60000), `plugins.stop_timeout_ms` (5000),
  `plugins.health_interval_ms` (30000), `plugins.restart_max_attempts` (5),
  `plugins.restart_backoff_initial_ms` (500), `plugins.sources` ([]), and per-plugin
  `[plugins.overrides.<name>]` tables.
- Source entry schema (chapter 09; shared by `plugins.sources` and `skills.sources`
  entries): `name`, `kind` (`registry` | `git` | `archive` | `path`), `location`,
  `priority` (100), `enabled` (true), `signature_required` (false), `timeout_ms` (300000).

## Glossary additions

| Term | One-line meaning |
|---|---|
| Tool Declaration | The single self-describing JSON contract document every tool registers (chapter 01). |
| ToolEvent | The five-kind ordered stream union a tool emits during execution: `progress`, `log`, `output_delta`, `artifact`, `result`. |
| Invocation pipeline | The fixed validate → permission → sandbox → execute → record order every invocation traverses. |
| Reserved namespace | A built-in tool namespace closed to third-party registration (ADR-070). |
| Spillover Artifact | The content-addressed Artifact holding tool output beyond the inline cap (ADR-071). |
| Effective timeout | min(declared timeout or configured default, configured cap), recorded per invocation. |
| Teardown budget | The grace-plus-kill window bounding termination escalation after cancel or timeout. |
| ARP handshake | The host-initiated `arp.initialize` exchange negotiating the protocol version and validating declared surfaces against the manifest (chapter 08). |
| Descriptor pinning | SHA-256 digests of approved MCP tool descriptors persisted at exposure approval; later drift suspends exposure (E-MCP-008). |
| Exposure approval | The recorded Approval that makes an MCP server's discovered surfaces agent-visible (chapter 06). |
| Package source | A configured acquisition endpoint of kind `registry`, `git`, `archive`, or `path`, consulted by discovery and resolution (chapter 09). |
| Registry index | The schema-versioned JSON document a `registry` source serves: packages with versions, kinds, locators, checksums, and signature references. |
| Resolution plan | The deterministic recorded output of `PackagePort.Resolve` — exact versions, sources, checksums, signature references — and the only input `Install` accepts. |
| Staging area | The scope-local directory where acquired archives are verified and extracted before installation; cleaned on every failure path. |
| Composition report | The structured record of every skill contribution, its source skill/version, and every override applied in one activation set (chapter 07). |
| Skill activation | Making a registered skill effective for a session, profile, or run after its requirements resolve (FR-SKILL-002). |

## Assumptions

Local list per Volume 0 chapter 05 (global numbers at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | A 1 MiB inline output cap serves the common tool-output distribution without routine spillover | Volume 12 benchmarks; Volume 7 context budgeting review | Adjust `tools.max_output_bytes` default — configuration, not contract, change (ADR-071 review condition) |
| Technical assumption | Built-in tools can honor cooperative cancellation within the 2 s grace budget | Teardown timing tests per platform (FR-TOOL-006 verification) | Tune `[tools]` teardown keys; kill escalation already bounds the worst case |
| Technical assumption | Declared idempotency intersected with envelope retryability covers the useful automatic-retry space | SM-10 field data and suite telemetry | ADR-072 review conditions trigger; policy-engine-driven retry rules considered |
| Product hypothesis | Typed `*.request` operation surfaces (rather than raw REST pass-through) are what agents need from integration tools | Beta/v1 usage telemetry and issue reports | Add policy-gated raw operations per service through the change procedure |

Local list per Volume 0 chapter 05 (global numbers at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | Line-delimited JSON-RPC framing with the 4 MiB cap and out-of-band artifacts serves plugin workloads within NFR-PLUG-002 budgets | Volume 12 benchmarks; ADR-009/ADR-079 review conditions | Transport migration behind the ARP contract (sockets or alternative framing); method surface unchanged |
| Technical assumption | The official SDK's supported MCP revision set covers the practical server ecosystem until the pin resolves | SM-15 interop scorecard trends per release | Accelerate SDK version adoption; escalate via the ADR-010 review conditions |
| Product hypothesis | Data-only skills express the majority of shareable procedural knowledge without executable hooks | Beta/v1 usage telemetry and issue signals | Executable behavior remains plugin territory; the skill format holds (INV-SKL-04) and authoring guidance improves |
| Technical assumption | Independent publishers adopt cosign signing gradually; unsigned-but-checksummed packages dominate early Beta | Signature-adoption statistics at phase gates (ADR-081 review condition) | Default policy for `unverified` content tightens later than planned; the mechanism and states are unchanged |

## Open questions

Every PENDING VALIDATION in chapters 00–04 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V6A-OQ-1 | `browser.control` driving mechanism: selection and packaging of an official automation standard (W3C WebDriver or equivalent documented protocol) — PENDING VALIDATION | Chapter 03; ADR-074 | No — tool is v1 phase | Implementation spike at v1 planning; outcome recorded via ADR amendment | Open |
| V6A-OQ-2 | Per-service API facts for integration tools (GitHub, GitLab, Jira, Notion, Slack, Linear, Docker Engine, Kubernetes): versions, auth modes, scopes, rate limits — PENDING VALIDATION per service | Chapter 03; ADR-074; RISK-TOOL-003 | No — phases gate each tool | Verify against official documentation at each tool's phase start; record per-service notes in this volume | Open |
| V6A-OQ-3 | WASM in-process plugin isolation as a possible future third tool channel — PENDING VALIDATION per ADR-009 (referenced by ADR-073) | ADR-073 | No — v2 candidate only | ADR-009 review-condition spike before any v2 commitment | Open |

Every PENDING VALIDATION in chapters 05–10 maps to a row here.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V6B-OQ-1 | Pinned/certified MCP protocol revision set — PENDING VALIDATION per ADR-010 rule 3 (depends on the 2026-07-28 specification RC cycle) | Chapters 05, 06; FR-MCP-001 | No — the SDK-supported set stands in until the pin | Fix the pin when the RC lands and the SDK declares support; resolve via ADR-010 review condition | Open |
| V6B-OQ-2 | OAuth-based MCP server authorization — PENDING VALIDATION per ADR-010 rule 4 (SDK client OAuth is experimental) | Chapter 05 (FR-MCP-004) | No — token/header authorization is the Beta path | Validate when the SDK graduates client OAuth; a superseding decision activates it with ADR-014 credential handling | Open |
| V6B-OQ-3 | OS-level sandbox mechanisms for the plugin and MCP-server tiers — PENDING VALIDATION per ADR-021 (Seatbelt, Landlock, namespaces/bubblewrap) | Chapters 05, 08 (FR-PLUG-003) | No — process-level containment is the MVP/Beta floor | Volume 9 per-platform mechanism validation before any isolation claim | Open |
| V6B-OQ-4 | Official Andromeda extension registry index: hosting, operation, and URL — PENDING VALIDATION per ADR-080 | Chapter 09 (FR-PLUG-006); ADR-080, ADR-081 | No — `git`/`archive`/`path` sources and third-party indexes work without it | Organizational decision; amend ADR-080 with the default-source and publisher-identity policy | Open |

## Cross-volume references

Volume 2: tool-and-action entity shapes, invariants, frozen states, write discipline
(chapters 04/09/10). Volume 3: frozen ToolPort/TerminalPort signatures elaborated here;
component boundaries; the SandboxPort-only launch rule. Volume 9: permission enum, scopes,
decisions, trust vocabulary, sandbox model (keystones FR-SEC-100, FR-SEC-101 referenced by
FR-TOOL-005/006). Volume 10: event envelope; configuration schema/precedence for the
`[tools]` keys; artifact retention. Volume 5: Authentication Layer for integration-tool
credentials (FR-AUTH-001 referenced by ADR-074). Volume 11: Git semantics and destructive
gates behind `git.exec`; product-side hosting integrations. Volume 8: CLI/TUI surfaces.
Volume 12: dispatch/teardown/persistence budgets. Volume 13: conformance suites and crash
injection. Within Volume 6 (fragment B): MCP bridging (FR-MCP-001), plugin surfaces
(FR-PLUG-001), skill format (FR-SKILL-001), package trust inputs.

Volume 2: entity shapes and invariants elaborated here — MCP Server / MCP Client Connection
(INV-MCPS-01..04, INV-MCPC-01..04), Skill (INV-SKL-01..04), Plugin (INV-PLG-01..04), Package
(INV-PKG-01..04), Extension (INV-EXT-01..04); frozen state enums (chapter 09); write
discipline (chapter 10). Volume 3: PackagePort contract elaborated here; SandboxPort-only
launch rule for stdio servers and plugins; ADR-030/031 layering (SDK containment per
ADR-010). Volume 4: Skill Engine execution semantics (chapter 08 there) consume the format
defined here; Prompt Engine slots receive skill contributions. Volume 5: capability enum
consumed by skill requirements; credential flows (FR-AUTH-001 referenced) behind MCP
authorization. Volume 7: MCP resources enter context as Context Items under Volume 7
budgets. Volume 8: `mcp`/`skill`/`plugin` CLI command groups and TUI surfaces present the
operations specified here. Volume 9: permission enum/scopes/decisions, trust vocabulary and
thresholds, sandbox tiers (keystones FR-SEC-100/101/102 referenced), Approval semantics,
redaction. Volume 10: configuration schema/precedence for `[mcp]`, `[skills]`, `[plugins]`;
event envelope (FR-OBS-001); retention. Volume 13: conformance suites (MCP conformance,
ARP fixtures), crash injection, offline suites. Volume 14: product updates deliberately
disjoint from extension packages; shared signing tooling patterns (ADR-013 ↔ ADR-081).
Within Volume 6 (fragment A): tool contract (FR-TOOL-001), invocation pipeline and machine
(FR-TOOL-008 chapter 04), SDK keystone (FR-SDK-001), `[tools]` keys, ADR-070 naming
complemented by ADR-077.
