# 10 — Platform Wireframes

This chapter specifies the TUI's platform and system screens: **provider, model,
configuration, plugins, skills, workflows, MCP, errors, help, command palette, quick
actions, update, and recovery**. The core working screens (start, workspace selection,
session, plan, execution, tool call, permission prompt, diff, git, files, context, memory,
logs, costs/tokens) are chapter 09's. The panel system, navigation, focus, keyboard/mouse
handling, resize behavior, and the small-terminal rules are chapter 07's; theming tiers and
the token-to-ANSI mapping are chapter 08's; the interaction patterns these screens compose
(view states, modals, toasts, search, pagination) are chapter 11's; glyph and accessibility
rules are chapter 12's.

## Reading the wireframes

- Every wireframe is drawn at 80 columns in the `ascii` glyph set (FR-TUI-067), the lowest
  tier every supported terminal renders; on richer tiers the identical layout draws with
  the `unicode` set and the chapter 08 token mapping. Wireframes are layout contracts, not
  pixel contracts: column proportions and element order are normative, exact spacing is not.
- `>` marks the focused row; the focused pane carries the emphasized border per chapter 08.
- `[key]` denotes the default keybinding of a registered action (ADR-110); all bindings are
  rebindable through the chapter 07 keymap.
- State words shown in screens (`available`, `degraded`, `running`, `failed`, …) are the
  frozen Volume 2 chapter 09 vocabularies, rendered verbatim — never paraphrased.
- The top bar and the bottom hint bar are the standard chrome shared with chapter 09
  screens; the bottom bar always shows `? help` and `ctrl+k palette` plus screen-local keys.

## The management frame

All thirteen screens instantiate one **management frame**: a list pane (left, ~40% width)
over a collection, and a detail pane (right) over the selected element, with the list
searchable (`/`), filterable (`f`), paginated, and virtualized per chapter 11. Overlay
screens (help, palette, quick actions) float above the current screen instead. On terminals
narrower than the two-pane threshold (chapter 07 resize rules), list and detail stack: the
list fills the screen and `enter` pushes the detail as a full-screen view with `esc`
returning.

### FR-TUI-060 — Platform screen catalog and management frame

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: TUI (Volume 8)
- Affected components: TUI
- Dependencies: ADR-110; ADR-006; chapter 07 shell contract; chapter 11 patterns
- Related risks: RISK-UX-080

#### Description

The TUI provides the thirteen platform screens of this chapter. Eight are management-frame
screens over runtime collections — providers, models, configuration, plugins, skills,
workflows, MCP servers, errors — and five are flow or overlay screens: help, command
palette, quick actions, update, recovery. Each screen MUST: (1) be reachable from the
command palette and from the navigation surface of chapter 07; (2) render entity states
using the frozen Volume 2 state names verbatim; (3) implement the chapter 11 canonical view
states (FR-UX-074); (4) expose every operation as a registered action (ADR-110) with its
permission names and destructiveness class; and (5) have a CLI equivalent for every
operation (driver parity, Volume 3 chapter 06 — the command families of chapters 03–06).
Screen phasing: provider, model, configuration, errors, help, palette, update, and recovery
screens are MVP; plugins, skills, workflows, and MCP screens ship with their subsystems
(Volume 6 phases) and until then the palette omits their actions and the navigation surface
omits their entries.

#### Motivation

Platform state (providers, extensions, configuration, failures, updates) must be
inspectable and operable without leaving the terminal session — PRD-008 — and must present
one consistent operating model rather than thirteen bespoke screens.

#### Actors

Users; the Runtime and ports supplying collection data; extensions contributing actions.

#### Preconditions

TUI started in a workspace (chapter 07); the backing subsystem for the screen is present.

#### Main flow

1. The user opens a screen via palette or navigation.
2. The list pane loads the collection through the owning port with a `loading` state.
3. Selection populates the detail pane; actions execute through the registry.

#### Alternative flows

- Narrow terminal: stacked list/detail navigation per the management-frame rules.
- Collection empty: the `empty` state with the screen's primary action (FR-UX-074).

#### Edge cases

- The backing subsystem is disabled or not yet at phase: the screen is absent from
  navigation and palette, not shown broken.
- A collection mutates while displayed (event-driven): the list updates in place; the
  focused element is preserved by identity (ULID), not by row index.
- An element is removed while focused: focus moves to the next element; the detail pane
  shows the tombstone summary if the entity has one (`removed` state) or empties.

#### Inputs

Collection queries via ports; Event Bus subscriptions for live updates; user navigation.

#### Outputs

Rendered screens; registered action invocations; no direct side effects of its own.

#### States

Screens render, never own, entity states; view-level states per FR-UX-074.

#### Errors

Load failures render the `error` view state with the envelope summary (FR-UX-001 fields);
the E-TUI family (chapter 07) covers TUI-internal failures.

#### Constraints

