package api

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	inventoryv1 "github.com/acme/inventory/v1"
)

func (h *productHandler) UpdateProduct(
	ctx context.Context,
	req *connect.Request[inventoryv1.UpdateProductRequest],
) (*connect.Response[inventoryv1.UpdateProductResponse], error) {
	id, err := uuid.FromString(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	params := productFromUpdate(req.Msg.GetItem(), id)

	row, err := h.store.UpdateProduct(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inventoryv1.UpdateProductResponse{
		Item: productToProto(row),
	}), nil
}
