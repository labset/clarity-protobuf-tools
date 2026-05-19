package mcp

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"
	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/mcp-tools/schema/gen/test/inventory/v1"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func (t *productTools) getProduct(ctx context.Context, req *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
	var protoReq inventoryv1.GetProductRequest
	if err := json.Unmarshal(req.Params.Arguments, &protoReq); err != nil {
		return nil, err
	}
	resp, err := t.handler.GetProduct(ctx, connect.NewRequest(&protoReq))
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(resp.Msg)
	if err != nil {
		return nil, err
	}
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: string(data)}},
	}, nil
}
