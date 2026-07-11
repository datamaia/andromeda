# Annex — Consolidated Permission Catalog

**Status:** Consolidated (Phase C). This annex is the corpus-wide index of the permission
model's three closed vocabularies — 13 permissions, 10 scope qualifiers, 7 decisions — and
of the permission bindings of the 20 built-in tools. It is a *reference view*: the
normative home of the enums, the evaluation algorithm, inheritance, revocation,
persistence, and audit obligations is Volume 9 chapter
[05](../volume-09-security/05-permission-model.md) (keystone FR-SEC-100, implemented by the
Permission Manager behind `PermissionPort`); the tool bindings are Volume 6 chapter
[03](../volume-06-tools-mcp-skills-plugins/03-builtin-tools-catalog.md)'s (FR-TOOL-007).
This annex mints nothing and renames nothing. All three enums are closed vocabularies:
adding a value requires an ADR through the Volume 0 change procedure.

## The decision path, in one paragraph

Every side-effecting action enters the Permission Manager through `PermissionPort`
(`Check`, `Request`, `RecordDecision` — Volume 3's frozen port) and resolves in a fixed
four-tier precedence ([ADR-121](adr/ADR-121.md)): **any matching deny → deny; else any
ask-forcing rule → ask; else any matching allow → allow; else ask.** The three-value
*evaluation* vocabulary (`allow` | `deny` | `ask`) is distinct from the seven-value
*decision* vocabulary below. Where interaction is not permitted (non-interactive CLI,
headless mode per ADR-032, CI, every `Check` call), `ask` resolves as deny with E-SEC-001
(FR-SEC-105) — no configuration key, environment variable, or flag converts `ask` to
`allow`. Evaluation failure is E-SEC-002 and resolves as deny (fail-closed,
[ADR-125](adr/ADR-125.md)). Every resolution produces exactly one Audit Record; built-in
defaults grant only `read` inside the opened workspace — everything else asks (FR-SEC-100).

## Permission enum (13 values)

Canonical `snake_case` names, frozen for the corpus. Extension declarations (tools,
plugins, MCP servers, skills — Volume 6) declare required permissions with exactly these
names; unknown names are validation errors, never forward-compatible values.

| Permission | Grants the class of action | Representative actions | Defined in |
|---|---|---|---|
| `read` | Read files and metadata within qualified paths | file read, directory listing, search, diff computation | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `write` | Create, modify, delete, or rename files within qualified paths | file write, patch apply, temp artifact creation outside sandbox temp | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `execute` | Run commands and executables matching qualified command patterns | terminal command, test runner, build invocation | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `network` | Open network connections to qualified hosts/domains | HTTP requests by tools, non-provider downloads | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `credential_access` | Resolve secret material through the Secret Store | provider key resolution, integration-tool token resolution | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `git_mutation` | Mutate repository state | stage, commit, branch create/switch, apply patch, worktree ops | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `process_spawn` | Spawn processes beyond a tool's own sandboxed execution | background workers, watchers, plugin child processes | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `container_access` | Interact with container or orchestration runtimes | docker/kubernetes tool operations (Volume 6 catalog) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `external_service_access` | Call third-party service APIs as an authenticated principal | GitHub/GitLab/Jira/Notion/Slack/Linear tool operations | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `clipboard` | Read or write the system clipboard | copy diff/result to clipboard | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `notifications` | Emit desktop/system notifications | completion or approval-needed notices | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `package_installation` | Install, update, or remove packages/extensions | plugin/skill/MCP server install (Volume 6), self-update apply (Volume 14) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `system_modification` | Change machine state outside the workspace | global config writes, shell profile edits, OS settings | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |

Boundary rules (Vol 9 ch 05): provider inference traffic through `ProviderPort` does not
consume `network` (it is authenticated, costed, and audited through Volumes 5 and 10 —
`network` governs *tool-originated* connections); `credential_access` gates any resolution
of secret material on behalf of a tool, plugin, or MCP server; and `system_modification` is
never grantable to extensions at `always_allow_policy`.

## Scope qualifiers (10 values)

Frozen names: `session`, `workspace`, `command`, `tool`, `provider`, `host`, `path`,
`domain`, `repository`, `organization`. Two groups:

- **Attachment qualifiers** — `session`, `workspace` — bind a grant's validity context,
  materializing as the Permission row's grant scope (Volume 2: `invocation` ⊂ `run` ⊂
  `session` ⊂ `workspace` ⊂ `global`, inherited evaluation-time only).
