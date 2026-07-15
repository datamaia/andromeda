package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/storage"
)

// backgroundAction backs /background: it launches an unattended `andromeda run` for the goal as a
// detached child process, logging to a file under .andromeda/background/, and returns immediately so
// the composer stays free. A background agent cannot answer approval prompts, so it runs with
// workspace write granted by default (revertible with /undo) and command execution only when the
// user opts in with --exec. The grants are stated plainly in the reply.
func (s *tuiSession) backgroundAction(_ context.Context, args string) string {
	args = strings.TrimSpace(args)
	allowExec := false
	if rest, ok := strings.CutPrefix(args, "--exec"); ok {
		allowExec = true
		args = strings.TrimSpace(rest)
	}
	if args == "" {
		return "usage: /background [--exec] <goal> — runs an unattended agent in the background\n" +
			"grants workspace write by default (revert with /undo); add --exec to also allow commands"
	}
	self, err := os.Executable()
	if err != nil {
		return "background: cannot locate the andromeda binary: " + err.Error()
	}
	logDir := filepath.Join(s.wd, storage.MarkerDir, "background")
	if err := os.MkdirAll(logDir, 0o750); err != nil {
		return "background: " + err.Error()
	}
	id := time.Now().UTC().Format("20060102-150405")
	logPath := filepath.Join(logDir, id+".log")
	logFile, err := os.Create(logPath) //nolint:gosec // fixed path under the workspace marker dir
	if err != nil {
		return "background: " + err.Error()
	}

	cmdArgs := []string{"run", args, "--provider", s.cfg.provider, "--model", s.cfg.model, "--allow-write"}
	if s.cfg.baseURL != "" {
		cmdArgs = append(cmdArgs, "--base-url", s.cfg.baseURL)
	}
	if s.cfg.apiKeyEnv != "" {
		cmdArgs = append(cmdArgs, "--api-key-env", s.cfg.apiKeyEnv)
	}
	if allowExec {
		cmdArgs = append(cmdArgs, "--allow-exec")
	}
	cmd := exec.Command(self, cmdArgs...) //nolint:gosec // self is our own binary; args are controlled
	cmd.Dir = s.wd
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	pid, err := startBackground(cmd)
	if err != nil {
		_ = logFile.Close()
		return "background: " + err.Error()
	}

	grants := "write"
	if allowExec {
		grants = "write + exec"
	}
	return fmt.Sprintf("started background task %s (pid %d · grants: %s)\n  goal: %s\n  log:  %s\n"+
		"runs unattended for this session; revert its file changes with /undo",
		id, pid, grants, args, relOr(s.wd, logPath))
}

// startBackground detaches and starts the process, reaping it so it never becomes a zombie, and
// returns its pid. It is a package var so tests can stub it — a test must never actually spawn the
// test binary.
var startBackground = func(cmd *exec.Cmd) (int, error) {
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	pid := cmd.Process.Pid
	go func() { _ = cmd.Wait() }()
	return pid, nil
}
