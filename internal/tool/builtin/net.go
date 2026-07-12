package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// maxHTTPBody bounds captured response bodies (parity with the other tools' 1 MiB cap).
const maxHTTPBody = 1 << 20

// HTTPRequest performs one HTTP request against an allowed host. It requires the `network`
// permission scoped to the target host; when a credential_ref is supplied it additionally
// requires `credential_access` and resolves the secret server-side into an Authorization bearer
// header — the secret value never appears in the arguments or the produced record. Phase: Beta.
type HTTPRequest struct {
	Client  *http.Client
	Secrets ports.SecretStorePort // optional; required only to resolve credential_ref
}

// NewHTTPRequest builds the http.request tool. A nil client uses http.DefaultClient.
func NewHTTPRequest(client *http.Client, secrets ports.SecretStorePort) HTTPRequest {
	if client == nil {
		client = http.DefaultClient
	}
	return HTTPRequest{Client: client, Secrets: secrets}
}

// Describe returns the http_request tool descriptor.
func (HTTPRequest) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "http_request", Namespace: "http", Version: "1",
		Description: "Perform one HTTP request against an allowed host",
		InputSchema: []byte(`{"type":"object","required":["method","url"],"properties":{` +
			`"method":{"type":"string"},"url":{"type":"string"},"headers":{"type":"object"},` +
			`"body":{"type":"string"},"credential_ref":{"type":"string"},` +
			`"timeout_ms":{"type":"integer"},"max_redirects":{"type":"integer"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"status":{"type":"integer"},"headers":{"type":"object"},"body":{"type":"string"},"truncated":{"type":"boolean"}}}`),
		Permissions:  []core.Permission{core.PermNetwork, core.PermCredentialAccess},
		Origin:       "builtin", TrustLevel: "trusted",
	}, nil
}

type httpInput struct {
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	CredentialRef string            `json:"credential_ref"`
	MaxRedirects  *int              `json:"max_redirects"`
}

// Validate requires method and a parseable url.
func (HTTPRequest) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in httpInput
	if err := json.Unmarshal(input, &in); err != nil || in.Method == "" || in.URL == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"method and url are required"}}, nil
	}
	if _, err := url.ParseRequestURI(in.URL); err != nil {
		return ports.ValidationResult{Valid: false, Findings: []string{"invalid url"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

// Resources requests network access to the target host, plus credential_access when a credential_ref is set.
func (HTTPRequest) Resources(input ports.JSON) ([]ports.PermissionQuery, error) {
	var in httpInput
	_ = json.Unmarshal(input, &in)
	host := ""
	if u, err := url.Parse(in.URL); err == nil {
		host = u.Hostname()
	}
	qs := []ports.PermissionQuery{{Permission: core.PermNetwork, Scope: core.ScopeDomain, Subject: host}}
	if in.CredentialRef != "" {
		qs = append(qs, ports.PermissionQuery{Permission: core.PermCredentialAccess, Scope: core.ScopeDomain, Subject: host})
	}
	return qs, nil
}

// Execute performs the HTTP request and returns its status, headers, and capped body.
func (t HTTPRequest) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in httpInput
	_ = json.Unmarshal(req.Input, &in)

	var bodyReader io.Reader
	if in.Body != "" {
		bodyReader = strings.NewReader(in.Body)
	}
	hreq, err := http.NewRequestWithContext(ctx, strings.ToUpper(in.Method), in.URL, bodyReader)
	if err != nil {
		return errEvent("could not build request: " + err.Error()), nil
	}
	for k, v := range in.Headers {
		hreq.Header.Set(k, v)
	}
	if in.CredentialRef != "" {
		token, err := t.resolveCredential(ctx, in.CredentialRef)
		if err != nil {
			return errEvent("credential resolution failed: " + err.Error()), nil
		}
		hreq.Header.Set("Authorization", "Bearer "+token)
	}

	client := t.clientFor(in)
	resp, err := client.Do(hreq)
	if err != nil {
		return errEvent("request failed: " + err.Error()), nil
	}
	defer func() { _ = resp.Body.Close() }()

	limited := io.LimitReader(resp.Body, maxHTTPBody+1)
	data, _ := io.ReadAll(limited)
	truncated := int64(len(data)) > maxHTTPBody
	if truncated {
		data = data[:maxHTTPBody]
	}

	headers := map[string]string{}
	for k := range resp.Header {
		headers[k] = resp.Header.Get(k)
	}
	out, _ := json.Marshal(map[string]any{
		"status": resp.StatusCode, "headers": headers, "body": string(data), "truncated": truncated,
	})
	return okEvent(string(out)), nil
}

// clientFor derives a client honoring max_redirects when set (0 means do not follow).
func (t HTTPRequest) clientFor(in httpInput) *http.Client {
	base := t.Client
	if base == nil {
		base = http.DefaultClient
	}
	if in.MaxRedirects == nil {
		return base
	}
	limit := *in.MaxRedirects
	c := *base
	c.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		if len(via) > limit {
			return http.ErrUseLastResponse
		}
		return nil
	}
	return &c
}

// resolveCredential reads a secret named by ref ("namespace/name", default namespace "auth") and
// returns its material for immediate use. The value is never returned to the agent or recorded.
func (t HTTPRequest) resolveCredential(ctx context.Context, ref string) (string, error) {
	if t.Secrets == nil {
		return "", fmt.Errorf("no secret store configured")
	}
	ns, name := "auth", ref
	if i := strings.IndexByte(ref, '/'); i >= 0 {
		ns, name = ref[:i], ref[i+1:]
	}
	v, err := t.Secrets.Get(ctx, ports.SecretRef{Namespace: ns, Name: name})
	if err != nil {
		return "", err
	}
	return string(v.Bytes()), nil
}

// Cancel is a no-op; the request is bounded by the Execute context.
func (HTTPRequest) Cancel(context.Context, core.ULID) error { return nil }
