package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BrowserFlowConfig configures an OAuth 2.0 authorization-code flow with PKCE via the system
// browser and a fixed localhost redirect (ADR-063). Only official, documented endpoints are used.
type BrowserFlowConfig struct {
	AuthorizeURL string
	TokenURL     string
	ClientID     string
	RedirectPath string            // e.g. "/auth/callback"
	RedirectPort int               // e.g. 1455
	Scopes       []string          // space-joined into the scope param
	ExtraAuth    map[string]string // additional authorization-URL parameters
	HTTPClient   *http.Client      // nil uses http.DefaultClient
}

func (c BrowserFlowConfig) redirectURI() string {
	return fmt.Sprintf("http://localhost:%d%s", c.RedirectPort, c.RedirectPath)
}

// BuildAuthorizeURL constructs the authorization URL for a PKCE flow.
func BuildAuthorizeURL(cfg BrowserFlowConfig, pkce PKCE, state string) string {
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {cfg.ClientID},
		"redirect_uri":          {cfg.redirectURI()},
		"scope":                 {strings.Join(cfg.Scopes, " ")},
		"code_challenge":        {pkce.Challenge},
		"code_challenge_method": {pkce.Method},
		"state":                 {state},
	}
	for k, v := range cfg.ExtraAuth {
		q.Set(k, v)
	}
	return cfg.AuthorizeURL + "?" + q.Encode()
}

// RunBrowserFlow performs the full browser authorization-code+PKCE flow: it binds the localhost
// callback, opens the browser to the authorization URL, waits for the redirect, validates the
// state, and exchanges the code for tokens. open launches the browser (injected for testability);
// it blocks until the callback fires, the context ends, or the flow fails.
func RunBrowserFlow(ctx context.Context, cfg BrowserFlowConfig, open func(string) error) (OAuthToken, error) {
	pkce, err := NewPKCE()
	if err != nil {
		return OAuthToken{}, err
	}
	state, err := RandomState()
	if err != nil {
		return OAuthToken{}, err
	}

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.RedirectPort))
	if err != nil {
		return OAuthToken{}, authErr("E-AUTH-006",
			fmt.Sprintf("cannot bind localhost:%d for the OAuth callback (is another login in progress?): %v", cfg.RedirectPort, err))
	}
	defer func() { _ = ln.Close() }()

	type result struct {
		code string
		err  error
	}
	resCh := make(chan result, 1)
	mux := http.NewServeMux()
	mux.HandleFunc(cfg.RedirectPath, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case q.Get("error") != "":
			writeCallbackPage(w, false)
			resCh <- result{err: authErr("E-AUTH-006", "authorization denied: "+q.Get("error"))}
		case q.Get("state") != state:
			writeCallbackPage(w, false)
			resCh <- result{err: authErr("E-AUTH-006", "OAuth state mismatch — the login was not completed safely")}
		default:
			writeCallbackPage(w, true)
			resCh <- result{code: q.Get("code")}
		}
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = srv.Serve(ln) }()
	defer func() { _ = srv.Close() }()

	if open != nil {
		_ = open(BuildAuthorizeURL(cfg, pkce, state)) // best-effort; caller also prints the URL
	}

	select {
	case <-ctx.Done():
		return OAuthToken{}, ctx.Err()
	case res := <-resCh:
		if res.err != nil {
			return OAuthToken{}, res.err
		}
		if res.code == "" {
			return OAuthToken{}, authErr("E-AUTH-006", "authorization callback carried no code")
		}
		return exchangeCode(ctx, cfg, res.code, pkce.Verifier)
	}
}

// exchangeCode swaps an authorization code (+ PKCE verifier) for tokens.
func exchangeCode(ctx context.Context, cfg BrowserFlowConfig, code, verifier string) (OAuthToken, error) {
	return postToken(ctx, cfg, url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {cfg.redirectURI()},
		"client_id":     {cfg.ClientID},
		"code_verifier": {verifier},
	})
}

// RefreshBrowserToken exchanges a refresh token for a fresh access token, preserving the refresh
// token if the server does not return a new one.
func RefreshBrowserToken(ctx context.Context, cfg BrowserFlowConfig, refresh string) (OAuthToken, error) {
	tok, err := postToken(ctx, cfg, url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refresh},
		"client_id":     {cfg.ClientID},
	})
	if err != nil {
		return OAuthToken{}, err
	}
	if tok.RefreshToken == "" {
		tok.RefreshToken = refresh
	}
	return tok, nil
}

// tokenResponse is the token endpoint's JSON body.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func postToken(ctx context.Context, cfg BrowserFlowConfig, form url.Values) (OAuthToken, error) {
	hc := cfg.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return OAuthToken{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := hc.Do(req)
	if err != nil {
		return OAuthToken{}, authErr("E-AUTH-006", "token request failed: "+err.Error())
	}
	defer func() { _ = resp.Body.Close() }()
	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return OAuthToken{}, authErr("E-AUTH-006", "token response was not valid JSON")
	}
	if tr.Error != "" {
		msg := tr.Error
		if tr.ErrorDesc != "" {
			msg += ": " + tr.ErrorDesc
		}
		return OAuthToken{}, authErr("E-AUTH-006", "authorization failed: "+msg)
	}
	if tr.AccessToken == "" {
		return OAuthToken{}, authErr("E-AUTH-006", "token endpoint returned no access token")
	}
	tok := OAuthToken{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		IDToken:      tr.IDToken,
		AccountID:    jwtAccountID(tr.IDToken, tr.AccessToken),
	}
	if exp := jwtExpiry(tr.AccessToken); !exp.IsZero() {
		tok.Expiry = exp
	} else if tr.ExpiresIn > 0 {
		tok.Expiry = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}
	return tok, nil
}

// writeCallbackPage renders the minimal page the browser shows after the redirect.
func writeCallbackPage(w http.ResponseWriter, ok bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	msg, sub := "You're signed in.", "You can close this tab and return to your terminal."
	if !ok {
		msg, sub = "Sign-in failed.", "Return to your terminal and try again."
	}
	_, _ = fmt.Fprintf(w, `<!doctype html><meta charset=utf-8><title>Andromeda</title>`+
		`<body style="background:#121417;color:#F5F2ED;font-family:system-ui,sans-serif;`+
		`display:flex;flex-direction:column;align-items:center;justify-content:center;height:100vh;margin:0">`+
		`<h1 style="color:#7C5CFF;font-weight:600">%s</h1><p style="color:#8C7B6E">%s</p></body>`, msg, sub)
}
