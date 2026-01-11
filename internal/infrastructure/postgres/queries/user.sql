-- name: CreateUser :one
INSERT INTO users (email, password_hash, role, status)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, role, status, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, status, created_at
FROM users
WHERE email = $1;