- **Resource qualifiers** — the remaining eight — constrain covered resources,
  materializing as the row's `resource_selector` (a JSON object whose keys are exactly
  these names). A present key must match; an absent key is unconstrained; a present key
  with an empty list matches nothing; unknown keys invalidate the grant or rule (E-SEC-002
  class, never silently ignored).

| Qualifier | Kind | Pattern form | Matching rule | Defined in |
|---|---|---|---|---|
| `session` | attachment | — | Grant dies with its Session (expiry recorded as a Session-machine side effect) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `workspace` | attachment | — | Grant persists in that workspace's database until revocation or TTL | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `command` | resource | space-separated word patterns | First word matches the executable basename or absolute path; per-word `*` globbing; a final bare `*` matches all remaining arguments; comparison on parsed argv, never raw shell text | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `tool` | resource | canonical dotted tool name (Volume 6 grammar) | Exact name or `namespace.*` | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `provider` | resource | provider label (Volume 2) | Exact match | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `host` | resource | hostname, IP literal, or CIDR | Exact hostname; IP within CIDR; no wildcard hostnames (use `domain`) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `path` | resource | glob, workspace-relative or absolute | `*` within one segment, `**` across segments; queried path symlink-resolved before matching | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `domain` | resource | DNS name, optional leading `*.` | Exact name, or suffix match where `*.example.com` covers subdomains but not the apex | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `repository` | resource | repository slug or absolute path | Exact match on the repository identity the Git Engine reports | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `organization` | resource | hosting namespace/organization name | Exact match; covers all repositories the platform attributes to it | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |

## Decision vocabulary (7 values)

Decisions are what a user (or policy) answers on an Approval; each maps deterministically
to persisted effects. Approval states are the frozen Volume 2 vocabulary (`requested`,
`granted`, `denied`, `expired`, `cancelled`; full machine in
[Vol 9 ch 09](../volume-09-security/09-approval-state-machine.md)).

| Decision | Persisted effect | Grant scope | Lifetime | Defined in |
|---|---|---|---|---|
| `allow_once` | One allow Permission row, invocation-scoped, referenced by the subject's `permission_ids` | `invocation` | Consumed by that invocation; a retry re-evaluates (Volume 4 gated-retry rule) | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `allow_for_session` | Allow Permission row attached to the Session | `session` | Until session end or revocation | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `allow_for_workspace` | Allow Permission row attached to the Workspace | `workspace` | Until revocation, or `permissions.workspace_grant_ttl` when non-zero | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `always_allow_policy` | Standing allow row at `global` scope (workspace scope when the prompt's selector is workspace-relative), origin `approval` | `global` or `workspace` | Until revocation; evaluated as standing policy thereafter | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `deny_once` | One deny Permission row, invocation-scoped | `invocation` | That invocation only | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `always_deny` | Standing deny row at `workspace` or `global` scope (same placement rule as `always_allow_policy`) | `workspace` or `global` | Until revocation | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |
| `ask_every_time` | No new grant; matching standing allow grants at `session` and `workspace` scope are revoked (`revoked_by: user`) so future requests prompt again | — | Immediate | [Vol 9 ch 05](../volume-09-security/05-permission-model.md) |

Decision constraints (Vol 9 ch 05): `always_allow_policy`/`always_deny` prompts display the
exact selector to be persisted; `always_allow_policy` is not offered for
`system_modification` requests nor for selectors unbounded on all resource qualifiers when
the permission is `execute`, `write`, or `credential_access` (such standing grants exist
only as explicit config rules); prompt dismissal is not `deny_once` — the Approval stays
pending until it expires, and expiry resolves denied-class.

## Policy rules and configuration

Standing policy lives in the `[permissions]` table (keys cataloged in
[catalog-config.md](catalog-config.md); protected against runtime overrides per FR-CFG-005):
`permissions.approval_timeout` (default `"10m"`), `permissions.workspace_grant_ttl`
(default `"0s"`), and `[[permissions.rules]]` entries with fields `name` (unique per config
layer), `permission` (enum value), `effect` (`allow` | `deny` | `ask`), plus any resource-
qualifier keys using the selector grammar above. Rules evaluate as virtual grants
(`origin_kind: policy`); project/workspace config layers attach at workspace scope, the
global layer at global scope; an `effect: ask` rule pins prompting even over a matching
allow. Invalid rules fail configuration validation entirely (E-CFG class at load per Volume
10); a permission rule is never applied best-effort.

## Built-in tools × permissions matrix (Volume 6, chapter 03)

Bindings as declared per tool by the catalog chapter, which is normative (FR-TOOL-007).
Legend: **R** — required (every operation of the tool needs it); **O** — optional,
operation-dependent (requested only when the arguments demand it, FR-TOOL-005). The
qualifying operation for each O is footnoted below the table. Permissions requested by no
built-in tool are omitted as columns and listed after the table.

