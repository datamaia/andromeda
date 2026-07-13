package main

import (
	"context"
	"os"
	"time"

	"github.com/datamaia/andromeda/internal/app"
	"github.com/datamaia/andromeda/internal/auth"
	"github.com/datamaia/andromeda/internal/tui"
)

// startProviderAuth implements tui.ProviderAuthFunc: for the ChatGPT OAuth provider it runs the
// browser sign-in on a background goroutine and streams progress (URL, then done/error); other
// providers — and an already signed-in ChatGPT — return nil, meaning no interactive auth is needed.
func (s *tuiSession) startProviderAuth(id string) <-chan tui.AuthEvent {
	if id != auth.OpenAIChatGPTProvider {
		return nil
	}
	if s.chatGPTSignedIn() {
		return nil
	}
	ch := make(chan tui.AuthEvent, 2)
	go func() {
		defer close(ch)
		ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
		defer cancel()
		tok, err := auth.RunBrowserFlow(ctx, auth.OpenAIChatGPTFlow(), func(u string) error {
			ch <- tui.AuthEvent{URL: u}
			return openBrowser(u)
		})
		if err != nil {
			ch <- tui.AuthEvent{Err: err}
			return
		}
		mgr, err := newAuthManager()
		if err != nil {
			ch <- tui.AuthEvent{Err: err}
			return
		}
		if err := mgr.StoreOAuthToken(ctx, auth.OpenAIChatGPTProvider, "default", tok); err != nil {
			ch <- tui.AuthEvent{Err: err}
			return
		}
		ch <- tui.AuthEvent{Done: true}
	}()
	return ch
}

// chatGPTSignedIn reports whether a usable ChatGPT token is stored (still valid, or refreshable).
func (s *tuiSession) chatGPTSignedIn() bool {
	ss, err := app.SecretStore()
	if err != nil {
		return false
	}
	tok, err := auth.New(ss).LoadOAuthToken(s.ctx, auth.OpenAIChatGPTProvider, "default")
	if err != nil {
		return false
	}
	return tok.AccessToken != "" && (!tok.Expired(0) || tok.RefreshToken != "")
}

// providerKeyEnvFor implements tui.ProviderKeyEnvFunc: the environment variable a provider still
// needs its API key in, or "" when no key is required or one is already present.
func providerKeyEnvFor(id string) string {
	info, ok := app.LookupProvider(id)
	if !ok || !info.KeyRequired || info.KeyEnv == "" {
		return ""
	}
	if os.Getenv(info.KeyEnv) != "" {
		return ""
	}
	return info.KeyEnv
}

// setProviderKey records a pasted API key into the process environment for this session so the
// provider can be built. It is never written to disk.
func setProviderKey(id, key string) {
	if info, ok := app.LookupProvider(id); ok && info.KeyEnv != "" {
		_ = os.Setenv(info.KeyEnv, key)
	}
}
