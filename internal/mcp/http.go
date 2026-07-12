package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/datamaia/andromeda/internal/ports"
)

// TokenSource yields the bearer token to present on each HTTP request. It is the binding point
// between MCP and authentication: wire it to the OAuth device grant (auth.Manager) so remote MCP
// servers are reached with a token minted by RFC 8628 device authorization and held in the Secret
// Store. Returning an empty token (and nil error) sends the request without an Authorization
// header (anonymous servers).
type TokenSource func(ctx context.Context) (string, error)

// BearerFromSecretStore is the concrete OAuth→MCP binding: it reads the access token that the
// device authorization grant stored (auth.Manager stores under namespace "auth", name
// "<provider>:<profile>") straight from the Secret Store on each request, so token rotation is
// picked up without rebuilding the client. The material is used immediately and never retained.
func BearerFromSecretStore(store ports.SecretStorePort, ref ports.SecretRef) TokenSource {
	return func(ctx context.Context) (string, error) {
		v, err := store.Get(ctx, ref)
		if err != nil {
			return "", err
		}
		return string(v.Bytes()), nil
	}
}

// NewHTTPClient returns an MCP client that speaks JSON-RPC over HTTP to endpoint, attaching a
// bearer token from tokenSource on every request. This is the JSON-response mode of the MCP
// Streamable HTTP transport (ADR-010); server-initiated SSE push beyond a single response frame
// is PENDING VALIDATION. A nil httpClient uses http.DefaultClient.
func NewHTTPClient(endpoint string, tokenSource TokenSource, httpClient *http.Client) *Client {
	return NewClient(newHTTPTransport(endpoint, tokenSource, httpClient))
}

// httpTransport adapts the newline-delimited JSON-RPC framing (jsonrpc.Conn) to HTTP: each
// written request frame becomes one POST whose response body is queued for the reader. It
// satisfies io.ReadWriteCloser so it drops straight into NewClient.
type httpTransport struct {
	endpoint string
	tokenSrc TokenSource
	hc       *http.Client

	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	cond     *sync.Cond
	writeBuf []byte // accumulates a partial outgoing frame until newline-terminated
	readBuf  []byte // response frames awaiting the reader
	closed   bool
}

func newHTTPTransport(endpoint string, tokenSrc TokenSource, hc *http.Client) *httpTransport {
	if hc == nil {
		hc = http.DefaultClient
	}
	ctx, cancel := context.WithCancel(context.Background())
	t := &httpTransport{endpoint: endpoint, tokenSrc: tokenSrc, hc: hc, ctx: ctx, cancel: cancel}
	t.cond = sync.NewCond(&t.mu)
	return t
}

// Write accumulates outgoing bytes and, on each complete newline-terminated frame, performs the
// HTTP round-trip. Response frames are queued for Read.
func (t *httpTransport) Write(p []byte) (int, error) {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	t.writeBuf = append(t.writeBuf, p...)
	frames := t.takeFramesLocked()
	t.mu.Unlock()

	for _, frame := range frames {
		t.roundTrip(frame)
	}
	return len(p), nil
}

// takeFramesLocked splits complete newline-terminated frames out of writeBuf.
func (t *httpTransport) takeFramesLocked() [][]byte {
	var frames [][]byte
	for {
		i := bytes.IndexByte(t.writeBuf, '\n')
		if i < 0 {
			break
		}
		frame := append([]byte(nil), t.writeBuf[:i]...)
		t.writeBuf = t.writeBuf[i+1:]
		if len(bytes.TrimSpace(frame)) > 0 {
			frames = append(frames, frame)
		}
	}
	return frames
}

// roundTrip POSTs one request frame and queues the resulting response frame(s). Any failure is
// surfaced back to the caller as a synthesized JSON-RPC error carrying the request's id, so a
// single failed call fails on its own rather than tearing down the whole connection.
func (t *httpTransport) roundTrip(frame []byte) {
	resp, err := t.doPost(frame)
	if err != nil {
		t.queue(synthErrorFrame(frame, err.Error()))
		return
	}
	for _, f := range resp {
		t.queue(f)
	}
}

func (t *httpTransport) doPost(frame []byte) ([][]byte, error) {
	req, err := http.NewRequestWithContext(t.ctx, http.MethodPost, t.endpoint, bytes.NewReader(frame))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if t.tokenSrc != nil {
		token, err := t.tokenSrc(t.ctx)
		if err != nil {
			return nil, fmt.Errorf("token source: %w", err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}
	httpResp, err := t.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", httpResp.StatusCode, strings.TrimSpace(string(body)))
	}
	return extractFrames(httpResp.Header.Get("Content-Type"), body), nil
}

// queue appends a newline-terminated frame to the read buffer and wakes a blocked reader.
func (t *httpTransport) queue(frame []byte) {
	t.mu.Lock()
	t.readBuf = append(t.readBuf, frame...)
	if len(frame) == 0 || frame[len(frame)-1] != '\n' {
		t.readBuf = append(t.readBuf, '\n')
	}
	t.cond.Signal()
	t.mu.Unlock()
}

// Read returns queued response bytes, blocking until some are available or the transport closes.
func (t *httpTransport) Read(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for len(t.readBuf) == 0 && !t.closed {
		t.cond.Wait()
	}
	if len(t.readBuf) == 0 && t.closed {
		return 0, io.EOF
	}
	n := copy(p, t.readBuf)
	t.readBuf = t.readBuf[n:]
	return n, nil
}

// Close cancels in-flight requests and unblocks the reader.
func (t *httpTransport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.cancel()
	t.cond.Broadcast()
	t.mu.Unlock()
	return nil
}

// extractFrames turns an HTTP response body into one or more compact JSON-RPC frames. For
// application/json it compacts the single object; for text/event-stream it extracts each SSE
// `data:` payload. Compaction guarantees one JSON object per newline for the scanner upstream.
func extractFrames(contentType string, body []byte) [][]byte {
	if strings.Contains(contentType, "text/event-stream") {
		return sseFrames(body)
	}
	if compact := compactJSON(body); compact != nil {
		return [][]byte{compact}
	}
	return nil
}

func sseFrames(body []byte) [][]byte {
	var frames [][]byte
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimRight(line, "\r")
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if compact := compactJSON([]byte(payload)); compact != nil {
			frames = append(frames, compact)
		}
	}
	return frames
}

func compactJSON(b []byte) []byte {
	b = bytes.TrimSpace(b)
	if len(b) == 0 {
		return nil
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return nil
	}
	return buf.Bytes()
}

// synthErrorFrame builds a JSON-RPC error response echoing the request's id, so a transport-level
// failure fails only the originating call.
func synthErrorFrame(reqFrame []byte, msg string) []byte {
	var req struct {
		ID json.RawMessage `json:"id"`
	}
	_ = json.Unmarshal(reqFrame, &req)
	id := req.ID
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	return []byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":%s,"error":{"code":-32000,"message":%q}}`, id, msg))
}
