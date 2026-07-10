#!/usr/bin/env python3
"""Andromeda specification linter.

Mechanical enforcement of Volume 0 (docs/spec/volume-00-conventions/):
identifier taxonomy and ownership, template completeness, cross-reference
resolution, embedded-example validity, and banned-content rules.

Usage:
    python3 scripts/spec_lint.py [--errors-only] [--json] [--strict] [--root PATH]

Exit codes: 0 = no errors (warnings allowed unless --strict), 1 = errors found,
2 = usage/environment error.

Python 3.11+, standard library only.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
import tomllib
from dataclasses import dataclass, field
from pathlib import Path

# --- Volume 0 mirrors (keep in sync with 03-id-taxonomy-and-ownership.md) ---

AREA_OWNERS: dict[str, str] = {
    "PRD": "volume-01",
    "ARCH": "volume-03",
    "PORT": "volume-03",
    "AGT": "volume-04",
    "WF": "volume-04",
    "PROV": "volume-05",
    "AUTH": "volume-05",
    "TOOL": "volume-06",
    "MCP": "volume-06",
    "SKILL": "volume-06",
    "PLUG": "volume-06",
    "SDK": "volume-06",
    "MEM": "volume-07",
    "CTX": "volume-07",
    "IDX": "volume-07",
    "UX": "volume-08",
    "CLI": "volume-08",
    "TUI": "volume-08",
    "SEC": "volume-09",
    "CFG": "volume-10",
    "OBS": "volume-10",
    "GIT": "volume-11",
    "GH": "volume-11",
    "PERF": "volume-12",
    "TEST": "volume-13",
    "REL": "volume-14",
}

PHASES = {"Core", "MVP", "Beta", "v1", "v2", "Future", "Out of Scope"}

FR_BULLETS = [
    "Type", "Status", "Priority", "Phase", "Source", "Owner",
    "Affected components", "Dependencies", "Related risks",
]
FR_SECTIONS = [
    "Description", "Motivation", "Actors", "Preconditions", "Main flow",
    "Alternative flows", "Edge cases", "Inputs", "Outputs", "States",
    "Errors", "Constraints", "Security", "Observability", "Performance",
    "Compatibility", "Acceptance criteria", "Verification method",
    "Traceability",
]
NFR_BULLETS = [
    "Category", "Priority", "Phase", "Metric", "Target", "Minimum threshold",
    "Measurement method", "Test environment", "Measurement frequency",
    "Owner", "Dependencies", "Risks", "Acceptance criteria",
]
RISK_BULLETS = [
    "Category", "Probability", "Impact", "Severity", "Mitigation",
    "Detection", "Owner", "Status",
]
ADR_BULLETS = ["Status", "Date", "Deciders", "Components", "Related requirements"]
ADR_SECTIONS = [
    "Context", "Problem", "Forces and constraints", "Alternatives considered",
    "Decision", "Rationale", "Positive consequences", "Negative consequences",
    "Risks", "Reversal plan", "Review conditions",
]

VAGUE_WORDS = [
    "fast", "easy", "intuitive", "robust", "scalable", "secure", "efficient",
    "modern", "advanced", "compatible", "powerful",
]
VAGUE_RE = re.compile(
    r"(?<![-\w])(" + "|".join(VAGUE_WORDS) + r")\b", re.IGNORECASE
)
NORMATIVE_RE = re.compile(r"\b(MUST NOT|MUST|SHOULD NOT|SHOULD|MAY)\b")
ALLOW_VAGUE = "<!-- lint:allow vague -->"

# Legacy product names that must never appear (extend if a rename ever occurs).
LEGACY_NAMES: list[str] = []
SPANISH_KEYWORDS = [
    "DEBE", "DEBERÁ", "DEBERÍA", "NO DEBE", "PUEDE",
    "PENDIENTE DE VALIDACIÓN", "FUERA DE ALCANCE", "DEPRECADO",
]
PLACEHOLDER_WORDS = ["TBD", "TODO", "lorem ipsum"]

MERMAID_TYPES = {
    "flowchart", "graph", "sequenceDiagram", "stateDiagram", "stateDiagram-v2",
    "classDiagram", "erDiagram", "journey", "gantt", "pie", "mindmap",
    "timeline", "quadrantChart", "gitGraph", "block-beta",
}

SKIP_DIRS = {"_spine", "_audit"}

ID_DEF_RE = re.compile(
    r"^#{1,4}\s+((?:FR|NFR|RISK|E)-[A-Z]{2,6}-\d{3}|(?:ADR|PRD)-\d{3})\s+—\s+\S"
)
ID_HEADING_LOOSE_RE = re.compile(
    r"^#{1,4}\s+((?:FR|NFR|RISK|E)-[A-Z]{2,6}-\d{3}|(?:ADR|PRD)-\d{3})\b"
)
ID_REF_RE = re.compile(
    r"\b((?:FR|NFR|RISK|E)-[A-Z]{2,6}-\d{3}|(?:ADR|PRD)-\d{3})\b"
)
LINK_RE = re.compile(r"!?\[[^\]]*\]\(([^)\s]+)\)")


@dataclass
class Finding:
    severity: str  # "error" | "warning"
    file: str
    line: int
    check: str
    message: str


@dataclass
class Document:
    path: Path
    rel: str
    lines: list[str]
    prose: list[tuple[int, str]] = field(default_factory=list)  # outside fences
    fences: list[tuple[int, str, list[str]]] = field(default_factory=list)


def parse_document(path: Path, root: Path, findings: list[Finding]) -> Document:
    rel = str(path.relative_to(root))
    lines = path.read_text(encoding="utf-8").splitlines()
    doc = Document(path=path, rel=rel, lines=lines)
    fence_open: int | None = None
    fence_info = ""
    fence_ticks = 0
    body: list[str] = []
    for i, line in enumerate(lines, start=1):
        stripped = line.lstrip()
        m = re.match(r"^(`{3,})(.*)$", stripped)
        if m:
            ticks, info = len(m.group(1)), m.group(2).strip()
            if fence_open is None:
                fence_open, fence_info, fence_ticks, body = i, info, ticks, []
                if not info:
                    findings.append(Finding(
                        "error", rel, i, "fence-language",
                        "fenced code block without a language tag"))
            elif ticks >= fence_ticks and not info:
                doc.fences.append((fence_open, fence_info, body))
                fence_open = None
            else:
                body.append(line)
            continue
        if fence_open is not None:
            body.append(line)
        else:
            doc.prose.append((i, line))
    if fence_open is not None:
        findings.append(Finding(
            "error", rel, fence_open, "fence-unclosed",
            "unclosed fenced code block"))
    return doc


def block_of(doc: Document, def_line: int) -> list[tuple[int, str]]:
    """Prose lines from a `### ID —` definition to the next ###-or-higher heading."""
    out: list[tuple[int, str]] = []
    started = False
    def_level = len(doc.lines[def_line - 1]) - len(doc.lines[def_line - 1].lstrip("#"))
    def_level = len(re.match(r"^#+", doc.lines[def_line - 1].strip()).group(0))
    for i, line in doc.prose:
        if i == def_line:
            started = True
            continue
        if not started:
            continue
        m = re.match(r"^(#{1,6})\s", line)
        if m and len(m.group(1)) <= def_level:
            break
        out.append((i, line))
    return out


