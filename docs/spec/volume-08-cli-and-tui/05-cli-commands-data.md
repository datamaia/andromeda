# 05 — CLI Commands: Data and Inspection

This chapter specifies the data families: `memory`, `context`, `index`, `git`, `logs`,
`trace`, and `export`. Chapter [02](02-cli-conventions.md) conventions apply throughout.
Every command in this chapter is **Phase: MVP** (they surface the offline guarantee list
and MVP minimum items: local memory queries, indexing, basic Git, logs). All operate fully
offline except where a section names `network` explicitly. FR-CLI-015 binds these sections
normatively.

## `andromeda memory`

Memory inspection and curation over MemoryStorePort (owner Volume 7; FR-MEM-001). Layer
names and retention semantics are Volume 7's; the CLI renders them verbatim, including the
frozen Memory Record status vocabulary (`active`, `archived`, `expired`, `deleted`).

```text
andromeda memory search <query> [--layer <layer>] [--limit <n>]
andromeda memory show <record-id>
andromeda memory add <text> [--layer <layer>]
andromeda memory delete <record-id>...
andromeda memory export [--layer <layer>] [--output <path>]
```

Behavior: `search` queries lexically (and semantically where an index provides it) and
renders records with provenance and trust attribution (Volume 7). `add` ingests a
user-authored record with provenance `user`; default layer per Volume 7. `delete` is a
destructive confirmation (hard deletion honoring cascade rules; deletion records persist).
`export` streams canonical JSON entity documents (Volume 2, chapter 10 export forms) —
NDJSON on stdout by default, a file with `--output`.

**Permissions:** none for `search`/`show`; `write` for `add`/`delete`; `write` for
`export --output`. **Exit codes:** 0, 1, 2, 3, 5, 8, 9.

**JSON result (`memory search`, `data`):**

```json
{
  "records": [
    {
      "record_id": "01JZX0MEM00000000000000000",
      "layer": "workspace",
      "status": "active",
      "summary": "Project uses table-driven tests exclusively",
      "provenance": "user",
      "created_at": "2026-07-10T16:40:00Z"
    }
  ]
}
```

**Examples:**

```text
$ andromeda memory search "testing conventions" --layer workspace   # valid
$ andromeda memory delete 01JZX0ME --yes                            # valid, scripted
$ andromeda memory add                                              # invalid
error[E-CLI-002]: invalid value for <text>: memory content is required
```

**Errors:** E-CLI-002/003; E-MEM family (Volume 7).

## `andromeda context`

Context assembly inspection and user steering over the Context Manager (owner Volume 7;
FR-CTX-001): what would enter a model request, and the user's pins and exclusions.

```text
andromeda context show [--session <id>]
andromeda context pin <path-or-item-id> [--session <id>]
andromeda context unpin <path-or-item-id> [--session <id>]
andromeda context exclude <path-or-glob> [--workspace]
```

Behavior: `show` renders the current assembly for the active (or named) session: context
items with source, priority, token counts, and budget utilization — the transparency view
of Principle 7 (context state). `pin` forces inclusion, `unpin` reverses it, `exclude`
blocks paths/globs from assembly (session-scoped by default, workspace-scoped with
`--workspace`); semantics and precedence of pins/exclusions are Volume 7's.

**Permissions:** none for `show`; `write` (context preferences) for mutations.
**Exit codes:** 0, 1, 2, 3.

**JSON result (`context show`, `data`):**

```json
{
  "session_id": "01JZX0A1B2C3D4E5F6G7H8J9K1",
  "budget_tokens": 32768,
  "used_tokens": 18240,
  "items": [
    {"source": "file", "ref": "src/auth/signup.go", "priority": "pinned", "tokens": 1240}
  ]
}
```

**Examples:**

```text
$ andromeda context pin src/auth/signup.go            # valid
$ andromeda context exclude "vendor/**" --workspace   # valid
$ andromeda context unpin                             # invalid
error[E-CLI-002]: invalid value for <path-or-item-id>: a pinned path or item is required
```

**Errors:** E-CLI-002; E-CTX family (Volume 7).

## `andromeda index`

Index lifecycle over IndexerPort (owner Volume 7; FR-IDX-001; frozen Index states). Works
without embeddings and without Internet (lexical indexes; ADR-020 governs semantic ones).

