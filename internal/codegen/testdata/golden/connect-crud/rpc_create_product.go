package api

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgtype"

	inventoryv1 "github.com/acme/inventory/v1"
)

func (h *productHandler) CreateProduct(
	ctx context.Context,
	req *connect.Request[inventoryv1.CreateProductRequest],
) (*connect.Response[inventoryv1.CreateProductResponse], error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	params := productFromCreate(req.Msg.GetItem(), id, now)

	row, err := h.store.CreateProduct(ctx, params)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inventoryv1.CreateProductResponse{
		Item: productToProto(row),
	}), nil
}
