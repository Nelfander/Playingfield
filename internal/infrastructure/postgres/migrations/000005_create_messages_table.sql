-- name: create_messages_table
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    
    -- If this is filled, it's a Project Group Chat
    project_id BIGINT REFERENCES projects(id) ON DELETE CASCADE,
    
    -- If this is filled, it's a Private DM
    receiver_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- This ensures a message is either for a project OR for a person
    CONSTRAINT check_message_type CHECK (
        (project_id IS NOT NULL AND receiver_id IS NULL) OR  
        (project_id IS NULL AND receiver_id IS NOT NULL)
    )
);