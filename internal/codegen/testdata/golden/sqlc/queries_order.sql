-- name: GetOrder :one
SELECT id, created_at, updated_at, deleted_at, quantity, billing_address, shipping_address
FROM acme_inventory.order
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListOrders :many
SELECT id, created_at, updated_at, deleted_at, quantity, billing_address, shipping_address
FROM acme_inventory.order
WHERE deleted_at IS NULL;

-- name: CreateOrder :one
INSERT INTO acme_inventory.order (
  id, created_at, updated_at, deleted_at, quantity, billing_address, shipping_address
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateOrder :one
UPDATE acme_inventory.order
SET quantity = $2, billing_address = $3, shipping_address = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteOrder :exec
UPDATE acme_inventory.order
SET deleted_at = NOW()
WHERE id = $1;

-- name: DeleteOrder :exec
DELETE FROM acme_inventory.order
WHERE id = $1;
