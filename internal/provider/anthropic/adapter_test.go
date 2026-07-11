package anthropic

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

func TestChatParsesMessages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/messages" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "k" || r.Header.Get("anthropic-version") != APIVersion {
			t.Errorf("missing anthropic headers")
		}
		io.WriteString(w, `{"content":[{"type":"text","text":"claude says hi"}],"usage":{"input_tokens":7,"output_tokens":4}}`)
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, APIKey: "k", Client: &provider.Client{HTTP: srv.Client()}})
	resp, err := a.Chat(context.Background(), ports.ChatRequest{
		Model:    "claude-x",
		Messages: []ports.Message{{Role: "user", Parts: []ports.ContentPart{{Text: "hi"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message.Parts[0].Text != "claude says hi" {
		t.Errorf("content = %q", resp.Message.Parts[0].Text)
	}
	if resp.Usage.InputTokens != 7 || resp.Usage.OutputTokens != 4 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

func TestStreamParsesContentBlockDeltas(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"A\"}}\n\n")
		io.WriteString(w, "event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"text\":\"B\"}}\n\n")
		io.WriteString(w, "event: message_delta\ndata: {\"type\":\"message_delta\",\"usage\":{\"output_tokens\":2}}\n\n")
		io.WriteString(w, "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n")
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, APIKey: "k", Client: &provider.Client{HTTP: srv.Client()}})
	st, err := a.ChatStream(context.Background(), ports.ChatRequest{Model: "m", Messages: []ports.Message{{Role: "user", Parts: []ports.ContentPart{{Text: "x"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	var out string
	for {
		ev, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if ev.Kind == "content" {
			out += ev.ContentDelta
		}
	}
	if out != "AB" {
		t.Errorf("streamed = %q", out)
	}
}

func TestSystemMessageExtracted(t *testing.T) {
	r := build(ports.ChatRequest{
		Messages: []ports.Message{
			{Role: "system", Parts: []ports.ContentPart{{Text: "be terse"}}},
			{Role: "user", Parts: []ports.ContentPart{{Text: "hi"}}},
		},
	}, false)
	if r.System != "be terse" {
		t.Errorf("system = %q", r.System)
	}
	if len(r.Messages) != 1 || r.Messages[0].Role != "user" {
		t.Errorf("messages = %+v", r.Messages)
	}
	if r.MaxTokens != DefaultMaxTokens {
		t.Errorf("max_tokens = %d", r.MaxTokens)
	}
}
