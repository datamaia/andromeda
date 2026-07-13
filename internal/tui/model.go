package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Tagline is the brand tagline shown on the start screen (ADR-026).
const Tagline = "Your terminal companion for shipping great software."

// Responder produces the reply to a submitted line. mode is the active interaction mode
// (agent | plan | shell) so the driver can enforce it: plan proposes without changing anything,
// shell runs the line as a command, agent runs the full tool-using loop. Injected so the Model is
// testable and so the driver can wire it to the real agent.
type Responder func(goal, mode string) string

// entry is one transcript line.
type entry struct {
	role string // "user" | "agent" | "system"
	text string
}

// Model is the Bubble Tea session model.
type Model struct {
	styles     Styles
	transcript []entry
	input      string
	provider   string
	model      string
	effort     string // reasoning effort, shown in the status bar only when set
	state      string
	started    time.Time
	now        time.Time
	width      int
	height     int
	respond    Responder
	quitting   bool

	// interaction mode shown in the status bar and enforced on the tool path: agent | plan | shell.
	mode string
	loop bool // loop mode (keep iterating on the last goal) — toggled by /loop

	quitArmed   bool // a single ctrl+c was pressed; a second one quits (esc never quits the app)
	discovering bool // discovering models off the UI thread before opening the model picker

	// provider picker (ctrl+p / "/provider"), configured via WithProviderMenu
	providers        []ProviderChoice
	onSelectProvider ProviderSelectFunc

	// modal selection list (provider or model): a generic overlay driven by menu.go.
	pickerOpen   bool
	pickerKind   string // "provider" | "model" — drives onboarding chaining and esc behaviour
	pickerTitle  string
	pickerItems  []pickerItem
	pickerCursor int
	pickerErr    string
	pickerApply  func(Model, string) (Model, error)

	// first-run onboarding: require a provider (and sign-in/key) then a model before chatting.
	onboarding  bool
	pendingAuth string // provider id awaiting a browser sign-in

	// interactive provider sign-in (ChatGPT OAuth): async, streamed like an agent run.
	providerAuth ProviderAuthFunc
	authEvents   <-chan AuthEvent
	authing      bool
	authProvider string
	authURL      string

	// in-TUI API-key entry for providers whose key is missing.
	providerKeyEnv     ProviderKeyEnvFunc
	setProviderKey     ProviderSetKeyFunc
	pendingKeyEnv      string // env var to prompt for once the picker closes
	pendingKeyProvider string
	keyEntry           bool
	keyProvider        string
	keyEnvName         string
	keyInput           string

	// slash-command palette ("/"): app-backed handlers and the highlighted row.
	actions       Actions
	paletteCursor int

	// async agent run (approval-capable): the driver's runner streams events the Model drains,
	// pausing on an ApprovalRequest to prompt the user (approve/reject + allow/denylist).
	runner         AgentRunner
	agentEvents    <-chan AgentEvent
	running        bool
	approval       *ApprovalRequest
	approvalCursor int
}

// New builds a session Model.
func New(provider, model string, respond Responder) Model {
	start := time.Now()
	m := Model{
		styles:   DefaultStyles(),
		provider: provider,
		model:    model,
		state:    "ready",
		mode:     "agent",
		started:  start,
		now:      start,
		respond:  respond,
		width:    80,
		height:   24,
	}
	// The start screen shows the brand splash (mascot + tagline); this system line is what remains
	// once the conversation begins, so it stays tagline-free to avoid duplicating the splash.
	m.transcript = append(m.transcript, entry{"system", "session ready · type a goal, enter to send"})
	return m
}

// tickMsg advances the status-bar clock once per second.
type tickMsg time.Time

// tick schedules the next one-second status-bar refresh.
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Init implements tea.Model. It starts the status-bar clock so the session timer ticks live.
func (m Model) Init() tea.Cmd { return tick() }

// Update implements tea.Model (Bubble Tea v2: keyboard input arrives as tea.KeyPressMsg).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tickMsg:
		m.now = time.Time(msg)
		if m.quitting {
			return m, nil
		}
		return m, tick()
	case agentEventMsg:
		return m.handleAgentEvent(msg)
	case authEventMsg:
		return m.handleAuthEvent(msg)
	case modelsMsg:
		return m.showModelPicker(msg.models)
	case tea.PasteMsg:
		return m.handlePaste(msg.Content)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// modelsMsg carries the result of asynchronous model discovery.
