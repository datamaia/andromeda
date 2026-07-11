package provider

import (
	"context"
	"errors"

	"github.com/datamaia/andromeda/internal/ports"
)

// Named pairs a provider label with an adapter.
type Named struct {
	Name    string
	Adapter ports.ProviderPort
}

// ChangeNotice reports a routing change (fallback) to the user layer (Transparent AI). The
// runtime wires this to the event bus / TUI; a nil notifier is allowed.
type ChangeNotice struct {
	From   string
	To     string
	Reason string
}

// Router implements ports.ProviderPort over a primary adapter and ordered fallbacks. It fails
// over only on retryable errors (connectivity, rate limit, server) and never on auth or
// bad-request errors, so a fallback can never mask a misconfiguration or run a costly retry on
// a request the provider correctly rejected.
type Router struct {
	primary  Named
	fallback []Named
	notify   func(ChangeNotice)
}

// NewRouter builds a Router. notify may be nil.
func NewRouter(primary Named, fallback []Named, notify func(ChangeNotice)) *Router {
	return &Router{primary: primary, fallback: fallback, notify: notify}
}

var _ ports.ProviderPort = (*Router)(nil)

func (r *Router) chain() []Named {
	return append([]Named{r.primary}, r.fallback...)
}

// retryable reports whether err is an E-PROV error marked retryable.
func retryable(err error) bool {
	var pe *ports.PortError
	if errors.As(err, &pe) {
		return pe.Retryable
	}
	return false
}

// Chat tries the primary then fallbacks on retryable failures.
func (r *Router) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	chain := r.chain()
	var lastErr error
	for i, n := range chain {
		resp, err := n.Adapter.Chat(ctx, req)
		if err == nil {
			if i > 0 && r.notify != nil {
				r.notify(ChangeNotice{From: chain[0].Name, To: n.Name, Reason: "fallback"})
			}
			return resp, nil
		}
		lastErr = err
		if !retryable(err) || ctx.Err() != nil {
			return ports.ChatResponse{}, err
		}
	}
	return ports.ChatResponse{}, lastErr
}

// ChatStream tries the primary then fallbacks on retryable failures to establish the stream.
func (r *Router) ChatStream(ctx context.Context, req ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	chain := r.chain()
	var lastErr error
	for i, n := range chain {
		st, err := n.Adapter.ChatStream(ctx, req)
		if err == nil {
			if i > 0 && r.notify != nil {
				r.notify(ChangeNotice{From: chain[0].Name, To: n.Name, Reason: "fallback"})
			}
			return st, nil
		}
		lastErr = err
		if !retryable(err) || ctx.Err() != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

// Embed dispatches to the primary (embeddings are not failed over — vectors must be consistent
// within one index generation).
func (r *Router) Embed(ctx context.Context, req ports.EmbedRequest) (ports.EmbedResponse, error) {
	return r.primary.Adapter.Embed(ctx, req)
}

// DiscoverModels aggregates models across the chain, primary first.
func (r *Router) DiscoverModels(ctx context.Context) ([]ports.ModelDescriptor, error) {
	var out []ports.ModelDescriptor
	seen := map[string]bool{}
	for _, n := range r.chain() {
		ms, err := n.Adapter.DiscoverModels(ctx)
		if err != nil {
			continue
		}
		for _, m := range ms {
			if !seen[m.ID] {
				seen[m.ID] = true
				out = append(out, m)
			}
		}
	}
	return out, nil
}

// Capabilities delegates to the primary.
func (r *Router) Capabilities(ctx context.Context, model string) (ports.CapabilitySet, error) {
	return r.primary.Adapter.Capabilities(ctx, model)
}

// CountTokens delegates to the primary.
func (r *Router) CountTokens(ctx context.Context, req ports.TokenCountRequest) (ports.TokenCount, error) {
	return r.primary.Adapter.CountTokens(ctx, req)
}
