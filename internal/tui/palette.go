package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// Actions are the app-backed operations slash commands invoke. The composition root wires them so
// the TUI keeps no app/provider imports; a nil action degrades to an informative message.
type Actions struct {
	Doctor func(ctx context.Context) string
	Update func(ctx context.Context) string
	// Memory runs the /memory text subcommands (add|search|rm|list); MemoryList backs the menu.
	Memory     func(ctx context.Context, args string) string
	MemoryList func(ctx context.Context) []MemoryNote
	// Collection returns the entries of a manageable capability set (kind: skills|mcp|workflows|
	// plugins) for its interactive menu; a zero CollectionView renders an empty state.
	Collection func(ctx context.Context, kind string) CollectionView
	Models     func(ctx context.Context) []string
	Config     func(ctx context.Context) string
	Logout     func(ctx context.Context, provider string) string
	Export     func(lines []string) string
	Init       func(ctx context.Context, provider, model string) string
	Files      func(ctx context.Context) []string          // workspace files for @-mention completion
	Context    func(ctx context.Context) []string          // workspace context lines for the /status panel
	Ontology   func(ctx context.Context, op string) string // op: build|show|rm — deterministic .ttl map
	Graph      func(ctx context.Context, op string) string // op: build|open|show|rm — visual map + viewer
	Skills     func(ctx context.Context) []SkillNote       // discovered skills for $-mention invocation
	// Permission runs the /permission text subcommands (allow|deny <cmd>, rm <list> <cmd>, list);
	// Permissions returns the current allow/deny policy for the interactive menu.
	Permission  func(ctx context.Context, args string) string
	Permissions func(ctx context.Context) PermissionView
}

// WithActions wires the app-backed slash-command handlers.
func (m Model) WithActions(a Actions) Model { m.actions = a; return m }

// slashCommand is one palette entry. aliases are alternate names that match in the palette and
// dispatch to the same handler (e.g. "mem" → memory).
type slashCommand struct {
	name    string
	desc    string
	aliases []string
	run     func(m Model, args string) (tea.Model, tea.Cmd)
}

// commandRegistry is the full set of slash commands (industry-standard basics + Andromeda modes).
func commandRegistry() []slashCommand {
	return []slashCommand{
		{name: "help", desc: "getting started, modes, and keybindings", aliases: []string{"?"}, run: cmdHelp},
		{name: "commands", desc: "every slash command with its aliases", run: cmdCommands},
		{name: "keys", desc: "show keybindings", run: cmdKeys},
		{name: "clear", desc: "clear the conversation", aliases: []string{"reset"}, run: cmdClear},
		{name: "compact", desc: "summarize the conversation so far", run: cmdCompact},
		{name: "status", desc: "show provider, model, mode, and session", run: cmdStatus},
		{name: "model", desc: "choose the model (/model <name> to set)", run: cmdModel},
		{name: "effort", desc: "reasoning effort (minimal|low|medium|high)", run: cmdEffort},
		{name: "theme", desc: "color theme (dark|light)", run: cmdTheme},
		{name: "provider", desc: "choose the provider", run: cmdProvider},
		{name: "login", desc: "switch or sign in to a provider", aliases: []string{"signin"}, run: cmdLogin},
		{name: "logout", desc: "sign out of the current provider", aliases: []string{"signout"}, run: cmdLogout},
		{name: "config", desc: "show resolved configuration", run: cmdConfig},
		{name: "permission", desc: "pre-approve or block shell commands (allow/deny)", aliases: []string{"perms", "allowlist"}, run: cmdPermission},
		{name: "init", desc: "scaffold AGENTS.md, andromeda.toml, .agents/ and .andromeda/", run: cmdInit},
		{name: "export", desc: "save the transcript to a file", run: cmdExport},
		{name: "doctor", desc: "run environment checks", run: cmdDoctor},
		{name: "update", desc: "check for updates", run: cmdUpdate},
		{name: "memory", desc: "manage workspace memory", aliases: []string{"mem"}, run: cmdMemory},
		{name: "workflows", desc: "run step-by-step workflow recipes", aliases: []string{"workflow"}, run: cmdWorkflows},
		{name: "ontology", desc: "build/inspect the workspace ontology (.ttl)", aliases: []string{"onto", "ttl"}, run: cmdOntology},
		{name: "graph", desc: "build/serve a visual graph of the workspace", aliases: []string{"viz"}, run: cmdGraph},
		{name: "mcp", desc: "manage MCP servers", run: cmdMCP},
		{name: "skills", desc: "manage skills", aliases: []string{"skill"}, run: cmdSkills},
		{name: "plugins", desc: "manage plugins", aliases: []string{"plugin"}, run: cmdPlugins},
		{name: "goal", desc: "set and run a goal (/goal <text>)", run: cmdGoal},
		{name: "loop", desc: "toggle loop mode", run: cmdLoop},
		{name: "agent", desc: "switch to agent mode", run: cmdMode("agent")},
		{name: "plan", desc: "switch to plan mode (no changes are made)", run: cmdMode("plan")},
		{name: "shell", desc: "switch to shell mode", run: cmdMode("shell")},
		{name: "quit", desc: "exit Andromeda", aliases: []string{"exit"}, run: cmdQuit},
	}
}

