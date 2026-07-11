# 04 — Open-Source Governance

This chapter is the authoritative governance model of the Andromeda project. The
repository files `GOVERNANCE.md` and `MAINTAINERS.md` summarize it (Volume 11 chapter
03); on any divergence, this chapter prevails and the summaries are corrected. Governance
changes follow the Volume 0 change procedure and the decision rules below.

Two decisions are inherited, not re-made here: the license is **Apache-2.0** repository-
wide with the inbound=outbound DCO model and the dependency-license policy of ADR-002;
the development platform, branch protection, review, and traceability rules are Volume
11's. Governance operates *through* those mechanisms.

## Principles

1. **Specification-governed.** The corpus (Volumes 0–15) is the project's contract with
   itself; normative changes land only through the change procedure. Governance cannot
   overrule the specification informally — it changes it formally.
2. **Recorded authority.** Every grant of authority (role, release, override, waiver) is
   a recorded artifact: a `MAINTAINERS.md` entry, an audited `policy-check` override
   (ADR-148), a waiver in gate evidence (Volume 13 chapter 04), or an ADR.
3. **Least surprise for contributors.** Rules that gate contributions are mechanical
   where possible (required checks) and documented where not; a contributor MUST be able
   to predict from public documents whether a change is acceptable.
4. **No pay-to-influence.** Sponsorship and employment buy no decision rights; decisions
   follow the role and process rules below.

## Roles

| Role | Who | Authority | How granted |
|---|---|---|---|
| User | Anyone using Andromeda | Files issues through the intake forms; participates in Discussions; votes informally through reactions and feedback channels | By use; no registration |
| Contributor | Anyone with a merged contribution (code, spec, docs, triage with recorded activity) | Everything a user has, plus attribution in the release notes stream; eligible for committer nomination | First merged PR (DCO-signed per ADR-002) |
| Committer | Regular contributors trusted with triage | Everything a contributor has, plus: issue triage rights (labels, milestones, project fields), PR review with non-binding approval, quarantine proposals (ADR-177), membership in the triage rotation | Nominated by any maintainer; confirmed by maintainer lazy consensus (5 working days); recorded in `MAINTAINERS.md` |
| Maintainer | The project's accountable owners | Everything a committer has, plus: binding review (CODEOWNERS), merge authority on owned paths, release authority (below), security-report handling, role confirmations, casting votes in formal decisions | Nominated by a maintainer after sustained committer contribution; confirmed by **two-thirds** of existing maintainers; recorded in `MAINTAINERS.md` with owned areas |

Additional hats (assignments, not ranks — always held by a maintainer, always recorded):

- **Release manager** — named per release issue; drives the Volume 14 pipeline for that
  release (authority below).
- **Security responder** — the on-point maintainer for the security inbox in the current
  rotation; the escalation roster (below) orders the fallbacks.
- **Moderator** — enforces the Code of Conduct for a given report; MUST NOT be a party to
  the reported incident.

Role hygiene: roles are per-person, never per-organization. A maintainer inactive for 6
months SHOULD be moved to emeritus status (retaining attribution, losing authority) by
maintainer consensus; emeritus maintainers are restored by a single confirming vote of
the active maintainers. A maintainer may be removed for cause (Code of Conduct violation,
sustained obstruction, credential misuse) by two-thirds of the *other* maintainers; the
removal record states the cause class, not private detail.

## Decision-making

Decision classes, from lightest to heaviest:

| Class | Examples | Mechanism |
|---|---|---|
| Routine | Ordinary PRs, triage, iteration planning | Volume 11 review rules: one approving review (code-owner where owned); lazy consensus |
| Significant | New dependencies, new CI gates, backlog reordering within a phase, committer confirmations | Lazy consensus among maintainers: proposal visible ≥ 5 working days in the tracker; silence is consent; any maintainer objection converts it to a formal decision |
| Formal | ADR-recorded decisions, phase-gate waivers, scope changes (RISK-PRD-004 class), maintainer confirmations, governance changes, deprecations | Maintainer vote: simple majority of active maintainers for ADRs and waivers; two-thirds for governance changes, maintainer confirmations/removals, and license-policy amendments. Votes are public in the tracker except conduct and security matters |
| Specification-normative | Anything altering Volumes 0–15 normative content | Formal decision **plus** the Volume 0 chapter 10 change procedure (issue → classification → templates → lint → review → registers) |

