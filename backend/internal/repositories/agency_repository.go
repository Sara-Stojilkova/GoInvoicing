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
	// Delete soft-deletes an agency (sets deleted_at). Used for cleanup on
	// partial registration failures.
	Delete(ctx context.Context, id uuid.UUID) error
}
