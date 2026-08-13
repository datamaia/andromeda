package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/datamaia/andromeda/internal/graph"
	"github.com/datamaia/andromeda/internal/ontology"
	"github.com/spf13/cobra"
)

// newGraphCommand builds `andromeda graph`, which scans the workspace, writes a visual graph model
// (graph.json + Markdown notes) under .andromeda/graph/, and can serve a small self-contained
// force-directed viewer over localhost. Bare invocation builds; subcommands manage and serve it.
func newGraphCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "graph",
		Short: "Build and view a visual graph of the workspace",
		Long: "Scan every file in the workspace (honoring .gitignore) and write a deterministic graph " +
			"model — graph.json plus a self-contained 3D viewer (graph-3d.html) and human-readable " +
			"Markdown notes — to .andromeda/graph/. Then `andromeda graph open` opens the offline 3D " +
			"viewer from disk, or `andromeda graph serve` serves it over localhost with live refresh.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { _, err := runGraphBuild(cmd); return err },
	}
	c.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "Scan the workspace and (re)write the graph",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { _, err := runGraphBuild(cmd); return err },
	})

	var port int
	var noOpen bool
	serve := &cobra.Command{
		Use:   "serve",
		Short: "Build the graph and serve the interactive viewer on localhost",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runGraphServe(cmd, port, noOpen) },
	}
	serve.Flags().IntVar(&port, "port", 0, "localhost port to bind (0 = pick a free port)")
	serve.Flags().BoolVar(&noOpen, "no-open", false, "do not open the system browser")
	c.AddCommand(serve)

	c.AddCommand(&cobra.Command{
		Use:   "open",
		Short: "Build the graph and open the offline 3D viewer (no server)",
		Long: "Build the graph and open .andromeda/graph/graph-3d.html directly in your browser — a " +
			"self-contained, dependency-free 3D view that runs from the filesystem with no localhost " +
			"server. Use `graph serve` instead when you want the live-refreshing localhost viewer.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return runGraphOpen(cmd) },
	})

	c.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the graph overview note (index.md)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runGraphShow(cmd) },
	})
	c.AddCommand(&cobra.Command{
		Use:   "rm",
		Short: "Delete the generated graph",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := graph.Remove(wd); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "removed .andromeda/graph")
			return nil
		},
	})
	return c
}

func runGraphBuild(cmd *cobra.Command) (*graph.Graph, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	m, err := ontology.Scan(cmd.Context(), wd)
	if err != nil {
		return nil, err
	}
	g, dir, err := graph.Write(wd, m)
	if err != nil {
		return nil, err
	}
	rel, relErr := filepath.Rel(wd, dir)
	if relErr != nil {
		rel = dir
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "graph written to %s\n%s\n", rel, g.Stats())
	return g, nil
}

func runGraphServe(cmd *cobra.Command, port int, noOpen bool) error {
	if _, err := runGraphBuild(cmd); err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
	defer stop()

	out := cmd.OutOrStdout()
	return graph.Serve(ctx, wd, port, func(url string) {
		_, _ = fmt.Fprintf(out, "\nviewer ready at %s  (press Ctrl+C to stop)\n", url)
		if !noOpen {
			_ = openBrowser(url)
		}
	})
}

// runGraphOpen rebuilds the graph and opens the self-contained 3D viewer from the filesystem via a
// file:// URL — no localhost server. The relative links in the viewer resolve because it lives two
// levels under the workspace root (.andromeda/graph/).
func runGraphOpen(cmd *cobra.Command) error {
	if _, err := runGraphBuild(cmd); err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	htmlPath := filepath.Join(graph.Dir(wd), "graph-3d.html")
	url := "file://" + filepath.ToSlash(htmlPath)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "opening %s\n", url)
	return openBrowser(url)
}

func runGraphShow(cmd *cobra.Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(graph.Dir(wd), "index.md")) //nolint:gosec // fixed path under the workspace marker dir
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no graph yet — run `andromeda graph build`")
		}
		return err
	}
	_, err = cmd.OutOrStdout().Write(data)
	return err
}