// customCommands are user-authored slash commands discovered from disk and injected by the driver
// (W5). They are merged after the built-ins; a custom command whose name collides with a built-in
// is dropped by mergedCommands so a user file can never shadow core behaviour.
func (m Model) mergedCommands() []slashCommand {
	base := commandRegistry()
	if len(m.customCommands) == 0 {
		return base
	}
	taken := map[string]bool{}
	for _, c := range base {
		taken[c.name] = true
		for _, a := range c.aliases {
			taken[a] = true
		}
	}
	out := base
	for _, cc := range m.customCommands {
		if taken[cc.Name] {
			continue
		}
		cc := cc // capture
		out = append(out, slashCommand{
			name: cc.Name, desc: cc.Desc,
			run: func(mm Model, args string) (tea.Model, tea.Cmd) { return mm.runCustomCommand(cc, args) },
		})
	}
	return out
}

// resolveCommand finds a command by name or alias.
func resolveCommand(name string, cmds []slashCommand) *slashCommand {
	for i := range cmds {
		if cmds[i].name == name {
			return &cmds[i]
		}
		for _, a := range cmds[i].aliases {
			if a == name {
				return &cmds[i]
			}
		}
	}
	return nil
}

// paletteActive reports whether the user is typing a slash command name (so the palette shows). A
// bare "/" opens it; a path typed as input ("/Users/…", "/tmp/x", "~/y") is a goal, not a command,
// so the palette stays closed and Enter submits it normally.
func (m Model) paletteActive() bool {
	if !strings.HasPrefix(m.input, "/") || strings.Contains(m.input, " ") {
		return false
	}
	return !strings.ContainsAny(strings.TrimPrefix(m.input, "/"), "/.~")
}

// filteredCommands ranks commands against the typed query (after "/"): name/alias prefix matches
// first, then substring matches, each group in registry order. An empty query returns everything.
func (m Model) filteredCommands() []slashCommand {
	q := strings.ToLower(strings.TrimPrefix(m.input, "/"))
	all := m.mergedCommands()
	if q == "" {
		return all
	}
	var prefix, substr []slashCommand
	for _, c := range all {
		switch matchRank(q, c) {
		case rankPrefix:
			prefix = append(prefix, c)
		case rankSubstr:
			substr = append(substr, c)
		}
	}
	return append(prefix, substr...)
}

const (
	rankNone = iota
	rankSubstr
	rankPrefix
)

// matchRank scores how a query matches a command's name or any alias: prefix beats substring.
func matchRank(q string, c slashCommand) int {
	best := rankNone
	for _, name := range append([]string{c.name}, c.aliases...) {
		switch {
		case strings.HasPrefix(name, q):
			return rankPrefix
		case strings.Contains(name, q):
			best = rankSubstr
		}
	}
	return best
}

