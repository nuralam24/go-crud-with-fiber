-- name: CreateUser :one
INSERT INTO users (id, name, phone, email, password_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, phone, email, password_hash, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, name, phone, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET name = $2,
    phone = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, phone, email, password_hash, created_at, updated_at;
