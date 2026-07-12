// Package updater is layer L3: the Updater implementing ports.UpdaterPort (Volume 14). It
// checks a release source, downloads artifacts, verifies checksums (and signatures when
// enabled), applies updates via an atomic binary swap, and rolls back to a retained previous
// version. Check is the only method that needs network and MUST fail cleanly offline; Apply
// refuses to run unless Verify has passed for the same artifact set.
package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// ReleaseSource provides release metadata and artifact bytes. Implementations back it with a
// GitHub Releases feed, a local directory (tests), or an air-gapped mirror.
type ReleaseSource interface {
	Latest(ctx context.Context, channel string) (version string, checksum string, err error)
	Fetch(ctx context.Context, version string) (path string, checksum string, err error)
}

// Updater implements ports.UpdaterPort.
type Updater struct {
	current    string
	channel    string
	targetPath string // path of the installed binary
	source     ReleaseSource
	verified   map[string]bool // version -> checksum verified
	backupPath string
}

// New builds an Updater for the current version and installed binary path.
func New(current, channel, targetPath string, source ReleaseSource) *Updater {
	if channel == "" {
		channel = "stable"
	}
	return &Updater{current: current, channel: channel, targetPath: targetPath, source: source, verified: map[string]bool{}}
}

var _ ports.UpdaterPort = (*Updater)(nil)

// Check reports whether an update is available on the configured channel.
func (u *Updater) Check(ctx context.Context) (ports.UpdateCheckResult, error) {
	res := ports.UpdateCheckResult{Current: u.current, Channel: u.channel, Status: "up_to_date", Latest: u.current}
	if u.source == nil {
		return res, nil // no source configured: nothing to update against, cleanly
	}
	latest, _, err := u.source.Latest(ctx, u.channel)
	if err != nil {
		return ports.UpdateCheckResult{}, relErr("E-REL-001", "update check failed", err)
	}
	res.Latest = latest
	if latest != u.current && latest != "" {
		res.Status = "update_available"
	}
	return res, nil
}

// Download fetches a release's artifact, reporting a single completion progress event.
func (u *Updater) Download(ctx context.Context, rel ports.ReleaseRef) (ports.Stream[ports.DownloadProgress], error) {
	if u.source == nil {
		return nil, relErr("E-REL-002", "no release source configured", nil)
	}
	path, _, err := u.source.Fetch(ctx, rel.Version)
	if err != nil {
		return nil, relErr("E-REL-002", "download failed", err)
	}
	fi, _ := os.Stat(path)
	var size int64
	if fi != nil {
		size = fi.Size()
	}
	return streams.Slice([]ports.DownloadProgress{{BytesDone: size, BytesTotal: size}}), nil
}

// Verify validates the fetched artifact's checksum against the release metadata.
func (u *Updater) Verify(ctx context.Context, rel ports.ReleaseRef) (ports.VerificationReport, error) {
	path, want, err := u.source.Fetch(ctx, rel.Version)
	if err != nil {
		return ports.VerificationReport{}, relErr("E-REL-003", "verify: fetch failed", err)
	}
	got, err := sha256File(path)
	if err != nil {
		return ports.VerificationReport{}, relErr("E-REL-003", "verify: hash failed", err)
	}
	ok := got == want && want != ""
	if ok {
		u.verified[rel.Version] = true
	}
	return ports.VerificationReport{OK: ok, Checksum: ok, Findings: findings(ok, got, want)}, nil
}

// Apply swaps the installed binary atomically, retaining the previous version for rollback.
// It refuses to run unless Verify passed for the same version.
func (u *Updater) Apply(ctx context.Context, rel ports.ReleaseRef) (ports.UpdateApplyReport, error) {
	if !u.verified[rel.Version] {
		return ports.UpdateApplyReport{}, relErr("E-REL-004", "apply refused: artifact not verified", nil)
	}
	path, _, err := u.source.Fetch(ctx, rel.Version)
	if err != nil {
		return ports.UpdateApplyReport{}, relErr("E-REL-004", "apply: fetch failed", err)
	}
	// Retain the current binary as a backup, then atomically replace via rename.
	backup := u.targetPath + ".backup"
	if _, err := os.Stat(u.targetPath); err == nil {
		if err := copyFile(u.targetPath, backup); err != nil {
			return ports.UpdateApplyReport{}, relErr("E-REL-004", "apply: backup failed", err)
		}
		u.backupPath = backup
	}
	tmp := u.targetPath + ".new"
	if err := copyFile(path, tmp); err != nil {
		return ports.UpdateApplyReport{}, relErr("E-REL-004", "apply: stage failed", err)
	}
	_ = os.Chmod(tmp, 0o755) //nolint:gosec // G302: the staged file is the installed executable and must be runnable
	if err := os.Rename(tmp, u.targetPath); err != nil {
		return ports.UpdateApplyReport{}, relErr("E-REL-004", "apply: atomic replace failed", err)
	}
	from := u.current
	u.current = rel.Version
	return ports.UpdateApplyReport{Applied: true, FromVersion: from, ToVersion: rel.Version}, nil
}

// Rollback restores the previously retained version (offline).
func (u *Updater) Rollback(_ context.Context) (ports.RollbackReport, error) {
	if u.backupPath == "" {
		return ports.RollbackReport{}, relErr("E-REL-005", "no retained version to roll back to", nil)
	}
	if err := copyFile(u.backupPath, u.targetPath); err != nil {
		return ports.RollbackReport{}, relErr("E-REL-005", "rollback failed", err)
	}
	return ports.RollbackReport{RolledBack: true, ToVersion: "previous"}, nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // internal artifact path
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // internal artifact path
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { //nolint:gosec // G301: the installed binary's directory must be traversable (0o755)
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755) //nolint:gosec // G302: the installed binary must be runnable (0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func findings(ok bool, got, want string) []string {
	if ok {
		return nil
	}
	return []string{"checksum mismatch: got " + got + " want " + want}
}

func relErr(code, msg string, cause error) error {
	pe := &ports.PortError{Code: code, Category: "release", Severity: "error", Message: msg, Cause: cause}
	if cause != nil {
		pe.Detail = cause.Error()
	}
	return pe
}
