# 05 — Scope and Phase Definitions

This chapter bounds the product (scope and out of scope) and defines the delivery phases used
by every requirement in the corpus. Per the single-home matrix (Volume 0, chapter 03), **this
chapter is the authoritative definition of the phases** `Core`, `MVP`, `Beta`, `v1`, `v2`,
`Future`, and `Out of Scope`. Volume 15 sequences work within these phases; it does not
redefine them.

## Scope

Andromeda's product scope comprises the following capability areas. Each area's requirements
are minted by its owning volume (Volume 0, chapter 03); this list is the product-level
commitment that those volumes elaborate.

| Capability area | Summary | Specified in |
|---|---|---|
| Agent runtime | Agent Engine, Planner, Execution Engine, agent profiles, runs, turns, tasks | Volume 4 |
| Workflows | Declared, resumable orchestrations with approvals, including specification-driven development | Volume 4 |
| Providers and models | Provider contract, adapters, capability declaration and negotiation, routing, fallback | Volume 5 |
| Authentication | Credential acquisition, storage integration, refresh, rotation, revocation — official mechanisms only | Volume 5 (flows), Volume 9 (storage model) |
| Tools | Tool contract, Tool Runtime, built-in filesystem/terminal/Git tools, Tool SDK | Volume 6 |
| MCP | MCP client: connections, discovery, tools/resources/prompts, lifecycle, trust | Volume 6 |
| Skills and plugins | Skill format and engine; plugin runtime over the Andromeda Runtime Protocol | Volume 6 |
| Memory, context, indexing | Memory Manager, Context Manager, lexical and semantic indexes | Volume 7 |
| CLI and TUI | Command grammar, interactive terminal interface, non-interactive and CI modes | Volume 8 |
| Security | Permission model, Sandbox Engine, Secret Store, Audit Log, threat model | Volume 9 |
| Configuration, storage, observability | `andromeda.toml`, precedence, Persistence Layer, events, logging, tracing, cost records, consent-based telemetry | Volume 10 |
| Git and GitHub | Git Engine; GitHub as the project's development platform; product-side Git-hosting integrations via official APIs | Volume 11 |
| Performance and reliability | Budgets, degradation, recovery semantics | Volume 12 |
| Testing and quality | Test strategy, conformance suites, offline suite, quality gates | Volume 13 |
| Distribution | Installation, updates, rollback, signing, packaging for macOS and Linux | Volume 14 |
| Roadmap and governance | Phased execution, open-source governance, contribution model | Volume 15 |

## Out of scope

The following are excluded from the product as specified by this corpus. `Out of Scope` items
MUST NOT be implemented; reclassification requires the Volume 0 change procedure. Items marked
`Future` in the phase definitions below are, by contrast, desirable-but-uncommitted.

1. **A hosted Andromeda service (SaaS).** This corpus defines a locally installed product
   only. No core capability may require a hosted Andromeda backend.
2. **Graphical clients.** No desktop GUI, web UI, or mobile client. The terminal (CLI/TUI) is
   the product surface.
3. **Model training, fine-tuning, or weight management.** Andromeda consumes inference through
   providers; it does not train, host, or distribute models.
4. **Full code-editor or IDE functionality.** No interactive editor buffers, no
   language-server-based editing UI, no window/project management beyond the TUI's own panels.
5. **Unofficial integrations of any kind.** Reverse engineering, captured cookies, extracted
   tokens, private APIs, undocumented endpoints, web-UI automation, and any mechanism contrary
   to a service's terms — categorically excluded (provided constraint, chapter 07).
6. **A general-purpose consumer assistant.** Andromeda targets engineering work; non-technical
   consumer scenarios are not designed for.
7. **Windows versions earlier than Windows 11** for the eventual native Windows phase.

## Phase definitions

Every requirement in the corpus carries exactly one phase from this table. Phases are
cumulative delivery stages: a capability's phase is the earliest stage in which it MUST be
functional and verified.

### Core

- **Meaning:** The architectural nucleus on which everything else depends: domain model,
  public contracts (provider, tool, configuration, event, error), runtime primitives
  (sessions, runs, tasks), the Permission Manager decision path, the Persistence Layer schema
  basis, and the Platform Abstraction Layer surface.
- **Entry:** The relevant specification volumes are authored and lint-clean; the contract
  surfaces the item exposes are defined.
- **Exit:** The item is implemented, covered by unit and contract tests, and consumed by at
  least one dependent component. Core is not itself a release: every Core item ships inside
  the MVP, but is built and stabilized first, because MVP items depend on it.
- **Discipline:** A Core contract may change freely until MVP exit; from MVP exit onward,
  changes follow the deprecation rules bound to SM-20 ([chapter 06](06-success-metrics.md)).

### MVP

- **Meaning:** The first public, usable, installable release. The MVP MUST be functional,
  viable, and usable for the end-to-end jobs UC-01, UC-02, UC-03, UC-09, and UC-11
  ([chapter 03](03-use-cases.md)) — without implementing the entire product.
