# 99 — Volume 14 Register

Machine-parseable register of everything Volume 14 minted, per Volume 0 chapters 02 and 03.
Merged into the Volume 0 registers at consolidation.

## Requirements index

### Functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| FR-REL-001 | Release pipeline and distribution channels | MVP | Release-audit CI job (inventory, grammar, channel mapping); Volume 13 release qualification |
| FR-REL-002 | Integrity metadata: checksums, signatures, SBOM, provenance | MVP | Tamper-fixture suite; release-audit re-verification; SM-18 harness inclusion |
| FR-REL-003 | Installation channels: Homebrew tap, shell installer, Linux packages | MVP | Per-release installation matrix on Tier 1 platforms; installer tamper fixtures |
| FR-REL-004 | Air-gapped installation and offline update sources | MVP | Offline suite with mirror fixtures under OS-level network disablement |
| FR-REL-005 | Update check, channel subscription, and notification | MVP | UpdaterPort contract tests with metadata fixtures; offline suite; egress capture |
| FR-REL-006 | Download, verification, and consent-gated apply | MVP | SM-18 update suite; fault injection per state; lock contention tests |
| FR-REL-007 | Update automation policy | Beta | Policy-matrix integration tests; idle-window scheduling tests; suspension fixtures |
| FR-REL-008 | Rollback of the installed version | MVP | SM-19 offline rollback harness; corruption and schema-boundary fixtures |
| FR-REL-009 | Installation layout and ownership detection | MVP | Per-owner fixtures; lazy-initialization filesystem snapshots |
| FR-REL-010 | Uninstallation with data preservation by default | MVP | Uninstall matrix with before/after snapshots; broken-install fixtures |
| FR-REL-011 | Explicit data removal | MVP | Purge fixtures with filesystem/credential snapshots; override-layout fixtures |
| FR-REL-012 | Semantic versioning of the product and public contracts | Core | Release-audit bump-class vs contract-diff consistency; upgrade-test matrix |
| FR-REL-013 | Deprecation policy | Beta | Deprecation-ledger reconciliation in release audit; warning emission tests |
| FR-REL-014 | Support windows, release branches, and backports | MVP | Upgrade-path matrix; branch-content audits; support-status computation tests |
| FR-REL-015 | Changelog and release notes | MVP | Release-audit note checks; changelog divergence test |
| FR-REL-016 | Machine conformance and update history | MVP | State-machine property suite; crash-injection per state; event/history reconciliation |

### Non-functional requirements

| ID | Title | Phase | Verification method |
|---|---|---|---|
| NFR-REL-001 | Update time (SM-18) | v1 | Automated N−1 → N update test per release, per-state instrumentation, p95 |
| NFR-REL-002 | Rollback time (SM-19) | v1 | Automated offline rollback test per release with egress capture, p95 |
| NFR-REL-003 | Public-contract stability (SM-20) | v1 | Contract-diff tooling per release; audit against the deprecation ledger |

### Risks

| ID | Title | Severity | Status |
|---|---|---|---|
| RISK-REL-001 | Failed or interrupted update leaves an unusable installation | High | Open |
| RISK-REL-002 | External signing and notarization dependencies | Medium | Open — V14-OQ-1/V14-OQ-3 pending |
| RISK-REL-003 | Binary–database version skew after rollback or workspace sync | Medium | Open |
| RISK-REL-004 | Support and compatibility obligations exceeding maintainer capacity | Medium | Open |

## ADRs minted

Volume 14 block allocation is 190–204 (Volume 0 chapter 03); this volume used 190–193;
194–204 remain permanent gaps unless minted by later amendment.

| ADR | Title | Status | Decision (one line) |
|---|---|---|---|
| [ADR-190](../annexes/adr/ADR-190.md) | Distribution artifact set, naming scheme, and package formats | Accepted | Fixed `andromeda_<v>_<os>_<arch>` grammar with uname-style tokens; deb/rpm/apk files attached (hosted repos Future); checksums-file signing; macOS Intel/universal PENDING VALIDATION |
| [ADR-191](../annexes/adr/ADR-191.md) | Update channels, subscription floors, and automation consent policy | Accepted | Closed channel enum stable/rc/beta/nightly with SemVer pre-release semantics and a maturity-floor offer rule; auto check/download/apply ladder with consent boundary at apply; majors never auto-applied |
| [ADR-192](../annexes/adr/ADR-192.md) | Retained-version rollback and binary–database pairing | Accepted | One retained version by default; offline binary-only rollback with an explicit schema-pairing dialogue; automatic restore on failed apply shares the mechanism |
| [ADR-193](../annexes/adr/ADR-193.md) | Support windows, release branches, and backport policy | Accepted | Pre-v1 latest-release-only; from v1 latest minor of current major (full fixes) plus previous major's last minor (security/integrity) for 12 months; cherry-pick-only `release/vX.Y` branches; closed backport list |

