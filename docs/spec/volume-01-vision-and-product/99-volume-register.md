# 99 — Volume 1 Register

Machine-parseable register of everything Volume 1 minted, per Volume 0 chapters 02 and 03.
Merged into the Volume 0 registers at consolidation.

## Requirements index

| ID | Title | Phase | Verification method |
|---|---|---|---|
| PRD-001 | Agent-Native Engineering Harness | Core | Traceability audit |
| PRD-002 | Vendor and Model Independence | Core | Traceability audit |
| PRD-003 | Local-First Operation | Core | Traceability audit |
| PRD-004 | Tool-First Execution | Core | Traceability audit |
| PRD-005 | Safe Agent Autonomy | Core | Traceability audit |
| PRD-006 | Transparent and Auditable Operation | Core | Traceability audit |
| PRD-007 | Extensible Open Platform | Core | Traceability audit |
| PRD-008 | First-Class Terminal Experience | Core | Traceability audit |
| PRD-009 | Human and Automation Parity | Core | Traceability audit |
| PRD-010 | Durable Sessions and Recoverable Work | Core | Traceability audit |
| PRD-011 | Portable Unix-First Platform | Core | Traceability audit |
| PRD-012 | Specification-Driven Workflows | Core | Traceability audit |
| PRD-013 | Sustainable Open-Source Delivery | Core | Traceability audit |
| RISK-PRD-001 | Provider API drift and deprecation | — | Risk register review at phase gates |
| RISK-PRD-002 | Provider access-policy changes restricting third-party clients | — | Risk register review at phase gates |
| RISK-PRD-003 | Maintainer bus factor | — | Risk register review at phase gates |
| RISK-PRD-004 | Scope creep versus MVP viability | — | Risk register review at phase gates |
| RISK-PRD-005 | Security incident caused by agent actions | — | Risk register review at phase gates |
| RISK-PRD-006 | Local-model capability gaps | — | Risk register review at phase gates |
| RISK-PRD-007 | MCP ecosystem instability | — | Risk register review at phase gates |
| RISK-PRD-008 | Competitive pace of funded incumbents | — | Risk register review at phase gates |
| RISK-PRD-009 | Contributor onboarding friction | — | Risk register review at phase gates |
| RISK-PRD-010 | Public-contract churn breaking extensions | — | Risk register review at phase gates |

Local (non-corpus) identifier series used by this volume: `UC-NN` (use cases, chapter 03),
`JTBD-N` (jobs to be done, chapter 02), `SM-NN` (success metrics, chapter 06), `PC-NN`
(provided constraints, chapter 07). These are chapter-local labels, not corpus identifiers; no
other volume may mint them, and references to them from other volumes name the volume and
chapter.

## Glossary additions

| Term | One-line meaning |
|---|---|
| Offline guarantee list | The eleven operations (Volume 1, chapter 04, Local First) that MUST work with no Internet connection when models, tools, indexes, memory, and repositories are local. |
| Tier 1 platform | A platform on which releases are built, tested, and supported, and whose full acceptance suite gates releases (Volume 1, chapter 05). |
| Tier 2 platform | A platform on which releases are built and smoke-tested; defects are accepted but do not gate releases (Volume 1, chapter 05). |
| MVP provider seed | The three MVP provider adapters: generic OpenAI-compatible, Anthropic (cloud), and Ollama (local) (Volume 1, chapter 05). |
| Success metric (SM) | A product-level measurable commitment in Volume 1, chapter 06, formalized as one or more NFRs by the named owning volume. |
| Phase gate | The audit at a phase boundary (Volume 1, chapter 05) verifying exit criteria, bound metrics, and the risk register before the next phase begins. |
| Run record | The persisted, correlation-ID-linked record of a run — configuration snapshot, provider/model identity, prompts references, tool-invocation sequence and results — sufficient for audit and replay (formalized in Volumes 4 and 10). |
| Replay mode | An execution mode that re-traverses a recorded run's decision and tool sequence from its run record without re-invoking providers (formalized in Volumes 4 and 10). |

## Assumptions

