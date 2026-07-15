package tui

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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
	role string // "user" | "agent" | "tool" | "system" | "greeting"
	text string
}

// Model is the Bubble Tea session model.
type Model struct {
	styles     Styles
	transcript []entry
	input      string
	provider   string
	model      string
	version    string // andromeda version shown in the header (injected via WithVersion)
	effort     string // reasoning effort, shown in the status bar only when set
	state      string
	started    time.Time
	now        time.Time
	width      int
	height     int
	respond    Responder
	quitting   bool

	// scrollOffset is how many lines the transcript is scrolled up from the bottom (0 = latest,
	// auto-following new output). PageUp/PageDown/Home/End move it so history is navigable.
	scrollOffset int

	// interaction mode shown in the status bar and enforced on the tool path: agent | plan | shell.
	mode string
	loop bool // loop mode (keep iterating on the last goal) — toggled by /loop

	quitArmed   bool // a single ctrl+c was pressed; a second one quits (esc never quits the app)
	discovering bool // discovering models off the UI thread before opening the model picker

	// provider picker (ctrl+p / "/provider"), configured via WithProviderMenu
	providers        []ProviderChoice
	onSelectProvider ProviderSelectFunc
	onSelectModel    func(string) // propagate a model choice to the driver so the agent uses it

	// navigable command menu (skills/mcp/workflows/plugins/ontology/graph/memory): a drill-in/back
	// stack driven by commandmenu.go. Non-empty means a menu overlay is on screen.
	menu []menuLevel

	// modal selection list (provider or model): a generic overlay driven by menu.go.
	pickerOpen   bool
	pickerKind   string // "provider" | "model" — drives onboarding chaining and esc behaviour
	pickerTitle  string
	pickerItems  []pickerItem
	pickerCursor int    // index into the FILTERED list
	pickerTop    int    // first visible filtered row (viewport scroll offset)
	pickerFilter string // type-to-filter query narrowing a long list (e.g. OpenRouter's 400+ models)
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

	// input history: submitted lines are recalled with ↑/↓. histIdx == len(promptHistory) means the
	// live draft (saved in histDraft the first time ↑ leaves it) is showing.
	promptHistory []string
	histIdx       int
	histDraft     string

	// reasoning effort selection (/effort) propagated to the driver so the agent runs with it.
	onSelectEffort func(string)

	// color theme name ("dark" | "light"), switched live by /theme.
	theme string

	// user-authored slash commands (/W5), discovered by the driver and merged into the palette.
	customCommands []CustomCommand

	// @-mention file completion: the workspace file list (loaded once from Actions.Files) and the
	// highlighted candidate.
	fileList    []string
	filesLoaded bool
	atCursor    int

	// plan-review handoff: after a plan-mode turn completes, an approve/refine/reject overlay opens.
	planReview       bool
	planReviewCursor int

	// workspace context (fetched once by the driver) and cumulative token usage, shown in the status
	// bar's context row and the /status panel.
	workspaceRoot string
	branch        string
	inTokens      int
	outTokens     int

	// tabbed /status panel (Overview | Usage | Tools | Context).
	statusPanel bool
	statusTab   int

	// async agent run (approval-capable): the driver's runner streams events the Model drains,
	// pausing on an ApprovalRequest to prompt the user (approve/reject + allow/denylist).
	runner         AgentRunner
	agentEvents    <-chan AgentEvent
	running        bool
	approval       *ApprovalRequest
	approvalCursor int

	// live streaming: the transcript index of the agent line currently being streamed and of the
	// open tool card (-1 = none); the cancel that Esc calls to interrupt the run; and whether the
	// terminal error is the result of that interrupt (so it reads "interrupted", not an error).
	streamIdx   int
	toolIdx     int
	cancelRun   func()
	interrupted bool
	spinner     int       // frame index for the "working" spinner
	runStarted  time.Time // when the current run began, for the live "working (Ns)" timer
}

// New builds a session Model.
func New(provider, model string, respond Responder) Model {
	start := time.Now()
	m := Model{
		styles:    DefaultStyles(),
		provider:  provider,
		model:     model,
		state:     "ready",
		mode:      "agent",
		started:   start,
		now:       start,
		respond:   respond,
		width:     80,
		height:    24,
		streamIdx: -1,
		toolIdx:   -1,
	}
	// The start screen shows the brand splash (mascot + tagline). This greeting is rendered *by* the
	// splash, so it carries the "greeting" role: it never prints as a transcript line, and — crucially
	// — it does not count as content, so the splash yields to the transcript the instant a command
	// emits output (see hasContent / bodyString).
	m.transcript = append(m.transcript, entry{"greeting", "session ready · type a goal, enter to send"})
	return m
}

