# 09 — Wireframes: Core Screens

This chapter specifies the core screen set with ASCII wireframes and their behavioral
prose. Frames show the **standard layout class at the 80×24 reference size** (FR-TUI-002);
the wide class adds the sidebar, the compact class collapses chrome per FR-TUI-006.
Wireframes are normative for structure — which panels exist, what content each carries,
which actions are offered — and illustrative for exact spacing and sample data. Golden
frames (NFR-TUI-002) are the pixel-exact authority once implementation exists; drift
between the two is RISK-TUI-003.

Conventions in the frames: `[x]` marks a key hint; `>` marks the selected row; `*` marks
the focused panel's title; box characters are drawn in ASCII here — the shell renders
Unicode box drawing with the ASCII fallback owned by chapter 12. All state words are the
frozen vocabularies of Volume 2, chapter 09. The status bar's right side always shows
session token and cost totals (PRD-006); its left side shows contextual keys. Platform
screens (provider, model, configuration, plugins, skills, workflows, MCP, errors, help,
palette, quick actions, update, recovery) are chapter 10's.

## Splash / start

```text
+------------------------------------------------------------------------------+
|                                                                              |
|                                                                              |
|                                   /\_____/\                                  |
|                              +   /  o   o  \    *                            |
|                                  \    ^    /   + +                           |
|                             * +   \ '---' /     *                            |
|                                   /       \                                  |
|                                  (  |   |  )                                 |
|                                  (__|___|__)                                 |
|                                                                              |
|                            a n d r o m e d a                                 |
|                                                                              |
|             Your terminal companion for shipping great software.             |
|                                                                              |
|                              v0.4.2 (stable)                                 |
|                                                                              |
|                                                                              |
|          [Enter] open workspace    [g w] choose workspace    [?] help        |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|          first run: config created at ~/.config/andromeda/config.toml        |
+------------------------------------------------------------------------------+
```

The splash renders the official marks (ADR-026): the ASCII cat mascot with four-pointed
star accents, the lowercase wordmark, and the tagline, plus the installed version and
channel. It shows per the `tui.splash` policy (FR-UX-043): `auto` shows it on first run
and on the first start after a version change, for at most 2 s or until any key; `always`
shows it on every start; `never` skips straight to the start screen. The bottom line
carries first-run information when applicable (created configuration path). Any key
dismisses it; it never blocks on network or workspace I/O — discovery proceeds behind it.
At the `none` tier the marks render in plain text; the cat is already pure ASCII by
design.

## Workspace selection

```text
+------------------------------------------------------------------------------+
| andromeda | workspaces                                     no workspace open |
+------------------------------------------------------------------------------+
| *Recent workspaces ----------------------------------------------------------|
|                                                                              |
| > myproject           ~/code/myproject          2 sessions   10 min ago      |
|   api-gateway         ~/code/api-gateway        7 sessions   yesterday       |
|   dotfiles            ~/dotfiles                1 session    3 days ago      |
|   experiments         ~/tmp/experiments         0 sessions   2 weeks ago     |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
| Details ---------------------------------------------------------------------|
| ~/code/myproject                                                             |
| git: main +2 ~1 | index: ready | profile: default                            |
| last session: 01JZX... "add retry logic to fetcher" (suspended)              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
| [Enter]open [n]new here [b]browse path [d]forget [q]quit                     |
+------------------------------------------------------------------------------+
```

Two panels: the recent-workspaces list (focused by default) and a read-only details panel
for the selection, showing path, VCS summary, index state (frozen Index states), active
profile, and the most recent session with its frozen Session state. `Enter` opens the
selected workspace (WorkspacePort `Open`) and proceeds to `tui.default_screen`; `n`
initializes the current directory (equivalent to `andromeda init`, chapter 03 semantics,
with its confirmation); `b` opens a path input overlay; `d` removes the entry from the
recents registry only — it never touches the workspace on disk, and says so in its
confirmation. An empty recents list renders the chapter 11 empty state with the same `n`
and `b` actions. Opening failures (E-AGT family from WorkspacePort) render in the details
panel with the error code and a retry key, leaving the list usable.

## Session

