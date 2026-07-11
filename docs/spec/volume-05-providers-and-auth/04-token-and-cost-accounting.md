# 04 — Token and Cost Accounting

This chapter defines how the Provider Layer produces the accounting ground truth behind
"what did this cost?" (Principle 7): the normalized usage report, Cost Record emission, the
pricing-table mechanism for estimates, and the token-counting/estimation hand-off. The Cost
Record entity and its invariants are Volume 2's (chapter 08, INV-COST-01..05); storage,
rollups, and retention are Volume 10's; presentation is Volume 8's. This volume owns *what
the Provider Layer reports and when*.

## The usage report

Every `ChatResponse`, terminal stream event, and `EmbedResponse` carries a **usage report**
populated exclusively from official provider accounting (INV-COST-01):

| Field | Presence | Meaning |
|---|---|---|
| `input_tokens` | when officially reported | Prompt tokens |
| `output_tokens` | when officially reported | Completion tokens |
| `cached_tokens` | when officially reported | Cache-served input tokens |
| `reasoning_tokens` | when officially reported | Reasoning tokens (Principle 7 — official counts only) |
| `reported_cost` | when officially reported | Monetary cost in integer micro-units + ISO 4217 currency, only where the provider declares `cost_reporting` |
| `fields_declared` | always | Which of the above the adapter's `UsageReporting` declaration promises, so absence is distinguishable from omission |

Adapters MUST NOT compute, extrapolate, or back-fill token counts: a provider that reports
nothing yields an empty report and downstream `cost_basis = unavailable`. Which providers
report which fields is declared per adapter (FR-PROV-002) and recorded in the catalog
(chapter 09) — per-provider reporting facts are PENDING VALIDATION against official
documentation where undocumented (register entry in `98-register-a.md`).

## Cost Record emission

The Provider Layer emits accounting data to the Runtime, which appends **Cost Records**
(Volume 2) — one per provider request that produced usage, including interrupted streams
with partial usage (delivered-usage snapshot, chapter 03). `cost_basis` resolves:

1. **`actual`** — the provider officially reported cost (`cost_reporting`); `cost_micros`
   copies the reported value.
2. **`estimated`** — no reported cost, but a pricing table covers the (provider, model)
   pair; `cost_micros` is computed from token counts × table prices. Estimates are never
   presented as actuals (INV-COST-03).
3. **`unavailable`** — neither reported cost nor applicable pricing data; token counts (if
   any) are recorded without a monetary value (INV-COST-01).

## Pricing tables

Andromeda ships **no built-in price data** (ADR-058): prices change at providers' discretion
and shipping them would assert external commercial facts the project cannot keep true.
Prices come from user-maintained configuration:

```toml
[providers.anthropic.pricing."example-cloud-model"]
input_per_million_micros = 3000000    # micro-units of `currency` per 1,000,000 input tokens
output_per_million_micros = 15000000
cached_input_per_million_micros = 300000
currency = "USD"
source = "provider pricing page, retrieved by the user"
effective_date = "2026-07-01"
```

Keys minted here (`[providers.<slug>.pricing."<model>"]`): `input_per_million_micros`,
`output_per_million_micros`, `cached_input_per_million_micros`,
`reasoning_per_million_micros`, `currency`, `source`, `effective_date` — all values integer
micro-units (INV-COST-02); `source` and `effective_date` are required (INV-MDL-04). Loaded
tables update the Model row's `pricing` attribute; the values in the example are
placeholders, not shipped facts.

### FR-PROV-030 — Token usage accounting

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: adapters, Provider Router, Runtime (Cost Record append), Context Manager
- Dependencies: FR-PROV-001, FR-PROV-002; Volume 2 INV-COST-01..05; ADR-058
- Related risks: RISK-PROV-003

#### Description

Every provider request MUST yield a usage report per the table above, populated only from
official provider accounting, with declared-versus-reported field bookkeeping. The layer
MUST hand the report to the Runtime for Cost Record emission on every completion,
including failed requests that consumed tokens and interrupted streams with determinable
partial usage. `CountTokens` MUST be served only via a documented counting mechanism
(`token_counting` capability); absent one, it returns E-PROV-006 so the Context Manager
falls back to its estimation strategy (Volume 7) — estimation never masquerades as
provider accounting.

#### Motivation

Honest token accounting is the substrate of cost transparency (PRD-006), context budgeting
(Volume 7), and budget enforcement (Volume 4); fabricated counts would poison all three.

#### Actors

Adapters (extract usage); router (attach attribution); Runtime (append records); Context
Manager (counting hand-off).

#### Preconditions

Request completed, failed-with-usage, or interrupted-with-determinable-usage.

#### Main flow

