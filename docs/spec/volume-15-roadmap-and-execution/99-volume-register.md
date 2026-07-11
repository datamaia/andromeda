# 99 — Volume 15 Register

Machine-parseable register of everything Volume 15 minted, per Volume 0 chapters 02 and
03. Merged into the Volume 0 registers at consolidation.

## Requirements index

Volume 15 mints no requirement identifiers — it consolidates phases and governance. Every
capability it sequences is defined and phased by its owning volume (Volumes 1–14
registers); every governance obligation binds through the change procedure and the
development-platform mechanics of Volume 11, not through new requirement identifiers.

| ID | Title | Phase | Verification method |
|---|---|---|---|
| — | *(none minted by this volume)* | — | — |

Local (non-corpus) identifier series used by this volume: `EP-NN` (epics, chapter 02),
`MS-N` (milestones, chapter 02), `W-N`/"wave" (iteration waves, chapter 02), and `R1`–`R9`
(risk-driven ordering rules, chapter 03). These are chapter-local planning labels, not
corpus identifiers; no other volume may mint them, and references to them from other
volumes name this volume and chapter.

## ADRs minted

None. The block available to this volume (numbers 205 and up, per Volume 0 chapter 03) is
untouched: no genuinely new architecture-level decision arose. Governance processes,
role rules, and the trademark policy are governance policy owned by chapter 04 of this
volume under the change procedure; the decisions this volume relies on are inherited and
referenced — ADR-002 (license), ADR-015 (versioning and commit conventions), ADR-013
(release tooling), ADR-193 (support windows), ADR-080/ADR-081 (extension distribution
and signatures), ADR-160/ADR-161 (benchmark methodology and gating).

## Error codes minted

None. Volume 15 defines no runtime behavior.

## Events minted

None. Development-process automation reports through CI check results and audit-filed
issues (Volume 11 convention), not through the product event bus.

## Config keys minted

None. Volume 15 owns no TOML table (Volume 0 chapter 03 configuration-table ownership).

## Glossary additions

| Term | One-line meaning |
|---|---|
| Epic (EP-NN) | A chapter-local Volume 15 planning label for a requirement-ID fan-out unit, realized on the platform as an issue with the `epic` label and a child task list (Volume 15, chapter 02). |
| Milestone (MS-N) | A chapter-local Volume 15 label for a named delivery checkpoint with an exit condition, mapped to a GitHub milestone (Volume 15, chapter 02). |
| Wave | A planning grouping of epics into iteration order consistent with the epic dependency graph; guidance, not a normative constraint (Volume 15, chapter 02). |
| Definition of Ready | The Volume 15 chapter 02 checklist an issue satisfies before entering an iteration (traceable, specified, bounded, unblocked, classified). |
| Definition of Done | The Volume 15 chapter 02 checklist a change satisfies before merge (reviewed, tested, gated, observable, documented, recorded). |
| Risk-driven ordering rule (R1–R9) | One of the standing Volume 15 chapter 03 rules by which open RISK-* entries reorder backlog work. |
| Escalation roster | The ordered maintainer list in `MAINTAINERS.md` that reassigns an unacknowledged security report after 2 business days (Volume 15, chapter 04). |
| Release manager | The maintainer named on a release issue who drives the Volume 14 pipeline and owns the qualification evidence for that release (Volume 15, chapter 04). |
| RFC | The Volume 15 chapter 04 request-for-comments process required before implementing changes to public contracts, MAJOR document changes, or governance changes; accepted RFCs conclude as ADRs. |
| Project health metrics | The consolidated Volume 15 chapter 05 metric checklist (feedback latency, traceability, coverage/mutation, flake, disclosure response, backport depth, contributor funnel, provenance distribution, ecosystem pulse, benchmark headroom) reviewed at every phase gate. |

## Assumptions