```text
+------------------------------------------------------------------------------+
| andromeda | session  myproject | main | anthropic:sonnet | run: running      |
+------------------------------------------------------------------------------+
| *Transcript ---------------------------------------------------------------- |
|                                                                              |
| you > Add retry logic to the fetcher and cover it with tests.                |
|                                                                              |
| agent > I'll plan this in three steps: wrap fetch calls with a               |
|   backoff policy, add unit tests for the retry paths, and update             |
|   the changelog. Creating the plan now...                                    |
|                                                                              |
|   plan 01JZX8 proposed (3 tasks)                    [g p] review plan        |
|                                                                              |
| agent > Task 1: editing internal/fetch/client.go                             |
|   ~ tool fs.edit running (2.1s)  internal/fetch/client.go                    |
|   | @@ -41,6 +41,18 @@ func (c *Client) Get(ctx ...                          |
|   ... streaming ...                                                          |
|                                                                              |
+------------------------------------------------------------------------------+
| > _                                                                          |
|   [Enter]send  [Ctrl+C]interrupt run                                         |
+------------------------------------------------------------------------------+
| [Tab]focus [g]goto [/]search [?]help [^P]palette   12.4k tok | $0.42 | run   |
+------------------------------------------------------------------------------+
```

The primary screen (MVP item 2): a virtualized transcript viewport (chapter 11
virtualization) and a single-line input that grows to at most 5 rows. The transcript
renders turns with role markers, streaming deltas coalesced per FR-TUI-007, inline tool
activity summaries (name, frozen Tool Invocation state, elapsed time, primary argument),
and inline references to artifacts (plans, diffs) with their go-to keys. The header shows
workspace, branch, `provider:model`, and the run's frozen state; the status bar totals
tokens and cost live (PRD-006). While a run is `running`, the input stays available for
queued follow-up (delivered per Volume 4 turn semantics); `Ctrl+C` requests interruption
with a confirmation modal. When the transcript is scrolled up during streaming, a "new
output below" pill appears instead of forced auto-scroll; `End` resumes following. With
no session yet, the empty state offers starting one and recent-sessions resume per the
Session `suspended` state.

## Plan

```text
+------------------------------------------------------------------------------+
| andromeda | plan  01JZX8 v2 | run 01JZX7 | state: proposed                   |
+------------------------------------------------------------------------------+
| *Tasks --------------------------------------------------------------------- |
|                                                                              |
| > 1. pending  Wrap fetch calls with backoff policy                           |
|      touches: internal/fetch/client.go, internal/fetch/retry.go              |
|      permissions: write                                                      |
|   2. pending  Unit tests for retry paths            depends on: 1            |
|      touches: internal/fetch/retry_test.go                                   |
|      permissions: write                                                      |
|   3. pending  Update CHANGELOG                       depends on: 1, 2        |
|      permissions: write                                                      |
|                                                                              |
| Detail --------------------------------------------------------------------- |
| Task 1 - Wrap fetch calls with backoff policy                                |
| Rationale: Get/Post call sites lack retry; transient 5xx fail runs.          |
| Approach: introduce retry.Do with exponential backoff, jitter, max 3.        |
| Verification: unit tests simulate 502 then success; go test ./internal/...   |
|                                                                              |
| Plan history: v1 superseded by v2 (revised after context review)             |
|                                                                              |
| [a]approve [r]request revision [x]reject [Enter]detail [g x]execution        |
+------------------------------------------------------------------------------+
```

