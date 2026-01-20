package projects

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type projectUserEntry struct {
	ProjectID int64
	UserID    int64
	Role      string
}

type FakeRepository struct {
	projects     []Project
	projectUsers []projectUserEntry
	nextID       int64
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		nextID:       1,
		projects:     []Project{},
		projectUsers: []projectUserEntry{},
	}
}

func (f *FakeRepository) Create(ctx context.Context, p Project) (*Project, error) {
	p.ID = f.nextID
	f.nextID++
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	f.projects = append(f.projects, p)
	return &p, nil
}

func (f *FakeRepository) GetAllByOwner(ctx context.Context, ownerID int64) ([]Project, error) {
	var res []Project
	for _, p := range f.projects {
		if p.OwnerID == ownerID {
			res = append(res, p)
		}
	}
	return res, nil
}

func (f *FakeRepository) GetByID(ctx context.Context, id int64) (*Project, error) {
	for _, p := range f.projects {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.New("project not found")
}

func (f *FakeRepository) DeleteProject(ctx context.Context, id int64, ownerID int64) error {
	for i, p := range f.projects {
		if p.ID == id && p.OwnerID == ownerID {
			// Remove the project from the slice
			f.projects = append(f.projects[:i], f.projects[i+1:]...)
			return nil
		}
	}
	return errors.New("project not found")
}

func (f *FakeRepository) AddUserToProject(ctx context.Context, projectID int64, userID int64, role string) error {
	// Optional: Check if project exists first to be realistic
	_, err := f.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	f.projectUsers = append(f.projectUsers, projectUserEntry{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	})
	return nil
}

func (f *FakeRepository) ListUsers(ctx context.Context, projectID int64) ([]sqlc.ListUsersInProjectRow, error) {
	var res []sqlc.ListUsersInProjectRow
	for _, pu := range f.projectUsers {
		if pu.ProjectID == projectID {
			res = append(res, sqlc.ListUsersInProjectRow{
				ID:   pu.UserID,
				Role: pgtype.Text{String: pu.Role, Valid: true},
			})
		}
	}
	return res, nil
}

func (f *FakeRepository) RemoveUserFromProject(ctx context.Context, projectID int64, userID int64) error {
	for i, pu := range f.projectUsers {
		if pu.ProjectID == projectID && pu.UserID == userID {
			f.projectUsers = append(f.projectUsers[:i], f.projectUsers[i+1:]...)
			return nil
		}
	}
	return errors.New("user not found in project")
}
