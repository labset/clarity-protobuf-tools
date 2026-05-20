package api

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/kafka-worker/schema/gen/test/inventory/v1"
)

func (h *productHandler) GetProduct(
	ctx context.Context,
	req *connect.Request[inventoryv1.GetProductRequest],
) (*connect.Response[inventoryv1.GetProductResponse], error) {
	id, err := uuid.FromString(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	row, err := h.store.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inventoryv1.GetProductResponse{
		Item: productToProto(row),
	}), nil
}
