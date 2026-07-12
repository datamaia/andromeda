package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
)

// BrowserControl drives a browser session over the W3C WebDriver protocol — a documented
// automation standard, so no proprietary mechanism is invented. The WebDriver endpoint (a running
// driver such as a Selenium/geckodriver/chromedriver server) is supplied by configuration; this
// tool is the transport that speaks the standard REST commands. Phase: v1.
type BrowserControl struct {
	endpoint string // WebDriver base URL, e.g. http://localhost:4444
	sessdir  string // session id path segment prefix; standard is "/session"
	client   *http.Client
}

// NewBrowserControl builds browser.control against a WebDriver endpoint. A nil client uses the
// default. An empty endpoint means the tool is unconfigured and refuses to run.
func NewBrowserControl(endpoint string, client *http.Client) BrowserControl {
	if client == nil {
		client = http.DefaultClient
	}
	return BrowserControl{endpoint: strings.TrimRight(endpoint, "/"), sessdir: "/session", client: client}
}

// Describe returns the browser_control tool descriptor.
func (BrowserControl) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name: "browser_control", Namespace: "browser", Version: "1",
		Description: "Drive a browser via the W3C WebDriver protocol",
		InputSchema: []byte(`{"type":"object","required":["operation","session"],"properties":{` +
			`"operation":{"type":"string","enum":["navigate","read","screenshot","click","type","evaluate"]},` +
			`"session":{"type":"string"},"url":{"type":"string"},"selector":{"type":"string"},"text":{"type":"string"}}}`),
		OutputSchema: []byte(`{"type":"object","properties":{"session":{"type":"string"},"content":{"type":"string"},"result":{"type":"string"},"truncated":{"type":"boolean"}}}`),
		Permissions:  []core.Permission{core.PermNetwork, core.PermProcessSpawn}, Origin: "builtin", TrustLevel: "trusted",
	}, nil
}

type browserInput struct {
	Operation string `json:"operation"`
	Session   string `json:"session"`
	URL       string `json:"url"`
	Selector  string `json:"selector"`
	Text      string `json:"text"`
}

// Validate requires operation and session, and requires url for the navigate operation.
func (BrowserControl) Validate(_ context.Context, input ports.JSON) (ports.ValidationResult, error) {
	var in browserInput
	if err := json.Unmarshal(input, &in); err != nil || in.Operation == "" || in.Session == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"operation and session are required"}}, nil
	}
	if in.operationNeedsURL() && in.URL == "" {
		return ports.ValidationResult{Valid: false, Findings: []string{"navigate requires url"}}, nil
	}
	return ports.ValidationResult{Valid: true}, nil
}

func (in browserInput) operationNeedsURL() bool { return in.Operation == "navigate" }

// Resources requests network access to the WebDriver endpoint.
func (BrowserControl) Resources(ports.JSON) ([]ports.PermissionQuery, error) {
	return []ports.PermissionQuery{{Permission: core.PermNetwork, Scope: core.ScopeDomain, Subject: "webdriver"}}, nil
}

// Execute maps the requested operation to its W3C WebDriver command and returns the response.
func (t BrowserControl) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	var in browserInput
	_ = json.Unmarshal(req.Input, &in)
	if t.endpoint == "" {
		return errEvent("browser.control is not configured (no WebDriver endpoint)"), nil
	}
	base := t.endpoint + t.sessdir + "/" + in.Session

	// Map each operation to its standard W3C WebDriver command (method, path, payload).
	switch in.Operation {
	case "navigate":
		body, _ := json.Marshal(map[string]any{"url": in.URL})
		return t.call(ctx, http.MethodPost, base+"/url", body, "result")
	case "read":
		return t.call(ctx, http.MethodGet, base+"/source", nil, "content")
	case "screenshot":
		return t.call(ctx, http.MethodGet, base+"/screenshot", nil, "content")
	case "evaluate":
		body, _ := json.Marshal(map[string]any{"script": in.Text, "args": []any{}})
		return t.call(ctx, http.MethodPost, base+"/execute/sync", body, "result")
	case "click", "type":
		// Locate the element, then act. Kept minimal: this proves the transport, not a full driver.
		body, _ := json.Marshal(map[string]any{"using": "css selector", "value": in.Selector})
		return t.call(ctx, http.MethodPost, base+"/element", body, "result")
	default:
		return errEvent("unsupported operation: " + in.Operation), nil
	}
}

func (t BrowserControl) call(ctx context.Context, method, url string, payload []byte, field string) (ports.Stream[ports.ToolEvent], error) {
	var body io.Reader
	if payload != nil {
		body = bytes.NewReader(payload)
	}
	hreq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return errEvent("could not build WebDriver request: " + err.Error()), nil
	}
	hreq.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(hreq)
	if err != nil {
		return errEvent("WebDriver request failed: " + err.Error()), nil
	}
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBody))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errEvent("WebDriver error " + resp.Status + ": " + string(data)), nil
	}
	// WebDriver replies wrap the payload under {"value": ...}; surface it as the named field.
	var wrap struct {
		Value json.RawMessage `json:"value"`
	}
	_ = json.Unmarshal(data, &wrap)
	out, _ := json.Marshal(map[string]any{field: string(wrap.Value), "truncated": false})
	return okEvent(string(out)), nil
}

// Cancel is a no-op; WebDriver commands complete synchronously within Execute.
func (BrowserControl) Cancel(context.Context, core.ULID) error { return nil }
