package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// ServiceEndpoint carries the validated, operator-supplied facts a service integration needs.
// Per ADR-074 the per-service base URL, auth mode, and scopes are PENDING VALIDATION, so they are
// NEVER hardcoded here — the composition resolves them from configuration ([services.<name>]) and
// the Secret Store. An empty BaseURL means the service is not configured.
type ServiceEndpoint struct {
	BaseURL       string
	CredentialRef string // "namespace/name" in the Secret Store; empty for unauthenticated calls
	AuthScheme    string // "bearer" (default), "token", or "raw"
	AuthHeader    string // default "Authorization"
}

// ServiceRequest is the transport-and-schema surface for an external service's official REST or
// GraphQL API (github/gitlab/jira/notion/linear/slack). It injects credentials from the Secret
// Store, composes the request against the configured base URL, executes it, and captures the
// response plus any rate-limit headers. The concrete operation→path mapping is a per-service fact
// that is PENDING VALIDATION, so the caller supplies method and path; `operation` is a label
// recorded on the result. Secret material never appears in the arguments or the record.
type ServiceRequest struct {
	service  string
	perms    []core.Permission
	client   *http.Client
	secrets  ports.SecretStorePort
	endpoint ServiceEndpoint
}

// NewServiceRequest builds a service-request tool. Callers use the per-service constructors below.
func NewServiceRequest(service string, perms []core.Permission, endpoint ServiceEndpoint, secrets ports.SecretStorePort, client *http.Client) ServiceRequest {
	if client == nil {
		client = http.DefaultClient
	}
	return ServiceRequest{service: service, perms: perms, client: client, secrets: secrets, endpoint: endpoint}
}

// Per-service constructors fix the tool name and permission surface from the catalog.
func NewGitHubRequest(e ServiceEndpoint, s ports.SecretStorePort, c *http.Client) ServiceRequest {
	return NewServiceRequest("github", []core.Permission{core.PermExternalServiceAccess, core.PermNetwork}, e, s, c)
}
func NewGitLabRequest(e ServiceEndpoint, s ports.SecretStorePort, c *http.Client) ServiceRequest {
	return NewServiceRequest("gitlab", []core.Permission{core.PermExternalServiceAccess, core.PermNetwork}, e, s, c)
}
func NewJiraRequest(e ServiceEndpoint, s ports.SecretStorePort, c *http.Client) ServiceRequest {
	return NewServiceRequest("jira", []core.Permission{core.PermExternalServiceAccess, core.PermNetwork}, e, s, c)
}
func NewNotionRequest(e ServiceEndpoint, s ports.SecretStorePort, c *http.Client) ServiceRequest {
	return NewServiceRequest("notion", []core.Permission{core.PermExternalServiceAccess, core.PermNetwork}, e, s, c)
}
func NewLinearRequest(e ServiceEndpoint, s ports.SecretStorePort, c *http.Client) ServiceRequest {
	return NewServiceRequest("linear", []core.Permission{core.PermExternalServiceAccess, core.PermNetwork}, e, s, c)
}
func NewSlackRequest(e ServiceEndpoint, s ports.SecretStorePort, c *http.Client) ServiceRequest {
	return NewServiceRequest("slack", []core.Permission{core.PermExternalServiceAccess, core.PermNetwork, core.PermNotifications}, e, s, c)
}

