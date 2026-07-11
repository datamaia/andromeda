# 03 — Built-in Tools Catalog

The 20 built-in tools, each with purpose, schema sketch, permissions, and phase. Built-ins
are `builtin` origin, carry the highest trust classification (Volume 2), live in reserved
namespaces (ADR-070), and pass the same pipeline as every other tool. Not all are MVP: the
delivery phasing follows ADR-074 (official public APIs only; integrations land Beta → v1 →
v2). Full input/output schemas are implementation artifacts validated by the conformance
suite; the sketches below fix each tool's contract shape. In sketches, a trailing `?` in a
field name marks an optional field; values describe types and meaning.

## Catalog conventions

- Paths are workspace-relative; escaping the workspace requires an explicit broader `read`/
  `write` scope grant (Volume 9 selectors).
- Every tool obeys the limits, truncation, and spillover rules of chapter 02 (ADR-071).
- Mutating operations declare their permission classes as `required` only when every
  operation of the tool needs them; operation-dependent classes are `optional` and are
  requested only when the arguments demand them (FR-TOOL-005).
- Network-facing tools never accept secret material in arguments: credentials travel as
  Secret Store references resolved by the Authentication Layer (`credential_access` gates
  any tool-visible credential resolution).

### Phase summary

| Phase | Tools |
|---|---|
| MVP | `fs.read`, `fs.write`, `fs.search`, `fs.replace`, `fs.patch`, `fs.diff`, `git.exec`, `terminal.exec` |
| Beta | `process.control`, `http.request`, `docker.control`, `sqlite.query`, `github.request` |
| v1 | `browser.control`, `kubernetes.control`, `gitlab.request`, `jira.request`, `slack.request` |
| v2 | `notion.request`, `linear.request` |

The MVP set realizes MVP minimum items 10–12 (Volume 1): `fs.read` covers reading and
directory listing, `fs.replace` is the edit primitive, `fs.search` is search; `git.exec`
covers status/diff/stage/commit/branch/log through the Git Engine; `terminal.exec` is the
PTY-backed terminal.

## Filesystem tools

### fs.read — MVP

Read a file or list a directory, with byte-range support for large files. Permissions:
`read`. Idempotent; cancellation cooperative.

```json
{
  "input": { "path": "string", "offset?": "integer >= 0", "limit?": "integer >= 1" },
  "output": { "kind": "file | directory", "content?": "string", "entries?": "array of name/kind/size", "size_bytes": "integer", "truncated": "boolean" }
}
```

### fs.write — MVP

Create or overwrite a file with given content; creates parent directories when asked.
Permissions: `write`. Not idempotent (overwrites record distinct File Changes); every write
produces a File Change row with before/after hashes (Volume 2 INV-FCH rules).

```json
{
  "input": { "path": "string", "content": "string", "create_parents?": "boolean", "overwrite?": "boolean, default false" },
  "output": { "path": "string", "op": "create | edit", "bytes_written": "integer", "after_hash": "string sha-256" }
}
```

### fs.search — MVP

Search file contents (literal or regular expression) and file names under a scope, with
ignore-rule awareness. Permissions: `read`. Idempotent.

```json
{
  "input": { "query": "string", "regex?": "boolean", "scope?": "string path", "glob?": "string filter", "max_results?": "integer" },
  "output": { "matches": "array of path/line/column/excerpt", "truncated": "boolean" }
}
```

### fs.replace — MVP

Exact-match or pattern replacement inside one file — the edit primitive. The match must be
unique unless `replace_all`. Permissions: `read`, `write`. Not idempotent.

```json
{
  "input": { "path": "string", "old": "string", "new": "string", "replace_all?": "boolean", "regex?": "boolean" },
  "output": { "path": "string", "replacements": "integer", "after_hash": "string sha-256" }
}
```

### fs.patch — MVP

Apply a unified diff (a Patch entity body) to the workspace atomically — all hunks or none.
Application through this tool moves the Patch `proposed` → `applied` and produces grouped
File Changes (INV-PATCH-03). Permissions: `write`. Not idempotent.

