package main

import (
	"context"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/settingstore"
)

// cannedProvider returns a fixed assistant reply (no tool calls), enough to drive a one-shot
// summarization through app.RunAgent.
type cannedProvider struct{ reply string }

func (p cannedProvider) Chat(context.Context, ports.ChatRequest) (ports.ChatResponse, error) {
	return ports.ChatResponse{Message: ports.Message{Role: "assistant",
		Parts: []ports.ContentPart{{Type: "text", Text: p.reply}}}}, nil
}
func (cannedProvider) ChatStream(context.Context, ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	return nil, nil
}
func (cannedProvider) Embed(context.Context, ports.EmbedRequest) (ports.EmbedResponse, error) {
	return ports.EmbedResponse{}, nil
}
func (cannedProvider) DiscoverModels(context.Context) ([]ports.ModelDescriptor, error) {
	return nil, nil
}
func (cannedProvider) Capabilities(context.Context, string) (ports.CapabilitySet, error) {
	return nil, nil
}
func (cannedProvider) CountTokens(context.Context, ports.TokenCountRequest) (ports.TokenCount, error) {
	return ports.TokenCount{}, nil
}

func longHistory(n int) []ports.Message {
	var h []ports.Message
	for i := 0; i < n; i++ {
		h = append(h, textMessage("user", "do step"), textMessage("assistant", "did step"))
	}
	return h
}

func TestCompactHistoryReplacesWithSummaryPair(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	s := &tuiSession{ctx: context.Background(), wd: t.TempDir(), sessionID: "s1",
		cfg: tuiConfig{provider: "groq", model: "x"}, prov: cannedProvider{reply: "SUMMARY-42"}}
	s.history = longHistory(3) // 6 messages

	status, ok := s.compactHistory(context.Background())
	if !ok || !strings.Contains(status, "compacted") {
		t.Fatalf("compact: ok=%v status=%q", ok, status)
	}
	// History collapses to a user→assistant pair carrying the summary (valid role alternation).
	if len(s.history) != 2 || s.history[0].Role != "user" || s.history[1].Role != "assistant" {
		t.Fatalf("history not a summary pair: %+v", s.history)
	}
	if !strings.Contains(messageText(s.history[1]), "SUMMARY-42") {
		t.Fatalf("summary not carried: %q", messageText(s.history[1]))
	}
}

func TestCompactShortConversationNoOp(t *testing.T) {
	s := &tuiSession{ctx: context.Background(), wd: t.TempDir(), prov: cannedProvider{reply: "x"}}
	s.history = []ports.Message{textMessage("user", "hi")}
	if status, ok := s.compactHistory(context.Background()); ok || !strings.Contains(status, "already short") {
		t.Fatalf("short compact: ok=%v status=%q", ok, status)
	}
}

func TestMaybeAutoCompact(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	s := &tuiSession{ctx: context.Background(), wd: t.TempDir(), sessionID: "s1",
		cfg: tuiConfig{provider: "groq", model: "x"}, prov: cannedProvider{reply: "SUM"}}
	s.history = longHistory(7) // 14 messages, 7 user turns (past the 6-turn threshold)

	// Off by default → no notice, history untouched.
	if got := s.maybeAutoCompact(context.Background()); got != "" {
		t.Fatalf("auto-compact should be off: %q", got)
	}
	if len(s.history) != 14 {
		t.Fatal("history changed while autocompact off")
	}
	// On + past threshold → compacts and returns a notice.
	s.autoCompact = true
	notice := s.maybeAutoCompact(context.Background())
	if !strings.Contains(notice, "auto-compact") {
		t.Fatalf("expected auto-compact notice: %q", notice)
	}
	if len(s.history) != 2 {
		t.Fatalf("history not compacted: %d msgs", len(s.history))
	}
	// Below threshold → no-op even when on.
	s.history = longHistory(1)
	if got := s.maybeAutoCompact(context.Background()); got != "" {
		t.Fatalf("below threshold should no-op: %q", got)
	}
}

func TestAutoCompactActionTogglesAndPersists(t *testing.T) {
	wd := t.TempDir()
	s := &tuiSession{ctx: context.Background(), wd: wd}

	if got := s.autoCompactAction(context.Background(), "on"); !strings.Contains(got, "ON") {
		t.Fatalf("on: %q", got)
	}
	if st, _ := settingstore.Load(wd); !st.AutoCompact {
		t.Fatal("on should persist AutoCompact=true")
	}
	if !s.autoCompact {
		t.Fatal("on should update the in-memory flag")
	}
	if got := s.autoCompactAction(context.Background(), "status"); !strings.Contains(got, "ON") {
		t.Fatalf("status: %q", got)
	}
	if got := s.autoCompactAction(context.Background(), ""); !strings.Contains(got, "OFF") {
		t.Fatalf("bare toggle should flip to OFF: %q", got)
	}
	if st, _ := settingstore.Load(wd); st.AutoCompact {
		t.Fatal("toggle should persist AutoCompact=false")
	}
	if got := s.autoCompactAction(context.Background(), "frob"); !strings.Contains(got, "usage") {
		t.Fatalf("unknown: %q", got)
	}
}

func TestRenderConversationSkipsNonChat(t *testing.T) {
	msgs := []ports.Message{
		textMessage("user", "hello"),
		textMessage("assistant", "hi"),
		textMessage("tool", "ignored"),
		textMessage("user", "   "),
	}
	got := renderConversation(msgs)
	if !strings.Contains(got, "User: hello") || !strings.Contains(got, "Assistant: hi") {
		t.Fatalf("render missing chat turns: %q", got)
	}
	if strings.Contains(got, "ignored") {
		t.Fatalf("render should drop tool/system turns: %q", got)
	}
}
