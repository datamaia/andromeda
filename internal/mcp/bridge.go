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

// Describe returns the tool descriptor for the bridged MCP tool, namespaced by server and
// defaulting to network/external-service permissions at an untrusted trust level.
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

// Validate reports the input as valid, delegating schema validation to the MCP server.
func (b *BridgedTool) Validate(_ context.Context, _ ports.JSON) (ports.ValidationResult, error) {
	// Schema validation is delegated to the server; the Runtime may add a JSON Schema check.
	return ports.ValidationResult{Valid: true}, nil
}

// Execute calls the MCP tool on its server and streams the result as output and terminal events.
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

// Cancel is a no-op; a bridged MCP tool call is not separately cancelable.
func (b *BridgedTool) Cancel(context.Context, core.ULID) error { return nil }
