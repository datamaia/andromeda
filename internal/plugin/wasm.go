package plugin

import (
	"context"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// WASM plugins (ADR-009 v2 mechanism) run untrusted extension logic inside a wazero sandbox —
// a pure-Go WebAssembly runtime with no host access unless explicitly granted. The guest ABI:
//   - export "alloc"(size i32) i32          — allocate `size` bytes, return the pointer
//   - export "run"(ptr i32, len i32) i64     — process the input at [ptr,len]; return the
//                                              result location packed as (ptr<<32 | len)
// The host reads the result from the guest's exported "memory". Guests that only compute may
// export a nullary "run_i32"(result i32) instead (used by the smoke test).

// WASMRuntime hosts WebAssembly plugin modules.
type WASMRuntime struct {
	rt wazero.Runtime
}

// NewWASMRuntime returns a WASM runtime bound to ctx (nothing from the host is exposed to
// guests: no filesystem, clock, or network — a wazero module with no imports).
func NewWASMRuntime(ctx context.Context) *WASMRuntime {
	return &WASMRuntime{rt: wazero.NewRuntime(ctx)}
}

// Close releases the runtime.
func (w *WASMRuntime) Close(ctx context.Context) error { return w.rt.Close(ctx) }

// WASMModule is an instantiated WASM plugin module.
type WASMModule struct {
	name string
	mod  api.Module
}

// Instantiate loads and instantiates a WASM module from its bytes.
func (w *WASMRuntime) Instantiate(ctx context.Context, name string, wasm []byte) (*WASMModule, error) {
	mod, err := w.rt.Instantiate(ctx, wasm)
	if err != nil {
		return nil, pluginErr("E-PLUG-020", "wasm instantiate failed: "+err.Error())
	}
	return &WASMModule{name: name, mod: mod}, nil
}

// CallI32 calls a nullary exported function returning an i32 (compute-only guests).
func (m *WASMModule) CallI32(ctx context.Context, fn string) (int32, error) {
	f := m.mod.ExportedFunction(fn)
	if f == nil {
		return 0, pluginErr("E-PLUG-021", "wasm export not found: "+fn)
	}
	res, err := f.Call(ctx)
	if err != nil {
		return 0, pluginErr("E-PLUG-022", "wasm call failed: "+err.Error())
	}
	if len(res) == 0 {
		return 0, pluginErr("E-PLUG-022", "wasm function returned no value")
	}
	return int32(res[0]), nil
}

// Run passes input bytes to the guest's alloc+run ABI and returns the result bytes.
func (m *WASMModule) Run(ctx context.Context, input []byte) ([]byte, error) {
	alloc := m.mod.ExportedFunction("alloc")
	run := m.mod.ExportedFunction("run")
	mem := m.mod.Memory()
	if alloc == nil || run == nil || mem == nil {
		return nil, pluginErr("E-PLUG-021", "wasm module does not implement the run ABI")
	}
	ptrRes, err := alloc.Call(ctx, uint64(len(input)))
	if err != nil {
		return nil, pluginErr("E-PLUG-022", "wasm alloc failed: "+err.Error())
	}
	ptr := uint32(ptrRes[0])
	if !mem.Write(ptr, input) {
		return nil, pluginErr("E-PLUG-022", "wasm memory write out of range")
	}
	out, err := run.Call(ctx, uint64(ptr), uint64(len(input)))
	if err != nil {
		return nil, pluginErr("E-PLUG-022", "wasm run failed: "+err.Error())
	}
	packed := out[0]
	outPtr := uint32(packed >> 32)
	outLen := uint32(packed & 0xffffffff)
	data, ok := mem.Read(outPtr, outLen)
	if !ok {
		return nil, pluginErr("E-PLUG-022", "wasm result read out of range")
	}
	return append([]byte(nil), data...), nil
}

// Close releases the module.
func (m *WASMModule) Close(ctx context.Context) error { return m.mod.Close(ctx) }

// AsTool bridges a WASM module's run ABI to a permission-mediated ToolPort (untrusted origin).
func (m *WASMModule) AsTool(description string, perms []core.Permission) ports.ToolPort {
	return &wasmTool{mod: m, description: description, perms: perms}
}

type wasmTool struct {
	mod         *WASMModule
	description string
	perms       []core.Permission
}

func (t *wasmTool) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "wasm_" + t.mod.name, Namespace: "wasm/" + t.mod.name, Version: "1",
		Description: t.description, InputSchema: []byte(`{"type":"object"}`),
		Permissions: t.perms, Origin: "plugin", TrustLevel: "untrusted",
	}, nil
}

func (t *wasmTool) Validate(context.Context, ports.JSON) (ports.ValidationResult, error) {
	return ports.ValidationResult{Valid: true}, nil
}

func (t *wasmTool) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	out, err := t.mod.Run(ctx, req.Input)
	if err != nil {
		return streams.Slice([]ports.ToolEvent{{Kind: "terminal", Terminal: true, Outcome: "error", Text: err.Error()}}), nil
	}
	return streams.Slice([]ports.ToolEvent{
		{Kind: "output", Text: string(out)},
		{Kind: "terminal", Terminal: true, Outcome: "success", Text: string(out)},
	}), nil
}

func (t *wasmTool) Cancel(context.Context, core.ULID) error { return nil }
