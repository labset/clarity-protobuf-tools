package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	inventoryv1 "github.com/labset/clarity-protobuf-tools/test/connect-crud-outbox/schema/gen/test/inventory/v1"
)

func TestUpdateProduct_FieldMask(t *testing.T) {
	env := setupTest(t)
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

	// Verify outbox events: create + update
	jobs := queryRiverJobs(t, ctx, env.pool)
	require.Len(t, jobs, 2)
	assert.Equal(t, "create_product", jobs[0].Kind)
	assert.Equal(t, "update_product", jobs[1].Kind)

	var args map[string]any
	require.NoError(t, json.Unmarshal(jobs[1].Args, &args))
	assert.Equal(t, id, args["entity_id"])
	assert.NotEmpty(t, args["occurred_at"])
	fieldMask, ok := args["field_mask"].([]any)
	require.True(t, ok)
	assert.Equal(t, []any{"name"}, fieldMask)
}

func TestUpdateProduct_NotFound(t *testing.T) {
	env := setupTest(t)
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
