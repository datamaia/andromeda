# 99 — Volume 3 Register

Machine-parseable register of everything Volume 3 minted, per Volume 0 chapters 02 and 03.
Merged into the Volume 0 registers at consolidation.

## Requirements index

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-ARCH-001 | Layered dependency rule | Core | depguard + import-graph test in CI (ADR-033); release audit |
| FR-ARCH-002 | Ports-and-adapters composition | Core | Static adapter-name checks; contract suites over doubles; startup failure injection |
| FR-ARCH-003 | Port name and signature freeze | Core | Spec-lint and consolidation audit; contract-diff tooling in CI |
| FR-ARCH-004 | Context propagation and cancellation on all ports | Core | Per-method cancellation cases in contract suites; leak and fault-injection tests |
| FR-ARCH-005 | Bounded process model | Core | Process-tree termination tests per platform; spawn-path static audit |
| FR-ARCH-006 | Supervised concurrency | Core | Naked-goroutine scan; scheduler contract suite; leak gates |
| FR-ARCH-007 | External IPC control surface | MVP | IPC conformance suite; cross-user security tests; CLI-parity tests |
| FR-ARCH-008 | Headless operating mode | Beta | Headless E2E suite; interactive/headless record-parity comparison |
| FR-ARCH-009 | Crash recovery and resumable state | MVP | SM-11 crash-injection suite; recovery idempotence tests; corruption fixtures |
| FR-ARCH-010 | Graceful shutdown ordering | MVP | Shutdown-ordering integration tests; escalation and wedged-child fixtures |
| FR-ARCH-011 | Extension through versioned public contracts | Core | Per-surface conformance suites; SM-16(b) matrix over extension origins; compatibility fixtures |
| FR-PORT-001 | Platform encapsulation | Core | Prohibited-construct scanner in CI; PAL conformance suite |
| FR-PORT-002 | PAL surface completeness and portability of signatures | Core | Per-surface conformance suites per Tier 1 platform; signature review |
| FR-PORT-003 | Directory resolution through the PAL with XDG semantics | Core | Golden path-resolution tests per platform; container CI job |
| FR-PORT-004 | Platform support matrix conformance | MVP | Tier 1 release gates; SM-17 measurement; startup-refusal tests |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-ARCH-001 | Dependency-rule enforcement in CI | Core | Required PR checks (depguard + graph test); release audit |
| NFR-ARCH-002 | Port contract stability | Beta | Contract-diff per release; Volume 14 release audit |
| NFR-ARCH-003 | Shutdown deadline | MVP | Instrumented shutdown timing in suites and field diagnostics |
| NFR-ARCH-004 | Leak-free termination | Core | Goroutine-leak gates per suite; post-shutdown process scans |
| NFR-PORT-001 | Tier 1 behavioral parity | MVP | Full acceptance suite per Tier 1 platform per release (SM-17) |
| NFR-PORT-002 | PAL conformance coverage | Core | Surface-to-suite inventory check in CI |
| NFR-PORT-003 | Single-binary deliverable with bounded prerequisites | MVP | Linkage inspection; clean-machine install tests |
| NFR-PORT-004 | Platform-conditional code containment | Core | Automated scanner on every PR |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-ARCH-001 | Layering erosion under delivery pressure | High | Open |
| RISK-ARCH-002 | Port interface churn during parallel authoring | High | Open |
| RISK-ARCH-003 | Task Scheduler as a critical single point | High | Open |
| RISK-ARCH-004 | Recovery divergence between recorded and actual state | High | Open |
| RISK-PORT-001 | macOS Intel viability gap | Medium | Open — PENDING VALIDATION |
| RISK-PORT-002 | Linux isolation primitive fragmentation | High | Open — PENDING VALIDATION |
| RISK-PORT-003 | PAL abstraction leaks blocking the Windows phase | High | Open |

## Port interfaces frozen (chapter 02)

ProviderPort, ToolPort, MemoryStorePort, IndexerPort, EventBusPort, PermissionPort,
SecretStorePort, SandboxPort, ConfigPort, GitPort, TerminalPort, WorkspacePort,
SessionStorePort, SchedulerPort, UpdaterPort, PackagePort, AuthPort, TelemetryPort — 18
ports; names and signatures frozen per FR-ARCH-003.

## ADRs minted

