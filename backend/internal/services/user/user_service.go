package services

import (
	"context"
	"time"

	domain "backend/internal/domain/user"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type UserService struct {
	repo repositories.UserRepository
}

func NewUserService(repo repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, name, email, role string, agencyID uuid.UUID) (*domain.User, error) {
	user := &domain.User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Role:      role,
		AgencyID:  agencyID,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) ListByAgency(ctx context.Context, agencyID uuid.UUID) ([]*domain.User, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var result []*domain.User
	for _, u := range all {
		if u.AgencyID == agencyID {
			result = append(result, u)
		}
	}
	return result, nil
}
