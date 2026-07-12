#!/usr/bin/env python3
"""Roadmap status automation (Volume 11 chapters 05/06, FR-GH-008).

Drives the "Andromeda Roadmap" Projects v2 board through its Status lifecycle in
response to pull-request and release events, and stamps the Target release field
when work ships. The intake half (issue -> Backlog) lives in the
`actions/add-to-project` step of `.github/workflows/project.yml`; this script is
the richer half that the default GITHUB_TOKEN cannot express.

Transitions (all forward-only -- a later state is never dragged backwards):

    pr-opened  --pr N     linked issues -> "In Review"
    pr-merged  --pr N     linked issues -> "Validation"
    release    --tag T     every "Validation" item -> "Released" + Target release = T
    drop       --issue N   remove the issue's item from the board (closed not-planned)

"Linked issue" means an issue named in the PR's `closingIssuesReferences`
(i.e. `Closes #N` / `Fixes #N`), resolved through the GraphQL API so it matches
GitHub's own definition rather than parsing the PR body ourselves.

Auth: shells out to `gh api graphql`, so it uses whatever token `gh` is
configured with. In CI that is the ROADMAP_PROJECT_TOKEN secret (a PAT with the
`project` scope); locally it is the ambient `gh` login. No token is ever read,
echoed, or logged by this script.

Config (env, all optional -- defaults target this repo's board):
    ANDROMEDA_PROJECT_OWNER   project owner login      (default: datamaia)
    ANDROMEDA_PROJECT_NUMBER  project number           (default: 1)
    GH_REPO / --repo          OWNER/NAME for PR lookups (default: datamaia/andromeda)

Exit codes: 0 success (including no-op), 2 usage error, 1 GraphQL/API failure.
"""
from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys

OWNER = os.environ.get("ANDROMEDA_PROJECT_OWNER", "datamaia")
NUMBER = int(os.environ.get("ANDROMEDA_PROJECT_NUMBER", "1"))

# Status lifecycle order. A transition only advances an item; it never regresses
# one (a merged PR must not knock a Released item back to Validation). "Blocked"
# is treated as orthogonal -- ranked next to In Progress so it can still advance.
STATUS_ORDER = [
    "Backlog",
    "Ready",
    "In Progress",
    "Blocked",
    "In Review",
    "Validation",
    "Done",
    "Released",
]


class GraphQLError(RuntimeError):
    """A `gh api graphql` call returned an error payload or failed to run."""


def gh_graphql(query: str, **variables: object) -> dict:
    """Run a GraphQL query/mutation via `gh api graphql` and return `data`.

    String variables are passed with `-f` (raw), everything else (ints, bools)
    with `-F` (typed) so `pr` arrives as an Int, not a string.
    """
    cmd = ["gh", "api", "graphql", "-f", f"query={query}"]
    for key, value in variables.items():
        if value is None:
            continue  # omit -> GraphQL sees the variable's default (null)
        if isinstance(value, str):
            cmd += ["-f", f"{key}={value}"]
        else:
            cmd += ["-F", f"{key}={json.dumps(value)}"]
    proc = subprocess.run(cmd, capture_output=True, text=True)
    if proc.returncode != 0:
        raise GraphQLError(proc.stderr.strip() or "gh api graphql failed")
    payload = json.loads(proc.stdout)
    if payload.get("errors"):
        raise GraphQLError(json.dumps(payload["errors"]))
    return payload["data"]


# --- board metadata -------------------------------------------------------

# Same selection set for a user- or org-owned board; `{holder}` is swapped in.
_META_QUERY = """
query($owner: String!, $number: Int!) {
  {holder}(login: $owner) {
    projectV2(number: $number) {
      id
      status: field(name: "Status") {
        ... on ProjectV2SingleSelectField { id options { id name } }
      }
      target: field(name: "Target release") {
        ... on ProjectV2FieldCommon { id }
      }
    }
  }
}
"""