def check_block_template(
    doc: Document, def_line: int, ident: str, bullets: list[str],
    sections: list[str], findings: list[Finding],
) -> None:
    block = block_of(doc, def_line)
    text_lines = [l for _, l in block]
    for b in bullets:
        pat = re.compile(rf"^\s*-\s+\*{{0,2}}{re.escape(b)}\*{{0,2}}\s*:")
        if not any(pat.match(l) for l in text_lines):
            findings.append(Finding(
                "error", doc.rel, def_line, "template",
                f"{ident}: missing metadata bullet '- {b}:'"))
    for s in sections:
        pat = re.compile(rf"^####\s+{re.escape(s)}\s*$")
        if not any(pat.match(l) for l in text_lines):
            findings.append(Finding(
                "error", doc.rel, def_line, "template",
                f"{ident}: missing section '#### {s}'"))
    # Phase value + non-empty acceptance/verification for FR/NFR
    for i, l in block:
        m = re.match(r"^\s*-\s+\*{0,2}Phase\*{0,2}\s*:\s*(.*)$", l)
        if m:
            phase = m.group(1).strip().rstrip(".")
            if phase not in PHASES:
                findings.append(Finding(
                    "error", doc.rel, i, "phase",
                    f"{ident}: Phase '{phase}' not in {sorted(PHASES)}"))
    if sections:  # FR/ADR-style: sections must be non-empty
        for s in ("Acceptance criteria", "Verification method"):
            if s in sections and _section_empty(text_lines, s):
                findings.append(Finding(
                    "error", doc.rel, def_line, "empty-section",
                    f"{ident}: section '#### {s}' is empty"))
    else:  # NFR-style bullets must be non-empty
        for b in ("Acceptance criteria", "Metric", "Measurement method"):
            if b in bullets:
                pat = re.compile(
                    rf"^\s*-\s+\*{{0,2}}{re.escape(b)}\*{{0,2}}\s*:\s*(.+)$")
                if not any(pat.match(l) and pat.match(l).group(1).strip()
                           for l in text_lines):
                    findings.append(Finding(
                        "error", doc.rel, def_line, "empty-field",
                        f"{ident}: bullet '- {b}:' is empty"))


