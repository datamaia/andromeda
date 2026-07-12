package builtin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/terminal"
)

func TestProcessControlListInspectTerminate(t *testing.T) {
	eng := terminal.New()
	ctx := context.Background()
	// Start a long-running supervised process.
	id, err := eng.Execute(ctx, ports.CommandSpec{Program: "sh", Args: []string{"-c", "sleep 30"}})
	if err != nil {
		t.Fatal(err)
	}
	tool := NewProcessControl(eng)

	// list shows it running.
	_, text := runTool(t, tool, `{"operation":"list"}`)
	var listRes struct {
		Processes []struct {
			ExecutionID string `json:"execution_id"`
			Running     bool   `json:"running"`
			State       string `json:"state"`
		} `json:"processes"`
	}
	_ = json.Unmarshal([]byte(text), &listRes)
	found := false
	for _, p := range listRes.Processes {
		if p.ExecutionID == string(id) {
			found = true
			if !p.Running || p.State != "running" {
				t.Fatalf("process should be running: %+v", p)
			}
		}
	}
	if !found {
		t.Fatalf("started process not listed: %s", text)
	}

	// inspect returns that single process.
	_, itext := runTool(t, tool, `{"operation":"inspect","execution_id":"`+string(id)+`"}`)
	if !json.Valid([]byte(itext)) {
		t.Fatalf("inspect returned invalid json: %s", itext)
	}

	// terminate kills it.
	if outcome, _ := runTool(t, tool, `{"operation":"terminate","execution_id":"`+string(id)+`"}`); outcome != "success" {
		t.Fatal("terminate should succeed")
	}
	outcome, _ := eng.Wait(ctx, id)
	if outcome.Status == "succeeded" {
		t.Fatalf("terminated process should not report success: %+v", outcome)
	}

	// After termination, list reports it not running (poll briefly for the finish goroutine).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_, lt := runTool(t, tool, `{"operation":"list"}`)
		var r struct {
			Processes []struct {
				ExecutionID string `json:"execution_id"`
				Running     bool   `json:"running"`
			} `json:"processes"`
		}
		_ = json.Unmarshal([]byte(lt), &r)
		stillRunning := false
		for _, p := range r.Processes {
			if p.ExecutionID == string(id) && p.Running {
				stillRunning = true
			}
		}
		if !stillRunning {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("process still reported running after terminate")
}

func TestProcessControlValidation(t *testing.T) {
	tool := NewProcessControl(terminal.New())
	if vr, _ := tool.Validate(context.Background(), []byte(`{"operation":"signal"}`)); vr.Valid {
		t.Fatal("signal without execution_id should be invalid")
	}
	if vr, _ := tool.Validate(context.Background(), []byte(`{"operation":"list"}`)); !vr.Valid {
		t.Fatal("list should be valid without execution_id")
	}
}

func TestProcessControlInspectUnknown(t *testing.T) {
	tool := NewProcessControl(terminal.New())
	if outcome, _ := runTool(t, tool, `{"operation":"inspect","execution_id":"nope"}`); outcome != "error" {
		t.Fatal("inspecting an unknown execution should error")
	}
}
