package pal

import (
	"context"
	"errors"
	"io"
	"time"
)

// The 19 platform surfaces of Volume 3 chapter 07. Method sets are minimal at the
// architecture-skeleton stage and grow additively as owning epics need them; the surface
// boundary is what matters — no platform dependency exists outside these interfaces.

// Filesystem abstracts file and directory operations with symlink-aware semantics.
type Filesystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm uint32) error
	MkdirAll(path string, perm uint32) error
	Remove(name string) error
	Stat(name string) (FileInfo, error)
	ReadDir(name string) ([]DirEntry, error)
}

// FileInfo is a minimal, platform-neutral file description.
type FileInfo struct {
	Name    string
	Size    int64
	Mode    uint32
	IsDir   bool
	ModTime time.Time
	Symlink bool
}

// DirEntry is one directory entry.
type DirEntry struct {
	Name  string
	IsDir bool
}

// Paths resolves and normalizes filesystem paths.
type Paths interface {
	Abs(path string) (string, error)
	Clean(path string) string
	Join(elem ...string) string
	Home() (string, error)
}

// Permissions abstracts POSIX permission checks and changes.
type Permissions interface {
	Chmod(name string, perm uint32) error
	IsExecutable(name string) (bool, error)
}

// Processes abstracts process creation and control (single process; trees via ProcessTrees).
type Processes interface {
	Start(ctx context.Context, spec ProcessSpec) (Process, error)
}

// ProcessSpec describes a process to start.
type ProcessSpec struct {
	Program string
	Args    []string
	Dir     string
	Env     []string
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
}

// Process is a running process handle.
type Process interface {
	PID() int
	Wait(ctx context.Context) (int, error)
	Kill() error
}

// Signals maps portable signal names to platform signals.
type Signals interface {
	Send(pid int, name string) error
}

// PTY abstracts pseudoterminal allocation.
type PTY interface {
	Open(ctx context.Context, spec ProcessSpec) (PTYHandle, error)
}

// PTYHandle is an allocated PTY with its child process.
type PTYHandle interface {
	io.ReadWriteCloser
	Resize(cols, rows int) error
	Process() Process
}

// Shell resolves the user's shell and quoting rules.
type Shell interface {
	Default() (string, error)
	Quote(arg string) string
}

// ErrCredentialNotFound is returned by a CredentialStore's Get/Delete when the service/account
// pair has no stored material. It is the platform-neutral not-found signal the Secret Store maps
// onto its own sentinel, so callers never depend on a specific OS keychain's error text.
var ErrCredentialNotFound = errors.New("pal: credential not found")

// CredentialStore is the platform keychain backend used by the Secret Store (ADR-014).
// Get and Delete report a missing item as ErrCredentialNotFound.
type CredentialStore interface {
	Get(service, account string) ([]byte, error)
	Set(service, account string, secret []byte) error
	Delete(service, account string) error
	Available() bool
}

// Notifications delivers desktop notifications where available.
type Notifications interface {
	Notify(title, body string) error
}

// Clipboard reads and writes the system clipboard.
type Clipboard interface {
	Read() (string, error)
	Write(text string) error
}

// Installer places and removes installed artifacts (used by the Updater).
type Installer interface {
	ReplaceBinary(ctx context.Context, newPath, targetPath string) error
	Backup(targetPath string) (string, error)
	RestoreBackup(backupPath, targetPath string) error
}

// Updater surfaces platform-specific update mechanics beyond the generic UpdaterPort.
type Updater interface {
	SelfPath() (string, error)
}

// Sandbox exposes the platform isolation mechanism selection (ADR-021 layered model).
type Sandbox interface {
	AvailableMechanisms() []string // e.g. "process", "seatbelt", "landlock"
}

// ToolDiscovery locates external tools on the platform (PATH resolution and well-known dirs).
type ToolDiscovery interface {
	Lookup(name string) (string, error)
}

// ConfigDirs resolves XDG-style configuration, data, cache, and runtime directories (ADR-022).
type ConfigDirs interface {
	ConfigHome() (string, error)
	DataHome() (string, error)
	CacheHome() (string, error)
	RuntimeDir() (string, error)
}

// FileLocking provides advisory file locks (used by the encrypted-file secret fallback and
// the single-writer workspace-database discipline).
type FileLocking interface {
	Acquire(ctx context.Context, path string, exclusive bool) (FileLock, error)
}

// FileLock is a held advisory lock.
type FileLock interface {
	Release() error
}

// LocalIPC provides Unix-domain-socket local IPC (ADR-012 external IPC) with peer verification.
type LocalIPC interface {
	Listen(ctx context.Context, path string) (Listener, error)
	Dial(ctx context.Context, path string) (Conn, error)
}

// Listener accepts local IPC connections.
type Listener interface {
	Accept(ctx context.Context) (Conn, error)
	Close() error
}

// Conn is a local IPC connection with verified peer credentials.
type Conn interface {
	io.ReadWriteCloser
	PeerUID() (int, error)
}

// TempFiles creates temporary files and directories with safe permissions.
type TempFiles interface {
	TempDir(pattern string) (string, error)
	TempFile(pattern string) (string, error)
}

// ProcessTrees terminates whole process trees (for sandbox teardown and cancellation).
type ProcessTrees interface {
	KillTree(pid int) error
}
