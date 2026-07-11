# 03 — Threats: Extensions and Supply Chain

This chapter enumerates the threats whose vector is **code and content Andromeda trusts to some
degree**: extensions the user installs (plugins, skills, MCP servers), repositories the user works
on, the dependencies Andromeda and its extensions build on, and the pipeline that builds, releases,
and updates the product. It also covers the inference supply chain — the providers and local models
that supply model behavior.

"Supply chain" is the umbrella framing of this chapter rather than a separate numbered threat: its
concrete, individually mitigable instances are dependency attacks (RISK-SEC-013), CI compromise
(RISK-SEC-014), release compromise (RISK-SEC-015), and update compromise (RISK-SEC-016). Grouping
this way keeps every numbered threat testable.

The frozen decisions this chapter builds on are: plugins as JSON-RPC 2.0 subprocesses over the
Andromeda Runtime Protocol (ADR-009), MCP over the official SDK (ADR-010), release tooling of
goreleaser / cosign / syft / SLSA provenance (ADR-013), credential storage per ADR-014, and
dependency-rule enforcement in CI (ADR-033). Signature and notarization availability for this
project is PENDING VALIDATION (Volume 1 signing viability note); this chapter designs so that
enabling signatures is a configuration change, not a redesign.

### RISK-SEC-010 — Malicious plugins

- Category: Extensions / supply chain
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Plugins run as isolated subprocesses (ADR-009) mediated by permission and sandbox; least-privilege grants declared per plugin; signature and trust gating at install (ADR-013 patterns); no in-process third-party code (ADR-073)
- Detection: Plugin permission-decision audit; ARP handshake and capability logs; anomaly between declared and observed permission use; install-time verification events
- Owner: Plugin Runtime (Volume 6) / Permission Manager (Volume 9)
- Status: Open

#### Asset

The host and the user's grants exposed to an installed plugin's tools and behavior.

#### Actor

A malicious extension author who publishes a plugin, or a compromised plugin whose maintainer or
build was subverted.

#### Vector

A plugin that requests broad permissions, exfiltrates data, or performs unauthorized side effects
through the tool surfaces it exposes over the Andromeda Runtime Protocol.

#### Preconditions

The plugin is installed and started (`running` state), its tool surfaces are registered, and it
holds or can obtain side-effecting grants.

#### Impact

The range of actions the plugin's granted permissions allow — file, command, network, or Git
effects. High impact because a plugin is code the user chose to run, yet still lower-trust than
Andromeda itself.

#### Prevention

Plugins execute only as separate subprocesses (ADR-009), never in-process (ADR-073), reached through
the Tool Runtime and contained by `SandboxPort` (ADR-021). Each plugin declares its permissions;
grants default to deny and are scoped per plugin. Install verifies package integrity and signature
state per ADR-013 patterns through the Package Manager; the trust classification of an unsigned or
untrusted-origin plugin gates what it may request without explicit acknowledgment.

#### Response

Anomalous permission use raises a security event and can transition the plugin toward `disabled` by
policy; the plugin subprocess is terminated through sandbox teardown, killing its process tree.

#### Recovery

The plugin is disabled or removed via the Package Manager (`removed` state), its grants revoked, and
its effects reviewed and reverted through the audit chain and Git Engine.

#### Residual risk

A plugin that behaves until a trigger, then acts within its granted permissions, cannot be caught at
install; least-privilege, sandbox containment, and per-action approval for dangerous classes bound
the damage.

#### Tests

Plugin isolation tests confirm subprocess execution and teardown; permission-scoping tests confirm a
plugin cannot exceed declared grants; install-verification tests confirm signature/trust gating;
ARP conformance tests exercise the handshake and capability negotiation.

### RISK-SEC-011 — Malicious skills

- Category: Extensions / supply chain
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Skills declare required tools/capabilities and compatible providers; skills confer no authority beyond the tools they invoke (which remain permission-mediated); signing and trust at install; untrusted-content labeling of skill prompt text
- Detection: Skill install verification events; tool-invocation audit for skill-driven actions; descriptor-diff on skill update
- Owner: Skill Engine (Volume 6) / Permission Manager (Volume 9)
- Status: Open

#### Asset

The agent's behavior and grants exposed to a loaded skill's prompts and its required tool set.

#### Actor

A malicious skill author, or a compromised skill package.

