# 03 — CLI Commands: Core

This chapter specifies the execution-and-setup command families: the bare root, `run`,
`plan`, `exec`, `init`, `session`, `config`, and `auth`. Chapter [02](02-cli-conventions.md)
conventions (global flags, output contract, stream discipline, confirmations, error
presentation) apply everywhere; sections below state only command-specific behavior. Every
command listed here is **Phase: MVP**. FR-CLI-013 at the end of the chapter binds these
sections normatively.

## `andromeda` (root)

Specified by FR-CLI-003 (chapter [01](01-cli-architecture.md)): bare interactive invocation
hands off to the TUI; bare non-interactive invocation prints short usage to stderr and
exits 2; `--version` aliases `version`.

```text
andromeda [global flags]
```

**Exit codes:** 2 (non-interactive bare invocation, `--json` bare invocation); after TUI
hand-off, the TUI session's closing outcome (0, 1, 8, 9). **Permissions:** none at the CLI
layer. **JSON result:** none (E-CLI-002 for `--json` bare). **Errors:** E-CLI-002; E-TUI
family after hand-off.

## `andromeda run`

Executes one agent run — the product's primary action (UC-01, UC-02) — streaming activity
while it executes and reporting the recorded outcome.

```text
andromeda run [flags] [<goal>...]
andromeda run --file <path>          # goal from file; "-" reads stdin
echo "goal" | andromeda run          # goal from piped stdin
```

| Flag | Default | Meaning |
|---|---|---|
| `--file`, `-f` | — | Goal text from file or stdin (`-`); exclusive with positional goal |
| `--agent` | configured default profile | Agent Profile name |
| `--provider` | profile's provider | Provider override for this run |
| `--model` | profile's model | Model override for this run |
| `--session` | new session | Attach the run to an existing session (ULID/prefix) |
| `--max-turns` | `[agent]` configuration (Volume 4) | Turn ceiling; exceeding cancels the run |
| `--budget-tokens` | unlimited | Token budget; exhaustion cancels with recorded reason |
| `--budget-cost` | unlimited | Cost budget in the accounting currency (Volume 5) |

Behavior: goal sources are exclusive (positional, `--file`, piped stdin) — more than one is
E-CLI-006. Human mode streams model output and tool-activity lines to stdout (payload) with
progress per FR-UX-003; `--json` streams NDJSON per FR-CLI-006. Interactive plan approval
and permission prompts follow the run's policy; non-interactively both resolve from policy
(FR-CLI-009). Interruption (`interrupt` signal, `--timeout`, budget exhaustion) cancels the
run through its context; the run records `cancelled` and the CLI exits 8.

**Permissions:** none for the command itself; every tool side effect inside the run is
mediated individually (Principle 8) and may touch any permission class its tools declare.
**Exit codes:** 0, 1, 2, 3, 4, 5, 6, 7, 8, 9.

**JSON result (`data`):**

```json
{
  "run_id": "01JZX0A1B2C3D4E5F6G7H8J9K0",
  "session_id": "01JZX0A1B2C3D4E5F6G7H8J9K1",
  "state": "completed",
  "turns": 6,
  "usage": {"input_tokens": 18234, "output_tokens": 4120},
  "cost": {"amount": "0.0000", "currency": "USD", "estimated": true},
  "files_changed": 3,
  "commands_executed": 2,
  "artifacts": [],
  "duration_ms": 184220
}
```

`state` is always a frozen Run state name (Volume 2, chapter 09).

**Examples:**

```text
$ andromeda run "add input validation to the signup handler"      # valid
$ andromeda run --json --no-input -f spec.md                      # valid, CI shape
$ andromeda run --file spec.md "also a positional goal"           # invalid
error[E-CLI-006]: flags --file and a positional goal cannot be combined
```

**Errors:** E-CLI-002/006/007/009; E-AGT family (Volume 4) for run failures; E-PROV family
exit 7; E-SEC family exit 5; E-CFG family exit 3.

