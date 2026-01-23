package messages

import (
	"context"
	"time"
)

type Message struct {
	ID          int64     `json:"id"`
	SenderID    int64     `json:"sender_id"`
	SenderEmail string    `json:"sender_email"`
	Content     string    `json:"content"`
	ProjectID   *int64    `json:"project_id,omitempty"`
	ReceiverID  *int64    `json:"receiver_id,omitempty"` // pointer for nullable, without pointer it defaults to 0 . with pointer it can be nil
	CreatedAt   time.Time `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, m Message) (*Message, error)
	GetByProject(ctx context.Context, projectID int64) ([]Message, error)
	GetDirectMessages(ctx context.Context, userA, userB int64) ([]Message, error)
}