type modelsMsg struct{ models []string }

// handlePaste appends pasted text (bracketed paste) to whatever field is focused — the API-key
// prompt or the input line — so long keys and multi-line goals can be pasted, not just typed.
func (m Model) handlePaste(text string) (tea.Model, tea.Cmd) {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, "\n", " ")
	switch {
	case m.keyEntry:
		m.keyInput += strings.TrimSpace(text)
	case m.approval != nil || m.pickerOpen || m.authing:
		// ignore paste while these overlays are focused
	default:
		m.input += text
	}
	return m, nil
}

// WithOnboarding starts the session in first-run mode: the provider picker opens immediately and a
// provider (plus sign-in/key) and model must be chosen before the chat is usable. Call it last, so
// the provider menu is already configured.
func (m Model) WithOnboarding() Model {
	m.onboarding = true
	nm, _ := m.openProviderPicker()
	return nm.(Model)
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Quit requires two ctrl+c presses (a single press arms; any other key disarms), so the session
	// is never torn down by a stray keypress. Esc never quits — it only cancels overlays or clears
	// the input. Ctrl+C works mid-overlay; cancelling the context unblocks any pending approver.
	if msg.Mod&tea.ModCtrl != 0 && msg.Code == 'c' {
		if m.quitArmed {
			m.quitting = true
			return m, tea.Quit
		}
		m.quitArmed = true
		return m, nil
	}
	if m.quitArmed {
		m.quitArmed = false // any other key cancels the pending quit
	}
	// The approval overlay captures all other keys while a run is paused awaiting a decision.
	if m.approval != nil {
		return m.handleApprovalKey(msg)
	}
	// The API-key prompt captures keys while the user pastes a credential.
	if m.keyEntry {
		return m.handleKeyEntryKey(msg)
	}
	// A modal picker (provider/model) captures all keys while open (esc there means "back").
	if m.pickerOpen {
		return m.handlePickerKey(msg)
	}
	// While a browser sign-in is in flight, ignore input (ctrl+c above still quits).
	if m.authing {
		return m, nil
	}
	// Shift+Tab cycles the interaction mode (agent → plan → shell), like other agent CLIs.
	if msg.Code == tea.KeyTab && msg.Mod&tea.ModShift != 0 {
		m.mode = nextMode(m.modeOrDefault())
		return m, nil
	}
	// While a slash-command name is being typed, the palette owns navigation keys (arrows/tab/enter
	// select a command; esc closes the palette). Other keys fall through to normal text editing so
	// the filter keeps narrowing as you type.
	if m.paletteActive() {
		if nm, cmd, handled := m.handlePaletteKey(msg); handled {
			return nm, cmd
		}
	}
	switch {
	case msg.Mod&tea.ModCtrl != 0 && msg.Code == 'p' && len(m.providers) > 0:
		return m.openProviderPicker()
	case msg.Code == tea.KeyEscape:
		m.input = "" // esc clears the line; it never quits the app
		return m, nil
	case msg.Code == tea.KeyEnter:
		return m.submit()
	case msg.Code == tea.KeyBackspace:
		if n := len(m.input); n > 0 {
			m.input = m.input[:n-1]
		}
	case msg.Text != "":
		// Printable input (including space) carries its characters in Text.
		m.input += msg.Text
	}
	return m, nil
}

func (m Model) submit() (tea.Model, tea.Cmd) {
	goal := strings.TrimSpace(m.input)
	if goal == "" {
		m.input = ""
		return m, nil
	}
	// A "/"-prefixed line is a slash command, not a goal — dispatch it to the registry.
	if strings.HasPrefix(goal, "/") {
		return m.runInput()
	}
	// Ignore new goals while a run is in flight (its approval overlay handles keys separately).
	if m.running {
		return m, nil
	}
	m.input = ""
	m.transcript = append(m.transcript, entry{"user", goal})
	mode := m.modeOrDefault()
	// Async path: a wired runner drives agent/plan runs so tool actions can pause for approval.
	if m.runner != nil && mode != "shell" {
		ch := m.runner(goal, mode)
		m.agentEvents = ch
		m.running = true
		m.state = "running"
		return m, waitAgent(ch)
	}
	// Synchronous path: shell mode, or when no runner is wired (e.g. unit tests).
	reply := "(no responder configured)"
	if m.respond != nil {
		m.state = "running"
		reply = m.respond(goal, mode)
		m.state = "ready"
	}
	m.transcript = append(m.transcript, entry{"agent", reply})
	return m, nil
}

