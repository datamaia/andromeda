package sandbox

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// Engine implements ports.SandboxPort with process-level controls (ADR-021 MVP layer).
type Engine struct {
	mu      sync.Mutex
	handles map[core.ULID]*handle
	environ []string // base environment to filter from (defaults to os.Environ)
}

type handle struct {
	spec        ports.SandboxSpec
	policy      ports.SandboxPolicy
	execs       map[ports.ExecutionID]*execution
	containment string
}

// New returns a sandbox Engine.
func New() *Engine {
	return &Engine{handles: map[core.ULID]*handle{}, environ: os.Environ()}
}

// NewWithEnv returns an Engine that filters from the given base environment (for tests).
func NewWithEnv(environ []string) *Engine {
	return &Engine{handles: map[core.ULID]*handle{}, environ: environ}
}

var _ ports.SandboxPort = (*Engine)(nil)

// Prepare allocates an execution environment and returns a handle.
func (e *Engine) Prepare(ctx context.Context, spec ports.SandboxSpec) (ports.SandboxHandle, error) {
	if err := ctx.Err(); err != nil {
		return ports.SandboxHandle{}, err
	}
	id := core.NewULID()
	e.mu.Lock()
	e.handles[id] = &handle{
		spec:        spec,
		execs:       map[ports.ExecutionID]*execution{},
		containment: "process",
	}
	e.mu.Unlock()
	return ports.SandboxHandle{ID: id, ContainmentLevel: "process"}, nil
}

// ApplyPolicy binds or tightens the effective policy on a handle.
func (e *Engine) ApplyPolicy(ctx context.Context, sb ports.SandboxHandle, policy ports.SandboxPolicy) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	h, ok := e.handles[sb.ID]
	if !ok {
		return secErr("E-SEC-030", "unknown sandbox handle")
	}
	h.policy = policy
	// Resolve the effective containment level (ADR-021): honor a request for OS-level isolation
	// only when the platform supports it; a downgrade to process-level is explicit and recorded.
	switch policy.Isolation {
	case "os", "auto":
		if osIsolationSupported() {
			h.containment = "os"
		} else {
			h.containment = "process"
		}
	default:
		h.containment = "process"
	}
	return nil
}

// ContainmentLevel returns the effective containment level of a handle (observable state).
func (e *Engine) ContainmentLevel(sb ports.SandboxHandle) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if h, ok := e.handles[sb.ID]; ok {
		return h.containment
	}
	return ""
}

// ExecuteIn starts a command inside the sandbox after policy checks and env filtering, and
// returns its execution ID. Direct process spawning outside this path is prohibited.
func (e *Engine) ExecuteIn(ctx context.Context, sb ports.SandboxHandle, cmd ports.CommandSpec) (ports.ExecutionID, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	e.mu.Lock()
	h, ok := e.handles[sb.ID]
	e.mu.Unlock()
	if !ok {
		return "", secErr("E-SEC-030", "unknown sandbox handle")
	}

	if !commandAllowed(h.policy, cmd.Program) {
		return "", secErr("E-SEC-031", "command denied by sandbox policy")
	}
	// Working directory must be the sandbox dir or within an allowed path.
	dir := cmd.Dir
	if dir == "" {
		dir = h.spec.WorkingDir
	}
	if dir != "" && dir != h.spec.WorkingDir {
		allowed := append(append([]string{}, h.policy.WritePaths...), h.policy.ReadPaths...)
		if h.spec.WorkingDir != "" {
			allowed = append(allowed, h.spec.WorkingDir)
		}
		if !pathWithin(dir, allowed) {
			return "", secErr("E-SEC-032", "working directory outside sandbox path policy")
		}
	}

	env := filterEnv(e.environ, h.spec.EnvAllow)
	// Per-command env overrides that are explicitly allow-listed pass through.
	for k, v := range cmd.Env {
		if containsExact(h.spec.EnvAllow, k) {
			env = append(env, k+"="+v)
		}
	}

	// Apply OS-level isolation when the handle's effective containment is "os".
	program, progArgs := cmd.Program, cmd.Args
	if h.containment == "os" {
		program, progArgs = wrapOSIsolation(h.policy, cmd.Program, cmd.Args)
	}
	wrapped := cmd
	wrapped.Program, wrapped.Args = program, progArgs

	timeout := time.Duration(h.policy.TimeLimitSec) * time.Second
	ex, err := startExecution(ctx, wrapped, dir, env, timeout)
	if err != nil {
		return "", secErr("E-SEC-033", "failed to start sandboxed command")
	}
	id := core.NewULID()
	e.mu.Lock()
	h.execs[id] = ex
	e.mu.Unlock()
	return id, nil
}

// Teardown terminates the full process tree of a handle's executions and releases it.
func (e *Engine) Teardown(ctx context.Context, sb ports.SandboxHandle) error {
	e.mu.Lock()
	h, ok := e.handles[sb.ID]
	if ok {
		delete(e.handles, sb.ID)
	}
	e.mu.Unlock()
	if !ok {
		return nil // idempotent
	}
	for _, ex := range h.execs {
		ex.killTree()
	}
	return nil
}

// Wait blocks until a sandboxed execution terminates and returns its outcome. Not part of the
// port (I/O and waiting flow through TerminalPort); provided for composition and tests.
func (e *Engine) Wait(ctx context.Context, sb ports.SandboxHandle, id ports.ExecutionID) (ports.CommandOutcome, error) {
	e.mu.Lock()
	h, ok := e.handles[sb.ID]
	var ex *execution
	if ok {
		ex = h.execs[id]
	}
	e.mu.Unlock()
	if ex == nil {
		return ports.CommandOutcome{}, secErr("E-SEC-030", "unknown execution")
	}
	return ex.wait(ctx), nil
}

func secErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "security", Severity: "error", Message: msg}
}
