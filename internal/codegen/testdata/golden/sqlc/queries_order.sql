-- name: GetOrder :one
SELECT id, created_at, updated_at, quantity, billing_address, shipping_address
FROM acme_inventory.order
WHERE id = $1;

-- name: ListOrders :many
SELECT id, created_at, updated_at, quantity, billing_address, shipping_address
FROM acme_inventory.order;

-- name: CreateOrder :exec
INSERT INTO acme_inventory.order (
  id, created_at, updated_at, quantity, billing_address, shipping_address
) VALUES (
  $1, $2, $3, $4, $5, $6
);

-- name: UpdateOrder :exec
UPDATE acme_inventory.order
SET created_at = $2, updated_at = $3, quantity = $4, billing_address = $5, shipping_address = $6
WHERE id = $1;

-- name: DeleteOrder :exec
DELETE FROM acme_inventory.order
WHERE id = $1;
