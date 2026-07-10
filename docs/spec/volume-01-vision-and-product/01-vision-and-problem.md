# 01 — Vision and Problem

## Vision

Andromeda is an open-source **AI Engineering Harness**: a local-first runtime in which software
engineers and automations direct AI agents that plan, act through tools, and account for every
action they take. It combines, in a single coherent product:

- an **agent runtime** — the only execution model for intelligent interaction;
- a **planning and execution system** — inspectable plans decomposed into supervised tasks;
- a **tool environment** — schema-typed, permissioned, observable tools as the sole mechanism by
  which agents affect the world;
- a **workflow coordinator** — declared, resumable orchestrations of agents, tools, and human
  approvals, including specification-driven development;
- an **extensible platform** — public contracts for providers, tools, skills, plugins, and the
  other extension surfaces enumerated in [chapter 04](04-goals-non-goals-principles.md);
- a **CLI and TUI client** — a terminal-native interface usable interactively by humans and
  non-interactively by scripts, CI pipelines, and other automations.

Andromeda is **local-first**, **vendor-agnostic**, **model-agnostic**, **tool-first**,
**agent-native**, **observable**, **auditable**, **safe by default**, and **usable by humans and
automations alike**. These nine properties are not slogans; each is stated normatively in
[chapter 04](04-goals-non-goals-principles.md) and bound to measurable targets in
[chapter 06](06-success-metrics.md).

The long-term ambition, recorded as product objectives in chapter 04, is for Andromeda to be one
of the most complete, sound, and extensible open-source AI agent engineering platforms available
— an ambition that is only meaningful because every aspect of it is translated into the
verifiable metrics of chapter 06 rather than asserted as marketing language.

### Product identity

The official product tagline is: **"Your terminal companion for shipping great software."**

The tagline condenses the identity boundary of this chapter: a *companion* (a harness the user
directs, not an autonomous replacement), in the *terminal* (the product surface — CLI and TUI),
for *shipping software* (engineering outcomes, verified and reviewed, not conversation). Brand
identity — palette, typography, mascot, and banner artwork (`docs/brand/`) — is specified as
design tokens in Volume 8 and recorded in the decision register (Volume 0, chapter 06); no
volume other than Volume 8 defines visual identity values.

## The problem

Engineering teams adopting AI agents in 2026 face a set of structural problems that no single
existing product resolves. Andromeda exists because these problems compound each other: fixing
one in isolation (for example, adding a local model to a cloud-first assistant) does not produce
a trustworthy engineering system.

### Problem 1 — Vendor lock-in

Most agentic coding tools are built by, or built around, a single model provider. Their prompts,
capabilities, pricing assumptions, and authentication flows are entangled with that provider.
When the provider changes prices, deprecates an API, restricts third-party access, or degrades a
model, the user has no exit path that preserves their workflows, sessions, memory, and tooling.

**Consequence for users:** switching costs grow with every session; teams cannot arbitrate
between providers on cost, quality, or policy; procurement and compliance teams inherit a
single point of failure.

### Problem 2 — Cloud dependence and privacy exposure

Tools that require a permanent connection to a vendor's cloud cannot be used in air-gapped
environments, on flights, in outages, or in organizations whose policies forbid sending source
code to external services. Even when local models are nominally supported, core functions
(indexing, memory, session persistence) often silently depend on remote services.

**Consequence for users:** entire industries (defense, health, finance, public sector) and
entire situations (offline work, provider outages) are excluded, and privacy-conscious
individuals must choose between capability and confidentiality.

### Problem 3 — Opaque agent behavior

Agents read files, run commands, and modify code. Many tools show the user a summary at best.
The user cannot reliably answer: which files were read, which commands ran, what was sent to
which provider, what it cost, which permissions were exercised, or why a change was made. When
something goes wrong there is no audit trail that survives the session.

**Consequence for users:** trust is unverifiable; incident response is guesswork; regulated
teams cannot adopt agents at all because they cannot evidence what the agent did.

### Problem 4 — Unsafe autonomy

Agent frameworks frequently execute side-effecting actions — deleting files, force-pushing,
installing packages, calling external services — without an explicit, scoped, auditable
permission model. Safety is implemented, if at all, as ad-hoc confirmation prompts that users
learn to click through.

**Consequence for users:** a single hallucinated command can destroy work or exfiltrate
secrets; organizations respond by banning agents rather than governing them.

### Problem 5 — Chat-centric design

Products designed as chatbots retrofit tools, planning, and execution onto a conversation loop.
The result is a split brain: a "chat" runtime with one memory and permission model, and an
"agent" runtime with another. Capabilities behave differently depending on which door the user
entered through.

**Consequence for users:** inconsistent behavior, duplicated configuration, and interaction
histories that cannot be replayed, audited, or automated.

### Problem 6 — Closed extension models

