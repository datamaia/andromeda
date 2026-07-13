package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/eventbus"
	"github.com/datamaia/andromeda/internal/ports"
)

// Tools is the mediated tool surface the agent uses (satisfied by the Tool Runtime).
type Tools interface {
	Describe(ctx context.Context, name string) (ports.ToolDescriptor, error)
	Invoke(ctx context.Context, name string, subjectCtx ports.PermissionQuery, input ports.JSON) (ports.Stream[ports.ToolEvent], error)
}

// Engine runs the agent loop.
type Engine struct {
	provider ports.ProviderPort
	tools    Tools
	sessions ports.SessionStorePort // optional; nil disables persistence
	bus      ports.EventBusPort     // optional
}

// New builds an Agent Engine. sessions and bus may be nil.
func New(provider ports.ProviderPort, tools Tools, sessions ports.SessionStorePort, bus ports.EventBusPort) *Engine {
	return &Engine{provider: provider, tools: tools, sessions: sessions, bus: bus}
}

// RunInput parameterizes a run.
type RunInput struct {
	SessionID     core.ULID
	WorkspaceID   core.ULID
	Goal          string
	System        string
	Model         string
	Effort        string // reasoning effort (minimal|low|medium|high); empty leaves it to the provider
	ToolNames     []string
	MaxIterations int

	// History seeds the conversation with prior turns so a run continues an existing session (the
	// System prompt is prepended once and the new Goal appended after this history).
	History []ports.Message

	// Sink, when non-null, receives incremental run events (streamed content deltas, tool-call
	// starts, tool results) as they happen, so a caller can render a live transcript. It is called
	// synchronously from the run goroutine; a slow sink backs the run up (intended flow control).
	Sink func(RunEvent)
}

// RunEvent is one incremental step of a streaming run delivered to RunInput.Sink.
type RunEvent struct {
	Kind       string     // "content" | "tool_call" | "tool_result"
	Content    string     // content delta (Kind == "content")
	ToolName   string     // tool being called / that produced a result
	ToolInput  ports.JSON // tool arguments (Kind == "tool_call")
	ToolResult string     // tool outcome (Kind == "tool_result")
}

// RunResult is the outcome of a run.
type RunResult struct {
	RunID        core.ULID
	State        string // frozen Run states (Volume 2 ch 09)
	FinalText    string
	Iterations   int
	ToolCalls    int
	InputTokens  int
	OutputTokens int

	// Messages is the conversation after the run (excluding the system prompt): prior history plus
	// this turn's user goal, any tool exchanges, and the final assistant reply. Feed it back as the
	// next run's History to continue the session.
	Messages []ports.Message
}

// DefaultMaxIterations bounds the loop when RunInput does not set one (Volume 4 default).
const DefaultMaxIterations = 50

// Run executes the plan–act–observe loop to completion or the iteration budget.
func (e *Engine) Run(ctx context.Context, in RunInput) (RunResult, error) {
	maxIter := in.MaxIterations
	if maxIter <= 0 {
		maxIter = DefaultMaxIterations
	}
	runID := core.NewULID()
	res := RunResult{RunID: runID, State: "running"}

	e.emit("run.started", runID, in.SessionID)
	e.persistRun(ctx, runID, in.SessionID)

	messages := make([]ports.Message, 0, len(in.History)+4)
	if in.System != "" {
		messages = append(messages, textMsg("system", in.System))
	}
	messages = append(messages, in.History...)
	messages = append(messages, textMsg("user", in.Goal))

	// Reasoning effort rides on ModelParams.Extra so adapters that support it (OpenAI-compatible
	// reasoning models) can forward "reasoning_effort"; adapters that don't simply ignore it.
	params := ports.ModelParams{}
	if in.Effort != "" {
		params.Extra = map[string]any{"reasoning_effort": in.Effort}
	}

	decls, err := e.declarations(ctx, in.ToolNames)
	if err != nil {
		res.State = "failed"
		return res, err
	}
	subjectCtx := ports.PermissionQuery{SessionID: in.SessionID, WorkspaceID: in.WorkspaceID}

	for res.Iterations < maxIter {
		if err := ctx.Err(); err != nil {
			res.State = "cancelled"
			return res, err
		}
		res.Iterations++

		resp, err := e.chat(ctx, ports.ChatRequest{Model: in.Model, Messages: messages, Tools: decls, Params: params}, in.Sink)
		if err != nil {
			res.State = "failed"
			e.emit("run.failed", runID, in.SessionID)
			return res, err
		}
		res.InputTokens += resp.Usage.InputTokens
		res.OutputTokens += resp.Usage.OutputTokens
		e.appendRecord(ctx, runID, "turn", resp)

		if len(resp.ToolCalls) == 0 {
			res.FinalText = textOf(resp.Message)
			// Record the final assistant turn so the returned conversation can seed the next run.
			messages = append(messages, ports.Message{Role: "assistant",
				Parts: []ports.ContentPart{{Type: "text", Text: res.FinalText}}})
			res.Messages = conversationMessages(messages)
			res.State = "completed"
			e.emit("run.completed", runID, in.SessionID)
			return res, nil
		}

		// Providers correlate a tool result with its call by id; synthesize one when the model omits
		// it so the round-trip is always valid (some OpenAI-compatible servers reject empty ids).
		for i := range resp.ToolCalls {
			if resp.ToolCalls[i].ID == "" {
				resp.ToolCalls[i].ID = fmt.Sprintf("call_%d_%d", res.Iterations, i)
			}
		}
		// Re-attach the tool calls to the assistant message so the next request carries them (the
		// provider needs the assistant's tool_calls immediately followed by each tool's result).
		asst := resp.Message
		for _, tc := range resp.ToolCalls {
			asst.Parts = append(asst.Parts, ports.ContentPart{
				Type: "tool_call", ToolCallID: tc.ID, ToolName: tc.Name, ToolInput: tc.Input,
			})
		}
		messages = append(messages, asst)
		for _, tc := range resp.ToolCalls {
			res.ToolCalls++
			emit(in.Sink, RunEvent{Kind: "tool_call", ToolName: tc.Name, ToolInput: tc.Input})
			result := e.runTool(ctx, tc, subjectCtx)
			emit(in.Sink, RunEvent{Kind: "tool_result", ToolName: tc.Name, ToolResult: result})
			e.appendRecord(ctx, runID, "tool_result", map[string]string{"tool": tc.Name, "result": result})
			messages = append(messages, ports.Message{
				Role:  "tool",
				Parts: []ports.ContentPart{{Type: "tool_result", ToolCallID: tc.ID, ToolName: tc.Name, Text: result}},
			})
		}
	}

	res.State = "failed"
	e.emit("run.failed", runID, in.SessionID)
	return res, &ports.PortError{Code: "E-AGT-001", Category: "agent", Severity: "error",
		Message: "run exceeded the iteration budget", Detail: fmt.Sprintf("max=%d", maxIter)}
}

