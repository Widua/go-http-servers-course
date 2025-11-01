-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, expires_at ,revoked_at,user_id)
VALUES (
	$1, NOW(),NOW(), NOW() + INTERVAL '1 hour' , NULL, $2
)
RETURNING *;

-- name: GetRefreshTokenByToken :one
SELECT * from refresh_tokens where token = $1;

-- name: RevokeAccessToToken :exec
UPDATE refresh_tokens SET updated_at = NOW(), revoked_at = NOW() WHERE token = $1;
