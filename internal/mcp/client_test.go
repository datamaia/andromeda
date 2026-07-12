package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// fakeServer speaks minimal MCP over one end of a net.Pipe, responding to initialize,
// tools/list, and tools/call.
func fakeServer(t *testing.T, conn net.Conn) {
	t.Helper()
	go func() {
		sc := bufio.NewScanner(conn)
		enc := json.NewEncoder(conn)
		for sc.Scan() {
			var req struct {
				ID     int64           `json:"id"`
				Method string          `json:"method"`
				Params json.RawMessage `json:"params"`
			}
			if json.Unmarshal(sc.Bytes(), &req) != nil {
				continue
			}
			var result any
			switch req.Method {
			case "initialize":
				result = map[string]any{"protocolVersion": ProtocolVersion, "serverInfo": map[string]any{"name": "fake", "version": "1.0"}}
			case "tools/list":
				result = map[string]any{"tools": []map[string]any{
					{"name": "echo", "description": "echo back", "inputSchema": map[string]any{"type": "object"}},
				}}
			case "tools/call":
				var p struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				}
				json.Unmarshal(req.Params, &p)
				result = map[string]any{"content": []map[string]any{{"type": "text", "text": "echoed:" + string(p.Arguments)}}, "isError": false}
			}
			enc.Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
		}
	}()
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	client, server := net.Pipe()
	fakeServer(t, server)
	c := NewClient(client)
	t.Cleanup(func() { c.Close() })
	return c
}

func TestInitializeAndListTools(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	info, err := c.Initialize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if info.ServerInfo.Name != "fake" {
		t.Errorf("server name = %q", info.ServerInfo.Name)
	}
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 || tools[0].Name != "echo" {
		t.Fatalf("tools = %+v", tools)
	}
}

func TestCallTool(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	res, err := c.CallTool(ctx, "echo", json.RawMessage(`{"x":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if res.Text() != `echoed:{"x":1}` {
		t.Errorf("call result = %q", res.Text())
	}
}

func TestBridgeToToolPort(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)
	tools, err := BridgeTools(ctx, c, "fake")
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 {
		t.Fatalf("bridged %d tools", len(tools))
	}
	d, _ := tools[0].Describe(ctx)
	if d.Name != "mcp_fake_echo" || d.Origin != "mcp" || d.TrustLevel != "untrusted" {
		t.Fatalf("descriptor = %+v", d)
	}
	st, err := tools[0].Execute(ctx, ports.ToolExecuteRequest{Input: json.RawMessage(`{"a":2}`)})
	if err != nil {
		t.Fatal(err)
	}
	var term ports.ToolEvent
	for {
		ev, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if ev.Terminal {
			term = ev
		}
	}
	if term.Outcome != "success" || term.Text != `echoed:{"a":2}` {
		t.Fatalf("bridged execute result = %+v", term)
	}
}
