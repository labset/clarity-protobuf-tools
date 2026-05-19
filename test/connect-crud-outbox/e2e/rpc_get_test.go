package e2e

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/schema/gen/test/inventory/v1"
)

func TestGetProduct(t *testing.T) {
	env := setupTest(t)
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

func TestGetProduct_NoOutboxEvent(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	created, err := env.client.CreateProduct(ctx, connect.NewRequest(&inventoryv1.CreateProductRequest{
		Item: &inventoryv1.Product{
			Name:   "ReadOnly",
			Price:  100,
			Status: inventoryv1.ProductStatus_PRODUCT_STATUS_ACTIVE,
		},
	}))
	require.NoError(t, err)
	id := created.Msg.GetItem().GetEntity().GetId()

	// Get should not enqueue a job
	_, err = env.client.GetProduct(ctx, connect.NewRequest(&inventoryv1.GetProductRequest{Id: id}))
	require.NoError(t, err)

	// List should not enqueue a job
	_, err = env.client.ListProducts(ctx, connect.NewRequest(&inventoryv1.ListProductsRequest{PageSize: 10}))
	require.NoError(t, err)

	// Only the create job should exist
	jobs := queryRiverJobs(t, ctx, env.pool)
	require.Len(t, jobs, 1)
	assert.Equal(t, "create_product", jobs[0].Kind)
}

func TestGetProduct_NotFound(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	_, err := env.client.GetProduct(ctx, connect.NewRequest(&inventoryv1.GetProductRequest{
		Id: "00000000-0000-0000-0000-000000000001",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
