# Andromeda Engineering Product Specification

**Version:** 0.1.0-draft · **Date:** 2026-07-11 · **Status:** In authoring

This document set is the primary source of truth for **Andromeda**, an open-source, local-first,
vendor-agnostic AI agent engineering platform (CLI + TUI). It combines the roles of a PRD, SRS,
Software Architecture Document, Security Architecture Document, UX/CLI/TUI Specification, API and
SDK Specification, Test Strategy, Release Engineering Plan, Open Source Governance Model, and
Operations & Observability Specification. It is written to be implementable by AI agents under
human supervision and maintainable, extensible, testable, auditable, and operable by human
engineers.

Normative conventions, identifier taxonomy, templates, and the change procedure are defined in
[Volume 0](volume-00-conventions/00-index.md) and bind every other volume.

## Volumes

| Volume | Title | Status |
|---|---|---|
| [0](volume-00-conventions/00-index.md) | Conventions and Document Control | In authoring |
| [1](volume-01-vision-and-product/00-index.md) | Vision, Problem, Scope, and Product | Pending |
| [2](volume-02-domain-model/00-index.md) | Domain Model | Pending |
| [3](volume-03-architecture/00-index.md) | System Architecture | Pending |
| [4](volume-04-agent-runtime/00-index.md) | Agent Runtime | Pending |
| [5](volume-05-providers-and-auth/00-index.md) | Providers, Models, and Authentication | Pending |
| [6](volume-06-tools-mcp-skills-plugins/00-index.md) | Tools, MCP, Skills, and Plugins | Pending |
| [7](volume-07-memory-context-indexing/00-index.md) | Memory, Context, and Indexing | Pending |
| [8](volume-08-cli-and-tui/00-index.md) | CLI and TUI | Pending |
| [9](volume-09-security/00-index.md) | Security | Pending |
| [10](volume-10-config-storage-observability/00-index.md) | Configuration, Storage, and Observability | Pending |
| [11](volume-11-git-and-github/00-index.md) | Git, GitHub, and Development Platforms | Pending |
| [12](volume-12-performance-and-reliability/00-index.md) | Performance and Reliability | Pending |
| [13](volume-13-testing-and-quality/00-index.md) | Testing and Quality | Pending |
| [14](volume-14-distribution/00-index.md) | Distribution, Installation, and Updates | Pending |
| [15](volume-15-roadmap-and-execution/00-index.md) | Roadmap and Execution | Pending |
| [Annexes](annexes/00-index.md) | Glossary, ADRs, Catalogs, Matrices, Checklists | Pending |

## Reading order

1. Volume 0 (conventions) — mandatory before contributing to or modifying the specification.
2. Volume 1 (product) → Volume 2 (domain) → Volume 3 (architecture) — the foundation.
3. Volumes 4–14 — subsystem specifications; each is self-contained given Volumes 0–3.
4. Volume 15 — phased roadmap and execution plan.
5. Annexes — consolidated reference material (ADR bodies, catalogs, matrices, checklists).

## Verification

Run the specification linter from the repository root:

```bash
python3 scripts/spec_lint.py            # full report
python3 scripts/spec_lint.py --errors-only
python3 scripts/spec_lint.py --json     # machine-readable output
```

The linter enforces the identifier taxonomy, template completeness, cross-reference resolution,
and embedded-example validity defined in Volume 0.
