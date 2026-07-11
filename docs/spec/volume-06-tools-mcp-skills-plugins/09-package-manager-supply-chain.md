# 09 — Package Manager and Supply Chain

The **Package Manager** (Volume 3, chapter 04) moves extension packages — kinds `plugin`,
`skill`, and `bundle` (Volume 2, Package entity) — from sources onto the machine and back
off: discovery, resolution, installation, verification, update, and removal, through the
frozen Package installation states whose full machine is in
[chapter 10](10-state-machines.md). This chapter owns the behavioral contract of
**PackagePort** (Volume 3, chapter 02) for extension packages; product binary updates are a
different process (UpdaterPort and the Update machine, Volume 14). Distribution design
follows [ADR-080](../annexes/adr/ADR-080.md) (sources now, marketplace Future),
[ADR-081](../annexes/adr/ADR-081.md) (signature policy), and
[ADR-083](../annexes/adr/ADR-083.md) (dependency resolution).

Boundary rules inherited from the corpus:

1. **Verification precedes activation.** No package content executes or registers before its
   checksum verifies (INV-PKG-01) and its signature policy passes (INV-PKG-02); the
   verification order is fixed in FR-PLUG-007.
2. **Permissioned mutation.** Install, update, and remove require the `package_installation`
   permission; acquisition from non-local sources additionally requires `network` under
   Volume 9 scope semantics (`domain` scope for the source host).
3. **No implicit operations.** The Package Manager never installs, updates, or removes
   anything except on explicit user command or recorded policy decision; there are no
   background update checks for extensions.
4. **Provenance is permanent.** Every package row records its source locator, checksum,
   signature state, and resolution plan; removal tombstones rather than erases (Volume 2
   retention rules), keeping the chain extension → package → source navigable offline
   (SM-13 pattern).

## Package format

A package is a gzip-compressed tar archive containing a `package.toml` metadata document at
the archive root plus the payload directory for its kind (a skill directory per ADR-078, a
plugin install tree per chapter 08, or bundle members). Archive naming:
`<name>-<version>-<kind>[-<os>-<arch>].tar.gz` — the platform suffix is present exactly when
the payload is platform-specific (plugin executables; skills are platform-independent).

```toml
[package]
name = "conventional-review"
version = "1.2.0"
kind = "skill"                      # plugin | skill | bundle
description = "Reviews diffs against project conventions."
publisher = "Ada Contributor <ada@example.com>"
license = "Apache-2.0"
homepage = "https://example.com/conventional-review"
contract_version = "1.0"            # public contract targeted (SM-20 tracking)

[package.dependencies]
# Package-name -> semver range. Permitted for kinds skill and bundle;
# plugins declare no package dependencies in contract 1.0 (ADR-083).
"diff-reading" = ">=1.0.0 <2.0.0"

[package.platforms]
# Required for kind = "plugin"; omitted for skills.
supported = []                      # e.g., ["darwin-arm64", "linux-amd64", "linux-arm64"]
```

The following metadata document is invalid — unknown kind, malformed range — and MUST fail
resolution with E-PLUG-009 (metadata findings are resolution-time findings; nothing is
downloaded for a package whose index entry is already malformed, and a malformed in-archive
document fails verification per FR-PLUG-007):

```toml invalid
[package]
name = "bad"
version = "1.0.0"
kind = "theme"                      # not a Volume 2 package kind

[package.dependencies]
"diff-reading" = "newest"           # not a semver range
```

Rules: `name` follows the skill/plugin name grammar (chapter 07), with the `andromeda-`
prefix reserved for official packages (Volume 2; unsigned use of the prefix is `invalid` per
ADR-081); `version` is SemVer (ADR-015); `kind` is the closed Volume 2 vocabulary; the
in-archive `package.toml` MUST agree with the index entry that advertised it (name, version,
kind) — disagreement is a verification failure (E-PLUG-011).

## Sources and the registry index

Acquisition operates against **declared package sources** (ADR-080), configured per scope as
entries of `plugins.sources` and `skills.sources` (keys minted in chapters 07 and 08; entry
schema minted here — both keys share it):

| Source entry field | Type | Required | Meaning |
|---|---|---|---|
| `name` | string | yes | Source identifier, unique per scope |
| `kind` | string | yes | `registry` \| `git` \| `archive` \| `path` (Volume 2 vocabulary) |
| `location` | string | yes | Index URL (`registry`), repository URL (`git`), archive URL/path (`archive`), directory (`path`) |
| `priority` | int | no (default 100) | Lower consults first (ADR-083 selection) |
| `enabled` | bool | no (default `true`) | Source participates in discovery/resolution |
| `signature_required` | bool | no (default `false`) | Absence of a signature counts as `invalid` for packages from this source (ADR-081 downgrade defense) |
| `timeout_ms` | int | no (default 300000) | Acquisition budget per fetch from this source |

