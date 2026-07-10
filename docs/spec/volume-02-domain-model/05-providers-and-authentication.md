# 05 — Providers and Authentication

This chapter defines the Provider aggregate (**Provider**, **Model**), the **Capability**
value vocabulary, and the Credential aggregate (**Credential**, **Authentication Session**).
The provider contract, adapter rules, capability negotiation, and authentication flows are
owned by Volume 5; credential *storage* is owned by Volume 9 (Secret Store). Everything here
is vendor-agnostic by construction: no attribute of any entity in this chapter may be
meaningful for only one vendor (Principle 1).

## Provider

Purpose: an adapter-backed source of model inference — a cloud service or a local server —
registered in configuration and reached through exactly one adapter.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `slug` | `string` | yes | Stable configuration key (e.g., `anthropic`, `local-ollama`); unique |
| `display_name` | `string` | yes | Human-facing name |
| `adapter` | `string` | yes | Adapter identifier implementing the provider contract (e.g., `openai_compatible`, `anthropic`, `ollama`, or an Extension-provided adapter); registry owned by Volume 5 |
| `endpoint` | `json` | no | Non-secret connection parameters (base URL, organization/project identifiers); shape owned by the adapter's declared config schema (Volume 5) |
| `declared_capabilities` | `json` | yes | Provider-level Capability values declared by the adapter |
| `auth_kind` | `enum` | yes | `none` \| `api_key` \| `oauth` \| `custom`; which authentication family the provider requires (flows in Volume 5) |
| `connection_state` | `enum` | yes | Canonical Provider connection state (chapter 09) |
| `last_verified_at` | `timestamp` | no | Last successful capability/health verification |
| `enabled` | `boolean` | yes | Provider participates in routing |
| `adapter_metadata` | `json` | no | Opaque adapter-owned data; never interpreted by the Core Domain (chapter 01, rule 5) |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `slug` unique (it is the `[providers.<slug>]` configuration
  table key, Volume 10).

### Relations

