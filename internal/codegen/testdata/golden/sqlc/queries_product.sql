-- name: GetProduct :one
SELECT id, created_at, updated_at, name, price
FROM acme_inventory.product
WHERE id = $1;

-- name: ListProducts :many
SELECT id, created_at, updated_at, name, price
FROM acme_inventory.product;

-- name: CreateProduct :exec
INSERT INTO acme_inventory.product (
  id, created_at, updated_at, name, price
) VALUES (
  $1, $2, $3, $4, $5
);

-- name: UpdateProduct :exec
UPDATE acme_inventory.product
SET created_at = $2, updated_at = $3, name = $4, price = $5
WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM acme_inventory.product
WHERE id = $1;
