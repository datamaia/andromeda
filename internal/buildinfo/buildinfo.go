// Package buildinfo exposes version and build metadata for the andromeda binary.
//
// Values are overridden at release time via -ldflags (see .goreleaser.yaml, ADR-013).
// In development builds they carry the placeholders below, augmented by the Go module
// build info when available.
package buildinfo

import "runtime/debug"

// Overridable via -ldflags "-X github.com/datamaia/andromeda/internal/buildinfo.version=...".
var (
	version = "0.0.0-dev"
	commit  = "unknown"
	date    = "unknown"
)

// Info is a snapshot of the binary's build metadata.
type Info struct {
	Version string
	Commit  string
	Date    string
	GoOS    string
	GoArch  string
}

// Get returns the current build metadata, filling gaps from the embedded module build info.
func Get() Info {
	i := Info{Version: version, Commit: commit, Date: date}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if i.Commit == "unknown" && s.Value != "" {
					i.Commit = s.Value
				}
			case "vcs.time":
				if i.Date == "unknown" && s.Value != "" {
					i.Date = s.Value
				}
			case "GOOS":
				i.GoOS = s.Value
			case "GOARCH":
				i.GoArch = s.Value
			}
		}
	}
	return i
}