No business logic in screens: every operation is a Runtime API or port call (Volume 3
driver rule). Screens MUST NOT cache entity state beyond the rendering frame in a way that
survives contradicting events (RISK-UX-080).

#### Security

Mutating actions carry their §5 permission names in the registry and are mediated by the
Permission Manager; destructive actions bind to FR-UX-072 confirmation tiers. Secret
material never renders: credential references and redacted forms only (Volume 9).

#### Observability

Screen navigation emits no events of its own; action execution emits `tui.action.invoked`
plus the owning subsystem's events. Load failures are logged with correlation IDs.

#### Performance

List panes virtualize (FR-UX-076); collection load renders feedback within the NFR-UX-077
deadline. Budgets are Volume 12's.

#### Compatibility

All screens render in every chapter 08 color tier and both glyph sets; layouts hold at
80×24 per the chapter 07 minimum-geometry rules.

#### Acceptance criteria

- Given any of the thirteen screens at phase, when the palette is opened, then the screen's
  open action and its element actions (in context) are listed (registry completeness).
- Given a provider in state `degraded`, when the provider screen renders it, then the exact
  string `degraded` appears — never a synonym (frozen-name case).
- Given a mutating action without a granted permission, when invoked, then the Permission
  Manager decision path runs and denial renders per FR-UX-001 (permission case).
- Given a collection load failure, when the screen renders, then the `error` view state
  shows the envelope's user message and recommended action (error case).
- Negative case: given the plugins subsystem before its phase, when navigation renders,
  then no plugins entry exists and `plugin.*` actions are absent from the palette.
- Observability case: given any action executed from a screen, when events are inspected,
  then `tui.action.invoked` carries the action identifier and source `screen`.

#### Verification method

teatest golden frames per screen per glyph set and color tier; registry-coverage test
(every screen action palette-reachable); driver-parity audit against chapters 03–06;
Volume 13 TUI suite.

#### Traceability

PRD-008, PRD-006; ADR-110, ADR-006, ADR-026; FR-UX-074, FR-UX-076, FR-UX-072; chapter 07
shell contract.

## Provider screen

```text
+------------------------------------------------------------------------------+
| andromeda / providers                          workspace lyra/andromeda    * |
+--------------------------------------+---------------------------------------+
| Providers (4)             / filter   | anthropic                       cloud |
|                                      |                                       |
| > anthropic           available      | State          available              |
|   ollama              available      | Auth           active   profile: work |
|   openai-compat       degraded       | Adapter        anthropic 1.4.0        |
|   azure-openai        disabled       | Endpoint       api.anthropic.com      |
|                                      | Capabilities   chat streaming         |
|                                      |                tool_calling vision    |
|                                      |                embeddings             |
|                                      | Models         12 discovered          |
|                                      | Last verified  2 min ago              |
|                                      | Requests p95   840 ms                 |
|                                      |                                       |
|                                      | [enter] models    [v] verify          |
|                                      | [a] authenticate  [x] disable         |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  f facets  tab pane  esc back      16-color |
+------------------------------------------------------------------------------+
```

The list pane shows every configured Provider with its frozen connection state
(`configured`, `verifying`, `available`, `degraded`, `unavailable`, `disabled`, `removed`).
The detail pane shows the adapter declaration surface (Volume 5): capability names exactly
as the Volume 5 enum spells them, authentication summary (never material — profile name and
Authentication Session state only), discovery count, and health figures. Actions:
`provider.verify` re-runs verification (permission `network`); `provider.authenticate`
launches the Volume 5 auth flow; `provider.disable`/`provider.enable` toggle routing
participation (confirmation tier 2 when runs are active); `enter` opens the model screen
scoped to the provider. Degraded and unavailable providers carry a one-line cause from the
last verification error envelope. Empty state: "No providers configured" with
`provider.add` as primary action.

## Model screen

```text
+------------------------------------------------------------------------------+
| andromeda / providers / anthropic / models                                  * |
+--------------------------------------+---------------------------------------+
| Models (12)               / filter   | claude-sonnet-4-5                     |
|                                      |                                       |
| > claude-sonnet-4-5      default     | Context window   200000 tokens        |
|   claude-haiku-4-5                   | Capabilities     chat  streaming      |
|   claude-opus-4-6                    |   tool_calling  structured_outputs    |
|   ...                                |   vision  reasoning  cancellation     |
|                                      |   token_usage_reporting               |
|                                      | Cost (per 1M)    in 3.00  out 15.00   |
|                                      | Discovered       via model_discovery  |
|                                      | Used by profile  default, review      |
|                                      |                                       |
|                                      | [enter] set default for workspace     |
|                                      | [p] use in current session            |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  esc providers                     16-color |
+------------------------------------------------------------------------------+
```

