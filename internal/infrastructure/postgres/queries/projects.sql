-- name: GetProjectByID :one
SELECT id, name, description, owner_id, created_at
FROM projects
WHERE id = $1;

-- name: CreateProject :one
INSERT INTO projects (name, description, owner_id)
VALUES ($1, $2, $3)
RETURNING id, name, description, owner_id, created_at;

-- name: ListProjectsByOwner :many
SELECT 
    p.id, 
    p.name, 
    p.description, 
    p.owner_id, 
    p.created_at,
    u.email AS owner_name
FROM projects p
LEFT JOIN users u ON p.owner_id = u.id
WHERE p.owner_id = $1 
   OR p.id IN (SELECT project_id FROM project_users WHERE user_id = $1)
ORDER BY p.created_at ASC;


