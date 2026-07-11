# Annexes — Index

The annexes are the consolidated reference layer of the specification corpus: full ADR
bodies, machine-oriented catalogs aggregated from the volume registers, cross-volume
matrices, and review checklists. Annex documents **aggregate and cite** — they mint no
requirement IDs, ADRs, error codes, events, or configuration keys, and they rename nothing.
Normative authority stays with the owning volumes (Volume 0, chapter 03 ownership rules);
where an annex and an owning chapter disagree, the chapter prevails and the annex is
corrected through the Volume 0 change procedure.

## ADR directory — [`adr/`](adr/)

One file per architecture decision record, `ADR-NNN.md`, in the full Volume 0 template. The
one-line index of all ADRs is [Volume 0, chapter 06](../volume-00-conventions/06-register-decisions.md).

**Count: 121 accepted ADRs.** Numbers are allocated in per-volume blocks (Volume 0,
chapter 03); unused numbers inside a block are permanent gaps unless minted by a later
amendment, so the numbering is intentionally non-contiguous:

| Block | Owner | Present | Gaps (reserved, unminted) |
|---|---|---|---|
| 001–039 | Foundation (Volumes 0–3; ADR-030..033 by Volume 3) | 001–033 | 034–039 |
| 040–054 | Volume 4 | 040–045, 048–054 | 046–047 |
| 055–069 | Volume 5 | 055–067 | 068–069 |
| 070–084 | Volume 6 | 070–074, 077–083 | 075–076, 084 |
| 085–099 | Volume 7 | 085–089 | 090–099 |
| 100–114 | Volume 8 | 100–114 | — |
| 115–129 | Volume 9 | 121–125 | 115–120, 126–129 |
| 130–144 | Volume 10 | 130–134, 137–141 | 135–136, 142–144 |
| 145–159 | Volume 11 | 145–149 | 150–159 |
| 160–174 | Volume 12 | 160–162 | 163–174 |
| 175–189 | Volume 13 | 175–177 | 178–189 |
| 190–204 | Volume 14 | 190–193 | 194–204 |
| 205 and up | Consolidation and later evolution | — | reserved |

## Catalogs

Six consolidated catalogs, aggregated from the `99-volume-register.md` files of Volumes 1–14
and kept mechanically checkable against the corpus (the error and event catalogs are the
consolidation targets named by Volume 0, chapter 03; ADR-016 and the event grammar bind
in-code definitions to them).

| File | Contents | Primary sources |
|---|---|---|
| `catalog-errors.md` | Every `E-<AREA>-NNN` error code with its ADR-016 envelope summary and exit-code mapping | Volume registers ("Error codes minted"); ADR-016 |
| `catalog-events.md` | Every event name (`<area>.<noun>.<verb-past>`), producer, and meaning; envelope semantics per Volume 10 | Volume registers ("Events minted"); FR-OBS-001 |
| `catalog-commands.md` | The full `andromeda` CLI command surface: syntax, flags, exit codes, JSON output schemas by reference | Volume 8, chapters 02–06 (FR-CLI-001) |
| `catalog-configuration.md` | Every `andromeda.toml` table and key with type, default, and owning volume; schema/precedence per Volume 10 | Volume registers ("Config keys minted"); FR-CFG-001 |
| `catalog-permissions.md` | The frozen permission, scope, and decision enums with every binding site (which actions require which permission) | Volume 9, chapter 05 (FR-SEC-100); tool/extension declarations (Volume 6) |
| `catalog-capabilities.md` | The provider capability enum and its resolution model (provenance, degradation strategies) | Volume 5, chapter 02 (FR-PROV-010/011; ADR-056) |

## Matrices

| File | Contents |
|---|---|
| [`compatibility-matrix.md`](compatibility-matrix.md) | Platform tiers and minimums; terminal capability tiers and the terminal matrix; provider adapters × phases × capability posture; MCP protocol revisions; Go toolchain and pinned dependency lines; configuration and schema version compatibility |
| `traceability-matrix.md` | The consolidated objective→requirement→ADR→verification chain generated at consolidation, backing the master matrix of Volume 0, chapter 09 and the FR-GH-001 automation |

## Checklists

| File | Contents |
|---|---|
| [`checklists.md`](checklists.md) | Release checklist (qualification gates, integrity artifacts, versioning, publication — Volumes 13/14); security review checklist (threat coverage, permission bindings, sandbox, secrets, audit — Volume 9); architecture review checklist (dependency matrix, ports, PAL, lifecycle, state machines — Volume 3) |

## Glossary

The corpus glossary has a single home: [Volume 0, chapter 04](../volume-00-conventions/04-glossary.md).
Volume-register "Glossary additions" sections merge there at consolidation; the annexes
carry no separate glossary.