The model screen lists the models a provider exposes (`DiscoverModels`), with declared
capabilities (Volume 5 enum names verbatim), context window, and cost figures *as reported
by the provider's declaration* — the TUI never invents pricing; when cost reporting is
undeclared the row shows `cost: not reported`. `model.set_default` writes workspace
configuration (Volume 10 precedence applies and the config screen shows the attribution);
`model.use_in_session` re-binds the active session's Agent Profile selection and, per
Volume 5, any provider/model change is notified in the session transcript. Switching models
mid-run follows Volume 4 run semantics; the action is disabled (with reason) while a turn
is in flight.

## Configuration screen

```text
+------------------------------------------------------------------------------+
| andromeda / configuration                        profile: default            |
+--------------------------------------+---------------------------------------+
| Tables                    / filter   | [providers.anthropic]                 |
|                                      |                                       |
|   [agent]                            | enabled = true          (workspace)   |
|   [providers]                        | endpoint = "https://api.anthropic..." |
| > [providers.anthropic]              |                         (project)     |
|   [tui]                              | auth_profile = "work"   (global)      |
|   [tui.theme]                        | timeout_ms = 60000      (default)     |
|   [permissions]                      |                                       |
|   [logging]                          | Source layers: flag > env > runtime   |
|                                      |  > project > workspace > profile      |
|                                      |  > global > defaults  (Volume 10)     |
|                                      |                                       |
| Validation: 0 errors, 1 warning      | [e] edit in $EDITOR   [enter] source  |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  esc back                          16-color |
+------------------------------------------------------------------------------+
```

The configuration screen renders the **resolved** configuration (`ConfigPort.Resolve`) as a
table tree; every value shows its source attribution layer in parentheses — the diagnostic
ConfigPort guarantees. `enter` on a value expands the full layer stack for that key (which
layers define it, which one won). Values of secret-typed keys render as references, never
material. The screen is read-only at MVP: `config.edit` hands off to the `config edit` CLI
flow (editor launch, then validation), keeping one write path; in-TUI value editing is
Beta. The validation strip surfaces the `Validate` report count; selecting it opens the
error detail with exact E-CFG messages (Volume 10). Empty state does not occur (defaults
always resolve); load failure renders the `error` state with exit-code-3 semantics
explained.

## Plugins screen

```text
+------------------------------------------------------------------------------+
| andromeda / plugins                                                         * |
+--------------------------------------+---------------------------------------+
| Plugins (3)               / filter   | andromeda-jira                        |
|                                      |                                       |
| > andromeda-jira        running      | State        running                  |
|   sec-scanner           stopped      | Version      0.9.2   ARP 1.0          |
|   legacy-notes          failed       | Origin       registry: community      |
|                                      | Trust        signed, verified         |
|                                      | Surfaces     2 tools, 1 command       |
|                                      | Permissions  network,                 |
|                                      |              external_service_access  |
|                                      | Restarts     0 in current session     |
|                                      | Last error   -                        |
|                                      |                                       |
|                                      | [s] stop  [r] restart  [x] disable    |
|                                      | [l] logs  [u] update  [del] remove    |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  esc back                          16-color |
+------------------------------------------------------------------------------+
```

Lists every registered Plugin with its frozen state (`registered`, `starting`, `running`,
`stopping`, `stopped`, `failed`, `disabled`, `removed`). The detail pane shows the ARP
handshake facts (protocol version), origin and trust/signature summary (Volume 6 vocabulary),
registered surfaces, granted permission names, and restart-policy counters. Actions map to
the Volume 6 lifecycle: stop/restart/disable/enable; `plugin.update` and `plugin.remove` go
through PackagePort (`package_installation` permission; remove is confirmation tier 3 —
typed name — because extension removal tombstones registration). `plugin.logs` opens the
logs screen (chapter 09) pre-filtered to the plugin's correlation IDs. A `failed` plugin
shows its last error envelope summary and the restart policy's next step.

## Skills screen

```text
+------------------------------------------------------------------------------+
| andromeda / skills                                                          * |
+--------------------------------------+---------------------------------------+
| Skills (5)                / filter   | conventional-review                   |
|                                      |                                       |
| > conventional-review    enabled     | Version      2.1.0                    |
|   sdd-scaffold           enabled     | Origin       workspace .andromeda/    |
|   release-notes          enabled     | Requires     tools: git.diff,         |
|   db-migrations          disabled    |   filesystem.read                     |
|   perf-triage            enabled     | Capabilities tool_calling             |
|                                      | Providers    any declaring required   |
|                                      |   capabilities                        |
|                                      | Composition  extends: review-base     |
|                                      | Validation   manifest ok, fixtures ok |
|                                      |                                       |
|                                      | [enter] view manifest  [x] disable    |
|                                      | [t] run fixtures       [del] remove   |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  esc back                          16-color |
+------------------------------------------------------------------------------+
```