#### Vector

A skill whose prompt content carries injection payloads, or whose required-tools declaration steers
the agent toward dangerous tools, or that changes behavior after the user trusted it.

#### Preconditions

The skill is installed and applied to a run; its prompts enter context and its required tools are
enabled.

#### Impact

Injected behavior and coerced tool use across every run that applies the skill. High impact because
a skill can shape the agent's whole approach, though it acts only through permission-mediated tools.

#### Prevention

A skill confers no authority of its own: it invokes tools, which remain permission-mediated
(keystone FR-SEC-100). Skill prompt text is untrusted-content labeled. Installation verifies package
integrity and signature state (ADR-013 patterns); an update that changes the declaration forces
re-consent. Required-tools and compatible-provider declarations are inspectable before the user
applies the skill.

#### Response

Anomalous skill-driven tool use raises a security event; the skill can be disabled by policy. A
changed declaration blocks application until re-consent.

#### Recovery

The skill is removed via the Package Manager (`removed`); its influence ends on the next run; effects
already produced are reviewed and reverted through the audit chain and Git Engine.

#### Residual risk

A persuasive skill can still lead the agent toward permitted-but-unwise actions; per-action approval
for dangerous classes and revertibility bound it.

#### Tests

Skill injection fixtures assert no unmediated side effect follows applied skill prompts;
required-tools tests confirm tools stay permission-gated; update-re-consent tests confirm a changed
declaration is re-approved.

### RISK-SEC-012 — Malicious repositories

- Category: Extensions / supply chain
- Probability: High
- Impact: High
- Severity: Critical
- Mitigation: Repository content treated as untrusted data; Git operations behind the adapter with no auto-execution of repo hooks or scripts; path and sandbox controls; permission and approval gating of any build/command the repo suggests
- Detection: Provenance on repo-sourced context; command and hook execution audit; path-policy and sandbox refusals; Git operation records
- Owner: Git Engine (Volume 11) / Sandbox Engine (Volume 9)
- Status: Open

#### Asset

The workspace, the host, and the agent loop when the user opens or works on an untrusted repository.

#### Actor

A remote content author who authors a repository the user clones or opens — a sample project, a
dependency source, a pull request under review.

#### Vector

A repository whose files carry injection payloads, whose build scripts or Git hooks attempt
execution, or whose structure (symlinks, path names) targets the traversal and symlink threats. This
is the aggregate, workspace-scale form of malicious files.

#### Preconditions

The user opens the repository as a workspace or the agent reads its content; a tool acts on repo
files or the agent is led to run repo-supplied commands.

#### Impact

The union of indirect injection, command injection, path traversal, and symlink attacks scoped to a
whole repository. High impact and Critical severity because working on untrusted repositories is a
core, frequent use case.

#### Prevention

Repository content is untrusted data. The Git Engine operates behind the ADR-025 adapter and does
not auto-execute repository hooks or build scripts; any command the repository suggests routes
through `PermissionPort` and, for dangerous classes, Approval, running under sandbox (ADR-021).
Path resolution enforces workspace-root policy (RISK-SEC-020) and symlink handling (RISK-SEC-021).
Git mutations require the `git_mutation` permission and are never silent.

#### Response

Attempts to execute repo-supplied code without consent are denied and audited; a repository showing
hostile structure can be excluded from context or closed. Path and symlink refusals emit E-SEC
decisions.

#### Recovery

Effects that executed revert through File Change / Git records; the repository can be closed and
removed from the workspace registry; poisoned memory/index entries are handled per RISK-SEC-007 /
RISK-SEC-008.

#### Residual risk

Content the user explicitly approves to build or run executes with their consent; the mitigation
makes that consent explicit and scoped rather than implicit. Residual exposure is user-approved
execution, minimized by naming concrete effects in prompts.

#### Tests

Hostile-repository fixtures assert hooks and build scripts do not auto-execute; path and symlink
fixtures assert containment; Git-mutation tests confirm permission gating and record creation.

### RISK-SEC-013 — Dependency attacks

- Category: Supply chain
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Pinned dependencies with checksums (Go module sums); dependency audit and vulnerability scanning in CI (ADR-033 patterns; SM-16); minimal dependency set; license policy (ADR-002); reproducible builds via goreleaser (ADR-013)
- Detection: Dependency and vulnerability scanning gate (CodeQL/Dependabot-class, SM-16); module-sum verification on build; SBOM diff (syft, ADR-013)
- Owner: Package Manager and CI (Volumes 6, 11) / Security (Volume 9)
- Status: Open