Local list per Volume 0, chapter 05 (global `AS-NNN`/`HY-NNN` numbers are minted at
consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Product hypothesis | Openly available local models with dependable tool calling exist for consumer hardware and continue to improve | SM-04 local conformance suite results per release | Local-first value narrows to non-agentic operations; RISK-PRD-006 escalates; degradation strategies in Volume 4 carry more weight |
| Product hypothesis | Andromeda's target users are terminal-proficient and prefer terminal surfaces over GUIs for this class of work | Adoption and feedback channels reviewed at phase gates | Revisit non-goal 11 (graphical clients) through the change procedure |
| Technical assumption | An implementation in the preferred language (PC-04, Go) can meet the SM-06 – SM-09 latency and memory budgets on Tier 1 platforms | Benchmark spike before MVP implementation begins, reported against chapter 06 reference conditions | The foundation technology-selection decision record is revisited per its reversal plan |
| Technical assumption | Official provider APIs expose enough surface (streaming, tool calling, usage accounting) to satisfy the Transparent AI visibility list without unofficial mechanisms | Volume 5 capability matrix cross-checked against each provider's official documentation during adapter development | Affected visibility items degrade to "unavailable for this provider," reported honestly per Principle 2 |
| Technical assumption | GitHub's platform features used by Volume 11 (Actions, releases, security tooling) remain available to open-source projects on workable terms | Release audits; PC-02 review at phase gates | Development-platform migration is a MAJOR document change; traceability chain is platform-portable by design |

## Open questions

Entries follow Volume 0, chapter 08; global `OQ-NNN` numbers are minted at consolidation.
Items marked PENDING VALIDATION below are the register entries for every `PENDING VALIDATION`
occurrence in this volume.

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V1-OQ-1 | Signing identities and macOS notarization credentials for releases (PENDING VALIDATION: signing viability) | Chapter 05, MVP minimum item 27 | No — checksums ship regardless; signing is a pipeline configuration change | Project owner obtains signing/notarization credentials; Volume 14 activates the signing pipeline | Open |
| V1-OQ-2 | macOS Intel (x86_64) Tier 2 support (PENDING VALIDATION: build/test capacity for the target) | Chapter 05, platform scope | No — Tier 1 platforms are unaffected | Volume 3 platform matrix decides after evaluating build/test capacity; falls to `Future` if not viable | Open |
| V1-OQ-3 | Which providers offer official account/subscription-based mechanisms for third-party clients (PENDING VALIDATION per provider, PC-07) | Chapter 07, PC-07 | No — API-key authentication is the baseline for all cloud adapters | Volume 5 documents the per-provider outcome from official documentation only | Open |
| V1-OQ-4 | Selection of pinned local models for the SM-04 conformance suite | Chapter 06, SM-04 measurement method | No — the suite design is model-independent | Volume 13 pins concrete models when the suite is authored; refreshed per release | Open |

## Cross-volume references

Expectations this volume places on other volumes; consolidation verifies each.

| This volume defines | Volume expected to formalize |
|---|---|
| Phase definitions (Core, MVP, Beta, v1, v2, Future, Out of Scope) | All volumes use them; Volume 15 sequences within them |
| SM-01, SM-04 (provider integration time, local-model conformance) | Volume 5 (NFRs in the PROV area) |
| SM-02, SM-03, SM-15 (tool creation, plugin creation, MCP conformance) | Volume 6 (NFRs in TOOL/PLUG/MCP areas) |
| SM-05 – SM-11 (offline, startup, latency, streaming, memory, tool-call reliability, recovery) | Volume 12 (PERF area), offline and crash-injection suites specified in Volume 13 |
| SM-12, SM-13 (run reproducibility, traceability) | Volume 10 (OBS area), audit-record semantics with Volume 9 |
| SM-14 (test coverage) | Volume 13 (TEST area) |
| SM-16 (security) | Volume 9 (SEC area) |
| SM-17 (portability) | Volume 3 (PORT area) |
| SM-18 – SM-20 (update time, rollback time, public-contract stability) | Volume 14 (REL area) |
| MVP provider seed (generic OpenAI-compatible, Anthropic, Ollama) | Volume 5 specifies the three adapters and phases the extended provider list |
| Platform tiers and Windows-as-later-phase rule | Volume 3 (platform matrix, minimum versions), Volume 14 (per-platform distribution) |
| MVP minimum and phase gates | Volume 15 (sequencing, milestones), Volume 13 (gating suites) |
| Offline guarantee list | Volume 13 (offline test suite), Volume 12 (offline-operation NFR) |
| Product principles (normative statements) | Volumes 4–14 specialize without weakening; Volume 9 owns permissions; Volume 5 owns capability vocabulary; Volume 6 owns the tool contract; Volume 10 owns event and logging envelopes |
| Error visibility fields (code, category, message, safe context, cause, recoverability, recommended action, correlation ID) | Volume 0 owns the scheme; area owners mint concrete `E-*` codes |
| Brand identity referenced by the tagline (palette, typography, mascot, banner) | Volume 8 (design tokens), decision register (Volume 0, chapter 06) |
