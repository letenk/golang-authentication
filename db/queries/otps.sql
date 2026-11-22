-- name: CreateOTP :one
INSERT INTO otps (user_id, code, purpose, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindValidOTP :one
SELECT * FROM otps
WHERE user_id = $1
  AND purpose = $2
  AND code = $3
  AND expires_at > now()
  AND used_at IS NULL
LIMIT 1;

-- Only an unused OTP can be consumed; 0 rows means it was already used.
-- name: MarkOTPUsed :execrows
UPDATE otps SET used_at = now()
WHERE id = $1 AND used_at IS NULL;

-- name: InvalidateOTPsByUserAndPurpose :exec
UPDATE otps SET used_at = now()
WHERE user_id = $1 AND purpose = $2 AND used_at IS NULL;
