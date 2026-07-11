# Volume 5 — Providers, Models, and Authentication

**Status:** Complete · **Owner:** Provider Layer / Authentication Layer (Volume 5)

Volume 5 is the single home of the provider contract and the capability enum (Volume 0,
chapter 03). It specifies the Provider Layer — the ProviderPort behavioral contract, the
adapter declaration set, capabilities and model discovery, streaming, tool calling, structured
outputs, token and cost accounting, resilience, routing, fallback, and error normalization —
and the Authentication Layer: authentication flows over official mechanisms only, the
credential lifecycle, the provider adapter catalog, local and offline operation, and the
Authentication Session and Provider connection state machines. Entity semantics come from
Volume 2 (chapter 05); port signatures are frozen in Volume 3 (chapter 02); credential storage
is Volume 9's.

## Chapters

| Chapter | File | Contents | Status |
|---|---|---|---|
| 01 — Provider Contract | `01-provider-contract.md` | Keystone FR-PROV-001; ProviderPort behavioral contract; the adapter declaration set; adapter registration and lifecycle | Complete |
| 02 — Capabilities and Model Discovery | `02-capabilities-model-discovery.md` | The closed capability enum; capability matrix; negotiation and verification; degradation strategies; model discovery; provider/model change notification | Complete |
| 03 — Streaming, Tool Calling, Structured Outputs | `03-streaming-toolcalling-structured-outputs.md` | Unified stream event model; tool-calling normalization; structured output modes and validation | Complete |
| 04 — Token and Cost Accounting | `04-token-and-cost-accounting.md` | Usage reports, Cost Record emission, pricing tables, token counting and estimation hand-off | Complete |
| 05 — Resilience, Routing, and Fallback | `05-resilience-routing-fallback.md` | Timeouts, retries, rate limits, circuit breakers, health verification, routing, fallback and its guard rules | Complete |
| 06 — Error Normalization | `06-error-normalization.md` | Normalization pipeline, mapping tables, and the E-PROV error catalog | Complete |
| 07 — Authentication Layer | `07-authentication-layer.md` | Keystone FR-AUTH-001; official authentication mechanisms; prohibited mechanisms; profiles | Complete |
| 08 — Credential Lifecycle | `08-credential-lifecycle.md` | SecretStorePort/AuthPort usage per ADR-014; acquisition, refresh, rotation, revocation | Complete |
| 09 — Provider Adapters Catalog | `09-provider-adapters-catalog.md` | The 19 named adapters with phase, auth method, transport, and capability notes | Complete |
| 10 — Local and Offline Operation | `10-local-and-offline-operation.md` | Volume 1 offline guarantees mapped to provider behavior | Complete |
| 11 — State Machines | `11-state-machines.md` | Full machines: Authentication Session, Provider connection | Complete |

The volume register `99-volume-register.md` consolidates everything this volume minted
(merged from the per-agent authoring fragments at the Phase B gate).