```toml
[[plugins.sources]]
name = "team-mirror"
kind = "registry"
location = "https://mirror.example.internal/andromeda/index.json"
priority = 10
signature_required = true
```

A `registry` source's location resolves to a **registry index**: a JSON document (validated
per ADR-024 against the index schema shipped with the binary) enumerating packages:

```json
{
  "schema_version": "1.0",
  "generated_at": "2026-07-10T12:00:00Z",
  "packages": [
    {
      "name": "conventional-review",
      "version": "1.2.0",
      "kind": "skill",
      "archive": "https://mirror.example.internal/pkgs/conventional-review-1.2.0-skill.tar.gz",
      "checksum": "sha256:6b1f4d8f4a3f0c2f9d8e7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b7c6d",
      "signature": "https://mirror.example.internal/pkgs/conventional-review-1.2.0-skill.sig",
      "publisher": "Ada Contributor <ada@example.com>",
      "contract_version": "1.0"
    }
  ]
}
```

Index rules: `schema_version` follows the SM-20 additive-evolution regime; entries lacking a
`checksum` are unusable (resolution skips them with a recorded finding); `signature` is a
locator for the cosign signature bundle where the publisher signs (ADR-081). `git` sources
are fetched through the Git Engine (GitPort; system git per ADR-025) and MUST contain either
an index document or package archives at documented paths; `path` and `archive` sources
bypass discovery and feed resolution directly. An **official Andromeda registry index** is
PENDING VALIDATION per ADR-080 (hosting is an organizational decision; tracked in this
volume's register); no default `registry` source ships enabled until it resolves.

### FR-PLUG-006 — Package sources, discovery, and dependency resolution

- Type: Functional
- Status: Approved
- Priority: P1
- Phase: Beta
- Source: Provided
- Owner: Package Manager (Volume 6)
- Affected components: Package Manager, Git Engine (git sources), Configuration Manager, CLI
- Dependencies: FR-PLUG-005; ADR-080, ADR-083, ADR-024, ADR-025
- Related risks: RISK-PLUG-003
- Keystone: none (supports FR-PLUG-001, FR-SKILL-001 distribution)

#### Description

The Package Manager MUST implement the source model above: the four source kinds with the
entry schema fields, per-scope configuration, and priority ordering. Discovery (list/search
by name, kind, and text over descriptions) MUST operate only on explicit command, over
enabled sources in priority order, using cached index documents when offline and marking
staleness. `PackagePort.Resolve` MUST implement ADR-083 exactly: transitive closure bounded
at depth 5 and 64 nodes, version-range intersection per package name, conflict failure with
every colliding constraint named, highest-satisfying-version selection with
first-source-wins by priority, deprecation exclusion with only-satisfier warning,
satisfied-in-place recognition of installed versions, and a complete resolution plan
(actions, exact versions, source locators, expected checksums, expected signature
references) as the sole output — with no side effects and no network beyond reading enabled
sources' indexes.

#### Motivation

Deterministic, explainable acquisition is the supply-chain control point: what gets
installed must be a pure, recorded function of what the user asked for and what the
configured sources offer.

#### Actors

Users running discovery/install commands; the Package Manager; configured sources; the Git
Engine for `git` sources.

#### Preconditions

At least one enabled source (for discovery/registry resolution); resolved configuration
available; for network sources, `network` permission per Volume 9.

#### Main flow

1. Discovery reads (cached) indexes from enabled sources in priority order and presents
   matches with kind, versions, publisher, and signature availability.
2. A resolution request names a package (and optionally a version or range).
3. `Resolve` builds the closure, intersects constraints, selects versions and sources, and
   returns the plan for user inspection (or immediate execution when the command combines
   resolve+install).

#### Alternative flows

- Offline with cached indexes: resolution proceeds against caches; the plan marks index age;
  acquisition of non-local archives fails later with E-PLUG-008 unless artifacts are
  pre-fetched.
- A source fails to load: discovery/resolution proceed over the remaining sources, recording
  the failure (E-PLUG-008 as a finding, not a fatal error, when other sources satisfy the
  request).

#### Edge cases

- The same name+version in two sources with different checksums: resolution selects by
  priority and records the divergence as a warning finding — a dependency-confusion signal
  (RISK-PLUG-003).
- A closure exceeding 64 nodes or depth 5 fails with the offending chain listed
  (E-PLUG-009).
- A range that only a deprecated version satisfies resolves with a deprecation warning
  carried into the plan (ADR-083 rule 3).
- Duplicate source names within one scope are a configuration validation error (Volume 10
  validation semantics; exit code 3 surface).

#### Inputs

Package requests (name, optional range), source configuration, index documents, installed
package registry.

#### Outputs

Discovery listings; resolution plans (recorded with subsequent installations);
`package.resolution.completed` events; findings.

#### States

Resolution occurs in the `resolving` state of the chapter 10 machine when driven by an
installation; standalone resolution (dry-run) touches no machine.

#### Errors

E-PLUG-008 (source unavailable), E-PLUG-009 (resolution failed). Configuration-shape errors
surface through Volume 10 validation.

#### Constraints

No network at startup; no resolution during installation (plans are inputs, ADR-083 rule 5);
index documents validated against the shipped schema before use (ADR-024).

#### Security

Sources are trust-relevant configuration: priority decides dependency-confusion outcomes;
`signature_required` per source hardens against signature stripping (ADR-081); index
documents are untrusted input (schema-validated, size-bounded, never executed).

#### Observability

`package.resolution.completed` carries the plan summary (package count, sources consulted,
cache ages); discovery and resolution findings appear in structured logs; plans are
persisted with installations for audit.

#### Performance

Resolution latency per NFR-PLUG-003(c); discovery over cached indexes performs no network
round trips.

#### Compatibility

Index `schema_version` follows additive evolution (SM-20); unknown index fields are ignored;
an index major above the supported one is rejected with the versions named (E-PLUG-008
compatibility class).

#### Acceptance criteria

- Given two enabled sources offering versions 1.1.0 and 1.2.0 of a requested package with
  range `>=1.0 <2.0`, when resolution runs, then the plan selects 1.2.0 from the
  priority-first source offering it, and re-running with identical inputs yields a
  byte-identical plan.
- Given dependencies whose ranges intersect to empty, when resolution runs, then E-PLUG-009
  names every colliding constraint and its declaring package, and no plan is produced.
- Negative case: given an index document failing schema validation, when discovery reads it,
  then the source is skipped with a recorded finding and other sources still serve.
- Permission case: given a `registry` source over HTTPS and no `network` grant for its
  domain, when discovery fetches, then the fetch is denied (E-SEC surfaced), cached data (if
  any) is used with staleness marked, and the denial is recorded.
- Observability case: `package.resolution.completed` is emitted exactly once per completed
  resolution with the plan hash.

#### Verification method

Resolver property tests (determinism, bound enforcement, intersection correctness) over
fixture indexes; multi-source integration tests including divergent-checksum and
schema-invalid indexes; offline-cache tests; permission-denial tests.

#### Traceability

ADR-080, ADR-083; FR-PLUG-005, FR-PLUG-007; Volume 2 Package `source` vocabulary;
RISK-PLUG-003.

## Package operations

Operations map to PackagePort and the frozen installation states (full machine in
[chapter 10](10-state-machines.md)):

| Operation | PackagePort path | State path | Permissions |
|---|---|---|---|
| Discover/search | (index reads only) | none | `network` for remote index refresh |
| Resolve (dry run) | `Resolve` | none | `network` for index refresh |
| Install | `Resolve` + `Install` | `requested` → … → `installed` | `package_installation`; `network` for remote acquisition |
| Verify | `Verify` | none (reads `installed`) | none (local reads) |
| Update | `Resolve` + `Install` of the newer version, then supersession of the prior | new row `requested` → … → `installed`; prior row per coexistence rules | `package_installation`; `network` |
| Uninstall | `Remove` | `installed` → `removed` | `package_installation` |

Registration of the delivered extensions (skill registration per FR-SKILL-001, plugin
registration per FR-PLUG-001, MCP server registration per chapter 05) happens during
`installing` and is part of the same failure atom: if any registration fails, the
installation fails and nothing stays partially active.

### FR-PLUG-005 — Extension package operations

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: Package Manager (Volume 6)
- Affected components: Package Manager, Skill Engine, Plugin Runtime, MCP Runtime, Permission Manager, Persistence Layer
- Dependencies: FR-PLUG-006, FR-PLUG-007; ADR-080, ADR-083; FR-SKILL-001, FR-PLUG-001
- Related risks: RISK-PLUG-003

#### Description

The Package Manager MUST implement install, update, verify, and remove for extension
packages through PackagePort semantics and the frozen installation states: `Install`
executes a previously produced resolution plan action by action, streaming `InstallEvent`
progress, acquiring archives into a staging area, verifying per FR-PLUG-007, placing files
recorded in `files_manifest` (INV-PKG-03), registering delivered extensions, and reaching
`installed` — or failing with nothing partially active (`failed`), or reverting an upgrade
to the prior good state (`rolled_back`). Updates are explicit: the user requests an update
(optionally after a manual check command); the newer version installs as a new Package row
and the prior version is superseded per the kind's coexistence rules (plugins: single active
version per name/scope, prior row `removed` after successful swap; skills: prior versions
remain registered per Volume 2 retention). `Remove` MUST deregister delivered extensions
(cascade per INV-EXT-03), delete exactly the files in `files_manifest`, tombstone the
Package row (`removed`), and refuse removal that would break an installed dependent's
resolution (conflict finding, E-PLUG-012) unless the user forces cascade removal explicitly.
Concurrent mutating operations on one package are serialized; independent packages may
proceed concurrently under scheduler pools (Volume 3 PackagePort concurrency).

