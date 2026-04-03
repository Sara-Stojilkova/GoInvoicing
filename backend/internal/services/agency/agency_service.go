package services

import (
	"context"

	domain "backend/internal/domain/agency"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type AgencyService struct {
	repo repositories.AgencyRepository
}

func NewAgencyService(repo repositories.AgencyRepository) *AgencyService {
	return &AgencyService{repo: repo}
}

func (s *AgencyService) Create(ctx context.Context, name string) (*domain.Agency, error) {
	panic("not implemented")
}

func (s *AgencyService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agency, error) {
	panic("not implemented")
}

func (s *AgencyService) List(ctx context.Context) ([]*domain.Agency, error) {
	panic("not implemented")
}
