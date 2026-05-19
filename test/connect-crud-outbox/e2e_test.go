package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/internal/test/inventory/v1/api"
	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/schema/gen/test/inventory/v1"
	"github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/schema/gen/test/inventory/v1/inventoryv1connect"
)

type testEnv struct {
	pool   *pgxpool.Pool
	river  *river.Client[pgx.Tx]
	server *httptest.Server
	client inventoryv1connect.ProductServiceClient
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctx := context.Background()

	// Start Postgres container
	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pgContainer.Terminate(context.Background()))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Run River migrations
	riverMigrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	require.NoError(t, err)
	_, err = riverMigrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	require.NoError(t, err)

	// Apply generated schema
	schema, err := os.ReadFile("internal/test/inventory/v1/sql/schema.sql")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, string(schema))
	require.NoError(t, err)

	// Create River client (no workers — we only inspect enqueued jobs)
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	require.NoError(t, err)

	// Start Connect server
	mux := http.NewServeMux()
	path, handler := api.NewProductServiceHandler(api.ProductDeps{
		Pool:  pool,
		River: riverClient,
	})
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := inventoryv1connect.NewProductServiceClient(
		http.DefaultClient,
		server.URL,
	)

	return &testEnv{
		pool:   pool,
		river:  riverClient,
		server: server,
		client: client,
	}
}

func TestCreateProduct(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	resp, err := env.client.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
		Item: &inventoryv1.Product{
			Name:   "Widget",
			Price:  1999,
			Status: inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE,
		},
	}))
	require.NoError(t, err)

	item := resp.Msg.GetItem()
	assert.NotEmpty(t, item.GetEntity().GetId())
	assert.Equal(t, "Widget", item.GetName())
	assert.Equal(t, int64(1999), item.GetPrice())
	assert.Equal(t, inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE, item.GetStatus())
	assert.NotNil(t, item.GetEntity().GetCreatedAt())
	assert.NotNil(t, item.GetEntity().GetUpdatedAt())
}

func TestGetProduct(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	created, err := env.client.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
		Item: &inventoryv1.Product{
			Name:   "Gadget",
			Price:  2999,
			Status: inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE,
		},
	}))
	require.NoError(t, err)
	id := created.Msg.GetItem().GetEntity().GetId()

	resp, err := env.client.GetProduct(ctx, connect.NewRequest(&inventoryv1.GetProductRequest{
		Id: id,
	}))
	require.NoError(t, err)

	item := resp.Msg.GetItem()
	assert.Equal(t, id, item.GetEntity().GetId())
	assert.Equal(t, "Gadget", item.GetName())
	assert.Equal(t, int64(2999), item.GetPrice())
}

func TestGetProduct_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	_, err := env.client.GetProduct(ctx, connect.NewRequest(&inventoryv1.GetProductRequest{
		Id: "00000000-0000-0000-0000-000000000001",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestListProducts_Pagination(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	// Create 3 products
	for i, name := range []string{"Alpha", "Beta", "Gamma"} {
		_, err := env.client.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
			Item: &inventoryv1.Product{
				Name:   name,
				Price:  int64((i + 1) * 100),
				Status: inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE,
			},
		}))
		require.NoError(t, err)
	}

	// List with page_size=2
	resp1, err := env.client.ListProducts(ctx, connect.NewRequest(&inventoryv1.ListProductsRequest{
		PageSize: 2,
	}))
	require.NoError(t, err)
	assert.Len(t, resp1.Msg.GetItems(), 2)
	assert.NotEmpty(t, resp1.Msg.GetNextPageToken())

	// Fetch next page
	resp2, err := env.client.ListProducts(ctx, connect.NewRequest(&inventoryv1.ListProductsRequest{
		PageSize:  2,
		PageToken: resp1.Msg.GetNextPageToken(),
	}))
	require.NoError(t, err)
	assert.Len(t, resp2.Msg.GetItems(), 1)
	assert.Empty(t, resp2.Msg.GetNextPageToken())
}

func TestUpdateProduct_FieldMask(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	created, err := env.client.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
		Item: &inventoryv1.Product{
			Name:   "Original",
			Price:  500,
			Status: inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE,
		},
	}))
	require.NoError(t, err)
	id := created.Msg.GetItem().GetEntity().GetId()

	// Update only the name
	resp, err := env.client.UpdateProduct(ctx, connect.NewRequest(&inventoryv1.UpdateProductRequest{
		Id: id,
		Item: &inventoryv1.Product{
			Name: "Updated",
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
	}))
	require.NoError(t, err)

	item := resp.Msg.GetItem()
	assert.Equal(t, "Updated", item.GetName())
	assert.Equal(t, int64(500), item.GetPrice())
	assert.Equal(t, inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE, item.GetStatus())
}

func TestUpdateProduct_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	_, err := env.client.UpdateProduct(ctx, connect.NewRequest(&inventoryv1.UpdateProductRequest{
		Id: "00000000-0000-0000-0000-000000000001",
		Item: &inventoryv1.Product{
			Name: "Nope",
		},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestDeleteProduct(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	created, err := env.client.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
		Item: &inventoryv1.Product{
			Name:   "ToDelete",
			Price:  100,
			Status: inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE,
		},
	}))
	require.NoError(t, err)
	id := created.Msg.GetItem().GetEntity().GetId()

	// Delete
	_, err = env.client.DeleteProduct(ctx, connect.NewRequest(&inventoryv1.DeleteProductRequest{
		Id: id,
	}))
	require.NoError(t, err)

	// Subsequent Get returns not found (soft-deleted)
	_, err = env.client.GetProduct(ctx, connect.NewRequest(&inventoryv1.GetProductRequest{
		Id: id,
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestDeleteProduct_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	ctx := context.Background()

	_, err := env.client.DeleteProduct(ctx, connect.NewRequest(&inventoryv1.DeleteProductRequest{
		Id: "00000000-0000-0000-0000-000000000001",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