func (e *Engine) runTool(ctx context.Context, tc ports.ToolCall, subj ports.PermissionQuery) string {
	st, err := e.tools.Invoke(ctx, tc.Name, subj, tc.Input)
	if err != nil {
		return "error: " + err.Error()
	}
	defer func() { _ = st.Close() }()
	var last string
	for {
		ev, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			return "error: " + err.Error()
		}
		if ev.Terminal {
			if ev.Outcome == "error" {
				return "error: " + ev.Text
			}
			last = ev.Text
		}
	}
	return last
}

// chat runs one turn. With no sink it is a plain non-streaming completion. With a sink it streams:
// content deltas are forwarded as they arrive (live typing) while text, tool calls, and usage are
// accumulated into the same ChatResponse the non-streaming path returns. A provider that cannot
// stream (ChatStream errors) transparently falls back to a single Chat call.
func (e *Engine) chat(ctx context.Context, req ports.ChatRequest, sink func(RunEvent)) (ports.ChatResponse, error) {
	if sink == nil {
		return e.provider.Chat(ctx, req)
	}
	stream, err := e.provider.ChatStream(ctx, req)
	if err != nil {
		resp, cerr := e.provider.Chat(ctx, req) // provider has no streaming — degrade to one shot
		if cerr == nil {
			if t := textOf(resp.Message); t != "" {
				sink(RunEvent{Kind: "content", Content: t})
			}
		}
		return resp, cerr
	}
	defer func() { _ = stream.Close() }()

	var text strings.Builder
	var calls []ports.ToolCall
	var usage ports.Usage
	for {
		ev, err := stream.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			return ports.ChatResponse{}, err
		}
		switch ev.Kind {
		case "content":
			text.WriteString(ev.ContentDelta)
			sink(RunEvent{Kind: "content", Content: ev.ContentDelta})
		case "tool_call":
			if ev.ToolCall != nil {
				calls = append(calls, *ev.ToolCall)
			}
		default: // "usage" | "terminal"
			if ev.Usage != nil {
				usage = *ev.Usage
			}
		}
	}
	return ports.ChatResponse{
		Message:   ports.Message{Role: "assistant", Parts: []ports.ContentPart{{Type: "text", Text: text.String()}}},
		ToolCalls: calls,
		Usage:     usage,
	}, nil
}

// emit forwards a run event to the sink when one is wired.
func emit(sink func(RunEvent), ev RunEvent) {
	if sink != nil {
		sink(ev)
	}
}

func (e *Engine) declarations(ctx context.Context, names []string) ([]ports.ToolDeclaration, error) {
	var out []ports.ToolDeclaration
	for _, n := range names {
		d, err := e.tools.Describe(ctx, n)
		if err != nil {
			return nil, err
		}
		out = append(out, ports.ToolDeclaration{Name: d.Name, Description: d.Description, InputSchema: d.InputSchema})
	}
	return out, nil
}

func (e *Engine) persistRun(ctx context.Context, runID, _ core.ULID) {
	if e.sessions == nil {
		return
	}
	_ = e.sessions.AppendRunRecords(ctx, runID, []ports.RunRecord{{Kind: "run_started"}})
}

func (e *Engine) appendRecord(ctx context.Context, runID core.ULID, kind string, payload any) {
	if e.sessions == nil {
		return
	}
	raw, _ := json.Marshal(payload)
	_ = e.sessions.AppendRunRecords(ctx, runID, []ports.RunRecord{{Kind: kind, Payload: raw}})
}

func (e *Engine) emit(name string, runID, sessionID core.ULID) {
	if e.bus == nil {
		return
	}
	_ = e.bus.Publish(context.Background(), eventbus.NewEvent(name, "agent-engine",
		eventbus.WithRun(runID), eventbus.WithSession(sessionID)))
}

// conversationMessages drops a leading system prompt so the returned history can be re-seeded (the
// system prompt is prepended fresh on the next run).
func conversationMessages(msgs []ports.Message) []ports.Message {
	if len(msgs) > 0 && msgs[0].Role == "system" {
		return msgs[1:]
	}
	return msgs
}

func textMsg(role, text string) ports.Message {
	return ports.Message{Role: role, Parts: []ports.ContentPart{{Type: "text", Text: text}}}
}

func textOf(m ports.Message) string {
	var s string
	for _, p := range m.Parts {
		if p.Type == "" || p.Type == "text" {
			s += p.Text
		}
	}
	return s
}