```json
{
  "input": { "patch_id?": "string ulid of a proposed Patch", "diff?": "string unified diff when not referencing a Patch row", "check_only?": "boolean dry-run" },
  "output": { "applied": "boolean", "files": "array of path/op/result", "rejected_hunks": "array" }
}
```

### fs.diff — MVP

Compute a unified diff between two files, a file and provided content, or the working tree
and a given base. Permissions: `read`. Idempotent.

```json
{
  "input": { "left": "string path or content ref", "right": "string path or content ref", "context_lines?": "integer" },
  "output": { "diff": "string unified diff", "binary": "boolean", "truncated": "boolean" }
}
```

## Version control and execution tools

### git.exec — MVP

Run one structured Git operation through the Git Engine (GitPort, ADR-025) — never a raw
shell string. Read operations (`status`, `diff`, `log`, `show`, `branch_list`) require
`read`; mutating operations (`stage`, `unstage`, `commit`, `branch_create`, `branch_switch`,
`apply_patch`, worktree operations) additionally require `git_mutation` (optional class,
requested per operation). No silent destructive operations: destructive Git actions are
Volume 11's to gate with confirmations. Not idempotent for mutations.

```json
{
  "input": { "operation": "enum of GitPort-backed operations", "args": "object per operation schema", "repo?": "string path, defaults to workspace repository" },
  "output": { "operation": "string", "result": "object per operation schema", "warnings": "array of string" }
}
```

### terminal.exec — MVP

Execute a command via the Terminal Engine (TerminalPort) inside a sandbox handle: argv array
(no shell interpolation by default; explicit `shell` mode is policy-gated), PTY or pipe
capture, streamed output (`output_delta` events), input writing for interactive flows.
Permissions: `execute`, `process_spawn`; `network` when the command policy grants egress
(Beta+ enforcement per ADR-021 layers). Not idempotent. The outcome mapping this volume owns
(Volume 2 recorded vocabulary): exit code 0 → `succeeded`; nonzero exit → `failed`;
effective-timeout expiry → `timed_out`; terminated by signal without exit — including
teardown escalation — → `killed`.

```json
{
  "input": { "argv": "array of string", "cwd?": "string path", "env?": "object of names passed per policy, values never persisted", "pty?": "boolean", "timeout_ms?": "integer", "stdin?": "string initial input" },
  "output": { "exit_code?": "integer", "signal?": "string", "outcome": "succeeded | failed | timed_out | killed", "stdout_ref?": "string artifact ulid", "stderr_ref?": "string artifact ulid", "output_tail": "string bounded tail", "truncated": "boolean" }
}
```

### process.control — Beta

Manage long-running processes previously started by Andromeda executions: list, inspect
resource usage, stream output, signal, terminate. Scope is strictly Andromeda-supervised
process trees (PAL Process Trees); arbitrary host processes are out of contract.
Permissions: `process_spawn`. Signal/terminate are not idempotent; queries are.

```json
{
  "input": { "operation": "list | inspect | stream | signal | terminate", "execution_id?": "string", "signal?": "string portable name" },
  "output": { "processes?": "array of execution/pid/state/cpu/memory", "delivered?": "boolean", "outcome?": "string" }
}
```

## Network and data tools

### http.request — Beta

Perform one HTTP request against an allowed host: method, URL, headers, body, redirects,
response capture under output caps. Permissions: `network` (host selectors);
`credential_access` optional, only when a `credential_ref` header source is used — secret
values never appear in arguments or records. Idempotent only for safe methods; declared
`idempotent = false` (the runtime never auto-retries mutating requests; agents retry safe
requests deliberately).

```json
{
  "input": { "method": "string", "url": "string", "headers?": "object", "body?": "string", "credential_ref?": "string secret reference resolved server-side", "timeout_ms?": "integer", "max_redirects?": "integer" },
  "output": { "status": "integer", "headers": "object", "body": "string", "body_ref?": "string artifact ulid when spilled", "truncated": "boolean" }
}
```

