package plugin

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"

	"github.com/datamaia/andromeda/internal/ports"
)

// fakePlugin serves ARP/MCP over one end of a pipe.
func fakePlugin(conn net.Conn) {
	go func() {
		sc := bufio.NewScanner(conn)
		enc := json.NewEncoder(conn)
		for sc.Scan() {
			var req struct {
				ID     int64  `json:"id"`
				Method string `json:"method"`
			}
			if json.Unmarshal(sc.Bytes(), &req) != nil {
				continue
			}
			var result any
			switch req.Method {
			case "initialize":
				result = map[string]any{"protocolVersion": "2025-06-18", "serverInfo": map[string]any{"name": "myplugin", "version": "0.1"}}
			case "tools/list":
				result = map[string]any{"tools": []map[string]any{{"name": "greet", "description": "greet someone", "inputSchema": map[string]any{"type": "object"}}}}
			case "tools/call":
				result = map[string]any{"content": []map[string]any{{"type": "text", "text": "hi from plugin"}}}
			}
			enc.Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
		}
	}()
}

func TestConnectAndBridgeTools(t *testing.T) {
	ctx := context.Background()
	client, server := net.Pipe()
	fakePlugin(server)
	rt := NewRuntime()

	in, err := rt.Connect(ctx, "myplugin", client)
	if err != nil {
		t.Fatal(err)
	}
	if in.State() != StateRunning {
		t.Fatalf("state = %q", in.State())
	}
	if in.info.ServerInfo.Name != "myplugin" {
		t.Errorf("server info = %+v", in.info)
	}

	tools, err := in.Tools(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 1 {
		t.Fatalf("bridged %d tools", len(tools))
	}
	d, _ := tools[0].Describe(ctx)
	if d.Name != "mcp_myplugin_greet" || d.Origin != "mcp" {
		t.Fatalf("descriptor = %+v", d)
	}

	st, _ := tools[0].Execute(ctx, ports.ToolExecuteRequest{Input: []byte(`{}`)})
	var text string
	for {
		ev, err := st.Next(ctx)
		if err == ports.ErrEndOfStream {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if ev.Terminal {
			text = ev.Text
		}
	}
	if text != "hi from plugin" {
		t.Errorf("tool result = %q", text)
	}

	if len(rt.List()) != 1 {
		t.Errorf("runtime should list one instance")
	}
	if err := in.Stop(); err != nil {
		t.Fatal(err)
	}
	if in.State() != StateStopped {
		t.Errorf("state after stop = %q", in.State())
	}
}

func TestSpawnRealSubprocess(t *testing.T) {
	// A tiny ARP server implemented in POSIX sh: read JSON-RPC lines, reply to initialize.
	// This exercises the real subprocess stdio wiring (Spawn).
	ctx := context.Background()
	script := `while IFS= read -r line; do
  id=$(printf '%s' "$line" | sed -n 's/.*"id":\([0-9]*\).*/\1/p')
  case "$line" in
    *initialize*) printf '{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"2025-06-18","serverInfo":{"name":"sh","version":"1"}}}\n' "$id" ;;
    *) printf '{"jsonrpc":"2.0","id":%s,"result":{}}\n' "$id" ;;
  esac
done`
	rt := NewRuntime()
	in, err := rt.Spawn(ctx, "sh-plugin", "sh", "-c", script)
	if err != nil {
		t.Fatalf("spawn: %v", err)
	}
	defer in.Stop()
	if in.State() != StateRunning || in.info.ServerInfo.Name != "sh" {
		t.Fatalf("instance = state %q info %+v", in.State(), in.info)
	}
}
