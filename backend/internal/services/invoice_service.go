package services

import (
	"context"
	"fmt"
	"time"

	"backend/internal/apperrors"
	"backend/internal/domain"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type InvoiceService struct {
	repo repositories.InvoiceRepository
}

func NewInvoiceService(repo repositories.InvoiceRepository) *InvoiceService {
	return &InvoiceService{repo: repo}
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, customerName string, amount float64, currency string, dueDate time.Time) (*domain.Invoice, error) {
	inv := &domain.Invoice{
		ID:           uuid.New(),
		CustomerName: customerName,
		Amount:       amount,
		Currency:     currency,
		IssuedAt:     time.Now(),
		DueDate:      dueDate,
		Status:       "draft",
	}
	if err := s.repo.Create(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *InvoiceService) MarkAsPaid(ctx context.Context, id uuid.UUID) error {
	inv, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("invoice %s: %w", id, apperrors.ErrNotFound)
	}
	if err := inv.MarkAsPaid(time.Now()); err != nil {
		return err
	}
	return s.repo.Update(ctx, inv)
}

func (s *InvoiceService) List(ctx context.Context) ([]*domain.Invoice, error) {
	return s.repo.List(ctx)
}

func (s *InvoiceService) ListOverdue(ctx context.Context, now time.Time) ([]*domain.Invoice, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var overdue []*domain.Invoice
	for _, inv := range all {
		if inv.IsOverdue(now) {
			overdue = append(overdue, inv)
		}
	}
	return overdue, nil
}

func (s *InvoiceService) GetSummary(ctx context.Context, now time.Time) (domain.Summary, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return domain.Summary{}, err
	}
	invoices := make([]domain.Invoice, len(all))
	for i, inv := range all {
		invoices[i] = *inv
	}
	return domain.SummarizeInvoices(invoices, now), nil
}
