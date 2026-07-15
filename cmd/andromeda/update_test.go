package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/buildinfo"
	"github.com/datamaia/andromeda/internal/ports"
)

func TestDescribeUpdate(t *testing.T) {
	base := buildinfo.Info{Commit: "abcdef1234567", Date: "2026-07-15T00:00:00Z", GoOS: "darwin", GoArch: "arm64"}

	t.Run("dev build explains itself and points at releases", func(t *testing.T) {
		i := base
		i.Version = "0.0.0-dev"
		got := describeUpdate(ports.UpdateCheckResult{Channel: "stable", Latest: "0.1.6"}, i, nil)
		for _, want := range []string{"development build", "latest published release: v0.1.6", "brew upgrade andromeda"} {
			if !strings.Contains(got, want) {
				t.Errorf("dev output missing %q:\n%s", want, got)
			}
		}
	})

	t.Run("release up to date", func(t *testing.T) {
		i := base
		i.Version = "0.1.6"
		got := describeUpdate(ports.UpdateCheckResult{Channel: "stable", Latest: "0.1.6", Status: "up_to_date"}, i, nil)
		if !strings.Contains(got, "latest release (v0.1.6)") {
			t.Errorf("up-to-date output = %s", got)
		}
	})

	t.Run("update available shows how to upgrade", func(t *testing.T) {
		i := base
		i.Version = "0.1.5"
		got := describeUpdate(ports.UpdateCheckResult{Channel: "stable", Latest: "0.1.6", Status: "update_available"}, i, nil)
		if !strings.Contains(got, "v0.1.5 → v0.1.6") || !strings.Contains(got, "brew upgrade") {
			t.Errorf("update-available output = %s", got)
		}
	})

	t.Run("offline is reported, not fatal", func(t *testing.T) {
		i := base
		i.Version = "0.1.6"
		got := describeUpdate(ports.UpdateCheckResult{Channel: "stable"}, i, errors.New("dial tcp: no route to host"))
		if !strings.Contains(got, "couldn't reach the release feed") || !strings.Contains(got, "no route to host") {
			t.Errorf("offline output = %s", got)
		}
	})
}

func TestIsDevVersion(t *testing.T) {
	for _, v := range []string{"", "dev", "0.0.0-dev", "1.2.3-dev.4"} {
		if !isDevVersion(v) {
			t.Errorf("isDevVersion(%q) = false, want true", v)
		}
	}
	for _, v := range []string{"0.1.6", "1.0.0"} {
		if isDevVersion(v) {
			t.Errorf("isDevVersion(%q) = true, want false", v)
		}
	}
}