## `andromeda plan`

Produces and manages Plans without executing them (inspectable autonomy, PRD-006; Plan
states per Volume 2, chapter 09; full machine Volume 4).

```text
andromeda plan new [flags] <goal>...      # produce a plan; state "proposed"
andromeda plan list [--session <id>] [--state <plan-state>]
andromeda plan show <plan-id>
andromeda plan approve <plan-id> [--run]  # approve; --run starts execution
andromeda plan reject <plan-id> [--reason <text>]
```

`plan new` accepts `--agent`, `--provider`, `--model`, `--session`, `--file` with `run`
semantics. `plan approve --run` is the two-step equivalent of `run` for policies that
require inspected plans; non-interactively it is the sanctioned way to execute with plan
inspection in between (UC-07). `plan show` renders steps, task derivations, and revision
lineage (`superseded` chains).

**Permissions:** none at the CLI layer (planning has no side effects; execution after
`--run` is mediated as in `run`). **Exit codes:** 0, 1, 2, 3, 4, 7, 8 (`approve --run`
adds 5, 6, 9).

**JSON result (`plan show`, `data`):**

```json
{
  "plan_id": "01JZX0PLAN0000000000000000",
  "run_id": "01JZX0RUN00000000000000000",
  "state": "proposed",
  "version": 1,
  "steps": [
    {"ordinal": 1, "summary": "Locate signup handler", "tasks": []}
  ],
  "supersedes": null
}
```

**Examples:**

```text
$ andromeda plan new "migrate config parsing to go-toml/v2"       # valid
$ andromeda plan approve 01JZX0PL --run --json                    # valid (unique prefix)
$ andromeda plan approve                                          # invalid
error[E-CLI-002]: invalid value for <plan-id>: a plan ULID or unique prefix is required
```

**Errors:** E-CLI-002; E-AGT family (planning/approval state conflicts, e.g., approving a
`superseded` plan); E-PROV family exit 7.

## `andromeda exec`

Single mediated execution without an agent loop: one tool invocation or one terminal
command, under exactly the same permission and sandbox mediation as agent-initiated actions
(PRD-004, PRD-005). The deterministic automation primitive (UC-08).

```text
andromeda exec tool <tool-name> [--input <json> | --input-file <path>]
andromeda exec command [--cwd <path>] [--env KEY=VALUE ...] -- <argv...>
```

Behavior: `exec tool` validates input against the tool's schema (ToolPort `Validate`) and
executes one Tool Invocation through its frozen states. `exec command` runs one Command
Execution through the Sandbox Engine and Terminal Engine; the child's stdout/stderr stream
through to the CLI's streams live; `--` is mandatory before argv. The child's own exit code
is reported in `data.child_exit_code` and summarized on stderr — Andromeda's process exit
code stays within the ADR-016 scheme: 0 when the child succeeded, 6 when it failed
(passthrough of arbitrary child codes would collide with the closed scheme; recorded as a
deliberate decision here).

**Permissions:** exactly what the tool declares (`exec tool`); `process_spawn` and
`execute` plus whatever the command's sandbox profile implies (`exec command`).
**Exit codes:** 0, 1, 2, 3, 5, 6, 8.

**JSON result (`exec tool`, `data`):**

```json
{
  "invocation_id": "01JZX0INV00000000000000000",
  "tool": "fs.search",
  "state": "succeeded",
  "result": {"status": "success", "output": {}},
  "child_exit_code": null,
  "duration_ms": 84
}
```

`state` uses the frozen Tool Invocation states; `result.status` the frozen Tool Result
vocabulary.

**Examples:**

```text
$ andromeda exec tool fs.search --input '{"pattern":"TODO_MARKER","path":"src/"}'   # valid
$ andromeda exec command --cwd ./svc -- go test ./...                               # valid
$ andromeda exec command go test                                                    # invalid
error[E-CLI-002]: invalid value for exec command: argv must follow "--"
```

