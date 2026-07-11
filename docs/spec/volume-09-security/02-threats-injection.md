# 02 — Threats: Injection and Content Manipulation

This chapter enumerates the threats whose vector is the **model context**: adversarial content
that reaches the model as if it were instruction, and adversarial declarations that misrepresent
what a tool or extension does. These are the threats most specific to an AI engineering harness,
because the agent's authority is driven by text it reads.

The controls these threats reference are defined normatively in chapters 05–09 of this volume and
in Volume 6 (tool, MCP, skill, and plugin trust). The threat entries reference controls by name
and by keystone identifier; they do not restate them.

A shared premise underlies the whole chapter: **content is data, not instruction.** Andromeda's
architecture treats file contents, tool results, memory records, index hits, and provider output
as untrusted-content labeled input. No embedded directive in that content confers authority; only
the permission model (keystone FR-SEC-100) and the human approval path (chapter 09) confer
authority to act.

### RISK-SEC-001 — Prompt injection (direct)

- Category: Injection / AI-specific
- Probability: High
- Impact: High
- Severity: Critical
- Mitigation: Permission mediation of every side effect (keystone FR-SEC-100); default-deny tool grants; human approval for dangerous actions; sandbox containment (keystone FR-SEC-101); no privileged interpretation of instruction-like input
- Detection: Approval and permission-decision audit trail; anomalous tool-invocation-rate metrics; `permission.decision.recorded` and denial events; run-record review
- Owner: Permission Manager (Volume 9) / Agent Engine (Volume 4)
- Status: Open

#### Asset

The user's permission grants, workspace data, host, and provider budget — everything the agent's
granted authority can reach in one run.

#### Actor

A remote content author or a party who can influence the text of the user's own prompt, including
content pasted by the user from an untrusted source.

#### Vector

Attacker-controlled instructions placed directly into the prompt stream (the user's message or
content the user copies in) that attempt to override system guidance, escalate the agent's
behavior, or trigger side effects the user did not intend.

#### Preconditions

The agent has at least one side-effecting tool enabled, and the injected instruction targets an
action within the agent's current grants or one the user can be led to approve.

#### Impact

Unauthorized tool use: file modification or deletion, command execution, network egress carrying
data, or Git mutation — bounded by the permissions in force. Impact is rated High because the
blast radius is the union of standing grants; it is not Critical-by-construction because approval
and sandbox layers cap it.

#### Prevention

The agent's output never carries authority: every side-effecting action is decided by
`PermissionPort` against standing grants and policy (Principle 8), and dangerous classes require a
human Approval (chapter 09). Tool grants default to deny; the sandbox (ADR-021) contains
execution regardless of what the model was told. Instruction-like text in input is not given
privileged interpretation; the Prompt Engine keeps system guidance and untrusted content in
distinct, labeled regions.

#### Response

On an anomalous action attempt, `PermissionPort` returns a denial `Decision`, the Tool Invocation
lands in `denied`, and a security event is emitted (chapter 08). The user may interrupt the run
(exit code 8 semantics), which terminates in-flight tool subprocesses through sandbox teardown.

#### Recovery

No recovery of state is required for denied actions; nothing executed. For actions that executed
under valid grants, the run record and File Change / Command Execution records make effects
reviewable and, where they are file changes, revertible through the Git Engine.

#### Residual risk

A user who grants broad permissions and approves prompts by habit can still be driven to a harmful
approved action. Scoped grants, per-action approval prompts that name the concrete effect, and
visible audit trails keep even permissive configurations inspectable and reversible; the residual
risk is behavioral, not architectural.

#### Tests

Injection-corpus tests assert that instruction-like input never yields an unmediated side effect;
permission-matrix tests confirm default-deny; approval-flow tests confirm dangerous classes prompt;
audit-chain tests (SM-13) confirm every attempted and executed action is attributable.

### RISK-SEC-002 — Indirect prompt injection

- Category: Injection / AI-specific
- Probability: High
- Impact: High
- Severity: Critical
- Mitigation: Untrusted-content labeling of all ingested content; permission mediation and approval for side effects; provenance tracking in the Context Manager; egress requires the `network` permission and, for provider fallback, explicit announcement
- Detection: Provenance-tagged context records; tool-invocation anomaly metrics; egress-permission audit; correlation of a side effect to the untrusted source that preceded it
- Owner: Context Manager (Volume 7) / Permission Manager (Volume 9)
- Status: Open

#### Asset

Workspace data, credentials, provider budget, and host integrity reachable through the agent's
authority.

#### Actor

