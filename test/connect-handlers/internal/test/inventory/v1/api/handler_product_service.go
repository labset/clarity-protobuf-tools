package api

import (
	"net/http"

	"connectrpc.com/connect"

	inventoryv1connect "github.com/labset/clarity-protobuf-tools/test/connect-handlers/schema/gen/test/inventory/v1/inventoryv1connect"
)

type productServiceHandler struct {
	inventoryv1connect.UnimplementedProductServiceHandler
}

// NewProductServiceHandler creates a Connect handler for ProductService.
func NewProductServiceHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	h := &productServiceHandler{}
	return inventoryv1connect.NewProductServiceHandler(h, opts...)
}
