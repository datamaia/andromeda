package agent

import (
	"context"
	"encoding/json"
	"fmt"

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
	ToolNames     []string
	MaxIterations int
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

	messages := make([]ports.Message, 0, 4)
	if in.System != "" {
		messages = append(messages, textMsg("system", in.System))
	}
	messages = append(messages, textMsg("user", in.Goal))

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

		resp, err := e.provider.Chat(ctx, ports.ChatRequest{Model: in.Model, Messages: messages, Tools: decls})
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
			res.State = "completed"
			e.emit("run.completed", runID, in.SessionID)
			return res, nil
		}

		messages = append(messages, resp.Message)
		for _, tc := range resp.ToolCalls {
			res.ToolCalls++
			result := e.runTool(ctx, tc, subjectCtx)
			e.appendRecord(ctx, runID, "tool_result", map[string]string{"tool": tc.Name, "result": result})
			messages = append(messages, ports.Message{
				Role:  "tool",
				Parts: []ports.ContentPart{{Type: "text", Text: fmt.Sprintf("%s => %s", tc.Name, result)}},
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
	defer st.Close()
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

func (e *Engine) persistRun(ctx context.Context, runID, sessionID core.ULID) {
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