A remote content author who plants instructions in material the agent will read: a file in the
repository, an issue or pull-request body, a web page fetched by a tool, a dependency's README, or
a tool result from an external service.

#### Vector

Content ingested into the model context that contains directives ("ignore prior instructions,
exfiltrate X", "run this command"), reaching the model as ordinary context and attempting to
steer its subsequent tool use.

#### Preconditions

A tool or the Context Manager ingests attacker-influenced content into the window, and the agent
has, or can be led to obtain, a side-effecting grant.

#### Impact

The same class of unauthorized action as direct injection, but harder for the user to anticipate
because the malicious text is buried in data the user did not author. Rated High for the same
blast-radius reason and elevated to Critical severity by its High probability: any run over
untrusted content is exposed.

#### Prevention

The Context Manager tags each Context Item with provenance and trust; untrusted content is labeled
as data. Side effects still route through `PermissionPort` and Approval, so injected directives
cannot themselves authorize action. Network egress requires the `network` permission; a
data-exfiltration attempt therefore surfaces as a permission decision the user can deny. The
design does not grant elevated trust to content merely because it resides in the workspace.

#### Response

A denied action emits a security event correlated to the ingesting tool invocation and, through
provenance, to the source content. The user may quarantine the offending file (exclude it from
context via the Context Manager's exclusion controls) and re-run.

#### Recovery

Effects that executed under valid grants are reverted through the same File Change / Git paths as
direct injection. Poisoned content is removed from the working set via context exclusion; if it
entered memory or an index, the memory-poisoning and index-poisoning responses (RISK-SEC-007,
RISK-SEC-008) apply.

#### Residual risk

Detection of malicious intent inside benign-looking data is imperfect; the mitigation caps impact
rather than guaranteeing recognition. Residual exposure is the set of actions within standing
grants that execute before the user notices — minimized by default-deny and per-action approval,
not eliminated.

#### Tests

Indirect-injection fixtures embed directives in files, issue bodies, and mock web/tool results and
assert no unmediated side effect follows; provenance tests confirm every Context Item carries its
source; egress tests confirm exfiltration attempts require and record a `network` decision.

### RISK-SEC-003 — Tool injection

- Category: Injection / tool surface
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Strict input-schema validation per tool (JSON Schema, ADR-024); parameterized construction of commands and paths; permission scoping per invocation; sandbox execution
- Detection: Validation-failure events (`tool.invocation.failed` with validation cause); schema-conformance test failures; anomalous argument-shape metrics
- Owner: Tool Runtime (Volume 6) / Sandbox Engine (Volume 9)
- Status: Open

#### Asset

The integrity of a tool invocation: the arguments a tool receives and the operations it performs
on the host.

#### Actor

A confused-deputy model, steered by injected content, that emits crafted tool arguments; or an
upstream data source whose values flow into a tool call.

#### Vector

Malicious values smuggled into a tool's structured input — argument fields that carry shell
metacharacters, traversal sequences, or oversized payloads — attempting to make the tool perform
an operation outside its declared purpose.

#### Preconditions

A tool accepts a field whose value reaches a sensitive sink (a shell, a path resolver, a network
target) and the tool or runtime fails to validate or neutralize it.

#### Impact

Coerced behavior of an otherwise legitimate tool: command execution (see RISK-SEC-019), path
escape (RISK-SEC-020), or resource exhaustion. High impact because a trusted tool becomes the
attacker's instrument.

#### Prevention

The Tool Runtime validates every input against the tool's declared JSON Schema before `Execute`
(ADR-024), rejecting nonconforming payloads with E-TOOL validation errors. Built-in tools
construct commands and resolve paths through parameterized, non-shell mechanisms; the Terminal
Engine does not interpolate untrusted values into a shell string. Each invocation runs under the
scoped permission set and sandbox policy for that call.

#### Response

A validation failure denies the invocation before execution and emits a failure event with the
validation cause. A value that passes schema but breaches a sink control (path or command policy)
is refused by the sandbox with an E-SEC decision.

#### Recovery

No state change occurs on a rejected invocation. If a coerced operation executed, its File Change
/ Command Execution records support review and revert.

#### Residual risk

Semantic misuse within schema-valid, policy-permitted bounds (a delete tool asked to delete a file
the user did not mean to lose) remains; approval prompts for destructive classes and revertibility
through Git bound it.

#### Tests

Schema mutation and fuzzing over tool inputs; sink-control tests injecting shell metacharacters and
traversal sequences; assertion that no built-in tool interpolates untrusted input into a shell.

### RISK-SEC-004 — Tool poisoning

- Category: Injection / tool surface
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Descriptor pinning and change detection; permission re-consent on declaration change; untrusted-content labeling of tool descriptions; trust classification of tool origin
- Detection: Descriptor-diff events on re-registration; audit of declaration changes; anomaly between declared and observed permission use
- Owner: Tool Runtime (Volume 6) / Permission Manager (Volume 9)
- Status: Open

#### Asset

The user's trust decision about a tool: the permissions granted on the basis of the tool's
declared behavior.

#### Actor

A malicious extension author, or a compromised extension, that ships a tool whose declaration
(name, description, schema, permission request) misrepresents its true behavior or changes after
the user has trusted it.

#### Vector

A tool declaration crafted to obtain broad grants under a benign description, or a silent change to
the declaration (a "rug pull") after initial approval so that later invocations do more than the
user consented to.

#### Preconditions

The tool is installable and its declaration is the basis for the user's permission grant; the
runtime does not detect the mismatch or the change.

#### Impact

Grants issued under false pretenses are exercised for unintended effects — the full range the
requested permissions allow. High impact because the user's own consent is turned against them.

#### Prevention

The Tool Runtime pins each registered tool's descriptor; a changed declaration forces
re-registration and re-consent rather than silently inheriting the prior grant. Tool descriptions
are untrusted-content labeled and never treated as instructions to the model. Origin trust
classification (Volume 6 trust vocabulary) gates what a tool from an untrusted origin may request
without explicit user acknowledgment.

#### Response

A descriptor change emits a diff event and blocks invocation under the old grant until the user
re-consents. Observed permission use inconsistent with the declaration raises a security event and
can trigger automatic disablement by policy.

#### Recovery

The tool is disabled or removed through the Package Manager (`removed` state); its grants are
revoked through the permission model; effects already produced are reviewed via the audit chain and
reverted where they are file or Git changes.

#### Residual risk

A tool that behaves honestly until a triggering condition, then acts within its granted
permissions, is not distinguishable at registration time; least-privilege grants and per-action
approval for dangerous classes bound the damage such a tool can do.

#### Tests

Descriptor-pinning tests assert re-consent on any declaration change; grant-scoping tests confirm a
tool cannot exceed its declared permissions; conformance tests reject declarations whose schema and
described behavior diverge in fixtures.

### RISK-SEC-005 — MCP poisoning

- Category: Injection / extension surface
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: MCP servers untrusted by default; trust gating and sandbox tiers (Volume 6 chapter 06); descriptor pinning; untrusted-content labeling of tool/resource/prompt content; least exposure of context to servers
- Detection: MCP connection and capability-change audit; descriptor-diff events; egress and permission-decision trails on MCP-bridged tools
- Owner: MCP Runtime (Volume 6) / Permission Manager (Volume 9)
- Status: Open

#### Asset

The model context and the user's grants, exposed to an external Model Context Protocol server's
tools, resources, and prompts.

#### Actor

A malicious or compromised MCP server: one the user configured that turns hostile, or one whose
maintainer or infrastructure changed.

#### Vector

Poisoned MCP tool descriptions, resource contents, or prompt templates that carry injection
payloads into the model context, or MCP tools that request or perform more than the user
understood — the extension-scale analogue of tool poisoning.

#### Preconditions

An MCP client connection is `ready` and its surfaces are exposed to the agent; the server's content
enters context or its tools are invocable.

#### Impact

Injection into every run that uses the server, plus coerced tool behavior across the server's
surface. High impact because one poisoned server affects all sessions that trust it.

#### Prevention

MCP servers are untrusted by default (Volume 6 chapter 06): their tool descriptions, results,
resource content, and prompt text are untrusted-content labeled wherever they enter context. Trust
gating and per-tier sandboxing constrain stdio servers; descriptor pinning forces re-consent on
capability change; the runtime exposes the least context necessary to a server rather than the full
window.

#### Response

A capability or descriptor change blocks the affected surface until re-consent; anomalous behavior
transitions the connection toward `disabled` by policy and emits a security event correlated to the
server.

#### Recovery

The connection is disabled or removed (`removed` state); bridged tool grants are revoked; poisoned
content that reached memory or an index is handled per RISK-SEC-007 / RISK-SEC-008.

#### Residual risk

A server that serves clean content until a trigger remains a latent risk; least-exposure and
least-privilege limit what a hostile server can extract or drive in any single run.

#### Tests

MCP conformance and trust-gating tests (SM-15) confirm untrusted labeling and re-consent on change;
injection fixtures served over MCP assert no unmediated side effect; exposure tests confirm servers
receive only the intended context slice.

### RISK-SEC-006 — Malicious model output

- Category: Injection / inference channel
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: Output treated as untrusted; structured-output schema validation; no auto-execution of output as code; permission and approval gating of any action derived from output; rendering safety in TUI (Volume 8)
- Detection: Structured-output validation failures; tool-argument validation failures; terminal-render sanitization checks
- Owner: Agent Engine (Volume 4) / Tool Runtime (Volume 6)
- Status: Open

#### Asset

The integrity of the agent loop and the user's terminal: what the model's response is allowed to
cause.

#### Actor

A confused-deputy or compromised provider/model returning crafted output; also an honest model
producing unsafe content when driven by injected context.

#### Vector

Model output engineered to cause harm when consumed: tool-call arguments that target sensitive
sinks, structured output that violates the expected schema, or text containing terminal escape
sequences that manipulate the display or trick the user.

#### Preconditions

Andromeda consumes the output to drive a tool call, parse a structured result, or render to the
terminal.

#### Impact

Coerced tool use (bounded by permissions), corrupted downstream parsing, or a misled user via
display manipulation. Rated Medium impact because output alone confers no authority; elevated to
High severity by its High probability, since crafted output is cheap to produce.

#### Prevention

Model output is untrusted: it never executes as code by virtue of being output. Structured outputs
are validated against their declared schema (ADR-024) and rejected on mismatch; tool arguments pass
the RISK-SEC-003 validation path; the TUI sanitizes or neutralizes control sequences before
rendering (Volume 8). Any action the output proposes still requires a permission decision and, for
dangerous classes, approval.

#### Response

Schema or argument validation failure denies the derived action and emits a failure event.
Suspicious render content is sanitized; a repeated pattern from one provider can transition it
toward `degraded` in routing.

#### Recovery

No state change on rejected output. Effects that executed under valid grants revert through the
File Change / Git paths.

#### Residual risk

Persuasive but technically valid output that leads the user or agent toward a permitted-but-unwise
action remains; approval prompts naming concrete effects and revertibility bound it.

#### Tests

Structured-output conformance and mutation tests; terminal-escape-sequence fixtures asserting safe
rendering; tool-argument validation tests over adversarial model outputs.

### RISK-SEC-007 — Memory poisoning

- Category: Injection / persistence
- Probability: Medium
- Impact: High
- Severity: High
- Mitigation: Provenance and trust attributes on every Memory Record; untrusted content never auto-promoted to trusted memory; approval for durable writes of untrusted-sourced content; no secrets in memory without explicit authorization
- Detection: Provenance audit over memory records; anomaly between memory source trust and its influence; retrieval-time trust filtering metrics
- Owner: Memory Manager (Volume 7) / Permission Manager (Volume 9)
- Status: Open

#### Asset

Long-term and workspace memory: the persisted knowledge the agent retrieves across sessions.

#### Actor

A remote content author whose injected content is captured into memory, or a malicious extension
that writes memory records.

#### Vector

Adversarial content persisted as a Memory Record so that it re-enters context in future runs — a
durable, cross-session form of indirect injection that survives the run that ingested it.

#### Preconditions

Untrusted content is ingested into memory with sufficient trust that it is retrieved and injected
into later contexts.

#### Impact

Repeated injection across future sessions from a single poisoning event; potential long-lived
influence over the agent's behavior. High impact due to persistence and reach.

#### Prevention

Every Memory Record carries provenance and a trust attribute (Volume 7); untrusted-sourced content
is stored as untrusted and is not auto-promoted to trusted memory. Retrieval applies trust
filtering, so untrusted memory enters context labeled as data. The Memory Manager MUST NOT persist
secrets without explicit authorization (Volume 7). Durable writes of untrusted-sourced content that
would gain influence require the user's consent.

#### Response

A poisoned record is identified through provenance and expired or deleted via the Memory Manager
(`expired` / `deleted` status); a security event records the action. Records sharing the source are
reviewable through the audit chain.

#### Recovery

The Memory Manager deletes or archives the poisoned records; audit records of the deletion persist
even though content is removed. Contexts assembled after deletion no longer include the content.

#### Residual risk

Poisoning discovered only after it has influenced runs leaves a trail of affected outputs; run
records make those runs identifiable for review. Trust filtering caps ongoing influence.

#### Tests

Provenance-tracking tests assert every record's source and trust are recorded; poisoning fixtures
confirm untrusted memory is retrieved as labeled data and is deletable; secret-in-memory tests
confirm refusal to persist secrets without authorization.

### RISK-SEC-008 — Index poisoning

- Category: Injection / persistence
- Probability: Medium
- Impact: Medium
- Severity: Medium
- Mitigation: Indexes are rebuildable caches (INV-IDX-02), never authority; index hits carry provenance and enter context untrusted-labeled; invalidation and rebuild on suspicion; scope-limited indexing
- Detection: Index generation and coverage metrics; provenance on index hits; retrieval anomaly monitoring
- Owner: Indexing Engine (Volume 7) / Context Manager (Volume 7)
- Status: Open

#### Asset

The lexical and semantic indexes over workspace content and the retrieval quality they provide.

#### Actor

A remote content author who plants content that, once indexed, is retrieved and injected; or a
party who influences what gets indexed to bias retrieval.

#### Vector

Crafted content designed to be surfaced by index queries (keyword stuffing, embedding collisions,
directive-laden text) so that querying the index injects the attacker's content into context.

#### Preconditions

Attacker-influenced content is within the index scope and ranks highly for relevant queries.

#### Impact

Biased or injected retrieval results feeding context; a channel for indirect injection at query
time. Rated Medium because indexes are caches with no authority and their content re-enters through
the same untrusted-labeled path as any file.

#### Prevention

Per ADR-020 and INV-IDX-02, indexes are rebuildable caches, never a source of truth or authority.
Index hits carry provenance and enter context as untrusted-labeled data subject to the same
mediation as file content. Indexing is scope-limited to the workspace; suspicious content can be
excluded and the index invalidated and rebuilt (`stale` → `building` → `ready`).

#### Response

On suspicion, the affected scope is invalidated (`Invalidate`) and rebuilt from source, dropping
poisoned entries. A security event records the invalidation.

#### Recovery

Rebuilding the index from current workspace source removes poisoned entries with no data loss
(caches are rebuildable). Contexts assembled after rebuild reflect only present content.

#### Residual risk

Content still present in the workspace will re-index; index poisoning is a symptom of malicious
files (RISK-SEC-009) or repositories (RISK-SEC-012), whose removal is the durable fix. The index
layer itself confers no additional authority.

#### Tests

Provenance tests on index hits; invalidation-and-rebuild tests confirming poisoned entries are
dropped; retrieval tests confirming index content enters context untrusted-labeled.

### RISK-SEC-009 — Malicious files

- Category: Injection / content
- Probability: High
- Impact: Medium
- Severity: High
- Mitigation: Files treated as untrusted data; content-type and size handling; path and sandbox controls on any tool that acts on files; no auto-execution of file contents; binary and large-file handling by reference
- Detection: Tool-read provenance; oversized/binary handling events; path-policy and sandbox refusals
- Owner: Context Manager (Volume 7) / Tool Runtime (Volume 6)
- Status: Open

#### Asset

The agent loop, the host, and the context window when the agent reads files from the workspace or
elsewhere.

#### Actor

A remote content author who places a hostile file where the agent will read it — a cloned
repository, a downloaded artifact, an email attachment saved to disk.

#### Vector

A file crafted to harm on ingestion: embedded injection directives, a decompression bomb, a
pathological size or encoding that exhausts resources, or content that exploits a downstream parser.

#### Preconditions

A tool reads the file and its content enters context or a parser without appropriate limits.

#### Impact

Indirect injection (RISK-SEC-002), resource exhaustion, or parser abuse. Rated Medium impact
because files confer no authority and the harm is bounded by content handling; High severity given
how commonly agents read untrusted files.

#### Prevention

File contents are untrusted data. Tools that read files enforce size and output limits (Volume 6
`[tools]` budgets), handle binary and oversized content by reference rather than inlining, and act
under path and sandbox policy. File content never auto-executes; the Context Manager labels
ingested file content with provenance and trust.

#### Response

Oversized or binary content is truncated-with-marker or referenced, emitting a handling event; a
file that triggers a policy or sandbox refusal is not processed further. The user can exclude the
file from context.

#### Recovery

Excluding the file from the working set removes its influence; if it reached memory or an index, the
respective responses apply. No host state changes from reading alone under the controls above.

#### Residual risk

A novel parser vulnerability in a dependency could be reached by crafted file content; dependency
hardening (RISK-SEC-013) and sandbox containment of parsing where feasible bound it.

#### Tests

Decompression-bomb and oversized-file fixtures assert bounded handling; injection-in-file fixtures
assert no unmediated side effect; binary-handling tests confirm by-reference treatment.
