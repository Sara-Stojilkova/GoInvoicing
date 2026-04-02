package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/internal/apperrors"
	"backend/internal/services"

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
		writeError(w, http.StatusInternalServerError, "failed to list invoices")
		return
	}
	writeJSON(w, http.StatusOK, invoices)
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
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CustomerName == "" || req.Currency == "" || req.Amount <= 0 {
		writeError(w, http.StatusBadRequest, "customer_name, amount, and currency are required")
		return
	}

	inv, err := h.svc.CreateInvoice(r.Context(), req.CustomerName, req.Amount, req.Currency, req.DueDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create invoice")
		return
	}
	writeJSON(w, http.StatusCreated, inv)
}

// POST /invoices/{id}/pay
func (h *InvoiceHandler) Pay(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	if err := h.svc.MarkAsPaid(r.Context(), id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			writeError(w, http.StatusNotFound, "invoice not found")
			return
		}
		if errors.Is(err, apperrors.ErrConflict) {
			writeError(w, http.StatusConflict, "invoice already paid")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to mark invoice as paid")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