Many tools expose no stable contract for adding a provider, a tool, or a workflow. Extensions,
where they exist, are plugins in name only: undocumented, unversioned, and broken by every
release. The Model Context Protocol (MCP) improved tool interoperability, but a protocol is not
a platform — lifecycle, trust, permissions, and observability of extensions remain unsolved in
most clients.

**Consequence for users:** teams cannot encode their own infrastructure, conventions, and
internal services into the agent's environment, so the agent remains a generic outsider in a
specific codebase.

### Problem 7 — Humans only, or automations only

Interactive assistants are rarely scriptable; automation frameworks are rarely pleasant to
operate interactively. Teams end up running two disjoint stacks — one for a developer at a
terminal, another for CI — with different behavior, different configuration, and different audit
characteristics.

**Consequence for users:** what worked on a laptop fails in CI; agent-driven automation cannot
be developed interactively and then promoted to unattended execution.

## How Andromeda answers the problem

| Problem | Andromeda's structural answer |
|---|---|
| Vendor lock-in | Provider Layer with a common contract; all provider-specific logic confined to adapters; routing and fallback under user policy ([chapter 04](04-goals-non-goals-principles.md), Vendor Agnostic) |
| Cloud dependence | Local-first core: the offline guarantee list in chapter 04 MUST work with no Internet connection when models, tools, indexes, memory, and repositories are local |
| Opaque behavior | Transparent AI: the visibility list of chapter 04 (provider, model, tokens, cost, tools, files, patches, commands, permissions, errors, retries, fallbacks, context and memory state, security events) plus structured events and audit records for every run |
| Unsafe autonomy | Safe by Default: every side-effecting action is bound to the Permission Manager; destructive operations require explicit confirmation unless a previously approved, narrowly delimited policy authorizes them |
| Chat-centric design | Agent Native: one runtime; informal conversation is an agent behavior over the same runtime, memory, context, permissions, and observability |
| Closed extension models | Open Architecture: public contracts for the full extension surface list, plus first-class MCP client support through official, documented protocol versions |
| Humans-or-automations | The CLI and the TUI drive the same Runtime; every interactive capability has a non-interactive equivalent with structured output and deterministic exit codes |

## Differentiation

Andromeda does not differentiate by claiming better models — it ships no models — but by the
properties of the harness around any model:

1. **The offline guarantee is contractual, not incidental.** The specific list of operations
   that work without Internet (chapter 04) is normative and is verified by an offline test
   suite (Volume 13). Competing products treat offline operation as a degraded curiosity.
2. **Capability honesty.** Providers and models declare capabilities; the Runtime detects them
   explicitly, never silently simulates a missing capability, and reports degradation to the
   user. Products that paper over capability gaps produce untrustworthy behavior that differs
   by backend.
3. **Auditability as a first-class output.** Every run produces structured events, cost
   records, and audit records; "what did the agent do?" is a query, not an archaeology project.
4. **Permissioned autonomy.** The permission model (Volume 9) with scoped grants, explicit
   decisions, and an append-only Audit Log is part of the core product, not an enterprise
   add-on.
5. **One runtime for laptop and CI.** Non-interactive operation with structured output is a
   product objective (chapter 04), not an afterthought, so agent behavior can be developed
   interactively and executed unattended without translation.
6. **A platform, not an application.** Providers, tools, skills, plugins, workflows, prompts,
   indexers, storage, authentication, telemetry exporters, Git integrations, commands, and
   policies are all extension points behind versioned public contracts.

## Experience references

The interaction quality Andromeda aspires to is comparable to the strongest terminal-based AI
engineering tools in the field: Codex CLI, Claude Code, OpenCode, Gemini CLI, Aider, Cursor
Agent, Goose, and Cline.

These references are used **only** to identify patterns of interaction quality, speed, clarity,
and ergonomics — for example: streaming responses with visible tool activity, inline diff
review, low-friction permission prompts, resumable sessions, and useful non-interactive modes.
The following rules bind every volume of this specification:

- Authors and implementers MUST NOT copy proprietary interfaces.
- Authors and implementers MUST NOT assume undocumented behavior of these products.
- Andromeda MUST NOT take technical dependencies on competitor products.

## What Andromeda is not

Non-goals are elaborated in [chapter 04](04-goals-non-goals-principles.md); the identity
boundary is stated here because it shapes every subsequent chapter. Andromeda is **not**:

1. A conventional chatbot.
2. A thin wrapper over provider APIs.
3. A full code editor.
4. A traditional IDE.
5. An interface limited to a single provider.
6. A collection of prompts.
7. A system dependent on cloud services to perform its core function.
8. A client that requires a permanent connection.
9. A system that executes destructive actions without permissions.
10. A product that hides from the user the actions it has performed.

Each "not" has a positive counterpart in the product principles: not a chatbot because it is
agent-native; not a wrapper because it is a runtime with planning, execution, memory, and
permissions; not single-provider because it is vendor-agnostic; not cloud-dependent because it
is local-first; not opaque or unsafe because it is transparent, observable, auditable, and safe
by default.
