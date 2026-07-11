# Volume 6 — Tools, MCP, Skills, and Plugins

**Status:** Authored (draft) · **Owner:** Tool Runtime / extension surfaces (Volume 6)

Volume 6 specifies Andromeda's action and extension surfaces: the tool contract and Tool SDK,
the tool lifecycle with permissions and trust, the built-in tool catalog, the Tool Invocation
state machine, MCP client support, the skill format and system, the plugin runtime over the
Andromeda Runtime Protocol, and the package manager with its supply-chain rules. Per Volume 0
chapter 03, this volume mints all `TOOL`, `SDK`, `MCP`, `SKILL`, and `PLUG` identifiers, owns
the behavioral contracts of `ToolPort`, `TerminalPort`, and `PackagePort` (Volume 3, chapter
02), and owns the full machines for Tool Invocation, Plugin, MCP Client Connection, and
Package installation (Volume 2, chapter 09). Keystones defined here: FR-TOOL-001 (tool
contract), FR-SDK-001 (extension SDK), FR-MCP-001 (MCP client support), FR-SKILL-001 (skill
format), FR-PLUG-001 (plugin runtime).

Foundations assumed: Volume 0 (conventions, glossary), Volume 1 (objectives, principles,
phases, MVP minimum), Volume 2 (entities and frozen states), Volume 3 (ports, components,
PAL), Volume 9 by reference (permission model, sandbox model, trust vocabulary).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — Tool SDK and Contract](01-tool-sdk-and-contract.md) | Keystones FR-TOOL-001 and FR-SDK-001; the full tool declaration, ToolPort behavioral contract, ToolEvent stream union, Extension SDK tool kit, SM-02 formalization |
| [02 — Tool Lifecycle, Permissions, and Trust](02-tool-lifecycle-permissions-trust.md) | Registration lifecycle, the fixed invocation pipeline, permission declaration and mediation, origin/trust rules, `[tools]` configuration keys, the E-TOOL error catalog |
| [03 — Built-in Tools Catalog](03-builtin-tools-catalog.md) | The 20 built-in tools with purpose, schema sketch, permissions, and phase — filesystem (read/write/search/replace/patch/diff), git, terminal, process, http, browser, docker, kubernetes, sqlite, github, gitlab, jira, notion, slack, linear |
| [04 — Tool Invocation State Machine](04-tool-invocation-state-machine.md) | The full machine over the frozen states, all twelve mandatory elements, tool-boundary events, FR-TOOL-008 |
| [05 — MCP Client and Runtime](05-mcp-client-and-runtime.md) | Keystone FR-MCP-001; connections, transports, discovery, registration, tools/resources/prompts, lifecycle, health, update, versioning |
| [06 — MCP Security and Conformance](06-mcp-security-and-conformance.md) | MCP trust model, isolation, supply-chain rules, conformance testing (SM-15) |
| [07 — Skill Format and System](07-skill-format-and-system.md) | Keystone FR-SKILL-001; manifest, versioning, composition, testing, publication, trust |
| [08 — Plugin Runtime and ARP](08-plugin-runtime-and-arp.md) | Keystone FR-PLUG-001; the Andromeda Runtime Protocol: framing, handshake, method surface, permissions, sandbox, lifecycle |
| [09 — Package Manager and Supply Chain](09-package-manager-supply-chain.md) | Install/discover/register/update/uninstall, dependencies, versioning, signatures, future marketplace |
| [10 — State Machines](10-state-machines.md) | Full machines: Plugin, MCP Client Connection, Package installation |
| [99 — Volume Register](99-volume-register.md) | Everything this volume minted (merged from authoring fragments at consolidation) |

## Reading guide

1. Chapter 01 is the contract hub: every tool, from every origin, is the chapter 01
   declaration plus the chapter 02 pipeline — chapters 05 and 08 only define how MCP and
   plugin tools arrive at that same boundary.
2. Chapter 03 fixes the built-in vocabulary agents and prompts rely on; phases there follow
   ADR-074, and only the eight MVP tools are commitments at MVP exit.
3. Chapters 04 and 10 own this volume's four state machines using the frozen names of
   Volume 2 chapter 09; no other chapter restates transitions.
4. Permission names, scopes, decisions, trust vocabulary, and sandbox policy content used
   throughout are Volume 9's; configuration schema and event envelope are Volume 10's.
