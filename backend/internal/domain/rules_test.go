package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEvaluateInvoiceStatus(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	past := now.Add(-48 * time.Hour)
	future := now.Add(48 * time.Hour)
	paidAt := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		invoice Invoice
		want    string
	}{
		{
			name:    "paid invoice",
			invoice: Invoice{ID: uuid.New(), DueDate: past, PaidAt: &paidAt, Status: "sent"},
			want:    "paid",
		},
		{
			name:    "paid invoice with future due date",
			invoice: Invoice{ID: uuid.New(), DueDate: future, PaidAt: &paidAt, Status: "sent"},
			want:    "paid",
		},
		{
			name:    "unpaid, past due date — overdue",
			invoice: Invoice{ID: uuid.New(), DueDate: past, Status: "sent"},
			want:    "overdue",
		},
		{
			name:    "unpaid, due one second ago — overdue",
			invoice: Invoice{ID: uuid.New(), DueDate: now.Add(-time.Second), Status: "sent"},
			want:    "overdue",
		},
		{
			name:    "draft invoice, future due date",
			invoice: Invoice{ID: uuid.New(), DueDate: future, Status: "draft"},
			want:    "draft",
		},
		{
			name:    "sent invoice, future due date",
			invoice: Invoice{ID: uuid.New(), DueDate: future, Status: "sent"},
			want:    "sent",
		},
		{
			name:    "draft invoice, past due date — overdue",
			invoice: Invoice{ID: uuid.New(), DueDate: past, Status: "draft"},
			want:    "overdue",
		},
		{
			name:    "unpaid, due exactly now — not overdue",
			invoice: Invoice{ID: uuid.New(), DueDate: now, Status: "sent"},
			want:    "sent",
		},
		{
			name:    "paid invoice preserves paid over overdue",
			invoice: Invoice{ID: uuid.New(), DueDate: past, PaidAt: &paidAt, Status: "overdue"},
			want:    "paid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateInvoiceStatus(tt.invoice, now)
			if got != tt.want {
				t.Errorf("EvaluateInvoiceStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
