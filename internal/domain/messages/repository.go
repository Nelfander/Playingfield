package messages

import "context"

type Repository interface {
	Create(ctx context.Context, m Message) (*Message, error)
	GetByProject(ctx context.Context, projectID int64) ([]Message, error)
	GetDirectMessages(ctx context.Context, userA, userB int64) ([]Message, error)
}