### sqlite.query — Beta

Run SQL against a SQLite database file in the workspace (user databases; Andromeda's own
state databases are refused — their integrity is governed by ADR-028/ADR-029, not by agent
SQL). Permissions: `read`; `write` optional for mutating statements, detected by statement
classification before execution. Queries idempotent; mutations not.

```json
{
  "input": { "database": "string path", "sql": "string", "params?": "array", "read_only?": "boolean, default true" },
  "output": { "columns": "array of string", "rows": "array of array", "row_count": "integer", "truncated": "boolean" }
}
```

### browser.control — v1

Drive a browser session for verification and research: navigate, read rendered content,
screenshot (as Artifact), evaluate selectors, basic interaction. The driving mechanism MUST
be an official automation standard (W3C WebDriver or an equivalent documented protocol);
the concrete mechanism and its platform packaging are PENDING VALIDATION (register entry
V6A-OQ-1). Permissions: `network`, `process_spawn`. Not idempotent.

```json
{
  "input": { "operation": "navigate | read | screenshot | click | type | evaluate", "url?": "string", "selector?": "string", "text?": "string", "session?": "string session handle" },
  "output": { "session": "string", "content?": "string", "artifact_ref?": "string screenshot artifact", "result?": "object", "truncated": "boolean" }
}
```

## Platform and service integration tools

Per ADR-074, each tool below uses its service's official, documented public API through the
Authentication Layer; per-service API versions, auth modes, scopes, and rate limits are
PENDING VALIDATION until verified at the tool's phase start (register entry V6A-OQ-2). All
declare `external_service_access` and `network` required, except the container-runtime
tools — `docker.control` and `kubernetes.control` — which declare `container_access` in
place of `external_service_access` (`docker.control` is additionally local by default, with
`network` optional); operation vocabularies below are contract sketches.

### docker.control — Beta

Operate the local Docker Engine through its official API socket: images, containers, logs,
build, compose-level operations where officially supported. Permissions: `container_access`;
`network` optional for remote engine endpoints. Queries idempotent; mutations not.

```json
{
  "input": { "operation": "ps | images | run | stop | rm | logs | build | inspect", "args": "object per operation" },
  "output": { "result": "object per operation", "logs_ref?": "string artifact ulid", "truncated": "boolean" }
}
```

### kubernetes.control — v1

Operate Kubernetes clusters via the official API server using kubeconfig contexts: get,
list, describe, apply, delete, logs, exec (exec additionally gated as `execute`).
Permissions: `container_access`, `network`; `execute` optional for pod exec. Queries
idempotent; mutations not.

```json
{
  "input": { "operation": "get | list | describe | apply | delete | logs | exec", "context?": "string kubeconfig context", "namespace?": "string", "manifest?": "string", "target?": "object kind/name" },
  "output": { "result": "object", "logs_ref?": "string artifact ulid", "truncated": "boolean" }
}
```

### github.request — Beta

Operate GitHub via its official REST API: repositories, issues, pull requests, checks,
releases — the product-side integration semantics (what a PR flow does) are Volume 11's;
this tool is the transport-and-schema surface agents call. Permissions:
`external_service_access`, `network`. Queries idempotent; mutations not.

```json
{
  "input": { "operation": "typed operation from the GitHub surface set", "owner?": "string", "repo?": "string", "params": "object per operation" },
  "output": { "result": "object per operation", "rate_limit": "object remaining/reset when reported", "truncated": "boolean" }
}
```

### gitlab.request — v1

The GitLab counterpart of `github.request` over GitLab's official REST API: projects,
issues, merge requests, pipelines. Permissions: `external_service_access`, `network`.

```json
{
  "input": { "operation": "typed operation from the GitLab surface set", "project?": "string", "params": "object per operation" },
  "output": { "result": "object per operation", "truncated": "boolean" }
}
```

### jira.request — v1

