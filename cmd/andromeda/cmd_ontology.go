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
		Long: "Scan every file in the workspace (honoring .gitignore) and write two deterministic " +
			"Turtle ontologies to .andromeda/ontology/: project.ttl (files, directories, and how they " +
			"relate) and code.ttl (an AST-level graph of Go packages, types, functions, and their " +
			"import/call/implements correlations) — a fast navigation surface for an AI or a person.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return runOntologyBuild(cmd) },
	}
	c.AddCommand(&cobra.Command{
		Use:   "build",
		Short: "Scan the workspace and (re)write the ontology (project.ttl + code.ttl)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runOntologyBuild(cmd) },
	})
	c.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the structural ontology (project.ttl)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runOntologyShow(cmd) },
	})
	c.AddCommand(&cobra.Command{
		Use:   "code",
		Short: "Print the AST-level code graph (code.ttl)",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return runOntologyCode(cmd) },
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
	m, cm, err := ontology.Generate(cmd.Context(), wd)
	if err != nil {
		return err
	}
	rel, relErr := filepath.Rel(wd, ontology.Dir(wd))
	if relErr != nil {
		rel = ontology.Dir(wd)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ontology written to %s\n%s\n%s\n", rel, m.Stats(), cm.Stats())
	return nil
}

func runOntologyShow(cmd *cobra.Command) error {
	return printOntologyFile(cmd, "project.ttl")
}

func runOntologyCode(cmd *cobra.Command) error {
	return printOntologyFile(cmd, "code.ttl")
}

func printOntologyFile(cmd *cobra.Command, name string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filepath.Join(ontology.Dir(wd), name)) //nolint:gosec // fixed path under the workspace marker dir
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no ontology yet — run `andromeda ontology build`")
		}
		return err
	}
	_, err = cmd.OutOrStdout().Write(data)
	return err
}
