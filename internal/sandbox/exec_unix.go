//go:build unix

package sandbox

import (
	"context"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
)

// execution is a running sandboxed command launched in its own process group so that Teardown
// can terminate the whole tree.
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
	var cancel context.CancelFunc = func() {}
	if timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	c := exec.CommandContext(runCtx, spec.Program, spec.Args...) //nolint:gosec // program is policy-checked upstream
	c.Dir = dir
	c.Env = env
	// New process group so the whole tree can be signalled at once.
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
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
	// Negative PID targets the whole process group.
	_ = syscall.Kill(-e.cmd.Process.Pid, syscall.SIGKILL)
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
		out.Status = "succeeded"
		out.ExitCode = 0
	} else if ee, ok := err.(*exec.ExitError); ok {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			out.Status = "killed"
			out.Signal = ws.Signal().String()
		} else {
			out.Status = "failed"
			out.ExitCode = ee.ExitCode()
		}
	} else {
		out.Status = "failed"
		out.ExitCode = -1
	}
	e.mu.Lock()
	e.done = true
	e.outcome = out
	e.mu.Unlock()
	return out
}
