package projects

import "context"

type Repository interface {
	Create(ctx context.Context, p Project) (*Project, error)
	GetAllByOwner(ctx context.Context, ownerID int64) ([]Project, error)
	GetByID(ctx context.Context, id int64) (*Project, error)
}
