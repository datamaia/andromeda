package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/datamaia/andromeda/internal/tui"
)

// defaultEditor is the fallback when neither $VISUAL nor $EDITOR is set.
func defaultEditor() string {
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vi"
}

// editorAction backs /editor: it writes the composer's current text to a temp file, opens it in the
// user's editor (suspending the TUI via tea.ExecProcess), and posts the composed prompt back as a
// tui.EditorMsg once the editor exits. The temp file is always cleaned up.
func (s *tuiSession) editorAction(seed string) tea.Cmd {
	f, err := os.CreateTemp("", "andromeda-prompt-*.md")
	if err != nil {
		return func() tea.Msg { return tui.EditorMsg{Err: err} }
	}
	path := f.Name()
	if seed != "" {
		_, _ = f.WriteString(seed)
	}
	_ = f.Close()

	name, args := editorCommand()
	args = append(args, path)
	cmd := exec.Command(name, args...) //nolint:gosec // editor name comes from the user's own $EDITOR
	return tea.ExecProcess(cmd, func(runErr error) tea.Msg {
		defer func() { _ = os.Remove(path) }()
		if runErr != nil {
			return tui.EditorMsg{Err: runErr}
		}
		data, err := os.ReadFile(path) //nolint:gosec // our own temp file
		if err != nil {
			return tui.EditorMsg{Err: err}
		}
		return tui.EditorMsg{Text: strings.TrimSpace(string(data))}
	})
}

// editorCommand resolves the editor to run: $VISUAL, then $EDITOR, then a sensible default. The
// value may include flags (e.g. "code --wait"), so it is split into a name and leading args.
func editorCommand() (string, []string) {
	ed := os.Getenv("VISUAL")
	if ed == "" {
		ed = os.Getenv("EDITOR")
	}
	if ed == "" {
		ed = defaultEditor()
	}
	fields := strings.Fields(ed)
	if len(fields) == 0 {
		return defaultEditor(), nil
	}
	return fields[0], fields[1:]
}
