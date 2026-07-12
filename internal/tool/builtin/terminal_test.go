package builtin

import (
	"encoding/json"
	"testing"

	"github.com/datamaia/andromeda/internal/terminal"
)

func TestTerminalRunCapturesOutput(t *testing.T) {
	tool := NewTerminalRun(terminal.New())
	outcome, text := runTool(t, tool, `{"command":"echo hello"}`)
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	var res struct {
		Output   string `json:"output"`
		ExitCode int    `json:"exit_code"`
		Status   string `json:"status"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Output != "hello\n" || res.ExitCode != 0 || res.Status != "succeeded" {
		t.Fatalf("result = %+v", res)
	}
}

func TestTerminalRunNonZeroExitFails(t *testing.T) {
	tool := NewTerminalRun(terminal.New())
	outcome, text := runTool(t, tool, `{"command":"exit 3"}`)
	if outcome != "error" {
		t.Fatalf("expected error outcome, got %s", outcome)
	}
	var res struct {
		ExitCode int    `json:"exit_code"`
		Status   string `json:"status"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.ExitCode != 3 || res.Status != "failed" {
		t.Fatalf("result = %+v", res)
	}
}

func TestTerminalRunValidation(t *testing.T) {
	tool := NewTerminalRun(terminal.New())
	vr, _ := tool.Validate(t.Context(), []byte(`{}`))
	if vr.Valid {
		t.Fatal("empty command should be invalid")
	}
}
