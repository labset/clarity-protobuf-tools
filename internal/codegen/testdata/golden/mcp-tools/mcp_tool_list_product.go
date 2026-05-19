package api

import (
	"context"

	"connectrpc.com/connect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"

	inventoryv1 "github.com/acme/inventory/v1"
)

type ListProductsInput struct {
	PageSize  int32  `json:"page_size"`
	PageToken string `json:"page_token"`
}

func (t *productTools) listProducts(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input ListProductsInput,
) (*mcp.CallToolResult, struct{}, error) {
	resp, err := t.handler.ListProducts(ctx, connect.NewRequest(&inventoryv1.ListProductsRequest{
		PageSize:  input.PageSize,
		PageToken: input.PageToken,
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