Foundation block allocation (Volume 0 chapter 03: ADR-001–039); this volume used 030–033.

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-030](../annexes/adr/ADR-030.md) | Architectural style: layered hexagonal (ports and adapters) | Accepted | L0 Core Domain / L1 Ports / L2 Engines / L3 Adapters / L4 Drivers / L5 composition root, with the chapter 01 dependency matrix normative |
| [ADR-031](../annexes/adr/ADR-031.md) | Go module and package layout | Accepted | Two modules (product + `sdk/`); `internal/ports` holds all 18 ports; one `internal/<component>/` subtree per glossary component; SDK mirrors contracts and never imports `internal/` |
| [ADR-032](../annexes/adr/ADR-032.md) | Headless server mode | Accepted | Included as an operating mode of the same binary over the ADR-012 IPC surface, policy-only permissions, invoker-owned lifecycle, Beta phase |
| [ADR-033](../annexes/adr/ADR-033.md) | Dependency-rule enforcement | Accepted | One layer manifest generates depguard rules; an import-graph test and a prohibited-construct scanner complete the required CI checks |

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-ARCH-001 | Component wiring failure | 3 / 1 |
| E-ARCH-002 | Port contract violation | 1 |
| E-ARCH-003 | IPC endpoint unavailable | 1 |
| E-ARCH-004 | IPC protocol version unsupported | — (client: 2) |
| E-ARCH-005 | Task submission rejected | 1 when foreground |
| E-ARCH-006 | Forced shutdown | 8 |
| E-ARCH-007 | Recovery reconciliation failure | 9 |
| E-PORT-001 | Unsupported platform | 3 |
| E-PORT-002 | Platform capability unavailable | 3 when fatal |
| E-PORT-003 | Credential store backend unavailable | 3 |

## Events minted

Per the Volume 0 event grammar; payloads and envelope semantics per Volume 10.

| Event | Emitted by | Meaning |
|---|---|---|
| `pal.platform.rejected` | PAL | Startup refused on an unsupported platform (E-PORT-001) |
| `pal.capability.degraded` | PAL | A surface capability is absent or degraded (E-PORT-002/003) |
| `pal.fallback.engaged` | PAL | Directory-resolution fallback engaged (ADR-022) |
| `runtime.recovery.completed` | Runtime | Startup recovery finished, with counts (FR-ARCH-009) |
| `runtime.shutdown.completed` | Runtime | Shutdown finished, orderly or forced, with step timings (FR-ARCH-010) |
| `scheduler.task.rejected` | Task Scheduler | Bounded-pool submission rejected (E-ARCH-005) |
| `ipc.client.connected` | IPC server | A verified same-user client completed the handshake |
| `ipc.request.rejected` | IPC server | A connection or request was rejected (identity, version, or endpoint state) |

## Glossary additions

| Term | One-line meaning |
|---|---|
| Port interface | One of the 18 frozen L1 interfaces of chapter 02 through which engines reach infrastructure and extensions reach Andromeda. |
| Adapter | An L3 implementation of a port against an external system (provider, SQLite, git, OS); importable only by the composition root. |
| Driver | An L4 entry-point surface (CLI, TUI, IPC server) that steers the Runtime and is imported by nothing. |
| Composition root | `cmd/andromeda` — the only production code that constructs and wires engines with adapters (FR-ARCH-002). |
| Dependency matrix | The chapter 01 table declaring every ALLOWED/PROHIBITED import direction between layers (FR-ARCH-001). |
| Layer manifest | The machine-readable package-to-layer assignment from which ADR-033 enforcement is generated. |
| PAL surface | One of the 19 platform abstractions of chapter 07, each with per-OS backends, capability probes, and a declared degradation policy. |
| Supervision tree | The chapter 08 context/group hierarchy under which all concurrent work runs (ADR-023, FR-ARCH-006). |
| Headless mode | The Beta operating mode of the same binary driven solely through the IPC surface with policy-only permissions (ADR-032, FR-ARCH-008). |
| Process family | One of the four sanctioned child-process categories (plugins, MCP stdio servers, sandboxed tool/terminal children, git operations), each with exactly one supervisor (FR-ARCH-005). |

## Assumptions

