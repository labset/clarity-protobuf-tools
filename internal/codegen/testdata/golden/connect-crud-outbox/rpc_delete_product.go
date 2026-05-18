package api

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"

	"github.com/acme/app/internal/acme/inventory/v1/outbox"
	inventoryv1 "github.com/acme/inventory/v1"
)

func (h *productHandler) DeleteProduct(
	ctx context.Context,
	req *connect.Request[inventoryv1.DeleteProductRequest],
) (*connect.Response[inventoryv1.DeleteProductResponse], error) {
	id, err := uuid.FromString(req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer tx.Rollback(ctx)

	rowsAffected, err := h.store.WithTx(tx).SoftDeleteProduct(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if rowsAffected == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("product not found"))
	}

	_, err = h.river.InsertTx(ctx, tx, outbox.DeleteProductEventArgs{
		EntityID:   id,
		OccurredAt: time.Now(),
	}, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inventoryv1.DeleteProductResponse{}), nil
}