**Errors:** E-CLI-002/006/009; E-TOOL family (Volume 6) exit 6; E-SEC family exit 5.

## `andromeda init`

Initializes the current (or named) directory as a workspace: creates `.andromeda/`
(ADR-022), initializes the workspace database (ADR-028), registers the workspace globally,
and schedules the initial lexical index build unless `--no-index`.

```text
andromeda init [<path>] [--profile <name>] [--force] [--no-index]
```

`--force` re-initializes an existing workspace — a destructive confirmation (FR-CLI-010):
it overwrites workspace-scoped settings (session/run history is preserved; the confirmation
text states exactly this). Without `--force`, an already-initialized path fails cleanly.

**Permissions:** `write` (workspace scope). **Exit codes:** 0, 1, 2, 3, 5, 8, 9 (9 on
database initialization/migration integrity failure per ADR-029).

**JSON result (`data`):**

```json
{
  "workspace_root": "/home/dev/project",
  "created": true,
  "profile": "default",
  "index_scheduled": true
}
```

**Examples:**

```text
$ andromeda init                                   # valid, current directory
$ andromeda init ~/work/svc --profile backend      # valid
$ andromeda init --force                           # valid; prompts (or E-CLI-003 with no consent path)
$ andromeda init /nonexistent/path                 # invalid
error[E-CLI-002]: invalid value for <path>: directory does not exist
```

**Errors:** E-CLI-002/003; E-AGT family (workspace open failures, Volume 4); E-CFG family
exit 3.

## `andromeda session`

Session management for automation parity with the TUI's session handling (PRD-009,
PRD-010; UC-11). Recorded grammar addition — see chapter 01.

```text
andromeda session list [--state <session-state>] [--limit <n>]
andromeda session show <session-id>
andromeda session resume <session-id>
andromeda session end <session-id>
```

Behavior: `list` renders sessions with frozen Session states; default `--limit 20`, most
recent first. `show` includes runs (frozen Run states), accounting summary, and
resumability. `resume` on a `suspended` session: interactive → TUI hand-off into the
session (FR-CLI-003 conditions apply); non-interactive → resumes `interrupted` runs to
their next terminal state, streaming per FR-CLI-006, honoring the never-silently-re-execute
invariant (side-effecting tasks require the Volume 4 re-approval rules). `end` moves a
session to `ended` (never resumable, always readable) — a destructive confirmation.

**Permissions:** none at the CLI layer (resumed work is mediated as in `run`).
**Exit codes:** `list`/`show` 0, 1, 2, 3, 9; `resume` 0, 1, 2, 3, 4, 5, 6, 7, 8, 9;
`end` 0, 1, 2, 3, 8.

**JSON result (`session list`, `data`):**

```json
{
  "sessions": [
    {
      "session_id": "01JZX0A1B2C3D4E5F6G7H8J9K1",
      "state": "suspended",
      "title": "signup validation work",
      "runs": 3,
      "last_activity": "2026-07-11T14:02:11Z"
    }
  ]
}
```

**Examples:**

```text
$ andromeda session list --state suspended          # valid
$ andromeda session resume 01JZX0A1 --json --no-input  # valid, CI resume
$ andromeda session end 01JZX0A1                    # valid; confirms (default No)
$ andromeda session resume                           # invalid
error[E-CLI-002]: invalid value for <session-id>: a session ULID or unique prefix is required
```

**Errors:** E-CLI-002/003; E-AGT family (resume conflicts); E-CFG family (storage) exit 3
or 9 per its envelope.

## `andromeda config`

Configuration surface over ConfigPort; schema, precedence, and validation are Volume 10's
(FR-CFG-001) — this command renders them.

```text
andromeda config get <key>
andromeda config set <key> <value> [--scope global|workspace|project]
andromeda config unset <key> [--scope global|workspace|project]
andromeda config list [--sources]
andromeda config validate [<path>]
andromeda config path
andromeda config edit [--scope global|workspace|project]
```

