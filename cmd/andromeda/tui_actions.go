package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/buildinfo"
	"github.com/datamaia/andromeda/internal/memory"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/skill"
	"github.com/datamaia/andromeda/internal/storage"
	"github.com/datamaia/andromeda/internal/tui"
	"github.com/datamaia/andromeda/internal/updater"
	"github.com/datamaia/andromeda/internal/workflow"
)

// sessionActions wires the TUI slash commands to the real app operations. Each handler returns the
// text the palette shows in the transcript; errors are formatted, not thrown, so a failing command
// never tears down the session.
func (s *tuiSession) sessionActions() tui.Actions {
	return tui.Actions{
		Doctor:    s.doctorAction,
		Update:    s.updateAction,
		Memory:    s.memoryAction,
		Workflows: s.workflowsAction,
		MCP:       s.mcpAction,
		Skills:    s.skillsAction,
		Models:    s.modelsAction,
	}
}

func (s *tuiSession) doctorAction(ctx context.Context) string {
	rep, err := app.Doctor(ctx, s.wd)
	if err != nil {
		return "doctor: " + err.Error()
	}
	var b strings.Builder
	status := "all checks passed"
	if !rep.OK() {
		status = "some checks failed"
	}
	b.WriteString("doctor · " + status)
	for _, c := range rep.Checks {
		mark := "✓"
		if !c.OK {
			mark = "✗"
		}
		fmt.Fprintf(&b, "\n  %s %-13s %s", mark, c.Name, c.Detail)
	}
	return b.String()
}

func (s *tuiSession) updateAction(ctx context.Context) string {
	self, _ := os.Executable()
	// No release source is wired for a from-source dev build; Check reports "up to date" cleanly.
	u := updater.New(buildinfo.Get().Version, "stable", self, nil)
	res, err := u.Check(ctx)
	if err != nil {
		return "update: " + err.Error()
	}
	if res.Status == "update_available" {
		return fmt.Sprintf("update available: %s → %s (channel %s)", res.Current, res.Latest, res.Channel)
	}
	return fmt.Sprintf("up to date: %s (channel %s)", res.Current, res.Channel)
}

func (s *tuiSession) memoryAction(ctx context.Context, args string) string {
	sub, rest, _ := strings.Cut(strings.TrimSpace(args), " ")
	rest = strings.TrimSpace(rest)
	db, err := storage.OpenWorkspaceDB(ctx, s.wd)
	if err != nil {
		return "memory: " + err.Error()
	}
	defer func() { _ = db.Close() }()
	store := memory.New(db)

	switch sub {
	case "add":
		if rest == "" {
			return "usage: /memory add <content>"
		}
		ids, err := store.Ingest(ctx, []ports.MemoryRecordDraft{{Layer: "workspace", Content: rest, Source: "tui"}})
		if err != nil {
			return "memory: " + err.Error()
		}
		return "remembered " + ids[0]
	case "search":
		if rest == "" {
			return "usage: /memory search <query>"
		}
		return formatMemory(store.Retrieve(ctx, ports.MemoryQuery{Text: rest, Limit: 20}))
	case "", "list":
		return formatMemory(store.Retrieve(ctx, ports.MemoryQuery{Limit: 20}))
	default:
		return "memory subcommands: list · add <content> · search <query>"
	}
}

func formatMemory(recs []ports.MemoryRecord, err error) string {
	if err != nil {
		return "memory: " + err.Error()
	}
	if len(recs) == 0 {
		return "no memories yet — add one with /memory add <content>"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "memory · %d record(s)", len(recs))
	for _, r := range recs {
		fmt.Fprintf(&b, "\n  [%s] %s", r.Layer, r.Content)
	}
	return b.String()
}

func (s *tuiSession) workflowsAction(context.Context) string {
	var b strings.Builder
	b.WriteString("workflows · built-in SDD stages (run: andromeda workflow run sdd)")
	for i, name := range workflow.SDDStageNames() {
		fmt.Fprintf(&b, "\n  %2d. %s", i+1, name)
	}
	return b.String()
}

func (s *tuiSession) mcpAction(context.Context) string {
	// MCP servers are declared in andromeda.toml under [mcp.servers] and connected at agent start;
	// none are wired into this session shell yet.
	return "mcp · no servers connected in this session — declare them under [mcp.servers] in andromeda.toml"
}

func (s *tuiSession) skillsAction(context.Context) string {
	dir := filepath.Join(s.wd, ".andromeda", "skills")
	ents, err := os.ReadDir(dir)
	if err != nil {
		return "skills · none found — add one under .andromeda/skills/<name>/skill.toml"
	}
	var found []string
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		sk, err := skill.Load(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		desc := sk.Manifest.Description
		if desc != "" {
			desc = " — " + desc
		}
		found = append(found, fmt.Sprintf("%s@%s%s", sk.Manifest.Name, sk.Manifest.Version, desc))
	}
	if len(found) == 0 {
		return "skills · none found — add one under .andromeda/skills/<name>/skill.toml"
	}
	sort.Strings(found)
	return "skills · " + fmt.Sprint(len(found)) + " available\n  " + strings.Join(found, "\n  ")
}

func (s *tuiSession) modelsAction(ctx context.Context) []string {
	if s.prov == nil {
		return nil
	}
	descs, err := s.prov.DiscoverModels(ctx)
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(descs))
	for _, d := range descs {
		ids = append(ids, d.ID)
	}
	sort.Strings(ids)
	return ids
}
