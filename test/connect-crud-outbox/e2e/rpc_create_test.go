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

func TestCreateProduct(t *testing.T) {
	env := setupTest(t)
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

	// Verify outbox event
	jobs := queryRiverJobs(t, ctx, env.pool)
	require.Len(t, jobs, 1)
	assert.Equal(t, "create_product", jobs[0].Kind)

	var args map[string]any
	require.NoError(t, json.Unmarshal(jobs[0].Args, &args))
	assert.Equal(t, item.GetEntity().GetId(), args["entity_id"])
	assert.NotEmpty(t, args["occurred_at"])
}
