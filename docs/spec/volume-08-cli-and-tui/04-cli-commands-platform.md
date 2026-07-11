# 04 — CLI Commands: Platform

This chapter specifies the platform-management families: `provider`, `model`, `tool`,
`plugin`, `skill`, `workflow`, and `mcp`. Chapter [02](02-cli-conventions.md) conventions
apply throughout. Command phases follow the surfaced subsystem: `provider`, `model`, and
`tool` are **MVP**; `plugin`, `skill`, `workflow`, and `mcp` are **Beta**, matching the
phases of their engines (Volumes 4 and 6) — the grammar is fixed now so the tree does not
shift when they land. FR-CLI-014 binds these sections normatively.

## `andromeda provider`

Provider registration and health over the Provider Layer (contract owner Volume 5;
FR-PROV-001). States rendered are the frozen Provider connection states (Volume 2,
chapter 09).

```text
andromeda provider list
andromeda provider show <name>
andromeda provider add <name> --adapter <adapter> [--endpoint <url>] [--set key=value ...]
andromeda provider remove <name>
andromeda provider test <name>
andromeda provider enable <name>
andromeda provider disable <name>
```

Behavior: `add` registers a configured Provider against an installed adapter; adapter
identity and per-adapter settings are validated per Volume 5's declaration set before any
row is written; credentials are never taken here (`auth login` owns them). `test` drives
verification (`verifying` → `available`/`degraded`/`unavailable`) and reports declared
capabilities discovered. `remove` is a destructive confirmation (registration and routing
configuration are removed; the row is tombstoned `removed`). `disable`/`enable` toggle
administrative exclusion (`disabled`).

**Permissions:** `write` (configuration) for mutations; `network` for `test`.
**Exit codes:** 0, 1, 2, 3, 5 (`test` adds 4, 7, 8).

**JSON result (`provider list`, `data`):**

```json
{
  "providers": [
    {
      "name": "local-ollama",
      "adapter": "ollama",
      "state": "available",
      "endpoint": "http://localhost:11434",
      "default": true
    }
  ]
}
```

**Examples:**

```text
$ andromeda provider add local-ollama --adapter ollama            # valid
$ andromeda provider test local-ollama --json                     # valid
$ andromeda provider add x --adapter nonexistent                  # invalid
error[E-CLI-002]: invalid value for --adapter: no installed adapter named "nonexistent"
```

**Errors:** E-CLI-002/003; E-PROV family (Volume 5) exit 7; E-AUTH family exit 4 on `test`
with failing credentials.

## `andromeda model`

Model catalog and capability inspection (capability enum owned by Volume 5).

```text
andromeda model list [--provider <name>] [--refresh]
andromeda model show <model-id> [--provider <name>]
andromeda model capabilities <model-id> [--provider <name>]
andromeda model default <model-id> [--provider <name>]
```

Behavior: `list` renders the local Model catalog; `--refresh` re-runs discovery
(ProviderPort `DiscoverModels`) and requires network for remote providers. `capabilities`
renders the declared CapabilitySet using the closed capability names (`chat`, `streaming`,
`tool_calling`, `parallel_tool_calling`, `structured_outputs`, `reasoning`, `vision`,
`audio_input`, `audio_output`, `embeddings`, `token_usage_reporting`, `cost_reporting`,
`model_discovery`, `cancellation`) — never inferred, never simulated (Principle 2).
`default` sets the profile-level default model; validation confirms the model exists in
the catalog.

**Permissions:** `write` (configuration) for `default`; `network` for `--refresh`.
**Exit codes:** 0, 1, 2, 3 (`--refresh` adds 4, 7, 8; `default` adds 5).

**JSON result (`model capabilities`, `data`):**

```json
{
  "provider": "local-ollama",
  "model": "qwen3:14b",
  "capabilities": ["chat", "streaming", "tool_calling"],
  "context_window": 32768,
  "declared_at": "2026-07-11T10:00:00Z"
}
```

**Examples:**

```text
$ andromeda model list --provider local-ollama                    # valid
$ andromeda model default qwen3:14b --provider local-ollama       # valid
$ andromeda model capabilities                                    # invalid
error[E-CLI-002]: invalid value for <model-id>: a model identifier is required
```

**Errors:** E-CLI-002; E-PROV family exit 7.

## `andromeda tool`

Tool registry surface over the Tool Runtime (contract owner Volume 6; FR-TOOL-001). Origin
and trust level are always visible (Principle 4).

