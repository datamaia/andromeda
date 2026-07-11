package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// scriptedProvider stands in for a live LLM: it emits a preset response sequence.
type scriptedProvider struct {
	responses []ports.ChatResponse
	calls     int
}

func (p *scriptedProvider) Chat(context.Context, ports.ChatRequest) (ports.ChatResponse, error) {
	r := p.responses[min(p.calls, len(p.responses)-1)]
	p.calls++
	return r, nil
}
func (p *scriptedProvider) ChatStream(context.Context, ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	return nil, nil
}
func (p *scriptedProvider) Embed(context.Context, ports.EmbedRequest) (ports.EmbedResponse, error) {
	return ports.EmbedResponse{}, nil
}
func (p *scriptedProvider) DiscoverModels(context.Context) ([]ports.ModelDescriptor, error) {
	return nil, nil
}
func (p *scriptedProvider) Capabilities(context.Context, string) (ports.CapabilitySet, error) {
	return nil, nil
}
func (p *scriptedProvider) CountTokens(context.Context, ports.TokenCountRequest) (ports.TokenCount, error) {
	return ports.TokenCount{}, nil
}

func assistantMsg(text string, calls ...ports.ToolCall) ports.ChatResponse {
	return ports.ChatResponse{
		Message:   ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: text}}},
		ToolCalls: calls,
	}
}

func TestRunAgentReadsAFileEndToEnd(t *testing.T) {
	ctx := context.Background()
	ws := t.TempDir()
	if err := os.WriteFile(filepath.Join(ws, "hello.txt"), []byte("secret-marker-42"), 0o600); err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(map[string]string{"path": filepath.Join(ws, "hello.txt")})
	prov := &scriptedProvider{responses: []ports.ChatResponse{
		assistantMsg("reading the file", ports.ToolCall{ID: "1", Name: "fs_read", Input: input}),
		assistantMsg("the file contains secret-marker-42"),
	}}

	res, err := RunAgent(ctx, RunAgentOptions{
		WorkspaceRoot: ws, Goal: "what is in hello.txt?", Model: "m", Provider: prov,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.State != "completed" {
		t.Fatalf("state = %q", res.State)
	}
	if res.ToolCalls != 1 {
		t.Errorf("tool calls = %d", res.ToolCalls)
	}
	if res.FinalText != "the file contains secret-marker-42" {
		t.Errorf("final = %q", res.FinalText)
	}
}

func TestRunAgentWriteDeniedWithoutAllowWrite(t *testing.T) {
	ctx := context.Background()
	ws := t.TempDir()
	target := filepath.Join(ws, "out.txt")
	input, _ := json.Marshal(map[string]string{"path": target, "content": "x"})
	// The scripted model asks to write, then reports what happened.
	prov := &scriptedProvider{responses: []ports.ChatResponse{
		assistantMsg("writing", ports.ToolCall{ID: "1", Name: "fs_write", Input: input}),
		assistantMsg("done"),
	}}
	// fs_write is not even registered without AllowWrite, so the tool call resolves to an error;
	// the agent still completes. The file must not exist.
	if _, err := RunAgent(ctx, RunAgentOptions{WorkspaceRoot: ws, Goal: "write a file", Model: "m", Provider: prov}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatal("file was written without --allow-write")
	}
}

func TestBuildProviderErrors(t *testing.T) {
	if _, err := BuildProvider(ProviderSpec{Name: "anthropic"}); err == nil {
		t.Error("anthropic without key should error")
	}
	if _, err := BuildProvider(ProviderSpec{Name: "bogus"}); err == nil {
		t.Error("unknown provider should error")
	}
	if _, err := BuildProvider(ProviderSpec{Name: "ollama"}); err != nil {
		t.Errorf("ollama should build: %v", err)
	}
}
