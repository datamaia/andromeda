package builtin

import (
	"context"
	"encoding/json"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/terminal"
)

// ProcessControl manages long-running processes previously started by Andromeda executions
// through the Terminal Engine: list, inspect, signal, and terminate. Scope is strictly
// Andromeda-supervised process trees; arbitrary host processes are out of contract. Phase: Beta.
type ProcessControl struct {
	Engine *terminal.Engine
}

// NewProcessControl builds the process.control tool over a Terminal Engine.
func NewProcessControl(e *terminal.Engine) ProcessControl { return ProcessControl{Engine: e} }

// Describe returns the process_control tool descriptor.
func (ProcessControl) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "process_control", Namespace: "process", Version: "1",
		Description: "Manage Andromeda-supervised processes: list, inspect, signal, terminate",
		InputSchema: []byte(`{"type":"object","required":["operation"],"properties":{` +
			`"operation":{"type":"string","enum":["list","inspect","signal","terminate"]},` +
			`"execution_id":{"type":"string"},"signal":{"type":"string"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"processes":{"type":"array"},"delivered":{"type":"boolean"},"outcome":{"type":"string"}}}`),
		Permissions:  []core.Permission{core.PermProcessSpawn}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type processInput struct {
	Operation   string `json:"operation"`
	ExecutionID string `json:"execution_id"`
	Signal      string `json:"signal"`
}

// Validate requires an operation, and execution_id for inspect, signal, and terminate.
func (ProcessControl) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in processInput
	if err := json.Unmarshal(input, &in); err != nil || in.Operation == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"operation is required"}}, nil
	}
	if (in.Operation == "inspect" || in.Operation == "signal" || in.Operation == "terminate") && in.ExecutionID == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{in.Operation + " requires execution_id"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests host-scoped process_spawn access.
func (ProcessControl) Resources(ports.JSON) ([]ports.PermissionQuery, error) {
	return []ports.PermissionQuery{{Permission: core.PermProcessSpawn, Scope: core.ScopeHost, Subject: "local"}}, nil
}

// Execute lists, inspects, signals, or terminates supervised processes via the Terminal Engine.
func (t ProcessControl) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in processInput
	_ = json.Unmarshal(req.Input, &in)

	switch in.Operation {
	case "list":
		return okEvent(marshalProcesses(t.Engine.Snapshot())), nil
	case "inspect":
		s, err := t.Engine.SnapshotOne(in.ExecutionID)
		if err != nil {
			return errEvent(err.Error()), nil
		}
		return okEvent(marshalProcesses([]terminal.ExecutionSnapshot{s})), nil
	case "signal":
		sig := ports.SignalName(in.Signal)
		if sig == "" {
			sig = "TERM"
		}
		if err := t.Engine.Signal(ctx, in.ExecutionID, sig); err != nil {
			return errEvent(err.Error()), nil
		}
		out, _ := json.Marshal(map[string]any{"delivered": true, "outcome": "signaled"})
		return okEvent(string(out)), nil
	case "terminate":
		if err := t.Engine.Signal(ctx, in.ExecutionID, "KILL"); err != nil {
			return errEvent(err.Error()), nil
		}
		out, _ := json.Marshal(map[string]any{"delivered": true, "outcome": "terminated"})
		return okEvent(string(out)), nil
	default:
		return errEvent("unsupported operation: " + in.Operation), nil
	}
}

// Cancel is a no-op; each operation completes synchronously within Execute.
func (ProcessControl) Cancel(context.Context, core.ULID) error { return nil }

func marshalProcesses(snaps []terminal.ExecutionSnapshot) string {
	procs := make([]map[string]any, 0, len(snaps))
	for _, s := range snaps {
		procs = append(procs, map[string]any{
			"execution_id": s.ID, "running": s.Running, "state": s.Status, "pid": s.PID, "exit_code": s.ExitCode,
		})
	}
	out, _ := json.Marshal(map[string]any{"processes": procs})
	return string(out)
}