// View implements tea.Model. Bubble Tea v2 returns a tea.View; AltScreen is requested here rather
// than as a program option. On the start screen (no exchanges yet) it shows the brand splash.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

// render produces the screen content as a plain (styled) string — also the unit-testable surface.
func (m Model) render() string {
	if m.quitting {
		return ""
	}
	if m.approval != nil {
		return m.renderApproval() + m.statusBar()
	}
	if m.keyEntry {
		return m.renderKeyEntry() + m.statusBar()
	}
	if m.pickerOpen {
		return m.renderPicker() + m.statusBar()
	}
	if m.discovering {
		return "\n  " + m.styles.Muted.Render("discovering models…") + "\n\n" + m.statusBar()
	}
	var b strings.Builder
	// The splash is the start-screen greeting; hide it while the command palette is open or a
	// sign-in is in flight, so the command list / progress messages and the prompt stay on screen.
	showSplash := m.atStart() && !m.paletteActive() && !m.authing
	if showSplash {
		b.WriteString(m.Splash(m.width))
	}
	for _, e := range m.transcript {
		switch e.role {
		case "user":
			b.WriteString(m.styles.User.Render("you ▸ ") + e.text + "\n")
		case "agent":
			b.WriteString(m.styles.Agent.Render("andromeda ▸ ") + e.text + "\n")
		default:
			// System lines are folded into the splash on the pure start screen, but shown whenever
			// the splash is hidden (e.g. during sign-in, so the browser URL is visible).
			if !showSplash {
				b.WriteString(m.styles.Muted.Render(e.text) + "\n")
			}
		}
	}
	b.WriteString("\n")
	if m.paletteActive() {
		b.WriteString(m.renderPalette())
	}
	b.WriteString(m.styles.Prompt.Render(m.promptSymbol()) + m.input + "▏\n")
	b.WriteString(m.statusBar())
	return b.String()
}

// nextMode cycles the interaction modes: agent → plan → shell → agent.
func nextMode(mode string) string {
	switch mode {
	case "plan":
		return "shell"
	case "shell":
		return "agent"
	default:
		return "plan"
	}
}

// promptSymbol reflects the active mode at the input line: "$" for shell, a "plan" cue for plan
// mode, and the default caret for agent mode.
func (m Model) promptSymbol() string {
	switch m.mode {
	case "shell":
		return "$ "
	case "plan":
		return "plan ❯ "
	default:
		return "❯ "
	}
}

// atStart reports whether the session has no user/agent exchanges yet (only the system banner).
func (m Model) atStart() bool {
	for _, e := range m.transcript {
		if e.role == "user" || e.role == "agent" {
			return false
		}
	}
	return true
}

// statusBar shows, live, what the session is running on: the active provider and model, the
// reasoning effort (only when set), the elapsed session time, and the current state.
func (m Model) statusBar() string {
	parts := []string{m.provider, m.model, m.modeOrDefault()}
	if m.effort != "" {
		parts = append(parts, "effort "+m.effort)
	}
	if m.loop {
		parts = append(parts, "loop")
	}
	parts = append(parts, m.uptime(), m.state)
	left := " " + strings.Join(parts, " · ") + " "
	hint := "  / commands · shift+tab: mode · ctrl+c ctrl+c: exit"
	if m.quitArmed {
		hint = "  press ctrl+c again to exit"
	}
	help := m.styles.Muted.Render(hint)
	bar := m.styles.StatusBar.Render(left)
	return lipgloss.JoinHorizontal(lipgloss.Left, bar, help)
}

// uptime is the session's elapsed wall-clock time, formatted compactly (M:SS, or H:MM:SS past
// an hour). It reads the last tick so rendering stays a pure function of Model state.
func (m Model) uptime() string {
	d := m.now.Sub(m.started)
	if d < 0 {
		d = 0
	}
	total := int(d.Seconds())
	h, mnt, s := total/3600, (total%3600)/60, total%60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, mnt, s)
	}
	return fmt.Sprintf("%d:%02d", mnt, s)
}

// Transcript returns the transcript lines as role:text pairs (for tests and export).
func (m Model) Transcript() []string {
	out := make([]string, 0, len(m.transcript))
	for _, e := range m.transcript {
		out = append(out, e.role+": "+e.text)
	}
	return out
}
