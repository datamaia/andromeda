# 06 — Success Metrics

The ambition stated in [chapter 01](01-vision-and-problem.md) — one of the most complete,
sound, and extensible open-source AI agent engineering platforms — is meaningful only as a set
of verifiable measurements. This chapter defines the product success metrics. Every product
objective in [chapter 04](04-goals-non-goals-principles.md) binds to one or more of these
metrics; the master traceability matrix (Volume 0, chapter 09) carries the binding.

## How to read this chapter

- Metric identifiers `SM-NN` are local to this volume (they are not corpus identifiers; the
  linter-governed namespaces of Volume 0, chapter 03 are unaffected).
- **This volume mints no NFR identifiers.** Each metric names the volume that formalizes it as
  one or more non-functional requirements in that volume's own NFR namespace, with the full
  NFR template (category, minimum threshold, test environment, measurement frequency).
- Targets below are **product-level commitments at v1**, with MVP interim levels where stated.
  Formalizing volumes MAY set stricter targets and MUST NOT set weaker ones without the
  Volume 0 change procedure.
- Where a target says "p95", the percentile is computed over the measurement population
  defined by the measurement method on the reference conditions below.

## Reference conditions

The following product-level reference conditions apply wherever a metric does not state
otherwise. Volume 12 owns the formal test-environment definitions and MUST bind each formalized
NFR to a concrete environment derived from these:

| Reference item | Product-level definition |
|---|---|
| Reference macOS hardware | Apple Silicon laptop, 16 GB unified memory, SSD |
| Reference Linux hardware | 4 vCPU / 8 GB RAM virtual machine (x86_64), SSD-class storage |
| Reference repository | Working tree of approximately 5,000 files / 1,000,000 lines, Git history included |
| Reference network | 50 Mbit/s sustained downstream, 40 ms RTT |
| Offline condition | All network interfaces disabled at the OS level |

## Metrics

