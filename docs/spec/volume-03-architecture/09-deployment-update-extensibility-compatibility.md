# 09 — Deployment, Update, Extensibility, and Compatibility

This chapter closes the volume outward: how the architecture deploys onto a machine, where
its state lives, how the Updater plugs into it, where third parties extend it, and how all of
its public contracts stay compatible over time. Distribution mechanics (installers, channels,
signing) are Volume 14's; extension contract details are Volume 6's; this chapter fixes the
architectural shape they elaborate.

## Deployment shapes

Andromeda deploys as **one installable executable** (`andromeda`, PRD-011) containing all 37
components, built-in tools/profiles/workflows/skills (in binary, per Volume 2's persistence
map), embedded migrations (ADR-029), and the default prompt templates. There are exactly two
deployment shapes, differing only in how the process is driven:

| Shape | Drivers active | Phase | Notes |
|---|---|---|---|
| Interactive / one-shot | CLI, TUI; IPC socket also available | MVP | Includes non-interactive one-shot CLI runs (Volume 8) |
| Persistent headless instance | IPC socket only | Beta | FR-ARCH-008, ADR-032; started and owned by its invoker — never a self-installed daemon |

No hosted service, no daemon registration, no privileged helper processes exist in any shape
(Volume 1 out-of-scope list; process-model rules, chapter
[08](08-processes-concurrency-ipc.md)).

### State footprint

Everything Andromeda writes, and where (per ADR-022 resolution and ADR-028 topology; exact
per-platform paths resolve via FR-PORT-003):

| Data | Location | Lifecycle |
|---|---|---|
| Global configuration `andromeda.toml` | Platform config directory | User-edited; migrated N-1 per the compatibility strategy below |
| Global database `global.db` | Platform data directory | Authoritative; forward-only migrations (ADR-029); credentials metadata global-only (ADR-028) |
| Workspace state `.andromeda/state.db` + `artifacts/` | Workspace root | Authoritative; travels with the workspace; independent migration age (ADR-029) |
| Index caches `.andromeda/index.db` | Workspace root | Rebuildable; dropped on layout change (ADR-028 rule 4) |
| Secret material | OS credential store or opt-in age-encrypted file (ADR-014) | Never in any database or export |
| Logs | Platform log location per ADR-022 | Rotated/retained per Volume 10 |
| Caches (downloads, staging) | Platform cache directory | Disposable at any time |
| Runtime (IPC socket, locks) | Per-instance runtime directory | Exists only while an instance runs |
| Retained previous version | Platform data directory (Updater area) | Enables offline rollback (SM-19) |

Uninstallation (procedure owned by Volume 14) removes the binary and, on user request, the
global footprint; workspace `.andromeda/` directories belong to their workspaces and are
never touched by product uninstall.

## Updater integration points

Volume 14 owns update behavior end-to-end (channels, cadence, consent, delta strategy,
signing per ADR-013). The architecture fixes the integration points it operates through:

1. **UpdaterPort** (chapter [02](02-port-interfaces.md)) is the only self-update surface;
   drivers and the scheduled check consume it.
2. **PAL Installer/Updater surfaces** perform detection and the atomic swap:
   package-manager-owned installs (Homebrew et al.) are detected and self-update defers to
   the owning manager (chapter 07 surface notes); direct installs use atomic
   replace-or-restore with a retained previous version.
3. **Update is permission-mediated and observable** like everything else: applying an update
   is a side-effecting action through PermissionPort, produces the frozen Update-process
   states as events, and lands in update history (global DB).
4. **Extension compatibility on update**: before `Apply`, the Updater consults PackagePort
   for installed extensions' declared compatibility (Volume 6 manifest field against the new
   version's contract versions) and reports incompatibilities per Volume 14 policy — the
   architecture guarantees the question is answerable from local registries, offline.
5. **Rollback is offline** (SM-19): retained artifacts plus forward-only databases mean
   rolling back the binary never requires network, and an older binary refuses newer schemas
   cleanly (ADR-029) rather than corrupting them.

Point 5 has one deliberate consequence: a rollback across a schema migration boundary
requires restoring the pre-migration backup (ADR-029 creates it automatically). Volume 14's
rollback procedure surfaces exactly this distinction.

