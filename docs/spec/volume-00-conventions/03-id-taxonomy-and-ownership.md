# 03 — Identifier Taxonomy and Ownership

## Identifier formats

| Kind | Format | Scope |
|---|---|---|
| Product objective | `PRD-NNN` | Global sequence, minted by Volume 1 |
| Functional requirement | `FR-<AREA>-NNN` | Per area, minted by the owning volume |
| Non-functional requirement | `NFR-<AREA>-NNN` | Per area, minted by the owning volume |
| Risk | `RISK-<AREA>-NNN` | Per area, minted by the owning volume |
| Architecture decision | `ADR-NNN` | Global sequence, block-allocated (below) |
| Error code | `E-<AREA>-NNN` | Per area, same ownership as requirements |

`NNN` is a three-digit number starting at `001`. Identifiers are never reused or renumbered;
retired items become `DEPRECATED`, and allocation gaps are permanent.

A heading `### <ID> — <Name>` **defines** the identifier. Definitions MUST be unique across the
corpus and MUST appear only inside the owning volume's directory (ADR bodies live in
`annexes/adr/`). All other occurrences are references and MUST resolve to a definition.

## Area ownership

Each area code has exactly one minting volume. No other volume may define identifiers in that
area. `AGT` and `PLUG` extend the mandated minimum area list.

| Area | Domain | Owning volume |
|---|---|---|
| `PRD` | Product objectives | Volume 1 |
| `ARCH` | Architecture | Volume 3 |
| `PORT` | Portability, platform abstraction, platform matrix | Volume 3 |
| `AGT` | Agent engine, planner, execution engine, prompt engine | Volume 4 |
| `WF` | Workflows, specification-driven development | Volume 4 |
| `PROV` | Providers, models, capabilities | Volume 5 |
| `AUTH` | Authentication, credentials | Volume 5 |
| `TOOL` | Tools and Tool SDK | Volume 6 |
| `MCP` | Model Context Protocol support | Volume 6 |
| `SKILL` | Skills | Volume 6 |
| `PLUG` | Plugins and plugin runtime | Volume 6 |
| `SDK` | Extension SDK surfaces | Volume 6 |
| `MEM` | Memory | Volume 7 |
| `CTX` | Context management | Volume 7 |
| `IDX` | Indexing | Volume 7 |
| `UX` | Cross-cutting user experience | Volume 8 |
| `CLI` | CLI grammar and commands | Volume 8 |
| `TUI` | Terminal UI | Volume 8 |
| `SEC` | Security, permissions, sandbox | Volume 9 |
| `CFG` | Configuration | Volume 10 |
| `OBS` | Observability, logging, telemetry | Volume 10 |
| `GIT` | Git engine and integration | Volume 11 |
| `GH` | GitHub platform and development process | Volume 11 |
| `PERF` | Performance and reliability | Volume 12 |
| `TEST` | Testing and quality | Volume 13 |
| `PORT` (tests) | Portability verification is specified in Volume 3, tested per Volume 13 | — |
| `REL` | Releases, distribution, updates | Volume 14 |

## ADR number allocation

ADR numbers are allocated in blocks so concurrent authors never collide. Unused numbers in a
block remain permanent gaps.

| ADR numbers (block) | Allocated to |
|---|---|
| 001 – 039 | Foundation decisions (Volumes 0–3 authoring) |
| 040 – 054 | Volume 4 |
| 055 – 069 | Volume 5 |
| 070 – 084 | Volume 6 |
| 085 – 099 | Volume 7 |
| 100 – 114 | Volume 8 |
| 115 – 129 | Volume 9 |
| 130 – 144 | Volume 10 |
| 145 – 159 | Volume 11 |
| 160 – 174 | Volume 12 |
| 175 – 189 | Volume 13 |
| 190 – 204 | Volume 14 |
| 205 and up | Consolidation and later evolution |

## Cross-reference rules

1. A volume mints identifiers only in its owned areas.
2. A volume references other subsystems by **entity name**, **component name**, **port
   interface name** (Volume 3), or a **keystone identifier** pre-listed in Volume 0 / the
   authoring spine — never by inventing an identifier in a foreign area.
3. Requirement-level cross-links between volumes are upgraded during consolidation; until then,
   name-based references are the norm.
