package user

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) RegisterUser(
	ctx context.Context,
	email string,
	hashedPassword string,
) (User, error) {

	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing.ID != 0 {
		return User{}, ErrUserAlreadyExists
	}

	u := User{
		Email:    email,
		Password: hashedPassword,
	}

	return s.repo.Create(ctx, u)
}
