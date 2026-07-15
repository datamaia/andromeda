package tui

import (
	"strings"
	"testing"
)

func lastText(m Model) string {
	if len(m.transcript) == 0 {
		return ""
	}
	return m.transcript[len(m.transcript)-1].text
}

// /help and /commands used to run the same handler. They must now be distinct: /help is the
// orientation guide, /commands is the exhaustive reference.
func TestHelpAndCommandsDiffer(t *testing.T) {
	m := New("ollama", "llama3", nil)
	hm, _ := cmdHelp(m, "")
	cm, _ := cmdCommands(m, "")
	help := lastText(hm.(Model))
	cmds := lastText(cm.(Model))
	if help == "" || cmds == "" {
		t.Fatal("both commands should produce output")
	}
	if help == cmds {
		t.Fatal("/help and /commands must not be identical")
	}
	if !strings.Contains(cmds, "/permission") || !strings.Contains(cmds, "/skills") {
		t.Fatalf("/commands should enumerate every command, got: %q", cmds)
	}
	if !strings.Contains(help, "modes") || !strings.Contains(help, "keybindings") {
		t.Fatalf("/help should read like a guide, got: %q", help)
	}
	if !strings.Contains(help, "/commands") {
		t.Fatalf("/help should point at /commands, got: %q", help)
	}
}
