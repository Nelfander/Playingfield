package projects

import (
	"context"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type Repository interface {
	Create(ctx context.Context, p Project) (*Project, error)
	GetAllByOwner(ctx context.Context, ownerID int64) ([]Project, error)
	GetByID(ctx context.Context, id int64) (*Project, error)
	DeleteProject(ctx context.Context, id int64, ownerID int64) error
	AddUserToProject(ctx context.Context, userID int64, projectID int64, role string) error
	RemoveUserFromProject(ctx context.Context, projectID int64, userID int64) error
	ListUsers(ctx context.Context, projectID int64) ([]sqlc.ListUsersInProjectRow, error)
}
