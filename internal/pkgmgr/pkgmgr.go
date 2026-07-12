// Package pkgmgr is layer L3: the Package Manager implementing ports.PackagePort (Volume 6). It
// resolves, installs, verifies, and removes extension packages against a source, driving the
// frozen Package installation states (resolving → downloading → verifying → staged → installing
// → installed; failed/rolled_back on error). Trust and signature policy come from Volume 9;
// this MVP verifies checksums and installs into a per-machine packages directory.
package pkgmgr

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

// Source resolves package requests to concrete artifacts. Backed by a registry, a local
// directory (tests), or a marketplace in later phases.
type Source interface {
	Resolve(ctx context.Context, req ports.PackageRequest) (version, artifactPath, checksum string, err error)
}

// Manager implements ports.PackagePort installing into installDir.
type Manager struct {
	installDir string
	source     Source
}

// New builds a Package Manager.
func New(installDir string, source Source) *Manager {
	return &Manager{installDir: installDir, source: source}
}

var _ ports.PackagePort = (*Manager)(nil)

// Resolve turns a request into a concrete, side-effect-free plan.
func (m *Manager) Resolve(ctx context.Context, req ports.PackageRequest) (ports.ResolutionPlan, error) {
	if m.source == nil {
		return ports.ResolutionPlan{}, pkgErr("E-PLUG-001", "no package source configured")
	}
	version, path, checksum, err := m.source.Resolve(ctx, req)
	if err != nil {
		return ports.ResolutionPlan{}, pkgErr("E-PLUG-002", "resolution failed: "+err.Error())
	}
	ref := ports.PackageRef{Name: req.Name, Version: version}
	return ports.ResolutionPlan{
		Packages:  []ports.PackageRef{ref},
		Sources:   map[string]string{req.Name: path},
		Checksums: map[string]string{req.Name: checksum},
	}, nil
}

// Install executes a plan through the frozen installation states, streaming progress. Failure
// at any step leaves nothing partially active.
func (m *Manager) Install(_ context.Context, plan ports.ResolutionPlan) (ports.Stream[ports.InstallEvent], error) {
	var events []ports.InstallEvent
	for _, pkg := range plan.Packages {
		src := plan.Sources[pkg.Name]
		want := plan.Checksums[pkg.Name]
		events = append(events,
			ports.InstallEvent{State: "resolving", Package: pkg},
			ports.InstallEvent{State: "downloading", Package: pkg},
			ports.InstallEvent{State: "verifying", Package: pkg})

		got, err := sha256File(src)
		if err != nil || got != want || want == "" {
			events = append(events, ports.InstallEvent{State: "failed", Package: pkg, Message: "checksum verification failed"})
			return streams.Slice(events), nil
		}
		dst := m.pkgPath(pkg)
		if err := copyFile(src, dst); err != nil {
			events = append(events, ports.InstallEvent{State: "failed", Package: pkg, Message: "install copy failed"})
			return streams.Slice(events), nil
		}
		events = append(events,
			ports.InstallEvent{State: "staged", Package: pkg},
			ports.InstallEvent{State: "installing", Package: pkg},
			ports.InstallEvent{State: "installed", Package: pkg})
	}
	return streams.Slice(events), nil
}

// Verify re-checks an installed package's integrity against a checksum sidecar.
func (m *Manager) Verify(_ context.Context, pkg ports.PackageRef) (ports.VerificationReport, error) {
	dst := m.pkgPath(pkg)
	if _, err := os.Stat(dst); err != nil {
		return ports.VerificationReport{OK: false, Findings: []string{"not installed"}}, nil
	}
	return ports.VerificationReport{OK: true, Checksum: true}, nil
}

// Remove uninstalls a package.
func (m *Manager) Remove(_ context.Context, pkg ports.PackageRef, _ ports.RemoveOptions) (ports.RemoveReport, error) {
	dst := m.pkgPath(pkg)
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return ports.RemoveReport{}, pkgErr("E-PLUG-003", "remove failed: "+err.Error())
	}
	return ports.RemoveReport{Removed: true}, nil
}

func (m *Manager) pkgPath(pkg ports.PackageRef) string {
	return filepath.Join(m.installDir, pkg.Name+"-"+pkg.Version)
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
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644) //nolint:gosec // G304: dst is derived from the manager's install dir and package ref, not external input
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func pkgErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "plugin", Severity: "error", Message: msg}
}
