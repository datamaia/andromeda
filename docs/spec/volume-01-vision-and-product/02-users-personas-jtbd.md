# 02 — Users, Personas, and Jobs to Be Done

## Target users

Andromeda targets people and systems that do engineering work in a terminal and need AI agents
they can direct, constrain, observe, and audit. The primary audiences are:

| Audience | Defining need |
|---|---|
| Individual developers (including open-source maintainers) | Leverage: one person shipping and maintaining more, without surrendering control of their repository or their spend |
| Platform and infrastructure engineers | Encoding organizational context — internal tools, services, conventions — into a governable agent environment |
| Enterprise engineers in regulated or security-sensitive contexts | Adoption that survives a security review: permissions, audit trails, local operation, official integrations only |
| Automation and AI engineers | Agents as unattended components: CI jobs, scheduled maintenance, pipeline steps with structured output and deterministic behavior |
| Privacy-focused and local-model users | Full capability with local inference and zero mandatory network egress |
| Extension authors | Building and distributing providers, tools, skills, plugins, and MCP integrations against stable public contracts |

Secondary audiences — quality engineers, security reviewers, release engineers — interact with
Andromeda through the same surfaces and are served by the same observability, permission, and
traceability guarantees; they are audiences of this specification (Volume 0, chapter 01) as much
as of the product.

Andromeda assumes terminal proficiency. It does not target non-technical end users, and it does
not attempt to hide the terminal behind a graphical shell (see the non-goals in
[chapter 04](04-goals-non-goals-principles.md)).

## Personas

Six personas anchor design decisions across the corpus. They are composites, not customer
records; volumes that make UX or workflow trade-offs SHOULD state which personas the trade-off
serves. Persona names are fictional.

### Persona 1 — Elena, individual open-source developer

- **Role:** Solo maintainer of several open-source libraries; freelance backend developer.
- **Environment:** macOS laptop (Apple Silicon), tmux + shell + editor, GitHub for everything.
  Pays for one cloud model subscription out of pocket; runs a local model for cheap bulk work.
- **Goals:** Triage issues, fix bugs, review contributions, and keep changelogs and releases
  moving across many repositories with a fraction of a full team's time. Keep spend predictable.
- **Frustrations:** Tools that lock her to one provider's pricing; agents that touch files she
  did not sanction; assistants whose sessions evaporate when the terminal closes; costs that are
  invisible until the invoice arrives.
- **How Andromeda serves her:** One harness across all repositories; per-workspace
  configuration and memory; visible token and cost accounting per run; scoped permissions so an
  agent can edit `src/` but never touch `.git/` history or her dotfiles without asking; session
  persistence and resume across days; cheap local-model routing for mechanical tasks.

### Persona 2 — Marcus, platform engineer at a product company

- **Role:** Senior engineer on a 12-person platform team serving ~200 developers.
- **Environment:** Linux workstations and build farm (x86_64 and arm64), internal service mesh,
  monorepo plus satellite repositories, self-hosted CI runners.
- **Goals:** Give every product team a paved road for agent-assisted engineering: standard
  tools, standard permissions, standard observability. Encode internal CLIs, deploy systems,
  and coding conventions as tools and skills the agents actually use.
- **Frustrations:** Assistants that cannot be extended with internal tooling; configuration
  that lives per-user instead of per-repository; no way to enforce organization policy (which
  providers, which permissions, which models) centrally; telemetry that leaves the network.
- **How Andromeda serves him:** The extension surfaces (tools, plugins, skills, policies,
  telemetry exporters) let his team package internal capabilities once and distribute them;
  project-level `andromeda.toml` puts policy in the repository; the Policy Engine and
  Permission Manager enforce organization defaults; structured events feed the team's own
  observability stack through exporters.

### Persona 3 — Priya, security-conscious enterprise engineer

- **Role:** Staff engineer in a financial-services firm with a strict security office.
- **Environment:** Locked-down Linux VDI; some air-gapped enclaves; every third-party tool
  passes a formal review; source code MUST NOT leave approved boundaries.
- **Goals:** Adopt agent-assisted engineering without violating data-handling policy; produce
  evidence for auditors: what ran, what it read, what it changed, who approved it.
- **Frustrations:** Vendors that cannot answer "what exactly leaves the machine?"; tools with
  hidden network calls; integrations that use reverse-engineered endpoints her security office
  would never approve; missing audit trails.
- **How Andromeda serves her:** Local-first operation with local models inside the enclave; the
  offline guarantee list means no hidden egress is required for core work; every integration
  uses official, documented mechanisms only (a provided constraint, see
  [chapter 07](07-constraints-dependencies-risks.md)); the Audit Log and structured events give
  her auditors an answer that is a query, not a promise; permissions are explicit, scoped,
  recorded decisions.

### Persona 4 — Jonas, AI automation builder

- **Role:** Automation engineer building agent-driven pipelines: nightly dependency updates,
  test triage, changelog generation, code-health reports.
- **Environment:** Headless Linux containers in CI (GitHub Actions and self-hosted runners); no
  TTY; everything must be reproducible, parseable, and exit-code honest.
