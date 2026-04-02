package domain

import (
	"errors"
	"testing"
	"time"

	"backend/internal/apperrors"

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

func TestMarkAsPaid(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		invoice Invoice
		wantErr error
	}{
		{"unpaid draft",   Invoice{ID: uuid.New(), Status: "draft"}, nil},
		{"unpaid sent",    Invoice{ID: uuid.New(), Status: "sent"}, nil},
		{"already paid",   Invoice{ID: uuid.New(), Status: "paid", PaidAt: timePtr(now.Add(-24 * time.Hour))}, apperrors.ErrConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := tt.invoice
			err := inv.MarkAsPaid(now)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("MarkAsPaid() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("MarkAsPaid() unexpected error: %v", err)
			}
			if inv.PaidAt == nil {
				t.Error("MarkAsPaid() did not set PaidAt")
			}
			if !inv.PaidAt.Equal(now) {
				t.Errorf("MarkAsPaid() PaidAt = %v, want %v", inv.PaidAt, now)
			}
			if inv.Status != "paid" {
				t.Errorf("MarkAsPaid() Status = %q, want %q", inv.Status, "paid")
			}
		})
	}
}

func timePtr(t time.Time) *time.Time { return &t }
