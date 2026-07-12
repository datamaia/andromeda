package mcp

import (
	"context"

	"github.com/datamaia/andromeda/internal/core"
	"github.com/datamaia/andromeda/internal/ports"
	"github.com/datamaia/andromeda/internal/streams"
)

// BridgedTool adapts one MCP-exposed tool to ports.ToolPort so the Tool Runtime mediates it
// like any built-in tool (permissions, trust, observability). MCP tools declare their origin as
// "mcp" and a lower default trust level.
type BridgedTool struct {
	client *Client
	info   ToolInfo
	server string
}

// BridgeTools discovers a server's tools and returns them as ToolPorts.
func BridgeTools(ctx context.Context, client *Client, server string) ([]ports.ToolPort, error) {
	infos, err := client.ListTools(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ports.ToolPort, 0, len(infos))
	for _, i := range infos {
		out = append(out, &BridgedTool{client: client, info: i, server: server})
	}
	return out, nil
}

func (b *BridgedTool) Describe(context.Context) (ports.ToolDescriptor, error) {
	return ports.ToolDescriptor{
		Name:        "mcp_" + b.server + "_" + b.info.Name,
		Namespace:   "mcp/" + b.server,
		Version:     "1",
		Description: b.info.Description,
		InputSchema: b.info.InputSchema,
		// MCP tools can do anything the server can; default to network + external service and
		// let the Tool Runtime's permission evaluation gate them (Volume 9 trust model).
		Permissions: []core.Permission{core.PermExternalServiceAccess, core.PermNetwork},
		Origin:      "mcp",
		TrustLevel:  "untrusted",
	}, nil
}

func (b *BridgedTool) Validate(_ context.Context, _ ports.JSON) (ports.ValidationResult, error) {
	// Schema validation is delegated to the server; the Runtime may add a JSON Schema check.
	return ports.ValidationResult{Valid: true}, nil
}

func (b *BridgedTool) Execute(ctx context.Context, req ports.ToolExecuteRequest) (ports.Stream[ports.ToolEvent], error) {
	res, err := b.client.CallTool(ctx, b.info.Name, req.Input)
	if err != nil {
		return streams.Slice([]ports.ToolEvent{{Kind: "terminal", Terminal: true, Outcome: "error", Text: err.Error()}}), nil
	}
	outcome := "success"
	if res.IsError {
		outcome = "error"
	}
	return streams.Slice([]ports.ToolEvent{
		{Kind: "output", Text: res.Text()},
		{Kind: "terminal", Terminal: true, Outcome: outcome, Text: res.Text()},
	}), nil
}

func (b *BridgedTool) Cancel(context.Context, core.ULID) error { return nil }