**Conflict resolution ladder.** (1) The PR/issue thread, with the disagreement stated as
alternatives; (2) a timeboxed Discussion (≤ 10 working days) producing a written option
summary; (3) maintainer vote per the class table; (4) for ties, the status quo stands —
a tie never forces change. There is no permanent tie-breaking chair; if the project later
adopts one, that is a governance change requiring two-thirds. Technical disputes about
already-specified behavior are not votes at all: the specification decides, and whoever
disputes it proposes a change through the procedure.

## RFC process

An RFC (request for comments) is required before implementation for changes that are
expensive to reverse:

- new or changed **public contracts** (provider contract, tool contract, ARP, skill
  format, workflow format, configuration schema, CLI structured-output schemas, event
  envelope — the SM-20 surface);
- **MAJOR document-set changes** (Volume 0 chapter 10 classification);
- **governance changes** to this chapter;
- anything a maintainer explicitly escalates to RFC.

Mechanics:

1. Open an issue through the architecture or research form (Volume 11 chapter 05) titled
   `RFC: <topic>`, containing: problem, constraints, at least two alternatives with
   trade-offs, the proposed decision, affected identifiers, and the migration/deprecation
   sketch where contracts change.
2. The RFC stays open for comment for a minimum of **10 working days**; the author
   revises in place with an edit log.
3. Disposition is a formal decision (maintainer vote). Accepted RFCs conclude as an ADR
   (next number in the appropriate block, 205 and up for post-consolidation evolution per
   Volume 0 chapter 03) plus the specification change itself; rejected RFCs are closed
   with the recorded reasons and remain citable.
4. An accepted RFC is not a blank check: the resulting PRs still pass every gate, and
   material deviation from the RFC's decision reopens it.

## ADR process

Per Volume 0 chapters 02/03/10, applied operationally: decisions and reversals require an
ADR using the full template, numbered from the appropriate block (205+ after
consolidation), stored under `docs/spec/annexes/adr/`, indexed in the Volume 0 decision
register, and linked from the changing volume's register. An ADR that supersedes another
marks the old one `Superseded by ADR-NNN`. Accepted ADRs bind until superseded —
including on maintainers who voted against them. Review conditions written into ADRs are
tracked at phase gates (chapter 01 gate operation).

## Security disclosure policy

The product-side mechanics are Volume 9 chapter 08 (incident response, audit evidence,
rotation flows); this section owns the *project-side* process.

1. **Private intake only.** Vulnerabilities are reported through the platform's private
   vulnerability reporting and the `SECURITY.md` instructions (Volume 11 chapter 03).
   Public issues describing suspected vulnerabilities are converted to private reports
   and the public copy minimized. The public security form collects only non-sensitive
   hardening work (FR-GH-006).
2. **Response commitment.** First substantive response within **3 business days**
   (NFR-SEC-003 / SM-16 c); tracking runs on the security inbox with timestamps, reviewed
   quarterly and at phase gates. The **escalation roster** — an ordered list of
   maintainers in `MAINTAINERS.md` — defines who acts when the on-point responder is
   unavailable: after 2 business days without acknowledgment, the next roster entry MUST
   take the report.
3. **Coordinated disclosure.** The project requests reporters allow a coordinated window
   while a fix is prepared; the project commits to a target of **90 days** from triage to
   public advisory, shortened when a fix ships earlier and extendable only by mutual
   agreement recorded in the advisory draft. Reporters are credited unless they decline.
4. **Fix and advisory.** Fixes land per rule R1 (chapter 03) at P0/P1; releases follow
   the backport policy (FR-REL-014, ADR-193) so every supported line receives the fix; a
   platform security advisory with severity, affected versions, and upgrade path
   accompanies publication; NFR-SEC-001 gates the release itself.
5. **Vulnerabilities in dependencies** follow the same path, coordinated upstream; the
   SBOM (ADR-013) identifies exposure per release.
6. **Product-recorded incidents** (Audit Log incident records, update verification
   failures per Volume 14) that indicate a project-side compromise — release, CI, or
   credential compromise (RISK-GH-001, RISK-SEC-014, RISK-SEC-015, RISK-SEC-016) —
   trigger this policy directly: yank affected releases (Release machine per Volume 14),
   rotate project credentials, publish an advisory.

## Code of Conduct