```text
andromeda tool list [--origin builtin|plugin|mcp] [--enabled|--disabled]
andromeda tool show <tool-name>
andromeda tool enable <tool-name>
andromeda tool disable <tool-name>
andromeda tool test <tool-name> [--input <json>]
```

Behavior: `show` renders the full declaration — identity, version, schemas, permission
declaration, timeouts, limits, origin, trust level. `test` performs schema and semantic
validation of `--input` via ToolPort `Validate` **without executing** (side-effect-free by
contract); execution is `exec tool` (chapter 03). `disable` removes the tool from agent
visibility workspace-wide; `enable` restores it.

**Permissions:** `write` (configuration) for enable/disable; none for reads and `test`.
**Exit codes:** 0, 1, 2, 3, 5, 6 (`test` maps validation findings to 0 with findings in
data when the input is merely invalid against the schema; 6 only for tool-runtime
failures).

**JSON result (`tool test`, `data`):**

```json
{
  "tool": "fs.write",
  "valid": false,
  "findings": [
    {"path": "$.content", "message": "required property missing"}
  ]
}
```

**Examples:**

```text
$ andromeda tool list --origin mcp                                # valid
$ andromeda tool test fs.write --input '{"path":"a.txt"}'         # valid (reports findings)
$ andromeda tool show                                             # invalid
error[E-CLI-002]: invalid value for <tool-name>: a tool name is required
```

**Errors:** E-CLI-002; E-TOOL family (Volume 6) exit 6.

## `andromeda plugin`

Plugin lifecycle over the Package Manager and Plugin Runtime (Volume 6; FR-PLUG-001;
Plugin and Package installation states are the frozen vocabularies).

```text
andromeda plugin list
andromeda plugin show <name>
andromeda plugin install <source> [--version <constraint>]
andromeda plugin uninstall <name>
andromeda plugin enable <name>
andromeda plugin disable <name>
```

Behavior: `<source>` is a local archive/path or a registry reference (registry phasing per
Volume 6); `install` streams the frozen installation states (`resolving` … `installed`) as
progress, verifying checksums/signatures per trust policy (Volume 9) before anything
activates — a failed step leaves nothing partially active. `uninstall` is a destructive
confirmation; it stops the plugin through its machine and tombstones the Package row.
`show` includes process state, declared surfaces (tools, commands), and trust level.

**Permissions:** `package_installation`; `network` for remote sources; `process_spawn` is
exercised by the Plugin Runtime at start, not by this command directly.
**Exit codes:** 0, 1, 2, 3, 5, 6, 8 (installation verification failures map per the
E-PLUG family envelope, exit 6; nothing maps to 9 — extension failures never corrupt
product state).

**JSON result (`plugin list`, `data`):**

```json
{
  "plugins": [
    {
      "name": "fmt-suite",
      "version": "1.2.0",
      "state": "running",
      "package_state": "installed",
      "trust": "third_party",
      "surfaces": {"tools": 2, "commands": 1}
    }
  ]
}
```

**Examples:**

```text
$ andromeda plugin install ./fmt-suite-1.2.0.tar.gz               # valid
$ andromeda plugin uninstall fmt-suite --yes                      # valid, scripted
$ andromeda plugin install                                        # invalid
error[E-CLI-002]: invalid value for <source>: a path or registry reference is required
```

**Errors:** E-CLI-002/003; E-PLUG family (Volume 6) exit 6.

## `andromeda skill`

Skill management over the Skill Engine (format owner Volume 6; FR-SKILL-001). Same
shape as `plugin` — skills are packaged extensions with the same installation states.

```text
andromeda skill list
andromeda skill show <name>
andromeda skill install <source> [--version <constraint>]
andromeda skill uninstall <name>
andromeda skill enable <name>
andromeda skill disable <name>
```

Behavior deltas from `plugin`: `show` renders the skill manifest surface — identity,
version, required tools and capabilities, compatible providers, composition/inheritance
declarations (Volume 6 vocabulary); skills have no process state (no `running`), only
registration and enablement. `install` validates the manifest against the Volume 6 format
before staging.

**Permissions:** `package_installation`; `network` for remote sources.
**Exit codes:** 0, 1, 2, 3, 5, 6, 8.

**JSON result (`skill show`, `data`):**

```json
{
  "name": "conventional-commits",
  "version": "2.0.1",
  "enabled": true,
  "requires": {"tools": ["git.commit"], "capabilities": ["tool_calling"]},
  "trust": "third_party"
}
```

**Examples:**