// tickMsg advances the status-bar clock once per second.
type tickMsg time.Time

// tick schedules the next one-second status-bar refresh.
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// spinnerTickMsg drives the working-spinner animation at a high frame rate while a run is active, so
// it feels fluid instead of advancing only once per second with the status clock. The loop runs for
// the life of the program; it only advances the frame while a run is active, so idle frames render
// identically and Bubble Tea's cell diff writes nothing.
type spinnerTickMsg struct{}

// spinnerFPS is the spinner animation interval (~12.5 fps) — fast enough to read as smooth motion.
const spinnerFPS = 80 * time.Millisecond

func spinnerTick() tea.Cmd {
	return tea.Tick(spinnerFPS, func(time.Time) tea.Msg { return spinnerTickMsg{} })
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// spinnerFrame is the current spinner glyph.
func (m Model) spinnerFrame() string {
	return spinnerFrames[m.spinner%len(spinnerFrames)]
}

// Init implements tea.Model. It starts the status-bar clock (once per second) and the fluid
// working-spinner loop (advances only during a run) so both animate live.
func (m Model) Init() tea.Cmd { return tea.Batch(tick(), spinnerTick()) }

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
		return m, tick() // the clock ticks once per second; the spinner animates on its own faster tick
	case spinnerTickMsg:
		// The spinner loop lives for the whole program: advance the frame only while a run is active
		// (idle frames are identical, so the renderer writes nothing), and stop rescheduling on quit.
		if m.quitting {
			return m, nil
		}
		if m.running {
			m.spinner++
		}
		return m, spinnerTick()
	case agentEventMsg:
		return m.handleAgentEvent(msg)
	case authEventMsg:
		return m.handleAuthEvent(msg)
	case modelsMsg:
		return m.showModelPicker(msg.models)
	case noticeMsg:
		return m.sys(msg.text), nil
	case tea.PasteMsg:
		return m.handlePaste(msg.Content)
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// wheelScrollLines is how many transcript lines one mouse-wheel notch moves.
const wheelScrollLines = 3

// handleMouseWheel scrolls the transcript with the mouse wheel: up reveals older output (back toward
// the oldest of the active session), down returns toward the latest. This is deliberately separate
// from ↑/↓, which recall previously submitted input lines.
func (m Model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseWheelUp:
		m.scrollOffset = m.clampScroll(m.scrollOffset + wheelScrollLines)
	case tea.MouseWheelDown:
		m.scrollOffset = m.clampScroll(m.scrollOffset - wheelScrollLines)
	}
	return m, nil
}

// modelsMsg carries the result of asynchronous model discovery.
type modelsMsg struct{ models []string }

