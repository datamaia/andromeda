package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/datamaia/andromeda/internal/ontology"
	"github.com/spf13/cobra"
)

// newOntologyCommand builds `andromeda ontology`, which deterministically scans the workspace and
// writes a Turtle (.ttl) ontology under .andromeda/ontology/ — a navigable structural map of the
// repo for context engineering. Bare invocation builds; subcommands manage the artifacts.
func newOntologyCommand() *cobra.Command {
	c := &cobra.Command{
		Use:   "ontology",
		Short: "Build a deterministic structural ontology (TTL) of the workspace",
		Long: "Scan every file in the workspace (honoring .gitignore) and write a deterministic " +
			"Turtle ontology to .andromeda/ontology/project.ttl describing how files, directories, " +
			"and data relate — a fast navigation surface for an AI or a person.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return runOntologyBuild(cmd) },
	}
	c.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "Scan the workspace and (re)write the ontology",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runOntologyBuild(cmd) },
	})
	c.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the current ontology (project.ttl)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runOntologyShow(cmd) },
	})
	c.AddCommand(&cobra.Command{
		Use:   "rm",
		Short: "Delete the generated ontology",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := ontology.Remove(wd); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "removed .andromeda/ontology")
			return nil
		},
	})
	return c
}

func runOntologyBuild(cmd *cobra.Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	m, err := ontology.Scan(cmd.Context(), wd)
	if err != nil {
		return err
	}
	path, err := ontology.Write(wd, m)
	if err != nil {
		return err
	}
	rel, relErr := filepath.Rel(wd, path)
	if relErr != nil {
		rel = path
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ontology written to %s\n%s\n", rel, m.Stats())
	return nil
}

func runOntologyShow(cmd *cobra.Command) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(ontology.Dir(wd), "project.ttl")) //nolint:gosec // fixed path under the workspace marker dir
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no ontology yet — run `andromeda ontology build`")
		}
		return err
	}
	_, err = cmd.OutOrStdout().Write(data)
	return err
}