Local list per Volume 0, chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | depguard (or a successor in the pinned golangci-lint) plus a Go-native graph test can express the full chapter 01 matrix including its exceptions | ADR-033 implementation spike at Core phase | Adopt a dedicated architecture-lint tool against the same manifest (ADR-033 alternative 3) |
| Technical assumption | The four-pool scheduler taxonomy (`interactive`, `tools`, `background`, `io`) is sufficient granularity for Volume 12's budgets | Volume 12 budget authoring; load benchmarks | Pools are configuration — the registry extends without contract changes |
| Technical assumption | UID-based peer verification on Unix domain sockets is available and reliable on all Tier 1 platforms via the PAL Local IPC surface | PAL conformance suite (Core) | Fall back to directory-permission-only access control with a documented weakening, or token handshake per Volume 9 |
| Technical assumption | Per-instance runtime directories (ADR-022) exist or are creatable with 0700 modes in all supported environments including containers and SSH sessions | Container/SSH CI jobs (FR-PORT-003 edge cases) | The Config Directories surface falls back to a state-root subdirectory, already specified |
| Product hypothesis | A persistent headless instance (ADR-032) is the shape editor and automation integrators actually want, versus embedding via the SDK | Beta-phase integration feedback (Volume 15 channels) | The IPC surface remains for tooling; SDK embedding would be a separate, additive decision |

## Open questions

Entries follow Volume 0, chapter 08; none blocks authoring. Every PENDING VALIDATION
occurrence in this volume maps to a row here (or to the originating ADR's own register
entry, noted).

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V3-OQ-1 | macOS Intel (x86_64) Tier 2 support — build/test capacity PENDING VALIDATION (also raised by Volume 1) | Chapter 07 platform matrix; RISK-PORT-001 | No — Tier 2 classification bounds the commitment | Validate CI capacity for Intel builds/smoke before MVP exit; record outcome via change procedure | Open |
| V3-OQ-2 | Linux OS-level isolation primitives — Landlock (kernel ≥ 5.13), namespaces, bubblewrap availability per reference distribution, PENDING VALIDATION per ADR-021 | Chapters 05/07 (Sandbox Engine, Sandbox surface); RISK-PORT-002 | No — process-level controls are the guaranteed floor | Per-mechanism validation during Beta implementation, resolved per ADR-021 review conditions | Open |
| V3-OQ-3 | Windows PAL backend mappings (named pipes, Credential Manager, ConPTY, Job Objects, known folders) — validation spike required before any v2 commitment | Chapter 07 Windows-future rules; RISK-PORT-003 | No — v2 candidate phase | Windows-phase spike validating each surface mapping; amends this volume | Open |
| V3-OQ-4 | MCP OAuth-based server authorization remains PENDING VALIDATION per ADR-010 (referenced by chapter 04) | Chapter 04, MCP Runtime | No — non-OAuth paths are the supported MVP surface | Tracked by ADR-010's review conditions; Volume 6 resolves | Open |
| V3-OQ-5 | go-git read-only fast paths remain individually PENDING VALIDATION per ADR-025 (referenced by chapter 05); official provider SDK adoption per adapter remains PENDING VALIDATION per ADR-019 (referenced by chapter 04) | Chapters 04/05 | No — subprocess git and stdlib HTTP are the defaults | Tracked by ADR-025/ADR-019 review conditions; Volumes 11/5 resolve | Open |

## Cross-volume references

The port-to-owner map other volumes elaborate against (full table with consumers in chapter
02):

| Port | Behavioral contract owner |
|---|---|
| ProviderPort, AuthPort | Volume 5 |
| ToolPort, TerminalPort | Volume 6 |
| PackagePort | Volume 6 (extension packages; release artifacts Volume 14) |
| MemoryStorePort, IndexerPort | Volume 7 |
| PermissionPort, SecretStorePort, SandboxPort | Volume 9 |
| ConfigPort, SessionStorePort, EventBusPort, TelemetryPort | Volume 10 |
| GitPort | Volume 11 |
| WorkspacePort | Volume 4 |
| SchedulerPort | Volume 3 (chapter 08); pool budgets Volume 12 |
| UpdaterPort | Volume 14 |

Other load-bearing references made by this volume: phases and platform tiers (Volume 1
chapter 05); entities, frozen states, and persistence conventions (Volume 2 chapters 01, 09,
10); event envelope and configuration schema (Volume 10); permission model and sandbox model
(Volume 9); test strategy and gates (Volume 13); distribution and update behavior
(Volume 14).
