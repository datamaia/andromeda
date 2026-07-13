package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// fakeJWT builds an unsigned JWT whose payload carries the given claims (header/signature are
// irrelevant — the code decodes the payload without verifying).
func fakeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	return "h." + base64.RawURLEncoding.EncodeToString(payload) + ".s"
}

func testFlow(tokenURL string) BrowserFlowConfig {
	return BrowserFlowConfig{
		AuthorizeURL: "https://auth.example/oauth/authorize",
		TokenURL:     tokenURL,
		ClientID:     "app_test",
		RedirectPath: "/auth/callback",
		RedirectPort: 1455,
		Scopes:       []string{"openid", "offline_access"},
		ExtraAuth:    map[string]string{"originator": "codex_cli_rs"},
	}
}

func TestBuildAuthorizeURLIncludesAllParams(t *testing.T) {
	u := BuildAuthorizeURL(testFlow(""), PKCE{Challenge: "chal", Method: "S256"}, "st8")
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatal(err)
	}
	q := parsed.Query()
	want := map[string]string{
		"response_type":         "code",
		"client_id":             "app_test",
		"redirect_uri":          "http://localhost:1455/auth/callback",
		"scope":                 "openid offline_access",
		"code_challenge":        "chal",
		"code_challenge_method": "S256",
		"state":                 "st8",
		"originator":            "codex_cli_rs",
	}
	for k, v := range want {
		if q.Get(k) != v {
			t.Errorf("authorize %s = %q, want %q", k, q.Get(k), v)
		}
	}
}

func TestExchangeCodeParsesTokensAndAccountID(t *testing.T) {
	access := fakeJWT(t, map[string]any{
		"https://api.openai.com/auth": map[string]any{"chatgpt_account_id": "acc_123"},
		"exp":                         time.Now().Add(time.Hour).Unix(),
	})
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": access, "refresh_token": "refresh_1", "id_token": "", "expires_in": 3600,
		})
	}))
	defer srv.Close()

	tok, err := exchangeCode(context.Background(), testFlow(srv.URL), "the-code", "the-verifier")
	if err != nil {
		t.Fatal(err)
	}
	// request shape
	for k, v := range map[string]string{
		"grant_type": "authorization_code", "code": "the-code",
		"redirect_uri": "http://localhost:1455/auth/callback", "client_id": "app_test",
		"code_verifier": "the-verifier",
	} {
		if gotForm.Get(k) != v {
			t.Errorf("exchange %s = %q, want %q", k, gotForm.Get(k), v)
		}
	}
	// parsed token
	if tok.AccessToken != access || tok.RefreshToken != "refresh_1" {
		t.Errorf("token = %+v", tok)
	}
	if tok.AccountID != "acc_123" {
		t.Errorf("account id = %q, want acc_123", tok.AccountID)
	}
	if tok.Expiry.IsZero() || tok.Expired(0) {
		t.Errorf("expiry not set/valid: %v", tok.Expiry)
	}
}

func TestRefreshBrowserTokenPreservesRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q", r.PostForm.Get("grant_type"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "new_access", "expires_in": 60})
	}))
	defer srv.Close()

	tok, err := RefreshBrowserToken(context.Background(), testFlow(srv.URL), "old_refresh")
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken != "new_access" {
		t.Errorf("access = %q", tok.AccessToken)
	}
	if tok.RefreshToken != "old_refresh" {
		t.Errorf("refresh should be preserved, got %q", tok.RefreshToken)
	}
}

func TestTokenEndpointErrorSurfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant", "error_description": "bad code"})
	}))
	defer srv.Close()
	_, err := exchangeCode(context.Background(), testFlow(srv.URL), "x", "y")
	if err == nil || !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("error = %v, want invalid_grant", err)
	}
}

func TestStoreLoadOAuthTokenRoundTrip(t *testing.T) {
	m := newManager(t)
	ctx := context.Background()
	in := OAuthToken{AccessToken: "a", RefreshToken: "r", AccountID: "acc_9", Expiry: time.Now().Add(time.Hour).Round(time.Second)}
	if err := m.StoreOAuthToken(ctx, "openai-chatgpt", "default", in); err != nil {
		t.Fatal(err)
	}
	out, err := m.LoadOAuthToken(ctx, "openai-chatgpt", "default")
	if err != nil {
		t.Fatal(err)
	}
	if out.AccessToken != in.AccessToken || out.RefreshToken != in.RefreshToken || out.AccountID != in.AccountID {
		t.Errorf("round-trip = %+v, want %+v", out, in)
	}
	if !out.Expiry.Equal(in.Expiry) {
		t.Errorf("expiry round-trip = %v, want %v", out.Expiry, in.Expiry)
	}
}
