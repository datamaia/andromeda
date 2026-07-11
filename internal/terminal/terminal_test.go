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
	id, _ := e.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "sleep 30"}})
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
	case <-time.After(5 * time.Second):
		t.Fatal("Wait did not return after kill")
	}
}

func TestUnknownExecution(t *testing.T) {
	if _, err := New().Wait(context.Background(), "nope"); err == nil {
		t.Error("expected error for unknown execution")
	}
}
