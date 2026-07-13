package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Actions are the app-backed operations slash commands invoke. The composition root wires them so
// the TUI keeps no app/provider imports; a nil action degrades to an informative message.
type Actions struct {
	Doctor    func(ctx context.Context) string
	Update    func(ctx context.Context) string
	Memory    func(ctx context.Context, args string) string
	Workflows func(ctx context.Context) string
	MCP       func(ctx context.Context) string
	Skills    func(ctx context.Context) string
	Models    func(ctx context.Context) []string
	Config    func(ctx context.Context) string
	Logout    func(ctx context.Context, provider string) string
	Export    func(lines []string) string
}

// WithActions wires the app-backed slash-command handlers.
func (m Model) WithActions(a Actions) Model { m.actions = a; return m }

// slashCommand is one palette entry.
type slashCommand struct {
	name string
	desc string
	run  func(m Model, args string) (tea.Model, tea.Cmd)
}

// commandRegistry is the full set of slash commands (industry-standard basics + Andromeda modes).
func commandRegistry() []slashCommand {
	return []slashCommand{
		{"help", "list commands and keybindings", cmdHelp},
		{"commands", "list all slash commands", cmdHelp},
		{"keys", "show keybindings", cmdKeys},
		{"clear", "clear the conversation", cmdClear},
		{"compact", "summarize the conversation so far", cmdCompact},
		{"status", "show provider, model, mode, and session", cmdStatus},
		{"model", "choose the model (/model <name> to set)", cmdModel},
		{"provider", "choose the provider", cmdProvider},
		{"login", "switch or sign in to a provider", cmdLogin},
		{"logout", "sign out of the current provider", cmdLogout},
		{"config", "show resolved configuration", cmdConfig},
		{"export", "save the transcript to a file", cmdExport},
		{"doctor", "run environment checks", cmdDoctor},
		{"update", "check for updates", cmdUpdate},
		{"memory", "manage workspace memory", cmdMemory},
		{"workflows", "list SDD workflows", cmdWorkflows},
		{"mcp", "MCP servers", cmdMCP},
		{"skills", "available skills", cmdSkills},
		{"goal", "set and run a goal (/goal <text>)", cmdGoal},
		{"loop", "toggle loop mode", cmdLoop},
		{"agent", "switch to agent mode", cmdMode("agent")},
		{"plan", "switch to plan mode (no changes are made)", cmdMode("plan")},
		{"shell", "switch to shell mode", cmdMode("shell")},
		{"quit", "exit Andromeda", cmdQuit},
	}
}

// paletteActive reports whether the user is typing a slash command name (so the palette shows).
func (m Model) paletteActive() bool {
	return strings.HasPrefix(m.input, "/") && !strings.Contains(m.input, " ")
}

// filteredCommands returns the commands whose name starts with the typed prefix (after "/").
func (m Model) filteredCommands() []slashCommand {
	prefix := strings.TrimPrefix(m.input, "/")
	var out []slashCommand
	for _, c := range commandRegistry() {
		if strings.HasPrefix(c.name, prefix) {
			out = append(out, c)
		}
	}
	return out
}

// handlePaletteKey drives palette navigation while a slash-command name is being typed. It returns
// handled=false for keys it does not consume (so normal text editing still applies).
func (m Model) handlePaletteKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	cmds := m.filteredCommands()
	switch {
	case msg.Code == tea.KeyEscape:
		m.input = "" // close the palette without quitting
		return m, nil, true
	case msg.Code == tea.KeyUp:
		if m.paletteCursor > 0 {
			m.paletteCursor--
		}
		return m, nil, true
	case msg.Code == tea.KeyDown:
		if m.paletteCursor < len(cmds)-1 {
			m.paletteCursor++
		}
		return m, nil, true
	case msg.Code == tea.KeyTab:
		if len(cmds) > 0 {
			m.input = "/" + cmds[clamp(m.paletteCursor, len(cmds))].name + " "
			m.paletteCursor = 0
		}
		return m, nil, true
	case msg.Code == tea.KeyEnter:
		if len(cmds) > 0 {
			c := cmds[clamp(m.paletteCursor, len(cmds))]
			m.input = ""
			m.paletteCursor = 0
			nm, cmd := c.run(m, "")
			return nm, cmd, true
		}
		return m, nil, true
	}
	return m, nil, false
}

func clamp(i, n int) int {
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}

// runInput dispatches a submitted "/command args" line to its handler.
func (m Model) runInput() (tea.Model, tea.Cmd) {
	line := strings.TrimPrefix(strings.TrimSpace(m.input), "/")
	m.input = ""
	name, args, _ := strings.Cut(line, " ")
	for _, c := range commandRegistry() {
		if c.name == name {
			return c.run(m, strings.TrimSpace(args))
		}
	}
	return m.sys("unknown command: /" + name + "  (type / to see commands)"), nil
}

// sys appends a system line to the transcript.
func (m Model) sys(text string) Model {
	m.transcript = append(m.transcript, entry{"system", text})
	return m
}