// noticeMsg carries the text of an off-thread action (e.g. a network-backed update check) to append
// to the transcript once it completes, so the UI never blocks while the work runs.
type noticeMsg struct{ text string }

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
	// The plan-review overlay captures keys while the user decides what to do with a proposed plan.
	if m.planReview {
		return m.handlePlanReviewKey(msg)
	}
	// The /status panel captures navigation keys while open.
	if m.statusPanel {
		return m.handleStatusKey(msg)
	}
	// The API-key prompt captures keys while the user pastes a credential.
	if m.keyEntry {
		return m.handleKeyEntryKey(msg)
	}
	// A modal picker (provider/model) captures all keys while open (esc there means "back").
	if m.pickerOpen {
		return m.handlePickerKey(msg)
	}
	// A navigable command menu captures keys while open (drill in / back / select).
	if m.menuOpen() {
		return m.handleMenuKey(msg)
	}
	// While a browser sign-in is in flight, ignore input (ctrl+c above still quits).
	if m.authing {
		return m, nil
	}
	// Shift+Tab cycles the interaction mode (agent → plan → shell), like other agent CLIs. A visible
	// transcript line confirms the switch, since the cycle passes through shell on the way back to agent.
	if msg.Code == tea.KeyTab && msg.Mod&tea.ModShift != 0 {
		m.mode = nextMode(m.modeOrDefault())
		m = m.sys("mode → " + m.mode + modeHint(m.mode))
		m.scrollOffset = 0
		return m, nil
	}
	// While a slash-command name (or argument) is being typed, the palette owns navigation keys
	// (arrows/tab/enter select; esc closes). Other keys fall through to normal text editing so the
	// filter keeps narrowing as you type.
	if m.menuKind() != "" {
		if nm, cmd, handled := m.handlePaletteKey(msg); handled {
			return nm, cmd
		}
	}
	// The @-mention menu owns navigation keys while an "@fragment" file token is being typed.
	if m.atActive() {
		if nm, cmd, handled := m.handleAtKey(msg); handled {
			return nm, cmd
		}
	}
	switch {
	case msg.Mod&tea.ModCtrl != 0 && msg.Code == 'p' && len(m.providers) > 0:
		return m.openProviderPicker()
	// Scroll the transcript history. Offset 0 follows new output; PageUp/ctrl-u scroll back.
	case msg.Code == tea.KeyPgUp:
		m.scrollOffset = m.clampScroll(m.scrollOffset + m.availHeight() - 1)
		return m, nil
	case msg.Code == tea.KeyPgDown:
		m.scrollOffset = m.clampScroll(m.scrollOffset - (m.availHeight() - 1))
		return m, nil
	case msg.Mod&tea.ModCtrl != 0 && msg.Code == 'u':
		m.scrollOffset = m.clampScroll(m.scrollOffset + m.availHeight()/2)
		return m, nil
	case msg.Mod&tea.ModCtrl != 0 && msg.Code == 'd':
		m.scrollOffset = m.clampScroll(m.scrollOffset - m.availHeight()/2)
		return m, nil
	case msg.Code == tea.KeyHome:
		m.scrollOffset = m.maxScroll()
		return m, nil
	case msg.Code == tea.KeyEnd:
		m.scrollOffset = 0
		return m, nil
	// ↑/↓ recall previous/next submitted lines (the palette owns these keys when it is open, handled
	// above). At the newest entry, ↓ restores the draft that was being typed before recall began.
	case msg.Code == tea.KeyUp:
		return m.historyPrev(), nil
	case msg.Code == tea.KeyDown:
		return m.historyNext(), nil
	case msg.Code == tea.KeyEscape:
		// Esc interrupts a running turn (cancel the run's context); otherwise it clears the line. It
		// never quits the app.
		if m.running && m.cancelRun != nil {
			m.interrupted = true
			m.cancelRun()
			m.state = "interrupting"
			return m, nil
		}
		m.input = ""
		return m, nil
	case msg.Code == tea.KeyEnter:
		return m.submit()
	case msg.Code == tea.KeyBackspace:
		if n := len(m.input); n > 0 {
			m.input = m.input[:n-1]
		}
		m.histIdx = len(m.promptHistory) // editing exits history recall
	case msg.Text != "":
		// Printable input (including space) carries its characters in Text.
		m.input += msg.Text
		m.histIdx = len(m.promptHistory) // editing exits history recall
		// Load the workspace file list the first time an "@" token appears, so the mention menu can
		// open immediately as the user keeps typing.
		if _, ok := atToken(m.input); ok && !m.filesLoaded {
			m = m.loadFiles()
		}
	}
	return m, nil
}

// pushHistory records a submitted line for ↑/↓ recall, collapsing a run of identical lines, and
// resets the recall cursor to the newest position.
func (m Model) pushHistory(line string) Model {
	line = strings.TrimSpace(line)
	if line != "" {
		if n := len(m.promptHistory); n == 0 || m.promptHistory[n-1] != line {
			m.promptHistory = append(m.promptHistory, line)
		}
	}
	m.histIdx = len(m.promptHistory)
	m.histDraft = ""
	return m
}

// historyPrev recalls the previous submitted line, saving the live draft the first time recall
// leaves it so ↓ can restore it.
func (m Model) historyPrev() Model {
	if len(m.promptHistory) == 0 {
		return m
	}
	if m.histIdx == len(m.promptHistory) {
		m.histDraft = m.input
	}
	if m.histIdx > 0 {
		m.histIdx--
		m.input = m.promptHistory[m.histIdx]
	}
	return m
}

// historyNext moves toward newer lines; stepping past the newest restores the saved draft.
func (m Model) historyNext() Model {
	if m.histIdx >= len(m.promptHistory) {
		return m
	}
	m.histIdx++
	if m.histIdx == len(m.promptHistory) {
		m.input = m.histDraft
	} else {
		m.input = m.promptHistory[m.histIdx]
	}
	return m
}

