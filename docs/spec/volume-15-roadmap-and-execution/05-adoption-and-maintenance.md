# 05 — Adoption and Maintenance

This chapter plans the product's life with its users: how each Volume 1 persona is
onboarded and at which phase Andromeda genuinely serves them, the standing maintenance
duties that keep a shipped product healthy, and the sustainability instruments — feedback
channels and project health metrics — that other volumes reference by name ("Volume 15
feedback channels", "Volume 15 project health metrics", "Volume 15 process audits").

## Adoption plan

### Personas mapped to onboarding paths

The six personas are Volume 1 chapter 02's; the onboarding paths below are commitments on
documentation, defaults, and sequencing — not new product behavior.

| Persona (Vol 1 ch 02) | Serving phase | Onboarding path | First-session success criterion |
|---|---|---|---|
| Elena — individual open-source developer | MVP | Install via Homebrew tap or shell installer (FR-REL-003) → `andromeda init` in a repository → add one cloud key (FR-AUTH-002) and the Ollama adapter → first interactive session with plan review and scoped permissions. Onboarding doc: the quickstart, cost-visibility guide (pricing tables per ADR-058), and permission-prompt walkthrough | A reviewed, committed change in her repository in the first session, with per-run cost basis visible |
| Amara — local-model privacy advocate | MVP | Offline-first quickstart: install from a downloaded artifact with checksum verification → local provider only (FR-PROV-083, FR-PROV-084) → verify the zero-egress posture with the documented offline checks (NFR-OBS-006, SM-05). Onboarding doc: the local-first guide, capability-honesty explanation (no silent simulation per Volume 5 chapter 02) | A full index-plan-edit-test-commit loop with zero network egress observed |
| Jonas — AI automation builder | MVP (interactive parity), Beta (headless) | CI quickstart: non-interactive mode (`--json`, `--no-input`, policy-only permissions per FR-SEC-105) → exit-code table and structured-output schemas (FR-CLI-006) → session resume in pipelines (FR-AGT-003). Beta adds the headless instance over IPC (FR-ARCH-008, ADR-032). Onboarding doc: the automation guide with a reference GitHub Actions job | A scripted run in CI producing parseable output and honest exit codes, resumable after interruption |
| Tomás — extension and tool author | Beta | SDK path: SDK tool template → conformance kits and test provider (FR-TEST-004, FR-TEST-006) → tool in a session within the SM-02 budget (NFR-SDK-001) → plugin over ARP (FR-PLUG-001) and skill packaging (FR-SKILL-001) → distribution through declared sources with signing (ADR-080, ADR-081). Onboarding doc: the extension author guide; the SM-02/SM-03 tutorials that CI itself walks | A working, tested, installable extension against a fake provider — no live credentials required |
| Marcus — platform engineer | Beta | Paved-road path: project-level `andromeda.toml` policy in the repository (FR-CFG-001 precedence), organization defaults via permission rules (FR-SEC-100 policy rules), internal tools/skills/plugins distributed through internal package sources (FR-PLUG-006), events exported to the team's stack through the consent-gated OTLP path (FR-OBS-013). Onboarding doc: the platform-team guide (policy, distribution, observability) | A second developer reproduces the paved road by cloning a repository — no per-user setup beyond credentials |
| Priya — security-conscious enterprise engineer | v1 | Review-first path: the security review dossier — permission model (FR-SEC-100), sandbox honesty (ADR-021 containment levels), audit chain (FR-SEC-111, SM-13), offline guarantees, official-mechanisms-only policy (FR-AUTH-001), SBOM/provenance (FR-REL-002) — then air-gapped installation (FR-REL-004) and enterprise proxies/trust anchors (FR-AUTH-006). Onboarding doc: the enterprise adoption guide, mapped to the evidence her auditors ask for | Her security office can answer "what leaves the machine, what ran, who approved it" from queries, not promises |

Sequencing consequence: MVP onboarding investment targets Elena, Amara, and Jonas (the
UC-01/UC-09/UC-11 jobs); Beta targets Tomás and Marcus (extension surfaces and paved
roads); the v1 dossier targets Priya (the guarantees only become claims when their gates
bind at v1 — NFR-SEC-001, NFR-PERF series, NFR-REL series). Publishing an enterprise
dossier before its gates bind would violate the honesty principle, so it is deliberately
last.

### Feedback channels

The channels other volumes cite for hypothesis validation (Volume 3 headless-shape
hypothesis, Volume 5 adapter-coverage hypothesis, Volume 7 ingestion-default hypothesis,
Volume 14 automation-default hypothesis, ADR-049 stage-granularity hypothesis):

1. **Tracker signal** — issue forms with mandatory requirement references; triage tags
   the register hypothesis an issue bears on, so phase gates can query "evidence for/
   against hypothesis X".
2. **Discussions** — qualitative feedback and design questions; summarized into gate
   evidence by the release manager per phase gate.
3. **Consent-based telemetry** — only under the Volume 10 consent regime (ADR-140,
   disabled by default); the collected-data catalog bounds what adoption analysis may
   use; absence of telemetry is never treated as absence of users (opt-in volume is
   accepted as low and self-selected per the Volume 10 register).
4. **Timed exercises** — the SM-01/SM-02/SM-03 phase-gate exercises double as usability
   evidence for the contributor-facing surfaces.

Each register hypothesis names its validation path; this chapter's duty is that the
channels exist, are queryable, and are reviewed at every phase gate.

## Maintenance plan

Standing duties once the MVP ships, each with its owner-of-record and cadence:

| Duty | Mechanism | Cadence | Source |
|---|---|---|---|
| Dependency updates | Grouped dependency PRs (gomod + actions) with the license-allowlist check (ADR-002) and full T0/T1 gates; SemVer-major bumps require a maintainer-reviewed changelog read | Weekly | Volume 11 chapter 03 (`dependabot.yml`); RISK-SEC-013 |
| Security patching | Advisory monitoring (CodeQL, dependency audit, secret scanning per FR-GH-010) feeding rule R1 scheduling; fixes backported per ADR-193's closed classes; NFR-SEC-001 gates publication | Continuous; release-gated | Volume 9, Volume 14 |
| Benchmark regression watch | Nightly benchmark lane with rolling baselines and the two-band regression gate (warn/fail per ADR-161); calibration failures investigated before results are trusted (RISK-PERF-002); trends reviewed at phase gates | Nightly; gate per release | Volume 12 (FR-PERF-005) |
| Provider drift watch | Scheduled live smoke against seed providers (recorded fixtures remain the gating path); E-PROV-008 rates monitored as drift telemetry; capability re-verification per configured cadence | Weekly scheduled lane | Volume 5 (RISK-PROV-001, RISK-PROV-080) |
| Git version matrix | Equivalence suite across the pinned git version matrix (NFR-GIT-001); floor review when a new git minor ships | Per release | Volume 11 (ADR-025, RISK-GIT-001) |
| MCP revision tracking | Interop scorecard trends (SM-15 b); SDK pin advanced per ADR-010 review conditions | Weekly interop job | Volume 6 (RISK-MCP-002) |
| Flake and determinism hygiene | Determinism lane; quarantine registry with the 14-day time box (ADR-177); dwell-time breaches escalate to P1 | Nightly; weekly triage | Volume 13 (NFR-TEST-005) |
| Traceability and process audit | Nightly chain audit (FR-GH-001); orphan and override reports reviewed; the Volume 9 "process audits" — verifying the security-inbox tracking and escalation roster actually function — run quarterly | Nightly; quarterly | Volume 11, Volume 9 (NFR-SEC-003) |
| Documentation currency | Docs pipeline gates (NFR-CLI-002 help-coverage walk; docs generation per FR-GH-011); user-facing changes land with docs per the Definition of Done | Per merge | Volume 8, Volume 11 |
| Data-retention and support-window duties | Deprecation ledger reconciliation, support-status computation, release-branch freeze/deletion at end of support | Per release | Volume 14 (FR-REL-013, FR-REL-014) |

Maintenance capacity is a planning input, not an afterthought: RISK-REL-004 (support
obligations exceeding maintainer capacity) is reviewed at every phase gate against the
backport queue depth and patch-release latency, and the support window (ADR-193) is the
lever deliberately kept narrow until capacity exists.

## Sustainability

### Project health metrics

The consolidated metric set reviewed at each phase gate (and cited by Volumes 9/11/14 as
"Volume 15" inputs). Each metric already has an owner and a mechanism; this table is the
review checklist, not a new measurement system:

| Metric | Definition and source | Watch condition |
|---|---|---|
| PR feedback latency | NFR-GH-002 check-run rollups (p85 targets) | Rising latency deters contributors (RISK-PRD-009) |
| Traceability completeness | NFR-GH-001 nightly audit orphan rate | Erosion signals process bypass (RISK-GH-002) |
| Coverage and mutation trend | NFR-TEST-003 / NFR-TEST-004 reports | Coverage rising while mutation score falls signals gaming (RISK-TEST-003) |
| Flake rate and quarantine dwell | NFR-TEST-005 registry | Gate credibility (RISK-TEST-002) |
| Disclosure first-response | NFR-SEC-003 inbox tracking | Roster health; bus factor (RISK-PRD-003) |
| Backport queue depth and patch latency | Volume 14 release records | Capacity versus support windows (RISK-REL-004) |
| Contributor funnel | Unique contributors per release; first-PR-to-second-PR conversion; committer/maintainer counts and areas two-deep | Bus factor and onboarding friction (RISK-PRD-003, RISK-PRD-009) |
| AI-provenance distribution | PR label distribution (`ai/none`/`ai/assisted`/`ai/generated`) from the Volume 11 reporting | Honesty of the project's own development claims; input to sustainability review |
| Extension ecosystem pulse | Extensions observed in the wild (registry/index listings where they exist, tracker mentions); SM-02/SM-03 exercise trends | PRD-007 vitality; contract stability pressure (RISK-PRD-010) |
| Benchmark trend headroom | Distance between nightly medians and NFR-PERF thresholds | Early warning before gates fail (RISK-PERF-001) |

### Sustainability posture

1. **Scope discipline is the primary sustainability tool.** The narrow support window
   (ADR-193), the closed MVP minimum, and rule R5 (chapter 03) exist so a small
   maintainer set can keep every published promise; widening promises requires widening
   capacity first (rule R9).
2. **Funding.** The project MAY accept sponsorship (platform sponsorship mechanisms or a
   fiscal host); funds are spent on project infrastructure (CI capacity, signing,
   benchmark hardware per ADR-160) by significant decision, with a published ledger
   summary per phase gate. Sponsorship buys no decision rights (chapter 04 principles).
   No sustainability model that contradicts ADR-002's license or the Out of Scope list
   (no hosted service) is entertained without the full change procedure.
3. **Contributor growth is engineered, not hoped for.** The SM-01/SM-02/SM-03 budgets
   make contribution cost a measured product property; curated onboarding work with
   named mentors is a standing triage duty; extension authorship is the widest funnel
   into the contributor pipeline (Tomás path above).
4. **Succession and continuity** follow the chapter 04 bus-factor obligations; the
   quarterly process audits verify them in practice.
5. **Honest ambition.** The competitive-pace risk (RISK-PRD-008) is answered with the
   specification's own strategy — vendor independence, local-first guarantees,
   auditable autonomy (PRD-002, PRD-003, PRD-005, PRD-006) — not with promise inflation;
   the phase gates exist so the project never claims what its gates have not verified.
