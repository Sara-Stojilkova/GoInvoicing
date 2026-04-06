package services

import (
	"context"
	"time"

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
	agency := &domain.Agency{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, agency); err != nil {
		return nil, err
	}
	return agency, nil
}

func (s *AgencyService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agency, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AgencyService) List(ctx context.Context) ([]*domain.Agency, error) {
	return s.repo.List(ctx)
}
