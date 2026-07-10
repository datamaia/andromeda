# 01 — Purpose, Audience, and Scope

## Purpose

This document set — the **Andromeda Engineering Product Specification** — is the primary source
of truth for the Andromeda project. It defines what Andromeda is, what it MUST do, how it MUST
behave, how it MUST be implemented, protected, integrated, distributed, tested, observed,
versioned, maintained, and evolved, and how development is conducted openly on GitHub with full
traceability between requirements, code, tests, and releases.

The specification MUST be precise enough that implementers — AI agents working under human
supervision, and human engineers — can build Andromeda without reinterpreting fundamental
requirements or inventing undocumented behavior.

It is not marketing material, not a commercial presentation, and not a conceptual vision
document. Every capability described here is bound to identifiers, phases, acceptance criteria,
and verification methods.

## Audience

| Audience | Primary use |
|---|---|
| AI implementation agents | Authoritative requirements, contracts, and acceptance criteria to implement against |
| Human maintainers and reviewers | Architecture, decision records, and change procedure for review and evolution |
| Contributors (open source) | Scope, conventions, governance, and contribution boundaries |
| Quality engineers | Test strategy, verification methods, quality gates |
| Security reviewers and auditors | Threat model, permission model, audit and traceability chains |
| Release engineers and operators | Distribution, update, observability, and operational specifications |

## Documentary scope

The set comprises Volumes 0–15 plus Annexes, as indexed in [docs/spec/README.md](../README.md).

**In scope:** product definition and phasing; domain model; system architecture; agent runtime;
providers, models, and authentication; tools, MCP, skills, and plugins; memory, context, and
indexing; CLI and TUI; security; configuration, storage, and observability; Git and GitHub
integration and project development process; performance and reliability; testing and quality;
distribution, installation, and updates; roadmap, execution, and open-source governance.

**Out of documentary scope:** implementation source code (except normative interface pseudocode,
schemas, and configuration examples); marketing collateral; legal advice; third-party product
documentation (referenced, never restated as authority).

## Language and self-containment

The specification is written in **English**. The project brief that seeded it is not part of the
repository; this document set is fully self-contained and does not depend on any external brief
to be interpreted.

## Product identity (fixed)

| Item | Value |
|---|---|
| Product name | `Andromeda` |
| Executable | `andromeda` |
| Main configuration file | `andromeda.toml` |
| Project-local directory | `.andromeda/` |
| Environment variable prefix | `ANDROMEDA_` |
| Main repository name | `andromeda` |
| GitHub organization / namespace | PENDING VALIDATION (see [open questions](08-register-open-questions.md)) |
| Main SDK name | `andromeda-sdk` |
| Internal protocol name | `Andromeda Runtime Protocol` |
| Official package prefix | `andromeda-` |

## Requirement precedence

When requirements conflict, the following order of precedence applies (highest first). Every
detected conflict MUST be documented in the [risk register](07-register-risks.md) or resolved by
an ADR; every exception MUST be justified by an ADR.

1. Safety of the user, the system, and credentials.
2. Legality and exclusive use of official mechanisms.
3. Data integrity and recoverability.
4. Mandatory functional requirements.
5. Privacy and local operation.
6. Compatibility and portability.
7. Open architecture and absence of vendor lock-in.
8. Observability and traceability.
9. Performance.
10. User experience.
11. Implementation convenience.

## Statement classification

The specification distinguishes explicitly between the following statement kinds. Authors MUST
NOT present assumptions as facts.

| Kind | Meaning | Recorded in |
|---|---|---|
| Provided requirement | Externally mandated behavior | In-volume requirement (`FR-*`/`NFR-*`), Source: `Provided` |
| Provided constraint | Externally mandated restriction | In-volume requirement or constraint list, Source: `Provided` |
| Design decision | Choice made by this specification | ADR (`ADR-NNN`) |
| Technical assumption | Believed true, unverified | [Assumptions register](05-register-assumptions.md) |
| Product hypothesis | Believed valuable, unvalidated | [Assumptions register](05-register-assumptions.md) |
| External information pending validation | Depends on unconfirmed external facts | Marked `PENDING VALIDATION` + [open questions](08-register-open-questions.md) |
| Open question | Unresolved, does not block progress | [Open questions register](08-register-open-questions.md) |
| Risk | Potential harm with probability and impact | `RISK-*` entries, [risk register](07-register-risks.md) |
| Discarded alternative | Considered and rejected | ADR "Alternatives considered" |