Lists installed Skills with enablement flags (Skills carry no frozen state machine; the
`enabled` flag and validation results are what the screen shows). Detail: manifest identity
and version, origin (workspace, global, package), required tools and capabilities (names
verbatim from the Volume 6 manifest and Volume 5 enum), provider compatibility as declared,
composition/inheritance chain, and the last validation outcome. `skill.view_manifest`
opens a read-only viewer; `skill.run_fixtures` executes the skill's test fixtures (Volume 6)
as a run visible in the execution screen; install/remove go through PackagePort with
`package_installation`. The screen's empty state offers `skill.install` and points to the
workspace skill directory convention.

## Workflows screen

```text
+------------------------------------------------------------------------------+
| andromeda / workflows                                     tab: [runs] defs   |
+--------------------------------------+---------------------------------------+
| Workflow Runs (2 active)  / filter   | sdd: feature/auth-cache               |
|                                      |                                       |
| > sdd  feature/auth-cache            | State     awaiting_approval           |
|        awaiting_approval             | Stage     implementation (7 of 14)    |
|   sdd  fix/flaky-index               | Started   14:02  elapsed 41 min       |
|        running                       | Approval  apply patch to src/index/   |
|   release-prep  v0.4.1               |           requested 2 min ago         |
|        completed                     | Progress  [######........] 7/14       |
|                                      | Agents    planner, implementer        |
|                                      | Artifacts 3 patches, 1 report         |
|                                      |                                       |
|                                      | [enter] open approval  [p] pause      |
|                                      | [c] cancel             [l] logs       |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  tab switch list  esc back         16-color |
+------------------------------------------------------------------------------+
```

Two tabs over one frame: **runs** (Workflow Run instances, frozen states `pending`,
`running`, `awaiting_approval`, `paused`, `interrupted`, `completed`, `failed`,
`cancelled`) and **defs** (installed workflow definitions with version and stage count).
For SDD runs the detail pane shows the Volume 4 stage sequence position as `stage (n of
14)` with a determinate progress bar (FR-UX-071). `workflow.open_approval` jumps to the
permission/approval prompt (chapter 09 screen); pause/resume/cancel map to Volume 4 run
semantics with tier-2 confirmation on cancel. An `interrupted` run shows the resume action
and the last persisted stage boundary. Empty runs tab: "No workflow runs yet" with
`workflow.start` primary action listing definitions.

## MCP screen

```text
+------------------------------------------------------------------------------+
| andromeda / mcp                                                             * |
+--------------------------------------+---------------------------------------+
| MCP Servers (3)           / filter   | github-mcp                            |
|                                      |                                       |
| > github-mcp            ready        | State       ready                     |
|   docs-search           reconnecting | Transport   stdio                     |
|   team-tools            disabled     | Protocol    negotiated at initialize  |
|                                      | Offers      14 tools, 2 resources,    |
|                                      |             1 prompt                  |
|                                      | Trust       user-approved, workspace  |
|                                      | Uptime      2 h 14 min                |
|                                      | Last error  -                         |
|                                      |                                       |
|                                      | [enter] browse offers  [r] reconnect  |
|                                      | [x] disable  [l] logs  [del] remove   |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  esc back                          16-color |
+------------------------------------------------------------------------------+
```

Lists MCP Client Connections with their frozen states (`configured`, `connecting`,
`initializing`, `ready`, `reconnecting`, `disconnected`, `failed`, `disabled`, `removed`).
Detail: transport kind, negotiated protocol summary, offered tools/resources/prompts
counts, trust summary per Volume 6's MCP trust model, uptime, and the last error envelope
for `failed`/`reconnecting`. `mcp.browse` opens the offers list (tools invocable per the
Volume 6 bridging rules); `mcp.reconnect` re-establishes transport (permissions `network`
and `external_service_access` as declared at registration); disable/remove follow the
Volume 6 lifecycle with tier-2/tier-3 confirmations respectively. A `reconnecting` row
shows attempt count and next retry time — visible degradation, per FR-UX-074.

## Errors screen

```text
+------------------------------------------------------------------------------+
| andromeda / errors                              session 01JX8...  (all: F)   |
+--------------------------------------+---------------------------------------+
| Recent errors (7)         / filter   | E-PROV-0xx  provider request failed   |
|                                      |                                       |
| > 14:23 E-PROV error  provider      | Severity     error                    |
|   14:21 E-TOOL error  tool          | Time         14:23:07                 |
|   14:05 E-CFG  warn   config        | User msg     Provider anthropic did   |
|   13:58 E-GIT  error  git           |   not respond within 60 s.            |
|   ...                                | Cause        timeout                  |
|                                      | Recoverable  yes - retry applies      |
|                                      | Recommended  retry; check provider    |
|                                      |   status page; see logs               |
|                                      | Correlation  01JX8Q7R...              |
|                                      | Exit code    7 (if CLI)               |
|                                      |                                       |
|                                      | [enter] open run  [y] copy diag       |
|                                      | [r] retry         [l] related logs    |
+--------------------------------------+---------------------------------------+
| ? help  ctrl+k palette  / filter  F all-sessions  esc back          16-color |
+------------------------------------------------------------------------------+
```

