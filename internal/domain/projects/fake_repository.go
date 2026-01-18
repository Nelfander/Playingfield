package projects

import (
	"context"
	"errors"
	"time"

	"github.com/nelfander/Playingfield/internal/infrastructure/postgres/sqlc"
)

type FakeRepository struct {
	projects []Project
	nextID   int64
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{nextID: 1}
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

func (f *FakeRepository) AddUserToProject(ctx context.Context, userID int64, projectID int64, role string) error {
	return nil
}

func (f *FakeRepository) RemoveUserFromProject(ctx context.Context, projectID int64, userID int64) error {
	// For now, we just return nil to satisfy the interface.
	// If you ever write a test for "Removing a User", you can add logic here.
	return nil
}
func (f *FakeRepository) ListUsers(ctx context.Context, projectID int64) ([]sqlc.ListUsersInProjectRow, error) {
	// Return an empty list for now so tests don't break
	return []sqlc.ListUsersInProjectRow{}, nil
}