| Tool | Phase | `read` | `write` | `execute` | `network` | `credential_access` | `git_mutation` | `process_spawn` | `container_access` | `external_service_access` | `notifications` |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `fs.read` | MVP | R | | | | | | | | | |
| `fs.write` | MVP | | R | | | | | | | | |
| `fs.search` | MVP | R | | | | | | | | | |
| `fs.replace` | MVP | R | R | | | | | | | | |
| `fs.patch` | MVP | | R | | | | | | | | |
| `fs.diff` | MVP | R | | | | | | | | | |
| `git.exec` | MVP | R | | | | | O¹ | | | | |
| `terminal.exec` | MVP | | | R | O² | | | R | | | |
| `process.control` | Beta | | | | | | | R | | | |
| `http.request` | Beta | | | | R³ | O⁴ | | | | | |
| `sqlite.query` | Beta | R | O⁵ | | | | | | | | |
| `docker.control` | Beta | | | | O⁶ | | | | R | | |
| `github.request` | Beta | | | | R | | | | | R | |
| `browser.control` | v1 | | | | R | | | R⁹ | | | |
| `kubernetes.control` | v1 | | | O⁷ | R | | | | R | | |
| `gitlab.request` | v1 | | | | R | | | | | R | |
| `jira.request` | v1 | | | | R | | | | | R | |
| `slack.request` | v1 | | | | R | | | | | R | O⁸ |
| `notion.request` | v2 | | | | R | | | | | R | |
| `linear.request` | v2 | | | | R | | | | | R | |

1. `git.exec`: read operations (`status`, `diff`, `log`, `show`, `branch_list`) require
   `read`; mutating operations (`stage`, `unstage`, `commit`, `branch_create`,
   `branch_switch`, `apply_patch`, worktree operations) additionally request
   `git_mutation`. Destructive Git actions defer to Volume 11 confirmation gates.
2. `terminal.exec`: `network` when the command policy grants egress (Beta+ enforcement per
   the ADR-021 sandbox layers).
3. `http.request`: `network` with `host` selectors.
4. `http.request`: `credential_access` only when a `credential_ref` header source is used —
   secret values never appear in arguments or records.
5. `sqlite.query`: `write` for mutating statements, detected by statement classification
   before execution; Andromeda's own state databases are refused by contract.
6. `docker.control`: local Docker Engine socket by default; `network` for remote engine
   endpoints.
7. `kubernetes.control`: pod `exec` is additionally gated as `execute`.
8. `slack.request`: notification-style posting additionally declares `notifications`.
9. `browser.control` additionally requires `process_spawn` (browser session processes) —
   both of its permissions are required, per its catalog entry.

No built-in tool requests `clipboard`, `package_installation`, or `system_modification`:
those permissions bind at other mediated surfaces — the TUI copy path (`clipboard`, Volume
8), the package operations of `plugin`/`skill`/`mcp` install flows (`package_installation`,
Volumes 6 and 8), and the updater's apply/rollback steps (`system_modification`, Volume
14) — see [catalog-commands.md](catalog-commands.md) for the command-level bindings. No
built-in tool exposes configuration mutation (FR-CFG-005 constraint).

## Consolidation notes

- **Coverage.** 13 permissions, 10 scope qualifiers, and 7 decisions reproduce the frozen
  enums of Volume 9 chapter 05 exactly; the 20 tool rows reproduce the permission and
  phase declarations of Volume 6 chapter 03 exactly (phase summary: 8 MVP, 5 Beta, 5 v1,
  2 v2). No register conflicts with a defining chapter.
- **Matrix values are declarations, not grants.** A tool's R/O cells state what its
  declaration requests; whether an invocation proceeds is decided per action through the
  evaluation algorithm — built-in defaults allow only `read` inside the opened workspace,
  so the first `write`, `execute`, `network`, or Git mutation in a fresh workspace always
  prompts (or is denied non-interactively with E-SEC-001, exit code 5).
- **`browser.control` footnote.** The chapter's per-tool entry lists `network` and
  `process_spawn` as its permissions without a required/optional split; both are rendered
  R here (note 9) per the catalog convention that classes are required when every
  operation needs them.
- **Integration-tool preamble.** Volume 6 chapter 03 states that platform/service
  integration tools declare `external_service_access` and `network` as required, except the
  container-runtime tools `docker.control` and `kubernetes.control`, which declare
  `container_access` in place of `external_service_access` (`docker.control` additionally
  local by default, `network` optional) — the matrix reflects the per-tool entries, which
  agree with that preamble.