The error center lists error envelopes (ADR-016) recorded in the current session, with a
toggle to all recent errors in the workspace. The list row shows time, code family,
severity, and source area; the detail pane renders the envelope's user-facing fields —
user message, cause classification, recoverability, retry policy, recommended action,
correlation ID, and exit-code mapping — using the FR-UX-001 presentation standard.
Technical messages and safe-context data appear under an expandable section; secret-bearing
fields never exist in envelopes by construction (Volume 9 redaction). `error.copy_diagnostics`
copies the redacted envelope per ADR-113; `error.retry` re-dispatches through the owning
command path only when the envelope's retry policy allows; `error.open_run` and
`error.related_logs` navigate by correlation ID. Errors surfaced as toasts (FR-UX-073)
always land here, so dismissal never loses information. The example row texts are
illustrative; concrete codes come from each area's catalog.

## Help overlay

```text
+------------------------------------------------------------------------------+
|                                                                              |
|   +----------------------- help - providers screen ----------------------+  |
|   | / search help...                                                     |  |
|   |                                                                      |  |
|   | This screen                                                          |  |
|   |   enter  open models        v  verify provider     a  authenticate  |  |
|   |   x      disable            /  filter list         f  facets        |  |
|   |                                                                      |  |
|   | Global                                                               |  |
|   |   ctrl+k palette            ?  this help            q  quit          |  |
|   |   tab    next pane          esc back                 g  git screen   |  |
|   |                                                                      |  |
|   | All actions and bindings: [enter] full reference (searchable)        |  |
|   | Version 0.4.0  -  docs: local `andromeda help`, andromeda.dev        |  |
|   +----------------------------------------------------------------------+  |
|                                                                              |
+------------------------------------------------------------------------------+
```

### FR-TUI-063 — Help overlay and keybinding reference

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: TUI (Volume 8)
- Affected components: TUI
- Dependencies: ADR-110; chapter 07 keymap
- Related risks: RISK-UX-080

#### Description

`?` (outside text inputs) and `f1` open the help overlay. It renders, from the ADR-110
registry: the focused screen's actions with current (not default) bindings first, then
global actions, then an `enter`-reachable full reference of every registered action —
searchable, grouped by screen, showing identifier, title, binding, and required
permissions. Help content is local: no network access, ever (PRD-003). The overlay is
modal per FR-UX-072, closes with `esc`, and never obscures an active permission prompt.

#### Motivation

Discoverability without leaving the terminal; the registry guarantees the help surface is
complete rather than curated.

#### Actors

Users.

#### Preconditions

TUI running; registry populated.

#### Main flow

1. User presses `?`; the overlay renders contextual actions from the registry.
2. `enter` expands to the searchable full reference; `esc` returns, `esc` again closes.

#### Alternative flows

- Accessible output mode: help renders as a numbered linear list (FR-TUI-065).

#### Edge cases

- Rebound keys: help always shows the *effective* binding from the keymap, including
  unbound actions (shown with `-`), so help never teaches stale keys.
- Terminal below minimum overlay size: help renders full-screen instead of floating.

#### Inputs

Registry entries; keymap state; search input.

#### Outputs

Rendered overlay; no side effects.

#### States

Overlay open/closed; no persisted state.

#### Errors

None of its own; a registry inconsistency is a defect surfaced by tests, not at runtime.

#### Constraints

Zero network; zero business logic; content derives from the registry and static local
documentation only.

#### Security

Help displays permission names per action — part of the transparency posture (PRD-006).

#### Observability

Opening help emits no event; executing an action from help emits `tui.action.invoked`
with source `help`.

#### Performance

Registry rendering is O(actions); search filters incrementally (FR-UX-076 rules).

#### Compatibility

Renders in both glyph sets and all color tiers; full-screen fallback below overlay
minimums.

#### Acceptance criteria

- Given any screen, when `?` is pressed, then every action available in that context
  appears with its effective binding (completeness case).
- Given a rebound action, when help renders, then the new binding is shown (staleness
  negative case).
- Given the offline condition, when help is used, then zero network attempts occur
  (observability/offline case).
- Given accessible output mode, when help opens, then a numbered linear list renders and
  every item is selectable by number (accessibility case).

#### Verification method

Registry-coverage test diffing help content against the registry; teatest golden frames;
offline suite assertion; keymap-rebind test.

#### Traceability

PRD-003, PRD-006, PRD-008; ADR-110; NFR-TUI-069; FR-UX-076.

## Command palette

```text
+------------------------------------------------------------------------------+
|  session: feature/auth-cache                                    (dimmed)     |
|   +--------------------------- command palette ------------------------+     |
|   | > prov ver_                                                        |     |
|   |                                                                    |     |
|   | > provider.verify        Verify provider connectivity          v  |     |
|   |   provider.add           Add a provider                            |     |
|   |   provider.authenticate  Authenticate provider                 a  |     |
|   |   x.jira.link-issue      Link Jira issue (plugin)                  |     |
|   |                                                                    |     |
|   |   recent: session.interrupt - diff.open - update.check             |     |
|   |   4 of 212 actions        [enter] run  [tab] complete  [esc] close |     |
|   +--------------------------------------------------------------------+     |
|                                                                              |
+------------------------------------------------------------------------------+
```

