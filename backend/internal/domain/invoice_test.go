package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIsOverdue(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)
	past := now.Add(-48 * time.Hour)
	future := now.Add(48 * time.Hour)
	paidAt := now.Add(-24 * time.Hour)

	tests := []struct {
		name    string
		invoice Invoice
		want    bool
	}{
		{
			name:    "unpaid, future due date",
			invoice: Invoice{ID: uuid.New(), DueDate: future},
			want:    false,
		},
		{
			name:    "unpaid, past due date",
			invoice: Invoice{ID: uuid.New(), DueDate: past},
			want:    true,
		},
		{
			name:    "unpaid, due exactly now",
			invoice: Invoice{ID: uuid.New(), DueDate: now},
			want:    false, // now.After(now) is false
		},
		{
			name:    "unpaid, due one second ago",
			invoice: Invoice{ID: uuid.New(), DueDate: now.Add(-time.Second)},
			want:    true,
		},
		{
			name:    "paid, past due date",
			invoice: Invoice{ID: uuid.New(), DueDate: past, PaidAt: &paidAt},
			want:    false,
		},
		{
			name:    "paid, future due date",
			invoice: Invoice{ID: uuid.New(), DueDate: future, PaidAt: &paidAt},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.invoice.IsOverdue(now)
			if got != tt.want {
				t.Errorf("IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}
