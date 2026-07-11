# Annex — Consolidated CLI Command Catalog

**Status:** Consolidated (Phase C). This annex is the corpus-wide index of the CLI command
tree specified in Volume 8, chapters
[03](../volume-08-cli-and-tui/03-cli-commands-core.md) (core),
[04](../volume-08-cli-and-tui/04-cli-commands-platform.md) (platform),
[05](../volume-08-cli-and-tui/05-cli-commands-data.md) (data and inspection), and
[06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) (maintenance). It is a
*reference view*: syntax, defaults, behavior rules, JSON `data` schemas, examples, and error
mappings live in the linked chapters, which are normative (bound by FR-CLI-013 through
FR-CLI-016). Grammar: keystone FR-CLI-001 and ADR-100 (resource-noun groups, a closed shared
verb vocabulary, depth ≤ 3, singular nouns, closed top level). Extension-contributed
commands mount under the reserved `x` namespace (ADR-104) and are not part of this closed
tree.

## Conventions that apply to every command

- **`--json` everywhere.** Every command offers structured output per FR-CLI-006 and
  ADR-101: a fixed eight-field result envelope (`schema`, `command`, `ok`, `exit_code`,
  `data`, `error`, `warnings`, `meta`); streaming commands emit NDJSON stream documents
  wrapping Volume 10 event envelopes before their terminal result envelope. The "JSON"
  column below distinguishes the result-envelope shape (`envelope`) from streaming
  (`NDJSON + envelope`) and export-document output.
- **Global flags** (FR-CLI-005, accepted by every command): `--help`/`-h`, `--json`,
  `--quiet`/`-q`, `--verbose`/`-v`, `--debug`, `--no-input`, `--yes`/`-y`,
  `--workspace`/`-C <path>`, `--profile <name>`, `--config <path>`,
  `--color auto|always|never`, `--timeout <duration>`; root only: `--version`.
- **Exit codes** follow the closed ADR-016 scheme (0–9); multi-item commands aggregate by
  the fixed severity order `9 > 5 > 4 > 3 > 7 > 6 > 8 > 2 > 1` (Volume 8 chapter 02).
- **Permissions** name Volume 9 enum values ([catalog](catalog-permissions.md)); "none (CLI
  layer)" means the command itself needs no grant while the work it starts is mediated
  per-action through PermissionPort.
- **Destructive confirmations** (FR-CLI-010) are marked "confirm"; `--yes` covers them and
  never covers permission approvals (ADR-102).

## Command tree

```text
andromeda
├── run · plan · exec · init · session · config · auth          (core, MVP)
├── provider · model · tool                                     (platform, MVP)
├── plugin · skill · workflow · mcp                             (platform, Beta; grammar reserved from Core)
├── memory · context · index · git · logs · trace · export      (data, MVP)
├── doctor · update · version · completion                      (maintenance, MVP)
└── x <extension> <command>                                     (extension mount, Beta; ADR-104)
```

## Core commands (Volume 8, chapter 03 — all MVP)

