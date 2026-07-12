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

// Responder produces the agent's reply to a user goal. Injected so the Model is testable and
// so the driver can wire it to the real agent.
type Responder func(goal string) string

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
}

// New builds a session Model.
func New(provider, model string, respond Responder) Model {
	start := time.Now()
	m := Model{
		styles:   DefaultStyles(),
		provider: provider,
		model:    model,
		state:    "ready",
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
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Mod&tea.ModCtrl != 0 && msg.Code == 'c':
		m.quitting = true
		return m, tea.Quit
	case msg.Code == tea.KeyEscape:
		m.quitting = true
		return m, tea.Quit
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
	m.input = ""
	if goal == "" {
		return m, nil
	}
	m.transcript = append(m.transcript, entry{"user", goal})
	reply := "(no responder configured)"
	if m.respond != nil {
		m.state = "running"
		reply = m.respond(goal)
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
	var b strings.Builder
	if m.atStart() {
		b.WriteString(m.Splash(m.width))
	}
	for _, e := range m.transcript {
		switch e.role {
		case "user":
			b.WriteString(m.styles.User.Render("you ▸ ") + e.text + "\n")
		case "agent":
			b.WriteString(m.styles.Agent.Render("andromeda ▸ ") + e.text + "\n")
		default:
			// The initial system line is folded into the splash on the start screen.
			if !m.atStart() {
				b.WriteString(m.styles.Muted.Render(e.text) + "\n")
			}
		}
	}
	b.WriteString("\n")
	b.WriteString(m.styles.Prompt.Render("❯ ") + m.input + "▏\n")
	b.WriteString(m.statusBar())
	return b.String()
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
	parts := []string{m.provider, m.model}
	if m.effort != "" {
		parts = append(parts, "effort "+m.effort)
	}
	parts = append(parts, m.uptime(), m.state)
	left := " " + strings.Join(parts, " · ") + " "
	help := m.styles.Muted.Render("  enter: send · esc: quit")
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
