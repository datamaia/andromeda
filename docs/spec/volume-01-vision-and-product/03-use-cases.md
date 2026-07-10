# 03 — Use Cases

This chapter presents fourteen concrete use cases that the product objectives of
[chapter 04](04-goals-non-goals-principles.md) exist to enable. They are narrative scenarios,
not functional requirements: the functional requirements that realize them are minted by the
owning volumes (Volumes 4–14) and traced back to objectives during consolidation.

Conventions: use-case identifiers `UC-NN` are local to this volume. "User" means a human
operator at the CLI or TUI; component names are those of the Volume 0 glossary. Each use case
lists its primary actors, preconditions, main flow, and outcome, and names the Jobs to Be Done
([chapter 02](02-users-personas-jtbd.md)) it serves.

## Interactive engineering

### UC-01 — Implement a feature from a specification

**Serves:** JTBD-1, JTBD-3. **Personas:** Elena, Marcus.

- **Actors:** User; agent; Planner; Execution Engine; Tool Runtime; Permission Manager.
- **Preconditions:** A workspace is open on a Git repository; a feature specification exists as
  a file in the repository; a provider with tool-calling capability is configured.
- **Flow:**
  1. The user starts a session in the TUI and asks the agent to implement the feature described
     in the specification file.
  2. The Context Manager assembles relevant context: the specification file, related source
     files located through the Index, and applicable workspace memory.
  3. The Planner produces an inspectable plan (read these files, modify those, add tests, run
     the test suite). The user reviews the plan and approves it.
  4. The Execution Engine dispatches tasks. File edits are applied through filesystem tools;
     each write inside the workspace is covered by a write permission granted for the session.
  5. The agent runs the test suite through the Terminal Engine; failures feed back into plan
     revision.
  6. The agent presents a consolidated patch. The user reviews the diff hunk by hunk, requests
     one adjustment, and accepts the result.
- **Outcome:** The feature is implemented and verified by tests; the run record contains the
  plan, every file read and modified, every command executed, tokens and cost, and the
  permission decisions. Nothing was changed outside the granted scope.

### UC-02 — Fix a failing test

**Serves:** JTBD-2. **Personas:** Elena, Marcus, Jonas.

- **Actors:** User; agent; Terminal Engine; Git Engine.
- **Preconditions:** Workspace open; the project's test command is configured or discoverable;
  at least one test fails.
- **Flow:**
  1. The user invokes a fix-oriented run, passing the failing test's identifier (or letting the
     agent run the suite to find failures).
  2. The agent executes the test through the Terminal Engine, captures the failure output, and
     correlates it with recent changes via the Git Engine (diff against the last passing state,
     when available).
  3. The agent forms a hypothesis, reads the implicated source files, and proposes a minimal
     patch, distinguishing "fix the code" from "fix the test" and stating which it chose and why.
  4. The user approves the patch; the agent re-runs the failing test, then the affected test
     subset.
- **Outcome:** The test passes; the patch is confined to the diagnosed fault; the justification
  and evidence (test output before and after) are part of the run record.

### UC-03 — Review a diff

**Serves:** JTBD-3, JTBD-7. **Personas:** Elena, Priya.

- **Actors:** User; agent; Git Engine.
- **Preconditions:** Workspace open on a repository with uncommitted changes, or a branch/range
  to compare.
- **Flow:**
  1. The user asks for a review of the working-tree diff (or a named revision range).
  2. The agent obtains the diff through the Git Engine — read-only; no Git mutation permission
     is requested.
  3. The agent analyzes the changes for defects, regressions, missing tests, and deviations
     from conventions recorded in workspace memory, citing file and line for each finding.
  4. Findings are presented as a structured list ordered by severity; the user navigates
     finding-by-finding in the TUI, accepting or dismissing each.
- **Outcome:** The user has an actionable review with evidence. No files were modified; the run
  record shows a read-only run, which is itself auditable evidence of the review having
  happened.

### UC-04 — Refactor across files

**Serves:** JTBD-1, JTBD-3. **Personas:** Marcus, Elena.