| Command | Synopsis | Key flags | Exit codes | JSON | Permissions | Defined in |
|---|---|---|---|---|---|---|
| `andromeda` (root) | Bare interactive invocation hands off to the TUI; bare non-interactive prints usage and exits 2 | global only; `--version` alias | 2; after TUI hand-off: 0, 1, 8, 9 | none (E-CLI-002 for `--json` bare) | none (CLI layer) | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `run` | `run [<goal>...]`, `run --file <path>`, or goal from piped stdin — execute one agent run, streaming activity | `--file/-f`, `--agent`, `--provider`, `--model`, `--session`, `--max-turns`, `--budget-tokens`, `--budget-cost` | 0–9 (full set) | NDJSON + envelope | none (CLI layer); every tool side effect mediated individually | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `plan new` | Produce a Plan without executing (state `proposed`) | `--agent`, `--provider`, `--model`, `--session`, `--file` | 0, 1, 2, 3, 4, 7, 8 | envelope | none (CLI layer) | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `plan list` | List plans | `--session`, `--state` | 0, 1, 2, 3, 4, 7, 8 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `plan show` | Render steps, task derivations, revision lineage | — | 0, 1, 2, 3, 4, 7, 8 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `plan approve` | Approve a plan; `--run` starts execution (two-step `run`) | `--run` | 0, 1, 2, 3, 4, 7, 8; with `--run` adds 5, 6, 9 | envelope (NDJSON with `--run`) | none; execution after `--run` mediated as in `run` | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `plan reject` | Reject a proposed plan | `--reason` | 0, 1, 2, 3, 4, 7, 8 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `exec tool` | One mediated Tool Invocation without an agent loop | `--input`, `--input-file` | 0, 1, 2, 3, 5, 6, 8 | envelope | exactly what the tool declares | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `exec command` | One sandboxed Command Execution; `--` before argv is mandatory; child exit code reported in `data`, never as process exit | `--cwd`, `--env` | 0, 1, 2, 3, 5, 6, 8 | envelope | `process_spawn`, `execute`, plus sandbox-profile implications | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `init` | Initialize a workspace: `.andromeda/`, workspace DB, global registration, initial lexical index | `--profile`, `--force` (confirm), `--no-index` | 0, 1, 2, 3, 5, 8, 9 | envelope | `write` (workspace scope) | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `session list` | List sessions with frozen Session states | `--state`, `--limit` (default 20) | 0, 1, 2, 3, 9 | envelope | none (CLI layer) | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `session show` | Session detail: runs, accounting, resumability | — | 0, 1, 2, 3, 9 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `session resume` | Resume a `suspended` session (interactive → TUI; non-interactive resumes `interrupted` runs, never silently re-executing) | — | 0–9 (full set) | NDJSON + envelope | none; resumed work mediated as in `run` | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `session end` | Move a session to `ended` (confirm; never resumable, always readable) | — | 0, 1, 2, 3, 8 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `config get` | Print the resolved value (with `--json`, source attribution) | — | 0, 1, 2, 3, 5 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `config set` / `config unset` | Validate-then-write a key at a scope; invalid values never reach disk | `--scope global\|workspace\|project` | 0, 1, 2, 3, 5 | envelope | `write` (configuration files) | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `config list` | Every resolved key; `--sources` shows the supplying layer | `--sources` | 0, 1, 2, 3, 5 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `config validate` | Check a file or the active configuration; reports all findings | `[<path>]` | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `config path` | Print resolved configuration file locations (ADR-022) | — | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `config edit` | Open the scoped file in the editor; validates on save (TTY required, E-CLI-005 otherwise) | `--scope` | 0, 1, 2, 3, 5 | envelope | `write` (configuration files) | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `auth login` | Run the adapter's declared auth method; hidden prompt or `--api-key-stdin`; never prints secret material | `--profile`, `--api-key-stdin` | 0, 1, 2, 3, 4, 5, 8 | envelope | `credential_access`; `network` for provider-reaching flows | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `auth logout` | Revoke (provider-side where officially supported) and delete local material | `--profile` | 0, 1, 2, 3, 4, 5, 8 | envelope | `credential_access`; `network` for provider-side revocation | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `auth status` | Authentication Session states (frozen names), no material | `[<provider>]` | 0, 1, 2, 3, 4, 5, 8 | envelope | `credential_access` | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `auth list` | List credential bindings per provider/profile | — | 0, 1, 2, 3, 4, 5, 8 | envelope | `credential_access` | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |
| `auth rotate` | Drive AuthPort `Rotate`; outcome in the Credential status vocabulary | `--profile` | 0, 1, 2, 3, 4, 5, 8 | envelope | `credential_access`; `network` | [Vol 8 ch 03](../volume-08-cli-and-tui/03-cli-commands-core.md) |

## Platform commands (Volume 8, chapter 04 — `provider`/`model`/`tool` MVP; `plugin`/`skill`/`workflow`/`mcp` Beta, grammar reserved from Core)