## Error codes minted

| Code | Name | Exit code |
|---|---|---|
| E-REL-001 | Update check failed | 1 |
| E-REL-002 | Release metadata invalid | 1 |
| E-REL-003 | Artifact download failed | 1 |
| E-REL-004 | Artifact verification failed | 9 |
| E-REL-005 | Update apply failed | 1 |
| E-REL-006 | Rollback failed or unavailable | 1 (9 when a restore attempt left verification-failed state) |
| E-REL-007 | Update already in progress | 1 |
| E-REL-008 | Externally managed installation | 1 |
| E-REL-009 | Unsupported upgrade path | 1 |
| E-REL-010 | Insufficient disk space for update | 1 |
| E-REL-011 | Release yanked | 1 |
| E-REL-012 | Update step timed out or cancelled | 8 |

## Events minted

Per the Volume 0 event grammar; envelope, ordering, delivery, persistence, retention, and
redaction semantics per Volume 10. Payloads are content-free (versions, channels, digests,
counts, ULIDs).

| Event | Emitted by | Meaning |
|---|---|---|
| `update.check.completed` | Updater | Check finished with `up_to_date` or `update_available` (FR-REL-005) |
| `update.check.failed` | Updater | Check failed (E-REL-001/002) |
| `update.artifact.downloaded` | Updater | Artifact finished staging (FR-REL-006) |
| `update.artifact.verified` | Updater | Verification passed for the staged set (FR-REL-002) |
| `update.verification.failed` | Updater | Verification failed (E-REL-004) |
| `update.version.applied` | Updater | New version active (`applied`) |
| `update.version.rolled_back` | Updater | Restore completed, automatic or manual (`rolled_back`) |
| `update.process.failed` | Updater | Update process terminated in `failed` |
| `update.state.changed` | Updater | Any Update machine transition (chapter 05) |
| `release.metadata.refreshed` | Updater | Local Release rows refreshed from the source |
| `release.yank.detected` | Updater | Installed or targeted version learned to be yanked |

## Config keys minted

Key content owned by this volume; schema, precedence, env-var mapping, and validation by
Volume 10 (single-home matrix).

| Table | Keys |
|---|---|
| `[update]` | `channel`, `source`, `auto_check`, `check_interval_hours`, `auto_download`, `auto_apply`, `notify`, `keep_versions`, `signature_policy` |
| `[update.timeouts]` | `check_seconds`, `download_seconds`, `apply_seconds` |

## Glossary additions

| Term | One-line meaning |
|---|---|
| Update channel | One of the closed set `stable`, `rc`, `beta`, `nightly`: a maturity filter over the single SemVer-ordered release line (ADR-191). |
| Offer rule | The ADR-191 selection: the highest non-yanked release whose channel maturity ≥ the subscriber's channel and whose upgrade path admits the installed version. |
| Retained version | The previously installed binary (plus verification metadata and supported schema versions) kept by the Updater to enable offline rollback (ADR-192). |
| Update lock | The machine-wide PAL file lock serializing mutating Updater operations (E-REL-007). |
| Release mirror | A filesystem or HTTPS root with an `index.json` and artifact files, serving the full update flow to air-gapped sites via `[update].source` (FR-REL-004). |
| Update history | The append-and-update global-database record of every Update machine instance (FR-REL-016). |
| Installation owner | The detected manager of the installed binary — `self`, `homebrew`, or `package` — deciding whether self-update proceeds or defers (FR-REL-009). |
| Support window | The published period during which a release line receives defined fix classes (ADR-193). |
| Backport | A cherry-picked, qualification-listed fix released as a patch on a supported release branch (FR-REL-014). |

## Assumptions

Local list per Volume 0, chapter 05 (global numbers minted at consolidation).