Operate Jira via its official REST API: issues, searches (JQL), transitions, comments.
Permissions: `external_service_access`, `network`.

```json
{
  "input": { "operation": "search | get_issue | create_issue | transition | comment", "params": "object per operation" },
  "output": { "result": "object per operation", "truncated": "boolean" }
}
```

### notion.request — v2

Operate Notion via its official API: pages, databases, blocks, queries. Permissions:
`external_service_access`, `network`.

```json
{
  "input": { "operation": "get_page | query_database | create_page | append_blocks | search", "params": "object per operation" },
  "output": { "result": "object per operation", "truncated": "boolean" }
}
```

### slack.request — v1

Operate Slack via its official Web API: post and read messages in authorized channels,
search where granted. Notification-style posting additionally declares `notifications`.
Permissions: `external_service_access`, `network`; `notifications` optional. Not idempotent
for posts.

```json
{
  "input": { "operation": "post_message | read_channel | search | list_channels", "channel?": "string", "text?": "string", "params?": "object" },
  "output": { "result": "object per operation", "truncated": "boolean" }
}
```

### linear.request — v2

Operate Linear via its official GraphQL API: issues, projects, cycles, comments.
Permissions: `external_service_access`, `network`.

```json
{
  "input": { "operation": "query | create_issue | update_issue | comment", "params": "object per operation" },
  "output": { "result": "object per operation", "truncated": "boolean" }
}
```

## Requirements

### FR-TOOL-007 — Built-in tool catalog and phasing

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Tool Runtime (Volume 6)
- Affected components: Tool Runtime, Git Engine, Terminal Engine, Sandbox Engine, Authentication Layer
- Dependencies: FR-TOOL-001, FR-TOOL-005, FR-TOOL-006; ADR-070, ADR-074; ADR-025
- Related risks: RISK-TOOL-003

#### Description

Andromeda MUST ship the 20 built-in tools of this catalog under the phase assignments of the
phase summary table, each conforming to the tool contract (FR-TOOL-001) with the permissions
and idempotency declarations stated per tool. The MVP set (`fs.read`, `fs.write`,
`fs.search`, `fs.replace`, `fs.patch`, `fs.diff`, `git.exec`, `terminal.exec`) MUST be
functional at MVP exit and realizes MVP minimum items 10–12 (Volume 1). Integration tools
MUST use only official, documented public APIs (ADR-074) with credentials as Secret Store
references; a phase's tools MUST NOT ship earlier without the change procedure, and no
catalog tool may be dropped without one.

#### Motivation

The catalog is the product's hands. Fixing names, shapes, permissions, and phases now lets
prompts, workflows (Volume 4), and documentation build against a stable vocabulary while
implementation lands in order of value density.

#### Actors

Agents invoking; users approving; tool maintainers; external services.

#### Preconditions

Tool Runtime, Sandbox Engine, and per-tool engine dependencies (Git Engine, Terminal Engine,
Authentication Layer) operational at the tool's phase.

#### Main flow

1. Built-ins materialize at startup for the current phase's set.
2. Agents resolve and invoke them through the standard pipeline.
3. Integration tools resolve credentials through the Authentication Layer at execution.

#### Alternative flows

- A service credential is absent: the integration tool fails with its declared tool-local
  unauthenticated condition under E-TOOL-006 and directs the user to the Volume 5 auth flow —
  the tool itself never prompts for secrets.

#### Edge cases

- `sqlite.query` against Andromeda's own state databases is refused by contract.
- `terminal.exec` with `shell` mode requested is policy-gated; the default is argv-only.
- `git.exec` destructive operations defer to Volume 11 confirmation gates; the tool never
  bypasses them.
- Offline operation: all MVP catalog tools are network-free; the offline suite (Volume 1
  guarantees) runs against them.

#### Inputs

Per-tool inputs above.

#### Outputs

Per-tool outputs above; File Change/Patch/Command Execution/Artifact records per Volume 2.

#### States

Invocations per chapter 04; no catalog-level state.