4. When two volumes appear to need the same requirement, the [single-home matrix](#single-home-matrix)
   decides the owner; the other volume references it.

## Single-home matrix

Each cross-cutting topic has exactly one authoritative home. Other volumes MUST reference, not
restate. (Volume 2 owns each entity's *state names*; the owning volume owns the full machine.)

| Topic | Authoritative home |
|---|---|
| Permission model, permission enum, decision semantics | Volume 9 |
| Sandbox model | Volume 9 |
| Credential storage | Volume 9 (model) / Volume 5 (provider auth flows) |
| Configuration schema, precedence, `andromeda.toml` | Volume 10 |
| Config key content per TOML table | Area owner (see below) |
| Capability enum and negotiation | Volume 5 |
| Tool contract | Volume 6 |
| Provider contract | Volume 5 |
| Specification-driven development workflow | Volume 4 |
| State machines | Owning volume of the entity's area |
| Event envelope and delivery semantics | Volume 10 (envelope) / area owners (event content) |
| Error envelope and exit codes | Volume 0 (scheme) / area owners (individual errors) |
| CLI command names and grammar | Volume 8 |
| Design tokens and theming | Volume 8 |
| Phase definitions (Core/MVP/Beta/v1/v2/Future/Out of Scope) | Volume 1 |
| Traceability chain and validation | Volume 0 / Volume 11 (automation) |

## Namespace conventions

### Error codes

`E-<AREA>-NNN`, same area ownership as requirements. Every error defined in the corpus MUST
declare: stable code, category, severity, user message, technical message, cause,
safe-to-log data, recoverability, retry policy, recommended action, exit-code mapping, HTTP
mapping where applicable, telemetry event, and security implications. The consolidated catalog
lives in `annexes/catalog-errors.md`.

### Events

Event names use `<area>[.<noun>].<verb-past>` in lowercase dot notation; the `<noun>` segment
is omitted when the event's subject is the area entity itself — e.g., `run.completed`,
`tool.invocation.denied`, `provider.request.failed`. The event envelope (version, producer,
correlation ID, timestamp, ordering, delivery semantics, persistence, retention, privacy,
redaction, compatibility, failure behavior) is defined in Volume 10; each area owner mints its
event names and payloads. Consolidated catalog: `annexes/catalog-events.md`.

### Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error |
| 2 | Usage error (invalid arguments or flags) |
| 3 | Configuration error |
| 4 | Authentication error |
| 5 | Permission denied (by Andromeda's permission model) |
| 6 | Tool execution failure |
| 7 | Provider failure |
| 8 | Timeout or cancellation |
| 9 | Integrity error (corrupted state, failed migration) |

Every CLI command definition in Volume 8 MUST state which exit codes it can produce.

### Configuration tables

Volume 10 owns the configuration schema, precedence, validation, and migration rules. The *key
content* of each TOML table belongs to its area owner:

| TOML table | Content owner |
|---|---|
| `[agent]`, `[workflows]` | Volume 4 |
| `[providers]`, `[providers.*]`, `[auth]` | Volume 5 |
| `[tools]`, `[plugins]`, `[mcp]`, `[skills]` | Volume 6 |
| `[memory]`, `[context]`, `[index]` | Volume 7 |
| `[cli]`, `[tui]`, `[tui.theme]` | Volume 8 |
| `[permissions]`, `[sandbox]`, `[security]` | Volume 9 |
| `[logging]`, `[telemetry]`, `[storage]` | Volume 10 |
| `[git]`, `[github]` | Volume 11 |
| `[update]` | Volume 14 |

### Environment variables

Environment variables use the `ANDROMEDA_` prefix and map to configuration keys by uppercasing
the dotted path with `_` separators and `__` for table nesting where ambiguous — the exact
mapping algorithm is defined in Volume 10. Example: `ANDROMEDA_TUI_THEME_MODE` ↔
`tui.theme.mode`.

### Permissions and capabilities

The permission enum (canonical `snake_case` names) is minted in Volume 9; the model/provider
capability enum is minted in Volume 5. Both are closed vocabularies: new values require an ADR.

## Traceability chain

Every requirement participates in the chain:

```text
Objective (PRD-NNN)
  → Requirement (FR-*/NFR-*)
  → Epic
  → Issue
  → Branch
  → Commit
  → Pull Request
  → Test
  → Check
  → Release
  → Artifact
```

Volume 0 chapter 09 holds the master matrix (objective → requirement → phase → verification);
Volume 11 defines the GitHub-side automation that enforces the chain from Epic onward
(issue/PR templates, branch naming, commit trailers, required checks).
