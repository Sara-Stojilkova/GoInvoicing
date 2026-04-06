package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/api"
	"backend/internal/apperrors"
	services "backend/internal/services/invoice"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type InvoiceHandler struct {
	svc *services.InvoiceService
}

func NewInvoiceHandler(svc *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

// GET /invoices
func (h *InvoiceHandler) List(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.svc.List(r.Context())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to list invoices")
		return
	}
	api.WriteJSON(w, http.StatusOK, invoices)
}

type createInvoiceRequest struct {
	CustomerName string    `json:"customer_name"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	DueDate      time.Time `json:"due_date"`
}

// POST /invoices
func (h *InvoiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CustomerName == "" || req.Currency == "" || req.Amount <= 0 {
		api.WriteError(w, http.StatusBadRequest, "customer_name, amount, and currency are required")
		return
	}

	inv, err := h.svc.CreateInvoice(r.Context(), req.CustomerName, req.Amount, req.Currency, req.DueDate)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to create invoice")
		return
	}
	api.WriteJSON(w, http.StatusCreated, inv)
}

// GET /api/invoices/summary
func (h *InvoiceHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetSummary(r.Context(), time.Now())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to get summary")
		return
	}
	api.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"paid_count":    summary.PaidCount,
		"unpaid_count":  summary.UnpaidCount,
		"overdue_count": summary.OverdueCount,
		"total_amount":  summary.TotalAmount,
	})
}

// POST /api/invoices/{id}/pay
func (h *InvoiceHandler) Pay(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	if err := h.svc.MarkAsPaid(r.Context(), id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "invoice not found")
			return
		}
		if errors.Is(err, apperrors.ErrConflict) {
			api.WriteError(w, http.StatusConflict, "invoice already paid")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "failed to mark invoice as paid")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