| Command | Synopsis | Key flags | Exit codes | JSON | Permissions | Defined in |
|---|---|---|---|---|---|---|
| `provider list` / `provider show` | Provider registry with frozen connection states | — | 0, 1, 2, 3, 5 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `provider add` | Register a Provider against an installed adapter; declaration validated before any row; never takes credentials | `--adapter`, `--endpoint`, `--set key=value` | 0, 1, 2, 3, 5 | envelope | `write` (configuration) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `provider remove` | Deregister (confirm; row tombstoned `removed`; refuses with dependents named) | — | 0, 1, 2, 3, 5 | envelope | `write` (configuration) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `provider test` | Drive verification (`verifying` → `available`/`degraded`/`unavailable`); reports discovered capabilities | — | 0, 1, 2, 3, 4, 5, 7, 8 | envelope | `network` | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `provider enable` / `provider disable` | Toggle administrative exclusion (`disabled`) | — | 0, 1, 2, 3, 5 | envelope | `write` (configuration) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `model list` | Local Model catalog; `--refresh` re-runs discovery | `--provider`, `--refresh` | 0, 1, 2, 3; `--refresh` adds 4, 7, 8 | envelope | `network` only for `--refresh` | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `model show` / `model capabilities` | Model detail; declared CapabilitySet with closed capability names — never inferred or simulated | `--provider` | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `model default` | Set the profile-level default model (validated against the catalog) | `--provider` | 0, 1, 2, 3, 5 | envelope | `write` (configuration) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `tool list` / `tool show` | Tool registry with origin and trust level always visible; `show` renders the full declaration | `--origin builtin\|plugin\|mcp`, `--enabled`/`--disabled` | 0, 1, 2, 3, 5, 6 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `tool enable` / `tool disable` | Toggle agent visibility workspace-wide | — | 0, 1, 2, 3, 5, 6 | envelope | `write` (configuration) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `tool test` | Schema and semantic validation of `--input` without executing (execution is `exec tool`) | `--input` | 0 with findings in data; 6 only for tool-runtime failures | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `plugin list` / `plugin show` | Plugin lifecycle view: process state, declared surfaces, trust level | — | 0, 1, 2, 3, 5, 6, 8 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `plugin install` | Install from path/archive/registry; streams frozen installation states; verification precedes any activation; idempotent on same version | `--version` | 0, 1, 2, 3, 5, 6, 8 (never 9) | NDJSON + envelope | `package_installation`; `network` for remote sources | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `plugin uninstall` | Stop through the machine and tombstone the Package row (confirm) | — | 0, 1, 2, 3, 5, 6, 8 | envelope | `package_installation` | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `plugin enable` / `plugin disable` | Toggle the plugin | — | 0, 1, 2, 3, 5, 6, 8 | envelope | `package_installation` | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `skill list` / `skill show` | Skill registry; `show` renders the manifest surface (requirements, composition, trust) | — | 0, 1, 2, 3, 5, 6, 8 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `skill install` / `skill uninstall` | Same shape as `plugin`; manifest validated against the Volume 6 format before staging; no process state | `--version` | 0, 1, 2, 3, 5, 6, 8 | NDJSON + envelope (install) | `package_installation`; `network` for remote sources | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `skill enable` / `skill disable` | Toggle registration enablement | — | 0, 1, 2, 3, 5, 6, 8 | envelope | `package_installation` | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `workflow list` / `workflow show` | Definitions and Workflow Runs (`--runs`) | `--runs` | 0–9 (family set) | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `workflow run` | Instantiate a Workflow Run, streaming stage transitions; gates block interactively or resolve from policy | `--input key=value`, `--session` | 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 | NDJSON + envelope | none (CLI layer); each step's actions mediated individually | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `workflow status` | Run state, stage, progress | — | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `workflow resume` / `workflow cancel` | Continue at the last persisted step boundary / request cancellation | — | 0–9 (family set) | NDJSON + envelope | none (CLI layer) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `workflow validate` | Check a definition file; reports all findings | `<path>` | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `mcp list` / `mcp show` | Server registrations with frozen MCP Client Connection states | — | 0, 1, 2, 3, 5 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `mcp add` | Register a server with exactly one transport (`--command` xor `--url`, else E-CLI-006); registration alone connects nothing | `--command`, `--url`, `--env` | 0, 1, 2, 3, 5 | envelope | `write` (configuration); connection-time permissions per Volume 6/9 policy | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `mcp remove` | Tombstone the registration (confirm) | — | 0, 1, 2, 3, 5 | envelope | `write` (configuration) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `mcp enable` / `mcp disable` | Toggle the registration | — | 0, 1, 2, 3, 5 | envelope | `write` (configuration) | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |
| `mcp status` / `mcp tools` | Connection states; tools discovered from a `ready` connection, marked origin `mcp` with server trust level | `[<name>]` | 0, 1, 2, 3, 5, 6, 8 | envelope | none | [Vol 8 ch 04](../volume-08-cli-and-tui/04-cli-commands-platform.md) |

