package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// The $-mention menu invokes any workspace resource that guides the agent, inline: typing "$" then a
// fragment opens a ranked list of everything discovered across .agents/.claude/.codex/.agent and
// .andromeda (plus .windsurf/.cursor for workflows) — skills, workflows, custom commands, and the
// workspace maps (ontology, graph, memory). Tab/Enter inserts "$name"; on submit every "$name" token
// is expanded, folding that resource's guidance into the agent run so it proceeds with that context.
// The list is supplied by the driver (Actions.Mentions) and cached on first use, keeping the TUI
// free of filesystem imports. This is the $-invocation convention Codex popularized, generalized
// beyond skills to everything that enriches or guides the agent.
const mentionMenuMax = 10

// Mention is one $-invokable workspace resource surfaced to the TUI for completion and expansion.
type Mention struct {
	Kind        string // skill | workflow | command | ontology | graph | memory
	Name        string
	Description string
	Path        string
	Body        string // guidance folded into the run when the mention is invoked
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

// loadMentions fetches and caches the workspace mention list from the driver (once per session).
func (m Model) loadMentions() Model {
	m.mentionsLoaded = true
	if m.actions.Mentions != nil {
		m.mentionList = m.actions.Mentions(context.Background())
	}
	return m
}

// matchMentions ranks the cached mentions against a fragment — name prefix first, then
// name/description substring — capped at mentionMenuMax. An empty fragment returns the first
// mentionMenuMax mentions (in discovery order: skills, workflows, commands, then maps).
func (m Model) matchMentions(frag string) []Mention {
	q := strings.ToLower(frag)
	var namePre, sub []Mention
	for _, s := range m.mentionList {
		ln := strings.ToLower(s.Name)
		switch {
		case q == "" || strings.HasPrefix(ln, q):
			namePre = append(namePre, s)
		case strings.Contains(ln, q) || strings.Contains(strings.ToLower(s.Description), q):
			sub = append(sub, s)
		}
	}
	out := append(namePre, sub...)
	if len(out) > mentionMenuMax {
		out = out[:mentionMenuMax]
	}
	return out
}

// mentionActive reports whether the $-mention menu should own navigation keys: a "$fragment" token
// is being typed and at least one mention matches it.
func (m Model) mentionActive() bool {
	frag, ok := dollarToken(m.input)
	if !ok {
		return false
	}
	return len(m.matchMentions(frag)) > 0
}

// acceptMention replaces the trailing "$fragment" with "$name " so the goal can invoke the resource.
func (m Model) acceptMention(name string) Model {
	i := strings.LastIndexAny(m.input, " \t")
	m.input = m.input[:i+1] + "$" + name + " "
	m.mentionCursor = 0
	return m
}

// handleMentionKey drives the $-mention menu: ↑/↓ move, Tab/Enter insert the highlighted mention,
// Esc drops the "$fragment". Returns handled=false for keys it does not consume (so typing keeps
// filtering).
func (m Model) handleMentionKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	frag, _ := dollarToken(m.input)
	items := m.matchMentions(frag)
	switch msg.Code {
	case tea.KeyEscape:
		i := strings.LastIndexAny(m.input, " \t")
		m.input = m.input[:i+1] // drop the $fragment, closing the menu
		m.mentionCursor = 0
		return m, nil, true
	case tea.KeyUp:
		if m.mentionCursor > 0 {
			m.mentionCursor--
		}
		return m, nil, true
	case tea.KeyDown:
		if m.mentionCursor < len(items)-1 {
			m.mentionCursor++
		}
		return m, nil, true
	case tea.KeyTab, tea.KeyEnter:
		if len(items) > 0 {
			m = m.acceptMention(items[clamp(m.mentionCursor, len(items))].Name)
		}
		return m, nil, true
	}
	return m, nil, false
}

// renderMentionMenu draws the ranked candidates above the prompt (reusing the palette's windowed row
// renderer), each labeled "$name · kind · description".
func (m Model) renderMentionMenu() string {
	frag, _ := dollarToken(m.input)
	items := m.matchMentions(frag)
	if len(items) == 0 {
		return ""
	}
	labels := make([]string, len(items))
	for i, s := range items {
		label := "$" + s.Name + "  · " + s.Kind
		if s.Description != "" {
			label += " ·  " + s.Description
		}
		labels[i] = label
	}
	return m.renderMentionRows(labels, clamp(m.mentionCursor, len(items)))
}

// renderMentionRows renders labels through the shared windowed row renderer but truncates each to the
// terminal width first, so a long description can't wrap and break the fixed footer layout.
func (m Model) renderMentionRows(labels []string, cur int) string {
	out := make([]string, len(labels))
	limit := max(20, m.width-6)
	for i, l := range labels {
		out[i] = ansi.Truncate(l, limit, "…")
	}
	return m.renderMenuRows(out, cur)
}

// mentionBlock renders one invoked mention's guidance under a kind-appropriate header, or "" when it
// has no body.
func mentionBlock(mn Mention) string {
	body := strings.TrimSpace(mn.Body)
	if body == "" {
		return ""
	}
	var header string
	switch mn.Kind {
	case "workflow":
		header = "## Workflow: " + mn.Name + " — follow these steps"
	case "command":
		header = "## Command: " + mn.Name
	case "ontology":
		header = "## Workspace ontology map"
	case "graph":
		header = "## Workspace graph map"
	case "memory":
		header = "## Workspace memory"
	default:
		header = "## Skill: " + mn.Name
	}
	return header + "\n" + body
}

// expandMentions rewrites a goal that references resources with "$name" by appending their guidance,
// so the agent runs with that context. Returns the augmented goal and the mentions that were
// activated (in first-mention order, de-duplicated), so the caller can record each activation.
func (m Model) expandMentions(goal string) (string, []Mention) {
	if len(m.mentionList) == 0 {
		return goal, nil
	}
	byName := make(map[string]Mention, len(m.mentionList))
	for _, mn := range m.mentionList {
		byName[strings.ToLower(mn.Name)] = mn
	}
	var used []Mention
	var blocks []string
	seen := map[string]bool{}
	for _, tok := range strings.Fields(goal) {
		if !strings.HasPrefix(tok, "$") {
			continue
		}
		name := strings.ToLower(strings.Trim(tok[1:], ".,;:!?)"))
		mn, ok := byName[name]
		if !ok || seen[mn.Name] {
			continue
		}
		seen[mn.Name] = true
		used = append(used, mn)
		if b := mentionBlock(mn); b != "" {
			blocks = append(blocks, b)
		}
	}
	if len(blocks) == 0 {
		return goal, used
	}
	return goal + "\n\n---\nActivated context — apply the following:\n\n" + strings.Join(blocks, "\n\n"), used
}
