package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// fakeAdapter returns preset responses/errors.
type fakeAdapter struct {
	name  string
	reply string
	err   error
	calls *int
}

func (f *fakeAdapter) Chat(_ context.Context, _ ports.ChatRequest) (ports.ChatResponse, error) {
	if f.calls != nil {
		*f.calls++
	}
	if f.err != nil {
		return ports.ChatResponse{}, f.err
	}
	return ports.ChatResponse{Message: ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Text: f.reply}}}}, nil
}
func (f *fakeAdapter) ChatStream(context.Context, ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	return nil, f.err
}
func (f *fakeAdapter) Embed(context.Context, ports.EmbedRequest) (ports.EmbedResponse, error) {
	return ports.EmbedResponse{}, f.err
}
func (f *fakeAdapter) DiscoverModels(context.Context) ([]ports.ModelDescriptor, error) {
	return []ports.ModelDescriptor{{ID: f.name + "-model"}}, nil
}
func (f *fakeAdapter) Capabilities(context.Context, string) (ports.CapabilitySet, error) {
	return nil, nil
}
func (f *fakeAdapter) CountTokens(context.Context, ports.TokenCountRequest) (ports.TokenCount, error) {
	return ports.TokenCount{}, nil
}

func TestRouterFailsOverOnRetryableError(t *testing.T) {
	var pCalls, fCalls int
	primary := &fakeAdapter{name: "primary", err: provErr(CodeConnectivity, "down", "", true, nil), calls: &pCalls}
	fallback := &fakeAdapter{name: "fallback", reply: "from-fallback", calls: &fCalls}

	var notice *ChangeNotice
	r := NewRouter(Named{"primary", primary}, []Named{{"fallback", fallback}}, func(n ChangeNotice) { notice = &n })
	resp, err := r.Chat(context.Background(), ports.ChatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Message.Parts[0].Text != "from-fallback" {
		t.Errorf("reply = %q", resp.Message.Parts[0].Text)
	}
	if pCalls != 1 || fCalls != 1 {
		t.Errorf("calls: primary=%d fallback=%d", pCalls, fCalls)
	}
	if notice == nil || notice.To != "fallback" {
		t.Errorf("expected a fallback change notice, got %+v", notice)
	}
}

func TestRouterDoesNotFailOverOnAuthError(t *testing.T) {
	var pCalls, fCalls int
	primary := &fakeAdapter{name: "primary", err: provErr(CodeAuth, "bad key", "", false, nil), calls: &pCalls}
	fallback := &fakeAdapter{name: "fallback", reply: "x", calls: &fCalls}
	r := NewRouter(Named{"primary", primary}, []Named{{"fallback", fallback}}, nil)
	_, err := r.Chat(context.Background(), ports.ChatRequest{})
	var pe *ports.PortError
	if !errors.As(err, &pe) || pe.Code != CodeAuth {
		t.Fatalf("want auth error surfaced, got %v", err)
	}
	if fCalls != 0 {
		t.Error("must not fail over on a non-retryable auth error (would mask misconfiguration)")
	}
}

func TestRouterDiscoverAggregates(t *testing.T) {
	r := NewRouter(Named{"a", &fakeAdapter{name: "a"}}, []Named{{"b", &fakeAdapter{name: "b"}}}, nil)
	models, _ := r.DiscoverModels(context.Background())
	if len(models) != 2 {
		t.Fatalf("models = %v", models)
	}
}
