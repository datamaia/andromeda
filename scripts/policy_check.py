#!/usr/bin/env python3
"""Workflow security policy check (Volume 11 chapter 06, FR-GH-012; ADR-149).

Mechanically enforces the workflow security posture on every file in .github/workflows:

  1. an explicit ``permissions`` block (top-level, or on every job);
  2. every ``uses:`` pinned to a full 40-hex commit SHA (not a mutable tag);
  3. no ``pull_request_target`` trigger that also checks out the pull-request head;
  4. privileged jobs (contents/id-token/packages: write) gated by an ``environment``;
  5. GitHub-hosted runner labels only.

Fails closed: an unparsable workflow is an error, not a skip. Also validates the label
taxonomy data (.github/labels.yml). Requires PyYAML.

Usage: python3 scripts/policy_check.py
"""
from __future__ import annotations

import glob
import re
import sys

import yaml

HOSTED_PREFIXES = ("ubuntu-", "macos-", "windows-")
SHA_RE = re.compile(r"^[0-9a-f]{40}$")
USES_RE = re.compile(r"^\s*-?\s*uses:\s*(\S+)", re.MULTILINE)
PRIVILEGED_WRITES = ("contents", "id-token", "packages")

errors: list[str] = []


def err(where: str, msg: str) -> None:
    errors.append(f"{where}: {msg}")


def workflow_on(wf: dict):
    """Return the ``on`` mapping, working around YAML parsing ``on`` as the boolean True."""
    if "on" in wf:
        return wf["on"]
    return wf.get(True)


def check_permissions(f: str, wf: dict) -> None:
    if "permissions" in wf:
        return  # a top-level block covers every job
    for name, job in (wf.get("jobs") or {}).items():
        if not isinstance(job, dict) or "permissions" not in job:
            err(f, f"job '{name}' has no permissions block and there is no top-level one (ADR-149 least privilege)")


def check_sha_pins(f: str, text: str) -> None:
    for ref in USES_RE.findall(text):
        if ref.startswith("./") or ref.startswith("docker://"):
            continue  # local or docker actions are not tag-pinned
        if "@" not in ref:
            err(f, f"action '{ref}' is not pinned (no @sha)")
            continue
        sha = ref.rsplit("@", 1)[1]
        if not SHA_RE.match(sha):
            err(f, f"action '{ref}' is not pinned to a full commit SHA (ADR-149)")


def check_pull_request_target(f: str, wf: dict, text: str) -> None:
    on = workflow_on(wf)
    triggers = on if isinstance(on, dict) else ([on] if isinstance(on, str) else (on or []))
    names = set(triggers.keys()) if isinstance(on, dict) else set(triggers)
    if "pull_request_target" not in names:
        return
    # pull_request_target runs with a read/write token in the base repo's context; checking out
    # PR-authored code under it is the classic RCE vector (ADR-149 rule 3).
    if re.search(r"ref:\s*\$\{\{\s*github\.event\.pull_request\.head", text):
        err(f, "pull_request_target checks out the PR head ref — prohibited (ADR-149)")


def check_privileged_environment(f: str, wf: dict) -> None:
    top = wf.get("permissions") or {}
    for name, job in (wf.get("jobs") or {}).items():
        if not isinstance(job, dict):
            continue
        perms = job.get("permissions", top)
        if not isinstance(perms, dict):
            continue
        privileged = [p for p in PRIVILEGED_WRITES if perms.get(p) == "write"]
        if privileged and "environment" not in job:
            err(f, f"job '{name}' is privileged ({', '.join(privileged)}: write) but is not environment-gated (ADR-149)")


def check_runners(f: str, wf: dict) -> None:
    labels: list[str] = []
    for name, job in (wf.get("jobs") or {}).items():
        if not isinstance(job, dict):
            continue
        ro = job.get("runs-on")
        if isinstance(ro, str) and "${{" not in ro:
            labels.append(ro)
        elif isinstance(ro, list):
            labels.extend(x for x in ro if isinstance(x, str) and "${{" not in x)
        # matrix-provided runner labels
        matrix = (job.get("strategy") or {}).get("matrix") if isinstance(job.get("strategy"), dict) else None
        if isinstance(matrix, dict):
            for key, vals in matrix.items():
                if key in ("os", "runner", "runs-on") and isinstance(vals, list):
                    labels.extend(x for x in vals if isinstance(x, str))
    for label in labels:
        if not label.startswith(HOSTED_PREFIXES):
            err(f, f"runner label '{label}' is not a GitHub-hosted runner (ADR-149: no self-hosted)")


def check_labels_data() -> None:
    path = ".github/labels.yml"
    try:
        data = yaml.safe_load(open(path))
    except FileNotFoundError:
        err(path, "missing label taxonomy data (FR-GH-007)")
        return
    except Exception as e:  # noqa: BLE001
        err(path, f"unparsable (fail closed): {e}")
        return
    if not isinstance(data, list) or not data:
        err(path, "must be a non-empty list of label definitions")
        return
    seen: set[str] = set()
    for i, item in enumerate(data):
        if not isinstance(item, dict) or "name" not in item or "color" not in item:
            err(path, f"entry {i} must have 'name' and 'color'")
            continue
        name, color = item["name"], str(item["color"])
        if name in seen:
            err(path, f"duplicate label '{name}'")
        seen.add(name)
        if not re.match(r"^[0-9a-fA-F]{6}$", color):
            err(path, f"label '{name}' color '{color}' must be 6 hex digits, no leading '#'")


def main() -> int:
    files = sorted(glob.glob(".github/workflows/*.yml"))
    if not files:
        err(".github/workflows", "no workflow files found")
    for f in files:
        text = open(f).read()
        try:
            wf = yaml.safe_load(text)
        except Exception as e:  # noqa: BLE001 — fail closed
            err(f, f"unparsable workflow (fail closed): {e}")
            continue
        if not isinstance(wf, dict):
            err(f, "workflow is not a mapping")
            continue
        check_permissions(f, wf)
        check_sha_pins(f, text)
        check_pull_request_target(f, wf, text)
        check_privileged_environment(f, wf)
        check_runners(f, wf)
    check_labels_data()

    if errors:
        print("policy-check: FAILED")
        for e in errors:
            print(f"  {e}")
        return 1
    print(f"policy-check: OK ({len(files)} workflows, labels data valid)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
