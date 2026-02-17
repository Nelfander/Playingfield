-- name: CreateUser :one
INSERT INTO users (email, password_hash, role, status)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, role, status, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, status, created_at
FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT id, email FROM users 
ORDER BY email ASC;



-- name: ScrubUserAccount :exec
-- Wipes identity but keeps the ID record for history/attachments
UPDATE users 
SET 
    email = 'deleted_' || id || '@playingfield.internal',
    password_hash = 'SCRUBBED',
    status = 'deleted'
WHERE id = $1;

-- name: DeleteProjectsByOwner :exec
-- Wipes projects the user created/owns
DELETE FROM projects WHERE owner_id = $1;

-- name: RemoveUserFromAllProjectMemberships :exec
-- Removes them from project_users (access control)
DELETE FROM project_users WHERE user_id = $1;

-- name: UnassignUserFromAllTasks :exec
-- Keeps the task, but clears the 'assigned_to' field
UPDATE tasks SET assigned_to = NULL WHERE assigned_to = $1;