def load_board() -> dict:
    """Resolve project id, Status field id + option map, Target release field id.

    Looked up at runtime rather than hardcoded so the automation survives a board
    that is recreated or has its options edited. Tries the `user` owner first and
    falls back to `organization`, so the same script works if the board is ever
    moved to an org.
    """
    holder = None
    for kind in ("user", "organization"):
        data = gh_graphql(_META_QUERY.replace("{holder}", kind), owner=OWNER, number=NUMBER)
        holder = (data.get(kind) or {}).get("projectV2")
        if holder is not None:
            break
    if holder is None:
        raise GraphQLError(f"project #{NUMBER} not found for owner {OWNER!r}")
    status = holder["status"]
    return {
        "project_id": holder["id"],
        "status_field_id": status["id"],
        "options": {o["name"]: o["id"] for o in status["options"]},
        "target_field_id": (holder.get("target") or {}).get("id"),
    }


# --- linked-issue resolution ---------------------------------------------

_CLOSING_ISSUES_QUERY = """
query($owner: String!, $name: String!, $pr: Int!) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $pr) {
      closingIssuesReferences(first: 50) { nodes { id number } }
    }
  }
}
"""

_ISSUE_ITEM_QUERY = """
query($issueId: ID!) {
  node(id: $issueId) {
    ... on Issue {
      number
      projectItems(first: 50) {
        nodes {
          id
          project { id }
          fieldValueByName(name: "Status") {
            ... on ProjectV2ItemFieldSingleSelectValue { name }
          }
        }
      }
    }
  }
}
"""


def closing_issue_ids(repo: str, pr: int) -> list[dict]:
    """Return [{id, number}] for issues the PR closes, per GitHub's own linkage."""
    owner, name = repo.split("/", 1)
    data = gh_graphql(_CLOSING_ISSUES_QUERY, owner=owner, name=name, pr=pr)
    ref = ((data.get("repository") or {}).get("pullRequest") or {})
    return (ref.get("closingIssuesReferences") or {}).get("nodes") or []


def issue_node_id(repo: str, number: int) -> str | None:
    """GraphQL node id for issue #number, or None if it doesn't exist."""
    owner, name = repo.split("/", 1)
    data = gh_graphql(
        "query($o:String!,$n:String!,$num:Int!){repository(owner:$o,name:$n)"
        "{issue(number:$num){id}}}",
        o=owner, n=name, num=number,
    )
    issue = ((data.get("repository") or {}).get("issue") or {})
    return issue.get("id")


def item_on_board(issue_id: str, project_id: str) -> dict | None:
    """The issue's project item on our board, or None if it isn't on it.

    Returns {"id": ..., "status": <current status name or None>}.
    """
    data = gh_graphql(_ISSUE_ITEM_QUERY, issueId=issue_id)
    node = data.get("node") or {}
    for item in (node.get("projectItems") or {}).get("nodes") or []:
        if (item.get("project") or {}).get("id") == project_id:
            value = item.get("fieldValueByName") or {}
            return {"id": item["id"], "status": value.get("name")}
    return None


# --- mutations ------------------------------------------------------------

_SET_SELECT = """
mutation($project: ID!, $item: ID!, $field: ID!, $option: String!) {
  updateProjectV2ItemFieldValue(input: {
    projectId: $project, itemId: $item, fieldId: $field,
    value: { singleSelectOptionId: $option }
  }) { projectV2Item { id } }
}
"""

_SET_TEXT = """
mutation($project: ID!, $item: ID!, $field: ID!, $text: String!) {
  updateProjectV2ItemFieldValue(input: {
    projectId: $project, itemId: $item, fieldId: $field,
    value: { text: $text }
  }) { projectV2Item { id } }
}
"""

_ARCHIVE = """
mutation($project: ID!, $item: ID!) {
  archiveProjectV2Item(input: { projectId: $project, itemId: $item }) {
    item { id }
  }
}
"""


def set_status(board: dict, item_id: str, status_name: str) -> None:
    gh_graphql(
        _SET_SELECT,
        project=board["project_id"],
        item=item_id,
        field=board["status_field_id"],
        option=board["options"][status_name],
    )


def archive_item(board: dict, item_id: str) -> None:
    gh_graphql(_ARCHIVE, project=board["project_id"], item=item_id)


def set_target_release(board: dict, item_id: str, tag: str) -> None:
    field = board["target_field_id"]
    if not field:
        print("  ! Target release field missing; skipping stamp", file=sys.stderr)
        return
    gh_graphql(_SET_TEXT, project=board["project_id"], item=item_id, field=field, text=tag)


