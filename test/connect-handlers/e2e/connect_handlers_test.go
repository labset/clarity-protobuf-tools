package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-handlers/schema/gen/test/inventory/v1"
	inventoryv1connect "github.com/labset/clarity-protobuf-tools/test/connect-handlers/schema/gen/test/inventory/v1/inventoryv1connect"

	api "github.com/labset/clarity-protobuf-tools/test/connect-handlers/internal/test/inventory/v1/api"
)

func TestConnectHandlers_Unimplemented(t *testing.T) {
	mux := http.NewServeMux()
	path, handler := api.NewProductServiceHandler(api.ProductServiceDeps{})
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := inventoryv1connect.NewProductServiceClient(
		http.DefaultClient,
		server.URL,
	)

	ctx := context.Background()

	t.Run("CreateProduct returns Unimplemented", func(t *testing.T) {
		_, err := client.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
			Name:  "Test",
			Price: 100,
		}))
		require.Error(t, err)
		assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
	})

	t.Run("GetProduct returns Unimplemented", func(t *testing.T) {
		_, err := client.GetProduct(ctx, connect.NewRequest(&inventoryv1.GetProductRequest{
			Id: "test-id",
		}))
		require.Error(t, err)
		assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
	})

	t.Run("ListProducts returns Unimplemented", func(t *testing.T) {
		_, err := client.ListProducts(ctx, connect.NewRequest(&inventoryv1.ListProductsRequest{
			PageSize: 10,
		}))
		require.Error(t, err)
		assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
	})
}
