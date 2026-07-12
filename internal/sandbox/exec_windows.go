//go:build windows

package sandbox

import (
	"context"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
)

// execution is a running sandboxed command on Windows. The command runs in a new process group
// so Teardown can terminate the whole tree via taskkill.
type execution struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	start  time.Time

	mu      sync.Mutex
	done    bool
	outcome ports.CommandOutcome
}

func startExecution(ctx context.Context, spec ports.CommandSpec, dir string, env []string, timeout time.Duration) (*execution, error) {
	runCtx := ctx
	cancel := context.CancelFunc(func() {})
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	c := exec.CommandContext(runCtx, spec.Program, spec.Args...) //nolint:gosec // policy-checked upstream
	c.Dir = dir
	c.Env = env
	if err := c.Start(); err != nil {
		cancel()
		return nil, err
	}
	return &execution{cmd: c, cancel: cancel, start: time.Now()}, nil
}

func (e *execution) killTree() {
	e.mu.Lock()
	done := e.done
	e.mu.Unlock()
	if done || e.cmd.Process == nil {
		return
	}
	// taskkill terminates the process and its children (/T) forcibly (/F).
	_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(e.cmd.Process.Pid)).Run()
	e.cancel()
}

func (e *execution) wait(_ context.Context) ports.CommandOutcome {
	e.mu.Lock()
	if e.done {
		o := e.outcome
		e.mu.Unlock()
		return o
	}
	e.mu.Unlock()

	err := e.cmd.Wait()
	e.cancel()
	out := ports.CommandOutcome{DurationMS: time.Since(e.start).Milliseconds()}
	if err == nil {
		out.Status, out.ExitCode = "succeeded", 0
	} else if ee, ok := err.(*exec.ExitError); ok {
		out.Status, out.ExitCode = "failed", ee.ExitCode()
	} else {
		out.Status, out.ExitCode = "failed", -1
	}
	e.mu.Lock()
	e.done = true
	e.outcome = out
	e.mu.Unlock()
	return out
}