## Extension points

The extension surfaces of Principle 6/PRD-007, each mapped to its architectural mechanism and
owning volume. This table is the authoritative map; FR-ARCH-011 binds its guarantees.

| Extension point | Mechanism (port / SDK surface) | Delivery | Owning volume |
|---|---|---|---|
| Providers | ProviderPort adapter: in-tree or plugin-provided via ARP provider surface | In-tree contribution; plugin package | Volume 5 |
| Tools | ToolPort via Extension SDK (in-tree built-ins), plugins (ARP), MCP servers | SDK; plugin/MCP packages | Volume 6 |
| Skills | Skill format loaded by the Skill Engine | Skill packages; workspace-local | Volume 6 |
| Workflows | Workflow definition format executed by the Workflow Engine | Packages; workspace-local | Volume 4 (format), Volume 6 (packaging) |
| Prompts | Prompt Engine template registry slots | Skills; packages; workspace overrides | Volume 4 |
| Indexers | IndexerPort implementations | In-tree; plugin-provided per Volume 7 phasing | Volume 7 |
| Storage | MemoryStorePort / port-level backend substitution | In-tree per Volume 7/10 phasing | Volumes 7/10 |
| Authentication | Auth mechanisms accompanying provider adapters (AuthPort flows) | With provider adapters | Volume 5 |
| Telemetry exporters | Telemetry pipeline exporter registration (ADR-011) | Plugin-provided; configuration | Volume 10 |
| Git integrations | Hosting integrations above GitPort (official APIs only) | In-tree; plugin-provided | Volume 11 |
| Commands | CLI command contributions | Plugin packages | Volume 8 (grammar), Volume 6 (mechanism) |
| TUI panels | Panel extension surface where viable | Phased per Volume 8 | Volume 8 |
| Policies | Policy vocabulary extensions (ADR-gated) | Configuration; Volume 9 process | Volume 9 |

Three architectural guarantees hold across every row: extensions implement **existing port
contracts** (no extension-only parallel APIs); extensions are subject to the **same
permission, sandbox, and observability regime** as built-ins (Principle 4's citizenship rule
generalized); and extension *distribution* always flows through the Package Manager's frozen
installation states, whatever the surface.

## Compatibility strategy

Compatibility is governed by ADR-015 (SemVer + Conventional Commits) applied to the **public
contract set** — the SM-20 list: provider contract, tool contract, plugin protocol (ARP),
skill format, workflow format, configuration schema, CLI structured-output schema, and the
event envelope — plus, from this volume, the port interfaces (`internal/ports`/`sdk/`
mirrors, NFR-ARCH-002) and the IPC protocol version (chapter 08).

| Contract family | Regime |
|---|---|
| Port interfaces and SDK | Frozen names/signatures (FR-ARCH-003); additive in minors; breaking only in a major with ≥ 1 minor deprecation window (NFR-ARCH-002, SM-20) |
| ARP and IPC protocol | Versioned handshakes (ADR-009, ADR-012); servers/instances support the current and previous protocol minor within a major (clients see E-ARCH-004 with the supported range otherwise) |
| Configuration (`andromeda.toml`) | **N-1 compatibility with migrations**: version N MUST read every configuration valid for N−1, migrating it per Volume 10's migration rules (automatic where lossless, guided otherwise) — deprecated keys warn for at least one minor before removal at a major (ADR-029's forward-only philosophy applied to config by Volume 10) |
| Databases | Forward-only migrations, pre-migration backups, clean refusal of future schemas with exit code 9 (ADR-029); workspace and global databases migrate independently |
| Serialized documents / events | `schema_version` per document kind; additive evolution preferred; readers reject unknown newer versions rather than guess (Volume 2 chapter 10) |
| Extension packages | Declare compatible contract version ranges in manifests (Volume 6); checked at install (PackagePort `Resolve`) and at product update (Updater integration point 4) |

The corpus-side counterpart: contract-diff tooling in CI (SM-20 measurement method) plus the
release audit (Volume 14) verify the regime mechanically; deviations are release blockers.

## Requirements

### FR-ARCH-011 — Extension through versioned public contracts

