# 04 — Threats: System, Credentials, and the Human Channel

This chapter enumerates the threats whose vector is the **operating-system surface** Andromeda
touches — command execution, filesystem paths, symlinks, process isolation, and privileges — plus
the **credential** threats (exfiltration, theft, log leakage) and the **human channel** (social
engineering of the user's approval decisions).

The controls are the sandbox specification (keystone FR-SEC-101; ADR-021), credential and secret
management (keystone FR-SEC-102; ADR-014), the permission model (keystone FR-SEC-100), and the
redaction and audit controls of chapters 07–08. The threat entries reference them by name and
keystone identifier.

A premise for the whole chapter: Andromeda cannot defend against an attacker who already holds the
user's OS account (stated in the overview). These threats concern keeping *lower-trust inputs* —
model output, tool arguments, repository content, extensions — from reaching the OS surface with
more authority than the user granted.

### RISK-SEC-019 — Command injection

- Category: System / execution
- Probability: High
- Impact: High
- Severity: Critical
- Mitigation: No shell interpolation of untrusted values; parameterized argv execution via TerminalPort/SandboxPort; the `execute` permission and approval for command classes; command allow/denylists evaluated by the Permission Manager; sandbox resource limits and env filtering (ADR-021)
- Detection: Command Execution records and `terminal.execution.*` events; permission-decision audit; sandbox-refusal events; anomalous-command metrics
- Owner: Terminal Engine (Volume 6) / Sandbox Engine (Volume 9)
- Status: Open

#### Asset

The host: its filesystem, processes, and network position reachable by an executed command.

#### Actor

A confused-deputy model steered by injection, a malicious tool or extension, or repository content
whose values flow into a command.

#### Vector

Untrusted input reaching a shell as command text — metacharacters, chained commands, or substituted
arguments — causing execution of operations the user did not authorize.

#### Preconditions

A path exists where untrusted input is concatenated into a shell string, or a command tool executes
arguments without a permission decision.

#### Impact

Arbitrary command execution under the user's identity, bounded by sandbox containment and the
permissions in force. High impact and Critical severity because command execution is the most direct
route to host compromise and is a frequent agent action.

#### Prevention

Andromeda does not interpolate untrusted values into a shell string: commands run as parameterized
argv through `TerminalPort`, launched exclusively via `SandboxPort.ExecuteIn` (direct spawning
outside the Sandbox Engine is a defect, ADR-009 note). Command execution requires the `execute`
permission; command classes the Permission Manager marks dangerous require Approval. Allowlists and
denylists are evaluated by the Permission Manager; the sandbox applies deny-by-default environment
passthrough, resource limits, and working-directory policy (ADR-021).

#### Response

A denied command lands the Tool Invocation in `denied` (exit code 5 semantics) and emits a security
event; a running command is terminated through sandbox teardown, killing its process tree.

#### Recovery

Effects of an executed command are captured as Command Execution and File Change records for review
and, where they are file changes, revert through the Git Engine. Nothing executes on a denied
command.

#### Residual risk

At MVP the sandbox provides process-level controls, not OS-level isolation (ADR-021): an approved
command can still, for example, open a network connection. OS-level isolation from Beta/v1 narrows
this and is PENDING VALIDATION per platform (tracked in the volume register). Least-privilege grants
and approval for dangerous classes bound the MVP residual.

#### Tests

Command-injection fixtures assert no shell interpolation of untrusted input; permission-matrix tests
confirm `execute` gating and approval for dangerous classes; sandbox tests confirm env filtering,
resource limits, and teardown.

### RISK-SEC-020 — Path traversal

- Category: System / filesystem
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Path canonicalization and workspace-root confinement; the `path` scope and `read`/`write` permissions per resolved path; rejection of paths escaping the permitted root; sandbox working-directory policy (ADR-021)
- Detection: Path-policy refusal events; File Change records with resolved absolute paths; audit of out-of-root access attempts
- Owner: Sandbox Engine (Volume 9) / Tool Runtime (Volume 6)
- Status: Open

#### Asset

Files outside the intended working scope: system files, other projects, credentials on disk.

#### Actor

A confused-deputy model, a malicious tool, or repository content supplying a crafted path.

#### Vector

A path containing `..` sequences, absolute prefixes, or encoded separators that resolves outside the
permitted workspace root, letting a file tool read or write where it should not.

#### Preconditions

A file tool accepts a path value and resolves it without canonicalization and root confinement.

#### Impact

Read of sensitive out-of-scope files (including secrets on disk) or write/overwrite outside the
workspace. High impact because it breaches the filesystem boundary the user assumed.

#### Prevention

File tools canonicalize every path and confine it to the permitted root before use; a resolved path
escaping the root is rejected with an E-SEC decision. Access is scoped by the `path` permission scope
and the `read` / `write` permissions per resolved path. The sandbox enforces working-directory and
path policy (ADR-021), so even an approved tool cannot act outside its allowed paths.

#### Response

An out-of-root resolution is denied and audited; the operation does not proceed. Repeated attempts
from one tool can trigger policy disablement.

#### Recovery

No state change on a rejected path. A write that occurred within scope but wrongly is reverted through
File Change / Git records.

#### Residual risk

Paths the user explicitly grants outside the default root (a legitimate multi-directory workflow)
widen scope by consent; scoping grants to specific paths rather than broad roots bounds this.

#### Tests

Traversal fixtures (`..`, absolute, encoded separators) assert rejection; canonicalization tests;
per-path permission tests; sandbox working-directory confinement tests.

### RISK-SEC-021 — Symlink attacks

- Category: System / filesystem
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Symlink resolution checked against the permitted root after following links; no-follow or verify-then-act on sensitive operations; sandbox path policy applied to the resolved target (ADR-021); atomic operations where the target could change (TOCTOU)
- Detection: Path-policy refusal events on resolved symlink targets; audit of symlink-resolved access; File Change records with final resolved paths
- Owner: Sandbox Engine (Volume 9) / Platform Abstraction Layer (Volume 3)
- Status: Open

#### Asset

Files the resolved target of a symlink points to, potentially outside the intended scope.

#### Actor

A remote content author who plants a symlink in a repository, or a local process that swaps a target
between check and use (time-of-check-to-time-of-use).

#### Vector

A symbolic link within the workspace whose target lies outside the permitted root, or whose target is
swapped after a permission check, redirecting a read or write to a sensitive file.

#### Preconditions

A file operation follows a symlink and either does not re-confine the resolved target or performs a
non-atomic check-then-act.

#### Impact

Read or write redirected outside scope — a path-traversal effect achieved through link resolution.
Medium impact and severity because it requires planting or racing a link and is bounded by the same
path controls once the target is resolved.

#### Prevention

File operations resolve symlinks and re-check the final target against the permitted root before
acting; sensitive operations use no-follow semantics or verify-then-act atomically to close the
TOCTOU window. The sandbox path policy applies to the *resolved* target, not the link path, so a link
pointing out of scope is refused (ADR-021). The Platform Abstraction Layer provides the per-OS
primitives for safe resolution.

#### Response

A resolved target outside scope is denied and audited; a detected mid-operation target change aborts
the operation.

#### Recovery

No state change on a rejected operation; a wrongful write within a race window is reverted through
File Change / Git records.

#### Residual risk

TOCTOU races cannot be eliminated on all platforms with equal strength; atomic operations and
resolved-target confinement reduce the window. Per-platform primitive availability is validated with
the sandbox mechanisms (ADR-021, PENDING VALIDATION).

#### Tests

Symlink fixtures pointing outside the root assert rejection; TOCTOU race tests on sensitive
operations; resolved-target confinement tests across platforms.

### RISK-SEC-022 — Secret exfiltration

- Category: Credentials / egress
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Secrets held only in the Secret Store and passed as references (ADR-014); deny-by-default environment passthrough to child processes (ADR-021); the `network` permission gates egress; redaction in logs, errors, memory, and events (chapter 07); least context exposure to providers
- Detection: Egress permission-decision audit; redaction-conformance checks; secret-access audit via SecretStorePort.Get; anomaly in outbound content
- Owner: Secret Store (Volume 9) / Context Manager (Volume 7)
- Status: Open

#### Asset

Secret material: provider API keys, OAuth tokens, and other credentials.

#### Actor

A confused-deputy model, a malicious tool or extension, or injected content attempting to read and
send secrets out.

#### Vector

An attempt to read secret material and transmit it — through a network call, a written file, a log
line, a provider request, or a tool result — to an attacker-controlled destination.

#### Preconditions

Secret material is accessible to a component that also has, or can obtain, an egress path.

#### Impact

Disclosure of credentials enabling account takeover or provider abuse at the user's expense. High
impact because credentials are the highest-value asset.

#### Prevention

Secrets live only in the Secret Store (OS keychain or age-encrypted fallback, ADR-014); only
`SecretStorePort` touches material, and only references (`secret_ref`) cross other ports. Child
processes receive deny-by-default filtered environments, so secrets in the environment do not leak to
spawned commands (ADR-021). Egress requires the `network` permission, surfacing an exfiltration
attempt as a deniable decision. Logs, errors (ADR-016 envelope), events, and memory are redacted
(chapter 07). The Context Manager sends the least necessary content to providers.

#### Response

An egress attempt carrying credential-shaped content is deniable at the `network` decision and
audited; every `SecretStorePort.Get` is audit-logged, so secret access is attributable. Suspected
exposure triggers rotation (AuthPort `Rotate`).

#### Recovery

Rotate or revoke exposed credentials (`rotated` / `revoked` status); review the run and secret-access
audit chain to scope the exposure; re-authenticate affected providers.

#### Residual risk

A user who grants broad `network` access and approves egress can still be socially engineered into
exfiltration (RISK-SEC-027); scoped egress by `domain` / `host`, redaction, and access auditing bound
and expose it. Content already sent cannot be recalled.

#### Tests

Redaction-conformance tests over logs, errors, events, and memory; environment-filtering tests
confirming secrets do not reach child processes; egress-gating tests; secret-access audit tests.

### RISK-SEC-023 — Credential theft

- Category: Credentials / storage
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: OS keychain storage with the age-encrypted file fallback (ADR-014); never plaintext at rest; credentials global-only in storage (ADR-028); zeroize-on-release value wrappers; access auditing; the `credential_access` permission
- Detection: SecretStorePort.Get audit trail; keychain-access anomalies; fallback-file integrity checks; credential-status transitions
- Owner: Secret Store (Volume 9) / Authentication Layer (Volume 5)
- Status: Open

#### Asset

Stored credential material at rest.

#### Actor

An attacker with read access to Andromeda's data files or environment, a malicious extension seeking
stored secrets, or malware on the host (bounded by the OS-account exclusion).

#### Vector

Reading credentials from storage — a plaintext file, an unprotected environment, or an unencrypted
fallback — rather than intercepting them in transit.

#### Preconditions

Credentials are stored in a location or form readable by a lower-trust component or a file-system
reader.

#### Impact

Direct theft of credentials with the same consequences as exfiltration. High impact.

#### Prevention

Credentials are stored only in the OS keychain (zalando/go-keyring) or, opt-in, an age-encrypted file
fallback — never plaintext at rest (ADR-014). Credentials are global-only in storage (ADR-028), not
duplicated per workspace. `SecretValue` is a zeroize-on-release wrapper callers MUST NOT persist or
log. Reading credential material requires the `credential_access` permission and is audit-logged.
Extensions never receive material, only references.

#### Response

Anomalous credential access raises a security event; suspected theft triggers rotation/revocation of
the affected credentials and re-authentication.

#### Recovery

Rotate or revoke the stolen credentials (`rotated` / `revoked`); re-authenticate providers; review the
access audit chain to scope which credentials were reachable.

#### Residual risk

The age-encrypted fallback's protection depends on the passphrase and host security; where the OS
keychain is available it is preferred. An attacker already holding the OS account is out of scope
(overview). Residual risk is the fallback threat model on hosts without a keychain.

#### Tests

Storage tests asserting no plaintext credential at rest; fallback encryption/decryption tests;
zeroize-on-release tests; `credential_access` gating and access-audit tests.

### RISK-SEC-024 — Sandbox escape

- Category: System / isolation
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Layered sandbox (ADR-021) with process-level controls at MVP and OS-level isolation from Beta/v1; deny-by-default env, path, and resource policy; effective containment level observable per execution; teardown of full process trees; no bypass of SandboxPort for launches
- Detection: Effective-containment-level records per execution; sandbox-refusal and teardown events; resource-limit-breach metrics; audit of launch paths
- Owner: Sandbox Engine (Volume 9) / Platform Abstraction Layer (Volume 3)
- Status: Open

#### Asset

The containment boundary around executed tools, commands, plugins, and MCP servers.

#### Actor

A malicious tool, command, plugin, or MCP server attempting to break out of its applied policy to
reach the broader host.

#### Vector

Exploiting a weakness in the isolation mechanism or a launch path that bypasses `SandboxPort` to
execute with more access than the applied policy allows.

#### Preconditions

Execution occurs with an isolation mechanism weaker than assumed, or a code path spawns a process
outside the Sandbox Engine.

#### Impact

Loss of containment: the executed code reaches host resources beyond its policy. High impact; the
severity of a breakout is the whole host surface, bounded by the user's own OS privileges.

#### Prevention

All executions enter through `SandboxPort.ExecuteIn`; direct process spawning outside the Sandbox
Engine is a defect caught structurally (ADR-009 note; ADR-033 prohibited-construct scanner). The
sandbox applies deny-by-default environment, path, and resource policy at MVP (process-level) and
OS-level isolation from Beta/v1 (Seatbelt, Landlock, namespaces/bubblewrap — each PENDING VALIDATION
per platform, ADR-021). The effective containment level is recorded per execution so a weaker layer
is never silently substituted; degradation to a weaker layer is explicit and observable. Teardown
terminates the full process tree.

#### Response

A detected bypass or unexpected containment level raises a security event; the execution is torn down;
the launch-path defect is treated as a release blocker.

#### Recovery

Terminate the offending process tree; review effects through Command Execution / File Change records;
disable the responsible tool or extension.

#### Residual risk

MVP process-level controls do not fully isolate an approved malicious binary from the host (ADR-021
honestly states this); OS-level isolation from Beta/v1 narrows it, contingent on per-platform
validation. This phase-dependent limit is documented to users (ADR-021) and tracked in the volume
register.

#### Tests

Launch-path audits asserting SandboxPort-only execution; containment-level recording tests;
env/path/resource-policy enforcement tests; teardown process-tree tests; per-platform isolation
validation once mechanisms are confirmed.

### RISK-SEC-025 — Privilege escalation

- Category: System / permissions
- Probability: Low
- Impact: High
- Severity: Medium
- Mitigation: Least-privilege grants with scopes (keystone FR-SEC-100); no privilege elevation by Andromeda; confused-deputy prevention through per-action mediation; grant persistence and revocation audited; approval for scope-widening
- Detection: Permission-decision and grant-change audit; anomaly between an action and its authorizing grant; scope-escalation attempts logged
- Owner: Permission Manager (Volume 9) / Policy Engine (Volume 9)
- Status: Open

#### Asset

The permission grant set and the authority it represents.

#### Actor

A malicious tool, extension, or injected agent behavior attempting to gain broader authority than
granted, including tricking a more-privileged component into acting on its behalf.

#### Vector

Widening scope without consent — reusing a grant beyond its scope, chaining a narrow grant into a
broad effect, or exploiting a component that holds authority (confused deputy) to perform a
privileged action.

#### Preconditions

A component can perform an action broader than its own grant, or a grant is applied beyond its
intended scope.

#### Impact

Actions performed with authority the user did not confer for that purpose. High impact; Low
probability given per-action mediation and scoping, yielding Medium severity.

#### Prevention

Andromeda runs under the user's own identity and never elevates OS privilege (ADR-032). Every
side-effecting action is mediated per invocation by `PermissionPort` against the specific grant and
scope (keystone FR-SEC-100); a grant does not transfer across scopes. Scope-widening requires a new
decision and, for dangerous classes, Approval. Grants persist with their scope and are auditable and
revocable; the Policy Engine constrains what may be auto-granted.

#### Response

An action exceeding its authorizing grant is denied and audited; a detected confused-deputy pattern
raises a security event and can trigger grant revocation.

#### Recovery

Revoke the over-broad or misused grants; review the audit chain to scope actions taken; re-establish
grants at the intended scope.

#### Residual risk

A user who grants broad, long-lived scopes reduces the benefit of per-action mediation; session- and
command-scoped grants and visible grant state bound this behavioral residual.

#### Tests

Scope-isolation tests confirming a grant does not apply beyond its scope; confused-deputy tests over
authority-holding components; grant-revocation and audit tests; approval-on-widening tests.

### RISK-SEC-026 — Log leakage

- Category: Credentials / observability
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Redaction of secrets and sensitive data in logs, errors, events, traces, and cost records (chapter 07; ADR-011 slog); safe-to-log field discipline in the error envelope (ADR-016); local-first logs with access under the host account; no secret material in telemetry
- Detection: Redaction-conformance tests over log/event/trace output; secret-pattern scanning of emitted records; audit of exported telemetry content
- Owner: Logging / Observability (Volume 10) / Security (Volume 9)
- Status: Open

#### Asset

The content of logs, events, traces, cost records, and any exported telemetry.

#### Actor

A reader of log output — a support recipient, a shared paste, an exported bundle — or a component that
writes sensitive data into a log.

#### Vector

Secret material or sensitive content written into a log line, an error message, an event payload, or a
telemetry export, disclosing it to whoever can read that output.

#### Preconditions

A component emits a record containing unredacted secret or sensitive content.

#### Impact

Disclosure of credentials or sensitive data through observability output — often shared widely during
debugging. Medium impact and severity because logs are local-first and access-bounded, but sharing
amplifies exposure.

#### Prevention

Logging redacts secrets and sensitive data before writing (chapter 07); the error envelope separates
user-facing and safe-to-log technical fields, and only safe-to-log context data is recorded (ADR-016).
Events carry identities, numbers, and codes — never content or secrets. `SecretValue` MUST NOT be
logged. Logs are structured slog JSON (ADR-011), local-first, readable under the host account.
Telemetry export carries no secret material and is consent-gated (Volume 10).

#### Response

A detected unredacted secret in output is a defect that blocks release; the affected credential is
rotated as a precaution; the emitting code path is corrected.

#### Recovery

Rotate any credential that appeared in a log; purge or re-scope the affected log/export; correct the
redaction gap.

#### Residual risk

Novel sensitive-data shapes may evade pattern-based redaction; field-discipline (safe-to-log
allowlisting rather than denylisting) in the envelope reduces reliance on pattern matching. Residual
risk is unrecognized sensitive shapes, bounded by allowlist discipline.

#### Tests

Redaction-conformance tests over logs, errors, events, and traces; secret-pattern scans of emitted
records in CI; telemetry-content tests asserting no secret material is exported.

### RISK-SEC-027 — Social engineering

- Category: Human channel
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Approval prompts name the concrete effect and scope (chapter 09); dangerous actions default to deny and require explicit consent; no dark patterns in prompts (Volume 8); scoped rather than blanket grants; visible audit of what was approved
- Detection: Approval-decision audit; correlation of an approved dangerous action to the prompt that authorized it; anomaly in approval rate
- Owner: Permission Manager (Volume 9) / TUI (Volume 8)
- Status: Open

#### Asset

The user's approval decisions — the human-in-the-loop control that gates dangerous actions.

#### Actor

A social engineer operating through injected content, a malicious extension's prompts, or crafted
model output that manipulates the user into approving a harmful action.

#### Vector

Content or output engineered to make a dangerous action look benign or urgent, so the user approves it
— defeating the permission model through the person rather than the code.

#### Preconditions

A dangerous action reaches an approval prompt and the user can be misled about its true effect.

#### Impact

The user authorizes a harmful action with their own consent — bypassing technical controls by
manipulating the decision-maker. High impact because consent overrides denial.

#### Prevention

Approval prompts state the concrete effect and scope of the action in Andromeda's own trusted UI
(chapter 09), not text supplied by the untrusted content, so the user sees what will actually happen.
Dangerous classes default to deny and require explicit consent. The TUI avoids dark patterns and
pre-selected dangerous defaults (Volume 8). Grants are offered at the narrowest useful scope
(`allow_once`, `allow_for_session`) rather than blanket policy. Every approval is audited, so a
misled decision is reviewable after the fact.

#### Response

An approved action that proves harmful is identifiable through the approval audit; the user can revoke
standing grants and interrupt in-flight runs. Patterns of manipulative content can inform future
prompt wording and detection.

#### Recovery

Revoke grants obtained through manipulation; revert file and Git effects through their records;
re-establish grants deliberately at the intended scope.

#### Residual risk

No technical control fully prevents a determined manipulation of a consenting user; trustworthy,
effect-naming prompts, default-deny, scoped grants, and after-the-fact auditability bound and expose
it. This residual is inherent to human-in-the-loop authority.

#### Tests

Approval-prompt tests asserting the effect and scope are rendered from Andromeda's trusted state, not
from untrusted content; default-deny tests for dangerous classes; audit tests linking each approved
action to its authorizing decision; TUI tests asserting no pre-selected dangerous default.
