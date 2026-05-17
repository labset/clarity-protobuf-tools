-- name: GetProduct :one
SELECT id, created_at, updated_at, deleted_at, name, price
FROM acme_inventory_v1.product
WHERE id = @id AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT id, created_at, updated_at, deleted_at, name, price
FROM acme_inventory_v1.product
WHERE deleted_at IS NULL;

-- name: ListProductsPaginated :many
SELECT id, created_at, updated_at, deleted_at, name, price
FROM acme_inventory_v1.product
WHERE deleted_at IS NULL
  AND (@cursor::uuid = '00000000-0000-0000-0000-000000000000'::uuid OR id > @cursor)
ORDER BY id
LIMIT @page_size;

-- name: CreateProduct :one
INSERT INTO acme_inventory_v1.product (
  id, created_at, updated_at, name, price
) VALUES (
  @id, @created_at, @updated_at, @name, @price
)
RETURNING *;

-- name: UpdateProduct :one
UPDATE acme_inventory_v1.product
SET name = @name, price = @price, updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteProduct :execrows
UPDATE acme_inventory_v1.product
SET deleted_at = NOW()
WHERE id = @id AND deleted_at IS NULL;

-- name: DeleteProduct :exec
DELETE FROM acme_inventory_v1.product
WHERE id = @id;
