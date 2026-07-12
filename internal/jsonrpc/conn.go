package jsonrpc

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string { return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message) }

// Conn is a JSON-RPC 2.0 client over a newline-delimited stream.
type Conn struct {
	rw      io.ReadWriteCloser
	enc     *json.Encoder
	writeMu sync.Mutex // serializes writes; never held with mu
	mu      sync.Mutex // guards nextID and the pending map only
	nextID  int64
	pending map[int64]chan Response
	closed  chan struct{}
	closeMu sync.Once
	readErr error
}

// New starts a Conn over rw and launches its read loop.
func New(rw io.ReadWriteCloser) *Conn {
	c := &Conn{
		rw:      rw,
		enc:     json.NewEncoder(rw),
		pending: map[int64]chan Response{},
		closed:  make(chan struct{}),
	}
	go c.readLoop()
	return c
}

func (c *Conn) readLoop() {
	sc := bufio.NewScanner(c.rw)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var resp Response
		if err := json.Unmarshal(line, &resp); err != nil {
			continue // ignore non-response frames (notifications)
		}
		c.mu.Lock()
		ch, ok := c.pending[resp.ID]
		if ok {
			delete(c.pending, resp.ID)
		}
		c.mu.Unlock()
		if ok {
			ch <- resp
		}
	}
	c.readErr = sc.Err()
	c.shutdown()
}

// Call sends a request and waits for the matching response or context cancellation.
func (c *Conn) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	select {
	case <-c.closed:
		return nil, fmt.Errorf("jsonrpc: connection closed")
	default:
	}
	var raw json.RawMessage
	if params != nil {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		raw = b
	}

	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan Response, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	// Write outside the state mutex so the read loop can deliver responses concurrently
	// (net.Pipe and pipes make Encode a blocking write). writeMu just serializes writers.
	c.writeMu.Lock()
	err := c.enc.Encode(Request{JSONRPC: "2.0", ID: id, Method: method, Params: raw})
	c.writeMu.Unlock()
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case <-c.closed:
		return nil, fmt.Errorf("jsonrpc: connection closed")
	case resp := <-ch:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Result, nil
	}
}

func (c *Conn) shutdown() {
	c.closeMu.Do(func() {
		close(c.closed)
		c.mu.Lock()
		for id, ch := range c.pending {
			close(ch)
			delete(c.pending, id)
		}
		c.mu.Unlock()
	})
}

// Close closes the underlying stream and the connection.
func (c *Conn) Close() error {
	c.shutdown()
	return c.rw.Close()
}
