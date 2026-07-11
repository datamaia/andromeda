# 10 — Completeness and Change Procedure

## Definition of completeness

The specification is complete when all of the following hold. These conditions are re-checked at
every release of the document set; the mechanical subset is enforced by `scripts/spec_lint.py`.

1. Every volume and chapter listed in the master index exists and is non-empty.
2. Every mandated domain entity (Volume 2) and architecture component (Volume 3) is defined with
   its full template.
3. Every requirement has a stable identifier, a phase, acceptance criteria, and a verification
   method.
4. Every component defines its errors, states, and observability.
5. Every dangerous action is bound to the permission model; every external integration uses
   official, documented mechanisms only.
6. Every workflow has states and transitions; every CLI command has exit codes; every
   configuration key has a place in the precedence chain.
7. Every metric has a measurement method; every risk has a mitigation; every `PENDING
   VALIDATION` has an open-questions entry.
8. The identifier corpus is unique, ownership-consistent, and cross-reference-resolvable.
9. No undocumented contradictions exist; conflicts are resolved by the precedence order in
   chapter 01 and recorded as ADRs or risks.
10. The MVP defined in Volume 1/15 is functional, viable, and does not require implementing the
    entire product.

## Completeness statement

Version 1.0.0 of this document set satisfies the definition of completeness above,
verified on 2026-07-11 by: `scripts/spec_lint.py` (0 errors, 0 warnings across the corpus —
identifier uniqueness and ownership, template completeness, phases, acceptance criteria,
verification methods, cross-reference resolution, register round-trips, embedded-example
validity) and a six-dimension consistency audit (coverage/traceability, terminology,
cross-volume contradictions, security/legality, verifiability/ambiguity, MVP viability),
whose confirmed findings were fixed and re-linted. Open questions and PENDING VALIDATION
items are consolidated in [chapter 08](08-register-open-questions.md); none blocks
implementation. The corpus holds 13 objectives, 482 requirements (266 FR, 102 NFR,
114 RISK), 121 ADRs, 222 error codes, and 322 events across Volumes 0–15 and the annexes.

## Document versioning

The document set is versioned as a whole using semantic versioning (`MAJOR.MINOR.PATCH`),
independent of product releases:

- **MAJOR** — a change that removes or redefines approved requirements or reverses an accepted
  ADR.
- **MINOR** — new requirements, new chapters, phase changes, new ADRs.
- **PATCH** — editorial fixes with no normative effect.

The current version is stated in `docs/spec/README.md`. Changes land through pull requests per
the process in Volume 11.

## Change procedure

1. **Propose.** Open an issue describing the motivation, the affected identifiers, and the kind
   of change (requirement, decision, editorial).
2. **Classify.** Decisions and reversals require an ADR (next number in the appropriate block).
   Requirement changes keep IDs stable: supersede with a new requirement and mark the old one
   `Status: DEPRECATED` pointing to its replacement.
3. **Author.** Apply the templates of chapter 02; respect area ownership of chapter 03; update
   the owning volume's `99-volume-register.md` and the relevant registers (05–08).
4. **Validate.** `scripts/spec_lint.py` MUST pass with no errors; warnings require review.
5. **Review.** Pull request per Volume 11 rules (human review is mandatory; AI-generated changes
   are labeled).
6. **Record.** Update the decision register (06), the master traceability matrix (09) when
   objectives/requirements change, and the document version.

## Audit obligations

Before each document-set release, an audit MUST verify: coverage, consistency, traceability,
ambiguity, unverifiable requirements, unjustified decisions, interfaces without defined errors,
states without transitions, commands without exit codes, configurations without precedence,
features without a delivery phase, risks without mitigation, circular dependencies,
contradictory requirements, unimplementable capabilities, unofficial dependencies, and
capabilities incompatible with local operation or the permission model.