#### Errors

E-TOOL family per chapter 02; tool-local codes per declaration.

#### Constraints

Official APIs only (ADR-074); reserved namespaces (ADR-070); per-service facts PENDING
VALIDATION until verified (V6A-OQ-1, V6A-OQ-2).

#### Security

Each tool's permission surface is stated above and enforced per FR-TOOL-005/006; catalog
tools are the most-invoked code in the product and receive the strictest conformance and
fault-injection coverage (NFR-TOOL-001).

#### Observability

Standard invocation events and spans; integration tools record rate-limit telemetry where
the service reports it.

#### Performance

Catalog tools are subject to the Volume 12 tool-dispatch and filesystem/git operation
budgets.

#### Compatibility

Tool versions follow SemVer independently per tool; schema-breaking changes bump MAJOR and
follow SM-20 deprecation.

#### Acceptance criteria

- Given MVP exit, when the acceptance suite runs, then the eight MVP tools pass the
  conformance suite on all Tier 1 platforms (compatibility case).
- Given the offline condition, when the MVP catalog tools execute against a local workspace,
  then zero network attempts are observed (negative/offline case).
- Given a `github.request` invocation without a configured credential, when executed, then a
  structured unauthenticated error returns and no secret prompt occurs inside the tool
  (error and security case).
- Given any catalog mutation tool, when it acts, then the corresponding Volume 2 records
  (File Change, Command Execution, Patch) exist and correlate (observability case).
- Given a pre-phase tool (e.g., `notion.request` before v2), when resolution is attempted,
  then E-TOOL-001 results — unshipped tools are absent, not stubbed (negative case).

#### Verification method

Per-tool conformance suites and golden fixtures (Volume 13); offline suite; Tier 1
platform-matrix acceptance runs; integration-tool contract tests against recorded official
API fixtures.

#### Traceability

PRD-003, PRD-004, PRD-008; Volume 1 MVP minimum items 10–12; ADR-025, ADR-070, ADR-074.

### NFR-TOOL-001 — Built-in tool contract conformance

- Category: Reliability
- Priority: P0
- Phase: MVP
- Metric: Fraction of shipped built-in tools passing the full tool contract conformance suite (schema validation, output conformance, truncation marking, cancellation-within-budget, permission mediation, record production) per release, per Tier 1 platform
- Target: 100% of shipped built-in tools, on every Tier 1 platform
- Minimum threshold: 100% (a failing built-in blocks the release)
- Measurement method: Conformance suite executed in CI on release candidates across the Tier 1 matrix; results attached to the release audit
- Test environment: Tier 1 platforms per Volume 1 chapter 05; reference repository per Volume 1 chapter 06
- Measurement frequency: Every release; full matrix at phase gates
- Owner: Tool Runtime (Volume 6)
- Dependencies: FR-TOOL-001, FR-TOOL-007
- Risks: RISK-TOOL-003
- Acceptance criteria: The release pipeline records a per-tool, per-platform conformance report with zero failures; any failure blocks publication until fixed or the tool is withdrawn through the change procedure.

## Risks

### RISK-TOOL-003 — External service API drift breaking integration tools

- Category: External dependency
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: Official-API-only policy with per-service PENDING VALIDATION gates (ADR-074);
  contract tests against recorded official fixtures re-verified on schedule; per-tool SemVer
  so breaking service changes ship as major tool versions with deprecation notes; MCP-bridge
  fallback documented for services whose APIs regress
- Detection: Scheduled integration test jobs against live services (consent-gated,
  rate-limit-respecting); field error-rate telemetry on E-TOOL-006 codes per integration tool
- Owner: Tool Runtime (Volume 6)
- Status: Open — PENDING VALIDATION per service (V6A-OQ-2)

External APIs version, deprecate, and change rate limits on their owners' schedules. The
catalog absorbs drift at the tool boundary: typed operations insulate agents from transport
details, and phasing (ADR-074) keeps unverified service facts out of committed contracts.