#### Asset

The integrity of the Andromeda binary and of extension packages: the code that actually runs.

#### Actor

A compromised dependency maintainer, or an attacker who publishes a malicious package version
(typosquatting, dependency confusion, account takeover).

#### Vector

A malicious or vulnerable third-party library pulled into Andromeda's build or into an extension,
executing with the process's authority.

#### Preconditions

A dependency is added or updated to a compromised version without detection before build and release.

#### Impact

Arbitrary behavior within the process — the strongest supply-chain foothold, since dependency code
runs unsandboxed inside Andromeda. High impact; probability Medium given the frequency of ecosystem
incidents balanced against pinning and scanning.

#### Prevention

Dependencies are pinned with verified module checksums; the dependency set is kept minimal;
vulnerability and dependency scanning gate releases (SM-16 target: zero known critical/high at
publication); the license policy (ADR-002) constrains what may be pulled in; builds are produced
reproducibly by goreleaser (ADR-013) with an SBOM (syft) so the shipped dependency set is auditable.

#### Response

A scan finding of critical/high severity blocks the release gate; a detected malicious version is
pinned out and the SBOM diff identifies affected artifacts.

#### Recovery

Roll the dependency back to a known-good pinned version, rebuild, and re-release; publish an advisory
per the disclosure policy (chapter 08). Affected users update through the verified update path.

#### Residual risk

A sophisticated, not-yet-known malicious version can pass scanning; minimal dependencies, pinning,
provenance (SLSA, ADR-013), and post-disclosure response bound the window. Whether a given third-party
scanner covers all relevant advisories is PENDING VALIDATION and tracked in the volume register.

#### Tests

Module-sum verification tests; CI dependency-audit gate tests; SBOM-generation and diff tests;
reproducible-build verification per release.

### RISK-SEC-014 — CI compromise

- Category: Supply chain
- Probability: Low
- Impact: High
- Severity: Medium
- Mitigation: Least-privilege CI tokens; pinned action versions; fork pull requests treated as untrusted; required checks and branch protection (Volume 11); provenance attestation (SLSA, ADR-013); secret scanning
- Detection: CI run audit; provenance verification of built artifacts; secret-scanning alerts; branch-protection and required-check enforcement
- Owner: CI / GitHub Actions (Volume 11) / Security (Volume 9)
- Status: Open

#### Asset

The build and release pipeline, and the signing/publishing credentials it holds.

#### Actor

An attacker who subverts CI: through a malicious pull request, a compromised action, or leaked CI
secrets.

#### Vector

Tampering with the build to inject code, or theft of pipeline credentials to publish a malicious
artifact — the pipeline as a distribution channel.

#### Preconditions

A CI workflow runs attacker-influenced code with access to secrets or publishing rights.

#### Impact

A tampered artifact distributed to all users, or leaked signing/publishing credentials. High impact;
Low probability given the pipeline controls, yielding Medium severity.

#### Prevention

CI tokens are least-privilege and scoped per job; third-party actions are pinned to specific
versions; pull requests from forks run without secrets and are treated as untrusted (Volume 11);
required checks and branch protection prevent unreviewed merges; artifacts carry SLSA provenance
(ADR-013) so consumers can verify the build's origin; secret scanning gates the repository.

#### Response

A provenance mismatch or a secret-scanning alert halts publishing; suspected compromise triggers
credential rotation and pipeline lockdown per the incident-response process (chapter 08).

#### Recovery

Rotate all pipeline credentials, rebuild from a verified clean state, re-attest provenance, and
re-release; yank any suspect release (`yanked` state, Volume 14) and advise users.

#### Residual risk

A determined attacker with a novel action-supply-chain foothold could still influence a build;
provenance attestation and post-incident verification shorten the window and support detection.

#### Tests

Pipeline-permission audits; pinned-action verification; fork-PR isolation tests; provenance
verification tests over released artifacts (Volume 11/14).

### RISK-SEC-015 — Release compromise