## Data and inspection commands (Volume 8, chapter 05 — all MVP; offline except declared network paths)

| Command | Synopsis | Key flags | Exit codes | JSON | Permissions | Defined in |
|---|---|---|---|---|---|---|
| `memory search` | Lexical (and semantic where indexed) query with provenance and trust attribution | `--layer`, `--limit` | 0, 1, 2, 3, 5, 8, 9 | envelope | none | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `memory show` | Record detail with frozen Memory Record status | — | 0, 1, 2, 3, 5, 8, 9 | envelope | none | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `memory add` | Ingest a user-authored record (provenance `user`) | `--layer` | 0, 1, 2, 3, 5, 8, 9 | envelope | `write` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `memory delete` | Hard deletion honoring cascade rules; deletion records persist (confirm; idempotent on `deleted`) | — | 0, 1, 2, 3, 5, 8, 9 | envelope | `write` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `memory export` | Canonical JSON entity documents; NDJSON to stdout or file | `--layer`, `--output` | 0, 1, 2, 3, 5, 8, 9 | export documents | `write` for `--output` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `context show` | Current assembly: items with source, priority, token counts, budget utilization | `--session` | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `context pin` / `context unpin` | Force or release inclusion of a path/item | `--session` | 0, 1, 2, 3 | envelope | `write` (context preferences) | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `context exclude` | Block paths/globs from assembly (session-scoped; `--workspace` widens) | `--workspace` | 0, 1, 2, 3 | envelope | `write` (context preferences) | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `index build` | Full build as a supervised operation; `--semantic` requires a declared `embeddings` provider — no silent lexical fallback | `--semantic`, `--path` | 0, 1, 2, 3, 7, 8 | NDJSON + envelope | `read`; `network` only for remote embeddings | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `index update` | Apply incremental changes | `--path` | 0, 1, 2, 3, 7, 8 | envelope | `read` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `index status` | Per-index state, generation, staleness (frozen Index states) | — | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `index search` | Query the index | `--semantic`, `--limit` | 0, 1, 2, 3, 7, 8 | envelope | none; `network` for remote semantic queries | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `index invalidate` | Mark scopes stale | `--path` | 0, 1, 2, 3 | envelope | `read` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `index remove` | Drop the index cache (confirm; rebuildable — never data loss) | — | 0, 1, 2, 3 | envelope | `read` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `git status` / `git diff` / `git log` / `git branch` (no name) | Read queries rendered from GitPort; `diff` streams and pages | `--staged`, `--limit`, `[<rev-spec>]` | 0, 1, 2, 3, 8 | envelope | none | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `git stage` / `git unstage` | Stage or unstage paths (attributed File Change records) | `<path>...` | 0, 1, 2, 3, 5, 8 | envelope | `git_mutation` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `git commit` | Create a commit; message content rules per ADR-015 enforced by Volume 11 | `--message` | 0, 1, 2, 3, 5, 8 | envelope | `git_mutation` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `git branch <name>` / `git switch` | Create or switch branches | `--create` | 0, 1, 2, 3, 5, 8 | envelope | `git_mutation` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `git apply` | Apply a reviewed Patch (`proposed` → `applied`) atomically or not at all | `<patch-id-or-path>` | 0, 1, 2, 3, 5, 8 | envelope | `git_mutation` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `logs` | Structured log inspection; `--follow` tails until interrupt (clean interrupt exits 0); records pre-redacted | `--follow`, `--level`, `--since`, `--session`, `--run`, `--limit` | 0, 1, 2, 3 | NDJSON of log records | none | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `trace list` / `trace show` | Correlated span tree of a run with token/cost rollup (UC-13 audit entry point) | `--session`, `--limit` | 0, 1, 2, 3 | envelope | none | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |
| `export session` / `export run` / `export memory` / `export audit` | Canonical Volume 2 export documents through the full redaction pipeline; stdout or `--output` (confirm on overwrite) | `--output`, `--layer`, `--since` | 0, 1, 2, 3, 5, 8, 9 | export documents (`andromeda.export.<entity>.v1`) | none for stdout; `write` for `--output` | [Vol 8 ch 05](../volume-08-cli-and-tui/05-cli-commands-data.md) |

