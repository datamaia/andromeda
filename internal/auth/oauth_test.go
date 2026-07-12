package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/datamaia/andromeda/internal/ports"
)

// fakeOAuthServer implements the device grant: device auth returns a code; the token endpoint
// returns authorization_pending once, then the access token.
func fakeOAuthServer(t *testing.T) (*httptest.Server, OAuthConfig) {
	t.Helper()
	var polls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/device", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"device_code":"DEV123","user_code":"WXYZ-1234","verification_uri":"https://example/device","interval":1,"expires_in":300}`))
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if atomic.AddInt32(&polls, 1) == 1 {
			w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		w.Write([]byte(`{"access_token":"ACCESS-TOKEN-OK"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	cfg := OAuthConfig{
		ClientID:      "cli",
		DeviceAuthURL: srv.URL + "/device",
		TokenURL:      srv.URL + "/token",
		HTTPClient:    srv.Client(),
		PollInterval:  5 * time.Millisecond,
	}
	return srv, cfg
}

func TestDeviceFlowStartAndPoll(t *testing.T) {
	ctx := context.Background()
	_, cfg := fakeOAuthServer(t)
	dc, err := StartDeviceFlow(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if dc.UserCode != "WXYZ-1234" {
		t.Fatalf("device code response = %+v", dc)
	}
	token, err := PollDeviceToken(ctx, cfg, dc)
	if err != nil {
		t.Fatal(err)
	}
	if token != "ACCESS-TOKEN-OK" {
		t.Errorf("token = %q", token)
	}
}

func TestAuthenticateDeviceStoresToken(t *testing.T) {
	ctx := context.Background()
	_, cfg := fakeOAuthServer(t)
	m := newManager(t)
	var shown DeviceCodeResponse
	if err := m.AuthenticateDevice(ctx, "acme", "default", cfg, func(dc DeviceCodeResponse) { shown = dc }); err != nil {
		t.Fatal(err)
	}
	if shown.UserCode == "" {
		t.Error("display callback did not receive the user code")
	}
	// The token is now stored and Authenticate(api_key) confirms presence.
	if _, err := m.Authenticate(ctx, ports.AuthSpec{Provider: "acme", Mechanism: "api_key"}); err != nil {
		t.Fatalf("stored token should authenticate: %v", err)
	}
}

func TestPollDeviceTokenContextCancel(t *testing.T) {
	_, cfg := fakeOAuthServer(t)
	cfg.TokenURL = cfg.DeviceAuthURL // always "pending"-less; force cancellation path
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := PollDeviceToken(ctx, cfg, DeviceCodeResponse{DeviceCode: "x", Interval: 1})
	if err == nil {
		t.Fatal("expected a cancellation/error")
	}
}
