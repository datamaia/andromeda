package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/datamaia/andromeda/internal/ports"
)

// Error codes in the E-PROV family (Volume 5). Retryability is set on the PortError.
const (
	CodeAuth         = "E-PROV-002" // authentication/authorization failure (not retryable)
	CodeRateLimit    = "E-PROV-003" // rate limited (retryable)
	CodeServer       = "E-PROV-004" // provider server error 5xx (retryable)
	CodeConnectivity = "E-PROV-005" // network/connectivity failure (retryable)
	CodeBadRequest   = "E-PROV-006" // malformed request or unsupported input (not retryable)
	CodeUnavailable  = "E-PROV-007" // capability unavailable for this provider/model
	CodeMalformed    = "E-PROV-008" // malformed provider response (not retryable)
)

// provErr builds an E-PROV PortError.
func provErr(code, msg, detail string, retryable bool, cause error) *ports.PortError {
	return &ports.PortError{
		Code: code, Category: "provider", Severity: "error",
		Message: msg, Detail: detail, Retryable: retryable, Cause: cause,
	}
}

// Unavailable returns the standard capability-unavailable error for a provider/model.
func Unavailable(capability string) error {
	return provErr(CodeUnavailable, "capability unavailable", capability, false, nil)
}

// Client wraps an *http.Client with a base URL and default headers for one provider endpoint.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Headers map[string]string
}

// PostJSON sends body as JSON to path and decodes a JSON response into out. Non-2xx responses
// and transport failures are mapped into the E-PROV family. For streaming, use Stream.
func (c *Client) PostJSON(ctx context.Context, path string, body any, out any) error {
	resp, err := c.post(ctx, path, body, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if err := mapStatus(resp.StatusCode, data); err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return provErr(CodeMalformed, "could not decode provider response", err.Error(), false, err)
		}
	}
	return nil
}

// PostStream sends body as JSON and returns the raw response body for SSE/stream parsing. The
// caller must close it. Non-2xx responses are mapped to E-PROV before returning.
func (c *Client) PostStream(ctx context.Context, path string, body any) (io.ReadCloser, error) {
	resp, err := c.post(ctx, path, body, true)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, mapStatus(resp.StatusCode, data)
	}
	return resp.Body, nil
}

func (c *Client) post(ctx context.Context, path string, body any, stream bool) (*http.Response, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, provErr(CodeBadRequest, "could not encode request", err.Error(), false, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, provErr(CodeBadRequest, "could not build request", err.Error(), false, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, provErr(CodeConnectivity, "could not reach provider", err.Error(), true, err)
	}
	return resp, nil
}

// GetJSON performs a GET and decodes JSON (used for model discovery).
func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return provErr(CodeBadRequest, "could not build request", err.Error(), false, err)
	}
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	resp, err := hc.Do(req)
	if err != nil {
		return provErr(CodeConnectivity, "could not reach provider", err.Error(), true, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if err := mapStatus(resp.StatusCode, data); err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return provErr(CodeMalformed, "could not decode provider response", err.Error(), false, err)
		}
	}
	return nil
}

func mapStatus(status int, body []byte) error {
	if status >= 200 && status < 300 {
		return nil
	}
	detail := truncate(string(body), 512)
	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return provErr(CodeAuth, "provider authentication failed", detail, false, nil)
	case status == http.StatusTooManyRequests:
		return provErr(CodeRateLimit, "provider rate limited", detail, true, nil)
	case status >= 500:
		return provErr(CodeServer, fmt.Sprintf("provider server error (%d)", status), detail, true, nil)
	case status == http.StatusBadRequest || status == http.StatusNotFound || status == http.StatusUnprocessableEntity:
		return provErr(CodeBadRequest, fmt.Sprintf("provider rejected request (%d)", status), detail, false, nil)
	default:
		return provErr(CodeServer, fmt.Sprintf("unexpected provider status %d", status), detail, status >= 500, nil)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
