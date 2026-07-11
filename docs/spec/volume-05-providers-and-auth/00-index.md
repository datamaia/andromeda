# Volume 5 — Providers, Models, and Authentication

**Status:** Authored (draft) · **Owner:** Provider Layer / Authentication Layer (Volume 5)

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
| 01 — Provider Contract | `01-provider-contract.md` | Keystone FR-PROV-001; ProviderPort behavioral contract; the adapter declaration set; adapter registration and lifecycle | Authored (draft) |
| 02 — Capabilities and Model Discovery | `02-capabilities-model-discovery.md` | The closed capability enum; capability matrix; negotiation and verification; degradation strategies; model discovery; provider/model change notification | Authored (draft) |
| 03 — Streaming, Tool Calling, Structured Outputs | `03-streaming-toolcalling-structured-outputs.md` | Unified stream event model; tool-calling normalization; structured output modes and validation | Authored (draft) |
| 04 — Token and Cost Accounting | `04-token-and-cost-accounting.md` | Usage reports, Cost Record emission, pricing tables, token counting and estimation hand-off | Authored (draft) |
| 05 — Resilience, Routing, and Fallback | `05-resilience-routing-fallback.md` | Timeouts, retries, rate limits, circuit breakers, health verification, routing, fallback and its guard rules | Authored (draft) |
| 06 — Error Normalization | `06-error-normalization.md` | Normalization pipeline, mapping tables, and the E-PROV error catalog | Authored (draft) |
| 07 — Authentication Layer | `07-authentication-layer.md` | Keystone FR-AUTH-001; official authentication mechanisms; prohibited mechanisms; profiles | Authored (draft) |
| 08 — Credential Lifecycle | `08-credential-lifecycle.md` | SecretStorePort/AuthPort usage per ADR-014; acquisition, refresh, rotation, revocation | Authored (draft) |
| 09 — Provider Adapters Catalog | `09-provider-adapters-catalog.md` | The 19 named adapters with phase, auth method, transport, and capability notes | Authored (draft) |
| 10 — Local and Offline Operation | `10-local-and-offline-operation.md` | Volume 1 offline guarantees mapped to provider behavior | Authored (draft) |
| 11 — State Machines | `11-state-machines.md` | Full machines: Authentication Session, Provider connection | Authored (draft) |

Register fragments `98-register-a.md` (chapters 01–06) and `98-register-b.md` (chapters
07–11) are merged into `99-volume-register.md` at consolidation.
