package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/apperrors"
	"backend/internal/repositories/postgres"
	"backend/internal/services"

	"github.com/google/uuid"
)

var (
	ctx = context.Background()
	now = time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
)

func newService() *services.InvoiceService {
	return services.NewInvoiceService(postgres.NewInvoiceRepo())
}

func TestCreateInvoice(t *testing.T) {
	tests := []struct {
		name         string
		customerName string
		amount       float64
		currency     string
		dueDate      time.Time
		wantStatus   string
	}{
		{"draft status on creation", "Acme Corp", 500.00, "USD", now.Add(48 * time.Hour), "draft"},
		{"draft even when due date is past", "Old Co", 100.00, "EUR", now.Add(-24 * time.Hour), "draft"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			inv, err := svc.CreateInvoice(ctx, tt.customerName, tt.amount, tt.currency, tt.dueDate)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if inv.CustomerName != tt.customerName {
				t.Errorf("CustomerName = %q, want %q", inv.CustomerName, tt.customerName)
			}
			if inv.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", inv.Status, tt.wantStatus)
			}
		})
	}
}

func TestMarkAsPaid(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(svc *services.InvoiceService) uuid.UUID
		wantErr error
	}{
		{
			name: "success",
			setup: func(svc *services.InvoiceService) uuid.UUID {
				inv, _ := svc.CreateInvoice(ctx, "Acme Corp", 500.00, "USD", now.Add(48*time.Hour))
				return inv.ID
			},
			wantErr: nil,
		},
		{
			name: "not found",
			setup: func(svc *services.InvoiceService) uuid.UUID {
				inv, _ := svc.CreateInvoice(ctx, "Acme Corp", 500.00, "USD", now.Add(48*time.Hour))
				id := inv.ID
				id[0] ^= 0xFF
				return id
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "already paid",
			setup: func(svc *services.InvoiceService) uuid.UUID {
				inv, _ := svc.CreateInvoice(ctx, "Acme Corp", 500.00, "USD", now.Add(48*time.Hour))
				svc.MarkAsPaid(ctx, inv.ID)
				return inv.ID
			},
			wantErr: apperrors.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			id := tt.setup(svc)
			err := svc.MarkAsPaid(ctx, id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("MarkAsPaid() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestListOverdue(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc *services.InvoiceService)
		wantCount int
	}{
		{
			name: "unpaid past due appears",
			setup: func(svc *services.InvoiceService) {
				svc.CreateInvoice(ctx, "Overdue Co", 100.00, "USD", now.Add(-48*time.Hour))
				svc.CreateInvoice(ctx, "Future Co", 200.00, "USD", now.Add(48*time.Hour))
			},
			wantCount: 1,
		},
		{
			name: "paid invoice is not overdue",
			setup: func(svc *services.InvoiceService) {
				inv, _ := svc.CreateInvoice(ctx, "Paid Co", 100.00, "USD", now.Add(-48*time.Hour))
				svc.MarkAsPaid(ctx, inv.ID)
			},
			wantCount: 0,
		},
		{
			name:      "empty repo",
			setup:     func(svc *services.InvoiceService) {},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			tt.setup(svc)
			overdue, err := svc.ListOverdue(ctx, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(overdue) != tt.wantCount {
				t.Errorf("len(overdue) = %d, want %d", len(overdue), tt.wantCount)
			}
		})
	}
}

func TestGetSummary(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(svc *services.InvoiceService)
		wantPaid    int
		wantOverdue int
		wantUnpaid  int
		wantTotal   float64
	}{
		{
			name: "mixed invoices",
			setup: func(svc *services.InvoiceService) {
				inv1, _ := svc.CreateInvoice(ctx, "Paid Co", 100.00, "USD", now.Add(-48*time.Hour))
				svc.CreateInvoice(ctx, "Overdue Co", 200.00, "USD", now.Add(-24*time.Hour))
				svc.CreateInvoice(ctx, "Future Co", 300.00, "USD", now.Add(48*time.Hour))
				svc.MarkAsPaid(ctx, inv1.ID)
			},
			wantPaid: 1, wantOverdue: 1, wantUnpaid: 2, wantTotal: 600.00,
		},
		{
			name:      "empty repo",
			setup:     func(svc *services.InvoiceService) {},
			wantPaid:  0, wantOverdue: 0, wantUnpaid: 0, wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			tt.setup(svc)
			summary, err := svc.GetSummary(ctx, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if summary.PaidCount != tt.wantPaid {
				t.Errorf("PaidCount = %d, want %d", summary.PaidCount, tt.wantPaid)
			}
			if summary.OverdueCount != tt.wantOverdue {
				t.Errorf("OverdueCount = %d, want %d", summary.OverdueCount, tt.wantOverdue)
			}
			if summary.UnpaidCount != tt.wantUnpaid {
				t.Errorf("UnpaidCount = %d, want %d", summary.UnpaidCount, tt.wantUnpaid)
			}
			if summary.TotalAmount != tt.wantTotal {
				t.Errorf("TotalAmount = %f, want %f", summary.TotalAmount, tt.wantTotal)
			}
		})
	}
}