#### Motivation

The operations are the product face of the supply chain: every extension on the machine got
there through this pipeline, which is what makes provenance, verification, and clean removal
provable properties rather than conventions.

#### Actors

Users running package commands; the Package Manager; governing components registering
delivered extensions; the Permission Manager.

#### Preconditions

A resolution plan (install/update); `package_installation` permission decision; staging
space available.

#### Main flow

1. Plan actions execute in order: acquire → verify → stage → install → register.
2. Each state transition persists before the next step (chapter 10 persistence rules) and
   emits its event.
3. On success the row is `installed`, `files_manifest` is complete, and delivered extensions
   are registered and enabled per their kinds' defaults.

#### Alternative flows

- Failure before `installing`: transition to `failed`; staging content discarded; nothing
  registered (INV-PKG-01 guarantee).
- Failure during an upgrade's `installing`: revert to the prior good state using the prior
  version's `files_manifest`; new row `rolled_back`; prior version remains `installed`.
- User cancellation mid-operation: the operation stops at the next safe boundary, cleanup
  runs, and the row records `failed` with the cancellation reason (E-PLUG-013).

#### Edge cases

- Installing a version already installed at the same scope: no-op reported as satisfied
  (plan marks it; exit code 0).
- Installing the same name at workspace scope over a global install: permitted; shadowing is
  reported at registration (chapter 07 rule for skills; plugins follow scope precedence at
  surface registration).
