package services

import (
	"context"

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
	panic("not implemented")
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	panic("not implemented")
}

func (s *UserService) ListByAgency(ctx context.Context, agencyID uuid.UUID) ([]*domain.User, error) {
	panic("not implemented")
}