### FR-TUI-061 — Command palette

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: TUI (Volume 8)
- Affected components: TUI
- Dependencies: ADR-110; ADR-104 (namespace mirror); FR-UX-072 (overlay rules)
- Related risks: RISK-UX-080

#### Description

`ctrl+k` opens the command palette: a centered modal overlay with an input line and a
result list over the ADR-110 registry, filtered to actions whose context predicate holds.
Matching is subsequence-based over identifier and title with exact-prefix ranked first,
then recency, then registry order; match count and total are always visible. Each row
shows identifier, title, and effective binding. `enter` executes the selection through the
identical permission/confirmation path as any other surface; `esc` closes without effect.
Extension actions appear under their `x.` prefix with an origin tag. Actions currently
unavailable in context do not appear (they are listed in the help full reference instead,
marked with their availability condition). The palette never blocks background work:
streams continue and surface as toasts while it is open.

#### Motivation

One reachable-from-anywhere entry point makes every capability discoverable and keyboard-
reachable — the enforcement surface for NFR-TUI-069 and the primary navigation accelerator.

#### Actors

Users; extensions contributing actions.

#### Preconditions

TUI running; registry populated.

#### Main flow

1. `ctrl+k` opens the palette with empty input showing recent actions.
2. Typing filters; `enter` executes the focused action; the palette closes.

#### Alternative flows

- Destructive selection: the FR-UX-072 confirmation tier runs before execution.
- Action requiring an ungranted permission: the Permission Manager path runs exactly as if
  invoked from its screen.

#### Edge cases

- Zero matches: the list shows "no matching actions" with the query kept editable — never
  a silent empty box.
- Palette opened while a modal confirmation is pending: refused with a status-bar notice;
  confirmations hold exclusive focus (FR-UX-072).
- Two actions with identical titles: identifiers disambiguate; identifiers are unique by
  registry construction.

#### Inputs

Registry; keymap; typed query; recency history (session-scoped, persisted with the
session record).

#### Outputs

Executed action or nothing; `tui.palette.opened` and `tui.action.invoked` events.

#### States

Open/closed; query string; focused row. Recency history persists per session (Volume 10
storage).

#### Errors

Execution errors belong to the executed action and render per FR-UX-001; the palette
itself only fails with the TUI (E-TUI family, chapter 07).

#### Constraints

Filtering runs on the UI thread within the chapter 11 coalescing budget; result list is
virtualized (FR-UX-076) — the registry may exceed the viewport.

#### Security

The palette grants nothing: permissions and confirmations attach to actions. Recency
history contains action identifiers only — no arguments, no content.

#### Observability

`tui.palette.opened` on open; `tui.action.invoked` with source `palette` on execution.

#### Performance

Keystroke-to-filtered-list within the NFR-UX-077 feedback deadline at 1,000 registered
actions.

#### Compatibility

Both glyph sets; all color tiers; full-screen fallback below overlay minimums; `ctrl+k`
rebindable where a terminal or multiplexer consumes it.

#### Acceptance criteria

- Given 212 available actions and query `prov ver`, when the palette filters, then
  `provider.verify` ranks first and the count row reads `n of 212` (ranking case).
- Given a destructive action selected, when `enter` is pressed, then the tier-appropriate
  confirmation renders before any effect (permission/confirmation case).
- Given an action absent from the current context, when queried, then it does not execute
  and does not appear (negative case).
- Given the palette open during a streaming run, when tokens arrive, then the stream
  continues and no token output is lost (concurrency case).
- Observability case: executing from the palette emits `tui.action.invoked` with source
  `palette` and the action identifier.

#### Verification method

teatest interaction scripts (ranking, zero-match, confirmation, concurrency); registry
coverage test; latency measurement per NFR-UX-077 method.

#### Traceability

PRD-008; ADR-110, ADR-104; NFR-TUI-069, NFR-UX-077; FR-UX-072, FR-UX-076.

## Quick actions

```text
+------------------------------------------------------------------------------+
|  ...session content (dimmed)...                                              |
|                                                                              |
|   +---------------------------- quick actions ---------------------------+   |
|   | 1 approve pending tool call        4 open diff of last change        |   |
|   | 2 interrupt current run            5 copy last error diagnostics     |   |
|   | 3 switch model for session         6 open providers screen           |   |
|   +----------------------------------------------------------------------+   |
+------------------------------------------------------------------------------+
```

### FR-TUI-062 — Quick actions

- Type: Functional
- Status: Approved
- Priority: P2
- Phase: Beta
- Source: Provided
- Owner: TUI (Volume 8)
- Affected components: TUI
- Dependencies: ADR-110; FR-TUI-061
- Related risks: RISK-UX-080

