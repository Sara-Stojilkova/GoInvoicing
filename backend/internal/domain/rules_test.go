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
	paidAt := timePtr(now.Add(-24 * time.Hour))

	tests := []struct {
		name    string
		invoice Invoice
		want    string
	}{
		{"paid invoice", Invoice{ID: uuid.New(), DueDate: past, PaidAt: paidAt, Status: "sent"}, "paid"},
		{"paid invoice with future due date", Invoice{ID: uuid.New(), DueDate: future, PaidAt: paidAt, Status: "sent"}, "paid"},
		{"unpaid, past due date — overdue", Invoice{ID: uuid.New(), DueDate: past, Status: "sent"}, "overdue"},
		{"unpaid, due one second ago — overdue", Invoice{ID: uuid.New(), DueDate: now.Add(-time.Second), Status: "sent"}, "overdue"},
		{"draft invoice, future due date", Invoice{ID: uuid.New(), DueDate: future, Status: "draft"}, "draft"},
		{"sent invoice, future due date", Invoice{ID: uuid.New(), DueDate: future, Status: "sent"}, "sent"},
		{"draft invoice, past due date — overdue", Invoice{ID: uuid.New(), DueDate: past, Status: "draft"}, "overdue"},
		{"unpaid, due exactly now — not overdue", Invoice{ID: uuid.New(), DueDate: now, Status: "sent"}, "sent"},
		{"paid preserves paid over overdue", Invoice{ID: uuid.New(), DueDate: past, PaidAt: paidAt, Status: "overdue"}, "paid"},
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
