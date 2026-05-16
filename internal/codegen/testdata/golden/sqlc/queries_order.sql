-- name: GetOrder :one
SELECT id, created_at, updated_at, deleted_at, quantity, billing_address, shipping_address
FROM acme_inventory_v1.order
WHERE id = @id AND deleted_at IS NULL;

-- name: ListOrders :many
SELECT id, created_at, updated_at, deleted_at, quantity, billing_address, shipping_address
FROM acme_inventory_v1.order
WHERE deleted_at IS NULL;

-- name: CreateOrder :one
INSERT INTO acme_inventory_v1.order (
  id, created_at, updated_at, quantity, billing_address, shipping_address
) VALUES (
  @id, @created_at, @updated_at, @quantity, @billing_address, @shipping_address
)
RETURNING *;

-- name: UpdateOrder :one
UPDATE acme_inventory_v1.order
SET quantity = @quantity, billing_address = @billing_address, shipping_address = @shipping_address, updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteOrder :exec
UPDATE acme_inventory_v1.order
SET deleted_at = NOW()
WHERE id = @id AND deleted_at IS NULL;

-- name: DeleteOrder :exec
DELETE FROM acme_inventory_v1.order
WHERE id = @id;