Renders the active Plan (frozen Plan states in the header; task rows carry frozen Task
states). The tasks panel lists ordered tasks with dependencies, touched paths, and
declared permissions from the plan record; the detail panel shows the selection's
rationale, approach, and verification notes. `a` approves (Plan `proposed` → `approved`,
via the Runtime; where policy requires an Approval, the permission prompt renders first);
`r` sends a revision request with a text overlay (Planner produces a successor version —
`revising`/`superseded` semantics are Volume 4's); `x` abandons after confirmation. Plan
history renders prior versions with their terminal states; selecting one shows it
read-only with a "superseded" banner. During `executing`, task states update live and the
screen becomes the plan-side twin of the execution screen.

## Execution

```text
+------------------------------------------------------------------------------+
| andromeda | execution  run 01JZX7 | state: running | elapsed 04:12           |
+------------------------------------------------------------------------------+
| *Tasks --------------------------------- | Current task ---------------------|
|                                          |                                   |
| > 1. running    Wrap fetch calls  02:58  | 1. Wrap fetch calls with backoff  |
|   2. ready      Unit tests               |                                   |
|   3. pending    Update CHANGELOG         | tool fs.edit succeeded 1.2s       |
|                                          | tool terminal.run executing 12s   |
|                                          |   $ go test ./internal/fetch/...  |
|                                          |   ok   internal/fetch  0.41s      |
|                                          |   ... streaming ...               |
|                                          |                                   |
|                                          | retries: 0 | writes: 2 files      |
|                                          |                                   |
|------------------------------------------+-----------------------------------|
| Timeline ------------------------------------------------------------------- |
| 14:02:11 run started            14:02:14 plan 01JZX8 approved                |
| 14:02:20 task 1 running         14:03:02 tool fs.edit succeeded              |
| 14:05:01 tool terminal.run executing                                         |
|                                                                              |
| [Enter]task detail [t]tool calls [p]pause [Ctrl+C]interrupt [g d]diff        |
+------------------------------------------------------------------------------+
```

Three panels: the task list (frozen Task states with elapsed times), the current-task
panel (live tool invocations with their frozen states and streamed output tail), and a
collapsed event timeline fed by the FR-TUI-007 bridge. `p` pauses the run (`paused`,
resumable); `Ctrl+C` opens the interruption confirmation (run → `cancelled` on confirm);
`Enter` opens the full task detail overlay; `t` jumps to the tool-calls screen filtered
to the current task. A run in `awaiting_approval` renders the blocking Approval reference
prominently in the current-task panel with a key to raise the permission prompt if it was
deferred. Terminal outcomes replace the live panels with a summary (outcome state,
duration, files changed, tokens, cost) and offer `g d` diff review — the UC-01 close.

## Tool call

```text
+------------------------------------------------------------------------------+
| andromeda | tool call  01JZXA | terminal.run | state: executing | 00:14      |
+------------------------------------------------------------------------------+
| Invocation ----------------------------------------------------------------- |
| tool: terminal.run v1.3.0 (builtin, trusted)   task: 1  run: 01JZX7          |
| permission: execute (workspace scope) - granted 14:02 via allow_for_session  |
| timeout: 120s   sandbox: process-level                                       |
|                                                                              |
| Input ---------------------------------------------------------------------- |
| { "command": "go test ./internal/fetch/...",                                 |
|   "cwd": ".", "capture_limit_kb": 512 }                                      |
|                                                                              |
| *Output (live) ------------------------------------------------------------- |
| === RUN   TestRetryBackoff                                                   |
| === RUN   TestRetryJitter                                                    |
| ok   internal/fetch  0.41s                                                   |
| ... streaming ...                                                            |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
| [c]cancel invocation [Enter]full input/output [l]related logs [Esc]back      |
+------------------------------------------------------------------------------+
```

Inspects one Tool Invocation: identity (name, version, origin, trust level from the tool
declaration — Volume 6), the permission that authorized it with its scope and decision
(PRD-006 attribution), effective timeout and sandbox level, schema-valid input as
formatted JSON, and the live output stream with capture-limit truncation markers.
`c` requests cooperative cancellation (ToolPort `Cancel`; invocation reaches `cancelled`
through its stream). Terminal states render the Tool Result summary — `succeeded` shows
output and artifacts; `failed`, `denied`, and `timed_out` show the E-TOOL-family envelope
with its recommended action. The list mode of this screen (reached by `g t`) shows all
invocations of the session with state, duration, and tool name, filterable per chapter 11.

## Permission prompt

```text
+------------------------------------------------------------------------------+
| andromeda | session myproject | main | anthropic:sonnet | run: awaiting_ap.. |
+------------------------------------------------------------------------------+
|                                                                              |
|   +--------------------- Permission required ---------------------------+    |
|   |                                                                     |    |
|   |  ! The agent requests: git_mutation (with network)                  |    |
|   |                                                                     |    |
|   |  command   git push origin main                                     |    |
|   |  scope     repository ~/code/myproject (remote: origin)             |    |
|   |  requested by  task 3 "Update CHANGELOG" - tool git.exec (push)     |    |
|   |  risk      pushes 2 commits to a shared remote; not undoable        |    |
|   |            locally                                                  |    |
|   |                                                                     |    |
|   |  [1] allow once                                                     |    |
|   |  [2] allow for this session                                         |    |
|   |  [3] allow for this workspace                                       |    |
|   |  [4] always allow (policy)                                          |    |
|   |  [5] deny once                                                      |    |
|   |  [6] always deny                                                    |    |
|   |  [7] ask every time                                                 |    |
|   |                                                                     |    |
|   |  expires in 04:58 - unanswered requests are denied                  |    |
|   +---------------------------------------------------------------------+    |
|                                                                              |
| [1-7]decide  [Enter]details  approval 01JZXB requested                       |
+------------------------------------------------------------------------------+
```

The approval modal (Approval states, Volume 2) traps focus per FR-TUI-003 and discards
typed-ahead input. It states the requested permission (frozen enum names from Volume 9;
a push binds `git_mutation` plus `network` — Volume 11 remote-operation rules), the
precise subject (command, path, or resource), the requesting task and tool
(attribution chain), and a risk statement from the tool's declaration. Digits map to the
frozen decision names, in Volume 9's decision-table order: `1` `allow_once`, `2`
`allow_for_session`, `3` `allow_for_workspace`, `4` `always_allow_policy`, `5`
`deny_once`, `6` `always_deny`, `7` `ask_every_time`. Where Volume 9's decision
constraints refuse `always_allow_policy` (`system_modification` requests, or selectors
unbounded on all resource qualifiers for `execute`/`write`/`credential_access`), its
entry is not rendered and its digit is inert; the remaining entries keep their digits.
`always_allow_policy` and `always_deny` persist a selector no broader than the subject
the modal displays (Volume 9 decision constraint). The expiry
countdown reflects the Approval's `expires_at`; expiry records `expired` and the subject
does not proceed. Decisions are explicit: `Esc` merely lowers the modal (status bar keeps
an "approval pending" badge and the run stays `awaiting_approval`). The danger marker
`!` and Danger styling follow chapter 08 rules — meaning survives every tier.

## Diff

```text
+------------------------------------------------------------------------------+
| andromeda | diff  patch 01JZXC | proposed | 2 files  +24 -3                  |
+------------------------------------------------------------------------------+
| Files ------------------------ | *Hunks -------------------------------------|
|                                |                                             |
| > M internal/fetch/client.go   | @@ -41,6 +41,18 @@ func (c *Client) Get     |
|   A internal/fetch/retry.go    |  41   ctx, cancel := context.With...        |
|                                |  42 + retried, err := retry.Do(ctx,         |
|                                |  43 +   c.policy, func() error {            |
|                                |  44 +     return c.do(req)                  |
|                                |  45 +   })                                  |
|                                |  46 + if err != nil {                       |
|                                |  47 +   return nil, err                     |
|                                |  48 + }                                     |
|                                |  49   defer cancel()                        |
|                                |                                             |
|                                | @@ -88,3 +100,9 @@ func (c *Client) Post    |
|                                |  ...                                        |
|--------------------------------+---------------------------------------------|
| internal/fetch/client.go: +12 -3 | hunk 1/3                                  |
|                                                                              |
|                                                                              |
| [a]apply patch [x]reject [o]open file [n/N]next/prev hunk [Esc]back          |
+------------------------------------------------------------------------------+
```

Reviews a Patch (recorded status vocabulary: `proposed`, `applied`, `rejected`,
`reverted`): file list with change markers (M/A/D/R), hunk viewport with added/removed
line styling (Primary-family for additions, Danger-family for removals, always with the
`+`/`-` textual markers so the no-color tier reads identically), and a per-file summary
line. `a` applies through GitPort `ApplyPatch` — a side-effecting action that follows the
permission model (`write`/`git_mutation` per the operation, Volume 9) and renders its
result per file; `x` records `rejected` with a confirmation. Applied patches remain
inspectable with an "applied" banner and a revert affordance (which itself creates a new
reviewable operation — no silent destructive operations, Volume 11). Word-level
highlighting and large-diff virtualization follow chapter 11.

## Git

```text
+------------------------------------------------------------------------------+
| andromeda | git  myproject | branch main <- origin/main | +2 ~1 | clean: no  |
+------------------------------------------------------------------------------+
| *Changes ------------------------------- | Preview --------------------------|
|                                          |                                   |
| staged:                                  | internal/fetch/client.go          |
| > M internal/fetch/client.go             | @@ -41,6 +41,18 @@                |
|   A internal/fetch/retry.go              |  + retried, err := retry.Do(...   |
| unstaged:                                |  + ...                            |
|   M README.md                            |                                   |
| untracked:                               |                                   |
|   internal/fetch/retry_test.go           |                                   |
|                                          |                                   |
|------------------------------------------+-----------------------------------|
| History -------------------------------------------------------------------- |
| 01c3f2a feat(fetch): add retry policy scaffolding        maia   2h ago       |
| 9ab01ee fix(cli): correct exit code on timeout           maia   1d ago       |
| 77e410b chore: bump linters                              bot    2d ago       |
|                                                                              |
|                                                                              |
| [s]stage [u]unstage [c]commit [b]branches [P]push [f]fetch [Enter]show       |
+------------------------------------------------------------------------------+
```

The Git Engine's interactive surface (GitPort; semantics Volume 11): header with branch,
upstream, ahead/behind, and cleanliness; a changes panel grouped as staged / unstaged /
untracked with per-file preview; and recent history. `s`/`u` stage and unstage the
selection; `c` opens the commit overlay (message editor with Conventional Commits
assistance per ADR-015 — commit messages carry change information only); `P` push and
other mutating operations route through the permission model (`git_mutation`) and render
their confirmations before execution. `Enter` on a history row opens the commit detail
(GitPort `Show`). Conflicted states render a dedicated conflicts group with resolution
guidance; destructive operations are never offered without an explicit confirmation
naming the exact refs affected (Volume 11 rules).

## Files

```text
+------------------------------------------------------------------------------+
| andromeda | files  myproject | 4,982 files | index: ready                    |
+------------------------------------------------------------------------------+
| *Tree ---------------------------------- | Preview --------------------------|
|                                          |                                   |
| v internal/                              | internal/fetch/client.go   210 L  |
|   v fetch/                               |                                   |
|     > client.go              M  8.2k     | package fetch                     |
|       retry.go               A  2.1k     |                                   |
|       retry_test.go          ?  3.4k     | import (                          |
|   > cli/                                 |   "context"                       |
|   > tui/                                 |   "net/http"                      |
| > docs/                                  |   "time"                          |
|   go.mod                        1.1k     | )                                 |
|   README.md                  M 12.0k     |                                   |
|                                          | // Client wraps HTTP access ...   |
|                                          |                                   |
|                                          |                                   |
|                                          |                                   |
|                                          |                                   |
|                                          |                                   |
| [Enter]open/collapse [v]view [/]filter [i]reveal ignored [g d]diff           |
+------------------------------------------------------------------------------+
```

A read-focused workspace browser: tree panel with VCS markers (M/A/D/?), sizes, and
ignore-aware listing (`i` toggles revealing ignored entries, marked distinctly), plus a
read-only preview with syntax-neutral rendering (line numbers, truncation markers for
large or binary files per Volume 7 handling rules). It is not an editor (editing is Out
of Scope per Volume 1); `v` opens the full-screen viewer overlay, and "open in $EDITOR"
is offered through the command palette, delegating to the environment. Filtering narrows
the tree per chapter 11 rules; selection state feeds the go-to-diff key when the file has
pending changes. Large trees virtualize; the tree never blocks on indexing — index state
appears in the header (frozen Index states) and preview degrades gracefully while
`building`.

## Context

```text
+------------------------------------------------------------------------------+
| andromeda | context  next request | budget 24.0k tok | used 17.8k (74%)      |
+------------------------------------------------------------------------------+
| Budget [##################----------]  74%  17.8k / 24.0k                    |
+------------------------------------------------------------------------------+
| *Items --------------------------------------------------------------------- |
|                                                                              |
| > P system prompt (profile: default)             1.2k   pinned               |
|   P memory: project conventions                  0.8k   pinned               |
|     transcript: last 6 turns                     6.4k   recency              |
|     file internal/fetch/client.go                3.9k   relevance 0.92       |
|     file internal/fetch/retry.go                 2.0k   relevance 0.87       |
|     tool result: go test output (truncated)      1.9k   recency              |
|     index hits: "backoff" (3 chunks)             1.6k   relevance 0.71       |
|                                                                              |
| Detail --------------------------------------------------------------------- |
| file internal/fetch/client.go - included by relevance (query match)          |
| source: workspace file, read 14:05:02, provenance: indexed content           |
| actions: pin, exclude, view full content                                     |
|                                                                              |
|                                                                              |
| [p]pin [x]exclude [Enter]inspect [r]refresh preview [Esc]back                |
+------------------------------------------------------------------------------+
```

A transparency surface over the Context Manager (Volume 7): the budget bar (token budget
for the next request against the active model's window), the ranked item list with per-
item token counts and inclusion reasons (pinned, recency, relevance score), and a detail
panel with provenance and source attribution. `p` pins (user pinning overrides ranking,
Volume 7), `x` excludes — both take effect for subsequent assembly and render
immediately in the preview; `r` recomputes the preview without sending anything. The
screen is read-mostly: it never edits content, only inclusion. When assembly data for a
model without token counting is estimated, counts render with a `~` estimation marker
(Volume 7 estimation strategy). Pinned items render the `P` marker so inclusion class
survives the no-color tier.

## Memory

```text
+------------------------------------------------------------------------------+
| andromeda | memory  workspace layer | 62 records | filter: none              |
+------------------------------------------------------------------------------+
| [1]session  [2]workspace  [3]long-term                                       |
+------------------------------------------------------------------------------+
| *Records ------------------------------------------------------------------- |
|                                                                              |
| > active   convention  "Tests live next to source files"        2026-07-01   |
|   active   decision    "Use exponential backoff for retries"    2026-07-11   |
|   active   preference  "Commit messages in English"             2026-06-20   |
|   archived note        "Legacy fetcher removed in v0.3"         2026-05-02   |
|                                                                              |
|                                                                              |
| Detail --------------------------------------------------------------------- |
| decision - "Use exponential backoff for retries"                             |
| status: active | layer: workspace | id: 01JZM3...                            |
| provenance: run 01JZX7, task 1, confirmed by user 14:06                      |
| retention: keep until superseded | trust: user-confirmed                     |
|                                                                              |
|                                                                              |
|                                                                              |
| [e]edit [d]delete [a]archive [n]new [Enter]inspect [/]search                 |
+------------------------------------------------------------------------------+
```

Browses Memory Records by layer tab (session, workspace, long-term — Volume 7 layers)
with the recorded status vocabulary (`active`, `archived`, `expired`, `deleted`), kind,
summary, and date. The detail panel shows full provenance (originating run/task, trust
level, confirmation), retention policy, and identity. `d` deletes through MemoryStorePort
`Delete` with a confirmation that states the cascade consequences (audit precedence:
deletion is itself recorded — Volume 7/9); `a` archives; `e` and `n` open the record
editor overlay for user-curated knowledge. Search spans summary and content lexically
(IndexerPort-backed where available). This screen is the user-facing guarantee that
memory is inspectable and correctable — nothing the agent remembers is hidden from the
operator (PRD-006).

## Logs

```text
+------------------------------------------------------------------------------+
| andromeda | logs  follow: on | level >= info | source: all | 1,204 lines     |
+------------------------------------------------------------------------------+
| *Lines --------------------------------------------------------------------- |
|                                                                              |
| 14:05:01.221 INFO  tool.terminal  exec started cmd=go test id=01JZXA         |
| 14:05:13.640 INFO  tool.terminal  exec completed exit=0 dur=12.4s            |
| 14:05:13.652 INFO  run.engine     task 1 completed dur=2m58s                 |
| 14:05:13.688 INFO  provider.http  request start model=sonnet stream=true     |
| 14:05:14.001 WARN  provider.http  retry 1/3 status=529 backoff=800ms         |
| 14:05:15.113 INFO  provider.http  stream complete tokens_out=412             |
| 14:05:15.120 INFO  memory.store   ingest 1 record layer=workspace            |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
|                                                                              |
| Detail: 14:05:14.001 WARN provider.http retry ... corr=01JZX7 span=8f21      |
|                                                                              |
| [f]follow [l]level [s]source [Enter]expand [c]copy line [/]search            |
+------------------------------------------------------------------------------+
```

A live tail over the structured log pipeline (Volume 10): level and source filters,
follow mode with the same scroll-up/"new lines below" behavior as the transcript, and a
detail line showing the selected record's correlation identifiers — the join keys to
runs, traces, and events (SM-13). `Enter` expands the full structured record as
formatted key/value content; `c` copies the raw line (clipboard semantics chapter 11).
Log content arrives pre-redacted (redaction is the pipeline's, not the TUI's); the TUI
adds no formatting that could re-join redacted fields. Levels style via the chapter 08
roles (WARN via Secondary-family with marker, ERROR via Danger with marker). Filters and
follow state are presentation-only and reset per session.

## Costs and tokens

```text
+------------------------------------------------------------------------------+
| andromeda | costs  session 01JZX6 | total 31.2k tok | $1.08 | budget 60%     |
+------------------------------------------------------------------------------+
| Session budget [############--------]  $1.08 / $1.80 (policy: warn 80%)      |
+------------------------------------------------------------------------------+
| *By run -------------------------------------------------------------------- |
|                                                                              |
| > run 01JZX7  "retry logic"      12.4k in / 3.1k out   $0.42   running       |
|   run 01JZX5  "explain module"    9.8k in / 2.2k out   $0.31   completed     |
|   run 01JZX2  "fix lint"         2.9k in / 0.8k out   $0.09   completed      |
|                                                                              |
| By provider / model -------------------------------------------------------- |
| anthropic sonnet     24.1k in / 5.6k out    $0.94    3 runs                  |
| ollama    qwen3      5.0k in / 0.5k out     $0.00    1 run                   |
|                                                                              |
| Detail --------------------------------------------------------------------- |
| run 01JZX7: 6 requests | cache reads 8.2k | est. next request 17.8k tok      |
| cost source: provider-reported (anthropic), none (ollama)                    |
|                                                                              |
|                                                                              |
| [Enter]run detail [g c]context [e]export [Esc]back                           |
+------------------------------------------------------------------------------+
```

Accounting transparency (PRD-006): session totals in the header, the budget bar with the
active policy threshold (budget policy semantics are Volume 4/10's; the TUI renders
state), per-run rows with token in/out, cost, and frozen Run state, per-provider/model
aggregation, and a detail panel showing request counts and the cost source — provider-
reported versus locally computed versus unavailable (`$0.00` with source `none` for local
providers, never invented prices; Volume 5 accounting rules). `e` exports the visible
aggregation through the CLI-equivalent export path (chapter 05 family). Values update
live from cost-record events; discrepancies resolve by authoritative read (FR-TUI-007
authority rule). Estimation markers (`~`, `est.`) distinguish estimates from reported
figures at every tier.

## Requirements

### FR-UX-043 — Splash and identity surfaces

- Type: Functional
- Status: Draft
- Priority: P2
- Phase: MVP
- Source: Provided
- Owner: TUI (Volume 8)
- Affected components: TUI
- Dependencies: ADR-026; FR-TUI-001 (start sequence); `tui.splash` key (chapter 07)
- Related risks: RISK-TUI-003

#### Description

The splash screen renders the official marks exactly as fixed by ADR-026: the ASCII cat
mascot with four-pointed star accents, the lowercase wordmark, the tagline "Your terminal
companion for shipping great software.", and the installed version with channel. Policy
via `tui.splash`: `auto` (default) shows it on first run and after version changes for at
most 2 s or until any key; `always` on every start; `never` skips it. The splash never
delays readiness: workspace discovery and Runtime startup proceed concurrently, and any
key dismisses immediately. First-run mode appends created-configuration information. The
marks are the only sanctioned rendering of the mascot in the TUI (no per-screen mascot
variants), keeping the identity stable and golden-testable.

#### Motivation

The brand marks are specification-fixed assets (ADR-026); the splash is their canonical
TUI surface. Bounding its duration and interruption keeps identity from taxing startup
(SM-06(b)).

#### Actors

Users starting the TUI; the shell start sequence.

#### Preconditions

Shell started per FR-TUI-001.

#### Main flow

1. Start sequence evaluates `tui.splash` policy.
2. Splash renders; discovery proceeds concurrently.
3. Key or timeout dismisses; the start screen follows.

#### Alternative flows

- `never`: direct to start screen.
- First run: extended footer content; dismissal identical.

#### Edge cases

- Terminal below 80×24: the splash centers what fits (mascot first to be dropped, tagline
  last) per compact-class rules; below 40×10 the size notice takes precedence.
- Version-change display when downgrading: shown identically (any version difference).

#### Inputs

`tui.splash`; version/channel; first-run marker.

#### Outputs

Rendered splash; dismissal into the start screen.

#### States

Not applicable — transient presentation.

#### Errors

None minted; rendering failures follow E-TUI-006.

#### Constraints

Maximum 2 s auto-display in `auto`; the mascot art is pure ASCII by specification (no
Unicode dependency).

#### Security

None.

#### Observability

Splash display is not separately evented (start timing lives in `tui.shell.started`).

#### Performance

Zero added latency to readiness (concurrent discovery); dismissal handled within the
SM-07 budget.

#### Compatibility

Renders at every tier including `none` (plain text marks).

#### Acceptance criteria

- Given a first run at 80×24, when the TUI starts, then the splash matches its golden
  frame including mascot, wordmark, tagline, version, and first-run footer.
- Given `tui.splash = "auto"` and an unchanged version, when the TUI starts a second
  time, then no splash renders.
- Given the splash visible, when any key is pressed at t < 2 s, then the start screen
  renders immediately and the keypress is consumed (not forwarded).
- Negative case: given `tui.splash = "never"`, when the TUI starts, then no frame
  contains the tagline.

#### Verification method

Golden frames (splash variants); scripted start-sequence tests over the policy matrix;
timing assertions in the PTY harness.

#### Traceability

PRD-008; ADR-026; MVP item 2; SM-06.

### FR-TUI-009 — Core screen inventory and content contract

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: TUI (Volume 8)
- Affected components: TUI; Runtime (read surfaces)
- Dependencies: FR-TUI-001–FR-TUI-007; chapters 10–11 for platform screens and patterns
- Related risks: RISK-TUI-003

#### Description

The MVP TUI ships all fourteen core screens of this chapter — splash, workspace
selection, session, plan, execution, tool call, permission prompt, diff, git, files,
context, memory, logs, costs — each with at least the panels, content elements, state
renderings, and actions its wireframe section specifies. Screens render frozen state
names verbatim (Volume 2, chapter 09); every side-effecting action offered by a screen
routes through the Runtime API and the permission model (no TUI-local mutations); every
screen defines behavior for its empty, loading, offline, and degraded conditions per
chapter 11 patterns. Removing a screen, panel, or listed action from the MVP set requires
the Volume 0 change procedure.

#### Motivation

MVP item 2 commits to an interactive surface with specific capabilities; an enumerated
content contract is what makes "functional TUI" verifiable rather than impressionistic,
and it anchors the golden-frame suite's scope.

#### Actors

TUI screens; Runtime read/write surfaces; the golden suite.

#### Preconditions

Shell running; workspace open (except splash and workspace selection).

#### Main flow

1. A screen activates via navigation.
2. It reads authoritative state, subscribes per FR-TUI-007, and renders its contracted
   content.
3. Contracted actions dispatch through the Runtime API.

#### Alternative flows

- Content source unavailable: the screen's degraded state renders (E-TUI-007 pattern)
  while navigation remains functional.

#### Edge cases

- Entities in states a screen has no bespoke rendering for (future enum additions are
  prohibited, but foreign records from newer schema versions can appear after downgrade
  refusal per ADR-029): render the state name verbatim with default styling rather than
  crashing or hiding the row.
- Concurrent CLI mutations (same workspace): screens converge on authoritative reads per
  FR-TUI-007's authority rule.

#### Inputs

Runtime reads; event subscriptions; user actions.

#### Outputs

Contracted renderings; Runtime API calls; screen events.

#### States

Renders every frozen vocabulary referenced by its wireframes; owns none.

#### Errors

Screen-level failures use E-TUI-006/E-TUI-007; domain errors render their owning
families' envelopes (E-TOOL, E-GIT, E-PROV, …) with code and recommended action.

#### Constraints

No screen may offer a side-effecting action that bypasses PermissionPort mediation
(SM-16(b) applies to the TUI surface).

#### Security

Permission prompt content requirements (precise subject, scope, attribution) are part of
this contract; approval rendering completeness is release-gating.

#### Observability

`tui.screen.changed` covers activation; screens add no bespoke events at MVP.

#### Performance

Screen activation renders first contentful frame within the SM-07 budget from cached
authoritative state, with fresh reads applied asynchronously.

#### Compatibility

All screens function at every size ≥ 80×24 and every tier (NFR-TUI-001, FR-UX-041).

#### Acceptance criteria

- Given the golden suite, when it enumerates screens, then all fourteen exist with their
  contracted panels and actions present at 80×24.
- Given each contracted side-effecting action, when triggered against an instrumented
  Runtime, then a PermissionPort evaluation precedes execution (mediation test).
- Given a screen whose data source is failing (fault injection), when activated, then
  its degraded state renders and `g`-navigation still works.
- Negative case: given the MVP build, when the screen registry is compared to this
  chapter, then no contracted screen is absent and no unspecified side-effecting action
  exists.

#### Verification method

Golden-frame enumeration suite; permission-mediation instrumentation tests (SM-16(b));
fault-injection screen matrix in Volume 13's TUI suite.

#### Traceability

PRD-006, PRD-008; MVP item 2; SM-16; FR-TUI-001–007; Volume 2 chapter 09 vocabularies.

## Risks

### RISK-TUI-003 — Wireframe and implementation drift

- Category: Process
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Wireframes are normative for structure only (this chapter's preamble); golden frames become the pixel authority at implementation and are updated through reviewed commits; FR-TUI-009's enumeration test pins the screen/panel/action inventory to this chapter; divergences that change the inventory require a specification change first
- Detection: FR-TUI-009 enumeration suite failures; consolidation and release audits comparing shipped screens to the chapter; review checklist item for TUI PRs
- Owner: TUI (Volume 8)
- Status: Open

ASCII wireframes cannot capture every rendering decision, and implementations
legitimately refine spacing and truncation. The risk is inventory drift — a panel or
action quietly disappearing, or an unspecified side-effecting action appearing. Binding
the inventory (not the pixels) to the specification keeps refinement free and drift
detectable.
