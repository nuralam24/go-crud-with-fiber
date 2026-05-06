-- name: GetAdminByEmail :one
SELECT id, email, password_hash, created_at, updated_at
FROM admins
WHERE email = $1;

-- name: GetUserByEmail :one
SELECT id, name, phone, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, email, role, token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, email, role, token_hash, expires_at, revoked_at, created_at, updated_at;

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, email, role, token_hash, expires_at, revoked_at, created_at, updated_at
FROM refresh_tokens
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW(),
    updated_at = NOW()
WHERE id = $1;