- Exposes 0..n **Model** (aggregate members).
- Targeted by 0..n **Authentication Session**; 0..n **Credential** rows may be issued for it.
- Referenced by **Agent Profile** `model_selector` and by **Turn**/**Cost Record** snapshots
  (by `slug`, not FK — records outlive registrations).

### Integrity invariants

1. **INV-PRV-01** — `slug` MUST be unique and MUST NOT be reused for a different service after
   deletion while any record referencing it is retained (slug reuse would corrupt historical
   attribution).
2. **INV-PRV-02** — A Provider MUST reference a registered adapter; all provider-specific
   behavior lives in that adapter (Principle 1). The Provider row itself carries no
   vendor-specific fields outside `adapter_metadata`.
3. **INV-PRV-03** — `declared_capabilities` MUST contain only values of the Volume 5
   Capability enum.
4. **INV-PRV-04** — `endpoint` MUST NOT contain secret material (keys, tokens, passwords);
   secrets travel only through the Credential aggregate.

### Lifecycle

Stateful — canonical connection states `configured`, `verifying`, `available`, `degraded`,
`unavailable`, `disabled`, `removed` (chapter 09); full machine owned by Volume 5.

### Persistence

Global database, table `providers` (provider registrations are machine-level; workspace
configuration may narrow or override selection per Volume 10 precedence). Retention: until
removed; rows referenced by history are tombstoned (`removed` state) rather than deleted.

### Versioning and serialization

Row versioning via `revision`. Serializes as canonical JSON without `adapter_metadata` unless
the export explicitly requests adapter data.

## Model

Purpose: a concrete inference target exposed by a provider, with declared capabilities. The
Runtime keys behavior off these declarations, never off model or vendor names (Principle 2).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `provider_id` | `ulid` | yes | Owning Provider |
| `name` | `string` | yes | Provider-namespaced model identifier as used on the wire (external identifier, stored verbatim) |
| `display_name` | `string` | no | Human-facing name |
| `capabilities` | `json` | yes | Capability values declared/detected for this model (Volume 5 negotiation) |
| `context_window_tokens` | `integer` | no | Declared context window, when known |
| `max_output_tokens` | `integer` | no | Declared output limit, when known |
| `pricing` | `json` | no | Per-token prices in integer micro-units + currency, when known; source recorded (used for Cost Record estimates, chapter 08) |
| `discovered` | `boolean` | yes | Row was produced by adapter discovery vs. explicit configuration |
| `first_seen_at` | `timestamp` | yes | First discovery/registration |
| `last_seen_at` | `timestamp` | yes | Most recent confirmation the model is still offered |
| `deprecated` | `boolean` | yes | No longer offered or scheduled for removal by the provider |
| `adapter_metadata` | `json` | no | Opaque adapter-owned data |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `(provider_id, name)` unique.
- Records elsewhere snapshot `provider_slug` + `model_name` strings; the Model row is the
  live registry, not the historical truth.

### Relations

- Belongs to exactly one **Provider** (aggregate member).
- Selected by **Agent Profile** `model_selector`; snapshotted by **Turn**, **Agent**, and
  **Cost Record**; bound to semantic **Index** embedding spaces (chapter 07).

### Integrity invariants

1. **INV-MDL-01** — `(provider_id, name)` MUST be unique; the same wire name under two
   providers is two Model rows.
2. **INV-MDL-02** — `capabilities` MUST contain only Volume 5 Capability enum values; absence
   of a capability means *not available* — the Runtime MUST NOT infer capabilities from
   `name` (Principle 2).
3. **INV-MDL-03** — Capability changes detected by verification MUST update the row and emit
   an Event; they never silently rewrite history (Turns keep what was used).
4. **INV-MDL-04** — `pricing`, when present, records its source and effective date; Cost
   Records computed from it are marked as estimates (chapter 08).

### Lifecycle

Stateless registry entry (`deprecated`/`discovered` are recorded flags; availability follows
the Provider connection state).

### Persistence

Global database, table `models`. Retention: rows persist while their Provider exists;
deprecated rows are kept for attribution.

### Versioning and serialization

Row versioning via `revision`. Serializes as canonical JSON; capability sets as sorted string
arrays.

## Capability

Purpose: a declared, machine-checkable ability of a model or provider (e.g., tool calling,
streaming, vision). Capability is a **value vocabulary**, not a stored aggregate: the closed
enum is minted and owned by Volume 5, and additions require an ADR (Volume 0, chapter 03).

### Attributes

Capability values appear in the model as strings inside capability sets. The vocabulary
definition (in Volume 5) gives each value:

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `name` | `enum` | yes | Canonical `snake_case` capability name (e.g., `tool_calling`, `streaming`, `vision`) |
| `level` | `enum` | yes | Declaration level: `provider` \| `model` \| `both` — where the value may legally appear |
| `meaning` | `text` | yes | Precise, testable definition of what declaring it promises |

### Identifiers

The `name` is the identity. There is no ULID and no table: the enum ships in code, versioned
with the provider contract (Volume 5).

### Relations

- Referenced by **Provider** `declared_capabilities`, **Model** `capabilities`, **Agent
  Profile** selectors, and workflow/tool requirement declarations (Volumes 4/6).

### Integrity invariants

1. **INV-CAP-01** — The Capability enum is closed: values are added only by ADR (Volume 0,
   chapter 03) and are never renamed or reused.
2. **INV-CAP-02** — Persisted capability sets MUST contain only enum values valid for their
   level (INV-PRV-03, INV-MDL-02); unknown values found at read time (e.g., after a
   downgrade) are surfaced as validation errors, not ignored.
3. **INV-CAP-03** — A capability that is not declared is absent: no component may simulate it
   silently (Principle 2).

### Lifecycle

Stateless value vocabulary.

### Persistence

Not persisted as rows; embedded as string values in owning entities.

### Versioning and serialization

The enum version travels with the provider contract version (Volume 5). Serialized as plain
strings.

## Credential

Purpose: a secret used to authenticate against a provider or service. The Credential entity is
**metadata plus a reference**: the secret material itself lives exclusively in the Secret
Store (Volume 9).

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `label` | `string` | yes | Human-facing name (e.g., `anthropic-personal`); unique |
| `provider_id` | `ulid` | no | Provider this credential is issued for; absent for service-generic credentials (e.g., Git hosting, Volume 11) |
| `service` | `string` | conditional | Service identifier when not provider-bound; required if `provider_id` is absent |
| `kind` | `enum` | yes | `api_key` \| `oauth_refresh_token` \| `basic` \| `custom` (families; flow details in Volume 5) |
| `secret_ref` | `string` | yes | Opaque handle into the Secret Store (Volume 9); resolves to the secret material only inside the Authentication Layer |
| `fingerprint` | `string` | no | Non-reversible hint for display (e.g., last four characters); derivation rules owned by Volume 9 |
| `status` | `enum` | yes | Recorded status: `active` \| `rotated` \| `revoked` \| `expired` (chapter 09 recorded vocabulary) |
| `rotated_to_id` | `ulid` | no | Successor Credential after rotation |
| `last_used_at` | `timestamp` | no | Last successful use |
| `expires_at` | `timestamp` | no | Known expiry of the underlying secret |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Natural key: `label` unique. `secret_ref` is unique but opaque —
  it identifies a Secret Store slot, never the secret value.

### Relations

- Issued for at most one **Provider** (or a named service).
- Derives 0..n **Authentication Session** (aggregate members).
- Rotation links to at most one successor **Credential**.

### Integrity invariants

1. **INV-CRED-01** — A Credential MUST NOT store secret material in any attribute, log,
   export, or serialization. `secret_ref` is the only bridge to the secret, and it resolves
   only through the Secret Store (Volume 9). This invariant is absolute: no debug mode, no
   export option, no migration may violate it.
2. **INV-CRED-02** — `fingerprint` MUST be non-reversible and MUST NOT reveal more than
   Volume 9's display rules allow.
3. **INV-CRED-03** — Revoking or rotating a Credential MUST invalidate all Authentication
   Sessions derived from it (cascade semantics in Volume 5).
4. **INV-CRED-04** — Deleting a Credential row MUST be preceded by deletion of its Secret
   Store slot; a dangling secret without metadata is a defect (orphan-sweep owned by
   Volume 9).

### Lifecycle

Recorded status only (`active`, `rotated`, `revoked`, `expired`); acquisition, rotation, and
revocation flows are owned by Volume 5, storage by Volume 9.

### Persistence

Global database, table `credentials` (credentials are machine-level, never inside a workspace
database — workspace exports must never be able to carry credential metadata by accident).
Retention: rows persist until deleted; revoked rows are kept while audit retention requires.

### Versioning and serialization

Row versioning via `revision`. Serialization for export **excludes** `secret_ref` (the
handle is machine-local); audit exports carry `id`, `label`, `status` only.

## Authentication Session

Purpose: the state of an authenticated identity against a provider — token lifecycle, expiry,
refresh — derived from a credential.

### Attributes

| Attribute | Type | Required | Meaning |
|---|---|---|---|
| `id` | `ulid` | yes | Primary key |
| `credential_id` | `ulid` | yes | Credential this session was established from |
| `provider_id` | `ulid` | yes | Provider authenticated against |
| `state` | `enum` | yes | Canonical Authentication Session state (chapter 09) |
| `token_ref` | `string` | no | Secret Store handle for the current short-lived token material (access tokens are secrets: INV-AUTHS-02) |
| `scopes` | `json` | no | Granted scopes/permissions as reported by the provider (non-secret) |
| `established_at` | `timestamp` | no | When the session first reached `active` |
| `expires_at` | `timestamp` | no | Current token expiry, when known |
| `last_refreshed_at` | `timestamp` | no | Last successful refresh |
| `failure` | `json` | no | Last failure summary (stable error code + safe context) |
| `created_at` | `timestamp` | yes | Creation instant |
| `updated_at` | `timestamp` | yes | Last committed change |
| `revision` | `integer` | yes | Optimistic-concurrency counter |

### Identifiers

- Primary key: `id`. Provider-side session/token identifiers, when any exist, are external
  identifiers kept inside the Secret Store payload, not in this row.

### Relations

- Derived from exactly one **Credential** (aggregate member); targets exactly one
  **Provider**.

### Integrity invariants

1. **INV-AUTHS-01** — An Authentication Session MUST reference exactly one Credential and one
   Provider, and the Credential's `provider_id`/`service` MUST be consistent with that
   Provider.
2. **INV-AUTHS-02** — Token material (access tokens, refresh responses, cookies) MUST live
   only in the Secret Store via `token_ref`; this row and its serializations never contain
   it (companion to INV-CRED-01).
3. **INV-AUTHS-03** — At most one Authentication Session per (Credential, Provider) pair may
   be in a non-terminal state at any time.
4. **INV-AUTHS-04** — Session establishment and refresh MUST use official provider mechanisms
   only (Volume 1 provided constraint; flows in Volume 5); the entity records no field that
   presumes an unofficial mechanism.

### Lifecycle

Stateful — canonical states `unauthenticated`, `authenticating`, `active`, `refreshing`,
`expired`, `failed`, `revoked` (chapter 09); full machine owned by Volume 5.

### Persistence

Global database, table `auth_sessions`. Retention: terminal rows pruned per Volume 5/9 policy
once audit windows pass.

### Versioning and serialization

Row versioning via `revision`. Never exported with `token_ref`; audit exports carry state
transitions only.
