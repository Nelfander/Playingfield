package projects

import (
	"context"
	"time"
)

type ProjectMember struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type Project struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	OwnerName   string    `json:"owner_name"`
}

type Repository interface {
	CreateProject(ctx context.Context, p Project) (*Project, error)
	Update(ctx context.Context, p Project) (*Project, error)
	GetAllByOwner(ctx context.Context, ownerID int64) ([]Project, error)
	GetByID(ctx context.Context, id int64) (*Project, error)
	DeleteProject(ctx context.Context, id int64, ownerID int64) error
	AddUserToProject(ctx context.Context, projectID int64, userID int64, role string) error
	RemoveUserFromProject(ctx context.Context, projectID int64, userID int64) error
	ListUsersInProject(ctx context.Context, projectID int64) ([]ProjectMember, error)
	UsersShareProject(ctx context.Context, userA, userB int64) (bool, error)
}
