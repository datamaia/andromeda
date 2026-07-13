package tui

import (
	"context"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The @-mention menu completes workspace file paths inline: typing "@" followed by a fragment opens
// a ranked list of matching files; Tab/Enter inserts "@path". The file list is supplied by the
// driver (Actions.Files) and cached on first use, keeping the TUI free of filesystem imports for
// discovery. atMenuMax bounds how many candidates are shown.
const atMenuMax = 10

// atToken returns the trailing "@fragment" token being typed (the text after the last whitespace),
// or ok=false when the input does not end in such a token.
func atToken(input string) (frag string, ok bool) {
	i := strings.LastIndexAny(input, " \t")
	last := input[i+1:]
	if strings.HasPrefix(last, "@") {
		return last[1:], true
	}
	return "", false
}

// loadFiles fetches and caches the workspace file list from the driver (once per session).
func (m Model) loadFiles() Model {
	m.filesLoaded = true
	if m.actions.Files != nil {
		m.fileList = m.actions.Files(context.Background())
	}
	return m
}

// matchFiles ranks the cached file list against a fragment in three tiers — whole-path prefix, then
// basename prefix, then path substring — capped at atMenuMax. An empty fragment returns the first
// atMenuMax files.
func (m Model) matchFiles(frag string) []string {
	q := strings.ToLower(frag)
	var pathPre, basePre, substr []string
	for _, f := range m.fileList {
		lf := strings.ToLower(f)
		base := strings.ToLower(filepath.Base(f))
		switch {
		case q == "" || strings.HasPrefix(lf, q):
			pathPre = append(pathPre, f)
		case strings.HasPrefix(base, q):
			basePre = append(basePre, f)
		case strings.Contains(lf, q):
			substr = append(substr, f)
		}
	}
	out := append(append(pathPre, basePre...), substr...)
	if len(out) > atMenuMax {
		out = out[:atMenuMax]
	}
	return out
}

// atActive reports whether the @-mention menu should own navigation keys: an "@fragment" token is
// being typed and at least one file matches it.
func (m Model) atActive() bool {
	frag, ok := atToken(m.input)
	if !ok {
		return false
	}
	return len(m.matchFiles(frag)) > 0
}

// acceptFileMention replaces the trailing "@fragment" with "@path " so the goal can reference the file.
func (m Model) acceptFileMention(path string) Model {
	i := strings.LastIndexAny(m.input, " \t")
	m.input = m.input[:i+1] + "@" + path + " "
	m.atCursor = 0
	return m
}

// handleAtKey drives the @-mention menu: ↑/↓ move, Tab/Enter insert the highlighted path, Esc drops
// the "@fragment". Returns handled=false for keys it does not consume (so typing keeps filtering).
func (m Model) handleAtKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	frag, _ := atToken(m.input)
	files := m.matchFiles(frag)
	switch msg.Code {
	case tea.KeyEscape:
		i := strings.LastIndexAny(m.input, " \t")
		m.input = m.input[:i+1] // drop the @fragment, closing the menu
		m.atCursor = 0
		return m, nil, true
	case tea.KeyUp:
		if m.atCursor > 0 {
			m.atCursor--
		}
		return m, nil, true
	case tea.KeyDown:
		if m.atCursor < len(files)-1 {
			m.atCursor++
		}
		return m, nil, true
	case tea.KeyTab, tea.KeyEnter:
		if len(files) > 0 {
			m = m.acceptFileMention(files[clamp(m.atCursor, len(files))])
		}
		return m, nil, true
	}
	return m, nil, false
}

// renderAtMenu draws the ranked file candidates above the prompt (reusing the palette's windowed row
// renderer for consistent scrolling).
func (m Model) renderAtMenu() string {
	frag, _ := atToken(m.input)
	files := m.matchFiles(frag)
	if len(files) == 0 {
		return ""
	}
	return m.renderMenuRows(files, clamp(m.atCursor, len(files)))
}
