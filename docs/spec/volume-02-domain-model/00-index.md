# Volume 2 — Domain Model

**Status:** Complete · **Owner:** Domain model author (Volume 2)

Volume 2 defines Andromeda's domain model: every domain entity, its attributes, identifiers,
relations, integrity invariants, lifecycle, persistence, versioning, and serialization. It is
the single authoritative home for entity shapes and for the **canonical state names** of every
stateful entity (Volume 0, chapter 03, single-home matrix). The *behavior* of each state
machine — transitions, guards, side effects, recovery, timeouts, cancellation, retries, and
errors — is owned by the volume named in [chapter 09](09-lifecycles-and-canonical-states.md).

Volume 2 is definitional: it mints no `FR-*`, `NFR-*`, or `RISK-*` identifiers and no error
codes. Behavioral requirements over these entities live in the owning volumes (4–14). This
volume mints three foundation decisions: ADR-027 (ULID entity identifiers), ADR-028 (workspace
and global database split), and ADR-029 (forward-only migrations).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Overview, Modeling Approach, and Aggregates](01-overview-and-aggregates.md) | Modeling rules, entity classification, aggregate map, domain views, identifier strategy |
| [02 — Workspace and Configuration](02-workspace-and-configuration.md) | Workspace, Project, Configuration Profile |
| [03 — Sessions, Agents, and Runs](03-sessions-agents-and-runs.md) | Session, Agent, Agent Profile, Run, Turn, Message, Plan, Task |
| [04 — Tools and Actions](04-tools-and-actions.md) | Tool, Tool Invocation, Tool Result, Approval, Permission, Artifact, File Change, Patch, Command Execution |
| [05 — Providers and Authentication](05-providers-and-authentication.md) | Provider, Model, Capability, Credential, Authentication Session |
| [06 — Extensibility and Distribution](06-extensibility-and-distribution.md) | Workflow, Workflow Run, Skill, Plugin, MCP Server, MCP Client Connection, Package, Extension, Release |
| [07 — Memory, Context, and Indexing](07-memory-context-and-indexing.md) | Memory Record, Context Item, Index, Embedding |
| [08 — Observability and Accounting](08-observability-and-accounting.md) | Event, Trace, Metric, Cost Record, Audit Record |
| [09 — Lifecycles and Canonical States](09-lifecycles-and-canonical-states.md) | Canonical state enums (frozen), full-machine ownership table, recorded status vocabularies |
| [10 — Persistence, Serialization, and Migrations](10-persistence-serialization-migrations.md) | SQLite schema conventions, database split, canonical JSON, schema versioning, forward-only migrations |
| [99 — Volume 2 Register](99-volume-register.md) | Entity catalog, state enum register, ADRs minted, glossary additions, assumptions, open questions |

## How to read this volume

Chapters 02–08 use one uniform entity template (defined in chapter 01): Purpose, Attributes,
Identifiers, Relations, Integrity invariants, Lifecycle, Persistence, Versioning and
serialization. Chapter 09 freezes state names for the whole corpus. Chapter 10 binds the model
to the Persistence Layer conventions of ADR-007 (SQLite via `modernc.org/sqlite`, WAL mode) and
ADR-022 (XDG directories, project state in `.andromeda/`).
