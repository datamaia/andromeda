# 04 — Goals, Non-Goals, and Product Principles

## Product objectives

The following objectives define what Andromeda must achieve as a product. They are the roots of
the traceability chain (Volume 0, chapter 03): every functional and non-functional requirement
in Volumes 2–15 traces to at least one objective here. Objective identifiers (`PRD-NNN`) are
minted by this volume and never renumbered.

Each objective binds to named success metrics from [chapter 06](06-success-metrics.md)
(`SM-NN`); an objective without a measurable consequence would be decoration, and this
specification does not carry decoration.

### PRD-001 — Agent-Native Engineering Harness

- Phase: Core
- Bound success metrics: SM-10, SM-12, SM-13

Andromeda provides exactly one execution model for intelligent interaction: agents that plan,
act through tools, and iterate under observation. Every product surface — CLI, TUI, workflows,
automations — drives this single runtime; there is no parallel "chat" pathway with different
memory, permissions, or observability.

**Rationale.** A second runtime is a second set of defects, a second security model, and a
second audit story. One runtime makes every capability improvement, every safety control, and
every trace apply everywhere at once (Problem 5, [chapter 01](01-vision-and-problem.md)).

### PRD-002 — Vendor and Model Independence

- Phase: Core
- Bound success metrics: SM-01, SM-04, SM-20

Andromeda works across model providers through a single provider contract, with all
provider-specific logic confined to adapters, and works across models through declared,
explicitly detected capabilities. Users can switch providers and models — between runs or
mid-session — without losing workflows, sessions, memory, tools, or configuration.

**Rationale.** Provider independence is the user's only durable defense against pricing,
policy, and quality changes they do not control (Problem 1). It is also what makes Andromeda's
open-source promise real: a harness usable only with one vendor is that vendor's client.

### PRD-003 — Local-First Operation

- Phase: Core
- Bound success metrics: SM-04, SM-05

Andromeda's core operates fully without Internet connectivity when models, tools, indexes,
memory, and repositories are local. The offline guarantee list in this chapter's Local First
principle is contractual and is verified by an offline test suite.

**Rationale.** Local operation serves the excluded environments and the privacy-conscious users
of [chapter 02](02-users-personas-jtbd.md) (Problem 2), and it disciplines the architecture:
components designed to work offline cannot grow hidden cloud dependencies.

### PRD-004 — Tool-First Execution

- Phase: Core
- Bound success metrics: SM-02, SM-10, SM-15

Agents affect the world only through tools: named, versioned, schema-typed capabilities with
declared permissions, timeouts, limits, and observability. Built-in tools, plugin tools, and
MCP-provided tools are equal citizens of the same contract.

**Rationale.** A single, typed action boundary is what makes permissioning, sandboxing,
tracing, testing, and extension possible at all. Everything Andromeda promises about safety and
audit is enforced at the tool boundary; anything that bypassed it would void those promises.

### PRD-005 — Safe Agent Autonomy

- Phase: Core
- Bound success metrics: SM-16, SM-13

Every side-effecting action is subject to the permission model; destructive operations require
explicit confirmation unless a previously approved policy authorizes a clearly delimited class
of actions. Denial is a first-class, recorded outcome that agents receive as structured input.

**Rationale.** Autonomy without governance gets agents banned (Problem 4). Andromeda's bet is
the opposite of "trust the model": trust the boundary, record the decisions, and let users
grant exactly as much autonomy as their context allows — from ask-every-time to policy-approved
unattended runs.

### PRD-006 — Transparent and Auditable Operation

- Phase: Core
- Bound success metrics: SM-13, SM-12

The user can always determine what Andromeda did on their behalf: provider, model, capabilities,
tokens, cost, tools executed, files read and modified, patches, commands, permission decisions,
results, errors, retries, fallbacks, and model changes — during the run and after it, from
persisted records.

**Rationale.** Verifiable trust beats asserted trust (Problem 3). Transparency is also the
precondition for the audit, compliance, and incident-response jobs (JTBD-7) that gate
enterprise adoption.

