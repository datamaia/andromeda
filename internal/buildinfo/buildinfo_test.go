package buildinfo

import "testing"

func TestGetReturnsDefaults(t *testing.T) {
	i := Get()
	if i.Version == "" {
		t.Error("version must not be empty")
	}
	if i.GoOS == "" || i.GoArch == "" {
		t.Errorf("expected GOOS/GOARCH to be populated, got %q/%q", i.GoOS, i.GoArch)
	}
}
