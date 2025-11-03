-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email,hashed_password)
VALUES (
	gen_random_uuid(), NOW(),NOW(), $1,$2
)
RETURNING *;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users where email = $1;

-- name: GetUserByID :one
SELECT * FROM users where id = $1;

-- name: UpdateUser :exec
UPDATE users SET updated_at = NOW(), email = $1, hashed_password = $2 where id = $3;

-- name: UpgradeUserToRed :exec
UPDATE users SET updated_at = NOW(), is_chirpy_red = true where id = $1;
