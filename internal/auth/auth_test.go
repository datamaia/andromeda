package auth

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/secret"
)

func newManager(t *testing.T) *Manager {
	t.Helper()
	fb, err := secret.NewFileBackend(filepath.Join(t.TempDir(), "s.age"), "pw")
	if err != nil {
		t.Fatal(err)
	}
	return New(secret.NewStore(fb))
}

func TestStoreAuthenticateListRevoke(t *testing.T) {
	ctx := context.Background()
	m := newManager(t)

	if err := m.StoreAPIKey(ctx, "anthropic", "default", "sk-123"); err != nil {
		t.Fatal(err)
	}
	h, err := m.Authenticate(ctx, ports.AuthSpec{Provider: "anthropic", Profile: "default", Mechanism: "api_key"})
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if h.Provider != "anthropic" || h.SessionID == "" {
		t.Fatalf("handle = %+v", h)
	}

	profiles, err := m.ListProfiles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 || profiles[0].Provider != "anthropic" {
		t.Fatalf("profiles = %+v", profiles)
	}

	if err := m.Revoke(ctx, h); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Authenticate(ctx, ports.AuthSpec{Provider: "anthropic", Mechanism: "api_key"}); err == nil {
		t.Fatal("authenticate should fail after revoke")
	}
}

func TestAuthenticateNoneNeedsNoCredential(t *testing.T) {
	ctx := context.Background()
	m := newManager(t)
	h, err := m.Authenticate(ctx, ports.AuthSpec{Provider: "ollama", Mechanism: "none"})
	if err != nil || h.Provider != "ollama" {
		t.Fatalf("none mechanism should succeed: %+v %v", h, err)
	}
}

func TestUnsupportedMechanism(t *testing.T) {
	ctx := context.Background()
	m := newManager(t)
	_, err := m.Authenticate(ctx, ports.AuthSpec{Provider: "x", Mechanism: "oauth_device"})
	pe, ok := err.(*ports.PortError)
	if !ok || pe.Code != "E-AUTH-004" {
		t.Fatalf("want E-AUTH-004, got %v", err)
	}
}

func TestEmptyKeyRejected(t *testing.T) {
	if err := newManager(t).StoreAPIKey(context.Background(), "p", "d", ""); err == nil {
		t.Fatal("empty key should be rejected")
	}
}
