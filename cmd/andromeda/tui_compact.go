package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/settingstore"
)

// compactSystem instructs the model to condense a conversation into a continuation brief. It runs
// as a plain completion (no tools, no approvals), so a summarization can never mutate the workspace.
const compactSystem = "You are a summarization engine. Condense the conversation the user provides " +
	"into a brief that lets a fresh session continue seamlessly. Preserve: decisions made, files and " +
	"code changed, facts and constraints established, and any open or pending tasks. Omit pleasantries " +
	"and meta-commentary. Do not use tools. Respond with only the summary."

// autoCompactTurns is the user-turn count at or above which /autocompact triggers before a turn.
const autoCompactTurns = 6

// compactHistory summarizes s.history in place, replacing it with a short user→assistant pair that
// carries the summary. The pair keeps provider role-alternation valid (Anthropic requires a
// user-first, strictly alternating transcript). A short conversation is left untouched. It returns a
// status line and whether the history was actually compacted.
func (s *tuiSession) compactHistory(ctx context.Context) (string, bool) {
	if app.CountTurns(s.history) < 2 {
		return "conversation is already short — nothing to compact", false
	}
	before := len(s.history)
	res, err := app.RunAgent(ctx, app.RunAgentOptions{
		WorkspaceRoot: s.wd,
		Provider:      s.prov,
		Model:         s.cfg.model,
		System:        compactSystem,
		Goal:          "Summarize this conversation so a fresh session can continue from it:\n\n" + renderConversation(s.history),
	})
	if err != nil {
		return "compact failed: " + err.Error(), false
	}
	summary := strings.TrimSpace(res.FinalText)
	if summary == "" {
		return "compact produced no summary; conversation left unchanged", false
	}
	s.history = []ports.Message{
		textMessage("user", "Summarize where we are so we can continue efficiently."),
		textMessage("assistant", summary),
	}
	s.persistSession("agent")
	return fmt.Sprintf("compacted %d messages into a summary — the agent keeps the gist, not the full log", before), true
}

// compactAction backs /compact (run off the UI thread by the TUI).
func (s *tuiSession) compactAction(ctx context.Context) string {
	status, _ := s.compactHistory(ctx)
	return status
}

// maybeAutoCompact compacts the history before a turn when auto-compaction is on and the
// conversation has grown past the threshold. It returns a notice to surface (empty when it did
// nothing) so the user always sees when their context was summarized. Called inside the run
// goroutine, so the summarization never blocks the UI thread.
func (s *tuiSession) maybeAutoCompact(ctx context.Context) string {
	if !s.autoCompact || app.CountTurns(s.history) < autoCompactTurns {
		return ""
	}
	status, ok := s.compactHistory(ctx)
	if !ok {
		return "" // a failed auto-compaction is silent; the turn proceeds with full context
	}
	return "auto-compact: " + status
}

// autoCompactAction backs /autocompact: it toggles (or sets) automatic compaction and persists the
// choice to the workspace settings store. Grammar: on | off | status | (bare toggles).
func (s *tuiSession) autoCompactAction(_ context.Context, args string) string {
	st, _ := settingstore.Load(s.wd)
	switch strings.TrimSpace(args) {
	case "", "toggle":
		st.AutoCompact = !st.AutoCompact
	case "on", "enable":
		st.AutoCompact = true
	case "off", "disable":
		st.AutoCompact = false
	case "status":
		return autoCompactStatus(st.AutoCompact)
	default:
		return "usage: /autocompact [on | off | status]"
	}
	if err := settingstore.Save(s.wd, st); err != nil {
		return "autocompact: " + err.Error()
	}
	s.autoCompact = st.AutoCompact
	return autoCompactStatus(st.AutoCompact) + "  ·  saved to " + relOr(s.wd, settingstore.Path(s.wd))
}

// autoCompactStatus renders the on/off state with a hint about the trigger.
func autoCompactStatus(on bool) string {
	if on {
		return fmt.Sprintf("autocompact ON — the conversation is summarized automatically past %d turns", autoCompactTurns)
	}
	return "autocompact OFF — compact manually with /compact"
}

// textMessage builds a single-part provider message.
func textMessage(role, text string) ports.Message {
	return ports.Message{Role: role, Parts: []ports.ContentPart{{Type: "text", Text: text}}}
}

// renderConversation flattens a stored transcript into labelled plain text for summarization,
// skipping empty and non-conversational (tool/system) turns.
func renderConversation(msgs []ports.Message) string {
	var b strings.Builder
	for _, m := range msgs {
		text := strings.TrimSpace(messageText(m))
		if text == "" {
			continue
		}
		switch m.Role {
		case "user":
			b.WriteString("User: ")
		case "assistant":
			b.WriteString("Assistant: ")
		default:
			continue
		}
		b.WriteString(text)
		b.WriteString("\n\n")
	}
	return strings.TrimSpace(b.String())
}
