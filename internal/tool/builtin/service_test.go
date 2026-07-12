package builtin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

func TestServiceRequestTransport(t *testing.T) {
	var gotAuth, gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath, gotMethod = r.Header.Get("Authorization"), r.URL.Path, r.Method
		w.Header().Set("X-RateLimit-Remaining", "42")
		w.Header().Set("X-RateLimit-Reset", "1700000000")
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	secrets := memSecrets{}
	_ = secrets.Set(context.Background(), ports.SecretRef{Namespace: "auth", Name: "github:default"}, ports.NewSecretValue([]byte("GH-TOKEN")), ports.SecretMeta{})
	endpoint := ServiceEndpoint{BaseURL: srv.URL, CredentialRef: "auth/github:default", AuthScheme: "token"}
	tool := NewGitHubRequest(endpoint, secrets, srv.Client())

	d, _ := tool.Describe(context.Background())
	if d.Name != "github_request" {
		t.Fatalf("name = %s", d.Name)
	}
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"operation": "get_repo", "method": "get", "path": "/repos/x/y"}))
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	if gotAuth != "token GH-TOKEN" || gotPath != "/repos/x/y" || gotMethod != "GET" {
		t.Fatalf("auth=%q path=%q method=%q", gotAuth, gotPath, gotMethod)
	}
	var res struct {
		Status    int            `json:"status"`
		Result    string         `json:"result"`
		RateLimit map[string]any `json:"rate_limit"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Status != 200 || res.Result != `{"ok":true}` {
		t.Fatalf("result = %+v", res)
	}
	if rem, _ := res.RateLimit["remaining"].(float64); rem != 42 {
		t.Fatalf("rate_limit = %+v", res.RateLimit)
	}
	// The secret value must never appear in the record.
	if strings.Contains(text, "GH-TOKEN") {
		t.Fatal("token leaked into the record")
	}
}

func TestServiceRequestUnconfiguredRefused(t *testing.T) {
	tool := NewSlackRequest(ServiceEndpoint{}, nil, nil) // no base URL
	vr, _ := tool.Validate(context.Background(), []byte(`{"operation":"post_message","path":"/chat.postMessage"}`))
	if vr.Valid {
		t.Fatal("an unconfigured service must be invalid")
	}
}

func TestServiceRequestBodyAndSlackPerms(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	tool := NewSlackRequest(ServiceEndpoint{BaseURL: srv.URL}, nil, srv.Client())

	// Slack declares the notifications permission in its surface.
	d, _ := tool.Describe(context.Background())
	var notif bool
	for _, p := range d.Permissions {
		if p == "notifications" {
			notif = true
		}
	}
	if !notif {
		t.Fatal("slack_request must declare notifications")
	}
	outcome, _ := runTool(t, tool, mustJSON(map[string]any{"operation": "post_message", "method": "post", "path": "/chat.postMessage", "body": `{"channel":"C1","text":"hi"}`}))
	if outcome != "success" {
		t.Fatal("post should succeed")
	}
	if gotBody != `{"channel":"C1","text":"hi"}` {
		t.Fatalf("server saw body = %q", gotBody)
	}
}
