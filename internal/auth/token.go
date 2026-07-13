package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
)

// OAuthToken is a stored OAuth credential bundle for a provider profile (ADR-014). Like all
// credential material it lives only behind a Secret Store reference and is never logged.
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	AccountID    string    `json:"account_id,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

// Expired reports whether the access token is at or past its expiry, with a skew margin so it is
// refreshed just before it actually lapses. A zero expiry is treated as never-expiring.
func (t OAuthToken) Expired(skew time.Duration) bool {
	if t.Expiry.IsZero() {
		return false
	}
	return time.Now().Add(skew).After(t.Expiry)
}

// StoreOAuthToken persists an OAuth credential bundle for a provider profile behind the Secret
// Store (JSON-encoded; the material never leaves the store in the clear).
func (m *Manager) StoreOAuthToken(ctx context.Context, provider, profile string, tok OAuthToken) error {
	if tok.AccessToken == "" {
		return authErr("E-AUTH-002", "empty access token")
	}
	data, err := json.Marshal(tok) //nolint:gosec // G117: intentionally serializing the OAuth token to persist it in the secret store
	if err != nil {
		return err
	}
	return m.secrets.Set(ctx, refFor(provider, profile),
		ports.NewSecretValue(data),
		ports.SecretMeta{Kind: "oauth_token", Provider: provider})
}

// LoadOAuthToken retrieves a stored OAuth credential bundle for a provider profile.
func (m *Manager) LoadOAuthToken(ctx context.Context, provider, profile string) (OAuthToken, error) {
	val, err := m.secrets.Get(ctx, refFor(provider, profile))
	if err != nil {
		return OAuthToken{}, authErr("E-AUTH-003", "no stored credential for "+provider)
	}
	var tok OAuthToken
	if err := json.Unmarshal(val.Bytes(), &tok); err != nil {
		return OAuthToken{}, authErr("E-AUTH-003", "stored credential for "+provider+" is not an OAuth token")
	}
	return tok, nil
}

// jwtAccountID extracts the ChatGPT account id from the first OpenAI OAuth JWT that carries it,
// reading the "https://api.openai.com/auth" claim's chatgpt_account_id field. The token is decoded
// without signature verification: it is already trusted, having come straight from the token
// exchange over TLS (this matches the reference Codex login implementations).
func jwtAccountID(tokens ...string) string {
	for _, t := range tokens {
		var claims struct {
			Auth struct {
				ChatGPTAccountID string `json:"chatgpt_account_id"`
			} `json:"https://api.openai.com/auth"`
		}
		if jwtClaims(t, &claims) == nil && claims.Auth.ChatGPTAccountID != "" {
			return claims.Auth.ChatGPTAccountID
		}
	}
	return ""
}

// jwtExpiry reads the exp claim (Unix seconds) from a JWT, or zero if absent/unparsable.
func jwtExpiry(token string) time.Time {
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if jwtClaims(token, &claims) == nil && claims.Exp > 0 {
		return time.Unix(claims.Exp, 0)
	}
	return time.Time{}
}

// jwtClaims decodes the (unverified) payload segment of a JWT into out.
func jwtClaims(token string, out any) error {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return authErr("E-AUTH-006", "malformed JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, out)
}
