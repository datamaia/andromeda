# 04 — Glossary

One vocabulary for the whole corpus. A concept has exactly one name; a name refers to exactly
one concept. New terms are added via each volume's `99-volume-register.md` and merged here during
consolidation. Entity semantics are owned by Volume 2; component semantics by Volume 3; this
glossary gives the corpus-wide one-line meaning.

## Domain entities

| Term | Meaning |
|---|---|
| Workspace | The root working environment Andromeda operates in: a directory tree plus its Andromeda state (`.andromeda/`), settings, and indexes. |
| Project | A logical unit inside a workspace with its own configuration profile and metadata (commonly a repository). |
| Session | A bounded interactive or non-interactive engagement with Andromeda, holding runs, context, and session memory; persistable and resumable. |
| Agent | An autonomous executor that plans and acts through tools under a profile, permissions, and observability. |
| Agent Profile | A named, versioned configuration of an agent: model/provider selection, prompts, tool set, permission defaults, and behavioral parameters. |
| Run | One top-level execution of an agent or workflow within a session, from intake to terminal state. |
| Turn | One request/response exchange inside a run. |
| Message | A single unit of conversation content (user, agent, system, or tool) inside a turn. |
| Plan | A structured, inspectable set of intended steps produced by the Planner for a run. |
| Task | A unit of executable work derived from a plan, with its own state machine. |
| Tool | A named, versioned, schema-typed capability an agent can invoke, with declared permissions and limits. |
| Tool Invocation | One call of a tool with concrete inputs, under a granted permission set. |
| Tool Result | The typed output (or error) of a tool invocation. |
| Approval | A recorded human decision granting or denying a requested action or permission. |
| Permission | A grant to perform a class of side-effecting action within a scope (see Volume 9). |
| Artifact | A durable output produced by a run (file, patch, report, export). |
| File Change | A recorded modification to a file (create, edit, delete, rename) attributable to a run. |
| Patch | A reviewable diff representing one or more file changes. |
| Command Execution | A recorded execution of a terminal command, with its inputs, environment policy, and outcome. |
| Provider | An adapter-backed source of model inference (cloud service or local server). |
| Model | A concrete inference target exposed by a provider, with declared capabilities. |
| Capability | A declared, machine-checkable ability of a model/provider (e.g., tool calling, streaming, vision); closed enum owned by Volume 5. |
| Credential | A secret used to authenticate against a provider or service; stored per Volume 9. |
| Authentication Session | The state of an authenticated identity against a provider (tokens, expiry, refresh). |
| Workflow | A declared, stateful orchestration of agents, tools, and approvals (see Volume 4). |
| Workflow Run | One execution instance of a workflow. |
| Skill | A packaged, versioned unit of procedural knowledge (prompts + required tools/capabilities) loadable by agents. |
| Plugin | An external extension process integrated through the plugin runtime and Andromeda Runtime Protocol. |
| MCP Server | An external Model Context Protocol server offering tools, resources, or prompts. |
| MCP Client Connection | Andromeda's managed connection to one MCP server. |
| Memory Record | A persisted unit of memory (session, workspace, long-term, semantic, episodic) with provenance and retention. |
| Context Item | One candidate unit of content assembled into a model request by the Context Manager. |
| Index | A queryable structure over workspace content (lexical or semantic). |
| Embedding | A vector representation of content used by semantic indexing and retrieval. |
| Event | A structured, versioned occurrence emitted on the event bus (envelope owned by Volume 10). |
| Trace | A correlated tree of spans describing one run's execution across components. |
| Metric | A named quantitative measurement emitted by the runtime. |
| Cost Record | An accounting entry for tokens/spend attributed to a run, provider, and model. |
| Audit Record | An immutable entry documenting a security-relevant action (who, what, when, under which permission). |
| Configuration Profile | A named set of configuration values selectable at global, project, or invocation level. |
| Package | A distributable unit (binary, plugin, skill) with version, checksum, and signature metadata. |
| Extension | Any third-party addition: tool, plugin, skill, provider adapter, exporter, or command. |
| Release | A published, versioned distribution of Andromeda with artifacts, notes, and provenance. |

## Architecture components