- Type: Functional
- Status: Draft
- Priority: P0
- Phase: Core
- Source: Provided
- Owner: Architecture (Volume 3)
- Affected components: Extension SDK, Plugin Runtime, MCP Runtime, Package Manager, all port-implementing components
- Dependencies: FR-ARCH-002, FR-ARCH-003; ADR-009, ADR-015; PRD-007
- Related risks: RISK-ARCH-002

#### Description

Every extension point in the [extension points](#extension-points) table MUST be reachable
exclusively through a versioned public contract — a port interface, the ARP protocol, or a
declared format — with no extension-only side doors: an extension MUST NOT be able to do
anything a built-in implementation of the same contract cannot, and MUST NOT bypass the
permission, sandbox, or observability regime of its surface. Each extension point's owning
volume specifies its contract under this rule.

#### Motivation

PRD-007's platform promise is only real if extension capability equals built-in capability
under equal governance; side doors would fork the security model exactly where third-party
code enters.

#### Actors

Extension authors; SDK; Plugin Runtime/MCP Runtime/Package Manager; owning volumes.

#### Preconditions

Chapter 02 ports and the SDK exist for the surface in question.

#### Main flow

1. An extension implements a contract (port via SDK, ARP surface, or format).
2. It is delivered through the Package Manager's installation states.
3. Its runtime registers it with the governing component; from registration onward it is
   indistinguishable in governance from a built-in.

#### Alternative flows

- Contract version mismatch at install or update: PackagePort resolution reports the
  incompatibility (compatibility strategy above); nothing partially activates.

#### Edge cases

- Extensions bundling several surfaces (a plugin providing tools and a telemetry exporter)
  register each surface with its own governing component; failure of one surface's
  registration does not half-activate the others (frozen Package installation states).
- Workspace-local, package-less extensions (local skills/workflows) still enter through
  format validation and trust policy — delivery differs, governance does not.

#### Inputs

Extension artifacts with manifests; contract version declarations.

#### Outputs

Registered, governed extensions; rejection reports on incompatibility.

#### States

Package installation machine (frozen, Volume 2); per-surface entity states (Plugin, MCP
Client Connection).

#### Errors

E-PLUG/E-MCP/E-SKILL/E-SDK families per surface (Volume 6); E-ARCH-002 when an extension
breaches its port contract at runtime.

#### Constraints

No `internal/` access from extensions (ADR-031); contract changes follow the compatibility
strategy of this chapter.

#### Security

Uniform citizenship: extension code executes under sandbox policy (plugins/MCP as
subprocesses per ADR-009/ADR-010), its capabilities are permission-mediated, and origin plus
trust level are always visible (Principle 4).

#### Observability

Extension-provided operations carry origin attribution in every record; registration and
failure events per surface family.

#### Performance

Extension overhead budgets (ARP round-trip) are Volume 12's per ADR-009's review condition.

#### Compatibility

Manifest-declared contract ranges checked at install and update (compatibility strategy);
SM-20 governs contract evolution.

#### Acceptance criteria

- Given the SM-16(b) enforcement test executed against a plugin-provided tool, when it
  attempts an unmediated side effect, then the result is identical to a built-in tool's:
  impossible or denied-and-recorded.
- Given a third-party provider adapter passing the Volume 5 conformance suite, when used in
  a run, then all Principle 7 attribution (provider, model, cost, capability set) is present
  exactly as for a seed adapter.
- Given an extension built against contract version X, when installed on a product whose
  supported range excludes X, then resolution fails with the declared incompatibility error
  and nothing activates.
- Negative case: given a plugin attempting an ARP method outside its negotiated
  capabilities, when invoked, then the Plugin Runtime rejects it and records the violation.

#### Verification method

Conformance suites per surface (Volume 13) executed against reference third-party extensions;
the SM-16(b) test matrix extended to extension origins; install/update compatibility test
fixtures; SM-02/SM-03 timed exercises validating the SDK path end-to-end.

#### Traceability

PRD-004, PRD-007; SM-02, SM-03, SM-15, SM-16, SM-20; ADR-009, ADR-010, ADR-015; FR-ARCH-002,
FR-ARCH-003; extension points table.
