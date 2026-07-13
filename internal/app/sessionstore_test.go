package app

import (
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func msg(role, text string) ports.Message {
	return ports.Message{Role: role, Parts: []ports.ContentPart{{Type: "text", Text: text}}}
}

func TestSessionStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	old := sessionsDirFn
	sessionsDirFn = func() (string, error) { return dir, nil }
	defer func() { sessionsDirFn = old }()

	conv := []ports.Message{msg("user", "add a login page"), msg("assistant", "sure, here is the plan")}
	s := StoredSession{
		ID: NewSessionID(), Provider: "groq", Model: "llama-3.3-70b", Mode: "agent",
		UpdatedAt: "2026-07-13T10:00:00Z", Title: SessionTitle(conv), Messages: conv,
	}
	if err := SaveSession(s); err != nil {
		t.Fatal(err)
	}

	got, err := LoadSession(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Provider != "groq" || got.Model != "llama-3.3-70b" || len(got.Messages) != 2 {
		t.Fatalf("loaded session = %+v", got)
	}
	if got.Title != "add a login page" {
		t.Errorf("title = %q", got.Title)
	}
	if CountTurns(got.Messages) != 1 {
		t.Errorf("turns = %d, want 1", CountTurns(got.Messages))
	}

	list, err := ListSessions()
	if err != nil || len(list) != 1 {
		t.Fatalf("list = %v (err %v)", list, err)
	}
	if LatestSessionID() != s.ID {
		t.Errorf("latest = %q, want %q", LatestSessionID(), s.ID)
	}

	if err := RemoveSession(s.ID); err != nil {
		t.Fatal(err)
	}
	if list, _ := ListSessions(); len(list) != 0 {
		t.Errorf("session not removed: %v", list)
	}
}

// Sessions are listed newest-first by UpdatedAt.
func TestListSessionsOrder(t *testing.T) {
	dir := t.TempDir()
	old := sessionsDirFn
	sessionsDirFn = func() (string, error) { return dir, nil }
	defer func() { sessionsDirFn = old }()

	older := StoredSession{ID: "20260101-000000-aaaaaa", UpdatedAt: "2026-01-01T00:00:00Z", Messages: []ports.Message{msg("user", "a")}}
	newer := StoredSession{ID: "20260713-000000-bbbbbb", UpdatedAt: "2026-07-13T00:00:00Z", Messages: []ports.Message{msg("user", "b")}}
	if err := SaveSession(older); err != nil {
		t.Fatal(err)
	}
	if err := SaveSession(newer); err != nil {
		t.Fatal(err)
	}
	list, err := ListSessions()
	if err != nil || len(list) != 2 {
		t.Fatalf("list = %v (err %v)", list, err)
	}
	if list[0].ID != newer.ID {
		t.Errorf("newest first: got %q", list[0].ID)
	}
}
