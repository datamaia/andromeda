package builtin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// stubBin writes an executable shell script that echoes its arguments, standing in for docker or
// kubectl so the tools' argv construction and output capture are testable without the real runtime.
func stubBin(t *testing.T, name, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("stub shell script not portable to windows in this test")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestDockerControlBuildsArgv(t *testing.T) {
	bin := stubBin(t, "docker", `echo "ARGS:$@"`)
	tool := NewDockerControl(bin)
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"operation": "ps", "args": []string{"-a"}}))
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	var res struct {
		Result   string `json:"result"`
		ExitCode int    `json:"exit_code"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Result != "ARGS:ps -a\n" || res.ExitCode != 0 {
		t.Fatalf("result = %+v", res)
	}
}

func TestDockerControlRejectsUnknownOperation(t *testing.T) {
	tool := NewDockerControl("docker")
	if vr, _ := tool.Validate(t.Context(), []byte(`{"operation":"nuke"}`)); vr.Valid {
		t.Fatal("unknown docker operation should be invalid")
	}
}

func TestDockerControlNonZeroExitIsError(t *testing.T) {
	bin := stubBin(t, "docker", `echo boom; exit 2`)
	tool := NewDockerControl(bin)
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"operation": "stop"}))
	if outcome != "error" {
		t.Fatal("non-zero docker exit should be an error outcome")
	}
	var res struct {
		ExitCode int `json:"exit_code"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.ExitCode != 2 {
		t.Fatalf("exit_code = %d", res.ExitCode)
	}
}

func TestKubernetesControlBuildsContextAndNamespace(t *testing.T) {
	bin := stubBin(t, "kubectl", `echo "ARGS:$@"`)
	tool := NewKubernetesControl(bin)
	outcome, text := runTool(t, tool, mustJSON(map[string]any{
		"operation": "get", "context": "prod", "namespace": "web", "args": []string{"pods"},
	}))
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	var res struct {
		Result string `json:"result"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Result != "ARGS:get --context prod --namespace web pods\n" {
		t.Fatalf("argv = %q", res.Result)
	}
}

func TestKubernetesExecRequestsExecutePermission(t *testing.T) {
	qs, _ := NewKubernetesControl("").Resources([]byte(`{"operation":"exec"}`))
	var exec bool
	for _, q := range qs {
		if q.Permission == "execute" {
			exec = true
		}
	}
	if !exec {
		t.Fatal("kubectl exec must additionally request execute")
	}
	// A non-exec op must not.
	qs, _ = NewKubernetesControl("").Resources([]byte(`{"operation":"get"}`))
	for _, q := range qs {
		if q.Permission == "execute" {
			t.Fatal("get must not request execute")
		}
	}
}

func TestContainerMissingBinaryIsToolError(t *testing.T) {
	tool := NewDockerControl(filepath.Join(t.TempDir(), "does-not-exist"))
	if outcome, _ := runTool(t, tool, `{"operation":"ps"}`); outcome != "error" {
		t.Fatal("a missing runtime binary should surface as a tool error, not a crash")
	}
}
