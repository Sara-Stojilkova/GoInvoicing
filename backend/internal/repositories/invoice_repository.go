package repositories

import (
	"context"

	"backend/internal/domain"

	"github.com/google/uuid"
)

type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error)
	List(ctx context.Context) ([]*domain.Invoice, error)
	Update(ctx context.Context, invoice *domain.Invoice) error
}
