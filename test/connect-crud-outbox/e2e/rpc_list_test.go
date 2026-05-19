package e2e

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/schema/gen/test/inventory/v1"
)

func TestListProducts_Pagination(t *testing.T) {
	env := setupTest(t)
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