- Removal of a package whose plugin is `running`: the plugin is stopped through its machine
  (chapter 10) before file deletion; a plugin that cannot be confirmed terminated blocks
  removal (INV-PLG-04, E-PLUG-012).
- A `bundle` installs members atomically: any member failure fails the bundle with nothing
  partially active.

#### Inputs

Resolution plans, archives, permission decisions, staging area, prior `files_manifest`
(upgrades, removals).

#### Outputs

Installed/updated/removed packages; registered/deregistered extensions; `package.*` events;
recorded plans and manifests.

#### States

The full chapter 10 Package installation machine; this requirement defines the operations
that drive it.

#### Errors

E-PLUG-008 – E-PLUG-013 per step; E-SEC family for permission denials; registration-step
errors surface from the delivered kind (E-SKILL-001/E-PLUG-006 classes) and fail the
installation.

#### Constraints

No resolution during installation; verification order fixed (FR-PLUG-007); one mutating
operation per package at a time; staging areas live under the scope's Andromeda state
directory (ADR-022 layout) and are cleaned on completion and at startup recovery.

#### Security

Every mutating operation requires `package_installation` (recorded decision); acquisition
respects `network` scopes; archives are untrusted input until verified — extraction enforces
path containment (no absolute paths, no `..`, symlinks not followed) with violations failing
verification (E-PLUG-011).

#### Observability

Every state transition emits exactly one `package.*` event with the package ULID, name,
version, and correlation IDs; operations append Audit Records (permission decisions,
verification outcomes); `InstallEvent` streams progress to the CLI/TUI live.

#### Performance

Per NFR-PLUG-003; staging-to-install is dominated by hashing and file placement, not
recomputation.

#### Compatibility

`contract_version` recorded per Extension (INV-EXT-04) at registration; installing a package
targeting an unsupported contract version fails resolution/verification with both versions
named (compatibility class of E-PLUG-009/E-PLUG-011).

