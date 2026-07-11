# Volume 13 — Testing and Quality

**Status:** Complete · **Owner:** Testing and quality (Volume 13)

Volume 13 defines how Andromeda is verified: the test strategy and pyramid, the closed
catalog of test types with tooling, gates, and phases, the test infrastructure (doubles,
determinism, coverage, mutation, test data, secret handling, CI parallelization), and the
quality gates and release qualification pipeline that stand between a change and a published
release. It formalizes SM-14 (test coverage) as NFRs in the `TEST` area and defines the suite
mechanics behind the metrics other volumes formalize (offline, portability, recovery,
contract stability). The testing toolchain is fixed by ADR-017; this volume specifies how it
is applied.

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Test Strategy and Pyramid](01-strategy-and-pyramid.md) | Strategy principles, execution tiers T0–T4, the pyramid with distribution constraints, test-to-requirement traceability annotations, merge-gate and determinism NFRs |
| [02 — Test Types Catalog](02-test-types-catalog.md) | The closed catalog of 32 test types across foundation, deterministic-output, generative, contract/conformance, surface, guarantee, lifecycle, operational, and process families — each with scope, tooling, gate tier, and phase; port contract kits; the offline suite |
| [03 — Fixtures, Fakes, Determinism, Coverage, and Mutation](03-fixtures-fakes-determinism.md) | Test-double taxonomy, provider test lanes, determinism controls, flaky-test quarantine, coverage and mutation gates, test data, secret handling, CI parallelization |
| [04 — Release Qualification and Quality Gates](04-release-qualification-and-gates.md) | Gate ladder, the S1–S6 qualification pipeline, the evidence bundle, waiver policy, qualification errors and events |
| [99 — Volume 13 Register](99-volume-register.md) | Everything this volume minted: requirements, ADRs, errors, events, glossary, assumptions, open questions |

## Decisions minted by this volume

| ADR | Decision |
|---|---|
| [ADR-175](../annexes/adr/ADR-175.md) | Mutation testing: scoped packages, phase-gated cadence, tooling pending validation |
| [ADR-176](../annexes/adr/ADR-176.md) | Provider test doubles: scripted fakes first, sanitized recordings second, live lanes non-gating |
| [ADR-177](../annexes/adr/ADR-177.md) | Flaky-test quarantine: build-tag exile with a time box, no blind retries |