def rank(status_name: str | None) -> int:
    """Order index for a status; unknown/None sorts before everything (-1)."""
    try:
        return STATUS_ORDER.index(status_name)  # type: ignore[arg-type]
    except ValueError:
        return -1


# --- board scan (release stamping) ---------------------------------------

_ITEMS_QUERY = """
query($project: ID!, $cursor: String) {
  node(id: $project) {
    ... on ProjectV2 {
      items(first: 100, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          fieldValueByName(name: "Status") {
            ... on ProjectV2ItemFieldSingleSelectValue { name }
          }
        }
      }
    }
  }
}
"""


def items_with_status(board: dict, status_name: str) -> list[str]:
    """Item ids on the board whose current Status equals `status_name`."""
    ids: list[str] = []
    cursor: str | None = None
    while True:
        data = gh_graphql(_ITEMS_QUERY, project=board["project_id"], cursor=cursor)
        conn = (data.get("node") or {}).get("items") or {}
        for item in conn.get("nodes") or []:
            if (item.get("fieldValueByName") or {}).get("name") == status_name:
                ids.append(item["id"])
        page = conn.get("pageInfo") or {}
        if page.get("hasNextPage"):
            cursor = page["endCursor"]
        else:
            return ids


# --- commands -------------------------------------------------------------


def advance_linked(board: dict, repo: str, pr: int, target: str) -> int:
    """Move every issue the PR closes to `target`, forward-only. Returns moves."""
    issues = closing_issue_ids(repo, pr)
    if not issues:
        print(f"PR #{pr}: no closing issue references; nothing to do")
        return 0
    moves = 0
    for issue in issues:
        item = item_on_board(issue["id"], board["project_id"])
        num = issue["number"]
        if item is None:
            print(f"  issue #{num}: not on the roadmap board; skipped")
            continue
        current = item["status"]
        if rank(current) >= rank(target):
            print(f"  issue #{num}: already at {current!r} (>= {target!r}); left as-is")
            continue
        set_status(board, item["id"], target)
        print(f"  issue #{num}: {current!r} -> {target!r}")
        moves += 1
    return moves


def release(board: dict, tag: str) -> int:
    """Every Validation item -> Released, stamped with the release tag."""
    item_ids = items_with_status(board, "Validation")
    if not item_ids:
        print(f"release {tag}: no items in Validation; nothing to ship")
        return 0
    for item_id in item_ids:
        set_target_release(board, item_id, tag)
        set_status(board, item_id, "Released")
        print(f"  item {item_id}: Validation -> Released ({tag})")
    return len(item_ids)


def drop(board: dict, repo: str, issue: int) -> int:
    """Archive the issue's board item (it left the board -- closed not-planned)."""
    node = issue_node_id(repo, issue)
    if node is None:
        print(f"issue #{issue}: not found; nothing to drop")
        return 0
    item = item_on_board(node, board["project_id"])
    if item is None:
        print(f"issue #{issue}: not on the roadmap board; nothing to drop")
        return 0
    archive_item(board, item["id"])
    print(f"  issue #{issue}: archived (left the board)")
    return 1


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Andromeda roadmap status automation")
    parser.add_argument(
        "--repo",
        default=os.environ.get("GH_REPO", "datamaia/andromeda"),
        help="OWNER/NAME for PR lookups (default: datamaia/andromeda)",
    )
    sub = parser.add_subparsers(dest="command", required=True)
    p_open = sub.add_parser("pr-opened", help="linked issues -> In Review")
    p_open.add_argument("--pr", type=int, required=True)
    p_merged = sub.add_parser("pr-merged", help="linked issues -> Validation")
    p_merged.add_argument("--pr", type=int, required=True)
    p_rel = sub.add_parser("release", help="Validation items -> Released + stamp tag")
    p_rel.add_argument("--tag", required=True)
    p_drop = sub.add_parser("drop", help="remove an issue's item from the board")
    p_drop.add_argument("--issue", type=int, required=True)
    args = parser.parse_args(argv)

    try:
        board = load_board()
        if args.command == "pr-opened":
            advance_linked(board, args.repo, args.pr, "In Review")
        elif args.command == "pr-merged":
            advance_linked(board, args.repo, args.pr, "Validation")
        elif args.command == "release":
            release(board, args.tag)
        elif args.command == "drop":
            drop(board, args.repo, args.issue)
    except GraphQLError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