#### Description

`ctrl+space` opens a compact overlay of at most 8 context-ranked actions from the registry,
numbered 1–8 for one-keystroke execution. Ranking inputs, in order: pending items demanding
attention (approvals, confirmations, recoverable errors), then actions relevant to the
focused element's state, then session recency. The set is deterministic for a given UI
state (testable), and every quick action is, by construction, also in the palette — quick
actions are a ranked subset, never an exclusive surface.

#### Motivation

The palette optimizes for search; quick actions optimize for the two-keystroke common case
— especially "something is waiting on me" moments.

#### Actors

Users.

#### Preconditions

TUI running.

#### Main flow

1. `ctrl+space` opens the overlay; digits execute; `esc` closes.

#### Alternative flows

- More than 8 candidates: the 8 highest-ranked render; the palette covers the rest.

#### Edge cases

- No applicable actions: overlay shows "nothing pending" with the palette hint — not an
  empty frame.
- A ranked action becomes unavailable between render and keypress (state changed): the
  execution re-checks context and refuses with a status notice instead of acting stale.

#### Inputs

Registry with context predicates; pending-attention queue; recency.

#### Outputs

Executed action; `tui.action.invoked` with source `quick_actions`.

#### States

Open/closed; no persistence beyond session recency shared with the palette.

#### Errors

As FR-TUI-061: errors belong to executed actions.

#### Constraints

Maximum 8 entries; single-row-per-action layout; renders within the NFR-UX-077 deadline.

#### Security

Identical mediation to all surfaces; a quick action for an approval opens the approval
prompt — it never approves directly with one keystroke.

#### Observability

`tui.action.invoked` with source `quick_actions`.

#### Performance

Ranking is O(available actions) over in-memory state.

#### Compatibility

Both glyph sets; all tiers; where `ctrl+space` does not transmit (terminal limitation),
the binding is remappable and the palette remains the guaranteed path.

#### Acceptance criteria

- Given a pending approval, when quick actions open, then the approval action is entry 1
  (ranking case).
- Given entry 2 is "interrupt current run", when `2` is pressed, then the interrupt
  confirmation (tier 2) renders — one keystroke never destroys work (confirmation case).
- Given a stale entry, when its digit is pressed after state changed, then execution is
  refused with a notice (negative case).
- Observability case: execution emits `tui.action.invoked` with source `quick_actions`.

#### Verification method

teatest scripted scenarios over ranking determinism, staleness refusal, and confirmation
binding.

#### Traceability

PRD-008; ADR-110; FR-TUI-061; FR-UX-072; NFR-UX-077.

## Update screen

```text
+------------------------------------------------------------------------------+
| andromeda / update                                     channel: stable        |
+------------------------------------------------------------------------------+
|                                                                              |
|  Installed   0.4.0        Available   0.4.1                                  |
|  State       downloading                                                     |
|                                                                              |
|  [############################..............]  64%   12.3 / 19.2 MB          |
|                                                                              |
|  Steps   checking        done                                                |
|          downloading     in progress                                         |
|          verifying       pending  (checksums, signatures per policy)         |
|          applying        pending                                             |
|                                                                              |
|  Release notes summary: fixes for index staleness; provider retry tuning.    |
|                                                                              |
|  [enter] continue  [c] cancel  [b] rollback to 0.3.9  [esc] background       |
+------------------------------------------------------------------------------+
| ? help  ctrl+k palette                                              16-color |
+------------------------------------------------------------------------------+
```

The update screen drives UpdaterPort through the frozen Update process states (`checking`,
`up_to_date`, `update_available`, `downloading`, `verifying`, `applying`, `applied`,
`failed`, `rolled_back`), showing the current state verbatim plus a step list mapping the
remaining machine path. Semantics are Volume 14's: the screen only presents. `update.check`
requires `network` and fails cleanly offline (the `offline` view state names the
constraint); `update.apply` requires `system_modification`, is a tier-2 confirmation, and
is refused while runs are active (with the list of blocking runs); `update.rollback` is
offered when a retained previous version exists and works offline. Download shows a
determinate progress bar (FR-UX-071); verification failure renders the E-REL envelope with
the recommended action and never offers apply. `esc` backgrounds a running download and
progress continues in the status bar.

## Recovery screen

```text
+------------------------------------------------------------------------------+
| andromeda / recovery                        found 2 interrupted items         |
+------------------------------------------------------------------------------+
|                                                                              |
|  Interrupted work was found from a previous process (state: interrupted).    |
|  Nothing has been resumed automatically.                                     |
|                                                                              |
|  > run    fix/flaky-index      interrupted 13 min ago                        |
|           last persisted: turn 12, task 3 running, 2 tasks completed         |
|    session refactor-auth       interrupted 2 days ago                        |
|           last persisted: suspended at turn 40                               |
|                                                                              |
|  [enter] resume   [i] inspect first   [d] discard...   [esc] not now         |
|                                                                              |
|  Discard requires typing the item name and keeps records; it never deletes   |
|  history. Resume continues from the last persisted state only.               |
+------------------------------------------------------------------------------+
| ? help  ctrl+k palette                                              16-color |
+------------------------------------------------------------------------------+
```

