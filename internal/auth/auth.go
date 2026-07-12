// Package auth is layer L3: the Authentication Layer implementing ports.AuthPort (Volume 5).
// It uses only official mechanisms (Volume 1 constraint): API-key intake and, in later
// increments, OAuth device/browser flows. Credential material lives only behind Secret Store
// references (ADR-014); this package never returns or logs secrets. A subscription never
// implies programmatic access.
package auth

import (
	"context"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// SecretNamespace is where credential material is stored.
const SecretNamespace = "auth"

// Manager implements ports.AuthPort over a Secret Store.
type Manager struct {
	secrets ports.SecretStorePort
}

// New returns an Authentication Layer over the given Secret Store.
func New(secrets ports.SecretStorePort) *Manager { return &Manager{secrets: secrets} }

var _ ports.AuthPort = (*Manager)(nil)

func refFor(provider, profile string) ports.SecretRef {
	if profile == "" {
		profile = "default"
	}
	return ports.SecretRef{Namespace: SecretNamespace, Name: provider + ":" + profile}
}

// StoreAPIKey stores an API key for a provider profile (the intake path used by the CLI). It is
// the only entry point that accepts raw material; Authenticate later confirms presence.
func (m *Manager) StoreAPIKey(ctx context.Context, provider, profile, key string) error {
	if key == "" {
		return authErr("E-AUTH-002", "empty API key")
	}
	return m.secrets.Set(ctx, refFor(provider, profile),
		ports.NewSecretValue([]byte(key)),
		ports.SecretMeta{Kind: "api_key", Provider: provider})
}

// Authenticate establishes an Authentication Session. For api_key it confirms that stored
// material exists (no network); OAuth flows are a later increment.
func (m *Manager) Authenticate(ctx context.Context, spec ports.AuthSpec) (ports.AuthenticationHandle, error) {
	switch spec.Mechanism {
	case "", "api_key", "none":
		if spec.Mechanism == "none" {
			// Local providers (e.g. Ollama) need no credential.
			return ports.AuthenticationHandle{SessionID: core.NewULID(), Provider: spec.Provider, Profile: spec.Profile}, nil
		}
		if _, err := m.secrets.Get(ctx, refFor(spec.Provider, spec.Profile)); err != nil {
			return ports.AuthenticationHandle{}, authErr("E-AUTH-003", "no stored credential for "+spec.Provider)
		}
		return ports.AuthenticationHandle{SessionID: core.NewULID(), Provider: spec.Provider, Profile: spec.Profile}, nil
	default:
		return ports.AuthenticationHandle{}, authErr("E-AUTH-004", "unsupported mechanism at this phase: "+spec.Mechanism)
	}
}

// Refresh renews a session. API-key sessions do not expire, so the handle is returned as-is.
func (m *Manager) Refresh(_ context.Context, h ports.AuthenticationHandle) (ports.AuthenticationHandle, error) {
	return h, nil
}

// Revoke deletes the stored credential and invalidates the session.
func (m *Manager) Revoke(ctx context.Context, h ports.AuthenticationHandle) error {
	return m.secrets.Delete(ctx, refFor(h.Provider, h.Profile))
}

// Rotate replaces a credential's material. For API keys this is a manual operation: the caller
// re-runs StoreAPIKey. Rotate reports the current status rather than inventing a rotation.
func (m *Manager) Rotate(_ context.Context, credentialID core.ULID) (ports.RotationReport, error) {
	return ports.RotationReport{CredentialID: credentialID, Status: "active"}, nil
}

// ListProfiles enumerates configured authentication profiles without any secret material.
func (m *Manager) ListProfiles(ctx context.Context) ([]ports.AuthProfile, error) {
	refs, err := m.secrets.List(ctx, ports.SecretScope{Namespace: SecretNamespace})
	if err != nil {
		return nil, err
	}
	out := make([]ports.AuthProfile, 0, len(refs))
	for _, r := range refs {
		provider, profile := "", "default"
		if i := strings.IndexByte(r.Name, ':'); i >= 0 {
			provider, profile = r.Name[:i], r.Name[i+1:]
		} else {
			provider = r.Name
		}
		out = append(out, ports.AuthProfile{Name: profile, Provider: provider})
	}
	return out, nil
}

func authErr(code, msg string) error {
	return &ports.PortError{Code: code, Category: "auth", Severity: "error", Message: msg}
}
