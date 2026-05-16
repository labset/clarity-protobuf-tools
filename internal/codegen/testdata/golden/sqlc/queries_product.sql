-- name: GetProduct :one
SELECT id, created_at, updated_at, deleted_at, name, price
FROM acme_inventory.product
WHERE id = @id AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT id, created_at, updated_at, deleted_at, name, price
FROM acme_inventory.product
WHERE deleted_at IS NULL;

-- name: CreateProduct :one
INSERT INTO acme_inventory.product (
  id, created_at, updated_at, name, price
) VALUES (
  @id, @created_at, @updated_at, @name, @price
)
RETURNING *;

-- name: UpdateProduct :one
UPDATE acme_inventory.product
SET name = @name, price = @price, updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteProduct :exec
UPDATE acme_inventory.product
SET deleted_at = NOW()
WHERE id = @id AND deleted_at IS NULL;

-- name: DeleteProduct :exec
DELETE FROM acme_inventory.product
WHERE id = @id;
