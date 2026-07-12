package mcp

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

// memSecretStore is a minimal in-memory ports.SecretStorePort for the binding test.
type memSecretStore map[string]ports.SecretValue

func (m memSecretStore) Get(_ context.Context, ref ports.SecretRef) (ports.SecretValue, error) {
	v, ok := m[ref.Namespace+"/"+ref.Name]
	if !ok {
		return ports.SecretValue{}, &ports.PortError{Code: "E-SEC-001", Category: "secret", Message: "not found"}
	}
	return v, nil
}
func (m memSecretStore) Set(_ context.Context, ref ports.SecretRef, v ports.SecretValue, _ ports.SecretMeta) error {
	m[ref.Namespace+"/"+ref.Name] = v
	return nil
}
func (m memSecretStore) Delete(_ context.Context, ref ports.SecretRef) error {
	delete(m, ref.Namespace+"/"+ref.Name)
	return nil
}
func (m memSecretStore) List(context.Context, ports.SecretScope) ([]ports.SecretRef, error) {
	return nil, nil
}

// httpMCPServer answers MCP JSON-RPC requests over HTTP (JSON-response mode) and records the
// Authorization header it saw, so the test can assert the bearer token was attached.
func httpMCPServer(t *testing.T, sawAuth *string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*sawAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		var req struct {
			ID     int64  `json:"id"`
			Method string `json:"method"`
		}
		_ = json.Unmarshal(body, &req)
		var result string
		switch req.Method {
		case "initialize":
			result = `{"protocolVersion":"2025-06-18","serverInfo":{"name":"remote","version":"1"}}`
		case "tools/list":
			result = `{"tools":[{"name":"echo","description":"echoes","inputSchema":{"type":"object"}}]}`
		case "tools/call":
			result = `{"content":[{"type":"text","text":"pong"}],"isError":false}`
		default:
			result = `{}`
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc":"2.0","id":` + itoa(req.ID) + `,"result":` + result + `}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func TestHTTPClientBearerAndRoundTrip(t *testing.T) {
	ctx := context.Background()
	var sawAuth string
	srv := httpMCPServer(t, &sawAuth)

	tokenSource := func(context.Context) (string, error) { return "ACCESS-TOKEN-OK", nil }
	c := NewHTTPClient(srv.URL, tokenSource, srv.Client())
	defer c.Close()

	info, err := c.Initialize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.ServerInfo.Name != "remote" {
		t.Fatalf("serverInfo = %+v", info)
	}
	if sawAuth != "Bearer ACCESS-TOKEN-OK" {
		t.Fatalf("Authorization header = %q, want bearer token", sawAuth)
	}

	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name != "echo" {
		t.Fatalf("tools = %+v", tools)
	}

	res, err := c.CallTool(ctx, "echo", json.RawMessage(`{"x":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if res.Text() != "pong" {
		t.Fatalf("call result = %q", res.Text())
	}
}

func TestHTTPClientBearerFromSecretStore(t *testing.T) {
	ctx := context.Background()
	var sawAuth string
	srv := httpMCPServer(t, &sawAuth)

	// Simulate the device grant having stored the access token where auth.Manager puts it.
	store := memSecretStore{}
	ref := ports.SecretRef{Namespace: "auth", Name: "acme:default"}
	_ = store.Set(ctx, ref, ports.NewSecretValue([]byte("DEVICE-GRANT-TOKEN")), ports.SecretMeta{})

	c := NewHTTPClient(srv.URL, BearerFromSecretStore(store, ref), srv.Client())
	defer c.Close()
	if _, err := c.Initialize(ctx); err != nil {
		t.Fatal(err)
	}
	if sawAuth != "Bearer DEVICE-GRANT-TOKEN" {
		t.Fatalf("bearer from secret store = %q", sawAuth)
	}
}

func TestHTTPClientAnonymousWhenNoToken(t *testing.T) {
	ctx := context.Background()
	var sawAuth string
	srv := httpMCPServer(t, &sawAuth)

	c := NewHTTPClient(srv.URL, nil, srv.Client()) // no token source
	defer c.Close()
	if _, err := c.Initialize(ctx); err != nil {
		t.Fatal(err)
	}
	if sawAuth != "" {
		t.Fatalf("expected no Authorization header, got %q", sawAuth)
	}
}

func TestHTTPClientErrorStatusFailsSingleCall(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	c := NewHTTPClient(srv.URL, func(context.Context) (string, error) { return "bad", nil }, srv.Client())
	defer c.Close()
	_, err := c.Initialize(ctx)
	if err == nil {
		t.Fatal("expected an error from a 401 response")
	}
	if !strings.Contains(err.Error(), "http 401") {
		t.Fatalf("error should surface the HTTP status, got %v", err)
	}
}

func TestHTTPClientSSEResponse(t *testing.T) {
	ctx := context.Background()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req struct {
			ID int64 `json:"id"`
		}
		_ = json.Unmarshal(body, &req)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("event: message\ndata: {\"jsonrpc\":\"2.0\",\"id\":" + itoa(req.ID) +
			",\"result\":{\"protocolVersion\":\"2025-06-18\",\"serverInfo\":{\"name\":\"sse\",\"version\":\"1\"}}}\n\n"))
	}))
	t.Cleanup(srv.Close)

	c := NewHTTPClient(srv.URL, nil, srv.Client())
	defer c.Close()
	info, err := c.Initialize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.ServerInfo.Name != "sse" {
		t.Fatalf("serverInfo over SSE = %+v", info)
	}
}
