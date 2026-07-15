package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// The $-mention menu invokes a workspace skill inline: typing "$" followed by a fragment opens a
// ranked list of discovered skills; Tab/Enter inserts "$name". On submit, every "$name" token is
// expanded — the referenced skill's instructions are folded into the agent run so it proceeds with
// that procedural knowledge (the convention Codex popularized). The skill list is supplied by the
// driver (Actions.Skills) and cached on first use, keeping the TUI free of filesystem imports.
const skillMenuMax = 10

// SkillNote is one discovered skill, as surfaced to the TUI for $-mention completion and invocation.
type SkillNote struct {
	Name        string
	Description string
	Path        string
	Body        string // instructions folded into the run when the skill is invoked
}

// dollarToken returns the trailing "$fragment" token being typed (the text after the last
// whitespace), or ok=false when the input does not end in such a token.
func dollarToken(input string) (frag string, ok bool) {
	i := strings.LastIndexAny(input, " \t")
	last := input[i+1:]
	if strings.HasPrefix(last, "$") {
		return last[1:], true
	}
	return "", false
}

// loadSkills fetches and caches the workspace skill list from the driver (once per session).
func (m Model) loadSkills() Model {
	m.skillsLoaded = true
	if m.actions.Skills != nil {
		m.skillList = m.actions.Skills(context.Background())
	}
	return m
}

// matchSkills ranks the cached skills against a fragment — name prefix first, then name/description
// substring — capped at skillMenuMax. An empty fragment returns the first skillMenuMax skills.
func (m Model) matchSkills(frag string) []SkillNote {
	q := strings.ToLower(frag)
	var namePre, sub []SkillNote
	for _, s := range m.skillList {
		ln := strings.ToLower(s.Name)
		switch {
		case q == "" || strings.HasPrefix(ln, q):
			namePre = append(namePre, s)
		case strings.Contains(ln, q) || strings.Contains(strings.ToLower(s.Description), q):
			sub = append(sub, s)
		}
	}
	out := append(namePre, sub...)
	if len(out) > skillMenuMax {
		out = out[:skillMenuMax]
	}
	return out
}

// skillActive reports whether the $-mention menu should own navigation keys: a "$fragment" token is
// being typed and at least one skill matches it.
func (m Model) skillActive() bool {
	frag, ok := dollarToken(m.input)
	if !ok {
		return false
	}
	return len(m.matchSkills(frag)) > 0
}

// acceptSkillMention replaces the trailing "$fragment" with "$name " so the goal can invoke the skill.
func (m Model) acceptSkillMention(name string) Model {
	i := strings.LastIndexAny(m.input, " \t")
	m.input = m.input[:i+1] + "$" + name + " "
	m.skillCursor = 0
	return m
}

// handleSkillKey drives the $-mention menu: ↑/↓ move, Tab/Enter insert the highlighted skill, Esc
// drops the "$fragment". Returns handled=false for keys it does not consume (so typing keeps filtering).
func (m Model) handleSkillKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	frag, _ := dollarToken(m.input)
	sk := m.matchSkills(frag)
	switch msg.Code {
	case tea.KeyEscape:
		i := strings.LastIndexAny(m.input, " \t")
		m.input = m.input[:i+1] // drop the $fragment, closing the menu
		m.skillCursor = 0
		return m, nil, true
	case tea.KeyUp:
		if m.skillCursor > 0 {
			m.skillCursor--
		}
		return m, nil, true
	case tea.KeyDown:
		if m.skillCursor < len(sk)-1 {
			m.skillCursor++
		}
		return m, nil, true
	case tea.KeyTab, tea.KeyEnter:
		if len(sk) > 0 {
			m = m.acceptSkillMention(sk[clamp(m.skillCursor, len(sk))].Name)
		}
		return m, nil, true
	}
	return m, nil, false
}

// renderSkillMenu draws the ranked skill candidates above the prompt (reusing the palette's windowed
// row renderer), each labeled "$name — description".
func (m Model) renderSkillMenu() string {
	frag, _ := dollarToken(m.input)
	sk := m.matchSkills(frag)
	if len(sk) == 0 {
		return ""
	}
	labels := make([]string, len(sk))
	for i, s := range sk {
		labels[i] = "$" + s.Name
		if s.Description != "" {
			labels[i] += "  —  " + s.Description
		}
	}
	return m.renderSkillRows(labels, clamp(m.skillCursor, len(sk)))
}

// renderSkillRows renders labels through the shared windowed row renderer but truncates each to the
// terminal width first, so a long skill description can't wrap and break the fixed footer layout.
func (m Model) renderSkillRows(labels []string, cur int) string {
	out := make([]string, len(labels))
	limit := max(20, m.width-6)
	for i, l := range labels {
		out[i] = ansi.Truncate(l, limit, "…")
	}
	return m.renderMenuRows(out, cur)
}

// expandSkillMentions rewrites a goal that references skills with "$name" by appending those skills'
// instructions, so the agent runs with the procedural context. Returns the augmented goal and the
// names of the skills that were activated (in first-mention order, de-duplicated).
func (m Model) expandSkillMentions(goal string) (string, []string) {
	if len(m.skillList) == 0 {
		return goal, nil
	}
	byName := make(map[string]SkillNote, len(m.skillList))
	for _, s := range m.skillList {
		byName[strings.ToLower(s.Name)] = s
	}
	var used []string
	var blocks []string
	seen := map[string]bool{}
	for _, tok := range strings.Fields(goal) {
		if !strings.HasPrefix(tok, "$") {
			continue
		}
		name := strings.ToLower(strings.Trim(tok[1:], ".,;:!?)"))
		s, ok := byName[name]
		if !ok || seen[s.Name] {
			continue
		}
		seen[s.Name] = true
		used = append(used, s.Name)
		if body := strings.TrimSpace(s.Body); body != "" {
			blocks = append(blocks, "## Skill: "+s.Name+"\n"+body)
		}
	}
	if len(blocks) == 0 {
		return goal, used
	}
	return goal + "\n\n---\nActivated skill instructions — follow these:\n\n" + strings.Join(blocks, "\n\n"), used
}
