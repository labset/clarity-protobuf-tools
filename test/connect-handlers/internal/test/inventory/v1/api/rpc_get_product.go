package api

import (
	"context"

	"connectrpc.com/connect"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-handlers/schema/gen/test/inventory/v1"
)

func (h *productServiceHandler) GetProduct(
	_ context.Context,
	_ *connect.Request[inventoryv1.GetProductRequest],
) (*connect.Response[inventoryv1.GetProductResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
