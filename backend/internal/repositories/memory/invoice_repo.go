package memory

import (
	"context"
	"fmt"
	"sync"

	"backend/internal/apperrors"
	"backend/internal/domain"

	"github.com/google/uuid"
)

type InvoiceRepo struct {
	mu       sync.RWMutex
	invoices map[uuid.UUID]*domain.Invoice
}

func NewInvoiceRepo() *InvoiceRepo {
	return &InvoiceRepo{invoices: make(map[uuid.UUID]*domain.Invoice)}
}

func (r *InvoiceRepo) Create(ctx context.Context, invoice *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.invoices[invoice.ID] = invoice
	return nil
}

func (r *InvoiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inv, ok := r.invoices[id]
	if !ok {
		return nil, fmt.Errorf("invoice %s: %w", id, apperrors.ErrNotFound)
	}
	return inv, nil
}

func (r *InvoiceRepo) List(ctx context.Context) ([]*domain.Invoice, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Invoice, 0, len(r.invoices))
	for _, inv := range r.invoices {
		result = append(result, inv)
	}
	return result, nil
}

func (r *InvoiceRepo) Update(ctx context.Context, invoice *domain.Invoice) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.invoices[invoice.ID]; !ok {
		return fmt.Errorf("invoice %s: %w", invoice.ID, apperrors.ErrNotFound)
	}
	r.invoices[invoice.ID] = invoice
	return nil
}