- **Actors:** User; agent; Indexing Engine; Execution Engine.
- **Preconditions:** Workspace open; the workspace Index is built (or is built on demand).
- **Flow:**
  1. The user requests a cross-cutting change — for example, renaming a concept and adjusting
     every call site, or extracting a shared module from duplicated code.
  2. The agent queries the Index to enumerate affected files and symbols, and presents the
     blast radius (file count, sites per file) before editing anything.
  3. The user approves; the Execution Engine applies the edits as a series of tasks with
     progress visible per file.
  4. The agent runs the build and tests; residual breakage is fixed iteratively within the
     approved scope.
  5. The result is presented as one reviewable patch spanning all affected files.
- **Outcome:** A consistent multi-file change with a single audit trail; files outside the
  enumerated scope are untouched, and the run record proves it.

### UC-05 — Onboard an unfamiliar codebase

**Serves:** JTBD-1, JTBD-10. **Personas:** Elena, Marcus.

- **Actors:** User; agent; Indexing Engine; Memory Manager.
- **Preconditions:** A repository the user does not know; a configured provider (cloud or
  local).
- **Flow:**
  1. The user opens the repository as a new workspace; Andromeda initializes `.andromeda/` and
     offers to index the tree.
  2. The user asks orientation questions: where is the entry point, how does deployment work,
     which tests cover module X.
  3. The agent answers from the Index and targeted file reads, citing paths for every claim.
  4. Durable findings the user confirms (build commands, conventions, architecture notes) are
     stored as workspace memory records for future sessions.
- **Outcome:** The user reaches productive work quickly, and the workspace has accumulated
  memory that makes every subsequent session start smarter — locally, in the repository's own
  Andromeda state.

## Workflows and automation

### UC-06 — Generate and run a specification-driven workflow

**Serves:** JTBD-12, JTBD-1. **Personas:** Marcus, Jonas.

- **Actors:** User; Workflow Engine; multiple agents; Permission Manager.
- **Preconditions:** Workspace open; a workflow definition for specification-driven development
  (intake → requirements → planning → task decomposition → implementation → validation →
  review) is available, either built in or defined by the team.
- **Flow:**
  1. The user starts the workflow with an intake document (a feature request).
  2. The Workflow Engine advances through declared states; each state binds agents, tools,
     permissions, and exit criteria.
  3. At declared approval gates (requirements sign-off, plan sign-off, final review), the
     workflow pauses and requests an explicit human approval; each approval is recorded.
  4. Implementation tasks execute as in UC-01, but under the workflow's narrower permission
     profile.
  5. The workflow completes, emitting its artifacts: the requirements document, the plan, the
     patch, the test evidence, and the full state-transition history.
- **Outcome:** A repeatable, resumable, audited engineering process. Interrupting the workflow
  at any state and resuming later (see UC-11) yields the same process with its history intact.

### UC-07 — Run non-interactively in CI

**Serves:** JTBD-5, JTBD-2. **Personas:** Jonas, Marcus.

- **Actors:** CI system (GitHub Actions job); CLI in non-interactive mode; Policy Engine.
- **Preconditions:** A repository checkout in a headless runner without a TTY; provider
  credentials injected via environment variables; a project policy file pre-granting the
  narrow permission set the job needs; JSON output enabled.
