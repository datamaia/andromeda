# Volume 8 — CLI and TUI

**Status:** Complete · **Owner:** CLI / TUI (Volume 8)

Volume 8 specifies Andromeda's product surface: the `andromeda` command-line interface —
architecture, grammar, conventions, and every command — and the interactive terminal user
interface — architecture, theming over the ADR-026 design tokens, wireframes, interaction
patterns, and accessibility/compatibility rules. Per Volume 0 chapter 03, this volume mints
all `CLI`, `TUI`, and `UX` identifiers, is the single home for CLI command names and grammar
and for design tokens and theming, and owns the `[cli]`, `[tui]`, and `[tui.theme]`
configuration tables. Keystones defined here: FR-CLI-001 (CLI grammar, chapter 01) and
FR-TUI-001 (TUI shell, chapter 07).

Foundations assumed: Volume 0 (conventions, glossary, exit codes), Volume 1 (objectives,
principles, phases, MVP minimum, success metrics), Volume 2 (entities and frozen state
vocabularies, rendered verbatim by both drivers), Volume 3 (driver position, ports, Runtime
API, PAL). Referenced by name where owned elsewhere: permission model and sandbox
(Volume 9), configuration schema and event envelope (Volume 10), Git semantics (Volume 11),
performance budgets (Volume 12).

## Chapters

| Chapter | Contents |
|---|---|
| [01 — CLI Architecture](01-cli-architecture.md) | Keystone FR-CLI-001; driver position, execution pipeline, the complete command tree, grammar rules, runtime mediation and parity, root/TUI hand-off, extension commands under `x` |
| [02 — CLI Conventions](02-cli-conventions.md) | Global flags, exit-code application, human + `--json` output contract (envelope, NDJSON), stream discipline, quiet/verbose/debug, non-interactive and CI modes, confirmations, `ANDROMEDA_*` environment variables, shell completion, error presentation, paging, progress, `[cli]` keys, `cli.*` events, E-CLI catalog |
| [03 — CLI Commands: Core](03-cli-commands-core.md) | Root, `run`, `plan`, `exec`, `init`, `session`, `config`, `auth` — syntax, flags, defaults, examples, errors, exit codes, JSON schemas, permissions |
| [04 — CLI Commands: Platform](04-cli-commands-platform.md) | `provider`, `model`, `tool`, `plugin`, `skill`, `workflow`, `mcp` — same specification shape |
| [05 — CLI Commands: Data](05-cli-commands-data.md) | `memory`, `context`, `index`, `git`, `logs`, `trace`, `export` — same specification shape; the offline and transparency surface |
| [06 — CLI Commands: Maintenance](06-cli-commands-maintenance.md) | `doctor`, `update`, `version`, `completion`; the post-command update notice |
| [07 — TUI Architecture](07-tui-architecture.md) | Keystone FR-TUI-001; Bubble Tea v2 model, panel system, navigation, focus, keyboard/mouse, resize, small-terminal behavior |
| [08 — Theming and Design Tokens](08-theming-and-design-tokens.md) | ADR-026 token-to-ANSI mapping, truecolor/256/16/no-color tiers, light-terminal fallback, the fixed Danger red with contrast ratios, `[tui.theme]` keys |
| [09 — Wireframes: Core](09-wireframes-core.md) | ASCII wireframes with prose: start/splash, workspace selection, session, plan, execution, tool call, permission prompt, diff, git, files, context, memory, logs, costs/tokens |
| [10 — Wireframes: Platform](10-wireframes-platform.md) | Provider, model, configuration, plugins, skills, workflows, MCP, errors, help, command palette, quick actions, update, recovery |
| [11 — Interaction Patterns](11-interaction-patterns.md) | Streaming rendering, spinners, progress, modals, confirmations, toasts, empty/loading/offline/degraded states, copy/paste, search, filtering, pagination, virtualization |
| [12 — Accessibility and Compatibility](12-accessibility-and-compatibility.md) | Accessibility, no-color mode, Unicode fallback, screen readers where viable, SSH, CI, terminal compatibility matrix |
| [99 — Volume Register](99-volume-register.md) | Everything this volume minted (merged from authoring fragments at consolidation) |

## Reading guide

1. Chapters 01 and 02 are the CLI's contract hub: every command chapter (03–06) states only
   its deltas from them, and extension commands inherit them mechanically (ADR-104).
2. The command tree in chapter 01 is the single grammar oracle; chapters 03–06 specify every
   node it declares, and nothing else.
3. Chapter 07 is the TUI's structural contract; chapters 09–11 render against it, and
   chapter 08's token mapping is consumed by both the TUI tiers and the CLI's 16-color-safe
   styling subset (ADR-103).
4. Both drivers render frozen state vocabularies (Volume 2, chapter 09) verbatim and
   implement no business logic: everything executes through the Runtime API and ports
   (FR-CLI-002; Volume 3, chapter 06).
