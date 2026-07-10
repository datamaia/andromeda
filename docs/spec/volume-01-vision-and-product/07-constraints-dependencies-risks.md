# 07 — Constraints, Dependencies, and Product Risks

## Provided constraints

The following constraints are externally mandated (Source: `Provided`, per the statement
classification of Volume 0, chapter 01). They are not design choices of this specification and
can be changed only by the project owner through the change procedure. Constraint identifiers
`PC-NN` are local to this volume.

- **PC-01 — English-language open-source project.** The specification, source code,
  identifiers, developer documentation, and community communication are written in English.
  The project is developed as open source; the license selection and governance model are
  decided in Volume 15 and recorded in the decision register.
- **PC-02 — GitHub as the development platform.** Code, issues, pull requests, discussions,
  projects, milestones, releases, CI (GitHub Actions), artifacts, security advisories, and
  traceability automation live on GitHub (Volume 11). This constrains the *project's*
  development platform; Git-hosting services as *product integrations* are a separate concern
  and remain multi-vendor.
- **PC-03 — Fixed product identity.** The executable is `andromeda`; the main configuration
  file is `andromeda.toml`; the remaining identity values (project directory, environment
  prefix, SDK name, protocol name, package prefix) are fixed in Volume 0, chapter 01.
- **PC-04 — Implementation-language preference: Go.** The project owner's stated preference
  for the implementation language is Go. This preference is an input — not a substitute — for
  the mandated technology evaluation (Rust, Go, TypeScript, Python across performance, binary
  distribution, memory safety, TUI ecosystem, concurrency, plugins, portability, and the other
  criteria listed in the brief-derived evaluation set). The final selection is documented as
  the foundation technology-selection decision record in the 001–039 allocation block
  (Volume 0, chapter 06), including alternatives considered and a reversal plan.
- **PC-05 — Platform order.** macOS and Linux come first (Tier 1 per
  [chapter 05](05-scope-and-phases.md)); native Windows 11 support is a later phase; WSL and
  native Windows support are distinct modalities, and WSL is not a substitute for native
  support.
- **PC-06 — Official mechanisms only.** Every integration with providers and services MUST use
  official, documented, authorized mechanisms. Andromeda MUST NOT use, under any circumstance:
  reverse engineering,
  captured cookies, tokens extracted from applications, private APIs, undocumented endpoints,
  automation of web interfaces to evade restrictions, or any mechanism contrary to a service's
  terms of service. This constraint ranks second in the requirement precedence order (Volume 0,
  chapter 01) — above functionality — and no capability justifies violating it.
- **PC-07 — A subscription does not imply programmatic access.** Andromeda MAY use a user's
  provider account or subscription only where an official, documented mechanism authorized for
  third-party clients exists. The existence of a subscription MUST NOT be treated as evidence
  of such a mechanism. Which providers offer account-based mechanisms usable by third-party
  clients, and under what terms, is PENDING VALIDATION per provider (open-questions register,
  [99-volume-register.md](99-volume-register.md)); Volume 5 records the per-provider outcome.

## Dependencies

Andromeda's product promises depend on external systems and ecosystems it does not control.
Each dependency is listed with its failure mode and the risk that covers it.

| Dependency | Nature | Failure mode | Covered by |
|---|---|---|---|
| Cloud provider inference APIs | Official, documented HTTP APIs of model providers | Breaking changes, deprecations, outages, quality regressions | RISK-PRD-001 |
| Provider access policies and terms | Legal/commercial conditions for third-party clients | Narrowed access, pricing shifts, restricted auth mechanisms | RISK-PRD-002 |
| Local inference servers | Ollama and OpenAI-compatible local servers as local providers | API drift, packaging changes, capability variance | RISK-PRD-001, RISK-PRD-006 |
| Openly available local models | Models with tool-calling capability runnable on user hardware | Capability gaps making local-first workflows underperform | RISK-PRD-006 |
| Model Context Protocol | Published MCP specification versions and the server ecosystem | Protocol churn, server quality variance, supply-chain exposure | RISK-PRD-007 |
| Operating-system facilities | Keychain / Secret Service, PTYs, signals, filesystems (via the Platform Abstraction Layer) | OS releases changing behavior; missing facilities on some systems | RISK-PRD-001 (pattern), Volume 3 platform risks |
| Git | Repository operations underpinning the Git Engine | Version behavior differences; repository-scale edge cases | Volume 11 risks |
| GitHub platform | The project's development, CI, release, and security infrastructure (PC-02) | Actions/API changes, outages, policy changes affecting OSS projects | RISK-PRD-003 (continuity), Volume 11 risks |
| Signing and notarization services | Signing identities and macOS notarization for releases | Unavailable credentials delay signed releases (checksums unaffected) | Signing viability note, chapter 05 |
| Terminal emulators | Rendering and input environment for the TUI | Capability variance (color, Unicode, resize behavior) | Volume 8 compatibility requirements |

Dependency-management rules:

