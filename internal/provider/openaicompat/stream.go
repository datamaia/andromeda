package openaicompat

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/provider"
)

// ChatStream performs a streaming completion, parsing Server-Sent Events into ChatEvents.
func (a *Adapter) ChatStream(ctx context.Context, req ports.ChatRequest) (ports.Stream[ports.ChatEvent], error) {
	body, err := a.client.PostStream(ctx, "/chat/completions", buildRequest(req, true))
	if err != nil {
		return nil, err
	}
	return newSSEStream(body), nil
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type sseStream struct {
	body       io.ReadCloser
	sc         *bufio.Scanner
	mu         sync.Mutex
	usage      ports.Usage
	terminated bool // terminal event already delivered
	closed     bool
}

func newSSEStream(body io.ReadCloser) *sseStream {
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &sseStream{body: body, sc: sc}
}

func (s *sseStream) Next(ctx context.Context) (ports.ChatEvent, error) {
	if err := ctx.Err(); err != nil {
		return ports.ChatEvent{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.terminated {
		return ports.ChatEvent{}, ports.ErrEndOfStream
	}
	for s.sc.Scan() {
		line := strings.TrimSpace(s.sc.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return s.terminal(), nil
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return ports.ChatEvent{}, provider.Unavailable("malformed stream chunk")
		}
		if chunk.Usage != nil {
			s.usage = ports.Usage{InputTokens: chunk.Usage.PromptTokens, OutputTokens: chunk.Usage.CompletionTokens, CostBasis: "reported"}
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			return ports.ChatEvent{Kind: "content", ContentDelta: chunk.Choices[0].Delta.Content}, nil
		}
		// otherwise keep scanning for the next meaningful event
	}
	if err := s.sc.Err(); err != nil {
		return ports.ChatEvent{}, provider.Unavailable("stream read error")
	}
	// Stream ended without an explicit [DONE]; emit a terminal event once.
	return s.terminal(), nil
}

// terminal marks the stream terminated and returns the single terminal event.
func (s *sseStream) terminal() ports.ChatEvent {
	s.terminated = true
	usage := s.usage
	return ports.ChatEvent{Kind: "terminal", Terminal: true, Usage: &usage}
}

func (s *sseStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closeLocked()
}

func (s *sseStream) closeLocked() error {
	if s.closed {
		return nil
	}
	s.closed = true
	return s.body.Close()
}