def _section_empty(lines: list[str], section: str) -> bool:
    try:
        start = next(i for i, l in enumerate(lines)
                     if re.match(rf"^####\s+{re.escape(section)}\s*$", l))
    except StopIteration:
        return False  # missing is reported separately
    for l in lines[start + 1:]:
        if re.match(r"^####\s+", l) or re.match(r"^#{1,3}\s", l):
            return True
        if l.strip():
            return False
    return True


def owner_dir_for(ident: str) -> str | None:
    if ident.startswith("ADR-"):
        return "annexes/adr"
    if ident.startswith("PRD-"):
        return AREA_OWNERS["PRD"]
    m = re.match(r"^(?:FR|NFR|RISK|E)-([A-Z]{2,6})-", ident)
    if m:
        return AREA_OWNERS.get(m.group(1))
    return None


def volume_of(rel: str) -> str | None:
    m = re.match(r"(volume-\d{2})", rel)
    return m.group(1) if m else None


def lint(root: Path) -> list[Finding]:
    findings: list[Finding] = []
    docs: list[Document] = []
    for path in sorted(root.rglob("*.md")):
        relparts = path.relative_to(root).parts
        if relparts and relparts[0] in SKIP_DIRS:
            continue
        docs.append(parse_document(path, root, findings))

    definitions: dict[str, tuple[str, int]] = {}

    # Pass 1: definitions, malformed headings, templates
    for doc in docs:
        for i, line in doc.prose:
            loose = ID_HEADING_LOOSE_RE.match(line.strip())
            if not loose:
                continue
            ident = loose.group(1)
            if not ID_DEF_RE.match(line.strip()):
                findings.append(Finding(
                    "error", doc.rel, i, "def-format",
                    f"heading for {ident} must use '### {ident} — Name'"))
                continue
            if ident in definitions:
                prev = definitions[ident]
                findings.append(Finding(
                    "error", doc.rel, i, "duplicate-id",
                    f"{ident} already defined at {prev[0]}:{prev[1]}"))
            else:
                definitions[ident] = (doc.rel, i)
            owner = owner_dir_for(ident)
            if owner is None:
                findings.append(Finding(
                    "error", doc.rel, i, "unknown-area",
                    f"{ident}: area not in ownership table"))
            elif not doc.rel.startswith(owner):
                findings.append(Finding(
                    "error", doc.rel, i, "ownership",
                    f"{ident} defined outside owning directory '{owner}/'"))
            if ident.startswith("FR-"):
                check_block_template(doc, i, ident, FR_BULLETS, FR_SECTIONS, findings)
            elif ident.startswith("NFR-"):
                check_block_template(doc, i, ident, NFR_BULLETS, [], findings)
            elif ident.startswith("RISK-"):
                check_block_template(doc, i, ident, RISK_BULLETS, [], findings)
            elif ident.startswith("ADR-"):
                check_block_template(doc, i, ident, ADR_BULLETS, ADR_SECTIONS, findings)

    # Pass 2: references, vague terms, denylist, links, fences content
    for doc in docs:
        for i, line in doc.prose:
            if line.lstrip().startswith(">"):
                continue
            for m in ID_REF_RE.finditer(line):
                ident = m.group(1)
                if ident not in definitions:
                    findings.append(Finding(
                        "error", doc.rel, i, "unresolved-ref",
                        f"reference to undefined identifier {ident}"))
            if ALLOW_VAGUE not in line and NORMATIVE_RE.search(line):
                vm = VAGUE_RE.search(line)
                if vm and not re.search(r"\d", line):
                    findings.append(Finding(
                        "warning", doc.rel, i, "vague-term",
                        f"vague term '{vm.group(1)}' in normative statement "
                        "without metric"))
            for word in SPANISH_KEYWORDS + LEGACY_NAMES:
                if re.search(rf"(?<![\w]){re.escape(word)}(?![\w])", line):
                    findings.append(Finding(
                        "error", doc.rel, i, "denylist",
                        f"forbidden string '{word}'"))
            for word in PLACEHOLDER_WORDS:
                if re.search(rf"\b{re.escape(word)}\b", line):
                    findings.append(Finding(
                        "error", doc.rel, i, "placeholder",
                        f"placeholder '{word}' is not allowed"))
            for lm in LINK_RE.finditer(line):
                target = lm.group(1)
                if target.startswith(("http://", "https://", "#", "mailto:")):
                    continue
                target_path = (doc.path.parent / target.split("#")[0]).resolve()
                if not target_path.exists():
                    findings.append(Finding(
                        "error", doc.rel, i, "broken-link",
                        f"relative link target not found: {target}"))

        for start, info, body in doc.fences:
            kind = info.split()[0] if info else ""
            invalid = "invalid" in info.split()[1:]
            content = "\n".join(body)
            if kind == "toml" and not invalid:
                try:
                    tomllib.loads(content)
                except tomllib.TOMLDecodeError as e:
                    findings.append(Finding(
                        "error", doc.rel, start, "toml-parse",
                        f"TOML block does not parse: {e}"))
            elif kind == "json" and not invalid:
                try:
                    json.loads(content)
                except json.JSONDecodeError as e:
                    findings.append(Finding(
                        "error", doc.rel, start, "json-parse",
                        f"JSON block does not parse: {e}"))
            elif kind == "mermaid":
                first = next((l.strip() for l in body if l.strip()), "")
                dtype = first.split()[0] if first else ""
                if dtype not in MERMAID_TYPES:
                    findings.append(Finding(
                        "warning", doc.rel, start, "mermaid-type",
                        f"unknown mermaid diagram type '{dtype}'"))
                for a, b in (("(", ")"), ("[", "]"), ("{", "}")):
                    if content.count(a) != content.count(b):
                        findings.append(Finding(
                            "warning", doc.rel, start, "mermaid-brackets",
                            f"unbalanced '{a}{b}' in mermaid block"))
                        break

    # Pass 3: volume registers and PENDING VALIDATION bookkeeping
    by_volume: dict[str, set[str]] = {}
    for ident, (rel, _line) in definitions.items():
        if ident.startswith(("FR-", "NFR-", "RISK-")):
            vol = volume_of(rel)
            if vol:
                by_volume.setdefault(vol, set()).add(ident)
    for doc in docs:
        if not doc.rel.endswith("99-volume-register.md"):
            continue
        vol = volume_of(doc.rel)
        listed = {m.group(1) for _, l in doc.prose for m in ID_REF_RE.finditer(l)
                  if m.group(1).startswith(("FR-", "NFR-", "RISK-"))}
        defined = by_volume.get(vol, set())
        for ident in sorted(defined - listed):
            findings.append(Finding(
                "error", doc.rel, 1, "register-missing",
                f"{ident} defined in {vol} but absent from its register"))
        for ident in sorted(listed - defined):
            findings.append(Finding(
                "error", doc.rel, 1, "register-orphan",
                f"{ident} listed in register but not defined in {vol}"))
    for vol, defined in sorted(by_volume.items()):
        if vol == "volume-00":
            continue
        if not any(d.rel == f"{vol}/99-volume-register.md" or
                   d.rel.startswith(f"{vol}/") and d.rel.endswith("99-volume-register.md")
                   for d in docs):
            findings.append(Finding(
                "error", f"{vol}/", 1, "register-absent",
                f"{vol} defines requirements but has no 99-volume-register.md"))

    oq = next((d for d in docs
               if d.rel.endswith("08-register-open-questions.md")), None)
    oq_text = "\n".join(l for _, l in oq.prose) if oq else ""
    for doc in docs:
        if doc.rel.endswith(("08-register-open-questions.md",)):
            continue
        pv_lines = [i for i, l in doc.prose if "PENDING VALIDATION" in l]
        if not pv_lines:
            continue
        vol = volume_of(doc.rel)
        register = next(
            (d for d in docs
             if vol and d.rel == f"{vol}/99-volume-register.md"), None)
        covered = ("PENDING VALIDATION" in oq_text) or (
            register and any("PENDING VALIDATION" in l or "Open questions" in l
                             for _, l in register.prose))
        if not covered:
            findings.append(Finding(
                "warning", doc.rel, pv_lines[0], "pending-validation",
                "PENDING VALIDATION used but no open-questions entry found"))

    return findings


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--root", type=Path,
                    default=Path(__file__).resolve().parents[1] / "docs" / "spec")
    ap.add_argument("--errors-only", action="store_true")
    ap.add_argument("--json", action="store_true", dest="as_json")
    ap.add_argument("--strict", action="store_true",
                    help="exit 1 on warnings too")
    args = ap.parse_args()

    if not args.root.is_dir():
        print(f"spec root not found: {args.root}", file=sys.stderr)
        return 2

    findings = lint(args.root)
    errors = [f for f in findings if f.severity == "error"]
    warnings = [f for f in findings if f.severity == "warning"]
    shown = errors if args.errors_only else findings

    if args.as_json:
        print(json.dumps(
            {"errors": len(errors), "warnings": len(warnings),
             "findings": [f.__dict__ for f in shown]}, indent=2))
    else:
        for f in sorted(shown, key=lambda f: (f.file, f.line)):
            print(f"{f.severity.upper():7} {f.file}:{f.line} [{f.check}] {f.message}")
        print(f"\n{len(errors)} error(s), {len(warnings)} warning(s) "
              f"across {args.root}")

    if errors or (args.strict and warnings):
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