The project adopts the **Contributor Covenant, version 2.1**, as `CODE_OF_CONDUCT.md`
(file placement per Volume 11 chapter 03), with the following project-specific bindings:
the enforcement contact is the private conduct address listed in `CODE_OF_CONDUCT.md`,
reaching all maintainers; a maintainer who is party to a report MUST recuse from its
handling; enforcement decisions (warning, temporary restriction, permanent ban from
project spaces) are made by at least two non-recused maintainers; conduct records are
private, retained by the maintainers, and never published beyond the sanction itself. The
Code of Conduct applies in all project spaces — repository, Discussions, review threads,
and any official channel — and to conduct outside them when it targets project
participants.

## Contribution guide outline

`CONTRIBUTING.md` (Volume 11 chapter 03) MUST cover, in this order:

1. **Ways to contribute** — code, specification, documentation, triage, testing on Tier 1
   platforms, extension authorship (pointing at the SDK).
2. **Before you start** — search the tracker; `good-first-issue` and `help-wanted`
   curation; when an RFC is required (this chapter); when to open a Discussion instead of
   an issue.
3. **Development setup** — toolchain per ADR-001/ADR-018, local hook installation
   (`scripts/hooks/`), running the T0 gate locally.
4. **The rules that gate merges** — branch naming and PR conventions (Volume 11 chapters
   03/04), DCO sign-off (ADR-002), commit-content rules and AI-provenance labeling
   (ADR-015, FR-GH-005), size limits, human-review requirement, Definition of Done
   ([chapter 02](02-epics-milestones-sequencing.md)).
5. **Specification changes** — the change procedure, templates, lint, and register duties
   (Volume 0 chapter 10).
6. **Testing expectations** — pyramid placement, coverage floors (NFR-TEST-003), flaky
   policy (ADR-177), secret handling in tests (FR-TEST-008).
7. **Security reporting** — pointer to `SECURITY.md`; never file vulnerabilities
   publicly.
8. **Governance pointer** — roles and how committer/maintainer nomination works (this
   chapter); Code of Conduct.

## Release authority

Release semantics, channels, support windows, and the pipeline are Volume 14's; this
section fixes *who may exercise them*:

| Release class | Authority required |
|---|---|
| Nightly | None beyond CI — produced automatically from `main` (ADR-191 channel semantics); no human sign-off, no support obligation |
| Patch (stable or release branch) | The release manager (any maintainer self-assigned on the release issue); backport list per ADR-193's closed classes |
| Minor | Release issue with scope freeze (Volume 11 issue form); release manager named; lazy maintainer consensus on the scope |
| Major | RFC + formal decision (simple majority) on the breaking set and deprecation compliance (FR-REL-013, SM-20) before the release branch is cut |
| Yank | Any two maintainers jointly, immediately, with a public advisory; yank is the Release machine's `yanked` state (Volume 14) and is never silent |

Invariants: publication is refused without a complete qualification evidence bundle
(FR-TEST-009, NFR-TEST-006) — no human authority overrides a missing bundle; waivers
inside the bundle follow the Volume 13 waiver policy and are formal decisions; release
jobs run only in the protected environment with the ADR-149 posture; the release manager
MUST NOT approve their own qualification waiver.

## Trademark policy

The project name "Andromeda", the wordmark, the ASCII cat mascot, and the four-pointed
star (ADR-026 identity set) identify *this* project and its official releases. No
trademark registration is claimed at authoring time; pursuing registration is PENDING
VALIDATION (register row V15-OQ-2) and this policy applies as project policy either way:

1. **Nominative use is welcome** — truthfully referring to the project, writing about
   it, teaching it, linking it.
2. **Unmodified redistribution** (checksum-identical official artifacts) MAY use the name
   and identity set, and MUST NOT imply endorsement of a bundling product.
3. **Modified distributions and forks** MUST NOT present themselves as "Andromeda" or use
   the identity set as their own mark. Apache-2.0 grants code rights, not naming rights
   (the license itself does not grant trademark use); forks name themselves and MAY state
   compatibility factually ("based on Andromeda", "works with Andromeda").
4. **Extensions** MAY use "for Andromeda" phrasing ("<name> for Andromeda") and MUST NOT
   use bare "Andromeda" as the leading element of their own name or namespace; reserved
   namespaces are enforced mechanically (ADR-070, ADR-077, ADR-104).
5. Community spaces and packages (distro packaging, mirrors) using the name in a purely
   descriptive way are acceptable; the project MAY request corrections where a use
   implies official status.

## Plugin and extension policy

