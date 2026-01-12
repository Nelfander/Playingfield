-- name: GetProjectByID :one
SELECT id, name, description, owner_id, created_at
FROM projects
WHERE id = $1;

-- name: CreateProject :one
INSERT INTO projects (name, description, owner_id)
VALUES ($1, $2, $3)
RETURNING id, name, description, owner_id, created_at;

-- name: ListProjectsByOwner :many
SELECT id, name, description, owner_id, created_at
FROM projects
WHERE owner_id = $1
ORDER BY created_at ASC;

