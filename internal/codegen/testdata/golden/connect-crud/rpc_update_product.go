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

	current, err := h.store.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	item := req.Msg.GetItem()
	mask := req.Msg.GetUpdateMask()
	params := productFromUpdate(item, id)

	if mask != nil && len(mask.GetPaths()) > 0 {
		allowed := make(map[string]bool)
		for _, p := range mask.GetPaths() {
			allowed[p] = true
		}
		if !allowed["name"] {
			params.Name = current.Name
		}
		if !allowed["price"] {
			params.Price = current.Price
		}
	}

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