// argContext reports the command whose argument is being typed: input is "/<cmd> <partial>" and
// <cmd> is a known command (so a path like "/Users/me x" is not mistaken for a command argument).
func (m Model) argContext() (name, partial string, ok bool) {
	if !strings.HasPrefix(m.input, "/") {
		return "", "", false
	}
	rest := strings.TrimPrefix(m.input, "/")
	name, partial, found := strings.Cut(rest, " ")
	if !found {
		return "", "", false
	}
	if resolveCommand(name, m.mergedCommands()) == nil {
		return "", "", false
	}
	return name, partial, true
}

// statusTabNames are the tabs of the /status panel (W6) and the argument completions for /status.
var statusTabNames = []string{"overview", "usage", "tools", "context"}

// argCandidates lists the static argument completions for the commands that take a small fixed set.
func argCandidates(name string) []string {
	switch name {
	case "effort":
		return []string{"minimal", "low", "medium", "high"}
	case "theme":
		return []string{"dark", "light"}
	case "status":
		return statusTabNames
	case "ontology":
		return []string{"build", "show", "adjust", "rm"}
	case "graph":
		return []string{"build", "open", "show", "adjust", "rm"}
	}
	return nil
}

// filteredArgs ranks a command's argument candidates against the typed partial (prefix then
// substring, like filteredCommands).
func filteredArgs(name, partial string) []string {
	cands := argCandidates(name)
	q := strings.ToLower(partial)
	if q == "" {
		return cands
	}
	var prefix, substr []string
	for _, c := range cands {
		switch {
		case strings.HasPrefix(c, q):
			prefix = append(prefix, c)
		case strings.Contains(c, q):
			substr = append(substr, c)
		}
	}
	return append(prefix, substr...)
}

// menuKind reports what the palette is completing: "cmd" a command name, "arg" a command argument,
// or "" (closed). It drives rendering and key capture.
func (m Model) menuKind() string {
	if m.paletteActive() {
		return "cmd"
	}
	if name, _, ok := m.argContext(); ok && len(argCandidates(name)) > 0 {
		return "arg"
	}
	return ""
}

// handlePaletteKey drives palette navigation while a slash-command name (or argument) is being
// typed. It returns handled=false for keys it does not consume (so normal text editing still
// applies and the filter keeps narrowing as you type).
func (m Model) handlePaletteKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	if m.menuKind() == "arg" {
		return m.handleArgKey(msg)
	}
	cmds := m.filteredCommands()
	switch msg.Code {
	case tea.KeyEscape:
		m.input = "" // close the palette without quitting
		return m, nil, true
	case tea.KeyUp:
		if m.paletteCursor > 0 {
			m.paletteCursor--
		}
		return m, nil, true
	case tea.KeyDown:
		if m.paletteCursor < len(cmds)-1 {
			m.paletteCursor++
		}
		return m, nil, true
	case tea.KeyTab:
		if len(cmds) > 0 {
			m.input = "/" + cmds[clamp(m.paletteCursor, len(cmds))].name + " "
			m.paletteCursor = 0
		}
		return m, nil, true
	case tea.KeyEnter:
		if len(cmds) > 0 {
			c := cmds[clamp(m.paletteCursor, len(cmds))]
			m = m.pushHistory("/" + c.name)
			m.input = ""
			m.paletteCursor = 0
			nm, cmd := c.run(m, "")
			return nm, cmd, true
		}
		// No command matches the typed prefix — don't swallow Enter; let submit() report it as an
		// unknown command (or treat it as a goal if it turns out to be a path).
		return m, nil, false
	}
	return m, nil, false
}

