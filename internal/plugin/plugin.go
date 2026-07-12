package plugin

import (
	"context"
	"io"
	"os/exec"
	"sync"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/mcp"
	"github.com/datamaia/andromeda/internal/ports"
)

// Frozen Plugin lifecycle states (Volume 2, chapter 09).
const (
	StateRegistered = "registered"
	StateStarting   = "starting"
	StateRunning    = "running"
	StateStopping   = "stopping"
	StateStopped    = "stopped"
	StateFailed     = "failed"
	StateDisabled   = "disabled"
	StateRemoved    = "removed"
)

// Instance is one running plugin (or MCP server) and its protocol client.
type Instance struct {
	ID     core.ULID
	Name   string
	client *mcp.Client
	proc   *exec.Cmd // nil for injected transports

	mu    sync.Mutex
	state string
	info  mcp.ServerInfo
}

// State returns the current lifecycle state.
func (in *Instance) State() string {
	in.mu.Lock()
	defer in.mu.Unlock()
	return in.state
}

// Runtime manages plugin instances.
type Runtime struct {
	mu        sync.Mutex
	instances map[core.ULID]*Instance
}

// NewRuntime returns a Plugin Runtime.
func NewRuntime() *Runtime { return &Runtime{instances: map[core.ULID]*Instance{}} }

// Connect attaches to a plugin over an already-established transport (used for tests and for
// out-of-process transports the caller owns). It initializes the protocol handshake.
func (r *Runtime) Connect(ctx context.Context, name string, rw io.ReadWriteCloser) (*Instance, error) {
	in := &Instance{ID: core.NewULID(), Name: name, client: mcp.NewClient(rw), state: StateStarting}
	info, err := in.client.Initialize(ctx)
	if err != nil {
		in.state = StateFailed
		return in, pluginErr("E-PLUG-010", "handshake failed: "+err.Error())
	}
	in.info = info
	in.state = StateRunning
	r.register(in)
	return in, nil
}

// Spawn launches a plugin subprocess and connects to its stdio. In production the launch path
// is the Sandbox Engine; this direct spawn is used where the caller has already applied policy.
func (r *Runtime) Spawn(ctx context.Context, name, program string, args ...string) (*Instance, error) {
	cmd := exec.CommandContext(ctx, program, args...) //nolint:gosec // program is policy-checked upstream
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, pluginErr("E-PLUG-011", "stdin pipe: "+err.Error())
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, pluginErr("E-PLUG-011", "stdout pipe: "+err.Error())
	}
	if err := cmd.Start(); err != nil {
		return nil, pluginErr("E-PLUG-011", "could not start plugin: "+err.Error())
	}
	in, err := r.Connect(ctx, name, &stdioRWC{in: stdin, out: stdout, cmd: cmd})
	if in != nil {
		in.proc = cmd
	}
	return in, err
}

// Tools bridges the plugin's tools to permission-mediated ToolPorts.
func (in *Instance) Tools(ctx context.Context) ([]ports.ToolPort, error) {
	return mcp.BridgeTools(ctx, in.client, in.Name)
}

// Stop terminates the plugin: closes the protocol connection and, for spawned instances, the
// process.
func (in *Instance) Stop() error {
	in.mu.Lock()
	in.state = StateStopping
	in.mu.Unlock()
	_ = in.client.Close()
	if in.proc != nil && in.proc.Process != nil {
		_ = in.proc.Process.Kill()
		_ = in.proc.Wait()
	}
	in.mu.Lock()
	in.state = StateStopped
	in.mu.Unlock()
	return nil
}

// List returns the running instances.
func (r *Runtime) List() []*Instance {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*Instance, 0, len(r.instances))
	for _, in := range r.instances {
		out = append(out, in)
	}
	return out
}

func (r *Runtime) register(in *Instance) {
	r.mu.Lock()
	r.instances[in.ID] = in
	r.mu.Unlock()
}

// stdioRWC adapts a subprocess's stdin/stdout to an io.ReadWriteCloser.
type stdioRWC struct {
	in  io.WriteCloser
	out io.ReadCloser
	cmd *exec.Cmd
}

func (s *stdioRWC) Read(p []byte) (int, error)  { return s.out.Read(p) }
func (s *stdioRWC) Write(p []byte) (int, error) { return s.in.Write(p) }
func (s *stdioRWC) Close() error {
	_ = s.in.Close()
	_ = s.out.Close()
	return nil
}

func pluginErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "plugin", Severity: "error", Message: msg}
}
