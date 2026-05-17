package api

import (
	"context"
	"encoding/base64"
	"fmt"

	"connectrpc.com/connect"
	"github.com/gofrs/uuid/v5"

	inventoryv1 "github.com/acme/inventory/v1"
)

func (h *productHandler) ListProducts(
	ctx context.Context,
	req *connect.Request[inventoryv1.ListProductsRequest],
) (*connect.Response[inventoryv1.ListProductsResponse], error) {
	pageSize := req.Msg.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50
	}

	var cursor uuid.UUID
	if token := req.Msg.GetPageToken(); token != "" {
		raw, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid page token"))
		}
		cursor, err = uuid.FromString(string(raw))
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid page token"))
		}
	}

	_ = cursor
	_ = pageSize

	rows, err := h.store.ListProducts(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*inventoryv1.Product, len(rows))
	for i, row := range rows {
		items[i] = productToProto(row)
	}

	resp := &inventoryv1.ListProductsResponse{
		Items: items,
	}

	return connect.NewResponse(resp), nil
}
