package ports

// Contract types in this package are minimal at the architecture-skeleton stage (EP-02):
// they fix the frozen method signatures (FR-ARCH-003) while their owning volumes flesh out
// fields additively in later epics. Types shared by more than one port are declared here once.

// Path is a filesystem path in the workspace, always POSIX-style as Git stores it.
type Path = string

// ExecutionID identifies one command execution, linking SandboxPort and TerminalPort.
type ExecutionID = ULID

// CommandSpec describes a command to run. PTY-versus-pipe mode and limits are part of it.
// Shared by SandboxPort.ExecuteIn and TerminalPort.Execute. Full contract: Volume 6.
type CommandSpec struct {
	Program string
	Args    []string
	Dir     Path
	Env     map[string]string
	PTY     bool
}

// VerificationReport is the result of an integrity/signature verification. Shared by
// UpdaterPort.Verify and PackagePort.Verify. Full contract: Volumes 14 and 6.
type VerificationReport struct {
	OK       bool
	Checksum bool
	Signed   bool
	Findings []string
}
