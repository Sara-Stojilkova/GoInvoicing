package memory

import (
	"context"
	"fmt"
	"sync"

	"backend/internal/apperrors"
	"backend/internal/domain"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type invoiceRepo struct {
	mu       sync.RWMutex
	invoices map[uuid.UUID]*domain.Invoice
}

func NewInvoiceRepo() repositories.InvoiceRepository {
	return &invoiceRepo{invoices: make(map[uuid.UUID]*domain.Invoice)}
}

func (r *invoiceRepo) Create(ctx context.Context, invoice *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *invoiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inv, ok := r.invoices[id]
	if !ok {
		return nil, fmt.Errorf("invoice %s: %w", id, apperrors.ErrNotFound)
	}
	return inv, nil
}

func (r *invoiceRepo) List(ctx context.Context) ([]*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Invoice, 0, len(r.invoices))
	for _, inv := range r.invoices {
		result = append(result, inv)
	}
	return result, nil
}

func (r *invoiceRepo) Update(ctx context.Context, invoice *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.invoices[invoice.ID]; !ok {
		return fmt.Errorf("invoice %s: %w", invoice.ID, apperrors.ErrNotFound)
	}
	r.invoices[invoice.ID] = invoice
	return nil
}