func (t ServiceRequest) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: t.service + "_request", Namespace: t.service, Version: "1",
		Description: "Transport surface for the " + t.service + " official API",
		InputSchema: []byte(`{"type":"object","required":["operation","path"],"properties":{` +
			`"operation":{"type":"string"},"method":{"type":"string"},"path":{"type":"string"},` +
			`"body":{"type":"string"},"params":{"type":"object"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"operation":{"type":"string"},"status":{"type":"integer"},"result":{"type":"string"},"rate_limit":{"type":"object"},"truncated":{"type":"boolean"}}}`),
		Permissions:  t.perms, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type serviceInput struct {
	Operation string `json:"operation"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Body      string `json:"body"`
}

func (t ServiceRequest) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in serviceInput
	if err := json.Unmarshal(input, &in); err != nil || in.Operation == "" || in.Path == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"operation and path are required"}}, nil
	}
	if t.endpoint.BaseURL == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{t.service + " is not configured ([services." + t.service + "].base_url)"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

func (t ServiceRequest) Resources(ports.JSON) ([]ports.PermissionQuery, error) {
	qs := make([]ports.PermissionQuery, 0, len(t.perms))
	for _, p := range t.perms {
		qs = append(qs, ports.PermissionQuery{Permission: p, Scope: core.ScopeProvider, Subject: t.service})
	}
	return qs, nil
}

func (t ServiceRequest) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in serviceInput
	_ = json.Unmarshal(req.Input, &in)
	if t.endpoint.BaseURL == "" {
		return errEvent(t.service + " is not configured"), nil
	}
	method := strings.ToUpper(in.Method)
	if method == "" {
		method = http.MethodGet
	}
	url := strings.TrimRight(t.endpoint.BaseURL, "/") + "/" + strings.TrimLeft(in.Path, "/")

	var body io.Reader
	if in.Body != "" {
		body = strings.NewReader(in.Body)
	}
	hreq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return errEvent("could not build request: " + err.Error()), nil
	}
	hreq.Header.Set("Accept", "application/json")
	if in.Body != "" {
		hreq.Header.Set("Content-Type", "application/json")
	}
	if t.endpoint.CredentialRef != "" {
		if err := t.injectAuth(ctx, hreq); err != nil {
			return errEvent("credential resolution failed: " + err.Error()), nil
		}
	}

	resp, err := t.client.Do(hreq)
	if err != nil {
		return errEvent("request failed: " + err.Error()), nil
	}
	defer resp.Body.Close()
	limited := io.LimitReader(resp.Body, maxHTTPBody+1)
	data, _ := io.ReadAll(limited)
	truncated := int64(len(data)) > maxHTTPBody
	if truncated {
		data = data[:maxHTTPBody]
	}

	out, _ := json.Marshal(map[string]any{
		"operation": in.Operation, "status": resp.StatusCode, "result": string(data),
		"rate_limit": rateLimit(resp.Header), "truncated": truncated,
	})
	return okEvent(string(out)), nil
}

func (t ServiceRequest) injectAuth(ctx context.Context, hreq *http.Request) error {
	if t.secrets == nil {
		return fmt.Errorf("no secret store configured")
	}
	ns, name := "auth", t.endpoint.CredentialRef
	if i := strings.IndexByte(t.endpoint.CredentialRef, '/'); i >= 0 {
		ns, name = t.endpoint.CredentialRef[:i], t.endpoint.CredentialRef[i+1:]
	}
	v, err := t.secrets.Get(ctx, ports.SecretRef{Namespace: ns, Name: name})
	if err != nil {
		return err
	}
	token := string(v.Bytes())
	header := t.endpoint.AuthHeader
	if header == "" {
		header = "Authorization"
	}
	switch t.endpoint.AuthScheme {
	case "token":
		hreq.Header.Set(header, "token "+token)
	case "raw":
		hreq.Header.Set(header, token)
	default: // "bearer"
		hreq.Header.Set(header, "Bearer "+token)
	}
	return nil
}

func (ServiceRequest) Cancel(context.Context, core.ULID) error { return nil }

// rateLimit surfaces common rate-limit headers when the service reports them.
func rateLimit(h http.Header) map[string]any {
	rl := map[string]any{}
	if v := h.Get("X-RateLimit-Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl["remaining"] = n
		}
	}
	if v := h.Get("X-RateLimit-Reset"); v != "" {
		rl["reset"] = v
	}
	return rl
}
