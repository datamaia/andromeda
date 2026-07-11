# Volume 3 — System Architecture

**Status:** Authored (draft) · **Owner:** Architecture (Volume 3)

Volume 3 defines Andromeda's system architecture: the layered, ports-and-adapters structure of
the codebase, the frozen port interfaces that every later volume elaborates, the full component
inventory with boundaries and dependency rules, the Platform Abstraction Layer contract and
platform support matrix, the process/concurrency/IPC model, and the deployment, update,
extensibility, and compatibility strategy. Volumes 4–15 are written against the boundaries and
contracts fixed here; per Volume 0, chapter 03, this volume mints all `ARCH` and `PORT`
identifiers and the port interface names other volumes reference.

Foundations assumed: Volume 0 (conventions, glossary, ID taxonomy), Volume 1 (objectives,
principles, phases, platform scope), Volume 2 (entities, aggregates, canonical states,
persistence conventions).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Architecture Overview](01-architecture-overview.md) | Logical architecture (layers, hexagonal structure), physical architecture, module layout, dependency matrix, FR-ARCH-001/FR-ARCH-002 |
| [02 — Port Interfaces](02-port-interfaces.md) | The 18 frozen port interfaces with typed pseudocode, per-method semantics, error classes, concurrency rules, and owning volumes |
| [03 — Components: Core and Runtime](03-components-core.md) | Core Domain, Runtime, Agent Engine, Planner, Execution Engine, Context Manager, Memory Manager, Prompt Engine |
| [04 — Components: Platform Services](04-components-platform.md) | Provider Layer, Authentication Layer, Tool Runtime, Plugin Runtime, Workflow Engine, Skill Engine, MCP Runtime |
| [05 — Components: Infrastructure](05-components-infrastructure.md) | Configuration Manager, Persistence Layer, Event Bus, Task Scheduler, Indexing Engine, Workspace Engine, Git Engine, Terminal Engine, Sandbox Engine, Permission Manager, Secret Store, Audit Log, Policy Engine |
| [06 — Components: Interface and Delivery](06-components-interface.md) | CLI, TUI, Updater, Package Manager, Extension SDK, Telemetry, Logging, Observability |
| [07 — Platform Abstraction Layer](07-platform-abstraction-layer.md) | PAL contract, the 19 platform surfaces, platform support matrix, XDG layout, Windows-future rules, FR/NFR-PORT-* |
| [08 — Processes, Concurrency, and IPC](08-processes-concurrency-ipc.md) | Process model, supervised task model, external IPC surface, crash recovery, shutdown semantics, E-ARCH errors |
| [09 — Deployment, Update, Extensibility, Compatibility](09-deployment-update-extensibility-compatibility.md) | Deployment shapes, state locations, Updater integration points, extension point map, compatibility strategy |
| [99 — Volume Register](99-volume-register.md) | Everything this volume minted: requirements, ADRs, error codes, events, glossary additions, assumptions, open questions |

## Reading guide for parallel volume authors

1. Chapter 02 is the spine: port names and signatures there are frozen. An owning volume
   elaborates semantics behind a port; it MUST NOT rename ports or methods or change signatures.
2. Chapters 03–06 give each component's boundary, dependencies, and phase; the owning volume of
   each component's area specifies its internal behavior.
3. Chapter 07 binds every OS-specific behavior to the PAL; no other volume may introduce
   platform-conditional behavior outside it.
