package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// The container-runtime tools operate their runtime through its official CLI (the documented
// interface): the Docker CLI over the Engine socket and kubectl over the API server. Each maps a
// structured operation to a fixed argv — never a raw shell string — so the surface is auditable
// and the concrete runtime facts stay in the runtime, not invented here.

// DockerControl operates the local Docker Engine via the docker CLI. Phase: Beta.
type DockerControl struct{ bin string }

// NewDockerControl builds docker.control. An empty bin uses "docker" on PATH.
func NewDockerControl(bin string) DockerControl {
	if bin == "" {
		bin = "docker"
	}
	return DockerControl{bin: bin}
}

// Describe returns the docker_control tool descriptor.
func (DockerControl) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "docker_control", Namespace: "docker", Version: "1",
		Description: "Operate the local Docker Engine via its CLI",
		InputSchema: []byte(`{"type":"object","required":["operation"],"properties":{` +
			`"operation":{"type":"string","enum":["ps","images","run","stop","rm","logs","build","inspect"]},` +
			`"args":{"type":"array","items":{"type":"string"}}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"result":{"type":"string"},"exit_code":{"type":"integer"},"truncated":{"type":"boolean"}}}`),
		Permissions:  []core.Permission{core.PermContainerAccess, core.PermNetwork}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

var dockerOps = map[string]bool{
	"ps": true, "images": true, "run": true, "stop": true, "rm": true, "logs": true, "build": true, "inspect": true,
}

// Validate requires an operation drawn from the supported Docker operation set.
func (DockerControl) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in cliOpInput
	if err := json.Unmarshal(input, &in); err != nil || in.Operation == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"operation is required"}}, nil
	}
	if !dockerOps[in.Operation] {
		return ports.ValidationResult{Valid: false, Findings: []string{"unsupported docker operation: " + in.Operation}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests host-scoped container access to the Docker Engine.
func (DockerControl) Resources(ports.JSON) ([]ports.PermissionQuery, error) {
	return []ports.PermissionQuery{{Permission: core.PermContainerAccess, Scope: core.ScopeHost, Subject: "docker"}}, nil
}

// Execute runs the docker CLI with the operation and its args as a fixed argv.
func (t DockerControl) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in cliOpInput
	_ = json.Unmarshal(req.Input, &in)
	argv := append([]string{in.Operation}, in.Args...)
	return runCLI(ctx, t.bin, argv), nil
}

// Cancel is a no-op; the CLI invocation is bounded by the Execute context.
func (DockerControl) Cancel(context.Context, core.ULID) error { return nil }

// KubernetesControl operates Kubernetes clusters via kubectl. Phase: v1.
type KubernetesControl struct{ bin string }

// NewKubernetesControl builds kubernetes.control. An empty bin uses "kubectl" on PATH.
func NewKubernetesControl(bin string) KubernetesControl {
	if bin == "" {
		bin = "kubectl"
	}
	return KubernetesControl{bin: bin}
}

var kubeOps = map[string]bool{
	"get": true, "list": true, "describe": true, "apply": true, "delete": true, "logs": true, "exec": true,
}

// Describe returns the kubernetes_control tool descriptor.
func (KubernetesControl) Describe(context.Context) (ports.ToolDescriptor, error) {
	// Pod exec is additionally gated as execute (catalog).
	return ports.ToolDescriptor{
		Name: "kubernetes_control", Namespace: "kubernetes", Version: "1",
		Description: "Operate Kubernetes clusters via kubectl",
		InputSchema: []byte(`{"type":"object","required":["operation"],"properties":{` +
			`"operation":{"type":"string","enum":["get","list","describe","apply","delete","logs","exec"]},` +
			`"context":{"type":"string"},"namespace":{"type":"string"},"args":{"type":"array","items":{"type":"string"}}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"result":{"type":"string"},"exit_code":{"type":"integer"},"truncated":{"type":"boolean"}}}`),
		Permissions:  []core.Permission{core.PermContainerAccess, core.PermNetwork, core.PermExecute}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

// Validate requires an operation drawn from the supported kubectl operation set.
func (KubernetesControl) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in kubeInput
	if err := json.Unmarshal(input, &in); err != nil || in.Operation == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"operation is required"}}, nil
	}
	if !kubeOps[in.Operation] {
		return ports.ValidationResult{Valid: false, Findings: []string{"unsupported kubectl operation: " + in.Operation}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests host-scoped container access, plus execute for the exec operation.
func (KubernetesControl) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in kubeInput
	_ = json.Unmarshal(input, &in)
	qs := []ports.PermissionQuery{{Permission: core.PermContainerAccess, Scope: core.ScopeHost, Subject: "kubernetes"}}
	if in.Operation == "exec" {
		qs = append(qs, ports.PermissionQuery{Permission: core.PermExecute, Scope: core.ScopeHost, Subject: "kubernetes"})
	}
	return qs, nil
}

// Execute runs kubectl with the operation, optional context/namespace flags, and args as a fixed argv.
func (t KubernetesControl) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in kubeInput
	_ = json.Unmarshal(req.Input, &in)
	argv := []string{in.Operation}
	if in.Context != "" {
		argv = append(argv, "--context", in.Context)
	}
	if in.Namespace != "" {
		argv = append(argv, "--namespace", in.Namespace)
	}
	argv = append(argv, in.Args...)
	return runCLI(ctx, t.bin, argv), nil
}

// Cancel is a no-op; the CLI invocation is bounded by the Execute context.
func (KubernetesControl) Cancel(context.Context, core.ULID) error { return nil }

type cliOpInput struct {
	Operation string   `json:"operation"`
	Args      []string `json:"args"`
}

type kubeInput struct {
	Operation string   `json:"operation"`
	Context   string   `json:"context"`
	Namespace string   `json:"namespace"`
	Args      []string `json:"args"`
}

// runCLI executes a runtime CLI with a fixed argv and captures its output under the output cap.
func runCLI(ctx context.Context, bin string, argv []string) ports.Stream[ports.ToolEvent] {
	cmd := exec.CommandContext(ctx, bin, argv...) //nolint:gosec // argv is built from a fixed op set, never a shell string
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	out := buf.String()
	truncated := false
	if len(out) > maxHTTPBody {
		out = out[:maxHTTPBody]
		truncated = true
	}
	exit := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			// Binary missing or not runnable: surface as a tool-local error, not a crash.
			return errEvent(strings.TrimSpace(bin + ": " + err.Error()))
		}
	}
	res, _ := json.Marshal(map[string]any{"result": out, "exit_code": exit, "truncated": truncated})
	if exit != 0 {
		return errEvent(string(res))
	}
	return okEvent(string(res))
}
