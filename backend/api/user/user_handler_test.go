package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domain "backend/internal/domain/user"
	"backend/internal/repositories/memory"
	services "backend/internal/services/user"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func newUserService() *services.UserService {
	return services.NewUserService(memory.NewUserRepo())
}

func withChiParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func mustCreateUser(t *testing.T, svc *services.UserService, agencyID uuid.UUID) *domain.User {
	t.Helper()
	user, err := svc.Create(context.Background(), "Alice", "alice@example.com", "member", agencyID)
	if err != nil {
		t.Fatalf("setup Create: %v", err)
	}
	return user
}

// --- Create ---

func TestUserHandlerCreate(t *testing.T) {
	agencyID := uuid.New()
	validBody := fmt.Sprintf(`{"name":"Alice","email":"alice@example.com","role":"member","agency_id":%q}`, agencyID)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid request",          validBody,                                                                              http.StatusCreated},
		{"malformed json",         `{bad json}`,                                                                           http.StatusBadRequest},
		{"missing name",           fmt.Sprintf(`{"email":"alice@example.com","role":"member","agency_id":%q}`, agencyID), http.StatusBadRequest},
		{"missing email",          fmt.Sprintf(`{"name":"Alice","role":"member","agency_id":%q}`, agencyID),              http.StatusBadRequest},
		{"missing role",           fmt.Sprintf(`{"name":"Alice","email":"alice@example.com","agency_id":%q}`, agencyID),  http.StatusBadRequest},
		{"missing agency_id",      `{"name":"Alice","email":"alice@example.com","role":"member"}`,                         http.StatusBadRequest},
		{"invalid agency_id uuid", `{"name":"Alice","email":"alice@example.com","role":"member","agency_id":"bad"}`,       http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewUserHandler(newUserService())

			r := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Create(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusCreated {
				var user domain.User
				if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if user.ID == uuid.Nil {
					t.Error("response user has zero ID")
				}
				if user.AgencyID != agencyID {
					t.Errorf("AgencyID = %v, want %v", user.AgencyID, agencyID)
				}
			}
		})
	}
}

// --- Get ---

func TestUserHandlerGet(t *testing.T) {
	agencyID := uuid.New()

	tests := []struct {
		name       string
		idStr      func(*services.UserService) string
		wantStatus int
	}{
		{
			name:       "invalid uuid",
			idStr:      func(*services.UserService) string { return "not-a-uuid" },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			idStr:      func(*services.UserService) string { return uuid.New().String() },
			wantStatus: http.StatusNotFound,
		},
		{
			name: "success",
			idStr: func(svc *services.UserService) string {
				return mustCreateUser(t, svc, agencyID).ID.String()
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserService()
			idStr := tt.idStr(svc)
			h := NewUserHandler(svc)

			r := httptest.NewRequest(http.MethodGet, "/api/users/"+idStr, nil)
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.Get(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var user domain.User
				if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if user.ID == uuid.Nil {
					t.Error("response user has zero ID")
				}
			}
		})
	}
}

// --- List by agency ---

func TestUserHandlerListByAgency(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()

	tests := []struct {
		name       string
		agencyID   string
		setup      func(*services.UserService)
		wantStatus int
		wantLen    int
	}{
		{
			name:       "missing agency_id",
			agencyID:   "",
			setup:      nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid agency_id uuid",
			agencyID:   "not-a-uuid",
			setup:      nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty agency returns empty array",
			agencyID:   agencyA.String(),
			setup:      nil,
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name:     "returns only users from requested agency",
			agencyID: agencyA.String(),
			setup: func(svc *services.UserService) {
				mustCreateUser(t, svc, agencyA)
				mustCreateUser(t, svc, agencyA)
				mustCreateUser(t, svc, agencyB)
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserService()
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewUserHandler(svc)

			url := "/api/users"
			if tt.agencyID != "" {
				url += "?agency_id=" + tt.agencyID
			}
			r := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			h.ListByAgency(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var got []*domain.User
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if len(got) != tt.wantLen {
					t.Errorf("len = %d, want %d", len(got), tt.wantLen)
				}
			}
		})
	}
}