1. The adapter extracts officially reported usage fields.
2. The router attaches provider slug, model name, and correlation ULIDs.
3. The Runtime appends the Cost Record (`cost_basis` per the resolution rules) and
   `provider.cost.recorded` is emitted.

#### Alternative flows

- Streaming: cumulative `usage_update` events refine the in-flight view; the terminal
  event's report is authoritative for the record.
- `CountTokens` with `token_counting`: the adapter calls the documented mechanism and
  returns the count with the model identity echoed.

#### Edge cases

- Provider reports totals only (no input/output split): the report carries what exists;
  absent fields stay absent — no arithmetic reconstruction.
- Duplicate usage on retries: each attempt that reached the provider and returned usage
  yields its own record (attempt-numbered), so retry spend is visible (chapter 05).
- Cancelled before dispatch: no record (nothing consumed).

#### Inputs

Wire responses; `UsageReporting` declaration; correlation metadata.

#### Outputs

Usage reports; Cost Records (via Runtime); `provider.cost.recorded` events; `TokenCount`
results.

#### States

None; records are append-only (INV-COST-04).

#### Errors

E-PROV-006 (no counting mechanism); extraction failures are E-PROV-008 on the request.

#### Constraints

No fabricated counts (INV-COST-01); integer micro-unit money only (INV-COST-02); records
immutable (INV-COST-04).

#### Security

Usage reports carry numbers and identities, never content; account-level identifiers in
provider usage payloads are dropped at extraction (Volume 9 redaction).

#### Observability

`provider.cost.recorded` (payload: provider slug, model, token fields, basis, attempt);
token/cost metrics per Volume 12 taxonomy; per-run aggregation views are derived (INV-COST-05).

#### Performance

Extraction is on the response path with negligible budget (Volume 12 envelope); record
append uses the SessionStorePort hot path budget.

#### Compatibility

Usage report shape is public contract surface (SM-20); new officially-reported fields are
additive.

#### Acceptance criteria

- Given a provider reporting full usage, when a request completes, then the Cost Record
  matches the wire-reported values exactly and `cost_basis` reflects the resolution rules.
- Given a provider reporting nothing, when a request completes, then the record carries no
  token values and `cost_basis = unavailable` — and no estimation appears anywhere labeled
  as usage.
- Given an interrupted stream with partial usage determinable, when it terminates, then the
  partial snapshot is recorded and flagged as interrupted.
- Negative case: an adapter computing counts locally and reporting them as official fails
  the conformance suite's accounting-honesty check.
- Observability case: for every suite run, Cost Records and `provider.cost.recorded` events
  are 1:1 with usage-bearing requests (SM-13 chain).

#### Verification method

Conformance suite accounting-honesty checks against recorded fixtures; fault-injection for
partial usage; SM-13 audit-chain test; Volume 13 contract tests for `CountTokens`.

#### Traceability

PRD-006; SM-12, SM-13; ADR-058; INV-COST-01..05; FR-PROV-031.

### FR-PROV-031 — Cost accounting and pricing tables

- Type: Functional
- Status: Draft
- Priority: P1
- Phase: MVP
- Source: Provided
- Owner: Provider Layer (Volume 5)
- Affected components: Provider Layer, Configuration Manager, Runtime, CLI/TUI cost views
- Dependencies: FR-PROV-030; ADR-058; Volume 2 INV-MDL-04
- Related risks: RISK-PROV-003

#### Description

Monetary cost MUST resolve per the `cost_basis` rules: `actual` only from official
`cost_reporting`, `estimated` only from user-supplied pricing tables carrying `source` and
`effective_date`, `unavailable` otherwise. Estimates MUST be computed as
`tokens × per-million price ÷ 1,000,000` per field, summed in integer micro-units, and MUST
be labeled as estimates everywhere they surface. Pricing tables validate at configuration
load (required fields, positive integers, ISO 4217 currency); invalid tables are rejected,
not partially applied.

#### Motivation

Users decide budgets on cost; a wrong number labeled honestly is recoverable, a fabricated
or mislabeled one is not (INV-COST-03).

#### Actors

Users (maintain tables); Configuration Manager (validate); Runtime (compute); cost views.

#### Preconditions

Usage report available; pricing table optionally configured.

#### Main flow

1. Cost resolution walks actual → estimated → unavailable.
2. Estimated costs compute per-field and sum; currency copies from the table.
3. The record persists with its basis; views label estimates (Volume 8).

#### Alternative flows

- Mixed availability (reported cost for some requests, table for others in one run): each
  record keeps its own basis; aggregates that mix bases MUST surface the mix (a run total
  is labeled "includes estimates" when any member is estimated).

#### Edge cases

- Table lacks a field the usage report has (e.g., reasoning tokens without
  `reasoning_per_million_micros`): the uncovered portion is excluded and the record is
  flagged partial-estimate in safe metadata.
- Currency differences across providers in one run: aggregates group by currency; no
  implicit conversion exists.
