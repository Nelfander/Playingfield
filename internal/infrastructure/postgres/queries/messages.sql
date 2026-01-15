-- name: CreateMessage :one
INSERT INTO messages (sender_id, content, project_id, receiver_id)
VALUES ($1, $2, $3, $4)
RETURNING id, sender_id, content, project_id, receiver_id, created_at;

-- name: GetProjectMessages :many
SELECT m.id, m.sender_id, m.content, m.project_id, m.created_at, u.email as sender_email
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE m.project_id = $1
ORDER BY m.created_at ASC;

-- name: GetDirectMessages :many
SELECT m.id, m.sender_id, m.content, m.receiver_id, m.created_at, u.email as sender_email
FROM messages m
JOIN users u ON m.sender_id = u.id
WHERE (m.sender_id = $1 AND m.receiver_id = $2)
   OR (m.sender_id = $2 AND m.receiver_id = $1)
ORDER BY m.created_at ASC;