```text
$ andromeda skill install ./skills/conventional-commits            # valid
$ andromeda skill enable conventional-commits                      # valid
$ andromeda skill uninstall                                        # invalid
error[E-CLI-002]: invalid value for <name>: an installed skill name is required
```

**Errors:** E-CLI-002/003; E-SKILL family (Volume 6) exit 6.

## `andromeda workflow`

Workflow definitions and Workflow Runs (owner Volume 4; FR-WF-001; frozen Workflow Run
states).

```text
andromeda workflow list [--runs]
andromeda workflow show <name-or-run-id>
andromeda workflow run <name> [--input <key=value> ...] [--session <id>]
andromeda workflow status <run-id>
andromeda workflow resume <run-id>
andromeda workflow cancel <run-id>
andromeda workflow validate <path>
```

Behavior: `run` instantiates a Workflow Run and streams stage transitions (human progress
per FR-UX-003; NDJSON per FR-CLI-006); approval gates block interactively or resolve from
policy non-interactively (FR-CLI-009) — an unresolved gate denies and the run records its
state honestly. `resume` continues `paused`/`interrupted` runs at the last persisted step
boundary (Volume 4 semantics). `cancel` requests cancellation; the run records
`cancelled`. `validate` checks a workflow definition file against the Volume 4 format and
reports all findings.

**Permissions:** none at the CLI layer; each workflow step's actions are mediated
individually.
**Exit codes:** 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 (`validate`: 0, 1, 2, 3).

**JSON result (`workflow status`, `data`):**

```json
{
  "run_id": "01JZX0WF000000000000000000",
  "workflow": "sdd-feature",
  "state": "awaiting_approval",
  "stage": "architecture",
  "stages_completed": 4,
  "started_at": "2026-07-11T09:12:00Z"
}
```

**Examples:**

```text
$ andromeda workflow run sdd-feature --input spec=./spec.md       # valid
$ andromeda workflow resume 01JZX0WF --json --no-input            # valid, CI resume
$ andromeda workflow run                                          # invalid
error[E-CLI-002]: invalid value for <name>: a workflow name is required
```

**Errors:** E-CLI-002; E-WF family (Volume 4); E-SEC family exit 5 for denied gates.

## `andromeda mcp`

MCP server registrations and connections over the MCP Runtime (owner Volume 6; FR-MCP-001;
ADR-010 official SDK; frozen MCP Client Connection states).

```text
andromeda mcp list
andromeda mcp show <name>
andromeda mcp add <name> (--command <argv> | --url <url>) [--env KEY=VALUE ...]
andromeda mcp remove <name>
andromeda mcp enable <name>
andromeda mcp disable <name>
andromeda mcp status [<name>]
andromeda mcp tools <name>
```

Behavior: `add` registers a server with exactly one transport — a stdio subprocess
(`--command`) or a remote endpoint (`--url`); transport support follows the official SDK
(ADR-010). Registration alone connects nothing (`configured`); connection policy is
Volume 6's. `status` renders connection states; `tools` lists tools discovered from a
`ready` connection, marked with origin `mcp` and the server's trust level. `remove` is a
destructive confirmation and tombstones the registration.

**Permissions:** `write` (configuration) for mutations; `external_service_access` and
`network` (remote) or `process_spawn` (stdio) are exercised at connection time under
Volume 6/9 policy.
**Exit codes:** 0, 1, 2, 3, 5 (`status`/`tools` add 6, 8).

**JSON result (`mcp status`, `data`):**

```json
{
  "servers": [
    {
      "name": "docs-server",
      "transport": "stdio",
      "state": "ready",
      "tools_discovered": 4
    }
  ]
}
```

**Examples:**

```text
$ andromeda mcp add docs-server --command "npx docs-mcp"          # valid
$ andromeda mcp tools docs-server --json                          # valid
$ andromeda mcp add docs-server --command "npx x" --url http://y  # invalid
error[E-CLI-006]: flags --command and --url cannot be combined
```

**Errors:** E-CLI-002/003/006; E-MCP family (Volume 6) exit 6.

## Requirements

### FR-CLI-014 — Platform command family behavior

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: CLI (Volume 8)
- Affected components: CLI; Provider Layer; Tool Runtime; Plugin Runtime; Skill Engine; Workflow Engine; MCP Runtime; Package Manager
- Dependencies: FR-CLI-001, FR-CLI-005–FR-CLI-012; FR-PROV-001 (Volume 5); FR-TOOL-001, FR-PLUG-001, FR-SKILL-001, FR-MCP-001 (Volume 6); FR-WF-001 (Volume 4)
- Related risks: RISK-CLI-001, RISK-CLI-003

