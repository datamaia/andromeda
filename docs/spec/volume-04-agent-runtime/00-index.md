# Volume 4 — Agent Runtime

**Status:** Authored (draft) · **Owner:** Agent runtime (Volume 4)

Volume 4 specifies the execution heart of Andromeda: the Agent Engine and its plan–act–observe
loop, the Planner, the Execution Engine, the Prompt Engine, the full state machines for the
core execution entities (Session, Run, Agent, Plan, Task), the Workflow Engine with the
specification-driven development workflow, the Workflow Run state machine, the Skill Engine's
runtime semantics, and the Task Scheduler's behavioral elaboration. Entity shapes and
invariants are Volume 2's; component boundary contracts are Volume 3's; this volume owns the
behavior. Volume 4 mints identifiers in the `AGT` and `WF` areas and ADRs 040–054.

## Chapters

| Chapter | Contents | Status |
|---|---|---|
| [01 — Agent Engine](01-agent-engine.md) | The agent loop (keystone FR-AGT-001): plan–act–observe over ports, turn handling, interruption and resume, delegation, budgets, workspace and session intake | Authored (draft) |
| [02 — Planner](02-planner.md) | Plan production and revision, direct-execution plans, plan approval interplay | Authored (draft) |
| [03 — Execution Engine](03-execution-engine.md) | Task dispatch over the Tool Runtime and SchedulerPort, approvals, retries, cancellation, error propagation | Authored (draft) |
| [04 — Prompt Engine](04-prompt-engine.md) | Versioned prompt templates, registry, deterministic rendering, profile parameters | Authored (draft) |
| [05 — Core State Machines](05-core-state-machines.md) | Full machines for Session, Run, Agent, Plan, and Task under the frozen state names | Authored (draft) |
| 06 — Workflow Engine and SDD (`06-workflow-engine-and-sdd.md`) | Keystone FR-WF-001: the workflow engine and the 14-stage specification-driven development workflow | Authored (draft) |
| 07 — Workflow Run State Machine (`07-workflow-run-state-machine.md`) | Full machine for Workflow Run | Authored (draft) |
| 08 — Skill Engine Runtime (`08-skill-engine-runtime.md`) | Execution semantics for skills (format is Volume 6's) | Authored (draft) |
| 09 — Task Scheduler (`09-task-scheduler.md`) | SchedulerPort behavioral elaboration and supervision per ADR-023 | Authored (draft) |
| 99 — Volume Register (`99-volume-register.md`) | Everything Volume 4 minted; assembled from the per-agent register fragments at consolidation | Pending merge |

## Reading order and dependencies

Chapters 01–05 form the agent runtime core and are best read in order; chapters 06–09 build
the workflow layer on top of them. Prerequisites from other volumes: Volume 2 chapters 03 and
09 (entity shapes and frozen state names), Volume 3 chapters 02, 03, and 08 (ports, component
boundaries, concurrency and recovery). Downstream volumes consume this volume's contracts:
Volume 6 (Tool Runtime dispatch), Volume 7 (context assembly for turns), Volume 8 (CLI/TUI
presentation of runs, plans, and tasks), Volume 9 (approvals raised by this volume's engines),
Volume 10 (persistence of the run record stream), and Volume 12 (budgets for the loop's
latency and concurrency).