```text
andromeda index build [--semantic] [--path <scope>]
andromeda index update [--path <scope>]
andromeda index status
andromeda index search <query> [--semantic] [--limit <n>]
andromeda index invalidate [--path <scope>]
andromeda index remove
```

Behavior: `build` runs a full build as a supervised background-capable operation with
FR-UX-003 progress; `--semantic` requires an embeddings-capable provider (declared
`embeddings` capability) and fails precisely when none is configured — never silently
degrading to lexical (Principle 2). `update` applies incremental changes; `status` renders
per-index state, generation, and staleness. `invalidate` marks scopes stale; `remove`
drops the index cache — destructive confirmation, with text stating the cache is
rebuildable (never data loss, ADR-028 rule).

**Permissions:** `read` (workspace files); `network` only for `--semantic` with a remote
embeddings provider. **Exit codes:** 0, 1, 2, 3, 7 (semantic embeddings provider failure),
8.

**JSON result (`index status`, `data`):**

```json
{
  "indexes": [
    {
      "index_id": "01JZX0IDX00000000000000000",
      "kind": "lexical",
      "state": "ready",
      "generation": 12,
      "coverage": {"files": 4980, "stale_paths": 0}
    }
  ]
}
```

**Examples:**

```text
$ andromeda index build                              # valid, lexical, offline
$ andromeda index search "http retry policy"         # valid
$ andromeda index build --semantic                   # invalid without embeddings provider
error[E-IDX-...]: ...                                # E-IDX family text (Volume 7), exit 7 when provider-caused
```

**Errors:** E-CLI-002/003; E-IDX family (Volume 7); E-PROV family exit 7 for embeddings
calls.

## `andromeda git`

Git operations through the Git Engine (GitPort; owner Volume 11; FR-GIT-001; ADR-025).
The reason this surface exists next to plain `git`: operations run through Andromeda are
permission-mediated and **attributed** — they produce File Change/Patch/Command records
correlated to sessions and runs (SM-13), exactly like agent-initiated Git actions.

```text
andromeda git status
andromeda git diff [<rev-spec>] [--staged]
andromeda git log [--limit <n>]
andromeda git stage <path>...
andromeda git unstage <path>...
andromeda git commit --message <text>
andromeda git branch [<name>]
andromeda git switch <name> [--create]
andromeda git apply <patch-id-or-path>
```

Behavior: read commands (`status`, `diff`, `log`, `branch` without name) render GitPort
results; `diff` streams hunks and pages per FR-UX-002. Mutations (`stage`, `unstage`,
`commit`, `branch <name>`, `switch`, `apply`) hold a permission decision before invoking
GitPort (Volume 3 rule) and follow Volume 11's operation semantics — including its
no-silent-destructive-operations rule; the CLI adds no Git behavior of its own. `apply`
applies a reviewed Patch (frozen status `proposed` → `applied`) atomically or not at all.
Commit message content rules (ADR-015: change information only) are Volume 11's and
enforced there.

**Permissions:** none for reads; `git_mutation` for mutations. **Exit codes:** 0, 1, 2,
3, 5, 8; E-GIT family failures map per Volume 11 envelopes (within 1/3/5/8).

**JSON result (`git status`, `data`):**

```json
{
  "branch": "feat/signup-validation",
  "ahead": 2,
  "behind": 0,
  "staged": [{"path": "src/auth/signup.go", "change": "modified"}],
  "unstaged": [],
  "untracked": ["notes.md"]
}
```

**Examples:**

```text
$ andromeda git diff --staged                         # valid
$ andromeda git commit --message "fix: validate signup email"   # valid
$ andromeda git commit                                # invalid
error[E-CLI-002]: invalid value for --message: a commit message is required
```

**Errors:** E-CLI-002; E-GIT family (Volume 11); E-SEC family exit 5 for denied mutations.

## `andromeda logs`

Local log inspection (Logging pipeline owner Volume 10; ADR-011). Offline guarantee item:
viewing logs never needs the network.

```text
andromeda logs [--follow] [--level <level>] [--since <time-or-duration>]
               [--session <id>] [--run <id>] [--limit <n>]
```