### PRD-007 — Extensible Open Platform

- Phase: Core
- Bound success metrics: SM-02, SM-03, SM-15, SM-20

Andromeda exposes versioned public contracts for its extension surfaces — providers, tools,
skills, workflows, prompts, indexers, storage, authentication, telemetry exporters, Git
integrations, commands, TUI panels where viable, and policies — plus first-class MCP client
support, so that third parties can add capability without forking the product.

**Rationale.** No core team can encode every organization's infrastructure and conventions
(Problem 6, JTBD-8, JTBD-11). The platform succeeds when the most valuable capability in a
given deployment was written by its users.

### PRD-008 — First-Class Terminal Experience

- Phase: Core
- Bound success metrics: SM-06, SM-07, SM-08, SM-09

Andromeda's CLI and TUI meet the interaction quality of the strongest terminal AI tools
(references and usage rules in [chapter 01](01-vision-and-problem.md)): streaming output with
visible tool activity, inline diff review, precise permission prompts, responsive input, and
graceful behavior in constrained terminals — within the latency and resource budgets of
chapter 06.

**Rationale.** The harness is used through its interface; an agent platform that is unpleasant
to operate loses to weaker platforms that are pleasant. Interaction quality is stated as
budgets and behaviors, not adjectives, so it can be tested.

### PRD-009 — Human and Automation Parity

- Phase: Core
- Bound success metrics: SM-10, SM-12

Every capability available interactively is available non-interactively: headless operation
without a TTY, structured (JSON) output, deterministic documented exit codes, policy-resolved
permissions with no interactive prompts, and time and cost budgets. Interactive and unattended
runs share the runtime, configuration, and observability.

**Rationale.** Agent behavior is developed interactively and operated unattended (JTBD-5,
UC-07). Two stacks would mean the CI behavior is never quite what was tested at the keyboard —
the exact failure mode Andromeda's personas already suffer.

### PRD-010 — Durable Sessions and Recoverable Work

- Phase: Core
- Bound success metrics: SM-11, SM-12

Sessions, runs, and workflow runs persist their state incrementally and survive interruption:
after a crash, disconnect, or shutdown, the user resumes with history, plan and task states,
permission grants, and accounting intact, and interrupted tasks are reported as interrupted,
never assumed complete.

**Rationale.** Long-running agent work is only viable if fragility is bounded (JTBD-9). Durable
state is also what makes runs auditable after the fact and reproducible enough to debug.

### PRD-011 — Portable Unix-First Platform

- Phase: Core
- Bound success metrics: SM-17, SM-18, SM-19

Andromeda ships as a single installable executable (`andromeda`) for macOS and Linux first,
with Unix as the reference behavior, all OS-specific behavior encapsulated in the Platform
Abstraction Layer, and an architecture that keeps native Windows support implementable in a
later phase without redesign.

**Rationale.** The platform scope ([chapter 05](05-scope-and-phases.md)) matches where the
target users work today; the PAL discipline is what keeps the second wave of platforms a port,
not a rewrite.

### PRD-012 — Specification-Driven Workflows

- Phase: Core
- Bound success metrics: SM-12, SM-13

Andromeda coordinates multi-step engineering processes as declared workflows — states,
transitions, agents, tools, permissions, approvals, artifacts — including a
specification-driven development workflow from intake through implementation, validation, and
review. Workflows are resumable and their histories are complete.

**Rationale.** Serious engineering is a process, not a prompt (JTBD-12). Declared workflows
turn team process into something executable, repeatable, and auditable, and they are where
human approval gates naturally live.

### PRD-013 — Sustainable Open-Source Delivery

- Phase: Core
- Bound success metrics: SM-14, SM-18, SM-19, SM-20

Andromeda is developed openly on GitHub in English, with requirement-to-release traceability,
tested and signed releases (signing subject to the viability conditions in chapter 05),
documented governance, and public contracts whose stability is measured and defended.

**Rationale.** The product's promises — vendor independence, local operation, transparency —
are only credible if the project itself is inspectable, verifiable, and not hostage to a single
maintainer or vendor (see the bus-factor risk in
[chapter 07](07-constraints-dependencies-risks.md)).

