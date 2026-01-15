-- name: AddUserToProject :one
INSERT INTO project_users (project_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING id, project_id, user_id, role;

-- name: RemoveUserFromProject :exec
DELETE FROM project_users
WHERE project_id = $1 AND user_id = $2;

-- name: ListUsersInProject :many
SELECT u.id, u.email, pu.role
FROM users u
JOIN project_users pu ON u.id = pu.user_id
WHERE pu.project_id = $1;

-- name: CheckSharedProject :one
SELECT EXISTS (
    SELECT 1 
    FROM project_users pu1
    JOIN project_users pu2 ON pu1.project_id = pu2.project_id
    WHERE pu1.user_id = sqlc.arg('sender_id') AND pu2.user_id = sqlc.arg('receiver_id')
) AS shared;


