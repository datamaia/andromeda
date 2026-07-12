package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	state      string
	width      int
	height     int
	respond    Responder
	quitting   bool
}

// New builds a session Model.
func New(provider, model string, respond Responder) Model {
	m := Model{
		styles:   DefaultStyles(),
		provider: provider,
		model:    model,
		state:    "ready",
		respond:  respond,
		width:    80,
		height:   24,
	}
	m.transcript = append(m.transcript, entry{"system", "andromeda — " + Tagline})
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyEnter:
		return m.submit()
	case tea.KeyBackspace:
		if n := len(m.input); n > 0 {
			m.input = m.input[:n-1]
		}
	case tea.KeyRunes, tea.KeySpace:
		m.input += string(msg.Runes)
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

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	for _, e := range m.transcript {
		switch e.role {
		case "user":
			b.WriteString(m.styles.User.Render("you ▸ ") + e.text + "\n")
		case "agent":
			b.WriteString(m.styles.Agent.Render("andromeda ▸ ") + e.text + "\n")
		default:
			b.WriteString(m.styles.Muted.Render(e.text) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(m.styles.Prompt.Render("❯ ") + m.input + "▏\n")
	b.WriteString(m.statusBar())
	return b.String()
}

func (m Model) statusBar() string {
	left := fmt.Sprintf(" %s · %s · %s ", m.provider, m.model, m.state)
	help := m.styles.Muted.Render("  enter: send · esc: quit")
	bar := m.styles.StatusBar.Render(left)
	return lipgloss.JoinHorizontal(lipgloss.Left, bar, help)
}

// Transcript returns the transcript lines as role:text pairs (for tests and export).
func (m Model) Transcript() []string {
	out := make([]string, 0, len(m.transcript))
	for _, e := range m.transcript {
		out = append(out, e.role+": "+e.text)
	}
	return out
}
