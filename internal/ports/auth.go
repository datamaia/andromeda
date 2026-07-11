package ports

import "context"

// AuthPort runs provider authentication flows — acquisition, refresh, revocation, rotation —
// always via official mechanisms (Volume 1 constraint). Contract owner: Volume 5. Errors:
// E-AUTH (exit code 4 at the CLI boundary).
type AuthPort interface {
	Authenticate(ctx context.Context, spec AuthSpec) (AuthenticationHandle, error)
	Refresh(ctx context.Context, h AuthenticationHandle) (AuthenticationHandle, error)
	Revoke(ctx context.Context, h AuthenticationHandle) error
	Rotate(ctx context.Context, credentialID ULID) (RotationReport, error)
	ListProfiles(ctx context.Context) ([]AuthProfile, error)
}

// AuthSpec selects the provider and official mechanism to authenticate with.
type AuthSpec struct {
	Provider  string
	Profile   string
	Mechanism string // "api_key" | "oauth_device" | "oauth_browser" | "none" | ...
}

// AuthenticationHandle references an active Authentication Session; it carries no material.
type AuthenticationHandle struct {
	SessionID ULID
	Provider  string
	Profile   string
}

// RotationReport describes the outcome of a credential rotation.
type RotationReport struct {
	CredentialID ULID
	Status       string // "active" | "rotated" | "revoked" | "expired"
}

// AuthProfile is a named credential+provider binding, without secret material.
type AuthProfile struct {
	Name     string
	Provider string
}
