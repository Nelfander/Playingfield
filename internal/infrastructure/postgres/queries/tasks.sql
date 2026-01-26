-- name: CreateTask :one
INSERT INTO tasks (project_id, title, description, status, assigned_to)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2,
    description = $3,
    status = $4,
    assigned_to = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: GetTaskByID :one
SELECT * FROM tasks WHERE id = $1;

-- name: ListTasksForProject :many
SELECT 
    t.*, 
    u.email as assignee_email
FROM tasks t
LEFT JOIN users u ON t.assigned_to = u.id
WHERE t.project_id = $1
ORDER BY t.created_at ASC;

-- name: RecordTaskActivity :exec
INSERT INTO task_activities (task_id, user_id, action, details)
VALUES ($1, $2, $3, $4);

-- name: GetTaskHistory :many
SELECT 
    ta.id, 
    ta.task_id, 
    ta.user_id, 
    u.email as user_email, -- Add this
    ta.action, 
    ta.details, 
    ta.created_at
FROM task_activities ta
JOIN users u ON ta.user_id = u.id -- Add this join
WHERE ta.task_id = $1
ORDER BY ta.created_at DESC;