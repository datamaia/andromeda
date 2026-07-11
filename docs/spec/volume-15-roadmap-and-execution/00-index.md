# Volume 15 — Roadmap, Execution, and Governance

**Status:** Complete · **Owner:** Project governance / roadmap (Volume 15)

Volume 15 is the consolidation volume for execution: it sequences the specified product
inside the phases Volume 1 chapter 05 defines, decomposes the committed phases into epics
and milestones, orders the backlog, and defines the open-source governance under which the
work is delivered and maintained. This volume mints **no** requirement identifiers, no
error codes, no events, and no configuration keys — every capability it schedules is
defined and owned elsewhere; this volume aggregates and references. Phase definitions
remain Volume 1's; requirement phases remain the minting volumes'; this volume records
drift as open questions instead of editing foreign registers.

Foundations assumed: Volume 0 (conventions, identifier taxonomy, priorities P0–P3, change
procedure), Volume 1 (phase definitions, MVP minimum, personas, success metrics), the
volume registers of Volumes 1–14 (the requirement index and phase column of each register
is this volume's primary input), Volume 11 (repository structure, issue and label
taxonomy, Projects operation, CI), Volume 13 (gate tiers and release qualification),
Volume 14 (release semantics, support windows, deprecation), ADR-002 (Apache-2.0),
ADR-015 (SemVer and commit conventions).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Phase Plan](01-phase-plan.md) | The consolidated capability set per phase (Core, MVP, Beta, v1, v2, Future, Out of Scope) aggregated from every volume register; the MVP conformance check against the Volume 1 chapter 05 twenty-seven-item minimum with drift flagged; per-phase entry criteria, exit criteria, and quality gates |
| [02 — Epics, Milestones, and Sequencing](02-epics-milestones-sequencing.md) | MVP and Beta decomposed into epics (chapter-local EP-NN labels) mapped to requirement ID sets; the epic dependency graph; the implementation sequence honoring architectural dependencies; milestones; Definition of Ready and Definition of Done |
| [03 — Backlog and Prioritization](03-backlog-and-prioritization.md) | The prioritization model (Volume 0 P0–P3 semantics crossed with phase); the v1, v2, and Future backlog by area; risk-driven ordering rules bound to existing RISK-* identifiers |
| [04 — Open-Source Governance](04-open-source-governance.md) | Roles, decision-making and conflict resolution, RFC and ADR processes, security disclosure policy, Code of Conduct, contribution guide outline, release authority, trademark policy, plugin/extension policy, deprecation and support policy, community channels, bus-factor mitigation |
| [05 — Adoption and Maintenance](05-adoption-and-maintenance.md) | Adoption plan mapping the Volume 1 personas to onboarding paths; maintenance plan (dependency updates, security patches, benchmark regression watch); sustainability and project health metrics |
| [99 — Volume Register](99-volume-register.md) | Machine-readable register: empty requirements index (this volume mints nothing), glossary additions, assumptions, open questions, cross-volume references |

## Reading guide

1. Chapter 01 answers *what ships when*: it is the corpus-wide phase view auditors and
   phase gates consume.
2. Chapter 02 answers *in what order and by what unit of work*: epics, dependencies,
   milestones, and the working definitions of ready and done.
3. Chapter 03 answers *what comes next and why*: the ordered backlog beyond Beta and the
   rules that let risk reorder it.
4. Chapters 04 and 05 answer *under whose rules and for how long*: governance of the
   project and the plan that keeps the product adopted, patched, and measured after v1.