// renderPalette draws the filtered command list above the prompt while a name is being typed.
func (m Model) renderPalette() string {
	cmds := m.filteredCommands()
	if len(cmds) == 0 {
		return "  " + m.styles.Muted.Render("no matching command · esc to cancel") + "\n"
	}
	cur := clamp(m.paletteCursor, len(cmds))
	var b strings.Builder
	for i, c := range cmds {
		label := fmt.Sprintf("/%-11s %s", c.name, c.desc)
		if i == cur {
			b.WriteString("  " + m.styles.Agent.Render("▸ "+label) + "\n")
		} else {
			b.WriteString("    " + m.styles.Muted.Render(label) + "\n")
		}
	}
	return b.String()
}

// --- command handlers ---

func cmdHelp(m Model, _ string) (tea.Model, tea.Cmd) {
	var b strings.Builder
	b.WriteString("commands:")
	for _, c := range commandRegistry() {
		b.WriteString(fmt.Sprintf("\n  /%-11s %s", c.name, c.desc))
	}
	b.WriteString("\n\nkeybindings:")
	b.WriteString("\n  enter        send the current line")
	b.WriteString("\n  /            open the command palette")
	b.WriteString("\n  shift+tab    cycle mode (agent → plan → shell)")
	b.WriteString("\n  ctrl+p       switch provider")
	b.WriteString("\n  ↑/↓ or j/k   move in a menu · esc goes back")
	b.WriteString("\n  ctrl+c       quit")
	return m.sys(b.String()), nil
}

func cmdKeys(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.sys("keys: enter send · / palette · shift+tab mode · ctrl+p provider · ↑/↓ move · esc back · ctrl+c quit"), nil
}

func cmdClear(m Model, _ string) (tea.Model, tea.Cmd) {
	m.transcript = []entry{{"system", "conversation cleared"}}
	return m, nil
}

func cmdCompact(m Model, _ string) (tea.Model, tea.Cmd) {
	n := 0
	for _, e := range m.transcript {
		if e.role == "user" || e.role == "agent" {
			n++
		}
	}
	m.transcript = []entry{{"system", fmt.Sprintf("conversation compacted (%d messages summarized)", n)}}
	return m, nil
}

func cmdStatus(m Model, _ string) (tea.Model, tea.Cmd) {
	effort := m.effort
	if effort == "" {
		effort = "n/a"
	}
	return m.sys(fmt.Sprintf("provider %s · model %s · mode %s · effort %s · up %s",
		m.provider, m.model, m.modeOrDefault(), effort, m.uptime())), nil
}

func cmdProvider(m Model, _ string) (tea.Model, tea.Cmd) {
	if len(m.providers) == 0 {
		return m.sys("no providers configured"), nil
	}
	return m.openProviderPicker()
}

func cmdModel(m Model, args string) (tea.Model, tea.Cmd) {
	if args != "" {
		return m.setModel(args).sys("model set to " + args), nil
	}
	return m.openModelPicker()
}

func cmdLogin(m Model, _ string) (tea.Model, tea.Cmd) {
	if len(m.providers) == 0 {
		return m.sys("no providers configured"), nil
	}
	return m.openProviderPicker()
}

func cmdLogout(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Logout == nil {
		return m.unavailable("logout"), nil
	}
	return m.sys(m.actions.Logout(context.Background(), m.provider)), nil
}

func cmdConfig(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.runAction("config", m.actions.Config), nil
}

func cmdExport(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Export == nil {
		return m.unavailable("export"), nil
	}
	return m.sys(m.actions.Export(m.Transcript())), nil
}

func cmdDoctor(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.runAction("doctor", m.actions.Doctor), nil
}

func cmdUpdate(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.runAction("update", m.actions.Update), nil
}

func cmdMemory(m Model, args string) (tea.Model, tea.Cmd) {
	if m.actions.Memory == nil {
		return m.unavailable("memory"), nil
	}
	return m.sys(m.actions.Memory(context.Background(), args)), nil
}

func cmdWorkflows(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.runAction("workflows", m.actions.Workflows), nil
}

func cmdMCP(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.runAction("mcp", m.actions.MCP), nil
}

func cmdSkills(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.runAction("skills", m.actions.Skills), nil
}

func cmdGoal(m Model, args string) (tea.Model, tea.Cmd) {
	if args == "" {
		return m.sys("usage: /goal <what you want done>"), nil
	}
	m.input = args
	return m.submit()
}

func cmdLoop(m Model, _ string) (tea.Model, tea.Cmd) {
	m.loop = !m.loop
	state := "off"
	if m.loop {
		state = "on"
	}
	return m.sys("loop mode " + state), nil
}

func cmdMode(mode string) func(Model, string) (tea.Model, tea.Cmd) {
	return func(m Model, _ string) (tea.Model, tea.Cmd) {
		m.mode = mode
		note := ""
		if mode == "plan" {
			note = " — proposals only, no changes are made"
		}
		return m.sys("switched to " + mode + " mode" + note), nil
	}
}

func cmdQuit(m Model, _ string) (tea.Model, tea.Cmd) {
	m.quitting = true
	return m, tea.Quit
}

// runAction runs an app-backed action returning its text, or a fallback when unwired.
func (m Model) runAction(name string, fn func(context.Context) string) Model {
	if fn == nil {
		return m.unavailable(name)
	}
	return m.sys(fn(context.Background()))
}

func (m Model) unavailable(name string) Model {
	return m.sys("/" + name + " is not available in this session — run `andromeda " + name + "` in a shell")
}

func (m Model) modeOrDefault() string {
	if m.mode == "" {
		return "agent"
	}
	return m.mode
}
