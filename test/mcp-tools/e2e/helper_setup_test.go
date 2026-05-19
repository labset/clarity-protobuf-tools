package e2e

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	inventoryv1connect "github.com/labset/clarity-protobuf-tools/test/mcp-tools/schema/gen/test/inventory/v1/inventoryv1connect"
	mcptools "github.com/labset/clarity-protobuf-tools/test/mcp-tools/internal/test/inventory/v1/mcp"
)

func newTestMcpServer() *mcpsdk.Server {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "test-mcp-tools",
		Version: "v1",
	}, nil)
	var handler inventoryv1connect.UnimplementedProductServiceHandler
	mcptools.RegisterProductTools(server, &handler)
	return server
}
