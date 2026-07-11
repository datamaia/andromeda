package openaicompat

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

func msg(role, text string) ports.Message {
	return ports.Message{Role: role, Parts: []ports.ContentPart{{Type: "text", Text: text}}}
}

func TestChatParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("missing auth header")
		}
		var req chatReq
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "gpt-x" || len(req.Messages) != 1 {
			t.Errorf("request = %+v", req)
		}
		io.WriteString(w, `{"choices":[{"message":{"content":"hi there"}}],"usage":{"prompt_tokens":5,"completion_tokens":2}}`)
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, APIKey: "k", Client: &provider.Client{HTTP: srv.Client()}})
	resp, err := a.Chat(context.Background(), ports.ChatRequest{Model: "gpt-x", Messages: []ports.Message{msg("user", "hello")}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message.Parts[0].Text != "hi there" {
		t.Errorf("content = %q", resp.Message.Parts[0].Text)
	}
	if resp.Usage.InputTokens != 5 || resp.Usage.OutputTokens != 2 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

func TestChatStreamParsesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n")
		io.WriteString(w, "data: {\"usage\":{\"prompt_tokens\":3,\"completion_tokens\":1}}\n\n")
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, Client: &provider.Client{HTTP: srv.Client()}})
	st, err := a.ChatStream(context.Background(), ports.ChatRequest{Model: "m", Messages: []ports.Message{msg("user", "hi")}})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	var content strings.Builder
	var sawTerminal bool
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch ev.Kind {
		case "content":
			content.WriteString(ev.ContentDelta)
		case "terminal":
			sawTerminal = true
			if ev.Usage == nil || ev.Usage.OutputTokens != 1 {
				t.Errorf("terminal usage = %+v", ev.Usage)
			}
		}
	}
	if content.String() != "Hello" {
		t.Errorf("streamed content = %q", content.String())
	}
	if !sawTerminal {
		t.Error("expected a terminal event")
	}
}

func TestAuthErrorMapsToE_PROV_002(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"bad key"}`)
	}))
	defer srv.Close()
	a := New(Config{BaseURL: srv.URL, Client: &provider.Client{HTTP: srv.Client()}})
	_, err := a.Chat(context.Background(), ports.ChatRequest{Model: "m"})
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != provider.CodeAuth || pe.Retryable {
		t.Fatalf("want non-retryable E-PROV-002, got %v", err)
	}
}

func TestCapabilitiesAndCountTokens(t *testing.T) {
	a := New(Config{BaseURL: "http://x"})
	caps, _ := a.Capabilities(context.Background(), "m")
	if !caps.Has("tool_calling") {
		t.Error("expected tool_calling capability")
	}
	if _, err := a.CountTokens(context.Background(), ports.TokenCountRequest{}); err == nil {
		t.Error("CountTokens should be unavailable")
	}
}