#### Description

The commands specified in this chapter — `provider`, `model`, `tool`, `plugin`, `skill`,
`workflow`, `mcp` — MUST behave exactly as their sections define. The family's shared
invariants: resource groups use the shared verb vocabulary (FR-CLI-001); every state
rendered is a frozen vocabulary name; origin and trust level are visible wherever
third-party surface is listed; installation-class operations stream frozen installation
states and never leave partial activations; removal-class operations are destructive
confirmations. The `provider`, `model`, and `tool` groups are MVP; `plugin`, `skill`,
`workflow`, and `mcp` groups are Beta — their grammar is reserved by this requirement from
Core so later arrival is additive.

#### Motivation

Platform management is how users exercise vendor independence (PRD-002) and extensibility
(PRD-007) without editing files; a uniform family shape is what makes fifteen subsystems
feel like one product (ADR-100).

#### Actors

Users; automation managing fleets of configurations (UC-08); the engines surfaced.

#### Preconditions

The surfaced subsystem is present at its phase; otherwise the group's commands fail with a
phase-honest diagnostic (E-CLI-001 semantics do not apply — the command exists; the
subsystem reports unavailability through its own family).

#### Main flow

1. A caller invokes a family command per its syntax.
2. The command executes through the owning engine's port/API.
3. Output, exit code, and records match the section's declarations.

#### Alternative flows

- `--json` on every command per FR-CLI-006; installation progress as NDJSON stream
  documents.

#### Edge cases

- `provider remove` with dependent profiles: refused with the E-PROV family's dependency
  error naming the dependents (never cascades silently).
- `plugin install` of an already-installed version: idempotent success with
  `data.changed: false`.
- `mcp add` transport exclusivity (`--command` xor `--url`): E-CLI-006.
- `tool test` on a disabled tool: validation still runs (declaration is readable);
  execution paths remain blocked.

#### Inputs

Resource names, sources (paths, registry references), version constraints, transport
declarations, workflow inputs.

#### Outputs

Listings, declarations, installation progress, per the sections' JSON schemas.

#### States

Frozen vocabularies rendered verbatim: Provider connection, Plugin, MCP Client Connection,
Package installation, Workflow Run, Tool Invocation (via `tool test` findings only — no
invocation is created).

#### Errors

E-CLI family plus the owning families per section: E-PROV (7), E-TOOL/E-PLUG/E-SKILL/E-MCP
(6), E-WF (per envelope), E-AUTH (4) where credentials are exercised.

#### Constraints

`provider add` never takes credentials; `tool test` never executes; installation never
activates unverified content (Volume 9 trust policy gate); no group bypasses the Package
Manager for artifact handling.

#### Security

Mutations require the permission classes stated per section; everything third-party is
labeled with origin and trust level; signature verification precedes activation per
Volume 6/9 policy.

#### Observability

Family commands emit chapter 02 lifecycle events; installations and connections emit their
owning volumes' event families, correlated to the invocation.

#### Performance

Listings are local reads; only `test`, `--refresh`, and remote installs touch the network,
and each declares it.

#### Compatibility

Grammar identical across platforms and phases; Beta groups parse at MVP and fail with
phase-honest diagnostics rather than "unknown command" (grammar reserved from Core).

#### Acceptance criteria

- Given the section syntax lines, when grammar golden tests parse each, then every line
  binds to its documented handler and every verb is in the shared or declared-domain set.
- Given `plugin install` of a tampered archive fixture, when verification fails, then no
  file is activated, the Package row records `failed`, exit is 6, and the envelope carries
  the E-PLUG family code (negative + security case).
- Given `provider test` against a stopped local server, when it completes, then the
  Provider records `unavailable`, exit is 7, and `data.state` matches (state-honesty
  case).
- Given any removal command non-interactively without `--yes`, when invoked, then
  E-CLI-003 and the resource is intact (permission/confirmation case).
- Observability: installation runs produce correlated `cli.command.*` and package events.

#### Verification method

Volume 13 CLI suite: grammar goldens, schema conformance per command, tampered-artifact
and stopped-server fixtures, confirmation matrix; UC-12 E2E journey (plugin + MCP) at
Beta.

#### Traceability

PRD-002, PRD-004, PRD-007; UC-10, UC-12; FR-PROV-001, FR-TOOL-001, FR-PLUG-001,
FR-SKILL-001, FR-MCP-001, FR-WF-001; ADR-100, ADR-104.
