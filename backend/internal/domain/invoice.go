package domain

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID           uuid.UUID  `json:"id"`
	CustomerName string     `json:"customer_name"`
	Amount       float64    `json:"amount"`
	Currency     string     `json:"currency"`
	IssuedAt     time.Time  `json:"issued_at"`
	DueDate      time.Time  `json:"due_date"`
	PaidAt       *time.Time `json:"paid_at"` // nil = unpaid
	Status       string     `json:"status"`  // "draft", "sent", "paid", "overdue"
}

func (i Invoice) IsPaid() bool {
	return i.PaidAt != nil
}

func (i Invoice) IsOverdue(now time.Time) bool {
	return !i.IsPaid() && now.After(i.DueDate)
}

func (i Invoice) DaysUntilDue(now time.Time) int {
	return int(i.DueDate.Sub(now).Hours() / 24)
}