#### Acceptance criteria

- Given a valid plan for a skill package from a `path` source, when installed, then the row
  reaches `installed`, `files_manifest` lists every written file with digests, the skill
  registers per FR-SKILL-001, and `package.installation.completed` is emitted once.
- Given an upgrade whose registration step fails, when installation aborts, then the new row
  is `rolled_back`, the prior version's files verify intact (`Verify` passes), and the prior
  extensions remain registered and enabled.
- Negative case: given a plan whose archive checksum mismatches at verification, when
  installation runs, then the row is `failed`, staging is empty afterwards, and nothing was
  registered (E-PLUG-011).
- Permission case: install without `package_installation` fails with the E-SEC denial
  recorded and no state machine instance created.
- Error case: removal of a package that another installed package depends on fails with
  E-PLUG-012 naming the dependents; forced cascade removal removes both with both tombstones
  recorded.
- Observability case: the full audit chain for an installation (decision → plan →
  verification → placement → registration) is navigable offline from the Package row.

#### Verification method

End-to-end operation tests over fixture sources and archives (all four source kinds);
crash-injection at every state boundary (chapter 10 recovery assertions); upgrade/rollback
tests; cascade-removal tests; audit-chain inspection per SM-13 pattern.

#### Traceability

PRD-007; ADR-080, ADR-081, ADR-083; INV-PKG-01..04, INV-EXT-03/04; chapters 07, 08, 10;
Volume 14 (product updates excluded).

## Verification and integrity

### FR-PLUG-007 — Package verification and integrity

- Type: Functional
- Status: Approved
- Priority: P0
- Phase: Beta
- Source: Provided
- Owner: Package Manager (Volume 6)
- Affected components: Package Manager, Policy Engine, Audit Log
- Dependencies: FR-PLUG-005; ADR-081; ADR-013 (tooling patterns)
- Related risks: RISK-PLUG-003

#### Description

Verification MUST run in fixed order during the `verifying` state, entirely before anything
is staged for activation: (1) **checksum** — SHA-256 of the acquired archive equals the
plan's expected checksum (INV-PKG-01; mismatch is E-PLUG-011, terminal for the attempt);
(2) **signature** — `signature_state` computed per ADR-081 (`verified` | `unverified` |
`invalid`), with `invalid` blocking unconditionally (INV-PKG-02) and `unverified` gated by
the trust policy in force (interactive default: prompt with provenance; non-interactive:
deny per PRD-009); (3) **content** — the archive extracts with path containment enforced,
the in-archive `package.toml` validates and agrees with the index entry, and the payload
validates for its kind (skill manifest per FR-SKILL-001, plugin manifest per FR-PLUG-001,
bundle member list). Verification MUST work offline when the artifacts, signature bundles,
and trust material are locally present (ADR-081 rule 5). Post-install,
`PackagePort.Verify` MUST re-check an installed package at any time: every `files_manifest`
entry present with matching digest, manifest consistency, and recorded signature state —
reporting tamper findings without modifying state (remediation is a user decision;
findings are audit-logged and surfaced).

#### Motivation

Verification is the enforcement point for the whole supply chain: everything downstream
(trust classification, activation policy, provenance) assumes the content is what its
metadata claims.

#### Actors

Package Manager; Policy Engine (trust policy); users deciding on `unverified` prompts;
Audit Log.

#### Preconditions

Acquired archive in staging; plan expectations (checksum, signature reference) available;
trust material configured where signatures are to be evaluated.

#### Main flow

1. Hash the archive; compare to the plan.
2. Evaluate the signature per ADR-081; compute `signature_state`; apply policy.
3. Extract with containment; validate metadata and payload for the kind.
4. Record all outcomes on the row; proceed to `staged`.

#### Alternative flows

- `unverified` + interactive: an Approval is raised showing name, version, publisher,
  source, and checksum; denial fails the installation with the decision recorded.
- `unverified` + non-interactive: denied (PRD-009); E-PLUG-011 with the policy cause.

#### Edge cases

- `signature_required = true` on the source makes a missing signature `invalid`, not
  `unverified` (ADR-081 downgrade defense).
- An archive entry with an absolute path, `..` traversal, or a symlink fails content
  verification naming the entry (E-PLUG-011); nothing extracts outside staging.
- Post-install `Verify` on a package whose files were modified reports each divergent path
  with expected/observed digests; the delivered extensions' own integrity rules
  (INV-SKL-02 for skills) additionally disable affected units at load time.