The technical contracts are Volume 6's (tool contract, ARP, skill format, package
sources, signature policy); the governance rules:

1. **Third-party extensions are their authors' works.** The project neither reviews nor
   warrants them; trust decisions ride on the Volume 6/9 mechanisms (declared
   permissions, sandbox tiers, checksums and cosign signatures per ADR-081, trust
   gating).
2. **Namespace discipline.** Built-in namespaces and the `mcp:`/`x` mounts are reserved
   (ADR-070, ADR-077, ADR-104); the project enforces reserved-namespace collisions
   mechanically and by registry policy, not by negotiation.
3. **Official registry.** Whether the project operates an official registry index — its
   hosting, publisher identity rules, and default-source status — is an open
   organizational decision (Volume 6 register V6B-OQ-4; ADR-080). Until decided,
   `git`/`archive`/`path` sources and third-party indexes are the supported distribution
   paths, and no extension is "official" except those shipped in this repository.
4. **Malicious-extension response.** A reported malicious extension is a security report
   (disclosure policy above). Where a project-operated index exists, the entry is
   delisted immediately on confirmation; advisories name affected versions.
5. **Contribution path to built-in status.** An extension becomes a built-in only through
   the ordinary specification change procedure (catalog amendment in the owning volume)
   plus code contribution under this chapter's rules — never by private arrangement.

## Deprecation and support policy

Owned by Volume 14 and referenced here as governance obligations: deprecations follow
FR-REL-013 (declared in release notes, warning emission, minimum one minor release of
overlap per SM-20); support windows and backports follow FR-REL-014 and ADR-193 (pre-v1:
latest release only; from v1: latest minor of the current major with full fixes, plus the
previous major's last minor receiving security/integrity fixes for 12 months;
cherry-pick-only release branches). Governance adds: a deprecation is a **formal
decision**; the deprecation ledger reconciliation in the release audit (Volume 14) is
part of the release manager's duty; and support-window changes are announced at least one
minor release before taking effect.

## Community channels

| Channel | Purpose | Rules |
|---|---|---|
| Issue tracker (forms) | Actionable work: bugs, features, all Volume 11 issue types | Forms only; blank issues disabled; triage per chapter 03 |
| GitHub Discussions | Design questions, Q&A, RFC pre-discussion, show-and-tell | Keeps Issues actionable (Volume 11 chapter 03); moderated per the Code of Conduct |
| Security inbox | Private vulnerability reports | Disclosure policy above; never public |
| Releases and changelog | Announcements: every release, deprecation, advisory | Generated per ADR-015/Volume 14; the single announcement channel of record |
| Roadmap project | Public plan of record | Volume 11 Projects operation; chapter 03 is its narrative |

Real-time chat is deliberately not committed at authoring time: a synchronous channel
without moderation capacity is a liability (RISK-PRD-003). Maintainers MAY establish one
by significant decision; if established, it MUST be listed in `README.md`, moderated
under the Code of Conduct, and treated as non-archival — decisions reached in chat do not
exist until recorded in the tracker.

## Bus-factor mitigation

RISK-PRD-003 (maintainer bus factor) is a standing governance concern; the mitigations
are obligations, not aspirations:

1. **Two-deep critical paths.** Every critical operational capability — release
   pipeline operation, signing material, security inbox, CI administration, the
   protected-environment secrets — MUST have at least two current holders/operators once
   the project has ≥ 2 maintainers; until then, the gap is a recorded risk reviewed at
   every phase gate (rule R9 of chapter 03).
2. **Credential inventory.** `MAINTAINERS.md` references an inventory of project
   credentials (what exists, where held, who can rotate); the inventory itself contains
   no secret material. Rotation on any maintainer departure is mandatory and audited
   (Volume 9 incident procedure applies on suspicion of exposure).
3. **Org-owned assets.** The repository, CI configuration, package identities (Homebrew
   tap, module path), and any registry identity are owned by the project organization,
   never by a personal account (the namespace decision itself is V11-OQ-1).
4. **Process over memory.** Everything the release manager and security responder do is
   executable from written procedure (Volume 14 pipeline, Volume 9 chapter 08, this
   chapter); phase-gate audits verify the procedures by exercising them, not by reading
   them.
5. **Succession by default.** Committer and maintainer pipelines are tended
   deliberately: curated `good-first-issue` work, named mentors on onboarding issues
   (Volume 11 label taxonomy), and nomination reviews at each phase gate.
