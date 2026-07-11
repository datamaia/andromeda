// Command andromeda is the composition root of the Andromeda AI engineering harness.
//
// This is the walking-skeleton entrypoint (EP-01). The full CLI grammar is specified in
// Volume 8 of docs/spec and is built out in later epics; for now the binary starts, reports
// its version, and exits with the codes defined in Volume 0 chapter 03 (ADR-016).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/buildinfo"
	"github.com/spf13/cobra"
)

// Exit codes (subset of the Volume 0 / ADR-016 scheme used at this stage).
const (
	exitOK      = 0
	exitGeneral = 1
	exitUsage   = 2
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	root := newRootCommand()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		// cobra already prints the error and usage for usage errors.
		return exitUsage
	}
	return exitOK
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "andromeda",
		Short:         "Andromeda — your terminal companion for shipping great software.",
		Long:          "Andromeda is an open-source, local-first, vendor-agnostic AI engineering harness (CLI + TUI).",
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.CompletionOptions.HiddenDefaultCmd = true
	root.AddCommand(newVersionCommand())
	root.AddCommand(newDoctorCommand())
	return root
}

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the environment and foundation (config, databases, events)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			rep, err := app.Doctor(context.Background(), wd)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			for _, c := range rep.Checks {
				mark := "ok  "
				if !c.OK {
					mark = "FAIL"
				}
				fmt.Fprintf(out, "[%s] %-13s %s\n", mark, c.Name, c.Detail)
			}
			if !rep.OK() {
				return fmt.Errorf("one or more checks failed")
			}
			fmt.Fprintln(out, "doctor: all checks passed")
			return nil
		},
	}
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build metadata",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			i := buildinfo.Get()
			_, err := fmt.Fprintf(cmd.OutOrStdout(),
				"andromeda %s (commit %s, built %s, %s/%s)\n",
				i.Version, i.Commit, i.Date, i.GoOS, i.GoArch)
			return err
		},
	}
}
