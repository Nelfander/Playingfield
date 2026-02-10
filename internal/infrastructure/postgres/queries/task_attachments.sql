-- name: CreateAttachment :one
INSERT INTO task_attachments (
    task_id, 
    user_id, 
    file_name, 
    file_key, 
    file_url
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetAttachmentsByTask :many
-- This will be useful for showing a list of files on the task page
SELECT 
    id, 
    task_id, 
    user_id, 
    file_name, 
    file_url, 
    created_at
FROM task_attachments 
WHERE task_id = $1 
ORDER BY created_at DESC;

-- name: DeleteAttachment :exec
-- For when a user wants to remove an attachment
DELETE FROM task_attachments 
WHERE id = $1;

-- name: GetAttachmentByID :one
-- Useful for verifying ownership before deletion
SELECT * FROM task_attachments WHERE id = $1;