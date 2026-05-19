package api

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/mcp-tools/schema/gen/test/inventory/v1"
)

func (h *productHandler) DeleteProduct(
	ctx context.Context,
	req *connect.Request[inventoryv1.DeleteProductRequest],
) (*connect.Response[inventoryv1.DeleteProductResponse], error) {
	id, err := uuid.FromString(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	rowsAffected, err := h.store.SoftDeleteProduct(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if rowsAffected == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("product not found"))
	}

	return connect.NewResponse(&inventoryv1.DeleteProductResponse{}), nil
}
