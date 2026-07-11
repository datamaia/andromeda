package app

import (
	"context"
	"fmt"

	"github.com/datamaia/andromeda/internal/config"
	"github.com/datamaia/andromeda/internal/eventbus"
	"github.com/datamaia/andromeda/internal/pal"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/datamaia/andromeda/internal/telemetry"
)

// Check is one diagnostic result.
type Check struct {
	Name   string
	OK     bool
	Detail string
}

// DoctorReport is the result of the environment diagnostic.
type DoctorReport struct {
	Checks []Check
}

// OK reports whether every check passed.
func (r DoctorReport) OK() bool {
	for _, c := range r.Checks {
		if !c.OK {
			return false
		}
	}
	return true
}

// Doctor exercises the MS-1 foundation end to end: it resolves configuration with attribution,
// opens the workspace and global databases (running migrations and backups), and emits an
// enveloped event to persisted storage. workspaceRoot is the directory to treat as the
// workspace (typically the current working directory).
func Doctor(ctx context.Context, workspaceRoot string) (DoctorReport, error) {
	var rep DoctorReport
	dirs := pal.NewConfigDirs()

	// 1. Configuration resolution with source attribution.
	cfg, err := config.Load(ctx, dirs, workspaceRoot)
	if err != nil {
		rep.Checks = append(rep.Checks, Check{"config", false, err.Error()})
		return rep, nil
	}
	resolved, err := cfg.Resolve(ctx, ports.ConfigQuery{Scope: "workspace"})
	if err != nil {
		rep.Checks = append(rep.Checks, Check{"config", false, err.Error()})
		return rep, nil
	}
	rep.Checks = append(rep.Checks, Check{
		"config", true,
		fmt.Sprintf("%d keys resolved (logging.level=%v from %s)",
			len(resolved.Values), resolved.Values["logging.level"], resolved.Sources["logging.level"]),
	})

	// 2. Global database.
	dataDir, err := dirs.DataHome()
	if err != nil {
		rep.Checks = append(rep.Checks, Check{"global-db", false, err.Error()})
		return rep, nil
	}
	gdb, err := storage.OpenGlobalDB(ctx, dataDir)
	if err != nil {
		rep.Checks = append(rep.Checks, Check{"global-db", false, err.Error()})
		return rep, nil
	}
	defer gdb.Close()
	gv, _ := gdb.SchemaVersion(ctx)
	rep.Checks = append(rep.Checks, Check{"global-db", true, fmt.Sprintf("%s (schema v%d)", gdb.Path(), gv)})

	// 3. Workspace database.
	wdb, err := storage.OpenWorkspaceDB(ctx, workspaceRoot)
	if err != nil {
		rep.Checks = append(rep.Checks, Check{"workspace-db", false, err.Error()})
		return rep, nil
	}
	defer wdb.Close()
	wv, _ := wdb.SchemaVersion(ctx)
	rep.Checks = append(rep.Checks, Check{"workspace-db", true, fmt.Sprintf("%s (schema v%d)", wdb.Path(), wv)})

	// 4. Emit an enveloped event to persisted storage, via the bus and the event store.
	bus := eventbus.New()
	defer bus.Close()
	tel := telemetry.New()
	_ = tel.EmitMetric(ctx, ports.MetricSample{Name: "doctor.runs", Value: 1})

	store := storage.NewEventStore(wdb)
	ev := eventbus.NewEvent("runtime.doctor.completed", "andromeda-doctor",
		eventbus.WithPayload([]byte(`{"ms":1}`)))
	if err := bus.Publish(ctx, ev); err != nil {
		rep.Checks = append(rep.Checks, Check{"events", false, err.Error()})
		return rep, nil
	}
	if _, err := store.Append(ctx, ev); err != nil {
		rep.Checks = append(rep.Checks, Check{"events", false, err.Error()})
		return rep, nil
	}
	n, _ := store.Count(ctx)
	rep.Checks = append(rep.Checks, Check{"events", true, fmt.Sprintf("emitted and persisted (%d total)", n)})

	return rep, nil
}
