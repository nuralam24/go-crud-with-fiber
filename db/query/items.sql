-- name: CreateItem :one
INSERT INTO items (id, title, description)
VALUES ($1, $2, $3)
RETURNING id, title, description, created_at, updated_at;

-- name: GetItemByID :one
SELECT id, title, description, created_at, updated_at
FROM items
WHERE id = $1;

-- name: ListItems :many
SELECT id, title, description, created_at, updated_at
FROM items
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
