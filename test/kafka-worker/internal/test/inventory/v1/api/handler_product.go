package api

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/labset/clarity-protobuf-tools/test/kafka-worker/internal/test/inventory/v1/db"
	inventoryv1connect "github.com/labset/clarity-protobuf-tools/test/kafka-worker/schema/gen/test/inventory/v1/inventoryv1connect"
)

// ProductDeps holds the dependencies for the ProductService handler.
type ProductDeps struct {
	Pool  *pgxpool.Pool
	River *river.Client[pgx.Tx]
}

type productHandler struct {
	inventoryv1connect.UnimplementedProductServiceHandler
	pool  *pgxpool.Pool
	river *river.Client[pgx.Tx]
	store *db.Queries
}

// NewProductServiceHandler creates a Connect handler for ProductService.
func NewProductServiceHandler(deps ProductDeps, opts ...connect.HandlerOption) (string, http.Handler) {
	h := &productHandler{
		pool:  deps.Pool,
		river: deps.River,
		store: db.New(deps.Pool),
	}
	return inventoryv1connect.NewProductServiceHandler(h, opts...)
}
