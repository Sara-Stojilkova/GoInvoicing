package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSummarizeInvoices(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	past := now.Add(-48 * time.Hour)
	future := now.Add(48 * time.Hour)
	paidAt := timePtr(now.Add(-24 * time.Hour))

	paid := func(amount float64) Invoice {
		return Invoice{ID: uuid.New(), Amount: amount, DueDate: future, PaidAt: paidAt, Status: "paid"}
	}
	overdue := func(amount float64) Invoice {
		return Invoice{ID: uuid.New(), Amount: amount, DueDate: past, Status: "sent"}
	}
	draft := func(amount float64) Invoice {
		return Invoice{ID: uuid.New(), Amount: amount, DueDate: future, Status: "draft"}
	}
	sent := func(amount float64) Invoice {
		return Invoice{ID: uuid.New(), Amount: amount, DueDate: future, Status: "sent"}
	}

	tests := []struct {
		name     string
		invoices []Invoice
		want     Summary
	}{
		{
			"empty list",
			[]Invoice{},
			Summary{PaidCount: 0, UnpaidCount: 0, OverdueCount: 0, TotalAmount: 0},
		},
		{
			"all paid",
			[]Invoice{paid(100), paid(200)},
			Summary{PaidCount: 2, UnpaidCount: 0, OverdueCount: 0, TotalAmount: 300},
		},
		{
			"all overdue",
			[]Invoice{overdue(50), overdue(75)},
			Summary{PaidCount: 0, UnpaidCount: 2, OverdueCount: 2, TotalAmount: 125},
		},
		{
			"all unpaid non-overdue",
			[]Invoice{draft(100), sent(200)},
			Summary{PaidCount: 0, UnpaidCount: 2, OverdueCount: 0, TotalAmount: 300},
		},
		{
			"mixed statuses",
			[]Invoice{paid(100), overdue(200), draft(50), sent(150)},
			Summary{PaidCount: 1, UnpaidCount: 3, OverdueCount: 1, TotalAmount: 500},
		},
		{
			"total amount includes paid invoices",
			[]Invoice{paid(500), overdue(300)},
			Summary{PaidCount: 1, UnpaidCount: 1, OverdueCount: 1, TotalAmount: 800},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SummarizeInvoices(tt.invoices, now)
			if got != tt.want {
				t.Errorf("SummarizeInvoices() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

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
