-- Soft-deleted users (deleted_at IS NOT NULL) are invisible to every query.

-- name: CreateUser :one
INSERT INTO users (name, email, phone, password, login_type, is_verified, created_by, updated_by, created_at, updated_at)
VALUES (sqlc.narg('name'), sqlc.narg('email'), sqlc.narg('phone'), @password, @login_type, false, @created_by, @created_by, now(), now())
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1 AND deleted_at IS NULL;

-- Partial update: a NULL argument leaves the column unchanged.
-- name: UpdateUserProfile :one
UPDATE users SET
  name = COALESCE(sqlc.narg('name'), name),
  phone = COALESCE(sqlc.narg('phone'), phone),
  updated_by = @updated_by,
  updated_at = now()
WHERE id = @id AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :execrows
UPDATE users SET password = @password, updated_by = @id, updated_at = now()
WHERE id = @id AND deleted_at IS NULL;

-- name: MarkUserEmailVerified :execrows
UPDATE users SET is_verified = true, verified_at = @verified_at, updated_by = @id, updated_at = now()
WHERE id = @id AND deleted_at IS NULL;

-- name: SoftDeleteUser :execrows
UPDATE users SET deleted_at = now(), deleted_by = @deleted_by, updated_by = @deleted_by, updated_at = now()
WHERE id = @id AND deleted_at IS NULL;