Behavior: renders structured log records as human lines (timestamp, level, component,
message, correlation ID under `--verbose`); `--json` emits NDJSON of the Volume 10 log
record schema. `--follow` tails live records until `interrupt` (exit 0 on clean
interrupt — following is not a timeout). Filters compose conjunctively. Records are
already redacted at write time (Volume 9/10); `logs` performs no unredaction.

**Permissions:** none (local read of the user's own logs). **Exit codes:** 0, 1, 2, 3.

**JSON result:** NDJSON stream of Volume 10 log records; no `data` wrapper for `--follow`
(stream documents per FR-CLI-006 rule 3, terminal envelope on exit).

**Examples:**

```text
$ andromeda logs --since 15m --level warn             # valid
$ andromeda logs --follow --run 01JZX0RUN             # valid
$ andromeda logs --since fortnight                    # invalid
error[E-CLI-002]: invalid value for --since: expected a timestamp or duration such as 15m, 2h
```

**Errors:** E-CLI-002; E-OBS family (Volume 10).

## `andromeda trace`

Run trace inspection over the Observability query surface (owner Volume 10): the
correlated span tree of one run (Trace entity; recorded status `ok`, `error`,
`interrupted`).

```text
andromeda trace list [--session <id>] [--limit <n>]
andromeda trace show <run-id>
```

Behavior: `show` renders the span tree with per-span timing, component, and outcome,
plus the run's token/cost rollup — the UC-13 audit entry point in the CLI. `--json` emits
the persisted trace records.

**Permissions:** none. **Exit codes:** 0, 1, 2, 3.

**JSON result (`trace show`, `data`):**

```json
{
  "run_id": "01JZX0RUN00000000000000000",
  "status": "ok",
  "duration_ms": 184220,
  "spans": [
    {"name": "planner.plan", "component": "Planner", "duration_ms": 2210, "children": []}
  ]
}
```

**Examples:**

```text
$ andromeda trace show 01JZX0RUN                      # valid
$ andromeda trace show                                # invalid
error[E-CLI-002]: invalid value for <run-id>: a run ULID or unique prefix is required
```

**Errors:** E-CLI-002; E-OBS family (Volume 10).

## `andromeda export`

Data portability: canonical JSON entity documents (Volume 2, chapter 10 export forms) for
sessions, runs, memory, and audit records.

```text
andromeda export session <session-id> [--output <path>]
andromeda export run <run-id> [--output <path>]
andromeda export memory [--layer <layer>] [--output <path>]
andromeda export audit [--since <time-or-duration>] [--output <path>]
```

Behavior: single-entity exports emit one canonical document; collection exports emit
NDJSON. Default destination is stdout; `--output` writes a file (refusing to overwrite
without `--yes` — a destructive confirmation). Exports pass the full redaction pipeline:
secret references are exported as references, never material; audit exports preserve
decision records verbatim (they contain no secrets by construction, Volume 9).

**Permissions:** none for stdout export; `write` for `--output`. **Exit codes:** 0, 1, 2,
3, 5, 8, 9.

**JSON result:** the canonical export document itself is the payload (its schemas are
Volume 2's export forms, referenced as `andromeda.export.<entity>.v1`); the FR-CLI-006
envelope wraps status when `--json` is set on a `--output` run.

**Examples:**

```text
$ andromeda export run 01JZX0RUN > run.json           # valid
$ andromeda export audit --since 7d --output audit.ndjson   # valid
$ andromeda export workspace                          # invalid
error[E-CLI-001]: unknown command or flag: "workspace"
  hint:   supported kinds: session, run, memory, audit
```

**Errors:** E-CLI-001/002/003/008; E-MEM/E-CFG families per source store.

## Requirements

### FR-CLI-015 — Data command family behavior

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: CLI (Volume 8)
- Affected components: CLI; Memory Manager; Context Manager; Indexing Engine; Git Engine; Logging; Observability; Persistence Layer
- Dependencies: FR-CLI-001, FR-CLI-005–FR-CLI-012; FR-MEM-001, FR-CTX-001, FR-IDX-001 (Volume 7); FR-GIT-001 (Volume 11); FR-OBS-001 (Volume 10)
- Related risks: RISK-CLI-003

#### Description

The commands specified in this chapter — `memory`, `context`, `index`, `git`, `logs`,
`trace`, `export` — MUST behave exactly as their sections define. Family invariants: every
command except declared network paths (`index --semantic` against remote embeddings)
operates fully offline; reads require no permissions, mutations require exactly the
declared classes; deletion-class operations are destructive confirmations; all rendered
states and statuses are frozen vocabulary names; exports are canonical Volume 2 documents
passed through the redaction pipeline.

#### Motivation

This family is the transparency and local-first surface (PRD-003, PRD-006): offline
guarantee items 2, 3, 5, and 11 are exercised through these exact commands, and UC-13's
audit journey begins at `trace`/`logs`/`export audit`.

#### Actors

Users; auditors (UC-13); scripts harvesting exports; the offline test suite (SM-05).

#### Preconditions

A workspace for workspace-scoped data (`memory`, `context`, `index`, `git`); global scope
suffices for `logs`, `trace`, `export session/run/audit`.

#### Main flow

1. A caller invokes a family command per its syntax.
2. It executes through the owning port/query surface.
3. Output, exit code, and records match the section's declarations.

#### Alternative flows

- `logs --follow` streams until interrupted, exiting 0 on clean interrupt.
- Collection exports stream NDJSON; single entities emit one document.

#### Edge cases

- `memory delete` of an already-`deleted` record: idempotent success with
  `data.changed: false`.
- `index build --semantic` with no embeddings capability configured: precise failure, no
  silent lexical fallback.
- `git apply` of an already-`applied` patch: refused with the E-GIT family's state
  conflict (frozen Patch status transitions).
- `export --output` onto an existing file: destructive confirmation.

#### Inputs

Queries, record/entity identifiers, path scopes, filters, output destinations.

#### Outputs

Listings, trees, streams, and canonical export documents per the sections' schemas.

#### States

Frozen vocabularies rendered verbatim: Memory Record status, Index states, Patch status,
Trace status, Session/Run states in exports.

#### Errors

E-CLI family plus E-MEM/E-CTX/E-IDX (Volume 7), E-GIT (Volume 11), E-OBS (Volume 10),
E-PROV (embeddings, exit 7).

#### Constraints

`logs`/`trace`/`export` are read-only over already-redacted stores and MUST NOT provide
any unredaction path; `git` mutations always pass PermissionPort before GitPort (Volume 3
rule); `index remove` affects caches only.

#### Security

Deletion and export honor Volume 9's audit precedence (deletion records persist; exports
never contain secret material); `git_mutation` scopes every repository mutation; exported
files are created with user-only permissions via the PAL.

#### Observability

Family commands emit chapter 02 lifecycle events; mutations produce their owning volumes'
records correlated per SM-13 — including `git` operations attributable identically to
agent-initiated ones.

#### Performance

Search/listing latency budgets are Volume 12's; `logs --follow` consumes bounded
subscription buffers (a slow terminal cannot stall the bus, ADR-012).

#### Compatibility

Offline behavior identical across Tier 1 platforms; export documents are
platform-independent canonical JSON.

#### Acceptance criteria

- Given the offline condition (OS network disabled), when `memory search`, `index build`,
  `git status`, `logs`, and `export run` execute against local fixtures, then all succeed
  with zero network attempts (SM-05 alignment).
- Given `index build --semantic` with no embeddings provider, when invoked, then the
  failure names the missing capability and exit is 7's class per the envelope — no lexical
  fallback occurred (negative case).
- Given `andromeda git commit` denied by policy, when invoked non-interactively, then exit
  is 5, no commit exists, and the denial is recorded (permission case).
- Given `export run` of a completed fixture run, when validated, then the document matches
  the Volume 2 export form and contains no secret-classified fields (security case).
- Observability: `git stage` produces File Change records correlated to the invocation
  (SM-13 case).

#### Verification method

Offline suite (SM-05) over this family; schema conformance for exports against Volume 2
forms; permission-denial fixtures; redaction leak tests on exports; Volume 13 CLI suite
goldens.

#### Traceability

PRD-003, PRD-005, PRD-006; UC-09, UC-13; offline guarantee items 2, 3, 5, 11; FR-MEM-001,
FR-CTX-001, FR-IDX-001, FR-GIT-001, FR-OBS-001; SM-05, SM-13.
