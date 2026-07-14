# Andromeda

**Your terminal companion for shipping great software.**

Andromeda is an open-source, local-first, vendor-agnostic AI engineering harness — a
CLI and interactive TUI that runs coding agents against your workspace, with your choice
of model provider and your credentials staying on your machine.

## Install

### macOS (recommended)

```sh
brew install datamaia/tap/andromeda
```

That single command taps the formula, verifies the checksum, and installs `andromeda`
onto your `PATH`. Upgrade later with:

```sh
brew upgrade andromeda
```

### Linux / macOS without Homebrew

Download the release binary for your platform (auto-detects OS and architecture):

```sh
mkdir -p ~/.local/bin && \
curl -fsSL "https://github.com/datamaia/andromeda/releases/download/v0.1.4/andromeda_0.1.4_$(uname -s|tr A-Z a-z)_$(uname -m|sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz" \
  | tar -xz -C ~/.local/bin andromeda && \
andromeda version
```

Make sure `~/.local/bin` is on your `PATH` (or extract into `/usr/local/bin` with `sudo`).
Checksums for every artifact are published alongside the release as `checksums.txt`.

### With Go

```sh
go install github.com/datamaia/andromeda/cmd/andromeda@latest
```

## Quick start

Launch the interactive TUI by running `andromeda` with no arguments:

```sh
andromeda
```

First run walks you through picking a provider and signing in. Or run a one-shot agent
task from the command line:

```sh
andromeda run "add a health-check endpoint" --provider openai-chatgpt --allow-write
```

Grant only the capabilities a task needs with `--allow-write`, `--allow-exec`, and
`--allow-network`; everything is read-only by default.

## Providers

Andromeda is vendor-agnostic. Supported providers include Anthropic, OpenAI, OpenAI via
your ChatGPT subscription (`openai-chatgpt`), Google Gemini, xAI, Groq, Cerebras,
OpenRouter, Hugging Face, and local models via Ollama or vLLM.

Authenticate once and credentials are stored in your OS keychain:

```sh
andromeda auth login openai-chatgpt   # browser OAuth (ChatGPT account)
andromeda auth add anthropic          # store an API key from an env var
andromeda provider check              # validate connectivity
```

## Common commands

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
| `andromeda doctor` | Diagnose your environment |
| `andromeda version` | Print the version |

## License

[Apache License 2.0](LICENSE).