- Category: Supply chain
- Probability: Low
- Impact: High
- Severity: Medium
- Mitigation: Checksums for every artifact (unconditional at MVP); cosign signatures and SLSA provenance where signing is viable (ADR-013, PENDING VALIDATION per Volume 1); verified publishing; yank capability
- Detection: Checksum and (where enabled) signature verification at install/update; provenance verification; release-audit gate (Volume 14)
- Owner: Release pipeline (Volume 14) / Security (Volume 9)
- Status: Open

#### Asset

Published release artifacts and the trust users place in them.

#### Actor

An attacker who tampers with an artifact after build or substitutes a malicious artifact at the
distribution point.

#### Vector

A modified binary, checksum, or package served to users in place of the authentic release.

#### Preconditions

An artifact or its integrity metadata is altered between build and the user's download without
detection.

#### Impact

Users install a tampered binary running with their authority. High impact; Low probability with
integrity metadata in place, yielding Medium severity.

#### Prevention

Every artifact publishes a checksum unconditionally at MVP; cosign signatures and SLSA provenance are
applied where signing is viable for this project (ADR-013) — signature and notarization availability
is PENDING VALIDATION (Volume 1 signing note) and tracked in the volume register. The Updater verifies
checksums and, when enabled, signatures and provenance before applying (`Verify` precedes `Apply`,
which MUST refuse unverified artifacts).

#### Response

A verification failure blocks installation/update; a confirmed tampered release is yanked (`yanked`
state) so update checks no longer offer it, and an advisory is published (chapter 08).

#### Recovery

Republish a clean, verified release; users on a bad version update to the clean one through the
verified path or roll back to a retained good version (SM-19).

#### Residual risk

Before signing is validated for this project, checksums plus provenance are the integrity floor; a
sophisticated substitution that also forges the checksum is bounded by transport integrity and, once
enabled, cryptographic signatures.

#### Tests

Checksum-verification tests; signature/provenance verification tests where enabled; yank-and-refuse
tests confirming update checks skip yanked releases (Volume 14).

### RISK-SEC-016 — Update compromise

- Category: Supply chain
- Probability: Low
- Impact: High
- Severity: Medium
- Mitigation: Updater verifies before apply (UpdaterPort: Apply refuses unverified artifacts); atomic replace-or-restore; offline-clean check; consent per policy; retained prior version for rollback
- Detection: Update verification-failure events; apply/rollback audit; update-history records (Volume 14)
- Owner: Updater (Volume 14) / Security (Volume 9)
- Status: Open

#### Asset

The installed binary and the update mechanism that replaces it.

#### Actor

A network attacker who tampers with update traffic, or an attacker who compromised the update
endpoint or metadata.

#### Vector

A malicious update delivered through the self-update path (check → download → verify → apply),
attempting to replace the installed binary with a hostile one.

#### Preconditions

An update is offered and accepted whose artifact or metadata was tampered with, and verification is
absent or bypassed.

#### Impact

Replacement of the trusted binary with a malicious one — a durable host compromise. High impact; Low
probability given verification-before-apply, yielding Medium severity.

#### Prevention

The Updater checks release metadata, downloads, and verifies checksums and (where enabled) signatures
and provenance before applying; `Apply` MUST refuse to run when `Verify` has not passed for the same
artifact set. Apply is atomic replace-or-restore, never leaving a half-replaced installation. Update
consent follows policy; the prior version is retained for offline rollback (SM-19). The check path is
the only one requiring network and MUST fail cleanly offline.

#### Response

A verification failure aborts the update with the installed version unchanged and emits an event; a
suspected compromise pauses the update channel pending investigation (chapter 08).

#### Recovery

Roll back to the retained prior version offline (`rolled_back` state); once a clean release exists,
update to it through the verified path.

#### Residual risk

The strength of the guarantee tracks signing viability, which is PENDING VALIDATION; checksums and
atomic apply are the floor until signatures are validated for this project.

#### Tests

Verify-before-apply tests asserting refusal of unverified artifacts; tampered-artifact fixtures;
atomic-apply crash-injection tests; offline rollback tests (Volume 14).

### RISK-SEC-017 — Compromised providers

- Category: Supply chain / inference channel
- Probability: Low
- Impact: High
- Severity: Medium
- Mitigation: Official documented endpoints only over authenticated transport (ADR-019); minimal necessary context sent (least exposure); credentials via Secret Store; provider output treated as untrusted (RISK-SEC-006); explicit user notification on provider/model change; no unsafe fallback (Volume 5 guard rules)
- Detection: Provider request/response audit; egress content policy; anomaly in provider behavior transitioning it toward `degraded`; cost and usage monitoring
- Owner: Provider Layer (Volume 5) / Security (Volume 9)
- Status: Open

