package builtin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// memSecrets is a minimal in-memory SecretStorePort for credential resolution tests.
type memSecrets map[string][]byte

func (m memSecrets) Get(_ context.Context, ref ports.SecretRef) (ports.SecretValue, error) {
	b, ok := m[ref.Namespace+"/"+ref.Name]
	if !ok {
		return ports.SecretValue{}, &ports.PortError{Code: "E-SEC-001", Category: "secret", Message: "not found"}
	}
	return ports.NewSecretValue(b), nil
}
func (m memSecrets) Set(_ context.Context, ref ports.SecretRef, v ports.SecretValue, _ ports.SecretMeta) error {
	m[ref.Namespace+"/"+ref.Name] = v.Bytes()
	return nil
}
func (m memSecrets) Delete(_ context.Context, ref ports.SecretRef) error {
	delete(m, ref.Namespace+"/"+ref.Name)
	return nil
}
func (memSecrets) List(context.Context, ports.SecretScope) ([]ports.SecretRef, error) {
	return nil, nil
}

func TestHTTPRequestBasic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Echo", r.Method)
		w.WriteHeader(201)
		w.Write([]byte("pong"))
	}))
	defer srv.Close()

	tool := NewHTTPRequest(srv.Client(), nil)
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"method": "post", "url": srv.URL, "body": "ping"}))
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	var res struct {
		Status  int               `json:"status"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Status != 201 || res.Body != "pong" || res.Headers["X-Echo"] != "POST" {
		t.Fatalf("result = %+v", res)
	}
}

func TestHTTPRequestCredentialRefResolvedNotLeaked(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	secrets := memSecrets{}
	_ = secrets.Set(context.Background(), ports.SecretRef{Namespace: "auth", Name: "acme:default"}, ports.NewSecretValue([]byte("SECRET-TOKEN")), ports.SecretMeta{})
	tool := NewHTTPRequest(srv.Client(), secrets)

	outcome, text := runTool(t, tool, mustJSON(map[string]any{"method": "get", "url": srv.URL, "credential_ref": "auth/acme:default"}))
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	if sawAuth != "Bearer SECRET-TOKEN" {
		t.Fatalf("server saw auth = %q", sawAuth)
	}
	// The secret value must not appear anywhere in the produced record.
	if strings.Contains(text, "SECRET-TOKEN") {
		t.Fatal("secret material leaked into the tool record")
	}
}

func TestHTTPRequestRedirectCap(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			w.Write([]byte("done"))
			return
		}
		http.Redirect(w, r, srv.URL+"/next", http.StatusFound) // infinite unless capped
	}))
	defer srv.Close()

	tool := NewHTTPRequest(srv.Client(), nil)
	// max_redirects=0 means do not follow: we should get the 302 itself, not an error loop.
	outcome, text := runTool(t, tool, mustJSON(map[string]any{"method": "get", "url": srv.URL, "max_redirects": 0}))
	if outcome != "success" {
		t.Fatalf("outcome = %s (%s)", outcome, text)
	}
	var res struct {
		Status int `json:"status"`
	}
	_ = json.Unmarshal([]byte(text), &res)
	if res.Status != http.StatusFound {
		t.Fatalf("status with no redirects = %d, want 302", res.Status)
	}
}

func TestHTTPRequestResourcesScopeHost(t *testing.T) {
	qs, err := NewHTTPRequest(nil, nil).Resources([]byte(`{"method":"get","url":"https://api.example.com/v1","credential_ref":"auth/x"}`))
	if err != nil {
		t.Fatal(err)
	}
	var net, cred bool
	for _, q := range qs {
		if q.Permission == "network" && q.Subject == "api.example.com" {
			net = true
		}
		if q.Permission == "credential_access" {
			cred = true
		}
	}
	if !net || !cred {
		t.Fatalf("resources = %+v", qs)
	}
}