- **Goals:** Run the same agent behavior unattended that he developed interactively; consume
  structured output downstream; retry and resume failed runs; bound cost and time per job.
- **Frustrations:** Interactive-only tools that hang waiting for a confirmation prompt in CI;
  output meant for eyes, not parsers; non-deterministic behavior between his laptop and the
  runner; no machine-readable record of what a run did.
- **How Andromeda serves him:** Non-interactive mode is a product objective, not a flag bolted
  on: JSON output, deterministic exit codes (Volume 0, chapter 03), pre-granted permission
  policies instead of prompts, session persistence for resume, and structured events he can
  archive as the run's record.

### Persona 5 — Amara, local-model privacy advocate

- **Role:** Independent researcher and developer; writes about privacy-preserving AI tooling.
- **Environment:** Linux desktop with a consumer GPU; runs local inference servers (Ollama and
  other OpenAI-compatible local servers); frequently fully offline by choice.
- **Goals:** A complete agent engineering workflow — index, plan, edit, test, commit — with
  zero bytes sent to any third party; honest reporting of what smaller local models can and
  cannot do.
- **Frustrations:** "Local support" that turns out to mean chat only, while indexing, memory,
  or updates phone home; tools that silently simulate missing model capabilities and produce
  garbage; cloud-default configuration that leaks before she can turn it off.
- **How Andromeda serves her:** The offline guarantee list is her daily reality, verified by
  the offline test suite; capability declaration and explicit detection mean Andromeda tells
  her when a local model lacks tool calling instead of faking it; telemetry is consent-based
  and local-first (Volume 10); a local provider is part of the MVP provider seed
  ([chapter 05](05-scope-and-phases.md)).

### Persona 6 — Tomás, extension and tool author

- **Role:** Developer-experience engineer and open-source contributor who builds integrations:
  an internal ticket-system tool, a company MCP server, a skill pack for his framework.
- **Environment:** macOS and Linux; publishes packages publicly and inside his company.
- **Goals:** Build against contracts that do not break every release; test extensions without
  a live provider; reach users through straightforward installation and discovery.
- **Frustrations:** Undocumented plugin APIs; silent breaking changes; no conformance suite to
  test against; extension work that dies with each host-product release.
- **How Andromeda serves him:** Versioned public contracts for every extension surface; the
  Extension SDK with test fixtures and a test provider (Volume 13); the public-contract
  stability metric in [chapter 06](06-success-metrics.md) makes breakage a measured defect, not
  an accepted cost; MCP support through official protocol versions means his MCP server works
  without Andromeda-specific hacks.

## Jobs to Be Done

The personas above hire Andromeda for the following jobs. Use cases in
[chapter 03](03-use-cases.md) trace to these jobs; product objectives in
[chapter 04](04-goals-non-goals-principles.md) exist to make these jobs succeed.

| # | Job statement | Primary personas |
|---|---|---|
| JTBD-1 | When I have a specified change to make, I want an agent to plan, implement, and verify it under my review, so I can ship more without lowering my standards. | Elena, Marcus |
| JTBD-2 | When a build or test fails, I want an agent to localize the fault and propose a reviewed fix, so I can resolve failures in minutes instead of hours. | Elena, Marcus, Jonas |
| JTBD-3 | When an agent proposes changes, I want to review every diff, command, and permission before it takes effect, so I stay accountable for my repository. | Elena, Priya |
| JTBD-4 | When I work in a restricted or offline environment, I want the full core workflow to run locally, so policy compliance does not cost me capability. | Priya, Amara |
| JTBD-5 | When I automate engineering chores, I want agent runs to be scriptable, parseable, and exit-code honest, so I can compose them into pipelines. | Jonas, Marcus |
| JTBD-6 | When I choose or change model providers, I want my workflows, sessions, tools, and memory to survive the switch, so the provider decision stays reversible and price-driven. | Elena, Marcus, Amara |
| JTBD-7 | When an agent has acted on my systems, I want a complete, queryable record of what it did and under which permissions, so I can audit, debug, and attribute every effect. | Priya, Marcus |
| JTBD-8 | When my team has internal tools and conventions, I want to expose them to agents through stable contracts, so agents work our way instead of a generic way. | Marcus, Tomás |
| JTBD-9 | When a session is interrupted — crash, disconnect, shutdown — I want to resume where I left off with state intact, so long-running work is not hostage to fragility. | Elena, Jonas |
| JTBD-10 | When I evaluate cost and model quality, I want per-run token, cost, and outcome visibility, so I can route work to the cheapest model that meets the bar. | Elena, Jonas, Amara |
| JTBD-11 | When I build an extension, I want documented, versioned, testable contracts, so my work survives host upgrades and reaches users predictably. | Tomás |
| JTBD-12 | When I define a repeatable engineering process, I want to encode it as a workflow with states, approvals, and artifacts, so execution is consistent across people, agents, and time. | Marcus, Jonas |

These jobs deliberately span the interactive/unattended boundary and the cloud/local boundary:
a product that satisfies JTBD-1 through JTBD-3 only interactively, or JTBD-4 only for chat,
would fail its personas. That constraint is why "usable by humans and automations" and
"local-first" are product principles rather than features.
