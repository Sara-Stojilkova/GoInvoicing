package domain

import (
	"math"
	"time"
)

func EvaluateInvoiceStatus(invoice Invoice, now time.Time) string {
	if invoice.IsPaid() {
		return "paid"
	}
	if invoice.IsOverdue(now) {
		return "overdue"
	}
	return invoice.Status
}

func CalculateLateFee(amount float64, daysLate int, rate float64) float64 {
	if daysLate <= 0 {
		return 0
	}
	compounded := amount * math.Pow(1+rate, float64(daysLate))
	return math.Round((compounded-amount)*100) / 100
}

type Summary struct {
	PaidCount    int
	UnpaidCount  int
	OverdueCount int
	TotalAmount  float64
}

func SummarizeInvoices(invoices []Invoice, now time.Time) Summary {
	var s Summary
	for _, inv := range invoices {
		s.TotalAmount += inv.Amount
		status := EvaluateInvoiceStatus(inv, now)
		switch status {
		case "paid":
			s.PaidCount++
		case "overdue":
			s.OverdueCount++
			s.UnpaidCount++
		default:
			s.UnpaidCount++
		}
	}
	return s
}
