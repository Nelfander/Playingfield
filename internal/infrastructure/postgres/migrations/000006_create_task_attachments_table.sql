CREATE TABLE task_attachments (
    id bigserial PRIMARY KEY,
    task_id bigint NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id bigint NOT NULL REFERENCES users(id),
    file_name TEXT NOT NULL,
    file_key TEXT NOT NULL,
    file_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for performance
CREATE INDEX idx_task_attachments_task_id ON task_attachments(task_id);