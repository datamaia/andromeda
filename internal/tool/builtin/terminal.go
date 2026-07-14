package builtin

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
	"github.com/datamaia/andromeda/internal/terminal"
)

// TerminalRun runs a shell command through the Terminal Engine and returns its captured output
// and exit status. It requires the `execute` permission; the Tool Runtime mediates. Phase: MVP.
type TerminalRun struct {
	Engine *terminal.Engine
}

// NewTerminalRun builds the terminal.run tool over a Terminal Engine.
func NewTerminalRun(e *terminal.Engine) TerminalRun { return TerminalRun{Engine: e} }

// Describe returns the terminal_run tool descriptor.
func (TerminalRun) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "terminal_run", Namespace: "terminal", Version: "1",
		Description: "Run a shell command and capture its output",
		InputSchema: []byte(`{"type":"object","required":["command"],"properties":{"command":{"type":"string"},"dir":{"type":"string"}}}`),
		Permissions: []core.Permission{core.PermExecute}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

// Validate requires a non-empty command.
func (TerminalRun) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in struct{ Command string }
	if err := json.Unmarshal(input, &in); err != nil || in.Command == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"command is required"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests execute permission scoped to the command's leading program. The full
// command line rides along in Command so a configured allowlist can match on argv, not just the
// binary name.
func (TerminalRun) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in struct{ Command string }
	_ = json.Unmarshal(input, &in)
	prog := in.Command
	if i := strings.IndexByte(prog, ' '); i > 0 {
		prog = prog[:i]
	}
	return []ports.PermissionQuery{{
		Permission: core.PermExecute, Scope: core.ScopeCommand,
		Subject: prog, Command: in.Command,
	}}, nil
}

// Execute runs the command through the Terminal Engine and returns its captured output and exit status.
func (t TerminalRun) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in struct {
		Command string `json:"command"`
		Dir     string `json:"dir"`
	}
	_ = json.Unmarshal(req.Input, &in)
	id, err := t.Engine.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", in.Command}, Dir: in.Dir})
	if err != nil {
		return errEvent("could not start command: " + err.Error()), nil
	}
	st, err := t.Engine.Stream(ctx, id)
	if err != nil {
		return errEvent(err.Error()), nil
	}
	var out strings.Builder
	for {
		c, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			return errEvent(err.Error()), nil
		}
		out.Write(c.Data)
	}
	outcome, _ := t.Engine.Wait(ctx, id)
	result, _ := json.Marshal(map[string]any{"output": out.String(), "exit_code": outcome.ExitCode, "status": outcome.Status})
	events := []ports.ToolEvent{{Kind: "output", Text: out.String()}}
	if outcome.Status == "succeeded" {
		events = append(events, ports.ToolEvent{Kind: "terminal", Terminal: true, Outcome: "success", Text: string(result)})
	} else {
		events = append(events, ports.ToolEvent{Kind: "terminal", Terminal: true, Outcome: "error", Text: string(result)})
	}
	return streams.Slice(events), nil
}

// Cancel is a no-op; the command is bounded by the Execute context.
func (TerminalRun) Cancel(context.Context, core.ULID) error { return nil }
