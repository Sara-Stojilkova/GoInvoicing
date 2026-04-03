package repositories

import (
	"context"

	domain "backend/internal/domain/agency"

	"github.com/google/uuid"
)

type AgencyRepository interface {
	Create(ctx context.Context, agency *domain.Agency) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Agency, error)
	List(ctx context.Context) ([]*domain.Agency, error)
}