| Kind | Statement | Validation path | If false |
|---|---|---|---|
| Technical assumption | GitHub Releases' documented public API remains usable for unauthenticated release-metadata checks at the default daily cadence within its published rate limits | Scheduled CI job exercising the check path against the live origin | Publish a static metadata index on project infrastructure and point the default `source` at it — the mirror mechanism (FR-REL-004) already covers the shape |
| Technical assumption | Same-filesystem rename-based binary replacement behaves atomically on Tier 1 platforms as the PAL Updater surface contract requires | PAL golden tests per platform; crash injection during the swap (FR-REL-016 suites) | Adopt the PAL's locked-file-aware strategy early (planned for Windows) and re-verify SM-18/SM-19 budgets |
| Product hypothesis | One retained version (`keep_versions = 1`) covers the dominant rollback scenario ("the update I just applied is bad") | Rollback usage and failure telemetry under consent; support channels | Raise the default and/or add multi-step rollback (additive per ADR-192 reversal plan) |
| Product hypothesis | `auto_check` on / `auto_download` and `auto_apply` off is the right default automation posture | Beta feedback (Volume 15); update-lag distribution under consent-based metrics | Change configuration defaults with release-note prominence; no contract impact (ADR-191 reversal plan) |

## Open questions

Entries follow Volume 0, chapter 08; none blocks authoring. Every PENDING VALIDATION
occurrence in this volume maps to a row here (or to the originating ADR's register entry,
noted).

| Local ID | Question / pending item | Raised in | Blocking? | Resolution path | Status |
|---|---|---|---|---|---|
| V14-OQ-1 | macOS Developer ID signing and notarization — PENDING VALIDATION per ADR-013 (Apple Developer account decision); until resolved, releases ship cosign verification plus Gatekeeper guidance and notes state signing status | ADR-013; chapter 01 | No — checksums and provenance are unconditional; signing enablement is configuration (Volume 1 signing viability) | Organizational decision; amend ADR-013 and the chapter 01 inventory | Open |
| V14-OQ-2 | macOS Intel (`darwin_x86_64`) Tier 2 builds and the `darwin_universal` artifact — PENDING VALIDATION, coupled to the Volume 1 platform-scope question on Intel build/test capacity | ADR-190; chapter 01 inventory | No — Tier 1 matrix is complete without them | Resolve with the Volume 1 Tier 2 decision; activate or delete the inventory rows | Open |
| V14-OQ-3 | In-binary signature verification implementation path (official Sigstore verification tooling embedded in the Updater vs external tooling) — PENDING VALIDATION; checksum verification is self-contained and unconditional either way | Chapter 01 (FR-REL-004 security); chapter 02 verification | No — `Verify`'s contract and `signature_policy` semantics are fixed independent of mechanism | Validation spike against the pinned cosign major (ADR-013); record the mechanism in a follow-up ADR from the 194–204 block | Open |

## Cross-volume references

Load-bearing references made by this volume, by name (requirement-level cross-links are
upgraded at consolidation):

| Referenced | Where used |
|---|---|
| UpdaterPort and PackagePort (Volume 3 chapter 02, frozen); PAL Installer/Updater, File Locking, Config Directories surfaces (Volume 3 chapter 07); deployment shapes and state footprint (Volume 3 chapter 09) | Chapters 01–03, 05 |
| Release entity, INV-REL-01..04, frozen Update and Release states (Volume 2 chapters 06/09) | Chapters 01, 02, 05 |
| ADR-004, ADR-013, ADR-014, ADR-015, ADR-016, ADR-022, ADR-023, ADR-027, ADR-028, ADR-029 (foundation decisions) | Throughout |
| Volume 8: `andromeda update` / `update check` / `update rollback` command surface, confirmation and non-interactive conventions, post-command update notice, `doctor` | Chapters 02, 03 |
| Volume 9: `network` and `system_modification` permissions, `always_allow_policy` decision, Audit Records, disclosure process, credential storage model | Chapters 02–04 |
| Volume 10: configuration schema/precedence/validation, event envelope and delivery, storage write discipline, config deprecation mechanics | Chapters 02, 04, 05 |
| Volume 11: CI pipelines, tag/branch conventions, release workflow protections, backport automation, `CHANGELOG.md` conventions | Chapters 01, 04 |
| Volume 12: reference machines, startup and disk budgets, SM-18/SM-19 harness environments | NFR test environments |
| Volume 13: release qualification, offline suite, installation/upgrade matrices, crash-injection harnesses | Verification methods throughout |
| Volume 1: phases and MVP minimum items 22/23/27, platform tiers, signing viability, offline guarantees; SM-18/SM-19/SM-20 formalized here | Chapters 01–04 |