#### Asset

The content Andromeda sends to providers (which may include workspace data) and the trust placed in
the model behavior received.

#### Actor

A malicious or compromised provider, or a network attacker between Andromeda and the provider.

#### Vector

A provider that harvests submitted content, returns crafted output, or claims capabilities it does
not honor; or interception of provider traffic.

#### Preconditions

Andromeda sends content to the provider and consumes its output; the provider or the channel is
hostile.

#### Impact

Data exposure of submitted content and injection through returned output. High impact; Low
probability for reputable providers over authenticated transport, yielding Medium severity.

#### Prevention

Andromeda reaches providers only through official, documented endpoints over authenticated transport
(ADR-019); credentials resolve from the Secret Store as references (ADR-014), never inline. The
Context Manager sends the minimal necessary context (least exposure). Provider output is untrusted
(RISK-SEC-006). Any change of provider or model is explicitly announced to the user, and fallback
obeys Volume 5 guard rules that prevent unsafe or data-exposing switches. The provider connection
machine transitions a misbehaving provider toward `degraded` or `unavailable`.

#### Response

Anomalous behavior degrades the provider in routing and emits a security event; the user is notified;
suspected credential exposure triggers rotation (AuthPort `Rotate`).

#### Recovery

Switch to another configured provider through announced, guarded fallback; rotate any credential that
may have been exposed; review runs that used the suspect provider through the audit chain.

#### Residual risk

Content already sent to a provider cannot be recalled; least-exposure minimizes what is sent, and
local-first operation (Ollama, generic OpenAI-compatible against a local server) avoids remote egress
entirely where privacy demands it. Whether a specific provider retains submitted content is governed
by that provider's terms and is PENDING VALIDATION per provider (Volume 5).

#### Tests

Egress-content tests confirming least exposure and credential-by-reference; provider-change
notification tests; fallback guard-rule tests (Volume 5); untrusted-output tests (RISK-SEC-006).

### RISK-SEC-018 — Compromised local models

- Category: Supply chain / inference channel
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Local model output treated as untrusted (RISK-SEC-006); local server reached over the configured local endpoint only; model provenance is the user's responsibility with capability-honesty checks; no elevated trust for locality
- Detection: Structured-output and tool-argument validation over local model output; capability-conformance checks (SM-04); local endpoint audit
- Owner: Provider Layer (Volume 5) / Security (Volume 9)
- Status: Open

#### Asset

The agent loop and outputs when driven by a locally served model.

#### Actor

A party who supplied a tampered model weight file, or a compromised local model server.

#### Vector

A local model that produces crafted output or dishonestly declares capabilities, or a local serving
process that behaves maliciously — the local analogue of a compromised provider, without the network
egress.

#### Preconditions

The user runs a model or serving stack of uncertain provenance and Andromeda consumes its output.

#### Impact

Malicious output driving coerced tool use (bounded by permissions) and dishonest capability claims
degrading reliability. Medium impact (no remote egress, output confers no authority) and Medium
severity.

#### Prevention

Local model output is untrusted and passes the RISK-SEC-006 validation path; locality confers no
elevated trust. Andromeda reaches the local model only through the configured local endpoint.
Capability declarations are verified against observed behavior by the conformance suite (SM-04); a
model that fails capability honesty is not treated as if it had the capability. Model-weight
provenance is the user's responsibility and stated as such (no overclaim). Because a local model
runs on the user's own host with no additional network egress, the primary residual concern is
coerced tool use, which the permission and sandbox layers already bound.

#### Response

Validation failures deny derived actions; capability-honesty failures exclude the model from
capability-dependent flows and notify the user.

#### Recovery

Switch to a trusted local or remote model; review runs that used the suspect model through the audit
chain; no credential rotation is needed (local serving is unauthenticated by configuration).

#### Residual risk

A local model can always produce misleading output; because output confers no authority and there is
no remote egress, residual risk is bounded to permitted tool use and reliability, not data exposure.

#### Tests

Local-provider conformance suite (SM-04) over pinned local models; capability-honesty tests;
structured-output and tool-argument validation over local model output.
