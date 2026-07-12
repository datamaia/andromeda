package tool_test

import (
	"context"
	"testing"

	"github.com/datamaia/andromeda/internal/permission"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/datamaia/andromeda/internal/tool"
)

// BenchmarkDispatch measures tool dispatch overhead: name lookup, input schema validation,
// permission evaluation, execute, and draining the outcome — over the permissiveTool double so
// the tool's own work is negligible. Relates to NFR-PERF-008 (tool dispatch overhead).
// Micro-benchmark tier, Volume 12 chapter 03.
func BenchmarkDispatch(b *testing.B) {
	ctx := context.Background()
	db, err := storage.OpenWorkspaceDB(ctx, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	rt := tool.NewRuntime(permission.NewManager(permission.NewStore(db)))
	if err := rt.Register(ctx, permissiveTool{}); err != nil {
		b.Fatal(err)
	}
	names := rt.Names()
	if len(names) == 0 {
		b.Fatal("no tools registered")
	}
	name := names[0]
	input := ports.JSON(`{"n":1}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st, err := rt.Invoke(ctx, name, ports.PermissionQuery{}, input)
		if err != nil {
			b.Fatal(err)
		}
		for {
			if _, err := st.Next(ctx); err != nil {
				break
			}
		}
	}
}
