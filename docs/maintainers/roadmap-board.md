# Roadmap board runbook

The **Andromeda Roadmap** GitHub Project (`FR-GH-008`) is the single tracker for all work.
Most of it is provisioned and driven by scripts; this runbook records what is automated,
how to provision it from scratch, and the one part GitHub's API cannot create — the **views**,
which must be added once in the UI.

- Board: <https://github.com/users/datamaia/projects/1>
- Automation script: [`scripts/project_sync.py`](../../scripts/project_sync.py)
- Automation workflow: [`.github/workflows/project.yml`](../../.github/workflows/project.yml)
- One-time provisioning: [`scripts/setup_project.sh`](../../scripts/setup_project.sh)

## Statuses (8)

`Backlog` · `Ready` · `In Progress` · `In Review` · `Blocked` · `Validation` · `Done` · `Released`

`Validation` holds merged work awaiting a phase gate or release; `Released` is set when the
change ships in a published release.

## Fields (9)

| Field | Kind | Populated by |
|---|---|---|
| Status | single select (the 8 above) | automation + humans |
| Area | single select (24 scopes) | triage (mirrors scope) |
| Phase | single select (Core, MVP, Beta, v1, v2, Future) | mirrors `phase:*` |
| Priority | single select (P0–P3) | mirrors `priority:*` |
| Size | single select (XS, S, M, L, XL) | triage estimate |
| Iteration | iteration (2-week cadence) | planning |
| Target release | text (`vX.Y.Z`) | release automation + milestones |
| Risk | single select (low, medium, high, critical) | mirrors `risk:*` |
| Requirements | text (corpus IDs) | issue form |

Statuses and fields are created by `scripts/setup_project.sh` plus the field/GraphQL calls it
prints; all nine exist on board #1 today.

## Automations (all live in `project.yml`)

Every job is guarded on the `ROADMAP_PROJECT_URL` repository variable, so the workflow is a no-op
until the board is provisioned. Projects v2 are outside the default `GITHUB_TOKEN`, so the jobs
authenticate with the `ROADMAP_PROJECT_TOKEN` secret (a PAT with the `project` scope).

| Trigger | Job | Effect |
|---|---|---|
| issue opened / `type:*` labeled | `intake` | add to board → `Backlog` |
| linked PR opened / reopened / ready | `in-review` | linked issues → `In Review` |
| linked PR merged | `validation` | linked issues → `Validation` |
| release published | `released` | every `Validation` item → `Released` + stamp Target release |
| issue closed as *not planned* | `dropped` | archive the item (it leaves the board) |

"Linked issue" is resolved through the PR's `closingIssuesReferences` (i.e. `Closes #N`), so it
matches GitHub's own linkage. Transitions are **forward-only**: a merged PR never drags a
`Released` item back to `Validation`, and status *regressions* (e.g. `Done` → `In Progress`) stay
manual by design. The status transitions run on `pull_request_target` (the only PR event that
exposes the secret to fork PRs and runs in the trusted base context); no job ever checks out or
runs PR-authored code, and untrusted values reach the script only through `env` — see the header
of `project.yml` and `scripts/policy_check.py`.

Run any transition by hand against the live board:

```bash
export GH_TOKEN="$(gh auth token)"   # a login with the 'project' scope
python3 scripts/project_sync.py pr-opened --pr 123
python3 scripts/project_sync.py pr-merged --pr 123
python3 scripts/project_sync.py release  --tag v0.3.0
python3 scripts/project_sync.py drop     --issue 123
```

## Views — UI-only (API limitation)

GitHub's public GraphQL API can create statuses, fields, and items, but it exposes **no
`createProjectV2View` / `updateProjectV2View` mutation** — Projects v2 views cannot be scripted.
These six views (three named + three filtered) are therefore the one manual step; everything else
above is automated. They are already created on board #1; the settings below are the source of
truth for re-creating them (**+ New view**) if the board is ever rebuilt.

| # | View name | Layout | Configuration |
|---|---|---|---|
| 1 | **Board** | Board | Column field: **Status**. Execution view. |
| 2 | **Roadmap** | Roadmap | Date fields: **Iteration** (marker) + **Target release**. Group by Phase. Planning view; this is the public roadmap. |
| 3 | **Triage** | Table | Group by **Area**; sort by Priority. Load/triage view. |
| 4 | **Incidents** | Table | Filter: `priority:P0 severity:critical,severity:high,severity:medium,severity:low`. Incident planning. |
| 5 | **MVP burn-down** | Table | Filter: `phase:MVP` (or label `phase:mvp`); group by Status. |
| 6 | **Security review** | Table | Filter: `label:"type:security"` (add `security-review` once that label exists). Security queue. |

Filter syntax notes: `field:Option` matches a single-select field's option (quote names with
spaces, e.g. `"Target release":v0.3.0`); `label:"name"` matches a label; space = AND, comma = OR
within one field. Adjust option names if the board's fields are renamed.

## Provision from scratch

```bash
# 1) grant the 'project' scope to the gh login, then create the board + statuses + fields
gh auth refresh -h github.com -s project,read:project
scripts/setup_project.sh --repo datamaia/andromeda

# 2) wire the workflow to it (prints exact commands at the end of setup_project.sh)
gh variable set ROADMAP_PROJECT_URL --repo datamaia/andromeda --body "<board url>"
#    ROADMAP_PROJECT_TOKEN must be a PAT with the 'project' scope
```

Then add the six views above in the UI. After that the board is fully self-maintaining.
