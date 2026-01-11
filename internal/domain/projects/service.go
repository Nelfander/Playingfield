package projects

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, name, description string, ownerID int64) (*Project, error) {
	p := Project{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}
	return s.repo.Create(ctx, p)
}

func (s *Service) ListProjects(ctx context.Context, ownerID int64) ([]Project, error) {
	return s.repo.GetAllByOwner(ctx, ownerID)
}