## Non-goals

Non-goals bound the product identity. They are restated as phased scope exclusions in
[chapter 05](05-scope-and-phases.md); re-scoping any of them requires the Volume 0 change
procedure.

1. **Not a conventional chatbot.** Andromeda does not ship a conversation product; informal
   conversation exists only as agent behavior on the single runtime (PRD-001).
2. **Not a thin API wrapper.** Forwarding prompts to a provider without planning, tools,
   permissions, memory, and observability is not an Andromeda mode.
3. **Not a full code editor or IDE.** Andromeda reads, analyzes, and modifies code through
   tools and presents diffs for review; it does not aim to provide interactive editor buffers,
   language-server-driven editing, or an IDE's window management. It complements editors
   rather than replacing them.
4. **Not a single-provider interface.** No feature may exist that structurally works with only
   one provider (PRD-002); provider-exclusive convenience belongs in adapters.
5. **Not a prompt collection.** Prompts are versioned internal assets of the Prompt Engine and
   skills; a library of copy-paste prompts is not the product.
6. **Not cloud-dependent.** No hosted Andromeda service is required for any core capability,
   and this corpus defines no hosted service (PRD-003).
7. **Not connection-requiring.** Permanent connectivity is never assumed; connectivity loss is
   a handled state, not a failure of the product (PRD-003, UC-09, UC-10).
8. **Not autonomously destructive.** Andromeda MUST NOT execute destructive actions without
   permissions — under any configuration, in any mode (PRD-005).
9. **Not opaque.** Andromeda MUST NOT hide performed actions from the user; there is no
   "quiet" mode that suppresses the record — quiet modes reduce presentation, never recording
   (PRD-006).
10. **Not a model host or trainer.** Andromeda ships no models, trains no models, and manages
    no model weights; it connects to providers (including local inference servers) through
    adapters.
11. **Not a graphical application.** No desktop GUI or web UI client is defined in this corpus
    (see Out of Scope, chapter 05); the terminal is the product surface.

## Product principles

Nine principles govern every design decision in Volumes 2–15. They are normative: each
principle's statements bind implementers, and volumes that specialize them (noted per
principle) may strengthen but MUST NOT weaken them. Where a principle conflicts with another
requirement, the precedence order of Volume 0, chapter 01 applies.

### Principle 1 — Vendor Agnostic

- Andromeda MUST NOT depend on a single provider.
- Every provider MUST be implemented as an adapter behind the common provider contract
  (Volume 5).
- The Core Domain and the Runtime MUST NOT contain provider-specific logic — no OpenAI-,
  Anthropic-, Google-, xAI-, Mistral-, or any other provider-specific branches outside that
  provider's adapter.
- Provider-specific code MUST remain inside its adapter, with the sole exception of shared
  types defined in the public contracts.
- Provider selection, routing, and fallback MUST be user- and policy-controlled, and every
  provider or model change MUST be reported to the user and recorded (see Principle 7).

### Principle 2 — Model Agnostic

- Andromeda MUST offer a coherent experience across models, and it MUST NOT claim that all
  models have identical capabilities.
- Every provider and every model MUST declare its capabilities using the capability vocabulary
  owned by Volume 5.
- The Runtime MUST use explicit capability detection — behavior keys off declared capabilities,
  never off model or vendor names.
- A capability that is not available:
  - MUST NOT be silently simulated;
  - MUST be reported to the user as unavailable;
  - MAY be degraded through a documented degradation strategy;
  - MUST produce a precise error when it is mandatory for a workflow.

### Principle 3 — Local First

- Andromeda MUST be able to operate with local models.
- The product core MUST function without an Internet connection when the session uses local
  models, local tools, local indexes, local memory, local repositories, and local
  documentation.
- The absence of Internet connectivity MUST NOT prevent any of the following (the **offline
  guarantee list**, verified per Volume 13):
  1. Opening a workspace.
  2. Querying local memory.
  3. Indexing files.
  4. Executing local tools.
  5. Using local Git.
  6. Using the terminal.
  7. Executing workflows whose steps are offline-capable.
  8. Creating patches.
  9. Reviewing diffs.
  10. Running tests.
  11. Viewing logs.

