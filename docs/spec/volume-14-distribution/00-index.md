# Volume 14 — Distribution, Installation, and Updates

**Status:** Authored (draft) · **Owner:** Updater / Package Manager / release engineering (Volume 14)

Volume 14 specifies how Andromeda reaches and leaves machines and how installations move
between versions: the release pipeline and distribution channels (keystone FR-REL-001), the
artifact integrity chain (checksums, cosign signatures, SBOM, provenance), installation
channels including air-gapped sites, the Updater behind the frozen UpdaterPort (checks,
consent-gated apply, automation, offline rollback), install/uninstall/data-removal
procedures, the SemVer/support/backport regime, and the full Update and Release state
machines. Per Volume 0 chapter 03, this volume mints all `REL` identifiers, the `[update]`
configuration key content, the `update.*`/`release.*` event names, and ADRs 190–204 (used:
190–193).

Foundations assumed: Volume 0 (conventions), Volume 1 (phases, MVP minimum items 22/23/27,
signing viability, platform tiers), Volume 2 (Release entity and INV-REL invariants; frozen
Update and Release states), Volume 3 (UpdaterPort signature; PAL Installer/Updater
surfaces; deployment shapes and state footprint), ADR-013 (release tooling), ADR-015
(SemVer + Conventional Commits), ADR-029 (forward-only migrations).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Distribution Channels](01-distribution-channels.md) | Pipeline overview, artifact inventory (ADR-190), FR-REL-001 (release pipeline, keystone), FR-REL-002 (checksums/signatures/SBOM/provenance), FR-REL-003 (Homebrew tap, shell installer, Linux packages), FR-REL-004 (air-gapped install, release mirrors), RISK-REL-002 |
| [02 — Updater and Rollback](02-updater-and-rollback.md) | UpdaterPort semantics, `[update]` keys, FR-REL-005 (check/channels/notification), FR-REL-006 (download/verify/apply), FR-REL-007 (automation, ADR-191), FR-REL-008 (rollback, ADR-192), NFR-REL-001/002, RISK-REL-001/003, E-REL catalog, `update.*`/`release.*` events |
| [03 — Installation, Uninstallation, and Data](03-install-uninstall-data.md) | Ownership model, data classes, FR-REL-009 (layout and ownership detection), FR-REL-010 (uninstall preserving data), FR-REL-011 (explicit purge) |
| [04 — Versioning, Support Windows, and Backports](04-versioning-support-backports.md) | Versioned surfaces, FR-REL-012 (SemVer and breaking changes), FR-REL-013 (deprecation policy), FR-REL-014 (support windows/backports/release branches, ADR-193), FR-REL-015 (changelog and release notes), NFR-REL-003 (SM-20), RISK-REL-004 |
| [05 — State Machines: Update and Release](05-state-machines.md) | Full machines over the frozen states (T1–T14, R1–R7, twelve elements each), FR-REL-016 (conformance and update history) |
| [99 — Volume Register](99-volume-register.md) | Everything this volume minted: requirements, ADRs, error codes, events, config keys, glossary additions, assumptions, open questions |

## Reading guide

1. Chapter 01 fixes what a release *is*; every other chapter consumes its artifact and
   verification vocabulary.
2. Chapter 02 is the behavioral contract behind UpdaterPort; Volume 8's `andromeda update`
   commands are its manual surface.
3. Chapters 03 and 04 govern the edges: machine on/off-boarding and the promises attached
   to version numbers over time.
4. Chapter 05 is normative for every state the words "update" and "release" carry anywhere
   in the corpus.
