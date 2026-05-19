package api

import (
	"context"

	"connectrpc.com/connect"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-handlers/schema/gen/test/inventory/v1"
)

func (h *productServiceHandler) ListProducts(
	_ context.Context,
	_ *connect.Request[inventoryv1.ListProductsRequest],
) (*connect.Response[inventoryv1.ListProductsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
