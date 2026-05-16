-- name: GetProduct :one
SELECT id, created_at, updated_at, deleted_at, name, price
FROM acme_inventory.product
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT id, created_at, updated_at, deleted_at, name, price
FROM acme_inventory.product
WHERE deleted_at IS NULL;

-- name: CreateProduct :one
INSERT INTO acme_inventory.product (
  id, created_at, updated_at, deleted_at, name, price
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateProduct :one
UPDATE acme_inventory.product
SET name = $2, price = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SoftDeleteProduct :exec
UPDATE acme_inventory.product
SET deleted_at = NOW()
WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM acme_inventory.product
WHERE id = $1;