### Principle 4 — Tool First

- Agents MUST act through tools; tools are first-class citizens of the product.
- Every tool — built-in, plugin-provided, or MCP-provided — MUST declare: identity,
  description, input schema, output schema, permission declaration, timeouts, lifecycle,
  versioning, error policy, observability, tests, limits, compatibility declaration, origin,
  and trust level (full contract in Volume 6).
- Tool citizenship is uniform: third-party tools MUST be subject to the same contract,
  permission model, sandboxing, and observability as built-in tools, with origin and trust
  level always visible.

### Principle 5 — Agent Native

- Every intelligent interaction MUST execute through agents on the single Runtime.
- A separate runtime named or shaped as "chat" MUST NOT exist.
- Informal conversation MAY exist as a behavior of an agent, and when it does it MUST use the
  same runtime, memory, context, permissions, traceability, and observability as every other
  agent behavior.

### Principle 6 — Open Architecture

- Every extensible capability SHOULD be implemented through public contracts (deviations
  require a documented decision per Volume 0).
- The system MUST allow extension of: providers, tools, skills, workflows, prompts, indexers,
  storage, authentication, telemetry exporters, Git integrations, commands, TUI panels (where
  viable), and policies.
- Public contracts MUST be versioned, and their stability is measured by SM-20.

### Principle 7 — Transparent AI

- The user MUST be able to know, for any session and run: the provider; the model; the active
  capabilities; tokens consumed; estimated cost and actual cost when determinable; the tools
  executed and their relevant parameters; files read; files modified; patches; commands
  executed; permissions requested, granted, and denied; results; errors; retries; fallbacks;
  model changes; context state; memory state; and security events.
- Andromeda MUST NOT promise to show a model's private internal reasoning.
- Andromeda MAY show: operational plans; reasoning summaries officially provided by the
  provider; generated justifications for decisions; executed steps; evidence; tool results;
  runtime traces; and reasoning token counts when the provider exposes them officially.
- Provider and model changes (manual, routing, or fallback) MUST be announced to the user and
  recorded — never silent (see UC-10).

### Principle 8 — Safe by Default

- Actions with side effects MUST be subject to permissions evaluated by the Permission Manager.
- Destructive operations MUST require explicit confirmation, unless a previously approved
  policy authorizes a clearly delimited class of actions — and such policies are themselves
  recorded, scoped, and revocable (Volume 9).
- Default configuration MUST err toward asking: broadening autonomy is always an explicit user
  or policy decision.

### Principle 9 — Observable by Default

- Every run MUST generate structured events (envelope per Volume 10).
- Every tool MUST be traceable: invocations, inputs summary, outcomes, and timing correlate to
  their run, task, and permission context.
- Every error MUST carry: a stable code, a category, a message, safe context, a cause, its
  recoverability, a recommended action, and a correlation ID (error scheme per Volume 0,
  chapter 03).
- Observability MUST be local-first: records exist and are queryable locally; export is
  optional and consent-based (Volume 10).

## Principles in tension

Principles collide in practice; the precedence order (Volume 0, chapter 01) resolves the
collisions. Three recurring cases are settled here so volumes do not re-litigate them:

1. **Transparency vs. terseness (Principle 7 vs. PRD-008).** Quiet and non-interactive modes
   reduce what is *presented*, never what is *recorded*. The full record remains queryable.
2. **Safety vs. automation (Principle 8 vs. PRD-009).** Unattended runs never bypass
   permissions; they resolve them from pre-approved policy, and anything not pre-granted is
   denied and recorded. A hung prompt in CI is a defect; so is a silent grant.
3. **Local-first vs. capability (Principle 3 vs. Principle 2).** When local models lack
   capabilities, Andromeda reports the gap and degrades per documented strategy; it neither
   blocks local use behind cloud features nor fakes cloud-grade capability locally.
