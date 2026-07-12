package terminal

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
)

func collect(t *testing.T, e *Engine, id ports.ExecutionID) string {
	t.Helper()
	st, err := e.Stream(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	for {
		c, err := st.Next(context.Background())
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		sb.Write(c.Data)
	}
	return sb.String()
}

func TestExecuteCapturesStdout(t *testing.T) {
	ctx := context.Background()
	e := New()
	id, err := e.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "printf 'hello world'"}})
	if err != nil {
		t.Fatal(err)
	}
	out := collect(t, e, id)
	outcome, err := e.Wait(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hello world") {
		t.Errorf("stdout = %q", out)
	}
	if outcome.Status != "succeeded" || outcome.ExitCode != 0 {
		t.Errorf("outcome = %+v", outcome)
	}
}

func TestNonZeroExit(t *testing.T) {
	ctx := context.Background()
	e := New()
	id, _ := e.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "exit 3"}})
	_ = collect(t, e, id)
	outcome, _ := e.Wait(ctx, id)
	if outcome.Status != "failed" || outcome.ExitCode != 3 {
		t.Errorf("outcome = %+v", outcome)
	}
}

func TestStdinWrite(t *testing.T) {
	ctx := context.Background()
	e := New()
	id, _ := e.Execute(ctx, ports.CommandSpec{Program: "cat"})
	if err := e.Write(ctx, id, []byte("piped input")); err != nil {
		t.Fatal(err)
	}
	// Close stdin by signalling cat to finish reading (EOF via terminate after a beat).
	ex, _ := e.lookup(id)
	_ = ex.stdin.Close()
	out := collect(t, e, id)
	if !strings.Contains(out, "piped input") {
		t.Errorf("cat output = %q", out)
	}
	_, _ = e.Wait(ctx, id)
}

func TestSignalStopsLongRunning(t *testing.T) {
	ctx := context.Background()
	e := New()
	id, _ := e.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "sleep 300"}})
	if err := e.Signal(ctx, id, ports.SignalKill); err != nil {
		t.Fatal(err)
	}
	done := make(chan ports.CommandOutcome, 1)
	go func() { o, _ := e.Wait(ctx, id); done <- o }()
	select {
	case o := <-done:
		if o.Status == "succeeded" {
			t.Error("killed process should not report success")
		}
	case <-time.After(30 * time.Second):
		t.Fatal("Wait did not return after kill")
	}
}

// TestSignalKillsProcessTree guards the orphaned-grandchild hang: the shell forks `sleep` as a
// child and waits, so killing only the shell would leave `sleep` holding the output pipes open,
// the pumps would never see EOF, and Wait would block. Signal must reach the whole process group.
func TestSignalKillsProcessTree(t *testing.T) {
	ctx := context.Background()
	e := New()
	id, _ := e.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "sleep 300 & wait"}})
	if err := e.Signal(ctx, id, ports.SignalKill); err != nil {
		t.Fatal(err)
	}
	done := make(chan ports.CommandOutcome, 1)
	go func() { o, _ := e.Wait(ctx, id); done <- o }()
	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("Wait hung after kill: the signal did not reach the whole process group")
	}
}

func TestUnknownExecution(t *testing.T) {
	if _, err := New().Wait(context.Background(), "nope"); err == nil {
		t.Error("expected error for unknown execution")
	}
}

func TestPTYModeCapturesOutput(t *testing.T) {
	if !ptySupported() {
		t.Skip("pty not supported")
	}
	ctx := context.Background()
	e := New()
	id, err := e.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "printf pty-hello"}, PTY: true})
	if err != nil {
		t.Fatal(err)
	}
	// Resize should succeed on a PTY execution.
	if err := e.Resize(ctx, id, 100, 40); err != nil {
		t.Errorf("resize: %v", err)
	}
	out := collect(t, e, id)
	if !strings.Contains(out, "pty-hello") {
		t.Errorf("pty output = %q", out)
	}
	outcome, _ := e.Wait(ctx, id)
	if outcome.Status != "succeeded" {
		t.Errorf("pty outcome = %+v", outcome)
	}
}

// TestPTYResizeRacesFinish drives Resize concurrently with the process exiting (and its PTY being
// closed in finish). Under -race this reproduces the File.Fd()-vs-Close data race if the PTY
// lifecycle is not serialized. It must remain race-clean.
func TestPTYResizeRacesFinish(t *testing.T) {
	if !ptySupported() {
		t.Skip("pty not supported")
	}
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		e := New()
		id, err := e.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "printf x"}, PTY: true})
		if err != nil {
			t.Fatal(err)
		}
		done := make(chan struct{})
		go func() {
			for j := 0; j < 50; j++ {
				_ = e.Resize(ctx, id, 80+j, 24)
			}
			close(done)
		}()
		_, _ = e.Wait(ctx, id) // triggers finish()/PTY close, racing the Resize loop
		<-done
	}
}
