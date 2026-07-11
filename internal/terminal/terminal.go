package terminal

import (
	"bufio"
	"context"
	"io"
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

// Execute starts a command in pipe mode and returns its execution ID.
func (e *Engine) Execute(ctx context.Context, spec ports.CommandSpec) (ports.ExecutionID, error) {
	cmd := exec.CommandContext(ctx, spec.Program, spec.Args...) //nolint:gosec // permission-checked upstream
	cmd.Dir = spec.Dir
	if len(spec.Env) > 0 {
		env := make([]string, 0, len(spec.Env))
		for k, v := range spec.Env {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}
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
		ex.mu.Unlock()
		ex.closeCh()
		close(ex.done)
	})
}

// Stream returns the execution's output stream (single consumer).
func (e *Engine) Stream(ctx context.Context, id ports.ExecutionID) (ports.Stream[ports.TerminalChunk], error) {
	ex, err := e.lookup(id)
	if err != nil {
		return nil, err
	}
	return ex.stream, nil
}

// Write sends input to the command's stdin.
func (e *Engine) Write(ctx context.Context, id ports.ExecutionID, input []byte) error {
	ex, err := e.lookup(id)
	if err != nil {
		return err
	}
	if ex.stdin == nil {
		return termErr("E-TOOL-012", "stdin not available", nil)
	}
	_, werr := ex.stdin.Write(input)
	return werr
}

// Signal delivers a portable signal to the command.
func (e *Engine) Signal(ctx context.Context, id ports.ExecutionID, sig ports.SignalName) error {
	ex, err := e.lookup(id)
	if err != nil {
		return err
	}
	return sendSignal(ex.cmd, sig)
}

// Resize is a no-op in pipe mode (no PTY).
func (e *Engine) Resize(ctx context.Context, id ports.ExecutionID, cols, rows int) error {
	_, err := e.lookup(id)
	return err
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
