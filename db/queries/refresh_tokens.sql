-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, device_name, device_id, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetRefreshTokenByToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- Active = not revoked and not expired.
-- name: ListActiveRefreshTokensByUserID :many
SELECT * FROM refresh_tokens
WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now()
ORDER BY created_at DESC;

-- name: CountActiveRefreshTokensByUserID :one
SELECT count(*) FROM refresh_tokens
WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now();

-- name: RevokeRefreshToken :execrows
UPDATE refresh_tokens SET revoked_at = now() WHERE token = $1;

-- Rotation only succeeds once per token; 0 rows means it was already revoked.
-- name: ReplaceRefreshToken :execrows
UPDATE refresh_tokens SET revoked_at = now(), replaced_by_token = $2
WHERE token = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshTokenByIDForUser :execrows
UPDATE refresh_tokens SET revoked_at = now()
WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL;

-- name: RevokeAllRefreshTokensByUserID :execrows
UPDATE refresh_tokens SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;
