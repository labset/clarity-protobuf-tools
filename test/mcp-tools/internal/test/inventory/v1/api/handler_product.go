package api

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/labset/clarity-protobuf-tools/test/mcp-tools/internal/test/inventory/v1/db"
	inventoryv1connect "github.com/labset/clarity-protobuf-tools/test/mcp-tools/schema/gen/test/inventory/v1/inventoryv1connect"
)

// ProductDeps holds the dependencies for the ProductService handler.
type ProductDeps struct {
	Pool *pgxpool.Pool
}

type productHandler struct {
	inventoryv1connect.UnimplementedProductServiceHandler
	store *db.Queries
}

// NewProductServiceHandler creates a Connect handler for ProductService.
func NewProductServiceHandler(deps ProductDeps, opts ...connect.HandlerOption) (string, http.Handler) {
	h := &productHandler{
		store: db.New(deps.Pool),
	}
	return inventoryv1connect.NewProductServiceHandler(h, opts...)
}
