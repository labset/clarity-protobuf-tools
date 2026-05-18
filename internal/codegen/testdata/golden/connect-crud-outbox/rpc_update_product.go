package api

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"github.com/acme/app/internal/acme/inventory/v1/outbox"
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

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer tx.Rollback(ctx)

	store := h.store.WithTx(tx)

	current, err := store.GetProduct(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	item := req.Msg.GetItem()
	mask := req.Msg.GetUpdateMask()
	params := productFromUpdate(item, id)

	var fieldMask []string
	if mask != nil && len(mask.GetPaths()) > 0 {
		fieldMask = mask.GetPaths()
		allowed := make(map[string]bool)
		for _, p := range fieldMask {
			allowed[p] = true
		}
		if !allowed["name"] {
			params.Name = current.Name
		}
		if !allowed["price"] {
			params.Price = current.Price
		}
		if !allowed["status"] {
			params.Status = current.Status
		}
	}

	row, err := store.UpdateProduct(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	_, err = h.river.InsertTx(ctx, tx, outbox.UpdateProductEventArgs{
		EntityID:   id,
		FieldMask:  fieldMask,
		OccurredAt: time.Now(),
	}, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inventoryv1.UpdateProductResponse{
		Item: productToProto(row),
	}), nil
}