1. Every dependency interaction goes through an owned abstraction (Provider Layer, MCP
   Runtime, Platform Abstraction Layer, Git Engine), so a dependency change is an adapter
   change, not a redesign.
2. No dependency may be load-bearing for the offline guarantee list (chapter 04): the offline
   core depends only on local facilities.
3. Third-party product documentation is referenced, never restated as authority (Volume 0,
   chapter 01); adapter behavior is verified against live conformance runs, not assumptions.

## Product risks

Product-level risks are defined below using the Volume 0 risk template. Area-specific risks
(security, provider, distribution, …) are minted by their owning volumes; this register holds
the risks that threaten the product as a whole. All risks are reviewed at every phase gate
(chapter 05), and each mitigation is bound to mechanisms this corpus actually specifies.

### RISK-PRD-001 — Provider API drift and deprecation

- Category: External dependency
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: All provider-specific logic is confined to adapters (chapter 04, Principle 1),
  so drift is absorbed at the adapter boundary. Each adapter declares the API versions and
  deprecations it tracks (Volume 5); the provider conformance suite runs on a schedule against
  live APIs where credentials permit; the generic OpenAI-compatible adapter provides a
  structural fallback path for OpenAI-compatible services; capability declaration limits the
  blast radius of a partial outage to the affected capabilities.
- Detection: Scheduled conformance CI runs failing per adapter; provider error-rate and
  error-category metrics (Volume 10) trending upward; adapter deprecation metadata flagging
  announced sunset dates at release audit.
- Owner: Provider Layer maintainers (Volume 5)
- Status: Open

### RISK-PRD-002 — Provider access-policy changes restricting third-party clients

- Category: External dependency
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: PC-06/PC-07 forbid unofficial workarounds, so the mitigation is architectural,
  not evasive: multi-provider support keeps every workflow portable to another provider;
  local providers are a policy-independent floor for core capability; Volume 5 documents,
  per provider, which authentication mechanisms are officially available to third-party
  clients (PENDING VALIDATION until confirmed per provider) and never ships speculative
  account-based access.
- Detection: Provider terms-of-service and developer-policy monitoring at each release audit;
  authentication failure categories in provider metrics; user reports triaged under a
  dedicated issue label.
- Owner: Product owner with Provider Layer maintainers
- Status: Open

### RISK-PRD-003 — Maintainer bus factor

- Category: Project sustainability
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Specification-first development is itself the primary mitigation: this corpus is
  precise enough for a competent implementer — human or AI agent under supervision — to
  continue the work (Volume 0, chapter 01). Additionally: governance in Volume 15 defines
  maintainer/committer roles and requires at least two release-authority holders as a v1 exit
  condition; the repository lives in an organization namespace, not a personal account
  (pending the namespace decision, Volume 0 open questions); processes (release, security
  response, ADR) are documented and automated in CI rather than held in one person's head.
- Detection: Bus-factor review at every phase gate: number of people with merge and release
  authority, contributor concentration (share of commits by top contributor), and unreviewed
  subsystem count reported in the governance section of each release audit.
- Owner: Product owner (governance, Volume 15)
- Status: Open

### RISK-PRD-004 — Scope creep versus MVP viability

- Category: Delivery
- Probability: High
- Impact: High
- Severity: Critical
- Mitigation: The MVP minimum (chapter 05) is change-controlled: additions require the Volume 0
  change procedure with recorded justification. Every capability carries a phase, and the
  viability assessment mandated by the brief (complexity, dependencies, maintenance cost, risk,
  security, implementability by agents) applies before any capability enters a committed
  phase; over-ambitious capabilities are moved to later phases with the decision recorded.
  Volume 15 sequences the MVP so that the end-to-end journeys (UC-01, UC-02, UC-03, UC-09,
  UC-11) are exercised early, exposing integration cost before breadth is added.
- Detection: Phase-gate audits comparing implemented scope against the MVP minimum;
  requirement-count trend per phase across document versions; MVP milestone burndown on
  GitHub Projects (Volume 11).
- Owner: Product owner
- Status: Open

### RISK-PRD-005 — Security incident caused by agent actions

- Category: Security (product trust)
- Probability: Medium
- Impact: High
- Severity: Critical
- Mitigation: The permission model, Sandbox Engine, and Audit Log are Core-phase components —
  the MVP never ships an unpermissioned execution path (chapter 04, Principle 8; SM-16
  enforces 100% Permission Manager mediation of side-effecting tool invocations). Destructive
  operations default to explicit confirmation; policies that widen autonomy are scoped,
  recorded, and revocable. Volume 9 defines the threat model, abuse tests, and the
  coordinated-disclosure process; Volume 13 includes fault-injection and permission-bypass
  test classes. Severity is classified Critical despite Medium probability because a single
  destructive incident attacks the product's central promise.
- Detection: Security events in the structured event stream; audit-chain tests (SM-13)
  failing on orphan side effects; permission-bypass attempts in the test suite; disclosure
  inbox and security advisories monitoring.
