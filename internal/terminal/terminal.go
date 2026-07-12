package terminal

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// Engine implements ports.TerminalPort with pipe-based streaming execution.
type Engine struct {
	mu    sync.Mutex
	execs map[ports.ExecutionID]*execution
	// maxOutput bounds captured bytes per stream (0 = 1 MiB default).
	maxOutput int64
}

type execution struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	ptyFile *os.File // set for PTY-mode executions
	stream  ports.Stream[ports.TerminalChunk]
	sendCh  func(ports.TerminalChunk) bool
	closeCh func()

	mu       sync.Mutex
	waited   bool
	outcome  ports.CommandOutcome
	start    time.Time
	waitOnce sync.Once
	done     chan struct{}
}

// New returns a Terminal Engine.
func New() *Engine {
	return &Engine{execs: map[ports.ExecutionID]*execution{}, maxOutput: 1 << 20}
}

var _ ports.TerminalPort = (*Engine)(nil)

// Execute starts a command and returns its execution ID. When spec.PTY is set and the platform
// supports it, the command runs under a pseudoterminal (merged output, resizable); otherwise it
// runs in pipe mode with tagged stdout/stderr.
func (e *Engine) Execute(ctx context.Context, spec ports.CommandSpec) (ports.ExecutionID, error) {
	if spec.PTY && ptySupported() {
		return e.executePTY(ctx, spec)
	}
	cmd := exec.CommandContext(ctx, spec.Program, spec.Args...) //nolint:gosec // permission-checked upstream
	cmd.Dir = spec.Dir
	if len(spec.Env) > 0 {
		env := make([]string, 0, len(spec.Env))
		for k, v := range spec.Env {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}
	// Run the child as its own process-group leader so Signal can reach the whole tree (a shell
	// plus anything it forks); otherwise an orphaned grandchild keeps the pipes open and Wait hangs.
	setpgid(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", termErr("E-TOOL-010", "stdout pipe", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", termErr("E-TOOL-010", "stderr pipe", err)
	}
	stdin, _ := cmd.StdinPipe()

	if err := cmd.Start(); err != nil {
		return "", termErr("E-TOOL-011", "could not start command", err)
	}

	st, send, closeFn := streams.Chan[ports.TerminalChunk](256)
	ex := &execution{cmd: cmd, stdin: stdin, stream: st, sendCh: send, closeCh: closeFn, start: time.Now(), done: make(chan struct{})}

	var pumps sync.WaitGroup
	pumps.Add(2)
	go ex.pump(stdout, "stdout", e.maxOutput, &pumps)
	go ex.pump(stderr, "stderr", e.maxOutput, &pumps)
	go func() {
		pumps.Wait()
		ex.finish()
	}()

	id := core.NewULID()
	e.mu.Lock()
	e.execs[id] = ex
	e.mu.Unlock()
	return id, nil
}

// executePTY runs a command under a pseudoterminal.
func (e *Engine) executePTY(ctx context.Context, spec ports.CommandSpec) (ports.ExecutionID, error) {
	cmd := exec.CommandContext(ctx, spec.Program, spec.Args...) //nolint:gosec // permission-checked upstream
	cmd.Dir = spec.Dir
	if len(spec.Env) > 0 {
		env := make([]string, 0, len(spec.Env))
		for k, v := range spec.Env {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}
	f, err := ptyStart(cmd)
	if err != nil {
		return "", termErr("E-TOOL-011", "could not allocate pty", err)
	}
	st, send, closeFn := streams.Chan[ports.TerminalChunk](256)
	ex := &execution{cmd: cmd, ptyFile: f, stream: st, sendCh: send, closeCh: closeFn, start: time.Now(), done: make(chan struct{})}

	var pumps sync.WaitGroup
	pumps.Add(1)
	go ex.pump(f, "pty", e.maxOutput, &pumps)
	go func() {
		pumps.Wait()
		ex.finish()
	}()

	id := core.NewULID()
	e.mu.Lock()
	e.execs[id] = ex
	e.mu.Unlock()
	return id, nil
}

func (ex *execution) pump(r io.Reader, streamName string, limit int64, wg *sync.WaitGroup) {
	defer wg.Done()
	br := bufio.NewReader(r)
	var written int64
	buf := make([]byte, 4096)
	for {
		n, err := br.Read(buf)
		if n > 0 {
			chunk := ports.TerminalChunk{Stream: streamName, Data: append([]byte(nil), buf[:n]...)}
			if written+int64(n) > limit {
				chunk.Truncated = true
				chunk.Data = chunk.Data[:maxInt64(0, limit-written)]
			}
			written += int64(n)
			if written <= limit || chunk.Truncated {
				ex.sendCh(chunk)
			}
		}
		if err != nil {
			return
		}
	}
}

func (ex *execution) finish() {
	ex.waitOnce.Do(func() {
		err := ex.cmd.Wait()
		out := ports.CommandOutcome{DurationMS: time.Since(ex.start).Milliseconds()}
		if err == nil {
			out.Status, out.ExitCode = "succeeded", 0
		} else if ee, ok := err.(*exec.ExitError); ok {
			out.Status, out.ExitCode = "failed", ee.ExitCode()
		} else {
			out.Status, out.ExitCode = "failed", -1
		}
		ex.mu.Lock()
		ex.outcome = out
		ex.waited = true
		// Close the PTY under the same lock that Resize/Write take, then clear it: creack/pty's
		// Setsize reaches through File.Fd() (outside the poll refcount), so a Resize racing this
		// Close is a genuine data race. Serializing here and nil-ing the handle closes that race.
		if ex.ptyFile != nil {
			_ = ex.ptyFile.Close()
			ex.ptyFile = nil
		}
		ex.mu.Unlock()
		ex.closeCh()
		close(ex.done)
	})
}

// Stream returns the execution's output stream (single consumer).
func (e *Engine) Stream(_ context.Context, id ports.ExecutionID) (ports.Stream[ports.TerminalChunk], error) {
	ex, err := e.lookup(id)
	if err != nil {
		return nil, err
	}
	return ex.stream, nil
}

// Write sends input to the command's stdin.
func (e *Engine) Write(_ context.Context, id ports.ExecutionID, input []byte) error {
	ex, err := e.lookup(id)
	if err != nil {
		return err
	}
	ex.mu.Lock()
	pty := ex.ptyFile
	ex.mu.Unlock()
	if pty != nil {
		// os.File.Write is refcount-safe against a concurrent Close, so it need not hold the lock.
		_, werr := pty.Write(input)
		return werr
	}
	if ex.stdin == nil {
		return termErr("E-TOOL-012", "stdin not available", nil)
	}
	_, werr := ex.stdin.Write(input)
	return werr
}

// Signal delivers a portable signal to the command.
func (e *Engine) Signal(_ context.Context, id ports.ExecutionID, sig ports.SignalName) error {
	ex, err := e.lookup(id)
	if err != nil {
		return err
	}
	return sendSignal(ex.cmd, sig)
}

// Resize sets the PTY window size (a no-op in pipe mode).
func (e *Engine) Resize(_ context.Context, id ports.ExecutionID, cols, rows int) error {
	ex, err := e.lookup(id)
	if err != nil {
		return err
	}
	// Hold the lock across Setsize: it reaches through File.Fd(), which must not overlap the
	// Close in finish() (they would race on the file descriptor teardown).
	ex.mu.Lock()
	defer ex.mu.Unlock()
	if ex.ptyFile != nil {
		return ptySetsize(ex.ptyFile, cols, rows)
	}
	return nil
}

// Wait blocks until the command terminates and returns its outcome.
func (e *Engine) Wait(ctx context.Context, id ports.ExecutionID) (ports.CommandOutcome, error) {
	ex, err := e.lookup(id)
	if err != nil {
		return ports.CommandOutcome{}, err
	}
	select {
	case <-ctx.Done():
		return ports.CommandOutcome{}, ctx.Err()
	case <-ex.done:
		ex.mu.Lock()
		defer ex.mu.Unlock()
		return ex.outcome, nil
	}
}

// ExecutionSnapshot is a point-in-time view of a supervised execution (process.control tool).
type ExecutionSnapshot struct {
	ID       ports.ExecutionID
	Running  bool
	Status   string
	ExitCode int
	PID      int
}

// Snapshot lists the engine's supervised executions and their current state. Scope is strictly
// Andromeda-supervised process trees; arbitrary host processes are never reported.
func (e *Engine) Snapshot() []ExecutionSnapshot {
	e.mu.Lock()
	type pair struct {
		id ports.ExecutionID
		ex *execution
	}
	pairs := make([]pair, 0, len(e.execs))
	for id, ex := range e.execs {
		pairs = append(pairs, pair{id, ex})
	}
	e.mu.Unlock()

	out := make([]ExecutionSnapshot, 0, len(pairs))
	for _, p := range pairs {
		s := ExecutionSnapshot{ID: p.id}
		select {
		case <-p.ex.done:
			p.ex.mu.Lock()
			s.Status, s.ExitCode = p.ex.outcome.Status, p.ex.outcome.ExitCode
			p.ex.mu.Unlock()
		default:
			s.Running, s.Status = true, "running"
		}
		if p.ex.cmd != nil && p.ex.cmd.Process != nil {
			s.PID = p.ex.cmd.Process.Pid
		}
		out = append(out, s)
	}
	return out
}

// SnapshotOne returns a single supervised execution's snapshot.
func (e *Engine) SnapshotOne(id ports.ExecutionID) (ExecutionSnapshot, error) {
	if _, err := e.lookup(id); err != nil {
		return ExecutionSnapshot{}, err
	}
	for _, s := range e.Snapshot() {
		if s.ID == id {
			return s, nil
		}
	}
	return ExecutionSnapshot{}, termErr("E-TOOL-013", "unknown execution", nil)
}

func (e *Engine) lookup(id ports.ExecutionID) (*execution, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	ex, ok := e.execs[id]
	if !ok {
		return nil, termErr("E-TOOL-013", "unknown execution", nil)
	}
	return ex, nil
}

func termErr(code, msg string, cause error) error {
	pe := &ports.PortError{Code: code, Category: "tool", Severity: "error", Message: msg, Cause: cause}
	if cause != nil {
		pe.Detail = cause.Error()
	}
	return pe
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
