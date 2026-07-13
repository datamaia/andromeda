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
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// toolAccum reassembles a streamed tool call: the id/name arrive in the first delta and the
// arguments stream in fragments across later deltas, keyed by choice index.
type toolAccum struct {
	id   string
	name string
	args strings.Builder
}

type sseStream struct {
	body       io.ReadCloser
	sc         *bufio.Scanner
	mu         sync.Mutex
	usage      ports.Usage
	tools      map[int]*toolAccum
	order      []int             // tool indices in first-seen order
	pending    []ports.ChatEvent // completed tool-call events queued to deliver before terminal
	flushed    bool              // tool calls have been converted to pending events
	terminated bool              // terminal event already delivered
	closed     bool
}

func newSSEStream(body io.ReadCloser) *sseStream {
	sc := bufio.NewScanner(body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &sseStream{body: body, sc: sc, tools: map[int]*toolAccum{}}
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
	// Drain any completed tool-call events queued at end-of-stream before the terminal event.
	if len(s.pending) > 0 {
		ev := s.pending[0]
		s.pending = s.pending[1:]
		return ev, nil
	}
	for s.sc.Scan() {
		line := strings.TrimSpace(s.sc.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			return s.finish(), nil
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return ports.ChatEvent{}, provider.Unavailable("malformed stream chunk")
		}
		if chunk.Usage != nil {
			s.usage = ports.Usage{InputTokens: chunk.Usage.PromptTokens, OutputTokens: chunk.Usage.CompletionTokens, CostBasis: "reported"}
		}
		if len(chunk.Choices) > 0 {
			d := chunk.Choices[0].Delta
			for _, tc := range d.ToolCalls {
				acc := s.tools[tc.Index]
				if acc == nil {
					acc = &toolAccum{}
					s.tools[tc.Index] = acc
					s.order = append(s.order, tc.Index)
				}
				if tc.ID != "" {
					acc.id = tc.ID
				}
				if tc.Function.Name != "" {
					acc.name = tc.Function.Name
				}
				acc.args.WriteString(tc.Function.Arguments)
			}
			if d.Content != "" {
				return ports.ChatEvent{Kind: "content", ContentDelta: d.Content}, nil
			}
		}
		// otherwise keep scanning for the next meaningful event
	}
	if err := s.sc.Err(); err != nil {
		return ports.ChatEvent{}, provider.Unavailable("stream read error")
	}
	// Stream ended without an explicit [DONE]; flush tool calls, then a terminal event.
	return s.finish(), nil
}

// finish flushes accumulated tool calls as ChatEvents (once), then delivers the terminal event.
func (s *sseStream) finish() ports.ChatEvent {
	if !s.flushed {
		s.flushed = true
		for _, idx := range s.order {
			t := s.tools[idx]
			s.pending = append(s.pending, ports.ChatEvent{
				Kind:     "tool_call",
				ToolCall: &ports.ToolCall{ID: t.id, Name: t.name, Input: ports.JSON(t.args.String())},
			})
		}
		if len(s.pending) > 0 {
			ev := s.pending[0]
			s.pending = s.pending[1:]
			return ev
		}
	}
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