// handleArgKey drives argument completion (e.g. "/effort me…"): ↑/↓ move, Enter completes and runs
// the command with the chosen argument, Tab completes into the input, Esc closes.
func (m Model) handleArgKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	name, partial, _ := m.argContext()
	args := filteredArgs(name, partial)
	switch msg.Code {
	case tea.KeyEscape:
		m.input = ""
		m.paletteCursor = 0
		return m, nil, true
	case tea.KeyUp:
		if m.paletteCursor > 0 {
			m.paletteCursor--
		}
		return m, nil, true
	case tea.KeyDown:
		if m.paletteCursor < len(args)-1 {
			m.paletteCursor++
		}
		return m, nil, true
	case tea.KeyTab:
		if len(args) > 0 {
			m.input = "/" + name + " " + args[clamp(m.paletteCursor, len(args))]
			m.paletteCursor = 0
		}
		return m, nil, true
	case tea.KeyEnter:
		chosen := strings.TrimSpace(partial)
		if len(args) > 0 {
			chosen = args[clamp(m.paletteCursor, len(args))]
		}
		if c := resolveCommand(name, m.mergedCommands()); c != nil && chosen != "" {
			m = m.pushHistory("/" + name + " " + chosen)
			m.input = ""
			m.paletteCursor = 0
			nm, cmd := c.run(m, chosen)
			return nm, cmd, true
		}
		return m, nil, false
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

// looksLikeSlashCommand reports whether a "/"-prefixed line is a command invocation rather than a
// filesystem path used as a goal. A command's name is a single bare word (letters, digits, "-", "_")
// with no path separator; "/Users/me/x" and "/tmp/out.md" are treated as goals, not commands.
func looksLikeSlashCommand(line string) bool {
	if !strings.HasPrefix(line, "/") {
		return false
	}
	name, _, _ := strings.Cut(strings.TrimPrefix(line, "/"), " ")
	if name == "" {
		return false
	}
	for _, r := range name {
		if r == '/' || r == '.' || r == '~' {
			return false // a path segment, not a command name
		}
	}
	return true
}

// runInput dispatches a submitted "/command args" line to its handler, resolving aliases and
// user-defined custom commands.
func (m Model) runInput() (tea.Model, tea.Cmd) {
	m = m.pushHistory(strings.TrimSpace(m.input))
	line := strings.TrimPrefix(strings.TrimSpace(m.input), "/")
	m.input = ""
	name, args, _ := strings.Cut(line, " ")
	if c := resolveCommand(name, m.mergedCommands()); c != nil {
		return c.run(m, strings.TrimSpace(args))
	}
	return m.sys("unknown command: /" + name + "  (type / to see commands)"), nil
}

// sys appends a system line to the transcript.
func (m Model) sys(text string) Model {
	m.transcript = append(m.transcript, entry{"system", text})
	return m
}

// renderPalette draws the command list, or (in argument mode) the argument candidates, above the
// prompt. Long lists scroll within a bounded window around the cursor with ↑/↓ more markers.
func (m Model) renderPalette() string {
	if m.menuKind() == "arg" {
		return m.renderArgMenu()
	}
	cmds := m.filteredCommands()
	if len(cmds) == 0 {
		return "  " + m.styles.Muted.Render("no matching command · esc to cancel") + "\n"
	}
	labels := make([]string, len(cmds))
	for i, c := range cmds {
		labels[i] = fmt.Sprintf("/%-11s %s", c.name, c.desc)
	}
	return m.renderMenuRows(labels, clamp(m.paletteCursor, len(cmds)))
}

// renderArgMenu draws the argument candidates for the command being completed.
func (m Model) renderArgMenu() string {
	name, partial, _ := m.argContext()
	args := filteredArgs(name, partial)
	if len(args) == 0 {
		return "  " + m.styles.Muted.Render("no matching option · esc to cancel") + "\n"
	}
	labels := make([]string, len(args))
	for i, a := range args {
		labels[i] = fmt.Sprintf("/%s %s", name, a)
	}
	return m.renderMenuRows(labels, clamp(m.paletteCursor, len(args)))
}

// renderMenuRows renders a list of labels with the cursor highlighted, scrolled within a window
// sized to the terminal so a long list never pushes the prompt off-screen.
func (m Model) renderMenuRows(labels []string, cur int) string {
	win := m.menuWindow()
	top := 0
	if cur >= win {
		top = cur - win + 1
	}
	if top > len(labels)-win {
		top = len(labels) - win
	}
	if top < 0 {
		top = 0
	}
	end := top + win
	if end > len(labels) {
		end = len(labels)
	}
	var b strings.Builder
	if top > 0 {
		b.WriteString("    " + m.styles.Muted.Render(fmt.Sprintf("↑ %d more", top)) + "\n")
	}
	for i := top; i < end; i++ {
		if i == cur {
			b.WriteString("  " + m.styles.Agent.Render("▸ "+labels[i]) + "\n")
		} else {
			b.WriteString("    " + m.styles.Muted.Render(labels[i]) + "\n")
		}
	}
	if end < len(labels) {
		b.WriteString("    " + m.styles.Muted.Render(fmt.Sprintf("↓ %d more", len(labels)-end)) + "\n")
	}
	return b.String()
}

// menuWindow is how many palette rows are shown at once, bounded to the terminal height.
func (m Model) menuWindow() int {
	h := m.height - 8 // splash/prompt/status/hints headroom
	if h < 5 {
		return 5
	}
	if h > 12 {
		return 12
	}
	return h
}

// --- command handlers ---

// cmdHelp is the orientation guide: what to do, the modes, inline sigils, and keybindings. It points
// at /commands for the exhaustive command reference (the two used to be identical dumps).
func cmdHelp(m Model, _ string) (tea.Model, tea.Cmd) {
	var b strings.Builder
	b.WriteString("Andromeda — your terminal companion\n")
	b.WriteString("\nType a goal and press enter. Commands start with / — press / to browse them, or run")
	b.WriteString("\n/commands for the full list.")
	b.WriteString("\n\nmodes  (shift+tab cycles):")
	b.WriteString("\n  agent   make changes with tools and approvals")
	b.WriteString("\n  plan    think through an approach; no edits are made")
	b.WriteString("\n  shell   run shell commands directly")
	b.WriteString("\n\ninline:")
	b.WriteString("\n  /   command palette      @   mention a file      $   invoke a skill")
	b.WriteString("\n\nkeybindings:")
	b.WriteString("\n  enter          send the current line")
	b.WriteString("\n  ↑/↓            recall previous/next inputs (navigate a menu when one is open)")
	b.WriteString("\n  mouse wheel    scroll the conversation, back to the oldest output of the session")
	b.WriteString("\n  PgUp/PgDn      scroll the conversation · ctrl+u/ctrl+d half-page · Home top · End latest")
	b.WriteString("\n  shift+tab      cycle mode: agent → plan → shell → agent")
	b.WriteString("\n  esc            interrupt a running turn / clear the line")
	b.WriteString("\n  ctrl+p         switch provider · ctrl+c ctrl+c quit")
	b.WriteString("\n\nmore: /commands (every command) · /status (session) · /doctor (environment)")
	return m.sys(b.String()), nil
}

// cmdCommands is the exhaustive command reference: every slash command (built-in and user-authored)
// with its aliases and description. Distinct from /help, which is the getting-started guide.
func cmdCommands(m Model, _ string) (tea.Model, tea.Cmd) {
	cmds := m.mergedCommands()
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "slash commands (%d):", len(cmds))
	for _, c := range cmds {
		name := "/" + c.name
		if len(c.aliases) > 0 {
			name += " /" + strings.Join(c.aliases, " /")
		}
		_, _ = fmt.Fprintf(&b, "\n  %-26s %s", name, c.desc)
	}
	b.WriteString("\n\ntip: type / then a fragment to filter; enter or → to run.")
	return m.sys(b.String()), nil
}

