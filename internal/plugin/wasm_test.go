package plugin

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// answerModule is a minimal, valid WebAssembly module:
//
//	(module (func (export "answer") (result i32) i32.const 42))
//
// Embedded as bytes so the test needs no wasm toolchain; it proves the wazero runtime loads,
// instantiates, and calls a guest export end to end.
var answerModule = []byte{
	0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // magic + version
	0x01, 0x05, 0x01, 0x60, 0x00, 0x01, 0x7f, // type: () -> i32
	0x03, 0x02, 0x01, 0x00, // func: 1 func of type 0
	0x07, 0x0a, 0x01, 0x06, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x00, 0x00, // export "answer" func 0
	0x0a, 0x06, 0x01, 0x04, 0x00, 0x41, 0x2a, 0x0b, // code: i32.const 42; end
}

func TestWASMInstantiateAndCall(t *testing.T) {
	ctx := context.Background()
	rt := NewWASMRuntime(ctx)
	defer rt.Close(ctx)

	mod, err := rt.Instantiate(ctx, "answerer", answerModule)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Close(ctx)

	v, err := mod.CallI32(ctx, "answer")
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Fatalf("answer() = %d, want 42", v)
	}
}

func TestWASMMissingExport(t *testing.T) {
	ctx := context.Background()
	rt := NewWASMRuntime(ctx)
	defer rt.Close(ctx)
	mod, _ := rt.Instantiate(ctx, "m", answerModule)
	if _, err := mod.CallI32(ctx, "nope"); err == nil {
		t.Fatal("expected an error for a missing export")
	}
}

func TestWASMInvalidModuleRejected(t *testing.T) {
	ctx := context.Background()
	rt := NewWASMRuntime(ctx)
	defer rt.Close(ctx)
	if _, err := rt.Instantiate(ctx, "bad", []byte{0x00, 0x01, 0x02, 0x03}); err == nil {
		t.Fatal("expected instantiation of an invalid module to fail")
	}
}

func TestRuntimeLoadWASMAndClose(t *testing.T) {
	ctx := context.Background()
	rt := NewRuntime()
	mod, err := rt.LoadWASM(ctx, "answerer", answerModule)
	if err != nil {
		t.Fatal(err)
	}
	v, err := mod.CallI32(ctx, "answer")
	if err != nil || v != 42 {
		t.Fatalf("answer() = %d, err = %v", v, err)
	}
	// Close releases the runtime and its modules; a second Close is a no-op.
	if err := rt.Close(ctx); err != nil {
		t.Fatal(err)
	}
	if err := rt.Close(ctx); err != nil {
		t.Fatalf("second Close should be a no-op, got %v", err)
	}
}

func TestWASMRunABIAbsentReportsError(t *testing.T) {
	ctx := context.Background()
	rt := NewWASMRuntime(ctx)
	defer rt.Close(ctx)
	mod, _ := rt.Instantiate(ctx, "m", answerModule) // has no alloc/run/memory

	// The bridged tool surfaces the missing ABI as a terminal error event (denial-as-data style).
	tool := mod.AsTool("compute", nil)
	d, _ := tool.Describe(ctx)
	if d.Origin != "plugin" || d.TrustLevel != "untrusted" {
		t.Fatalf("descriptor = %+v", d)
	}
	st, _ := tool.Execute(ctx, ports.ToolExecuteRequest{Input: []byte(`{}`)})
	var term ports.ToolEvent
	for {
		ev, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if ev.Terminal {
			term = ev
		}
	}
	if term.Outcome != "error" {
		t.Fatalf("expected an error outcome for a module without the run ABI, got %+v", term)
	}
}
