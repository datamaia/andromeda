package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// scriptedProvider returns a preset sequence of responses, one per Chat call, recording the last
// request so tests can assert what the engine sent.
type scriptedProvider struct {
	responses []ports.ChatResponse
	calls     int
	lastReq   ports.ChatRequest
}

func (p *scriptedProvider) Chat(_ context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	p.lastReq = req
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

// fakeTools records invocations and returns a preset result.
type fakeTools struct {
	result  string
	invoked []string
}

func (f *fakeTools) Describe(_ context.Context, name string) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{Name: name, Description: "fake", InputSchema: []byte("{}")}, nil
}
func (f *fakeTools) Invoke(_ context.Context, name string, _ ports.PermissionQuery, _ ports.JSON) (ports.Stream[ports.ToolEvent], error) {
	f.invoked = append(f.invoked, name)
	return streams.Slice([]ports.ToolEvent{{Kind: "terminal", Terminal: true, Outcome: "success", Text: f.result}}), nil
}

func assistant(text string, calls ...ports.ToolCall) ports.ChatResponse {
	return ports.ChatResponse{
		Message:   ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: text}}},
		ToolCalls: calls,
		Usage:     ports.Usage{InputTokens: 10, OutputTokens: 5},
	}
}

func TestLoopExecutesToolThenFinishes(t *testing.T) {
	ctx := context.Background()
	input, _ := json.Marshal(map[string]string{"path": "/x"})
	prov := &scriptedProvider{responses: []ports.ChatResponse{
		assistant("let me read the file", ports.ToolCall{ID: "1", Name: "fs_read", Input: input}),
		assistant("the file says hello"),
	}}
	tools := &fakeTools{result: `{"content":"hello"}`}
	e := New(prov, tools, nil, nil)

	res, err := e.Run(ctx, RunInput{Goal: "read /x", Model: "m", ToolNames: []string{"fs_read"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.State != "completed" {
		t.Fatalf("state = %q", res.State)
	}
	if res.FinalText != "the file says hello" {
		t.Errorf("final = %q", res.FinalText)
	}
	if res.Iterations != 2 || res.ToolCalls != 1 {
		t.Errorf("iterations=%d toolCalls=%d", res.Iterations, res.ToolCalls)
	}
	if len(tools.invoked) != 1 || tools.invoked[0] != "fs_read" {
		t.Errorf("invoked = %v", tools.invoked)
	}
	if res.InputTokens != 20 || res.OutputTokens != 10 {
		t.Errorf("usage accumulation wrong: %+v", res)
	}
}

func TestLoopFinishesImmediatelyWithoutTools(t *testing.T) {
	ctx := context.Background()
	prov := &scriptedProvider{responses: []ports.ChatResponse{assistant("done, no tools needed")}}
	e := New(prov, &fakeTools{}, nil, nil)
	res, err := e.Run(ctx, RunInput{Goal: "hi", Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Iterations != 1 || res.ToolCalls != 0 || res.FinalText != "done, no tools needed" {
		t.Fatalf("res = %+v", res)
	}
}

func TestBudgetExhaustionFails(t *testing.T) {
	ctx := context.Background()
	input, _ := json.Marshal(map[string]string{"path": "/x"})
	// The model always asks for a tool → never finishes.
	prov := &scriptedProvider{responses: []ports.ChatResponse{
		assistant("again", ports.ToolCall{ID: "1", Name: "fs_read", Input: input}),
	}}
	e := New(prov, &fakeTools{result: "x"}, nil, nil)
	res, err := e.Run(ctx, RunInput{Goal: "loop", Model: "m", ToolNames: []string{"fs_read"}, MaxIterations: 3})
	if err == nil {
		t.Fatal("expected a budget-exhaustion error")
	}
	if res.State != "failed" || res.Iterations != 3 {
		t.Fatalf("res = %+v", res)
	}
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-AGT-001" {
		t.Fatalf("want E-AGT-001, got %v", err)
	}
}

func TestCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	prov := &scriptedProvider{responses: []ports.ChatResponse{assistant("x")}}
	e := New(prov, &fakeTools{}, nil, nil)
	res, err := e.Run(ctx, RunInput{Goal: "g", Model: "m"})
	if err == nil || res.State != "cancelled" {
		t.Fatalf("expected cancellation, got state=%q err=%v", res.State, err)
	}
}

// A run seeds the prior conversation, forwards reasoning effort, and returns the continued history.
func TestRunSeedsHistoryAndReturnsConversation(t *testing.T) {
	ctx := context.Background()
	prior := []ports.Message{
		{Role: "user", Parts: []ports.ContentPart{{Type: "text", Text: "first goal"}}},
		{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: "first reply"}}},
	}
	prov := &scriptedProvider{responses: []ports.ChatResponse{assistant("second reply")}}
	e := New(prov, &fakeTools{}, nil, nil)

	res, err := e.Run(ctx, RunInput{Goal: "second goal", Model: "m", History: prior, Effort: "high"})
	if err != nil {
		t.Fatal(err)
	}
	// The provider saw the prior history plus the new goal (2 + 1).
	if got := len(prov.lastReq.Messages); got != 3 {
		t.Fatalf("provider saw %d messages, want 3 (2 history + new goal)", got)
	}
	// Reasoning effort rode along on ModelParams.Extra.
	if prov.lastReq.Params.Extra["reasoning_effort"] != "high" {
		t.Errorf("effort not forwarded: %v", prov.lastReq.Params.Extra)
	}
	// The returned conversation is history + user goal + final assistant (2 + 1 + 1).
	if got := len(res.Messages); got != 4 {
		t.Fatalf("res.Messages = %d, want 4", got)
	}
	if last := res.Messages[3]; last.Role != "assistant" || textOf(last) != "second reply" {
		t.Errorf("final message = %+v", last)
	}
}
