package api

import (
	"context"

	"connectrpc.com/connect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"

	inventoryv1 "github.com/acme/inventory/v1"
)

type GetProductInput struct {
	ID string `json:"id"`
}

func (t *productTools) getProduct(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input GetProductInput,
) (*mcp.CallToolResult, struct{}, error) {
	resp, err := t.handler.GetProduct(ctx, connect.NewRequest(&inventoryv1.GetProductRequest{
		Id: input.ID,
	}))
	if err != nil {
		return nil, struct{}{}, err
	}
	data, err := protojson.Marshal(resp.Msg)
	if err != nil {
		return nil, struct{}{}, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, struct{}{}, nil
}
