# Andromeda

**Your terminal companion for shipping great software.**

[![Website](https://img.shields.io/badge/web-andromedacli.com-7C5CFF)](https://andromedacli.com)
[![Docs](https://img.shields.io/badge/docs-andromedacli.com%2Fdocs-7C5CFF)](https://andromedacli.com/docs)
[![CI](https://github.com/datamaia/andromeda/actions/workflows/ci.yml/badge.svg)](https://github.com/datamaia/andromeda/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/datamaia/andromeda?sort=semver)](https://github.com/datamaia/andromeda/releases)
[![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go&logoColor=white)](go.mod)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

Andromeda is an open-source, **local-first, vendor-agnostic** AI engineering harness — a
CLI and interactive TUI that runs coding agents against your workspace. Bring your own model
provider; your credentials stay in your OS keychain and never leave your machine.

## Why Andromeda

- **Local-first & private** — no andromeda cloud, no telemetry. Credentials live in your OS
  keychain (Keychain / Secret Service / Windows Credential Manager).
- **Vendor-agnostic** — one interface over Anthropic, OpenAI (API key *or* ChatGPT
  subscription), Gemini, xAI, Groq, Cerebras, OpenRouter, Hugging Face, and local models
  (Ollama, vLLM). Switch providers with a flag.
- **Safe by default** — every side-effecting action (write, run command, network, git
  mutation) is read-only until you grant it. Approve interactively, or pre-approve vetted
  commands with a [`[permission]` allowlist](#configuration).
- **CLI *and* TUI** — script one-shot agent runs, or drop into a full-screen terminal UI with
  slash commands, `@`-file mentions, a plan/approve flow, and mouse-scrollable history.
- **Persistent sessions** — conversations are remembered across turns and can be resumed with
  `--continue` / `--resume`.
- **`AGENTS.md`-aware** — project guidance in `AGENTS.md` is folded into the agent's context on
  every run. Scaffold it (and `andromeda.toml`) with `/init`.

## Install

### macOS (recommended)

```sh
brew install datamaia/tap/andromeda
```

Homebrew taps the cask, verifies the checksum, and links `andromeda` onto your `PATH`.
Upgrade later with `brew upgrade andromeda`.

> **Untrusted tap?** Recent Homebrew versions ask you to trust a third-party tap before
> installing its casks. If you see `Refusing to load cask … from untrusted tap`, run
> `brew trust datamaia/tap` once, then re-run the install. (The `curl` script below needs no
> trust step.)

### Linux / macOS (script)

```sh
curl -fsSL https://andromedacli.com/install | bash
```

Detects your OS/architecture, downloads the matching release binary, and installs it to
`/usr/local/bin` (or `~/.local/bin`). Override with `ANDROMEDA_VERSION` or `ANDROMEDA_INSTALL_DIR`.

### Windows (PowerShell)

```powershell
irm https://andromedacli.com/install.ps1 | iex
```

Downloads the release archive, verifies its SHA256 against the release checksums, installs
`andromeda.exe` to `%LOCALAPPDATA%\Programs\andromeda`, and adds it to your user `PATH`.

### With Go

```sh
go install github.com/datamaia/andromeda/cmd/andromeda@latest
```

Prebuilt binaries and checksums for every release are on the
[releases page](https://github.com/datamaia/andromeda/releases).

### Windows Defender / SmartScreen

The Windows build isn't code-signed yet, so Microsoft Defender may flag `andromeda.exe` — usually
SmartScreen's "unknown publisher" prompt, or the ASR rule **"Use advanced protection against
ransomware"** blocking it. This is a **false positive**: as a coding agent, Andromeda legitimately
reads and writes many files in your workspace, which trips the anti-ransomware heuristic. It never
encrypts, exfiltrates, or touches files outside the workspace you point it at, and every release is
SHA256-checksummed (`checksums.txt`) and cosign-signed.

**On your own PC** (you're the admin) — allow it once in an elevated PowerShell:

```powershell
$exe = "$env:LOCALAPPDATA\Programs\andromeda\andromeda.exe"
Add-MpPreference -AttackSurfaceReductionOnlyExclusions $exe        # the ASR ransomware rule
Add-MpPreference -ControlledFolderAccessAllowedApplications $exe   # if Controlled Folder Access is on
```

…or via **Windows Security → Virus & threat protection → Ransomware protection → Allow an app
through Controlled folder access**.

**On a managed (company) PC** — ASR rules are enforced by IT and you can't self-exclude. Send this
to your IT/security team:

> Please add an exclusion from the Microsoft Defender ASR rule **"Use advanced protection against
> ransomware"** (GUID `c1db55ab-c21a-4637-bb3f-a12568109d35`) for **Andromeda CLI**, an open-source
> (Apache-2.0) developer tool — https://github.com/datamaia/andromeda
> - **Executable:** `C:\Users\<user>\AppData\Local\Programs\andromeda\andromeda.exe`
> - **Why:** it's a coding agent that reads/writes files in the developer's working repo; that
>   legitimate multi-file write pattern trips the ransomware heuristic (false positive). It doesn't
>   encrypt or exfiltrate anything; releases are SHA256-checksummed and cosign-signed.
> - **Suggested exclusion** (Intune / Defender for Endpoint → *ASR only exclusions*, or admin
>   PowerShell): `Add-MpPreference -AttackSurfaceReductionOnlyExclusions "C:\Users\<user>\AppData\Local\Programs\andromeda\andromeda.exe"`

Authenticode code signing — which removes the warning at the source — is wired and ready to activate
(the `binary_signs` block in `.goreleaser.yaml`); it turns on once a signing certificate is
provisioned.

## Quick start

Launch the interactive TUI by running `andromeda` with no arguments:

```sh
andromeda
```

The first run walks you through picking a provider and signing in. Or run a one-shot agent task
from the command line:

```sh
andromeda run "add a health-check endpoint" --provider openai-chatgpt --allow-write
```

Capabilities are opt-in per run — `--allow-write`, `--allow-exec`, `--allow-network`. Without
them the agent is read-only.

## Documentation

Full documentation lives at **[andromedacli.com/docs](https://andromedacli.com/docs)** — install,
providers & authentication, the engineering harness, working with agents, configuration, and the
command reference. A printable getting-started guide is at
**[andromedacli.com/guide](https://andromedacli.com/guide)**.

## Configuration

`andromeda.toml` configures a project (run `/init` in the TUI to scaffold it, `AGENTS.md`, and
the `.agents/` capability dirs). It is plain TOML, layered global → workspace → project → env → flags:

```toml
[provider]
default = "openai-chatgpt"

# Commands the agent may run WITHOUT an approval prompt, matched by argv prefix.
# Anything not listed still asks; entries in `deny` are always refused.
[permission]
allow = ["git status", "git diff", "go build ./...", "go test ./..."]
deny  = ["git push --force", "rm -rf"]
```

Sign in once and credentials are stored in your OS keychain:

```sh
andromeda auth login openai-chatgpt   # browser OAuth (ChatGPT account)
andromeda auth add anthropic          # store an API key from an env var
andromeda provider check              # validate connectivity
```

## Commands

| Command | Description |
| --- | --- |
| `andromeda` | Launch the interactive TUI (default) |
| `andromeda run <goal>` | Run an agent to accomplish a goal in the workspace |
| `andromeda --continue` | Reopen the most recent session |
| `andromeda --resume <id>` | Resume a specific saved session |
| `andromeda sessions list` | List saved TUI sessions |
| `andromeda provider list` | List supported model providers |
| `andromeda model list` | List models the configured provider exposes |
| `andromeda memory add <text>` | Add a workspace memory record |
| `andromeda ontology build` | Write a deterministic structural map of the repo (`.andromeda/ontology/project.ttl`) |
| `andromeda graph serve` | Build the workspace graph and open an interactive viewer on localhost |
| `andromeda doctor` | Diagnose your environment |
| `andromeda version` | Print the version |

## Development

Requires Go 1.25+. Common tasks:

```sh
go build ./...      # build
go test ./...       # test
make ci             # full local gate (fmt, vet, lint, build, test, coverage)
```

The codebase follows a hexagonal architecture: `internal/tui` and provider adapters depend only
on ports (`internal/ports`), with wiring in `cmd/andromeda`.

## License

[Apache License 2.0](LICENSE).
