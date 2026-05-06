-- name: CreateItem :one
INSERT INTO items (id, title, description, brand_id)
VALUES ($1, $2, $3, $4)
RETURNING id, title, description, brand_id, created_at, updated_at;

-- name: GetItemByID :one
SELECT i.id,
       i.title,
       i.description,
       i.brand_id,
       i.created_at,
       i.updated_at,
       b.id AS brand_ref_id,
       b.name AS brand_name,
       b.slug AS brand_slug
FROM items i
LEFT JOIN brands b ON b.id = i.brand_id
WHERE i.id = $1;

-- name: ListItems :many
SELECT i.id,
       i.title,
       i.description,
       i.brand_id,
       i.created_at,
       i.updated_at,
       b.id AS brand_ref_id,
       b.name AS brand_name,
       b.slug AS brand_slug
FROM items i
LEFT JOIN brands b ON b.id = i.brand_id
ORDER BY i.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountItems :one
SELECT COUNT(*)::bigint
FROM items;
