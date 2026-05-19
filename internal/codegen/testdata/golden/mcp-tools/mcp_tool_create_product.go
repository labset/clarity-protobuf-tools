package api

import (
	"context"

	"connectrpc.com/connect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"

	inventoryv1 "github.com/acme/inventory/v1"
)

type CreateProductInput struct {
	Name   string `json:"name"`
	Price  int64  `json:"price"`
	Status string `json:"status"`
}

func (t *productTools) createProduct(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input CreateProductInput,
) (*mcp.CallToolResult, struct{}, error) {
	resp, err := t.handler.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
		Item: &inventoryv1.Product{
			Name:   input.Name,
			Price:  input.Price,
			Status: inventoryv1.ProductStatus(inventoryv1.ProductStatus_value[input.Status]),
		},
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
