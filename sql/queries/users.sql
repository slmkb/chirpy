-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, password_hash)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE email = $1;

-- name: DropUsers :exec
DELETE FROM users;