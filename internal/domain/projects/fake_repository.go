package projects

import (
	"context"
	"errors"
	"fmt"
	"time"
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

func (f *FakeRepository) CreateProject(ctx context.Context, p Project) (*Project, error) {
	p.ID = f.nextID
	f.nextID++
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	f.projects = append(f.projects, p)

	// return a pointer to the version actually stored in the slice
	return &f.projects[len(f.projects)-1], nil
}

func (f *FakeRepository) Update(ctx context.Context, p Project) (*Project, error) {
	//  find the existing project by id
	for i, proj := range f.projects {
		if proj.ID == p.ID {
			// update the record in the slice
			f.projects[i] = p
			return &f.projects[i], nil
		}
	}
	return nil, fmt.Errorf("project not found in fake repo")
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
	for i := range f.projects {
		if f.projects[i].ID == id {
			// return a pointer to the actual element in the slice
			return &f.projects[i], nil
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

func (f *FakeRepository) ListUsersInProject(ctx context.Context, projectID int64) ([]ProjectMember, error) {
	var res []ProjectMember
	for _, pu := range f.projectUsers {
		if pu.ProjectID == projectID {
			// mapping to clean domain struct
			res = append(res, ProjectMember{
				ID:    pu.UserID,
				Email: "fake@example.com",
				Role:  pu.Role,
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

func (f *FakeRepository) UsersShareProject(ctx context.Context, userA, userB int64) (bool, error) {
	// Track which projects each user belongs to
	userAProjects := make(map[int64]bool)

	for _, pu := range f.projectUsers {
		if pu.UserID == userA {
			userAProjects[pu.ProjectID] = true
		}
	}

	for _, pu := range f.projectUsers {
		if pu.UserID == userB && userAProjects[pu.ProjectID] {
			return true, nil
		}
	}

	return false, nil
}
