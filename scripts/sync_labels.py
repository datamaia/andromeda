#!/usr/bin/env python3
"""Synchronize the repository label taxonomy from .github/labels.yml (FR-GH-007).

Idempotent upsert: creates each label, or updates its color/description if it already exists.
It does not delete labels absent from the file (pruning is intentionally out of scope so shared
platform labels are never removed). Uses the `gh` CLI, which authenticates from the ambient
GITHUB_TOKEN in Actions; needs `issues: write`.

Usage: python3 scripts/sync_labels.py [--repo OWNER/REPO] [--dry-run]
"""
from __future__ import annotations

import argparse
import subprocess
import sys

import yaml


def gh(args: list[str]) -> subprocess.CompletedProcess:
    return subprocess.run(["gh", *args], capture_output=True, text=True)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", default=None)
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--file", default=".github/labels.yml")
    a = ap.parse_args()

    labels = yaml.safe_load(open(a.file))
    if not isinstance(labels, list):
        print(f"sync-labels: {a.file} must be a list", file=sys.stderr)
        return 1

    repo_args = ["--repo", a.repo] if a.repo else []
    created = updated = 0
    for item in labels:
        name = item["name"]
        color = str(item["color"])
        desc = item.get("description", "")
        if a.dry_run:
            print(f"would upsert: {name} #{color} — {desc}")
            continue
        # `gh label create --force` creates or updates in one call.
        r = gh(["label", "create", name, "--color", color, "--description", desc, "--force", *repo_args])
        if r.returncode != 0:
            print(f"sync-labels: failed for '{name}': {r.stderr.strip()}", file=sys.stderr)
            return 1
        # `--force` prints "created" or "updated"; count best-effort.
        if "updat" in (r.stdout + r.stderr).lower():
            updated += 1
        else:
            created += 1

    if not a.dry_run:
        print(f"sync-labels: OK ({created} created, {updated} updated, {len(labels)} total)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
