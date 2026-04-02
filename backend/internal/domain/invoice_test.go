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

func TestDaysUntilDue(t *testing.T) {
	now := time.Date(2026, 3, 31, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		invoice Invoice
		want    int
	}{
		{"due in 2 days",  Invoice{DueDate: now.Add(48 * time.Hour)}, 2},
		{"due in 1 day",   Invoice{DueDate: now.Add(24 * time.Hour)}, 1},
		{"due today",      Invoice{DueDate: now}, 0},
		{"overdue 1 day",  Invoice{DueDate: now.Add(-24 * time.Hour)}, -1},
		{"overdue 7 days", Invoice{DueDate: now.Add(-7 * 24 * time.Hour)}, -7},
		{"due in 26 hours (round down)", Invoice{DueDate: now.Add(26 * time.Hour)}, 1},
		{"due in 25 hours (still 1 day)", Invoice{DueDate: now.Add(25 * time.Hour)}, 1},
		{"due in 23 hours (less than a day)", Invoice{DueDate: now.Add(23 * time.Hour)}, 0},
		{"overdue 26 hours (round up toward zero)", Invoice{DueDate: now.Add(-26 * time.Hour)}, -1},
		{"overdue 25 hours (still -1 day)", Invoice{DueDate: now.Add(-25 * time.Hour)}, -1},
		{"overdue 23 hours (less than a day)", Invoice{DueDate: now.Add(-23 * time.Hour)}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.invoice.DaysUntilDue(now)
			if got != tt.want {
				t.Errorf("DaysUntilDue() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateLateFee(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		daysLate int
		rate     float64
		want     float64
	}{
		{"zero days late",     100, 0,   0.05, 0},
		{"negative days late", 100, -1,  0.05, 0},
		{"1 day at 5%",        100, 1,   0.05, 5.00},
		{"2 days at 5%",       100, 2,   0.05, 10.25},
		{"30 days at 1%",      100, 30,  0.01, 34.78},
		{"1 day at 10%",       1000, 1,  0.10, 100.00},
		{"zero amount",        0,   10,  0.05, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateLateFee(tt.amount, tt.daysLate, tt.rate)
			if got != tt.want {
				t.Errorf("CalculateLateFee(%v, %v, %v) = %v, want %v", tt.amount, tt.daysLate, tt.rate, got, tt.want)
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
