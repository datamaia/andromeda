package ollama

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

func TestChatParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %s", r.URL.Path)
		}
		io.WriteString(w, `{"message":{"role":"assistant","content":"local reply"},"prompt_eval_count":8,"eval_count":3}`)
	}))
	defer srv.Close()

	a := New(Config{BaseURL: srv.URL, Client: &provider.Client{HTTP: srv.Client()}})
	resp, err := a.Chat(context.Background(), ports.ChatRequest{
		Model:    "llama",
		Messages: []ports.Message{{Role: "user", Parts: []ports.ContentPart{{Text: "hi"}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message.Parts[0].Text != "local reply" {
		t.Errorf("content = %q", resp.Message.Parts[0].Text)
	}
	if resp.Usage.InputTokens != 8 || resp.Usage.OutputTokens != 3 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}

func TestDiscoverModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("path = %s", r.URL.Path)
		}
		io.WriteString(w, `{"models":[{"name":"llama3"},{"name":"qwen"}]}`)
	}))
	defer srv.Close()
	a := New(Config{BaseURL: srv.URL, Client: &provider.Client{HTTP: srv.Client()}})
	models, err := a.DiscoverModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != 2 || models[0].ID != "llama3" {
		t.Fatalf("models = %v", models)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	a := New(Config{})
	caps, _ := a.Capabilities(context.Background(), "m")
	if !caps.Has("chat") {
		t.Error("expected chat capability")
	}
	if _, err := a.ChatStream(context.Background(), ports.ChatRequest{}); err == nil {
		t.Error("Ollama ChatStream should be unavailable at MVP")
	}
}
