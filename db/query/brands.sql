-- name: CreateBrand :one
INSERT INTO brands (id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, name, slug, created_at, updated_at;

-- name: ListBrands :many
SELECT id, name, slug, created_at, updated_at
FROM brands
ORDER BY created_at DESC;
