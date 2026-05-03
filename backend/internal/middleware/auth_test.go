package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const testSecret = "test-jwt-secret-for-unit-tests"

func makeToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return s
}

func validClaims(userID, agencyID uuid.UUID) jwt.MapClaims {
	return jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour).Unix(),
		"app_metadata": map[string]interface{}{
			"agency_id": agencyID.String(),
			"role":      "admin",
		},
	}
}

func TestAuthenticate_StatusCodes(t *testing.T) {
	userID   := uuid.New()
	agencyID := uuid.New()

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "valid token",
			authHeader: "Bearer " + makeToken(t, testSecret, validClaims(userID, agencyID)),
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong scheme",
			authHeader: "Basic abc123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "tampered signature",
			authHeader: "Bearer " + makeToken(t, "wrong-secret", validClaims(userID, agencyID)),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired token",
			authHeader: "Bearer " + makeToken(t, testSecret, jwt.MapClaims{
				"sub": userID.String(),
				"exp": time.Now().Add(-time.Hour).Unix(),
				"app_metadata": map[string]interface{}{
					"agency_id": agencyID.String(),
					"role":      "admin",
				},
			}),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing app_metadata",
			authHeader: "Bearer " + makeToken(t, testSecret, jwt.MapClaims{
				"sub": userID.String(),
				"exp": time.Now().Add(time.Hour).Unix(),
			}),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid agency_id in app_metadata",
			authHeader: "Bearer " + makeToken(t, testSecret, jwt.MapClaims{
				"sub": userID.String(),
				"exp": time.Now().Add(time.Hour).Unix(),
				"app_metadata": map[string]interface{}{
					"agency_id": "not-a-uuid",
					"role":      "admin",
				},
			}),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing role in app_metadata",
			authHeader: "Bearer " + makeToken(t, testSecret, jwt.MapClaims{
				"sub": userID.String(),
				"exp": time.Now().Add(time.Hour).Unix(),
				"app_metadata": map[string]interface{}{
					"agency_id": agencyID.String(),
					// no "role"
				},
			}),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "non-uuid sub",
			authHeader: "Bearer " + makeToken(t, testSecret, jwt.MapClaims{
				"sub": "not-a-uuid",
				"exp": time.Now().Add(time.Hour).Unix(),
				"app_metadata": map[string]interface{}{
					"agency_id": agencyID.String(),
					"role":      "admin",
				},
			}),
			wantStatus: http.StatusUnauthorized,
		},
	}

	handler := middleware.Authenticate(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthenticate_InjectsContext(t *testing.T) {
	userID   := uuid.New()
	agencyID := uuid.New()
	token    := makeToken(t, testSecret, validClaims(userID, agencyID))

	var gotUserID   uuid.UUID
	var gotAgencyID uuid.UUID
	var gotRole     string

	handler := middleware.Authenticate(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID   = r.Context().Value(middleware.ContextUserID).(uuid.UUID)
		gotAgencyID = r.Context().Value(middleware.ContextAgencyID).(uuid.UUID)
		gotRole     = r.Context().Value(middleware.ContextRole).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	if gotUserID != userID {
		t.Errorf("userID: got %v, want %v", gotUserID, userID)
	}
	if gotAgencyID != agencyID {
		t.Errorf("agencyID: got %v, want %v", gotAgencyID, agencyID)
	}
	if gotRole != "admin" {
		t.Errorf("role: got %q, want %q", gotRole, "admin")
	}
}