- **Entry:** Core exit criteria met for all Core items the MVP consumes.
- **Exit (all of the following):**
  1. Every item in the [MVP minimum](#mvp-minimum) below is functional.
  2. The acceptance test suite and the main E2E journey pass on all Tier 1 platforms
     ([platform scope](#platform-scope)).
  3. The offline guarantee list (chapter 04, Local First) passes the offline test suite with a
     local provider.
  4. Installation and basic update work on macOS and Linux per Volume 14.
  5. Releases are produced by GitHub Actions with checksums; signing per the
     [signing viability note](#signing-viability) below.
- **Discipline:** Adding scope to the MVP after Volume 15 sequences it requires the change
  procedure and a recorded justification (see the scope-creep risk, RISK-PRD-004).

### Beta

- **Meaning:** Hardening and breadth between MVP and v1: reliability and performance targets
  enforced as gates, public contracts stabilized, migration paths in place, and the extension
  surfaces exercised by real third-party extensions.
- **Entry:** MVP shipped (MVP exit criteria met and release published).
- **Exit (all of the following):**
  1. The v1 requirement set (per Volume 15) is feature-complete.
  2. Public contracts are frozen as release candidates; remaining changes are additive or go
     through deprecation.
  3. The chapter 06 metrics that bind to Beta targets meet those targets on Tier 1 platforms.
  4. Upgrade and rollback paths from every Beta build to the v1 candidate are tested.
- **Discipline:** Beta builds are public; breaking changes are permitted only with migration
  notes, and each is a recorded decision.

### v1

- **Meaning:** The first stable release. Semantic versioning guarantees attach to the public
  contracts: breaking changes only in a major release, preceded by the deprecation window
  defined in Volume 14.
- **Entry:** Beta exit criteria met.
- **Exit:** v1 is published and enters maintenance (backports and fixes per the support policy
  in Volume 15). Exit occurs when v2 ships or v1 support ends per that policy.

### v2

- **Meaning:** The next major expansion. Candidate v2 scope recorded now: native Windows 11
  support (x86_64; arm64 subject to viability), and the marketplace/distribution expansions
  Volumes 6 and 14 classify as post-v1. v2 scope is confirmed by the change procedure when v1
  ships.
- **Entry:** v1 shipped; v2 scope approved and sequenced in Volume 15.
- **Exit:** v2 published under the same stability regime as v1.

### Future

- **Meaning:** Desirable and coherent with the vision, but uncommitted: no timeline, no
  resources, no dependency from any committed phase. A committed requirement MUST NOT depend
  on a `Future` item.
- **Entry/Exit:** Not a delivery stage; items enter and leave `Future` only through the change
  procedure (typically `Future` → `v2`/later, or `Future` → `Out of Scope`).

### Out of Scope

- **Meaning:** Excluded from the product as specified. An `Out of Scope` item MUST NOT be
  implemented, and no requirement may depend on one.
- **Entry/Exit:** Classification and any reclassification occur only through the Volume 0
  change procedure, with the master traceability matrix updated accordingly.

## MVP minimum

The MVP includes, at minimum, all of the following. Each maps to the component names of the
Volume 0 glossary and to the volume that specifies it. Changing this list requires the change
procedure with recorded justification.

| # | MVP item | What "functional" means at MVP exit | Volume |
|---|---|---|---|
| 1 | Functional CLI | The command grammar's MVP subset executes with documented exit codes, JSON output, and non-interactive mode | 8 |
| 2 | Functional TUI | Interactive sessions with streaming, plan/task visibility, diff review, and permission prompts | 8 |
| 3 | Agent runtime | Agent Engine drives plan–act–observe loops over the single Runtime | 4 |
| 4 | Basic Planner | Produces inspectable, revisable plans for single-agent runs | 4 |
| 5 | Execution Engine | Executes plan tasks with task states, cancellation, and error propagation | 4 |
| 6 | Context Manager | Selects and budgets context items for model requests | 7 |
| 7 | Tool Runtime | Registers, validates, permissions, executes, and observes tools | 6 |
| 8 | Permission Manager | Evaluates permission requests with the MVP decision set; records every decision | 9 |
| 9 | Workspace Engine | Discovers and opens workspaces; manages `.andromeda/` state | 4 (behavior), 2 (entities) |
| 10 | Terminal | Terminal Engine: PTY-backed command execution with capture and limits | 6 |
| 11 | Filesystem tools | Read, write, edit, list, search within permissioned scopes | 6 |
| 12 | Basic Git | Git Engine: status, diff, stage, commit, branch, log | 11 |
| 13 | Provider abstraction | Provider contract with capability declaration; adapters as the only provider-specific code | 5 |
| 14 | At least one cloud provider | See [MVP provider seed](#mvp-provider-seed) | 5 |
| 15 | At least one local provider | See [MVP provider seed](#mvp-provider-seed) | 5 |
| 16 | Streaming | Streamed model output and live tool activity in TUI; streamed structured output in CLI | 5, 8 |
| 17 | Configuration | `andromeda.toml` with documented precedence and validation | 10 |
| 18 | Logging | Structured, redacted, local logs | 10 |
| 19 | Session persistence | Sessions and runs persist incrementally and resume after interruption | 10 (storage), 4 (semantics) |
| 20 | macOS | Tier 1 support per [platform scope](#platform-scope) | 3, 14 |
| 21 | Linux | Tier 1 support per [platform scope](#platform-scope) | 3, 14 |
| 22 | Installation | Documented install paths per Volume 14 for both Tier 1 operating systems | 14 |
| 23 | Basic update | `andromeda update`: check, download, verify, apply | 14 |
| 24 | Unit and integration tests | Test pyramid base in place with coverage gate SM-14 | 13 |
| 25 | Main E2E test | The UC-01 journey automated end-to-end | 13 |
| 26 | GitHub Actions | CI building, testing, and releasing on every mainline change | 11 |
| 27 | Signed releases (when viable) | Checksums always; signatures/notarization per [signing viability](#signing-viability) | 14 |

## MVP provider seed

The brief mandates at least one cloud and one local provider at MVP. This specification seeds
the Provider Layer with **three** adapters, a deliberate and justified widening:

| Adapter | Kind | Why it is in the seed |
|---|---|---|
| Generic OpenAI-compatible adapter | Cloud or local (endpoint-configured) | One adapter covers the many services and local servers that expose OpenAI-compatible APIs, maximizing reachable providers per unit of adapter work (SM-01) |
| Anthropic adapter | Cloud | A second, structurally different API family proves the provider contract is not shaped around one vendor's API (Principle 1); satisfies the "≥ 1 cloud provider" minimum |
| Ollama adapter | Local | Satisfies the "≥ 1 local provider" minimum and anchors the offline guarantee (UC-09) with the most common local-serving path |

**Justification for exceeding the minimum:** two cloud-facing adapters are required to verify
vendor agnosticism empirically — with a single adapter, provider-specific assumptions in the
contract are undetectable. The generic OpenAI-compatible adapter simultaneously serves cloud
and local endpoints, so the marginal cost of the third adapter is low and the de-risking value
is high. Integration of the remaining providers named in the extended integration list (OpenAI,
Google Gemini, OpenRouter, Azure OpenAI, Groq, Together, DeepSeek, xAI, Mistral, LM Studio,
vLLM, LiteLLM, llama.cpp, LocalAI, FastChat, Text Generation WebUI, and future providers) is
classified by phase in Volume 5; presence on that list does not imply MVP support.

## Platform scope

Unix is the reference behavior; all OS-specific behavior lives in the Platform Abstraction
Layer (Volume 3). Platform tiers at MVP:

| Platform | Tier | Phase | Notes |
|---|---|---|---|
| macOS on Apple Silicon (arm64) | Tier 1 | MVP | Primary development and support target |
| Linux x86_64 | Tier 1 | MVP | Reference distributions and minimum kernel defined in Volume 3 |
| Linux arm64 | Tier 1 | MVP | Same behavior contract as x86_64 |
| macOS on Intel (x86_64) | Tier 2 | MVP when viable | PENDING VALIDATION — depends on build/test capacity for the aging target; see the open-questions register ([99-volume-register.md](99-volume-register.md)) |
| Windows 11 native (x86_64; arm64 subject to viability) | — | v2 (candidate) | Native support: PowerShell, Windows Terminal, ConPTY, Windows paths, ACLs, Credential Manager — per the future Windows chapter of Volume 3 |
| Other Unix or Unix-like systems | — | Future | Considered where support cost is reasonable; no commitment |

**Tier 1** means: releases are built, tested, and supported on the platform, and the full
acceptance suite gates releases on it. **Tier 2** means: releases are built and smoke-tested;
defects are accepted but do not gate releases.

Rules:

- Windows 11 native support is a later phase. The architecture MUST NOT preclude it: the
  Platform Abstraction Layer boundaries (Volume 3) are designed against both POSIX and Windows
  primitives from the start.
- WSL is NOT a substitute for native Windows support. Andromeda running under WSL is Linux
  support in a Windows-hosted VM; it MUST be treated as a distinct modality from native
  Windows support in documentation, testing, and the platform matrix, and it MUST NOT be
  marketed as Windows support.
- Minimum OS versions, reference distributions, shells, and terminal emulators are minted by
  Volume 3, not here.

## Signing viability

MVP item 27 commits to signed releases **when viable**. Checksums (published digests for every
artifact) are unconditionally required at MVP. Cryptographic signatures and macOS
notarization additionally depend on external prerequisites — signing identities and
notarization credentials issued by third parties — whose availability for this project is
PENDING VALIDATION (open-questions register, [99-volume-register.md](99-volume-register.md)).
Volume 14 defines the signing pipeline such that enabling signatures is a configuration change,
not a redesign; until validated, releases ship with checksums and provenance metadata, and the
release notes state the signing status explicitly.
