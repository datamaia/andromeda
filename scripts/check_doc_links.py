#!/usr/bin/env python3
"""Validate that relative Markdown links in tracked files resolve (Volume 11, FR-GH-011).

The public repository has no separate documentation site and the specification is a private
companion document, so rather than publishing, this gate keeps the tracked Markdown honest:
every relative link (and image) must point at a file that exists in the checkout. External
(http/https/mailto) links and pure ``#anchor`` links are not fetched.

Usage: python3 scripts/check_doc_links.py
"""
from __future__ import annotations

import os
import re
import subprocess
import sys

LINK_RE = re.compile(r"!?\[[^\]]*\]\(([^)]+)\)")


def tracked_markdown() -> list[str]:
    out = subprocess.run(["git", "ls-files", "*.md"], capture_output=True, text=True).stdout
    return [line for line in out.splitlines() if line]


def main() -> int:
    errors: list[str] = []
    checked = 0
    for f in tracked_markdown():
        base = os.path.dirname(f)
        text = open(f, encoding="utf-8").read()
        for m in LINK_RE.finditer(text):
            target = m.group(1).strip()
            if not target:
                continue
            target = target.split()[0]  # drop an optional "title"
            if target.startswith("<") and target.endswith(">"):
                target = target[1:-1]
            low = target.lower()
            if low.startswith(("http://", "https://", "mailto:", "#", "tel:")):
                continue
            path = target.split("#", 1)[0]  # drop any anchor fragment
            if not path:
                continue
            checked += 1
            resolved = os.path.normpath(os.path.join(base, path))
            if not os.path.exists(resolved):
                errors.append(f"{f}: broken link -> {target}")

    if errors:
        print("check-doc-links: FAILED")
        for e in errors:
            print(f"  {e}")
        return 1
    print(f"check-doc-links: OK ({checked} relative links resolve)")
    return 0


if __name__ == "__main__":
    sys.exit(main())