Local list per Volume 0, chapter 05 (global `AS-NNN`/`HY-NNN` numbers are minted at
consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | The GitHub platform features this volume's governance operates through (private vulnerability reporting, Discussions, protected environments, Projects, milestone automation) remain available to open-source projects on workable terms — mirrors the Volume 1 and Volume 11 platform assumptions | Phase-gate reviews of platform terms; Volume 11 platform audit | Governance mechanics re-bind to the successor platform per the Volume 11 migration path; the role/decision model of chapter 04 is platform-independent by design |
| Product hypothesis | A small maintainer set can keep every published support and disclosure promise (ADR-193 windows, 3-business-day first response) through v1 without burnout | Chapter 05 health metrics at each phase gate: backport queue depth, patch latency, disclosure-response tracking | Narrow the support window further (announced per chapter 04), pause optional catalog tranches (rule R9), and prioritize maintainer growth before feature work |
| Product hypothesis | The Contributor Covenant 2.1 bindings of chapter 04 fit the project's moderation capacity and community scale | Conduct-report handling reviews at the quarterly process audits | Amend the Code of Conduct bindings via a governance change (two-thirds decision) |
| Product hypothesis | Two-week iterations and the chapter 02 wave structure fit a part-time, small-team contribution cadence | Iteration completion rates and epic cycle times on the Roadmap project | Re-plan cadence and wave assignments — planning change only; the dependency graph, not the waves, is normative |

## Open questions

Entries follow Volume 0, chapter 08; global `OQ-NNN` numbers are minted at consolidation.
Rows exist for every PENDING VALIDATION occurrence in this volume and for every drift
finding chapter 01 mandates recording.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V15-OQ-1 | MVP widening drift: the aggregated MVP set exceeds the Volume 1 chapter 05 twenty-seven-item minimum in eight areas (memory and indexing subsystems, sandbox/audit breadth, IPC surface, traces/metrics/telemetry stack, provider cost/resilience breadth, TUI platform breadth, distribution breadth, performance operations). Should the non-load-bearing remainder re-phase to Beta? | Chapter 01, MVP conformance check (direction 2); RISK-PRD-004 | No — the minimum is a floor; drift is schedule exposure, not contradiction | Each owning volume confirms its MVP phasing as load-bearing for an MVP exit criterion or re-phases via the change procedure; disposition required before Beta entry (chapter 01 Beta entry criteria) | Open |
| V15-OQ-2 | Trademark registration for the Andromeda name and identity set — PENDING VALIDATION (chapter 04 applies the naming policy as project policy regardless of registration status) | Chapter 04, trademark policy | No — the policy binds as project policy either way | Organizational decision on pursuing registration, with legal review; outcome recorded via governance change and, if registered, an ADR from the 205+ block | Open |
| V15-OQ-3 | Release signing and macOS notarization enablement — PENDING VALIDATION, mirroring V1-OQ-1 (Volume 1 register) and V14-OQ-1 (Volume 14 register); chapter 01 carries it in MVP-minimum item 27 | Chapter 01, MVP conformance table, item 27 | No — checksums and provenance ship unconditionally | Owned by the Volume 1/Volume 14 register rows; this volume tracks it only as a phase-gate checklist item | Open |
| V15-OQ-4 | v2 confirmation hand-off: which marketplace/distribution expansions classified post-v1 by Volumes 6 and 14 enter confirmed v2 scope, given their own pending validations (official registry index per ADR-080; hosted package repositories per ADR-190) | Chapter 01 (v2 section); chapter 03 (v2 backlog) | No — v2 is a candidate phase confirmed at v1 ship | The change-procedure decision confirming v2 scope when v1 ships, consuming the resolutions of V6B-OQ-4 (Volume 6 register) and the Volume 14 register items | Open |

## Cross-volume references

Obligations other volumes place on Volume 15, and where this volume satisfies each:

| Expectation (source) | Satisfied in |
|---|---|
| Sequencing within the Volume 1 phases without redefining them (Volume 1 chapter 05) | Chapter 01 (phase plan), chapter 02 (epics and waves) |
| MVP minimum and phase gates consumed for sequencing and milestones (Volume 1 register) | Chapter 01 conformance check; chapter 02 milestones MS-6/MS-8 |
| Governance model behind the repository files `GOVERNANCE.md`/`MAINTAINERS.md` (Volume 11 chapter 03) | Chapter 04 |
| Disclosure governance, security-inbox process, and escalation roster behind NFR-SEC-003 (Volume 9 chapter 08) | Chapter 04 (security disclosure policy); chapter 05 (quarterly process audits) |
| Maintainer process and process audits verifying disclosure tracking (Volume 9 register) | Chapters 04 and 05 |
| Project health metrics consuming backport queue depth and patch latency (Volume 14 register, RISK-REL-004) | Chapter 05 (health metrics; maintenance plan) |
| PR provenance distribution as sustainability input (Volume 11 register) | Chapter 05 (health metrics) |
| Feedback channels for register hypotheses (Volume 3, Volume 5, Volume 7, Volume 14 registers; ADR-049) | Chapter 05 (feedback channels) |
| Phase gates consuming benchmark gate results (Volume 12 register) | Chapter 01 (phase-gate operation); chapter 05 (benchmark regression watch) |
| Release authority and support policy operation (Volume 14 chapters 01/04) | Chapter 04 (release authority; deprecation and support policy) |
| License and inbound-dependency policy (ADR-002) — referenced, never re-decided | Chapters 04 and 05 |

Load-bearing references this volume makes: phase definitions and MVP minimum (Volume 1
chapter 05); every Requirements index and Phase column of the Volume 1–14 registers
(chapter 01 aggregation); issue/label/Projects mechanics and repository files (Volume 11
chapters 03/05); gate tiers, waiver policy, and qualification evidence (Volume 13
chapters 01/04); release semantics, channels, deprecation, and support windows (Volume 14);
incident response and audit semantics (Volume 9 chapter 08); personas and jobs (Volume 1
chapter 02); success metrics SM-01 through SM-20 (Volume 1 chapter 06) as gate bindings.