- **Flow:**
  1. The CI job invokes the CLI with a task (for example, "update the changelog from merged PRs
     since the last tag"), non-interactive mode, a time budget, and a cost budget.
  2. Andromeda detects the absence of a TTY, confirms non-interactive mode, and resolves every
     permission from pre-approved policy — no prompt is ever emitted; any permission not
     pre-granted is denied and recorded.
  3. The run executes; structured events stream to the job log; the final result — patch,
     summary, token and cost accounting — is written as JSON to stdout.
  4. The job parses the JSON, opens a pull request with the patch, and archives the run record
     as a build artifact. The CLI exits with code 0; had the run failed, the documented nonzero
     exit code would have failed the job deterministically (exit codes per Volume 0,
     chapter 03).
- **Outcome:** Unattended agent work with the same runtime, permissions, and observability as
  interactive use — plus machine-parseable results and honest exit codes.

### UC-08 — Scripted batch maintenance across repositories

**Serves:** JTBD-5, JTBD-6. **Personas:** Jonas, Elena.

- **Actors:** User's shell script; CLI; Provider Layer.
- **Preconditions:** Several local repository clones; a routing policy that sends mechanical
  tasks to a low-cost model (possibly local) and escalates on failure.
- **Flow:**
  1. A script loops over repositories, invoking the CLI non-interactively in each: apply a
     mechanical migration (for example, a configuration format change), run tests, produce a
     patch file.
  2. Per-repository runs are independent sessions with independent records; the script collects
     JSON results and patch artifacts.
  3. In two repositories the low-cost model fails validation; per policy, the run either
     escalates to the configured stronger model — reporting the model change explicitly in the
     run record — or exits with the documented failure code for the script to triage.
- **Outcome:** Fleet-scale maintenance with per-repository auditability, cost control through
  routing, and no hidden model substitution.

## Local and offline operation

### UC-09 — Work fully offline with a local model

**Serves:** JTBD-4, JTBD-10. **Personas:** Amara, Priya.

- **Actors:** User; agent; local provider (local inference server); all local engines.
- **Preconditions:** A local inference server is running with a model that declares
  tool-calling capability; the workspace, its Index, and its memory are local; the machine has
  no network connectivity.
- **Flow:**
  1. The user opens the workspace. Startup, workspace discovery, and session creation involve
     no network access.
  2. The user runs an implement-and-test task as in UC-01. Planning, file edits, terminal
     commands, Git operations, diff review, and log inspection all execute locally.
  3. Andromeda displays the active provider and model, and the capability set actually declared
     by the local model; a capability the model lacks (for example, vision) is reported as
     unavailable rather than simulated.
  4. The user reviews the patch, runs the tests, and commits with the Git Engine — still
     offline.
- **Outcome:** The complete core workflow — open workspace, query local memory, index files,
  run local tools, local Git, terminal, offline-capable workflows, create patches, review
  diffs, run tests, view logs — succeeded with zero network egress, exactly as the local-first
  guarantee in [chapter 04](04-goals-non-goals-principles.md) requires.

### UC-10 — Switch provider mid-work without losing anything

**Serves:** JTBD-6. **Personas:** Elena, Amara.

- **Actors:** User; Provider Layer; Context Manager.
- **Preconditions:** Two providers configured (for example, one cloud, one local); an active
  session with history, memory, and granted permissions.
- **Flow:**
  1. Mid-session, the cloud provider begins failing (outage or rate limiting). Andromeda
     surfaces the provider errors with their category and retry behavior — no silent retry
     storm.
  2. The user switches the session to the local provider (or an automatic fallback policy does,
     announcing the change explicitly in the session and the run record).
  3. Capability differences are reported: the session continues with the capabilities the new
     model declares; workflow steps requiring a missing mandatory capability fail with a
     precise error instead of degrading silently.
  4. Session history, memory, permissions, and workspace state carry over unchanged; only the
     inference backend changed.
- **Outcome:** The provider decision is reversible in-flight. Every provider or model change is
  visible to the user and recorded — never silent.

## Extension and platform

### UC-11 — Recover an interrupted session

**Serves:** JTBD-9. **Personas:** Elena, Jonas.

- **Actors:** User; Persistence Layer; Workflow Engine (when a workflow was active).
- **Preconditions:** A session (or workflow run) was interrupted — process killed, machine
  rebooted, SSH connection dropped — after its state had been persisted incrementally.
- **Flow:**
  1. The user relaunches Andromeda in the same workspace. The TUI's recovery surface lists
     resumable sessions with their last persisted state and timestamp.
  2. The user selects the interrupted session. Andromeda restores conversation history, plan
     state, task states, permission grants with their original scopes, and cost accounting.
  3. A task that was mid-execution at the interruption is reported as interrupted — not
     assumed completed — and is re-planned or re-run under the same permissions after user
     confirmation.
  4. Work continues from the restored state.
- **Outcome:** No loss of persisted work. The session record documents the interruption and
  the recovery as events, so even the failure is auditable.

### UC-12 — Extend Andromeda with a plugin and an MCP server

**Serves:** JTBD-8, JTBD-11. **Personas:** Tomás, Marcus.

- **Actors:** Extension author; plugin runtime; MCP Runtime; Permission Manager.
- **Preconditions:** The Extension SDK and its documentation are installed; the author has an
  internal service (ticket system) to integrate.
- **Flow:**
  1. Tomás scaffolds a plugin with the Extension SDK; the plugin declares a tool ("query
     tickets") with input/output schemas, permission requirements (external service access,
     scoped to one domain), timeouts, and version.
  2. He tests it against the SDK's fixtures and the test provider — no live model needed — then
     installs it into his workspace; Andromeda records the plugin's identity, version, and
     declared permissions at installation.
  3. Separately, his team already runs an MCP server for their build system. He registers it;
     Andromeda connects via a documented MCP protocol version, discovers its tools, and maps
     them into the Tool Runtime under the same permission model as native tools.
  4. In a session, the agent uses both — the plugin tool and the MCP tools — with each
     invocation individually permissioned, traced, and attributed to its origin.
- **Outcome:** Organizational capabilities become first-class agent tools without forking
  Andromeda. Third-party tools obey the same contracts, permissions, and observability as
  built-in ones — tool-first citizenship, regardless of origin.

## Accountability

### UC-13 — Audit what an agent did

**Serves:** JTBD-7. **Personas:** Priya, Marcus.

- **Actors:** Auditor (may be the user, a security reviewer, or a script); Audit Log;
  Observability.
- **Preconditions:** One or more completed runs exist; audit records and events were persisted
  (the default).
- **Flow:**
  1. The auditor queries the run history for a time window and workspace: which runs executed,
     under which agent profiles, providers, and models.
  2. For a selected run, the auditor retrieves: the plan; every tool invocation with inputs
     summary and result status; every file read and modified with the patches applied; every
     command executed; every permission requested, granted, and denied, and by whom; token and
     cost records; errors, retries, fallbacks, and any model changes; and security-relevant
     events.
  3. The auditor exports the correlated record (events, trace, audit records share correlation
     IDs) as a machine-readable artifact for the compliance archive.
  4. A spot check reconstructs one file's change history: each modification maps to a specific
     run, task, tool invocation, and permission decision.
- **Outcome:** "What did the agent do, and who allowed it?" is answered from records, not
  memory. A run that had hypothetically attempted an unpermitted action would appear in the
  same records as a denied request and a security event.

### UC-14 — Deny a dangerous action and continue safely

**Serves:** JTBD-3, JTBD-7. **Personas:** Priya, Elena.

- **Actors:** User; agent; Permission Manager; Audit Log.
- **Preconditions:** An interactive session with default (conservative) permission policy.
- **Flow:**
  1. While fixing a build, the agent proposes running a command that would delete a directory
     and reinstall dependencies. This class of action — destructive, outside prior grants —
     triggers an explicit permission request.
  2. The TUI presents the request precisely: the exact command, the affected paths, the
     permission class and scope, and the agent's stated reason.
  3. The user denies it once. The denial is recorded; the agent is informed of the denial as a
     structured result, not a silent failure.
  4. The agent adapts: it proposes a narrower alternative (delete only the stale cache
     subdirectory). The user grants that permission for this invocation only.
  5. The narrower command runs; both the denial and the scoped grant appear in the audit trail.
- **Outcome:** Safety without a dead end: denial is a first-class outcome the agent can reason
  about, destructive actions never execute without explicit permission, and the whole
  negotiation is auditable.

## Coverage matrix

| Use case | Jobs served | Primary volume(s) that will formalize the behavior |
|---|---|---|
| UC-01 Implement a feature from a specification | JTBD-1, JTBD-3 | 4, 6, 7, 8 |
| UC-02 Fix a failing test | JTBD-2 | 4, 6, 11 |
| UC-03 Review a diff | JTBD-3, JTBD-7 | 4, 8, 11 |
| UC-04 Refactor across files | JTBD-1, JTBD-3 | 4, 6, 7 |
| UC-05 Onboard an unfamiliar codebase | JTBD-1, JTBD-10 | 7, 8 |
| UC-06 Specification-driven workflow | JTBD-12, JTBD-1 | 4 |
| UC-07 Non-interactive CI run | JTBD-5, JTBD-2 | 8, 9, 10 |
| UC-08 Scripted batch maintenance | JTBD-5, JTBD-6 | 5, 8 |
| UC-09 Fully offline with a local model | JTBD-4, JTBD-10 | 5, 7, 12 |
| UC-10 Provider switch mid-work | JTBD-6 | 5 |
| UC-11 Recover an interrupted session | JTBD-9 | 4, 10, 12 |
| UC-12 Plugin and MCP extension | JTBD-8, JTBD-11 | 6 |
| UC-13 Audit an agent's actions | JTBD-7 | 9, 10 |
| UC-14 Deny a dangerous action | JTBD-3, JTBD-7 | 9 |