- Trust material referencing an unreachable transparency-log endpoint offline: the signature
  evaluates against local material; unevaluable online-only steps are recorded as not
  evaluated, never silently passed (ADR-081 rule 5).

#### Inputs

Archives, plans, signature bundles, trust material, policy documents, `files_manifest`
(post-install verify).

#### Outputs

Verification verdicts on the Package row (`checksum` confirmed, `signature_state`); audit
records; tamper reports.

#### States

`verifying` in the chapter 10 machine; post-install `Verify` reads `installed` without
transitions.

#### Errors

E-PLUG-011 (all verification failure classes: checksum, signature-invalid, policy denial,
content violation); E-SEC family for the Approval path.

#### Constraints

Fixed order (checksum before signature before content); no partial extraction outside
staging; verification results are immutable facts on the row for that attempt.

#### Security

This requirement is the supply-chain control (SM-16 posture for extensions): fail-closed at
every step, decisions recorded, downgrade defenses per source. Verification failures never
leak archive content into diagnostics beyond entry names and digests.

#### Observability

`package.verification.failed` on any failure with the failure class;
verification outcomes in the installation audit chain; post-install tamper findings emit
the same event with the `post_install` marker.

#### Performance

Verification cost is dominated by hashing (NFR-PLUG-003 budgets); signature evaluation adds
no network round trips when material is local.

#### Compatibility

Signature technology pinned per ADR-081 (cosign v3 line); index/plan expected-checksum
format is `sha256:<hex>`; future digest algorithms enter via the change procedure.

#### Acceptance criteria

- Given a package whose archive hash equals the plan checksum and whose signature validates
  against configured trust material, when verified, then `signature_state = verified` and
  installation proceeds without prompts.
- Given a tampered archive (any byte changed), when verified, then E-PLUG-011 reports the
  checksum mismatch, the row is `failed`, and staging is cleaned.
- Given a signature that fails validation, when verified, then installation blocks
  regardless of any configuration and the failure is audit-logged (INV-PKG-02).
- Permission case: given `unverified` state in a non-interactive run, when policy applies,
  then installation is denied with the recorded decision and E-PLUG-011 names the policy
  cause.
- Negative case: given an archive containing `../escape`, when content verification runs,
  then extraction stops, the entry is named, and no file exists outside staging.
- Observability case: post-install `Verify` on a modified file reports the path with both
  digests and emits `package.verification.failed` with the `post_install` marker.

#### Verification method

Verification matrix tests (checksum × signature × policy × source flags); tamper fixtures
(bit flips, entry injection, symlink smuggling); offline verification tests with local
trust material; audit-chain assertions.

#### Traceability

INV-PKG-01..03; ADR-081, ADR-013; FR-PLUG-005; RISK-PLUG-003; Volume 9 trust policy and
Approval semantics.

## Supply-chain rules

1. **One pipeline.** Every extension package — regardless of source kind — passes the same
   resolve → acquire → verify → stage → install → register pipeline; there is no trusted
   side door. Local `path` sources skip network acquisition, never verification.
2. **Source policy is security policy.** Source priority, enablement, and
   `signature_required` are the installation's dependency-confusion and downgrade defenses
   (ADR-081, ADR-083); they live in versioned configuration and are auditable.
3. **No implicit updates, ever.** Extension updates occur only on explicit command; there is
   no background fetching of indexes or artifacts.
4. **Provenance survives removal.** Tombstoned Package and Extension rows keep name,
   version, source, checksum, and signature state, so historical runs remain attributable
   (SM-12/SM-13).
5. **Product and extensions stay separate.** The Updater (Volume 14) never installs
   extension packages; the Package Manager never touches the product binary. The two
   pipelines share verification patterns (ADR-013/ADR-081) but no state.

### RISK-PLUG-003 — Extension supply-chain compromise

- Category: Security / supply chain
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: fixed verification order with fail-closed checksum and signature gates
  (FR-PLUG-007, INV-PKG-01/02); per-source `signature_required` and priority policy
  (ADR-081, ADR-083); path-contained extraction; deterministic resolution with recorded
  plans; no implicit updates; post-install `Verify` tamper detection; trust classification
  gating activation/start (chapters 07, 08)
- Detection: E-PLUG-011 rates; divergent-checksum findings across sources; post-install
  `Verify` findings; audit-chain review; Volume 9 threat-model monitoring
- Owner: Package Manager (Volume 6) with Volume 9 threat model
- Status: Open

