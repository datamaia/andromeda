package tool

import (
	"bytes"
	"context"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// ResourceScoped is an optional tool interface: a tool that can derive the concrete permission
// queries its input implies (e.g. the specific paths it will read/write) implements it so the
// Runtime evaluates path-level permissions rather than coarse tool-level ones.
type ResourceScoped interface {
	Resources(input ports.JSON) ([]ports.PermissionQuery, error)
}

// Runtime registers tools and mediates their invocation under validation and permissions.
type Runtime struct {
	perms       ports.PermissionPort
	tools       map[string]ports.ToolPort
	schemas     map[string]*jsonschema.Schema // compiled InputSchema per tool (ADR-024, FR-TOOL-002)
	interactive bool                          // route "ask" outcomes through the interactive approver
}

// RuntimeOption configures a Runtime.
type RuntimeOption func(*Runtime)

// WithInteractive makes the Runtime evaluate permissions through the interactive Request path, so
// an "ask" outcome raises an approval prompt (via the Manager's approver) instead of failing closed.
func WithInteractive() RuntimeOption { return func(r *Runtime) { r.interactive = true } }

// NewRuntime returns a Tool Runtime that decides permissions through perms.
func NewRuntime(perms ports.PermissionPort, opts ...RuntimeOption) *Runtime {
	r := &Runtime{perms: perms, tools: map[string]ports.ToolPort{}, schemas: map[string]*jsonschema.Schema{}}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Register adds a tool by its declared name and compiles its input JSON Schema. A tool whose
// declared InputSchema is not a valid JSON Schema is rejected at registration (FR-TOOL-002:
// an incomplete or malformed declaration is never partially accepted).
func (r *Runtime) Register(ctx context.Context, t ports.ToolPort) error {
	d, err := t.Describe(ctx)
	if err != nil {
		return err
	}
	if d.Name == "" {
		return fmt.Errorf("tool has no name")
	}
	if len(d.InputSchema) > 0 {
		sch, err := compileSchema(d.Name, d.InputSchema)
		if err != nil {
			return toolErr("E-TOOL-002", "invalid input schema for "+d.Name+": "+err.Error())
		}
		r.schemas[d.Name] = sch
	}
	r.tools[d.Name] = t
	return nil
}

// compileSchema compiles a tool's declared JSON Schema document.
func compileSchema(name string, schema ports.JSON) (*jsonschema.Schema, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return nil, err
	}
	c := jsonschema.NewCompiler()
	url := "mem://tool/" + name
	if err := c.AddResource(url, doc); err != nil {
		return nil, err
	}
	return c.Compile(url)
}

// Names returns the registered tool names.
func (r *Runtime) Names() []string {
	out := make([]string, 0, len(r.tools))
	for n := range r.tools {
		out = append(out, n)
	}
	return out
}

// Describe returns a registered tool's descriptor.
func (r *Runtime) Describe(ctx context.Context, name string) (ports.ToolDescriptor, error) {
	t, ok := r.tools[name]
	if !ok {
		return ports.ToolDescriptor{}, toolErr("E-TOOL-001", "unknown tool: "+name)
	}
	return t.Describe(ctx)
}

// Invoke validates input, evaluates permissions, and drives the tool, returning its ToolEvent
// stream. A validation failure returns a Go error (the invocation could not start). A
// permission denial is delivered as data: a single terminal error event.
func (r *Runtime) Invoke(ctx context.Context, name string, subjectCtx ports.PermissionQuery, input ports.JSON) (ports.Stream[ports.ToolEvent], error) {
	t, ok := r.tools[name]
	if !ok {
		return nil, toolErr("E-TOOL-001", "unknown tool: "+name)
	}
	desc, err := t.Describe(ctx)
	if err != nil {
		return nil, err
	}
	// Structural validation first: the declared JSON Schema is enforced by the Runtime, not left
	// to each tool (ADR-024, FR-TOOL-002). Then the tool's own semantic Validate runs.
	if sch := r.schemas[name]; sch != nil {
		inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(input))
		if err != nil {
			return nil, toolErr("E-TOOL-002", "input is not valid JSON for "+name)
		}
		if err := sch.Validate(inst); err != nil {
			return nil, toolErr("E-TOOL-002", "input schema validation failed for "+name+": "+schemaMessage(err))
		}
	}
	vr, err := t.Validate(ctx, input)
	if err != nil {
		return nil, err
	}
	if !vr.Valid {
		return nil, toolErr("E-TOOL-002", "input validation failed for "+name)
	}

	// Evaluate permissions. Prefer path-level queries when the tool provides them. In interactive
	// mode an "ask" outcome is routed to the approver (Request); otherwise it fails closed (Check).
	queries := r.permissionQueries(t, desc, subjectCtx, input)
	for _, q := range queries {
		d, err := r.decide(ctx, q)
		if err != nil || d.Outcome != core.OutcomeAllow {
			return streams.Slice([]ports.ToolEvent{{
				Kind: "terminal", Terminal: true, Outcome: "error",
				Text: fmt.Sprintf("permission denied: %s on %s", q.Permission, q.Subject),
			}}), nil
		}
	}

	req := ports.ToolExecuteRequest{
		InvocationID: core.NewULID(),
		Input:        input,
		Granted:      desc.Permissions,
	}
	return t.Execute(ctx, req)
}

// decide evaluates one permission query, prompting via the approver when the Runtime is
// interactive and falling back to the non-interactive check otherwise.
func (r *Runtime) decide(ctx context.Context, q ports.PermissionQuery) (ports.Decision, error) {
	if r.interactive {
		return r.perms.Request(ctx, ports.PermissionRequest{Query: q, Interactive: true})
	}
	return r.perms.Check(ctx, q)
}

func (r *Runtime) permissionQueries(t ports.ToolPort, desc ports.ToolDescriptor, subj ports.PermissionQuery, input ports.JSON) []ports.PermissionQuery {
	if rs, ok := t.(ResourceScoped); ok {
		if qs, err := rs.Resources(input); err == nil && len(qs) > 0 {
			// carry the subject's session/workspace context onto each derived query
			for i := range qs {
				qs[i].SessionID, qs[i].WorkspaceID = subj.SessionID, subj.WorkspaceID
			}
			return qs
		}
	}
	qs := make([]ports.PermissionQuery, 0, len(desc.Permissions))
	for _, p := range desc.Permissions {
		qs = append(qs, ports.PermissionQuery{
			Permission: p, Scope: core.ScopeTool, Subject: desc.Name,
			SessionID: subj.SessionID, WorkspaceID: subj.WorkspaceID,
		})
	}
	return qs
}

// schemaMessage renders a jsonschema validation error compactly (the full tree is verbose).
func schemaMessage(err error) string {
	if ve, ok := err.(*jsonschema.ValidationError); ok {
		msg := ve.Error()
		if i := bytes.IndexByte([]byte(msg), '\n'); i > 0 {
			return msg[:i]
		}
		return msg
	}
	return err.Error()
}

func toolErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "tool", Severity: "error", Message: msg}
}
