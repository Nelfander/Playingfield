package projects

import (
	"context"
	"errors"
	"time"
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
