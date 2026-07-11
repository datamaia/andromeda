# 09 — Master Traceability Matrix

Maps every product objective (`PRD-NNN`, defined in Volume 1, chapter
[04](../volume-01-vision-and-product/04-goals-non-goals-principles.md)) to the
requirements that realize it, and summarizes what every volume minted. Generated at
consolidation from the volumes' `99-volume-register.md` files, whose Requirements index
tables the spec linter cross-checks against the defining chapters. The corpus holds 13
objectives and 482 requirements (266 FR, 102 NFR, 114 RISK).

## Objective → requirements

Mapping rule: a requirement realizes an objective when its own definition cites that
objective — for an FR, a `PRD` citation in its `#### Traceability` section (present in
every FR by template); for an NFR, whose bullet template carries no Traceability
section, a citation anywhere in its definition block. The table therefore lists the
principal (direct, self-declared) realizers, not the transitive closure: a requirement
that supports an objective only through a dependency on a listed requirement is not
repeated. Risks never realize objectives and are indexed in chapter
[07](07-register-risks.md). Identifiers appear in register order (volume, then
authoring order); every objective has at least four direct realizers.

| Objective | Title | Realizing requirements | Volumes |
|---|---|---|---|
| [PRD-001](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Agent-Native Engineering Harness | FR-AGT-001, FR-PROV-013, FR-PROV-043, FR-CLI-002, FR-CLI-013 | 4, 5, 8 |
| [PRD-002](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Vendor and Model Independence | FR-ARCH-001, FR-ARCH-002, FR-PROV-001, FR-PROV-002, FR-PROV-010, FR-PROV-011, FR-PROV-012, FR-PROV-020, FR-PROV-021, FR-PROV-022, FR-PROV-042, FR-PROV-050, FR-AUTH-001, FR-AUTH-002, FR-AUTH-003, FR-AUTH-004, FR-AUTH-005, FR-AUTH-006, FR-AUTH-008, FR-PROV-080, FR-PROV-081, FR-PROV-082, FR-MCP-001, FR-CTX-001, FR-CTX-002, FR-IDX-003, FR-CLI-014 | 3, 5, 6, 7, 8 |
| [PRD-003](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Local-First Operation | FR-PROV-001, FR-PROV-012, FR-PROV-081, FR-PROV-083, FR-PROV-084, FR-PROV-085, FR-TOOL-007, FR-MEM-001, FR-MEM-006, FR-MEM-007, FR-MEM-010, FR-IDX-001, FR-IDX-003, FR-IDX-005, FR-CLI-015, FR-TUI-063, FR-UX-074, FR-CFG-001, FR-CFG-002, FR-OBS-010, FR-OBS-012, FR-PERF-003, NFR-PERF-024, FR-REL-004 | 5, 6, 7, 8, 10, 12, 14 |
| [PRD-004](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Tool-First Execution | FR-ARCH-011, FR-AGT-001, FR-AGT-010, FR-PROV-021, FR-TOOL-001, FR-TOOL-002, FR-TOOL-005, FR-TOOL-007, FR-TOOL-008, FR-MCP-001, FR-PLUG-001, FR-CLI-014, FR-GIT-001, FR-GIT-009 | 3, 4, 5, 6, 8, 11 |
| [PRD-005](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Safe Agent Autonomy | FR-ARCH-004, FR-ARCH-005, FR-AGT-001, FR-AGT-004, FR-AGT-005, FR-AGT-007, FR-AGT-008, FR-AGT-009, FR-AGT-010, FR-AGT-011, FR-AGT-012, FR-WF-001, FR-WF-004, FR-WF-007, FR-PROV-040, FR-PROV-041, FR-PROV-043, FR-AUTH-002, FR-AUTH-009, FR-AUTH-011, FR-TOOL-001, FR-TOOL-005, FR-TOOL-006, FR-TOOL-008, FR-MCP-006, FR-PLUG-003, FR-MEM-003, FR-MEM-004, FR-MEM-008, FR-CTX-004, FR-CTX-005, FR-CLI-002, FR-CLI-004, FR-CLI-009, FR-CLI-010, FR-CLI-013, FR-CLI-015, FR-TUI-003, FR-UX-070, FR-UX-072, FR-SEC-100, FR-SEC-101, FR-SEC-102, FR-SEC-104, FR-SEC-105, FR-SEC-111, FR-SEC-112, FR-SEC-113, FR-CFG-011, FR-OBS-004, FR-GIT-001, FR-GIT-009, FR-PERF-001, FR-PERF-006, FR-REL-005, FR-REL-006, FR-REL-007 | 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14 |
| [PRD-006](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Transparent and Auditable Operation | FR-ARCH-007, FR-AGT-001, FR-AGT-002, FR-AGT-004, FR-AGT-007, FR-AGT-008, FR-AGT-009, FR-AGT-010, FR-AGT-013, FR-AGT-014, FR-AGT-015, FR-WF-001, FR-WF-003, FR-WF-007, FR-WF-008, FR-PROV-001, FR-PROV-011, FR-PROV-013, FR-PROV-030, FR-PROV-031, FR-PROV-040, FR-PROV-042, FR-PROV-043, FR-PROV-050, FR-AUTH-008, FR-AUTH-009, FR-AUTH-011, FR-TOOL-003, FR-TOOL-004, FR-TOOL-008, FR-MCP-006, FR-PLUG-003, FR-MEM-001, FR-MEM-002, FR-MEM-003, FR-MEM-004, FR-MEM-005, FR-MEM-006, FR-MEM-008, FR-MEM-009, FR-CTX-001, FR-CTX-003, FR-CTX-004, FR-CTX-005, FR-CTX-006, FR-CTX-007, FR-IDX-001, FR-CLI-006, FR-CLI-008, FR-CLI-013, FR-CLI-015, FR-CLI-016, FR-UX-001, FR-UX-003, FR-TUI-007, FR-TUI-009, FR-TUI-060, FR-TUI-063, FR-TUI-064, FR-UX-072, FR-UX-074, FR-UX-075, FR-SEC-100, FR-SEC-101, FR-SEC-102, FR-SEC-104, FR-SEC-111, FR-SEC-112, FR-SEC-113, FR-CFG-005, FR-CFG-011, FR-OBS-001, FR-OBS-002, FR-OBS-004, FR-OBS-006, FR-OBS-007, FR-OBS-009, FR-GIT-001, FR-GIT-009, FR-GH-001, FR-PERF-003, FR-PERF-006, FR-REL-006, FR-REL-016 | 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 14 |
| [PRD-007](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Extensible Open Platform | FR-ARCH-001, FR-ARCH-002, FR-ARCH-003, FR-ARCH-011, FR-AGT-013, FR-WF-002, FR-WF-008, FR-PROV-002, FR-TOOL-001, FR-TOOL-003, FR-TOOL-004, FR-SDK-001, FR-MCP-001, FR-SKILL-001, FR-PLUG-001, FR-PLUG-005, FR-MEM-009, FR-CLI-004, FR-CLI-014 | 3, 4, 5, 6, 7, 8 |
| [PRD-008](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | First-Class Terminal Experience | FR-ARCH-004, FR-ARCH-006, FR-AGT-002, FR-WF-009, FR-PROV-020, FR-TOOL-006, FR-TOOL-007, FR-CTX-006, FR-CLI-001, FR-CLI-003, FR-CLI-005, FR-CLI-007, FR-CLI-008, FR-CLI-012, FR-CLI-013, FR-CLI-016, FR-UX-002, FR-UX-003, FR-TUI-001, FR-TUI-002, FR-TUI-003, FR-TUI-004, FR-TUI-005, FR-TUI-006, FR-TUI-007, FR-TUI-008, FR-TUI-009, FR-UX-040, FR-UX-041, FR-UX-042, FR-UX-043, FR-TUI-060, FR-TUI-061, FR-TUI-062, FR-TUI-063, FR-TUI-065, FR-TUI-066, FR-TUI-067, FR-UX-070, FR-UX-071, FR-UX-073, FR-UX-074, FR-UX-075, FR-UX-076, FR-CFG-005, FR-GIT-003, FR-PERF-002 | 3, 4, 5, 6, 7, 8, 10, 11, 12 |
| [PRD-009](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Human and Automation Parity | FR-ARCH-007, FR-ARCH-008, FR-AGT-001, FR-AGT-009, FR-WF-003, FR-WF-004, FR-PROV-022, FR-PROV-043, FR-AUTH-004, FR-TOOL-005, FR-MEM-004, FR-CLI-001, FR-CLI-002, FR-CLI-005, FR-CLI-006, FR-CLI-007, FR-CLI-009, FR-CLI-011, FR-CLI-013, FR-CLI-016, FR-TUI-001, FR-TUI-065, FR-TUI-068, FR-SEC-100, FR-SEC-105, FR-CFG-001, FR-CFG-005, FR-OBS-011, FR-REL-005, FR-REL-007 | 3, 4, 5, 6, 7, 8, 9, 10, 14 |
| [PRD-010](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Durable Sessions and Recoverable Work | FR-ARCH-004, FR-ARCH-006, FR-ARCH-009, FR-ARCH-010, FR-AGT-001, FR-AGT-003, FR-AGT-006, FR-AGT-008, FR-AGT-011, FR-AGT-012, FR-AGT-015, NFR-AGT-002, FR-WF-005, FR-WF-006, FR-WF-009, FR-WF-010, FR-AUTH-010, FR-TOOL-008, FR-MEM-001, FR-MEM-007, FR-CTX-005, FR-CTX-007, FR-IDX-004, FR-IDX-006, FR-CLI-013, FR-TUI-064, FR-UX-076, FR-SEC-113, FR-CFG-001, FR-CFG-009, FR-OBS-006, FR-PERF-001, FR-PERF-004, FR-REL-006, FR-REL-008, FR-REL-016 | 3, 4, 5, 6, 7, 8, 9, 10, 12, 14 |
| [PRD-011](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Portable Unix-First Platform | FR-ARCH-001, FR-ARCH-005, FR-PORT-001, FR-PORT-002, FR-PORT-003, FR-PORT-004, FR-AGT-006, FR-TUI-067, FR-TUI-068, FR-CFG-002, FR-REL-003, FR-REL-009 | 3, 4, 8, 10, 14 |
| [PRD-012](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Specification-Driven Workflows | FR-WF-001, FR-WF-002, FR-WF-003, FR-WF-005 | 4 |
| [PRD-013](../volume-01-vision-and-product/04-goals-non-goals-principles.md) | Sustainable Open-Source Delivery | FR-GH-001, FR-GH-002, FR-GH-004, FR-PERF-005, FR-REL-001, FR-REL-004 | 11, 12, 14 |

## Per-volume summary

Requirement and error-code counts aggregate each register's `## Requirements index`
and `## Error codes minted` sections; event and configuration-key counts are the
distinct names itemized in `## Events minted` and `## Config keys minted`. For
configuration keys, itemized keys of parameterized tables (for example
`[providers.<slug>]`, `[mcp.servers.<name>]`) and itemized fields of array-of-tables
entries count once each; open-ended families (per-server overrides of runtime-wide
defaults, per-plugin override tables, permission-rule resource qualifiers, the reserved
`[tui.keymap]` table) add nothing beyond their parent key. Volume 1 additionally
defines the 13 `PRD-*` objectives, which are not requirements and are counted nowhere
below; Volumes 2 and 15 mint none of the six kinds counted here (their registers carry
entities, canonical states, ADRs, glossary, and process material).

| Volume | FRs | NFRs | RISKs | E-codes | Events minted | Config keys minted |
|---|---|---|---|---|---|---|
| Volume 1 — Vision, Problem, Scope, and Product | 0 | 0 | 10 | 0 | 0 | 0 |
| Volume 2 — Domain Model | 0 | 0 | 0 | 0 | 0 | 0 |
| Volume 3 — System Architecture | 15 | 8 | 7 | 10 | 8 | 0 |
| Volume 4 — Agent Runtime | 25 | 6 | 7 | 24 | 83 | 29 |
| Volume 5 — Providers, Models, and Authentication | 33 | 7 | 9 | 30 | 39 | 60 |
| Volume 6 — Tools, MCP, Skills, and Plugins | 31 | 9 | 10 | 39 | 48 | 46 |
| Volume 7 — Memory, Context, and Indexing | 23 | 5 | 5 | 17 | 20 | 46 |
| Volume 8 — CLI and TUI | 48 | 10 | 11 | 17 | 16 | 19 |
| Volume 9 — Security | 14 | 6 | 27 | 14 | 23 | 21 |
| Volume 10 — Configuration, Storage, and Observability | 24 | 10 | 7 | 28 | 37 | 30 |
| Volume 11 — Git, GitHub, and Development Platforms | 22 | 4 | 7 | 21 | 20 | 21 |
| Volume 12 — Performance and Reliability | 6 | 28 | 4 | 4 | 7 | 0 |
| Volume 13 — Testing and Quality | 9 | 6 | 6 | 6 | 10 | 0 |
| Volume 14 — Distribution, Installation, and Updates | 16 | 3 | 4 | 12 | 11 | 12 |
| Volume 15 — Roadmap, Execution, and Governance | 0 | 0 | 0 | 0 | 0 | 0 |
| **Total** | **266** | **102** | **114** | **222** | **322** | **284** |

## Requirement → phase → verification

The full row-level table — every FR, NFR, and RISK in the corpus with its title,
phase, verification method, owning volume, and a link to its defining chapter — is the
[traceability matrix annex](../annexes/traceability-matrix.md), one table per volume,
generated from the same registers under the rule stated there.

## Continuation into development automation

Traceability does not stop at the specification boundary. The downstream chain —
requirement → epic → issue → branch → commit → pull request → test → check → release →
artifact — is operated and mechanically enforced on the development platform by
FR-GH-001 (traceability automation, with NFR-GH-001 as its completeness measure),
specified in Volume 11, chapter
[07](../volume-11-git-and-github/07-traceability-automation.md); the validators,
nightly audit, and per-release chain reports are defined there and are not restated
here.
