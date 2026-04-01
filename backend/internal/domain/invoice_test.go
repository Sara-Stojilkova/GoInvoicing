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

	tests := []struct {
		name    string
		invoice Invoice
		want    bool
	}{
		{"unpaid, future due date", Invoice{ID: uuid.New(), DueDate: future}, false},
		{"unpaid, past due date", Invoice{ID: uuid.New(), DueDate: past}, true},
		{"unpaid, due exactly now", Invoice{ID: uuid.New(), DueDate: now}, false},
		{"unpaid, due one second ago", Invoice{ID: uuid.New(), DueDate: now.Add(-time.Second)}, true},
		{"paid, past due date", Invoice{ID: uuid.New(), DueDate: past, PaidAt: timePtr(now.Add(-24 * time.Hour))}, false},
		{"paid, future due date", Invoice{ID: uuid.New(), DueDate: future, PaidAt: timePtr(now.Add(-24 * time.Hour))}, false},
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

func timePtr(t time.Time) *time.Time { return &t }
