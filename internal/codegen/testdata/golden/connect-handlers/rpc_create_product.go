package api

import (
	"context"

	"connectrpc.com/connect"

	inventoryv1 "github.com/acme/inventory/v1"
)

func (h *productServiceHandler) CreateProduct(
	_ context.Context,
	_ *connect.Request[inventoryv1.CreateProductRequest],
) (*connect.Response[inventoryv1.CreateProductResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
