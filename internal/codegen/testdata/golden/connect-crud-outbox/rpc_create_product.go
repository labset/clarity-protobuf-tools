package api

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/acme/app/internal/acme/inventory/v1/outbox"
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

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer tx.Rollback(ctx)

	row, err := h.store.WithTx(tx).CreateProduct(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	_, err = h.river.InsertTx(ctx, tx, outbox.CreateProductEventArgs{
		EntityID:   id,
		OccurredAt: now.Time,
	}, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&inventoryv1.CreateProductResponse{
		Item: productToProto(row),
	}), nil
}