func cmdKeys(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.sys("keys: enter send · ↑/↓ recall inputs (move in menus) · PgUp/PgDn scroll · / palette · @ files · shift+tab mode · ctrl+p provider · esc back · ctrl+c ctrl+c quit"), nil
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

func cmdStatus(m Model, args string) (tea.Model, tea.Cmd) {
	return m.openStatusPanel(args), nil
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

// effortLevels are the reasoning-effort settings offered by /effort.
var effortLevels = []string{"minimal", "low", "medium", "high"}

func validEffort(s string) bool {
	for _, l := range effortLevels {
		if l == s {
			return true
		}
	}
	return false
}

// cmdEffort sets the reasoning effort directly (/effort high) or opens a picker (/effort).
func cmdEffort(m Model, args string) (tea.Model, tea.Cmd) {
	if args != "" {
		if !validEffort(args) {
			return m.sys("effort must be one of: " + strings.Join(effortLevels, " | ")), nil
		}
		return m.setEffort(args), nil
	}
	return m.openEffortPicker()
}

// setEffort records the effort on the view and propagates it to the driver (so the agent uses it).
func (m Model) setEffort(id string) Model {
	m.effort = id
	if m.onSelectEffort != nil {
		m.onSelectEffort(id)
	}
	return m.sys("effort → " + id)
}

func (m Model) openEffortPicker() (tea.Model, tea.Cmd) {
	items := make([]pickerItem, 0, len(effortLevels))
	for _, l := range effortLevels {
		items = append(items, pickerItem{id: l, display: l})
	}
	m.pickerKind = "effort"
	return m.openPicker("Reasoning effort", items, m.effort, func(mm Model, id string) (Model, error) {
		return mm.setEffort(id), nil
	})
}

// cmdTheme switches the color theme directly (/theme light) or opens a picker (/theme).
func cmdTheme(m Model, args string) (tea.Model, tea.Cmd) {
	if args != "" {
		nm, err := m.setTheme(args)
		if err != nil {
			return m.sys(err.Error()), nil
		}
		return nm, nil
	}
	return m.openThemePicker()
}

// setTheme swaps the live style set. The Markdown renderer reads m.styles, so it re-themes too.
func (m Model) setTheme(name string) (Model, error) {
	switch name {
	case "light":
		m.styles = LightStyles()
		m.theme = "light"
	case "dark":
		m.styles = DefaultStyles()
		m.theme = "dark"
	default:
		return m, fmt.Errorf("theme must be dark or light")
	}
	return m.sys("theme → " + name), nil
}

func (m Model) themeName() string {
	if m.theme == "" {
		return "dark"
	}
	return m.theme
}

func (m Model) openThemePicker() (tea.Model, tea.Cmd) {
	items := []pickerItem{{id: "dark", display: "dark"}, {id: "light", display: "light"}}
	m.pickerKind = "theme"
	return m.openPicker("Color theme", items, m.themeName(), func(mm Model, id string) (Model, error) {
		return mm.setTheme(id)
	})
}

// CustomCommand is a user-authored slash command (a Markdown prompt template) injected by the
// driver. $ARGUMENTS expands to the whole argument string; $1..$9 to positional words (W5).
type CustomCommand struct {
	Name     string
	Desc     string
	Template string
}

// WithCustomCommands injects user-defined slash commands discovered from disk by the driver.
func (m Model) WithCustomCommands(cmds []CustomCommand) Model {
	m.customCommands = cmds
	return m
}

// WithEffortSelect wires a callback that propagates a reasoning-effort choice to the driver.
func (m Model) WithEffortSelect(fn func(string)) Model {
	m.onSelectEffort = fn
	return m
}

// runCustomCommand expands a custom command's template with the given arguments and submits it as a
// goal (the transcript records the invocation; the model receives the expanded prompt).
func (m Model) runCustomCommand(cc CustomCommand, args string) (tea.Model, tea.Cmd) {
	m.transcript = append(m.transcript, entry{"system", "/" + cc.Name + " " + args})
	m.input = expandTemplate(cc.Template, args)
	return m.submit()
}

// expandTemplate substitutes $ARGUMENTS and $1..$9 in a custom-command template.
func expandTemplate(tpl, args string) string {
	out := strings.ReplaceAll(tpl, "$ARGUMENTS", args)
	fields := strings.Fields(args)
	for i := 1; i <= 9; i++ {
		val := ""
		if i-1 < len(fields) {
			val = fields[i-1]
		}
		out = strings.ReplaceAll(out, "$"+strconv.Itoa(i), val)
	}
	return out
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

func cmdInit(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Init == nil {
		return m.unavailable("init"), nil
	}
	return m.sys(m.actions.Init(context.Background(), m.provider, m.model)), nil
}

func cmdDoctor(m Model, _ string) (tea.Model, tea.Cmd) {
	return m.runAction("doctor", m.actions.Doctor), nil
}

func cmdUpdate(m Model, _ string) (tea.Model, tea.Cmd) {
	if m.actions.Update == nil {
		return m.unavailable("update"), nil
	}
	// The update check may hit the network; run it off the UI thread and show progress immediately so
	// the interface never freezes while it waits.
	fn := m.actions.Update
	return m.sys("checking for updates…"), func() tea.Msg {
		return noticeMsg{text: fn(context.Background())}
	}
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
