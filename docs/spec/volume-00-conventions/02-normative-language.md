# 02 — Normative Language and Templates

## Normative keywords

The following keywords, always written in uppercase, carry normative force throughout the
specification. They follow the spirit of RFC 2119 / RFC 8174 with three project-specific
additions (PENDING VALIDATION, OUT OF SCOPE, DEPRECATED).

| Keyword | Meaning |
|---|---|
| **MUST** | Mandatory requirement. Non-compliance is a defect. |
| **MUST NOT** | Prohibited behavior. Occurrence is a defect. |
| **SHOULD** | Recommended behavior; deviation requires a documented justification (ADR or register entry). |
| **SHOULD NOT** | Discouraged behavior; occurrence requires a documented justification. |
| **MAY** | Optional behavior. |
| **PENDING VALIDATION** | A decision or integration that depends on unconfirmed external information. Every occurrence MUST have a matching entry in the [open questions register](08-register-open-questions.md) or the volume's own register. |
| **OUT OF SCOPE** | Functionality that is not part of the defined phase or of the product at all. |
| **DEPRECATED** | Behavior kept temporarily for compatibility and scheduled for removal. |

Rules:

- MUST/MUST NOT are never used for recommendations; SHOULD/MAY are never used for mandatory
  behavior.
- Lowercase "must", "should", "may" in prose carry no normative force; authors SHOULD avoid them
  where they could be misread as normative.

## Banned vague terms

The following adjectives MUST NOT appear in normative statements unless accompanied — in the same
statement or an explicitly referenced requirement — by a metric, a verifiable criterion, or a
precise definition:

> fast, easy, intuitive, robust, scalable, secure, efficient, modern, advanced, compatible,
> powerful.

The linter flags occurrences at warning level (see [Linting conventions](#linting-conventions)).
Product-principle names such as "Safe by Default" and proper terms such as "OpenAI-compatible
API" are acceptable because they are defined precisely in the glossary and Volume 5 respectively.

## Requirement statement style

- Each requirement is a numbered, self-contained unit using the templates below, with a stable
  identifier per [chapter 03](03-id-taxonomy-and-ownership.md).
- Acceptance criteria use Given/When/Then where it improves precision and MUST include negative
  cases, error cases, permission cases, and observability effects where applicable.
- Requirements are never deleted or renumbered; superseded requirements are marked
  `Status: DEPRECATED` with a pointer to their replacement.

## Functional requirement template

Every functional requirement MUST use exactly this structure. The metadata bullets and `####`
section headings below are enforced verbatim by the linter; additional sections MAY be appended
after `#### Traceability`.

```markdown
### FR-AREA-NNN — Name

- Type:
- Status:
- Priority:
- Phase:
- Source:
- Owner:
- Affected components:
- Dependencies:
- Related risks:

#### Description

#### Motivation

#### Actors

#### Preconditions

#### Main flow

#### Alternative flows

#### Edge cases

#### Inputs

#### Outputs

#### States

#### Errors

#### Constraints

#### Security

#### Observability

#### Performance

#### Compatibility

#### Acceptance criteria

#### Verification method

#### Traceability
```

Field values:

- **Type:** `Functional`.
- **Status:** `Draft` | `Approved` | `DEPRECATED`.
- **Priority:** `P0` (blocking) | `P1` (high) | `P2` (normal) | `P3` (low).
- **Phase:** `Core` | `MVP` | `Beta` | `v1` | `v2` | `Future` | `Out of Scope`.
- **Source:** `Provided` (external mandate) | `Derived` (follows from another requirement) |
  `Design` (introduced by this specification).
- **Owner:** owning component or role.

## Non-functional requirement template

```markdown
### NFR-AREA-NNN — Name

- Category:
- Priority:
- Phase:
- Metric:
- Target:
- Minimum threshold:
- Measurement method:
- Test environment:
- Measurement frequency:
- Owner:
- Dependencies:
- Risks:
- Acceptance criteria:
```

**Category** is one of: `Performance`, `Reliability`, `Security`, `Privacy`, `Usability`,
`Accessibility`, `Portability`, `Compatibility`, `Observability`, `Maintainability`,
`Scalability`, `Compliance`.

## Architecture decision record template

Full ADR bodies live in [`annexes/adr/`](../annexes/00-index.md), one file per ADR, indexed in
the [decision register](06-register-decisions.md).

```markdown
### ADR-NNN — Title

- Status:
- Date:
- Deciders:
- Components:
- Related requirements:

#### Context

#### Problem

#### Forces and constraints

#### Alternatives considered

#### Decision

#### Rationale

#### Positive consequences

#### Negative consequences

#### Risks

#### Reversal plan

#### Review conditions
```

**Status** is one of: `Proposed` | `Accepted` | `Superseded by ADR-NNN` | `DEPRECATED`.

## Risk entry template

```markdown
### RISK-AREA-NNN — Name

- Category:
- Probability:
- Impact:
- Severity:
- Mitigation:
- Detection:
- Owner:
- Status:
```

**Probability** and **Impact** are `Low` | `Medium` | `High`; **Severity** is the resulting
`Low` | `Medium` | `High` | `Critical` classification.

## Formatting conventions

- All documentation is Markdown with hierarchical headings; one `#` title per file.
- Tables for comparisons and enumerable facts; numbered lists for sequences; bullet lists for
  enumerations.
- Fenced code blocks MUST declare a language tag (`text` for plain content). Interfaces use
  typed pseudocode fenced as `pseudo`; configuration uses `toml`; GitHub Actions and manifests
  use `yaml`; schemas use `json`; CLI/TUI wireframes use `text`.
- Mermaid diagrams (` ```mermaid `) are used for architecture, state machines, sequences, and
  flows. A diagram never substitutes for prose: every diagram MUST be accompanied by a textual
  description of its components, relations, and constraints.
- Every stateful entity's machine defines: initial state, terminal states, transitions, events,
  guards, side effects, persistence, recovery, timeouts, cancellation, retries, and errors.

## Linting conventions

`scripts/spec_lint.py` enforces this volume mechanically. Conventions the linter understands:

- **Intentionally invalid examples.** The specification MUST include invalid configuration and
  schema examples. Fence them as ` ```toml invalid ` or ` ```json invalid ` so the linter skips
  parsing; unmarked `toml`/`json` blocks MUST parse.
- **Vague-term suppression.** A line that legitimately uses a banned adjective (e.g., quoting a
  product principle) MAY carry the trailing comment `<!-- lint:allow vague -->`.
- **Definition headings.** A heading of the form `### <ID> — <Name>` *defines* that identifier;
  any other occurrence is a *reference*. Definitions are unique corpus-wide and permitted only
  in the identifier's owning volume (chapter 03).
- **Volume registers.** Every content volume (1–15) ends with `99-volume-register.md`, whose
  requirement index MUST match the definitions present in that volume, in both directions.