Behavior: `get` prints the resolved value; with `--json`, value plus source attribution.
`set`/`unset` validate before writing (ConfigPort `Validate`) — an invalid key or value
never reaches disk; default `--scope` is `workspace` inside a workspace, else `global`.
`list --sources` shows every resolved key with its supplying layer. `validate` checks a
named file (or the active configuration) and reports **all** findings. `path` prints
resolved configuration file locations (ADR-022). `edit` opens the scoped file in
`cli.editor`/`VISUAL`/`EDITOR` (TTY required — E-CLI-005 otherwise) and validates on save,
refusing to keep an edit that fails validation without explicit retry/discard choice.

**Permissions:** `write` (configuration files) for `set`/`unset`/`edit`; none for reads.
**Exit codes:** 0, 1, 2, 3, 5 (`edit` adds none beyond these; validation failures are 3).

**JSON result (`config get`, `data`):**

```json
{
  "key": "cli.color",
  "value": "auto",
  "source": "default",
  "scope": null
}
```

**Examples:**

```text
$ andromeda config set cli.color never              # valid
$ andromeda config get providers.default --json     # valid
$ andromeda config set cli.color sometimes          # invalid
error[E-CFG-...]: ...                               # E-CFG family, exit 3 (Volume 10 text)
```

**Errors:** E-CLI-002/005; E-CFG family (Volume 10) exit 3.

## `andromeda auth`

Authentication flows over AuthPort — official mechanisms only (FR-AUTH-001, Volume 5);
storage through the Secret Store (FR-SEC-102, Volume 9). The CLI never prints secret
material.

```text
andromeda auth login <provider> [--profile <name>] [--api-key-stdin]
andromeda auth logout <provider> [--profile <name>]
andromeda auth status [<provider>]
andromeda auth list
andromeda auth rotate <provider> [--profile <name>]
```

Behavior: `login` runs the provider adapter's declared auth method (Volume 5): API-key
intake (interactive hidden prompt, or `--api-key-stdin` for automation), or
device/browser flows where the provider documents them — interactive-only; non-interactive
invocations without `--api-key-stdin` fail with E-CLI-005 naming that flag. `--profile`
names multiple credential bindings per provider. `logout` revokes (provider-side where
officially supported) and deletes local material. `status` renders Authentication Session
states (frozen names) without material. `rotate` drives AuthPort `Rotate`; the Credential
status vocabulary records the outcome.

**Permissions:** `credential_access`; `network` for flows that reach the provider.
**Exit codes:** 0, 1, 2, 3, 4, 5, 8.

**JSON result (`auth status`, `data`):**

```json
{
  "authentications": [
    {
      "provider": "anthropic",
      "profile": "default",
      "state": "active",
      "expires_at": null
    }
  ]
}
```

**Examples:**

```text
$ andromeda auth login anthropic                          # valid, prompts (hidden input)
$ printf %s "$API_KEY" | andromeda auth login anthropic --api-key-stdin   # valid, CI
$ andromeda auth login anthropic --no-input               # invalid without stdin key
error[E-CLI-005]: this operation needs an interactive terminal and none is available
  hint:   pass the key via --api-key-stdin
```

**Errors:** E-CLI-002/005/007; E-AUTH family (Volume 5) exit 4; E-SEC family (keychain
denial) exit 5.

## Requirements

### FR-CLI-013 — Core command family behavior

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: CLI (Volume 8)
- Affected components: CLI; Runtime; Authentication Layer; Configuration Manager; Workspace Engine
- Dependencies: FR-CLI-001, FR-CLI-002, FR-CLI-005–FR-CLI-012; FR-AGT-001 (agent loop, Volume 4); FR-AUTH-001 (Volume 5); FR-CFG-001 (Volume 10)
- Related risks: RISK-CLI-001, RISK-CLI-003

#### Description

