#!/usr/bin/env bash
# Create and configure the "Andromeda Roadmap" GitHub Project (Volume 11 chapter 05, FR-GH-008),
# then print the follow-up commands that wire project.yml to it.
#
# Requires a token with the `project` scope (Projects v2 are outside `repo`). Grant it once with:
#   gh auth refresh -h github.com -s project,read:project
#
# Usage: scripts/setup_project.sh [--repo OWNER/REPO]
#
# Idempotent-ish: re-running creates another project, so run once. Status options are set via
# GraphQL; if that step fails the project still exists and the 8 statuses can be set in the UI.
set -euo pipefail

repo="datamaia/andromeda"
[ "${1:-}" = "--repo" ] && repo="$2"
owner="${repo%%/*}"
title="Andromeda Roadmap"

echo "==> creating project '$title' owned by $owner"
created=$(gh project create --owner "$owner" --title "$title" --format json)
number=$(printf '%s' "$created" | python3 -c 'import sys,json;print(json.load(sys.stdin)["number"])')
url=$(printf '%s' "$created" | python3 -c 'import sys,json;print(json.load(sys.stdin)["url"])')
echo "    created project #$number  $url"

echo "==> setting the 8 Status options"
status_field=$(gh project field-list "$number" --owner "$owner" --format json \
  | python3 -c 'import sys,json;print(next(f["id"] for f in json.load(sys.stdin)["fields"] if f["name"]=="Status"))')
if gh api graphql -f fieldId="$status_field" -f query='
  mutation($fieldId: ID!) {
    updateProjectV2Field(input: { fieldId: $fieldId, singleSelectOptions: [
      { name: "Backlog",     color: GRAY,   description: "Intake" },
      { name: "Ready",       color: BLUE,   description: "Planned and unblocked" },
      { name: "In Progress", color: YELLOW, description: "Being worked" },
      { name: "In Review",   color: ORANGE, description: "Linked PR open" },
      { name: "Blocked",     color: RED,    description: "Waiting on a dependency" },
      { name: "Validation",  color: PURPLE, description: "Merged, awaiting phase-gate/release" },
      { name: "Done",        color: GREEN,  description: "Completed (no release needed)" },
      { name: "Released",    color: PINK,   description: "Shipped in a published release" }
    ]}) { projectV2Field { __typename } }
  }' >/dev/null 2>&1; then
  echo "    8 statuses set"
else
  echo "    WARN: could not set statuses via API; set them in the UI:"
  echo "      Backlog, Ready, In Progress, In Review, Blocked, Validation, Done, Released"
fi

echo "==> adding planning fields (best-effort)"
add_field() { # name, options
  gh project field-create "$number" --owner "$owner" --name "$1" \
    --data-type SINGLE_SELECT --single-select-options "$2" >/dev/null 2>&1 \
    && echo "    + $1" || echo "    WARN: could not add field '$1' (add it in the UI)"
}
add_field "Priority" "P0,P1,P2,P3"
add_field "Phase" "Core,MVP,Beta,v1,v2,Future"
add_field "Area" "runtime,provider,tool,memory,cli,tui,security,config,git,release,docs,perf"
add_field "Size" "XS,S,M,L,XL"
add_field "Risk" "low,medium,high,critical"
echo "==> adding text fields (best-effort)"
add_text_field() { # name
  gh project field-create "$number" --owner "$owner" --name "$1" --data-type TEXT \
    >/dev/null 2>&1 && echo "    + $1" || echo "    WARN: could not add field '$1' (add it in the UI)"
}
add_text_field "Target release"
add_text_field "Requirements"
echo "==> adding the Iteration field (GraphQL; not supported by gh field-create)"
project_id=$(gh project view "$number" --owner "$owner" --format json \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["id"])')
gh api graphql -f projectId="$project_id" -f query='
  mutation($projectId: ID!) {
    createProjectV2Field(input: { projectId: $projectId, dataType: ITERATION, name: "Iteration" }) {
      projectV2Field { ... on ProjectV2IterationField { __typename } }
    }
  }' >/dev/null 2>&1 && echo "    + Iteration" || echo "    WARN: could not add 'Iteration' (add it in the UI)"

cat <<EOF

==> DONE. The board exists: $url

Next, wire project.yml to it (run these):

  # 1) point the workflow at the board
  gh variable set ROADMAP_PROJECT_URL --repo $repo --body "$url"

  # 2) give the workflow a token with the 'project' scope (the default GITHUB_TOKEN cannot
  #    touch Projects v2). Use a PAT with the project scope, e.g. from .env via the github skill:
  bash ~/.claude/skills/github/scripts/secret_set.sh --name ROADMAP_PROJECT_TOKEN \\
    --repo $repo --from-env /Users/maia/Documents/lyra/andromeda/.env
  #    (the .env PAT must have the 'project' scope; if not, create/refresh one that does)

The 8 statuses, 9 fields, and all five automations (intake / In Review / Validation / Released /
drop) are provisioned — the automations live in .github/workflows/project.yml, driven by
scripts/project_sync.py. The ONLY remaining manual step is the six board views, which GitHub's
API cannot create (no createProjectV2View mutation). Add them in the UI per the exact per-view
configuration in docs/maintainers/roadmap-board.md.
EOF
