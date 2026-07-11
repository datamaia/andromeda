package ports

import (
	"context"

	"github.com/datamaia/andromeda/internal/core"
)

// ToolPort is the uniform execution contract for every tool regardless of origin
// (Principle 4). The Tool Runtime mediates all access. Contract owner: Volume 6. Errors:
// E-TOOL. A tool error is data (a Tool Result), not a transport failure; Execute returns a
// Go error only when the invocation could not start or the stream broke.
type ToolPort interface {
	Describe(ctx context.Context) (ToolDescriptor, error)
	Validate(ctx context.Context, input JSON) (ValidationResult, error)
	Execute(ctx context.Context, req ToolExecuteRequest) (Stream[ToolEvent], error)
	Cancel(ctx context.Context, invocationID ULID) error
}

// ToolDescriptor is the full tool declaration: identity, version, schemas, permissions,
// timeouts, limits, origin, trust level. Full contract: Volume 6.
type ToolDescriptor struct {
	Name         string
	Namespace    string
	Version      string
	Description  string
	InputSchema  JSON
	OutputSchema JSON
	Permissions  []core.Permission
	Origin       string // "builtin" | "plugin" | "mcp"
	TrustLevel   string
}

// ValidationResult reports whether input passed schema and semantic validation.
type ValidationResult struct {
	Valid    bool
	Findings []string
}

// ToolExecuteRequest carries the invocation ULID, validated input, granted permission set,
// and effective limits for one Tool Invocation.
type ToolExecuteRequest struct {
	InvocationID ULID
	Input        JSON
	Granted      []core.Permission
	Limits       ToolLimits
}

// ToolLimits are the effective per-invocation limits (timeout, output size, etc.).
type ToolLimits struct {
	TimeoutSeconds int
	MaxOutputBytes int64
}

// ToolEvent is one ordered event from a running invocation; exactly one terminal event
// becomes the Tool Result.
type ToolEvent struct {
	Kind     string // "progress" | "output" | "log" | "artifact" | "terminal"
	Text     string
	Artifact *ArtifactRef
	Terminal bool
	Outcome  string // on terminal: "success" | "error"
}

// ArtifactRef references a durable output produced by a tool.
type ArtifactRef struct {
	ID   ULID
	Path Path
	Kind string
}

// TerminalPort is PTY-backed command execution with streaming capture, signals, and resize.
// All executions enter through SandboxPort (ExecuteIn); the ExecutionID links the two ports.
// Contract owner: Volume 6. Errors: E-TOOL.
type TerminalPort interface {
	Execute(ctx context.Context, spec CommandSpec) (ExecutionID, error)
	Stream(ctx context.Context, id ExecutionID) (Stream[TerminalChunk], error)
	Write(ctx context.Context, id ExecutionID, input []byte) error
	Signal(ctx context.Context, id ExecutionID, sig SignalName) error
	Resize(ctx context.Context, id ExecutionID, cols int, rows int) error
	Wait(ctx context.Context, id ExecutionID) (CommandOutcome, error)
}

// SignalName is a portable signal name; the PAL maps it to platform mechanics.
type SignalName string

const (
	SignalInterrupt SignalName = "interrupt"
	SignalTerminate SignalName = "terminate"
	SignalKill      SignalName = "kill"
)

// TerminalChunk is one ordered output chunk (stdout/stderr tagged or merged PTY bytes).
type TerminalChunk struct {
	Stream    string // "stdout" | "stderr" | "pty"
	Data      []byte
	Truncated bool
}

// CommandOutcome is the recorded result of a command execution.
type CommandOutcome struct {
	Status     string // "succeeded" | "failed" | "timed_out" | "killed"
	ExitCode   int
	Signal     string
	DurationMS int64
}
