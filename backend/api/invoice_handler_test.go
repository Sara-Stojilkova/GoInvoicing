package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "backend/internal/domain/invoice"
	"backend/internal/repositories/memory"
	services "backend/internal/services/invoice"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func newService() *services.InvoiceService {
	return services.NewInvoiceService(memory.NewInvoiceRepo())
}

func withChiParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func mustCreate(t *testing.T, svc *services.InvoiceService) *domain.Invoice {
	t.Helper()
	inv, err := svc.CreateInvoice(
		context.Background(), "Acme", 100, "USD",
		time.Now().Add(24*time.Hour),
	)
	if err != nil {
		t.Fatalf("setup CreateInvoice: %v", err)
	}
	return inv
}

func TestInvoiceHandlerList(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*services.InvoiceService)
		wantStatus int
		wantLen    int
	}{
		{"empty repo returns empty array", nil, http.StatusOK, 0},
		{"returns all created invoices", func(svc *services.InvoiceService) {
			mustCreate(t, svc)
			mustCreate(t, svc)
		}, http.StatusOK, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewInvoiceHandler(svc)

			r := httptest.NewRequest(http.MethodGet, "/api/invoices/", nil)
			w := httptest.NewRecorder()
			h.List(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			var got []*domain.Invoice
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestInvoiceHandlerCreate(t *testing.T) {
	validBody := `{"customer_name":"Acme","amount":100,"currency":"USD","due_date":"2026-12-01T00:00:00Z"}`

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid request", validBody, http.StatusCreated},
		{"malformed json", `{bad json}`, http.StatusBadRequest},
		{"missing customer", `{"amount":100,"currency":"USD"}`, http.StatusBadRequest},
		{"missing currency", `{"customer_name":"Acme","amount":100}`, http.StatusBadRequest},
		{"amount zero", `{"customer_name":"Acme","amount":0,"currency":"USD"}`, http.StatusBadRequest},
		{"amount negative", `{"customer_name":"Acme","amount":-5,"currency":"USD"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewInvoiceHandler(newService())

			r := httptest.NewRequest(http.MethodPost, "/api/invoices/", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Create(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusCreated {
				var inv domain.Invoice
				if err := json.Unmarshal(w.Body.Bytes(), &inv); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if inv.ID == uuid.Nil {
					t.Error("response invoice has zero ID")
				}
				if inv.CustomerName != "Acme" {
					t.Errorf("customer_name = %q, want %q", inv.CustomerName, "Acme")
				}
				if inv.Status != "draft" {
					t.Errorf("status = %q, want %q", inv.Status, "draft")
				}
			}
		})
	}
}

func TestInvoiceHandlerPay(t *testing.T) {
	tests := []struct {
		name       string
		idStr      func(*services.InvoiceService) string
		wantStatus int
	}{
		{
			"invalid uuid",
			func(*services.InvoiceService) string { return "not-a-uuid" },
			http.StatusBadRequest,
		},
		{
			"invoice not found",
			func(*services.InvoiceService) string { return uuid.New().String() },
			http.StatusNotFound,
		},
		{
			"success",
			func(svc *services.InvoiceService) string {
				return mustCreate(t, svc).ID.String()
			},
			http.StatusNoContent,
		},
		{
			"already paid",
			func(svc *services.InvoiceService) string {
				inv := mustCreate(t, svc)
				_ = svc.MarkAsPaid(context.Background(), inv.ID)
				return inv.ID.String()
			},
			http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			idStr := tt.idStr(svc)
			h := NewInvoiceHandler(svc)

			r := httptest.NewRequest(http.MethodPost, "/api/invoices/"+idStr+"/pay", nil)
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.Pay(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestInvoiceHandlerSummary(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*services.InvoiceService)
		wantStatus  int
		wantPaid    float64
		wantUnpaid  float64
		wantOverdue float64
	}{
		{
			"empty repo — all counts zero",
			nil,
			http.StatusOK,
			0, 0, 0,
		},
		{
			"two draft invoices — both unpaid",
			func(svc *services.InvoiceService) {
				mustCreate(t, svc)
				mustCreate(t, svc)
			},
			http.StatusOK,
			0, 2, 0,
		},
		{
			"one paid one unpaid",
			func(svc *services.InvoiceService) {
				inv := mustCreate(t, svc)
				_ = svc.MarkAsPaid(context.Background(), inv.ID)
				mustCreate(t, svc)
			},
			http.StatusOK,
			1, 1, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newService()
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewInvoiceHandler(svc)

			r := httptest.NewRequest(http.MethodGet, "/api/invoices/summary", nil)
			w := httptest.NewRecorder()
			h.Summary(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			var got map[string]float64
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			for _, key := range []string{"paid_count", "unpaid_count", "overdue_count", "total_amount"} {
				if _, ok := got[key]; !ok {
					t.Errorf("response missing key %q", key)
				}
			}
			if got["paid_count"] != tt.wantPaid {
				t.Errorf("paid_count = %v, want %v", got["paid_count"], tt.wantPaid)
			}
			if got["unpaid_count"] != tt.wantUnpaid {
				t.Errorf("unpaid_count = %v, want %v", got["unpaid_count"], tt.wantUnpaid)
			}
			if got["overdue_count"] != tt.wantOverdue {
				t.Errorf("overdue_count = %v, want %v", got["overdue_count"], tt.wantOverdue)
			}
		})
	}
}