The attack shapes are known: a compromised source rewrites index entries; a hostile
publisher ships a poisoned update; a name-squatting package rides priority
misconfiguration; a post-install attacker edits installed files. The pipeline bounds each:
checksums bind content to plans, signatures bind content to publishers, priority and
per-source policy bound where names may come from, and re-verification makes local
tampering detectable. Residual risk concentrates in the `unverified` path a user explicitly
approves — addressed by provenance-forward consent UX and Volume 9's threat entries.

## Marketplace (Future)

A curated marketplace — hosted browsing, publisher accounts, moderation, ratings — is
classified **Future** (ADR-080). Its committed compatibility floor: the marketplace presents
as a `registry`-kind source using the registry index schema and the FR-PLUG-007 pipeline
unchanged. No requirement in this corpus depends on its existence; scope confirmation
belongs to the v2 process (Volume 1 phase definitions).

## Performance

### NFR-PLUG-003 — Extension package operation latency

- Category: Performance
- Priority: P2
- Phase: Beta
- Metric: (a) wall-clock install of a 1 MB skill package from a local `archive` source,
  request to `installed`; (b) wall-clock install of a 20 MB plugin package from a local
  `archive` source; (c) resolution-plan production for a request over enabled sources whose
  cached indexes total 1,000 package entries
- Target: (a) ≤ 2 s p95; (b) ≤ 5 s p95; (c) ≤ 500 ms p95
- Minimum threshold: (a) ≤ 4 s p95; (b) ≤ 10 s p95; (c) ≤ 1000 ms p95
- Measurement method: benchmark harness over fixture packages and indexes, 50 iterations,
  p95, per release; network transfer excluded by construction (local sources)
- Test environment: Volume 1 reference hardware, both reference machines per Volume 12
  formalization
- Measurement frequency: per release; regression-tracked in the Volume 12 benchmark suite
- Owner: Package Manager (Volume 6)
- Dependencies: FR-PLUG-005, FR-PLUG-006, FR-PLUG-007
- Risks: RISK-PLUG-003
- Acceptance criteria: Benchmark report shows all three percentiles within target on both
  reference machines; a minimum-threshold breach blocks release per Volume 12 gating.

## Events minted (package.*)

Envelope, ordering, delivery, persistence, retention, privacy, and failure behavior per
Volume 10 (FR-OBS-001). All payloads carry the package ULID, name, version, kind, and
correlation IDs; none carry archive content.

| Event | Version | Producer | Consumers | Payload summary |
|---|---|---|---|---|
| `package.resolution.completed` | 1 | Package Manager | CLI/TUI, Observability | plan hash, package count, sources consulted, cache ages, warning count |
| `package.installation.started` | 1 | Package Manager | CLI/TUI, Observability, Audit Log | plan hash, source kind, scope |
| `package.installation.completed` | 1 | Package Manager | CLI/TUI, Observability, Audit Log | duration, files count, signature state, registered extension kinds |
| `package.installation.failed` | 1 | Package Manager | CLI/TUI, Observability, Audit Log | failing state, error code, reason class |
| `package.verification.failed` | 1 | Package Manager | CLI/TUI, Observability, Audit Log | failure class (checksum/signature/policy/content), post_install marker |
| `package.rollback.completed` | 1 | Package Manager | CLI/TUI, Observability, Audit Log | restored version, failing state that triggered it |
| `package.removal.completed` | 1 | Package Manager | CLI/TUI, Observability, Audit Log | cascade flag, deregistered extension kinds |

## Error codes (E-PLUG-008 – E-PLUG-013)

### E-PLUG-008 — Package source unavailable

- Category: Connectivity
- Severity: Error
- User message: "Package source '<name>' could not be read: <cause summary>."
- Technical message: source kind and location class, transport error or index schema-version incompatibility, cache age if a cache was used
- Cause: network failure, missing path, git fetch failure, malformed or incompatibly versioned index
- Safe-to-log data: source name, kind, error class, index schema versions where applicable
- Recoverability: recoverable (fix source, restore connectivity, or use another source)
- Retry policy: none automatic; discovery/resolution proceed over remaining sources where possible
- Recommended action: check the source configuration and reachability; refresh or re-add the index
- Exit-code mapping: 1
- HTTP mapping: transport status recorded as cause class when an HTTP source reported one
- Telemetry event: `package.installation.failed` (when fatal to an operation); otherwise logged finding
- Security implications: unreadable sources fail visibly — resolution never silently narrows its candidate set without a recorded finding

### E-PLUG-009 — Package resolution failed

