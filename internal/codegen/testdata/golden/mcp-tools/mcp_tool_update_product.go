package api

import (
	"context"

	"connectrpc.com/connect"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	inventoryv1 "github.com/acme/inventory/v1"
)

type UpdateProductInput struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Price      int64    `json:"price"`
	Status     string   `json:"status"`
	UpdateMask []string `json:"update_mask"`
}

func (t *productTools) updateProduct(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	input UpdateProductInput,
) (*mcp.CallToolResult, struct{}, error) {
	req := &inventoryv1.UpdateProductRequest{
		Id: input.ID,
		Item: &inventoryv1.Product{
			Name:   input.Name,
			Price:  input.Price,
			Status: inventoryv1.ProductStatus(inventoryv1.ProductStatus_value[input.Status]),
		},
	}
	if len(input.UpdateMask) > 0 {
		req.UpdateMask = &fieldmaskpb.FieldMask{Paths: input.UpdateMask}
	}
	resp, err := t.handler.UpdateProduct(ctx, connect.NewRequest(req))
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
