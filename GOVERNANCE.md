# Governance

This is a summary. The full governance model is specified in
[Volume 15, chapter 04](docs/spec/volume-15-roadmap-and-execution/04-open-source-governance.md).

## Roles

- **Users** — anyone using Andromeda. Feedback and issues welcome.
- **Contributors** — anyone whose change has been merged.
- **Committers** — trusted contributors with triage and review responsibilities.
- **Maintainers** — accountable for a subsystem's direction, review, and release quality;
  listed in [MAINTAINERS.md](MAINTAINERS.md).

## Decisions

- Routine changes: lazy consensus through pull requests with mandatory human review.
- Architectural or contract-affecting changes: an **ADR** (see
  [Volume 0, chapter 10](docs/spec/volume-00-conventions/10-completeness-and-change-procedure.md))
  and, where broad, an **RFC** in Discussions.
- Disagreements escalate to the maintainers; unresolved ties are decided by the release
  authority. Conflict resolution and RFC process details live in Volume 15.

## Release authority

Releases are cut only by maintainers holding release authority. A **minimum of two**
release-authority holders is a v1 exit condition (bus-factor mitigation, RISK-PRD-003).

## Licensing and trademark

Code is licensed under [Apache-2.0](LICENSE) (ADR-002). Use of the "Andromeda" name follows
the trademark policy in Volume 15.

## Amending governance

Changes to this model follow the specification change procedure (Volume 0, chapter 10) and
require maintainer approval.
