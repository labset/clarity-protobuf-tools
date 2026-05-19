package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/schema/gen/test/inventory/v1"
)

func TestDeleteProduct(t *testing.T) {
	env := setupTest(t)
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

	// Verify outbox events: create + delete
	jobs := queryRiverJobs(t, ctx, env.pool)
	require.Len(t, jobs, 2)
	assert.Equal(t, "create_product", jobs[0].Kind)
	assert.Equal(t, "delete_product", jobs[1].Kind)

	var args map[string]any
	require.NoError(t, json.Unmarshal(jobs[1].Args, &args))
	assert.Equal(t, id, args["entity_id"])
	assert.NotEmpty(t, args["occurred_at"])
}

func TestDeleteProduct_NotFound(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	_, err := env.client.DeleteProduct(ctx, connect.NewRequest(&inventoryv1.DeleteProductRequest{
		Id: "00000000-0000-0000-0000-000000000001",
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}