### FR-TUI-064 — Error center and recovery screens

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: TUI (Volume 8)
- Affected components: TUI
- Dependencies: FR-UX-001 (presentation); ADR-016 (envelope); ADR-113 (copy); Volume 4
  resume semantics; SessionStorePort
- Related risks: RISK-UX-080, RISK-UX-079

#### Description

Two screens close the failure loop. The **error center** lists recorded error envelopes
(session scope, toggleable to workspace scope) and renders the full ADR-016 envelope
fields per FR-UX-001, with navigation by correlation ID to runs and logs, redacted-
diagnostics copy (ADR-113), and retry only where the envelope's retry policy permits. The
**recovery screen** renders at TUI start whenever `SessionStorePort.MarkInterrupted`
reports non-terminal work (frozen state `interrupted`): it lists interrupted sessions and
runs with their last persisted position, and offers resume, inspect, and discard. Resume
follows Volume 4 semantics (continuation from persisted state only — interrupted work is
never assumed complete, PRD-010); nothing resumes automatically; discard is a tier-3
confirmation (typed name) and marks work closed without deleting records.

#### Motivation

Failures and interruptions are product moments, not edge cases: PRD-010 makes recovery a
core promise, and PRD-006 requires errors to be inspectable with their full context.

#### Actors

Users; SessionStorePort; the Runtime resume path.

#### Preconditions

Error center: at least the session store readable. Recovery: startup scan completed.

#### Main flow

1. TUI start runs the interrupted scan; when non-empty, the recovery screen renders before
   any session opens.
2. The user resumes, inspects, or discards each item, or defers with `esc`.
3. Errors accumulate in the error center as envelopes are recorded; toast-surfaced errors
   deep-link here.

#### Alternative flows

- Resume fails (e.g., workspace moved): the failure envelope renders inline on the
  recovery row; the item stays listed.
- `esc` defers recovery: a persistent status-bar badge shows the interrupted count until
  addressed.

#### Edge cases

- Interrupted items from a newer schema version: shown read-only with the ADR-029
  integrity explanation; resume disabled with reason (exit-code-9 semantics at the CLI).
- Zero errors: the error center's `empty` state states that explicitly — an empty error
  list is information.
- Hundreds of interrupted items (crash loop): the list virtualizes and offers a bulk
  discard action, itself tier-3.

#### Inputs

`MarkInterrupted` results; run snapshots (`LoadRun`); recorded envelopes; user choices.

#### Outputs

Resume invocations to the Runtime; discard state transitions; copied diagnostics.

#### States

Renders the frozen `interrupted` state and terminal outcomes; owns no machine of its own.

#### Errors

Resume and discard failures render their envelopes inline; the screen never masks a
failure with a generic message (FR-UX-001).

#### Constraints

The recovery screen MUST render before any automatic session activation; automatic resume
MUST NOT occur regardless of configuration.

#### Security

Discard confirmations are tier 3; diagnostics copies are redacted (ADR-113; RISK-UX-079);
error detail renders only envelope-safe fields by default.

#### Observability

Recovery decisions execute registered actions (`tui.action.invoked`); the underlying
resume/discard transitions emit their Volume 4 entity events with correlation IDs.

#### Performance

The startup scan is bounded by `MarkInterrupted`'s single call; list rendering virtualizes
(FR-UX-076); budgets per Volume 12 session-restore targets.

#### Compatibility

Both glyph sets; all tiers; the recovery screen holds at 80×24 with stacked layout.

#### Acceptance criteria

- Given an interrupted run, when the TUI starts, then the recovery screen lists it with
  its last persisted position and nothing has resumed (PRD-010 case).
- Given a discard, when confirmed by typed name, then the item leaves the pending list and
  its records remain loadable (destructive-confirmation case).
- Given a resume failure, when rendered, then the envelope appears inline and the item
  remains actionable (error case).
- Given an error with retry policy "not retryable", when its detail renders, then no retry
  action is offered (negative case).
- Permission case: copying diagnostics emits `tui.clipboard.copied` with kind
  `error_diagnostics` and byte count, never content.
- Observability case: resume emits the run's transition events with the run's correlation
  ID reachable from the error center by ID.

#### Verification method

Crash-injection suite (SM-11 method) driving the recovery screen via teatest; envelope
rendering golden tests; retry-policy conformance checks; clipboard audit assertion.

#### Traceability

PRD-010, PRD-006; ADR-016, ADR-029, ADR-113; FR-UX-001, FR-UX-073, FR-UX-076; Volume 4
resume semantics.
