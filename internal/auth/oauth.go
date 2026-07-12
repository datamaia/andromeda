package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuthConfig describes a provider's official OAuth endpoints for the device authorization
// grant (RFC 8628, ADR-063). Only official, documented endpoints are used.
type OAuthConfig struct {
	ClientID      string
	DeviceAuthURL string
	TokenURL      string
	Scopes        []string
	HTTPClient    *http.Client
	// PollInterval overrides the server-provided polling interval (used for fast tests).
	PollInterval time.Duration
}

// DeviceCodeResponse is the device authorization endpoint's response.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
	ExpiresIn       int    `json:"expires_in"`
}

// StartDeviceFlow requests a device and user code. The caller displays UserCode and
// VerificationURI to the user, then calls PollDeviceToken.
func StartDeviceFlow(ctx context.Context, cfg OAuthConfig) (DeviceCodeResponse, error) {
	form := url.Values{"client_id": {cfg.ClientID}}
	if len(cfg.Scopes) > 0 {
		form.Set("scope", strings.Join(cfg.Scopes, " "))
	}
	var dc DeviceCodeResponse
	if err := postForm(ctx, cfg, cfg.DeviceAuthURL, form, &dc); err != nil {
		return DeviceCodeResponse{}, authErr("E-AUTH-005", "device authorization request failed: "+err.Error())
	}
	if dc.DeviceCode == "" {
		return DeviceCodeResponse{}, authErr("E-AUTH-005", "device authorization returned no device code")
	}
	return dc, nil
}

// PollDeviceToken polls the token endpoint until the user authorizes, then returns the access
// token. It honors the server's interval, slow-down responses, and the context deadline.
func PollDeviceToken(ctx context.Context, cfg OAuthConfig, dc DeviceCodeResponse) (string, error) {
	interval := time.Duration(dc.Interval) * time.Second
	if cfg.PollInterval > 0 {
		interval = cfg.PollInterval
	}
	if interval <= 0 {
		interval = time.Second
	}
	for {
		form := url.Values{
			"client_id":   {cfg.ClientID},
			"device_code": {dc.DeviceCode},
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		}
		var tok struct {
			AccessToken string `json:"access_token"`
			Error       string `json:"error"`
		}
		if err := postForm(ctx, cfg, cfg.TokenURL, form, &tok); err != nil {
			return "", authErr("E-AUTH-006", "token request failed: "+err.Error())
		}
		switch {
		case tok.AccessToken != "":
			return tok.AccessToken, nil
		case tok.Error == "authorization_pending":
			// keep polling
		case tok.Error == "slow_down":
			interval += time.Second
		case tok.Error == "":
			return "", authErr("E-AUTH-006", "token endpoint returned neither a token nor an error")
		default:
			return "", authErr("E-AUTH-006", "authorization failed: "+tok.Error)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}
}

func postForm(ctx context.Context, cfg OAuthConfig, endpoint string, form url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	hc := cfg.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}

// AuthenticateDevice runs the full device authorization grant for a provider profile, stores the
// resulting access token behind a Secret Store reference, and returns the session handle. The
// display callback surfaces the user code and verification URL to the user.
func (m *Manager) AuthenticateDevice(ctx context.Context, provider, profile string, cfg OAuthConfig, display func(DeviceCodeResponse)) error {
	dc, err := StartDeviceFlow(ctx, cfg)
	if err != nil {
		return err
	}
	if display != nil {
		display(dc)
	}
	token, err := PollDeviceToken(ctx, cfg, dc)
	if err != nil {
		return err
	}
	return m.StoreAPIKey(ctx, provider, profile, token)
}