- Owner: Security area owner (Volume 9)
- Status: Open

### RISK-PRD-006 — Local-model capability gaps

- Category: Product hypothesis
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: Capability honesty (Principle 2) prevents the worst failure mode — pretending:
  missing capabilities are reported, degradation strategies are documented, and mandatory-
  capability workflows fail precisely. The local conformance suite (SM-04) maintains an
  evidence-based picture of which local serving paths and models sustain the agent loop;
  routing policies escalate mechanically-failing tasks to stronger models with the change
  announced; planner behavior is specified to tolerate weaker models via smaller, verifiable
  steps (Volume 4).
- Detection: SM-04 conformance results per pinned model per release; tool-call reliability
  (SM-10) segmented by provider class (cloud vs. local) in runtime metrics; user-reported
  local-model defect rate under a dedicated label.
- Owner: Provider Layer maintainers with Agent Runtime owners (Volumes 4, 5)
- Status: Open

### RISK-PRD-007 — MCP ecosystem instability

- Category: External dependency
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Andromeda supports documented MCP protocol versions only, with explicit version
  negotiation (Volume 6); the MCP conformance suite (SM-15) gates releases; MCP servers run
  under the same permission, trust, and isolation model as all third-party tools (Principle 4),
  so a misbehaving server is contained; the pinned reference-server interop set decouples
  release quality from ecosystem churn.
- Detection: SM-15 conformance and interop job failures; MCP specification release monitoring
  at release audit; connection-failure and protocol-error metrics from the MCP Runtime.
- Owner: MCP area owner (Volume 6)
- Status: Open

### RISK-PRD-008 — Competitive pace of funded incumbents

- Category: Market
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: Differentiate on structural properties that conflict with incumbent business
  models — local-first guarantees, vendor agnosticism, contractual auditability, open
  extension surfaces (chapter 01, Differentiation) — rather than racing feature-for-feature;
  phase discipline (RISK-PRD-004 mitigation) prevents reactive scope chasing; the experience
  references of chapter 01 keep interaction quality benchmarked against the state of the art
  without copying it; open-source extensibility lets the community add breadth the core team
  cannot.
- Detection: Landscape review at every phase gate documenting notable capability shifts in the
  reference products (public information only); adoption trend proxies (downloads, stars,
  extension counts) reviewed per release.
- Owner: Product owner
- Status: Open

### RISK-PRD-009 — Contributor onboarding friction

- Category: Community
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: The extension-effort metrics (SM-01, SM-02, SM-03) make onboarding cost a
  measured, defended property; SDK templates and fixtures give contributors a working starting
  point; the spec linter automates convention enforcement so review feedback is mechanical
  where possible; Volume 15 defines the contribution guide, labeled entry-level issues, and
  explicit paths from contributor to committer; the modular architecture (Volume 3) bounds how
  much a contributor must understand to change one area.
- Detection: Time from first PR to first merged PR for new contributors; PR abandonment rate;
  count of distinct contributors per release; recurring confusion themes in discussions,
  reviewed at phase gates.
- Owner: Product owner (governance, Volume 15)
- Status: Open

### RISK-PRD-010 — Public-contract churn breaking extensions

- Category: Platform integrity
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: SM-20 sets the stability target (0 breaking changes outside a major release,
  deprecation window ≥ 1 minor release) enforced by contract-diff tooling in CI; public
  contracts freeze as release candidates at Beta exit (chapter 05); the Extension SDK ships
  conformance fixtures so extension authors detect breakage before users do; versioned
  contracts and the deprecation policy live in Volume 14.
- Detection: Contract-diff findings on every release candidate; extension breakage reports
  triaged under a dedicated label; SM-20 compliance section in each release audit.
- Owner: Extension SDK area owner (Volume 6) with release engineering (Volume 14)
- Status: Open

## Risk posture summary

| Risk | Severity | Primary defense in this corpus |
|---|---|---|
| RISK-PRD-001 | High | Adapter isolation + conformance suites |
| RISK-PRD-002 | High | Multi-provider portability + local floor + official mechanisms only |
| RISK-PRD-003 | High | Specification-first continuity + governance |
| RISK-PRD-004 | Critical | Change-controlled MVP minimum + phase gates |
| RISK-PRD-005 | Critical | Core-phase permissions, sandbox, audit |
| RISK-PRD-006 | High | Capability honesty + local conformance evidence |
| RISK-PRD-007 | Medium | Versioned protocol support + containment |
| RISK-PRD-008 | High | Structural differentiation + phase discipline |
| RISK-PRD-009 | Medium | Measured extension effort + automated conventions |
| RISK-PRD-010 | High | Contract-diff enforcement + deprecation policy |

The two Critical risks share one property: they are self-inflicted failure modes — shipping too
much, or shipping unsafe autonomy. Both defenses are therefore internal discipline mechanisms
(phase gates, Core-phase safety components) rather than external contingencies, and both are
verified mechanically rather than by intention.