| Term | Meaning |
|---|---|
| Core Domain | Pure domain model and invariants; no I/O, no provider or platform specifics. |
| Runtime | The composed engine layer that executes sessions, runs, and workflows. |
| Agent Engine | Drives the agent loop: planning, acting, observing, iterating. |
| Planner | Produces and revises plans from goals and context. |
| Execution Engine | Executes plan tasks, dispatching tools and managing task states. |
| Context Manager | Selects, ranks, budgets, and assembles context items for model requests. |
| Memory Manager | Ingests, stores, retrieves, and expires memory records. |
| Provider Layer | Common provider contract plus per-provider adapters. |
| Authentication Layer | Credential acquisition, storage integration, refresh, rotation, revocation. |
| Tool Runtime | Registers, validates, sandboxes, executes, and observes tools. |
| Plugin Runtime | Manages plugin processes and their protocol lifecycle. |
| Workflow Engine | Executes workflow definitions as state machines with approvals and resumability. |
| Skill Engine | Loads, validates, composes, and applies skills. |
| Prompt Engine | Renders versioned prompt templates with context and profile parameters. |
| MCP Runtime | Manages MCP client connections, discovery, and lifecycle. |
| Configuration Manager | Loads, validates, migrates, and resolves configuration with precedence. |
| Telemetry | Collects and exports metrics/usage signals under consent policy. |
| Logging | Structured, redacted, local-first logs. |
| Observability | Correlates logs, events, traces, metrics, and cost records. |
| Indexing Engine | Builds and incrementally updates lexical/semantic indexes. |
| Workspace Engine | Discovers, opens, snapshots, and manages workspaces and projects. |
| Sandbox Engine | Applies isolation policies to tool/command execution. |
| Permission Manager | Evaluates permission requests against grants, scopes, and policies. |
| Git Engine | Encapsulated Git operations (status, diff, commit, branch, worktree, …). |
| Terminal Engine | PTY management and command execution with capture and limits. |
| CLI | The `andromeda` command-line interface. |
| TUI | The interactive terminal user interface. |
| Updater | Checks, downloads, verifies, applies, and rolls back updates. |
| Package Manager | Installs, verifies, updates, and removes packages/extensions. |
| Extension SDK | Public contracts for building tools, plugins, skills, and adapters. |
| Persistence Layer | SQLite-backed storage for sessions, memory, indexes, audit, and state. |
| Event Bus | In-process typed publish/subscribe channel for events. |
| Task Scheduler | Schedules and supervises concurrent tasks with cancellation and backpressure. |
| Platform Abstraction Layer (PAL) | Encapsulates OS-specific behavior (filesystem, processes, signals, PTY, credential store, …). |
| Secret Store | The credential-storage abstraction over OS keychains and the encrypted-file fallback. |
| Audit Log | Append-only store of audit records. |
| Policy Engine | Evaluates configured policies (permissions, telemetry, provider routing constraints). |

## Product and process terms

| Term | Meaning |
|---|---|
| AI Engineering Harness | The product category Andromeda belongs to: a runtime + tool environment for engineering with AI agents. |
| Andromeda Runtime Protocol | The JSON-RPC 2.0-based protocol between Andromeda and plugin processes (see the decision register, chapter 06). |
| Phase | A delivery stage: `Core`, `MVP`, `Beta`, `v1`, `v2`, `Future`, `Out of Scope` (defined in Volume 1). |
| Keystone requirement | A hub requirement pre-listed for cross-volume reference. |
| Local-first | Core functionality operates without Internet when models, tools, indexes, memory, and repositories are local. |
| Vendor-agnostic | No provider-specific logic outside that provider's adapter. |
| Model-agnostic | Behavior driven by declared capabilities, not by assumptions about specific models. |
| Design token | A named brand value (color, typography role) from the brand definition; the closed set is specified in Volume 8. |
| Volume register | The `99-volume-register.md` file at the end of each content volume, listing everything that volume minted. |
| Single-home matrix | The table in chapter 03 assigning each cross-cutting topic exactly one authoritative volume. |
| Spec linter | `scripts/spec_lint.py`, the mechanical enforcement of Volume 0. |
