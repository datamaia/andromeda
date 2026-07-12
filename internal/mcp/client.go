package mcp

import (
	"context"
	"encoding/json"
	"io"

	"github.com/datamaia/andromeda/internal/jsonrpc"
)

// ProtocolVersion is the MCP revision the client advertises. The pinned/certified revision set
// is PENDING VALIDATION per ADR-010; this is the value negotiated at initialize.
const ProtocolVersion = "2025-06-18"

// Client is an MCP client over a JSON-RPC connection.
type Client struct {
	conn *jsonrpc.Conn
}

// NewClient wraps a byte stream (typically a server subprocess's stdio) in an MCP client.
func NewClient(rw io.ReadWriteCloser) *Client {
	return &Client{conn: jsonrpc.New(rw)}
}

// ServerInfo is returned by initialize.
type ServerInfo struct {
	ProtocolVersion string `json:"protocolVersion"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

// Initialize performs the MCP handshake.
func (c *Client) Initialize(ctx context.Context) (ServerInfo, error) {
	params := map[string]any{
		"protocolVersion": ProtocolVersion,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "andromeda", "version": "0.0.0-dev"},
	}
	raw, err := c.conn.Call(ctx, "initialize", params)
	if err != nil {
		return ServerInfo{}, err
	}
	var info ServerInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return ServerInfo{}, err
	}
	return info, nil
}

// ToolInfo describes an MCP-exposed tool.
type ToolInfo struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ListTools enumerates the server's tools.
func (c *Client) ListTools(ctx context.Context) ([]ToolInfo, error) {
	raw, err := c.conn.Call(ctx, "tools/list", map[string]any{})
	if err != nil {
		return nil, err
	}
	var out struct {
		Tools []ToolInfo `json:"tools"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out.Tools, nil
}

// CallResult is the result of a tools/call.
type CallResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

// Text concatenates the textual content of a call result.
func (r CallResult) Text() string {
	var s string
	for _, c := range r.Content {
		if c.Type == "text" {
			s += c.Text
		}
	}
	return s
}

// CallTool invokes a tool by name with JSON arguments.
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (CallResult, error) {
	params := map[string]any{"name": name}
	if len(args) > 0 {
		params["arguments"] = args
	}
	raw, err := c.conn.Call(ctx, "tools/call", params)
	if err != nil {
		return CallResult{}, err
	}
	var res CallResult
	if err := json.Unmarshal(raw, &res); err != nil {
		return CallResult{}, err
	}
	return res, nil
}

// Close closes the MCP connection.
func (c *Client) Close() error { return c.conn.Close() }