## Maintenance commands (Volume 8, chapter 06 — all MVP)

| Command | Synopsis | Key flags | Exit codes | JSON | Permissions | Defined in |
|---|---|---|---|---|---|---|
| `doctor` | Closed, identified check set (`binary`, `paths`, `config`, `storage`, `git`, `secretstore`, `terminal`, `workspace`, `disk`, plus `providers`/`update` with `--network`); diagnosis-only, continues past failures, aggregated exit | `--network`, `--check <id>` | 0, 1, 2, 3, 4, 5, 7, 8, 9 (severity aggregation) | envelope (`data.checks`) | none by default; `network` with `--network` | [Vol 8 ch 06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) |
| `update` | Full flow check → consent (confirm) → download → verify → apply over the frozen Update states; unattended form `--yes --no-input` | `--channel` | 0, 1, 2, 3, 5, 8, 9 | NDJSON + envelope | `network` (check/download); `system_modification` (apply) | [Vol 8 ch 06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) |
| `update check` | Check only; `up_to_date` and `update_available` both exit 0; the only subcommand requiring the network | `--channel` | 0, 1, 2, 3, 8 | envelope | `network` | [Vol 8 ch 06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) |
| `update rollback` | Restore the retained prior version from local artifacts, fully offline (confirm naming both versions) | — | 0, 1, 2, 3, 5, 8, 9 | envelope | `system_modification` | [Vol 8 ch 06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) |
| `version` | Build identity; no configuration load, no workspace discovery, no network; works in a broken environment; `--version` is the byte-identical alias | — | 0, 1 (output failure), 2 | envelope | none | [Vol 8 ch 06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) |
| `completion bash` / `completion zsh` / `completion fish` | Shell completion script to stdout; behavior per FR-CLI-012 | — | 0, 1, 2 | envelope (`data.script`, `data.shell`) | none | [Vol 8 ch 06](../volume-08-cli-and-tui/06-cli-commands-maintenance.md) |

The post-command **update notice** (chapter 06) is not a command: one throttled stderr line
from the locally cached check result, never performing network access, suppressed under
`--quiet`, `--json`, CI mode, non-TTY stderr, and inside
`update`/`version`/`completion`/`doctor`.

## Consolidation notes

- **Coverage.** All 25 top-level command groups of chapters 03–06 appear above with every
  subcommand the chapters specify (grouped rows share their chapter's common exit-code and
  permission declarations; per-subcommand deltas noted in place). The `x` extension mount
  (ADR-104) is grammar-reserved and carries no fixed subcommands to catalog.
- **`session` group.** Chapter 03 marks `andromeda session` as a recorded grammar addition
  relative to the chapter map's family list (root, `run`, `plan`, `exec`, `init`, `config`,
  `auth`); this catalog follows the chapter.
- **Exit-code sets** are reproduced from each command's section. Where a chapter states a
  group set plus subcommand additions (e.g., `provider test` adds 4, 7, 8), the row shows
  the effective per-command set.
- **Permissions** rows reproduce the chapters' "Permissions:" declarations; parenthetical
  scope glosses ("workspace scope", "configuration files") are the chapters' wording, not
  new qualifiers.
- **Phases.** Every cataloged command is MVP except the `plugin`, `skill`, `workflow`, and
  `mcp` groups (Beta — grammar reserved from Core so later arrival is additive, FR-CLI-014).