The commands specified in this chapter — root, `run`, `plan`, `exec`, `init`, `session`,
`config`, `auth` — MUST behave exactly as their sections define: the stated syntax, flags,
defaults, behavior rules, permissions, exit-code sets, and JSON `data` payloads are
normative. A command producing an exit code outside its declared set, output outside its
declared schema, or a side effect outside its declared permission classes is a defect.

#### Motivation

The core family is the MVP's contract surface (MVP item 1): UC-01, UC-02, UC-07, UC-09,
and UC-11 execute through exactly these commands, so their precision is the difference
between a specified product and a suggestive one.

#### Actors

Users; scripts and CI; the TUI (parity counterpart); Volume 13 test suites.

#### Preconditions

Per command as stated in its section; the family shares chapter 02's conventions.

#### Main flow

1. A caller invokes a family command per its syntax.
2. The command executes through the Runtime API/ports per its behavior rules.
3. Output, exit code, and records match the section's declarations.

#### Alternative flows

- Streaming variants (`run`, `plan approve --run`, `session resume`, `exec command`)
  follow FR-CLI-006 rule 3.

#### Edge cases

- All ULID arguments accept unique prefixes ≥ 6 characters (FR-CLI-001 rule 4).
- Goal-source exclusivity (`run`), argv separator requirement (`exec command`), and
  re-initialization protection (`init`) are each individually tested edge rules.

#### Inputs

Per command: goals, identifiers, configuration keys/values, credentials via stdin, argv
passthrough.

#### Outputs

Per command: streamed activity, result envelopes with the declared `data` schemas, records
persisted by the Runtime.

#### States

Commands render frozen state vocabularies verbatim (Session, Run, Plan, Task, Tool
Invocation, Authentication Session); the CLI owns no machine.

#### Errors

E-CLI family (chapter 02) plus each section's declared foreign families with their mapped
exit codes.

#### Constraints

No family command may prompt outside the confirmation/approval presenters; `auth` never
renders secret material in any mode; `exec` reports child exit codes in data, never as the
process exit code.

#### Security

`run`/`exec`/`session resume` execute arbitrary mediated side effects — their sections bind
them to per-action permission evaluation; `auth` touches `credential_access` exclusively
through AuthPort/Secret Store; `config set` cannot write an invalid configuration.

#### Observability

Family commands emit the chapter 02 lifecycle events; runs/invocations they start are fully
correlated per SM-13.

#### Performance

`run` streaming overhead under SM-08; cold-start commands under SM-06a (Volume 12 owns
both).

#### Compatibility

Identical behavior on Tier 1 platforms; goal text and config values are UTF-8.

#### Acceptance criteria

- Given the section syntax lines as a corpus, when grammar golden tests parse each, then
  every line binds to its documented handler (syntax conformance).
- Given `run --json` completing a two-tool goal against a local provider fixture, when the
  final envelope is validated, then it matches the `run` schema and `state` is a frozen Run
  state.
- Given `exec command -- false`, when it completes, then process exit is 6 and
  `data.child_exit_code` is 1 (child-exit rule).
- Given `auth login` with `--json --no-input` and no `--api-key-stdin`, when invoked, then
  E-CLI-005 renders in the envelope and exit is 1; no credential row is touched (negative +
  permission case).
- Given `init --force` declined at the prompt, when inspected, then the workspace is
  unmodified (confirmation case).
- Observability: each of the above emits `cli.command.started` and a terminal lifecycle
  event with matching correlation IDs.

#### Verification method

Volume 13 CLI suite: per-command golden and schema-conformance tests, PTY prompt tests,
non-interactive matrix, crash/interrupt tests for `run`/`session resume` (SM-11 alignment);
UC-01/UC-07/UC-11 E2E journeys.

#### Traceability

PRD-001, PRD-005, PRD-006, PRD-008, PRD-009, PRD-010; UC-01, UC-02, UC-07, UC-09, UC-11;
MVP items 1, 19, 23 context; FR-AGT-001, FR-AUTH-001, FR-CFG-001.