func (m Model) submit() (tea.Model, tea.Cmd) {
	goal := strings.TrimSpace(m.input)
	if goal == "" {
		m.input = ""
		return m, nil
	}
	// A "/"-prefixed line is a slash command, not a goal — but only when it looks like one. A bare
	// filesystem path (e.g. "/Users/me/project") also starts with "/", so it must reach the agent
	// as a goal instead of being rejected as an unknown command.
	if looksLikeSlashCommand(goal) {
		return m.runInput()
	}
	// Ignore new goals while a run is in flight (its approval overlay handles keys separately).
	if m.running {
		return m, nil
	}
	m = m.pushHistory(goal)
	m.input = ""
	m.scrollOffset = 0 // jump to the bottom so the new exchange is visible
	m.transcript = append(m.transcript, entry{"user", goal})
	mode := m.modeOrDefault()
	// Async path: a wired runner drives agent/plan runs so tool actions can pause for approval and
	// content streams in live. The cancel interrupts the run when the user presses Esc.
	if m.runner != nil && mode != "shell" {
		ch, cancel := m.runner(goal, mode)
		m.agentEvents = ch
		m.cancelRun = cancel
		m.running = true
		m.interrupted = false
		m.streamIdx = -1
		m.toolIdx = -1
		m.spinner = 0
		m.runStarted = m.now
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

// View implements tea.Model. Bubble Tea v2 returns a tea.View; AltScreen and mouse tracking are
// requested here rather than as program options. On the start screen (no exchanges yet) it shows the
// brand splash. MouseModeCellMotion lets the wheel scroll the transcript (see handleMouseWheel);
// without it the terminal translates the wheel into ↑/↓ keys, which would recall input history.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// render produces the screen content as a plain (styled) string — also the unit-testable surface.
func (m Model) render() string {
	if m.quitting {
		return ""
	}
	// The top header banner is always visible — the live session status stays on screen even while an
	// overlay owns the rest of the view or the transcript is scrolled back.
	header := m.headerString()
	if m.approval != nil {
		return header + m.wrap(m.renderApproval()) + m.statusBar()
	}
	if m.keyEntry {
		return header + m.wrap(m.renderKeyEntry()) + m.statusBar()
	}
	if m.pickerOpen {
		return header + m.wrap(m.renderPicker()) + m.statusBar()
	}
	if m.menuOpen() {
		return header + m.wrap(m.renderMenu()) + m.statusBar()
	}
	if m.statusPanel {
		return header + m.wrap(m.renderStatusPanel()) + m.statusBar()
	}
	if m.discovering {
		return header + "\n  " + m.styles.Muted.Render("discovering models…") + "\n\n" + m.statusBar()
	}
	// The transcript scrolls within the space between the header and a fixed footer (spinner, prompt,
	// hints), so the prompt is always visible and history is navigable with PageUp/PageDown/Home/End.
	avail := m.availHeight()
	visible, above := viewportLines(m.bodyString(), avail, m.scrollOffset)
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(visible)
	b.WriteString("\n")
	if above > 0 {
		b.WriteString(m.wrap(m.styles.Muted.Render(fmt.Sprintf("  ↑ %d more · PgUp/PgDn scroll · End latest", above))))
	}
	b.WriteString("\n")
	b.WriteString(m.footerString())
	return b.String()
}

// bodyString builds the scrollable transcript: the start-screen splash, or the user/agent/tool/
// system lines wrapped to the terminal width.
func (m Model) bodyString() string {
	var b strings.Builder
	// The splash is the start-screen greeting; it shows only until there is real content to display
	// (a conversation turn OR any command output). It is also hidden while the palette is open or a
	// sign-in is in flight, so the command list / progress messages and the prompt stay on screen.
	showSplash := !m.hasContent() && m.menuKind() == "" && !m.atActive() && !m.authing
	if showSplash {
		b.WriteString(m.Splash(m.width))
	}
	for _, e := range m.transcript {
		switch e.role {
		case "user":
			b.WriteString(m.wrap(m.styles.User.Render("you ▸ ")+e.text) + "\n")
		case "agent":
			b.WriteString(m.wrap(m.styles.Agent.Render("andromeda ▸ ")+renderMarkdown(e.text, m.styles)) + "\n")
		case "tool":
			b.WriteString(m.wrap(m.styles.Tool.Render("  "+e.text)) + "\n")
		case "greeting":
			// Rendered by the splash; never shown as a transcript line.
			continue
		default: // system — command output and notices, always visible
			b.WriteString(m.wrap(m.styles.Muted.Render(e.text)) + "\n")
		}
	}
	return b.String()
}

// footerString builds the fixed bottom block: an optional working spinner, the command palette when
// open, the prompt line, and the status bar.
func (m Model) footerString() string {
	var b strings.Builder
	if m.running {
		status := "working"
		if m.state == "streaming" {
			status = "responding"
		}
		b.WriteString(m.wrap(m.styles.Agent.Render(m.spinnerFrame()+" ")+
			m.styles.Muted.Render(status+m.runElapsed()+" · esc to interrupt")) + "\n")
	}
	switch {
	case m.planReview:
		b.WriteString(m.renderPlanReview())
	case m.menuKind() != "":
		b.WriteString(m.renderPalette())
	case m.atActive():
		b.WriteString(m.renderAtMenu())
	default:
		// A thin rule + blank line give the compose area room to breathe, so it reads like a proper
		// chat box separated from the conversation above (only when no overlay owns the space).
		b.WriteString(m.styles.Muted.Render(strings.Repeat("─", max(1, m.width))) + "\n\n")
	}
	b.WriteString(m.wrap(m.styles.Prompt.Render(m.promptSymbol())+m.input+"▏") + "\n\n")
	b.WriteString(m.statusBar())
	return b.String()
}

// availHeight is the number of transcript rows that fit between the top header and the footer
// (reserving one line for the "↑ N more" scroll indicator).
func (m Model) availHeight() int {
	avail := m.height - lipgloss.Height(m.headerString()) - lipgloss.Height(m.footerString()) - 1
	if avail < 1 {
		avail = 1
	}
	return avail
}

// bodyLineCount is the number of wrapped transcript lines (for scroll clamping).
func (m Model) bodyLineCount() int {
	return len(strings.Split(strings.TrimRight(m.bodyString(), "\n"), "\n"))
}

// maxScroll is the furthest the transcript can scroll up (0 when it all fits).
func (m Model) maxScroll() int {
	if mx := m.bodyLineCount() - m.availHeight(); mx > 0 {
		return mx
	}
	return 0
}

// clampScroll keeps a scroll offset within [0, maxScroll].
func (m Model) clampScroll(n int) int {
	if n < 0 {
		return 0
	}
	if mx := m.maxScroll(); n > mx {
		return mx
	}
	return n
}

// viewportLines returns the visible slice of body (offset lines up from the bottom) and how many
// lines remain hidden above it. Offset 0 shows the latest content and follows new output.
func viewportLines(body string, avail, offset int) (string, int) {
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	total := len(lines)
	if total <= avail {
		return strings.Join(lines, "\n"), 0
	}
	if maxOff := total - avail; offset > maxOff {
		offset = maxOff
	}
	if offset < 0 {
		offset = 0
	}
	top := total - avail - offset
	return strings.Join(lines[top:top+avail], "\n"), top
}

// wrap word-wraps a styled line to the terminal width (ANSI-aware, so styling is preserved and the
// visible width is what counts). A long unbroken token (a path, a URL) is hard-broken rather than
// allowed to overflow. With no known width it is a no-op.
func (m Model) wrap(s string) string {
	if m.width <= 1 {
		return s
	}
	return ansi.Wrap(s, m.width, "")
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

// modeHint is a short reminder of what a mode does, shown when switching.
func modeHint(mode string) string {
	switch mode {
	case "plan":
		return "  (propose only — no changes)"
	case "shell":
		return "  (runs your shell commands)"
	default:
		return "  (acts with tools; asks approval)"
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

// hasContent reports whether the transcript holds anything beyond the start-screen greeting: a
// user/agent/tool line, or a system notice emitted by a command. Once true, the splash yields to the
// transcript so command output and messages are actually visible (this is what makes slash commands
// like /graph or /skills show their result immediately after launch).
func (m Model) hasContent() bool {
	for _, e := range m.transcript {
		if e.role != "greeting" {
			return true
		}
	}
	return false
}

// headerString is the persistent top banner: a live status the user can always see, even scrolled
// back. Line 1 (identity) is the brand mark plus provider/model/mode; line 2 (context) is the git
// branch, workspace, token usage, run state, and elapsed time; line 3 is a rule. Each line is
// clamped to the terminal width so it never wraps or bleeds.
func (m Model) headerString() string {
	title := m.styles.Title.Render("▌ Andromeda")
	if m.version != "" {
		title += m.styles.Muted.Render(" " + m.version)
	}
	// The active mode is a colored badge (violet=agent, amber=plan, green=shell) so it's unmistakable.
	id := title + "  " + modeBadge(m.modeOrDefault()) +
		m.styles.Muted.Render("  · "+m.provider+" · "+m.model)
	if m.effort != "" {
		id += m.styles.Muted.Render(" · effort " + m.effort)
	}
	if m.loop {
		id += m.styles.Muted.Render(" · loop")
	}

	var ctx []string
	if m.branch != "" {
		ctx = append(ctx, "⎇ "+m.branch)
	}
	if m.workspaceRoot != "" {
		ctx = append(ctx, filepath.Base(m.workspaceRoot))
	}
	if m.inTokens > 0 || m.outTokens > 0 {
		ctx = append(ctx, fmt.Sprintf("tok ↑%s ↓%s", humanCount(m.inTokens), humanCount(m.outTokens)))
	}
	ctx = append(ctx, m.state, "⏱ "+m.uptime())
	ctxLine := m.styles.Muted.Render("  " + strings.Join(ctx, " · "))

	rule := m.styles.Muted.Render(strings.Repeat("─", max(1, m.width)))
	return m.clampLine(id) + "\n" + m.clampLine(ctxLine) + "\n" + m.clampLine(rule)
}

// clampLine truncates a styled line to the terminal width (ANSI-aware) so it never overflows.
func (m Model) clampLine(s string) string {
	if m.width > 1 && ansi.StringWidth(s) > m.width {
		return ansi.Truncate(s, m.width, "…")
	}
	return s
}

// statusBar is the slim bottom line: keybinding hints (the live session status now lives in the top
// header). It drops to a quit prompt when ctrl+c is armed, and truncates on a narrow terminal.
func (m Model) statusBar() string {
	hint := "  ↑↓ history · PgUp/PgDn scroll · / commands · @ files · shift+tab mode · ctrl+c×2 exit"
	if m.quitArmed {
		hint = "  press ctrl+c again to exit"
	}
	h := m.styles.Muted.Render(hint)
	if m.width > 1 && ansi.StringWidth(h) > m.width {
		return ansi.Truncate(h, m.width, "…")
	}
	return h
}

// humanCount abbreviates large token counts (1234 → 1.2k, 1500000 → 1.5M).
func humanCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return strconv.Itoa(n)
	}
}

// WithVersion records the running andromeda version so the header can display it.
func (m Model) WithVersion(v string) Model {
	m.version = v
	return m
}

// WithWorkspace records the workspace root and git branch (fetched once by the driver) for the
// status bar's context row and the /status Context tab.
func (m Model) WithWorkspace(root, branch string) Model {
	m.workspaceRoot = root
	m.branch = branch
	return m
}

// HistoryEntry is one restored transcript line for a resumed session (Role is "user" or "agent").
type HistoryEntry struct {
	Role string
	Text string
}

// WithHistory re-seeds the transcript from a resumed session so the prior conversation is visible
// (the agent's own memory is restored separately by the driver). It replaces the default greeting.
func (m Model) WithHistory(entries []HistoryEntry) Model {
	if len(entries) == 0 {
		return m
	}
	m.transcript = []entry{{"system", fmt.Sprintf("resumed session · %d messages restored", len(entries))}}
	for _, e := range entries {
		role := e.Role
		if role != "user" && role != "agent" {
			role = "system"
		}
		m.transcript = append(m.transcript, entry{role, e.Text})
	}
	return m
}

// runElapsed is the time the current run has been going, shown next to the working spinner (e.g.
// " (12s)" or " (1:05)"). Empty when no run has started. It reads the last clock tick.
func (m Model) runElapsed() string {
	if m.runStarted.IsZero() {
		return ""
	}
	d := m.now.Sub(m.runStarted)
	if d < 0 {
		d = 0
	}
	s := int(d.Seconds())
	if s < 60 {
		return fmt.Sprintf(" (%ds)", s)
	}
	return fmt.Sprintf(" (%d:%02d)", s/60, s%60)
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