- `effective_date` in the future: the table entry is inert until that date and the config
  validation warns.

#### Inputs

Usage reports; pricing tables; officially reported costs.

#### Outputs

Cost Records with basis; labeled cost views' source data.

#### States

None.

#### Errors

Configuration validation failures per Volume 10 (E-CFG family) for malformed tables; this
volume defines the semantic rules above.

#### Constraints

No shipped price data (ADR-058); no floating-point money (INV-COST-02); no silent currency
conversion.

#### Security

Pricing tables are non-secret configuration; `source` strings are user-provided and
redaction-exempt but length-bounded per Volume 10 validation.

#### Observability

Basis distribution metrics (share of actual/estimated/unavailable per provider) per Volume
12 taxonomy; `provider.cost.recorded` carries the basis.

#### Performance

Cost computation is integer arithmetic on the response path; negligible within Volume 12
budgets.

#### Compatibility

Pricing key set is public configuration surface (Volume 10 schema versioning applies).

#### Acceptance criteria

- Given a pricing table covering a model, when a usage-bearing request completes without
  reported cost, then the record is `estimated` with the arithmetic above, and the CLI/TUI
  cost views label it as an estimate.
- Given no table and no reported cost, when the request completes, then the record is
  `unavailable` and views show token counts without a monetary figure.
- Negative case: a table entry missing `source` or `effective_date` is rejected at load
  with the offending key named; no entry from that table applies.
- Compatibility case: aggregates never sum across currencies; a two-currency run shows two
  totals.
- Observability case: basis is queryable per record and per aggregate (SM-13 chain intact).

#### Verification method

Unit tests over resolution and arithmetic (including rounding at micro-unit granularity);
configuration validation tests; CLI/TUI labeling tests (Volume 8/13).

#### Traceability

PRD-006; ADR-058; INV-COST-01..05, INV-MDL-04; FR-PROV-030.

## Estimation hand-off and budgets

Where accounting inputs are missing, responsibilities split cleanly:

- **Prompt-size estimation** (before a request): Context Manager (Volume 7), triggered by
  E-PROV-006 from `CountTokens`. Its estimates budget context assembly and are never
  written into Cost Records.
- **Budget enforcement** (during a run): run/session budget semantics are Volume 4's. The
  Provider Layer supplies the enforcement inputs — live per-run accumulated tokens and cost
  by basis — and refuses fallback candidates that violate cost guard rules (chapter 05).
- **Aggregation and retention**: rollups, pruning, and export are Volume 10's (INV-COST-05
  divergence rule applies).

## Accounting observability

| Event | Version | Producer | Payload (summary) | Correlation |
|---|---|---|---|---|
| `provider.cost.recorded` | 1 | Provider Layer / Runtime | provider slug, model name, token fields, `cost_micros`+currency when present, `cost_basis`, attempt number | run, turn, trace ULIDs |

Envelope per Volume 10; payloads carry numbers and identities only.

### NFR-PROV-003 — Accounting completeness

- Category: Observability
- Priority: P1
- Phase: MVP
- Metric: Fraction of provider requests in instrumented suite runs that (a) produce a Cost Record with a declared `cost_basis`, and (b) contain zero token or cost values not traceable to official provider accounting or a labeled pricing-table estimate
- Target: 100% for both (a) and (b)
- Minimum threshold: 100% — accounting honesty is an identity property (INV-COST-01), not a tunable
- Measurement method: Record-completeness validator over all integration and E2E suite runs: enumerate provider requests from spans, resolve each to its Cost Record, verify basis provenance against recorded wire fixtures
- Test environment: CI suites on Tier 1 platforms with recorded fixtures and live seed providers
- Measurement frequency: Every release; regressions block per SM-12/SM-13 governance
- Owner: Provider Layer (Volume 5)
- Dependencies: FR-PROV-030, FR-PROV-031
- Risks: RISK-PROV-003
- Acceptance criteria: The per-release validator report shows zero requests without records, zero records with undeclared basis, and zero values failing provenance resolution.

### RISK-PROV-003 — Stale or wrong pricing data misleads users

- Category: Product / data quality
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: No shipped price data (ADR-058); mandatory `source` and `effective_date` on every table entry; estimates always labeled (INV-COST-03); mixed-basis aggregates flagged; `actual` reported cost always wins over tables
- Detection: Basis-distribution metrics (a provider drifting from `actual` to `estimated` signals reporting changes); user-visible `effective_date` in cost views; divergence between reported and estimated cost where both exist is surfaced in cost views
- Owner: Provider Layer (Volume 5)
- Status: Open

The residual risk is a user maintaining an outdated table and trusting estimates; labeling,
dating, and divergence surfacing bound the harm to a visible, correctable configuration
issue rather than silent misinformation.