| ID | Metric | Definition | Measurement method | Target | Formalized as NFR in |
|---|---|---|---|---|---|
| SM-01 | Provider integration time | Person-hours for a contributor (familiar with the language, new to the codebase) to implement a new provider adapter for a documented HTTP inference API, from SDK template to passing the provider conformance suite | Timed reference-integration exercise, executed at each phase gate (Beta, v1) and sampled from real contribution records (PR open-to-merge effort) | ≤ 16 person-hours (2 person-days) for an OpenAI-compatible-class API | Volume 5 |
| SM-02 | Tool creation time | Person-hours to create a working tool with the Extension SDK: schemas, permission declaration, tests, registered and invocable in a session | Timed exercise against the SDK tool template at each phase gate; verified by SDK tutorial walkthrough in CI | ≤ 4 person-hours | Volume 6 |
| SM-03 | Plugin creation time | Person-hours to build, install, and invoke a plugin exposing one tool over the Andromeda Runtime Protocol, from SDK scaffold to first successful invocation | Timed exercise against the SDK plugin template at each phase gate | ≤ 8 person-hours (1 person-day) | Volume 6 |
| SM-04 | Local-model compatibility | Pass rate of the local-provider conformance suite (agent loop, tool calling, streaming, capability declaration honesty) against models that declare the required capabilities, on ≥ 2 local serving paths (Ollama adapter; generic OpenAI-compatible adapter against a local server) | Automated conformance suite (Volume 13) run per release against pinned local models on reference hardware | ≥ 95% of applicable conformance checks pass on each supported local serving path | Volume 5 |
| SM-05 | Offline operation | Fraction of the offline guarantee list (chapter 04, Local First: 11 operations) that completes under the offline condition with a local provider | Offline test suite: OS-level network disablement, full UC-09 journey plus per-operation checks, per release | 100% of the 11 guaranteed operations; 0 network-access attempts observed | Volume 12 |
| SM-06 | Startup time | (a) Cold start: CLI process start to first byte of output for the version command; (b) TUI start: process start to interactive prompt in an indexed workspace | Automated benchmark harness, 50 iterations per platform per release, p95, reference hardware | (a) ≤ 200 ms p95; (b) ≤ 500 ms p95 | Volume 12 |
| SM-07 | Interaction latency | TUI input-to-render: keystroke or navigation action to completed screen update, excluding model inference time | Instrumented TUI event timestamps under scripted interaction replay, p95, reference hardware | ≤ 50 ms p95 | Volume 12 |
| SM-08 | Streaming latency | Andromeda-added overhead per streamed chunk: provider chunk receipt to rendered output (TUI) or emitted structured output (CLI) | Instrumented timestamps with a mock streaming provider (isolates Andromeda overhead from provider latency), p95 | ≤ 50 ms p95 added overhead | Volume 12 |
| SM-09 | Memory usage | (a) Resident set size of an idle interactive session in the reference repository (excluding any local model server processes); (b) RSS after an 8-hour scripted session | Process accounting in the benchmark harness per release | (a) ≤ 300 MB; (b) ≤ 600 MB | Volume 12 |
| SM-10 | Tool-call reliability | Fraction of tool invocations that terminate within their declared timeout in either a schema-valid result or a structured error envelope (no hangs, crashes, or malformed results) | Runtime metrics over the integration and E2E suites plus fault-injection runs; production figure from local metrics of consenting installations | ≥ 99.5% | Volume 12 |
| SM-11 | Session recovery | (a) Fraction of interrupted sessions (crash-injection at randomized points) that resume with zero loss of persisted turns and intact permission grants; (b) time to restore | Crash-injection suite per release: kill −9, SIGKILL during tool execution, simulated power loss between writes | (a) ≥ 99%; (b) ≤ 5 s p95 | Volume 12 |
| SM-12 | Run reproducibility | Fraction of runs whose persisted run record (configuration snapshot, provider and model identity, prompt references, full tool-invocation sequence with results) is complete and sufficient to replay the recorded decision-and-tool sequence with zero divergence in replay mode | Record-completeness validator over all suite runs; replay-mode divergence test per release | 100% record completeness; 0 divergence on replay of recorded runs | Volume 10 |
| SM-13 | Traceability | Fraction of side-effecting actions (file changes, command executions, Git mutations, tool network access) attributable — via persisted records sharing correlation IDs — to their run, task, tool invocation, and permission decision | Automated audit-chain test: enumerate side effects in instrumented runs, resolve each to its full record chain, per release | 100%; 0 orphan side effects | Volume 10 |
| SM-14 | Test coverage | Line coverage of the product source, with a stricter bar for the Core Domain and public contract packages | Coverage instrumentation in CI, enforced as a merge gate; reported per release | MVP: ≥ 70% overall, ≥ 85% Core Domain and contracts; v1: ≥ 80% overall, ≥ 90% Core Domain and contracts | Volume 13 |
| SM-15 | MCP compatibility | (a) Pass rate of the MCP client conformance suite for each documented MCP protocol version Andromeda declares support for; (b) interoperation success (connect, discover, invoke one tool) against a maintained reference set of ≥ 10 public MCP servers | Conformance suite in CI per release; scheduled interop job against the pinned reference-server set | (a) 100% of applicable checks; (b) ≥ 95% of reference servers | Volume 6 |
| SM-16 | Security | (a) Known vulnerabilities of critical or high severity open in a published release; (b) fraction of side-effecting tool invocations mediated by the Permission Manager; (c) time to first response for coordinated disclosures | (a) Dependency and code scanning (CodeQL, Dependabot, secret scanning) gating releases; (b) enforcement test that attempts unmediated side effects; (c) security-inbox tracking | (a) 0 at publication; (b) 100%; (c) ≤ 3 business days | Volume 9 |
| SM-17 | Portability | Fraction of the acceptance suite passing identically on all Tier 1 platforms (chapter 05), with any platform-specific behavior documented in the platform matrix | Full acceptance suite in CI on every Tier 1 platform per release; divergence audit against the platform matrix | 100% pass on every Tier 1 platform; 0 undocumented behavioral differences | Volume 3 |
| SM-18 | Update time | Wall-clock time for the basic update path: check, download, verify, apply, and report, on the reference network | Automated update test from release N−1 to N per release, p95 | ≤ 60 s p95 end-to-end; ≤ 10 s p95 excluding download transfer | Volume 14 |
| SM-19 | Rollback time | Wall-clock time to restore the previously installed version after an update, using locally retained artifacts (no re-download) | Automated rollback test per release, p95 | ≤ 30 s p95, executable offline | Volume 14 |
| SM-20 | Public-contract stability | Breaking changes to public contracts (provider contract, tool contract, plugin protocol, skill format, workflow format, configuration schema, CLI structured-output schema, event envelope) shipped outside a major release, and deprecation-window compliance | Contract-diff tooling in CI comparing released contract schemas; release audit per Volume 14 | 0 breaking changes outside a major release; every breaking change preceded by a deprecation period of ≥ 1 minor release | Volume 14 |

## Metric governance

1. **Formalization.** Each formalizing volume MUST translate its assigned metrics into NFRs
   using the Volume 0 NFR template, preserving or strengthening the targets above and adding
   minimum thresholds, test environments, and measurement frequency. The consolidation pass
   (Volume 0, chapter 09) verifies that every SM row is covered by at least one NFR.
2. **Phase applicability.** Unless a row states an MVP interim level, targets bind at v1 and
   are measured (without gating) from MVP onward, so trends exist before the gate applies.
   SM-05, SM-13, SM-16(b), and SM-17 bind at MVP exit: offline guarantee, traceability,
   permission mediation, and Tier 1 portability are identity properties, not maturity
   properties.
3. **Measurement honesty.** A metric measured with a weaker method than specified (for
   example, coverage measured only on a subset of packages) counts as unmeasured. Unmeasured
   bound metrics block the phase gates that depend on them (chapter 05).
4. **Change control.** Weakening any target or removing any metric is a MAJOR document change
   under Volume 0, chapter 10.