- Category: Dependency
- Severity: Error
- User message: "Cannot resolve '<package>': <cause class>."
- Technical message: not-found across enabled sources, empty constraint intersection (every colliding constraint with its declaring package), closure bound violation (offending chain), malformed metadata findings, or contract-version incompatibility
- Cause: missing package/version, conflicting semver ranges, depth/node bound exceeded, invalid metadata, unsupported `contract_version`
- Safe-to-log data: package names and versions, constraint list, bound values, source names
- Recoverability: recoverable by pinning versions, adjusting sources, or updating declarations
- Retry policy: none (deterministic against the same inputs)
- Recommended action: inspect the named constraints; pin an explicit version or update the conflicting dependents
- Exit-code mapping: 1
- HTTP mapping: not applicable
- Telemetry event: `package.installation.failed`
- Security implications: fail-closed resolution — no partial or heuristic plans

### E-PLUG-010 — Package download failed

- Category: Connectivity
- Severity: Error
- User message: "Downloading '<package>' from '<source>' failed."
- Technical message: transfer error class, bytes transferred versus expected size, budget (`timeout_ms`) if expired
- Cause: network interruption, source-side error, size mismatch, acquisition budget expiry
- Safe-to-log data: package name/version, source name, byte counts, elapsed
- Recoverability: recoverable; re-running the installation re-acquires
- Retry policy: one automatic re-attempt per archive within the same operation; further retries are user-driven
- Recommended action: retry; check source health; pre-fetch archives for offline installs
- Exit-code mapping: 1
- HTTP mapping: transport status recorded as cause class where reported
- Telemetry event: `package.installation.failed`
- Security implications: partial downloads never enter verification as complete artifacts; staging is cleaned on failure

### E-PLUG-011 — Package verification failed

- Category: Integrity
- Severity: Critical
- User message: "Package '<name>' failed verification and was not installed: <failure class>."
- Technical message: failure class (checksum mismatch with both digests, invalid signature with trust-material reference, policy denial for unverified content, content violation naming the archive entry, metadata disagreement), post_install marker for re-verification findings
- Cause: tampered or corrupted archive, failed signature validation, trust-policy denial, path-containment violation, index/archive metadata disagreement
- Safe-to-log data: package name/version, failure class, digests, entry names; never archive content
- Recoverability: not recoverable for the artifact (the content is untrusted); recoverable by obtaining a clean artifact
- Retry policy: none — deterministic against the same artifact
- Recommended action: re-acquire from a trusted source; report the source if digests diverge; review trust material for signature failures
- Exit-code mapping: 9
- HTTP mapping: not applicable
- Telemetry event: `package.verification.failed`
- Security implications: the primary supply-chain control (INV-PKG-01/02); failures are audit-logged with full provenance and nothing partially activates

### E-PLUG-012 — Package installation conflict

- Category: Conflict
- Severity: Error
- User message: "Cannot <install|remove> '<name>': <conflict summary>."
- Technical message: coexistence violation (INV-PKG-04), file-placement collision with an existing `files_manifest` entry of another package, removal blocked by named installed dependents, or removal blocked by an unterminated plugin process (INV-PLG-04)
- Cause: conflicting installed state, dependents, or live processes
- Safe-to-log data: package identities, conflicting paths or dependent names
- Recoverability: recoverable by resolving the conflict (remove/stop the other party, or force cascade where offered)
- Retry policy: none (deterministic until state changes)
- Recommended action: follow the named conflict; use explicit cascade removal where dependents are acceptable to remove
- Exit-code mapping: 1
- HTTP mapping: not applicable
- Telemetry event: `package.installation.failed`
- Security implications: file-collision refusal prevents one package from silently overwriting another's content

### E-PLUG-013 — Package operation interrupted or cancelled

- Category: Cancellation
- Severity: Warning
- User message: "The package operation on '<name>' was <cancelled|interrupted>; no partial installation remains."
- Technical message: state at interruption, cleanup actions taken (staging removal, rollback), recovery reconciliation marker when produced at startup
- Cause: user cancellation, context cancellation, process crash or shutdown mid-operation
- Safe-to-log data: package name/version, state at interruption, cleanup summary
- Recoverability: recoverable — re-run the operation; the recorded plan makes the retry identical
- Retry policy: user-driven re-run; no automatic resumption
- Recommended action: re-run the operation; run `Verify` on the prior version after an upgrade interruption
- Exit-code mapping: 8
- HTTP mapping: not applicable
- Telemetry event: `package.installation.failed`
- Security implications: interruption never leaves partially active content — cleanup and rollback guarantees are the chapter 10 machine's recovery rules
