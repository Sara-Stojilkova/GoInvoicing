package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "backend/internal/domain/agency"
	"backend/internal/repositories/memory"
	services "backend/internal/services/agency"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func newAgencyService() *services.AgencyService {
	return services.NewAgencyService(memory.NewAgencyRepo())
}

func withChiParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func mustCreateAgency(t *testing.T, svc *services.AgencyService) *domain.Agency {
	t.Helper()
	agency, err := svc.Create(context.Background(), "Acme Corp")
	if err != nil {
		t.Fatalf("setup Create: %v", err)
	}
	return agency
}

// --- Create ---

func TestAgencyHandlerCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid request",  `{"name":"Acme Corp"}`, http.StatusCreated},
		{"malformed json", `{bad json}`,           http.StatusBadRequest},
		{"missing name",   `{}`,                   http.StatusBadRequest},
		{"empty name",     `{"name":""}`,           http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAgencyHandler(newAgencyService())

			r := httptest.NewRequest(http.MethodPost, "/api/agencies", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Create(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusCreated {
				var agency domain.Agency
				if err := json.Unmarshal(w.Body.Bytes(), &agency); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if agency.ID == uuid.Nil {
					t.Error("response agency has zero ID")
				}
				if agency.Name != "Acme Corp" {
					t.Errorf("Name = %q, want %q", agency.Name, "Acme Corp")
				}
			}
		})
	}
}

// --- Get ---

func TestAgencyHandlerGet(t *testing.T) {
	tests := []struct {
		name       string
		idStr      func(*services.AgencyService) string
		wantStatus int
	}{
		{
			name:       "invalid uuid",
			idStr:      func(*services.AgencyService) string { return "not-a-uuid" },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			idStr:      func(*services.AgencyService) string { return uuid.New().String() },
			wantStatus: http.StatusNotFound,
		},
		{
			name: "success",
			idStr: func(svc *services.AgencyService) string {
				return mustCreateAgency(t, svc).ID.String()
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newAgencyService()
			idStr := tt.idStr(svc)
			h := NewAgencyHandler(svc)

			r := httptest.NewRequest(http.MethodGet, "/api/agencies/"+idStr, nil)
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.Get(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var agency domain.Agency
				if err := json.Unmarshal(w.Body.Bytes(), &agency); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if agency.ID == uuid.Nil {
					t.Error("response agency has zero ID")
				}
			}
		})
	}
}

// --- List ---

func TestAgencyHandlerList(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*services.AgencyService)
		wantStatus int
		wantLen    int
	}{
		{
			name:       "empty returns empty array",
			setup:      nil,
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name: "returns all agencies",
			setup: func(svc *services.AgencyService) {
				svc.Create(context.Background(), "Acme Corp")
				svc.Create(context.Background(), "Globex")
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newAgencyService()
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewAgencyHandler(svc)

			r := httptest.NewRequest(http.MethodGet, "/api/agencies", nil)
			w := httptest.NewRecorder()
			h.List(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			var got []*domain.Agency
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}
			if len(got) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}
