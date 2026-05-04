package services_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/apperrors"
	agencyDomain "backend/internal/domain/agency"
	"backend/internal/repositories/memory"
	services "backend/internal/services/auth"

	"github.com/google/uuid"
)

func newSvc(t *testing.T, supabaseURL string) *services.AuthService {
	t.Helper()
	return services.NewAuthService(supabaseURL, "anon-key", "svc-key", memory.NewAgencyRepo(), memory.NewUserRepo())
}

func TestLogin_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/token" || r.URL.Query().Get("grant_type") != "password" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"tok123","refresh_token":"ref456","expires_in":3600,"token_type":"bearer"}`))
	}))
	defer server.Close()

	result, err := newSvc(t, server.URL).Login(context.Background(), "user@example.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "tok123" {
		t.Errorf("access_token: got %q, want %q", result.AccessToken, "tok123")
	}
	if result.RefreshToken != "ref456" {
		t.Errorf("refresh_token: got %q, want %q", result.RefreshToken, "ref456")
	}
	if result.ExpiresIn != 3600 {
		t.Errorf("expires_in: got %d, want 3600", result.ExpiresIn)
	}
	if result.TokenType != "bearer" {
		t.Errorf("token_type: got %q, want %q", result.TokenType, "bearer")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant","error_description":"Invalid login credentials"}`))
	}))
	defer server.Close()

	_, err := newSvc(t, server.URL).Login(context.Background(), "user@example.com", "wrong")
	if !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Errorf("got %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_UnauthorizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	_, err := newSvc(t, server.URL).Login(context.Background(), "user@example.com", "wrong")
	if !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Errorf("got %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// server never responds; context is cancelled before the call reaches it
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before the call

	_, err := newSvc(t, server.URL).Login(ctx, "user@example.com", "pass")
	if err == nil {
		t.Fatal("expected an error for cancelled context")
	}
	if errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Error("context cancellation should not map to ErrInvalidCredentials")
	}
}

func TestLogin_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_server_error"}`))
	}))
	defer server.Close()

	_, err := newSvc(t, server.URL).Login(context.Background(), "user@example.com", "pass")
	if err == nil {
		t.Fatal("expected an error for server error response")
	}
	if errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Error("server error should not map to ErrInvalidCredentials")
	}
}

func makeSupabaseServer(t *testing.T, userID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/auth/v1/signup":
			fmt.Fprintf(w, `{"id":%q,"email":"test@example.com"}`, userID)
		case r.Method == http.MethodPut && r.URL.Path == "/auth/v1/admin/users/"+userID:
			fmt.Fprintf(w, `{"id":%q}`, userID)
		case r.Method == http.MethodDelete && r.URL.Path == "/auth/v1/admin/users/"+userID:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
}

func TestRegister_NewAgency(t *testing.T) {
	userID := "11111111-1111-1111-1111-111111111111"
	server := makeSupabaseServer(t, userID)
	defer server.Close()

	agencyRepo := memory.NewAgencyRepo()
	userRepo := memory.NewUserRepo()
	svc := services.NewAuthService(server.URL, "anon-key", "svc-key", agencyRepo, userRepo)

	result, err := svc.Register(context.Background(), services.RegisterRequest{
		FullName:   "Jane Doe",
		Email:      "jane@example.com",
		Password:   "password123",
		AgencyName: "Acme Co",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != "admin" {
		t.Errorf("role: got %q, want %q", result.Role, "admin")
	}
	if !result.Activated {
		t.Error("want activated=true for first user of a new agency")
	}
	if result.UserID.String() != userID {
		t.Errorf("userID: got %v, want %v", result.UserID, userID)
	}
	if result.AgencyID == (uuid.UUID{}) {
		t.Error("want non-zero AgencyID")
	}
}

func TestRegister_ExistingAgency(t *testing.T) {
	userID := "22222222-2222-2222-2222-222222222222"
	server := makeSupabaseServer(t, userID)
	defer server.Close()

	agencyRepo := memory.NewAgencyRepo()
	userRepo := memory.NewUserRepo()

	agencyID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	agencyRepo.Create(context.Background(), &agencyDomain.Agency{ID: agencyID, Name: "Existing Co"})

	svc := services.NewAuthService(server.URL, "anon-key", "svc-key", agencyRepo, userRepo)
	result, err := svc.Register(context.Background(), services.RegisterRequest{
		FullName: "Bob Smith",
		Email:    "bob@example.com",
		Password: "password123",
		AgencyID: &agencyID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != "member" {
		t.Errorf("role: got %q, want %q", result.Role, "member")
	}
	if result.Activated {
		t.Error("want activated=false for member joining existing agency")
	}
	if result.AgencyID != agencyID {
		t.Errorf("agencyID: got %v, want %v", result.AgencyID, agencyID)
	}
}

func TestRegister_AgencyNotFound(t *testing.T) {
	agencyRepo := memory.NewAgencyRepo()
	userRepo := memory.NewUserRepo()
	svc := services.NewAuthService("http://unused", "anon-key", "svc-key", agencyRepo, userRepo)

	nonExistent := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	_, err := svc.Register(context.Background(), services.RegisterRequest{
		FullName: "Bob Smith",
		Email:    "bob@example.com",
		Password: "password123",
		AgencyID: &nonExistent,
	})
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/auth/v1/signup" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error":"user_already_exists"}`))
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	agencyRepo := memory.NewAgencyRepo()
	userRepo := memory.NewUserRepo()
	svc := services.NewAuthService(server.URL, "anon-key", "svc-key", agencyRepo, userRepo)

	_, err := svc.Register(context.Background(), services.RegisterRequest{
		FullName:   "Jane Doe",
		Email:      "jane@example.com",
		Password:   "password123",
		AgencyName: "Acme Co",
	})
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Errorf("got %v, want ErrConflict", err)
	}

	// Verify the newly-created agency was cleaned up
	agencies, listErr := agencyRepo.List(context.Background())
	if listErr != nil {
		t.Fatalf("list agencies: %v", listErr)
	}
	if len(agencies) != 0 {
		t.Errorf("agency cleanup: got %d agencies remaining, want 0", len(agencies))
	}
}

func TestRegister_CleanupOnSetAppMetadataFailure(t *testing.T) {
	userID := "55555555-5555-5555-5555-555555555555"
	deleteCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/auth/v1/signup":
			fmt.Fprintf(w, `{"id":%q,"email":"test@example.com"}`, userID)
		case r.Method == http.MethodPut && r.URL.Path == "/auth/v1/admin/users/"+userID:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/auth/v1/admin/users/"+userID:
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	agencyRepo := memory.NewAgencyRepo()
	userRepo := memory.NewUserRepo()
	svc := services.NewAuthService(server.URL, "anon-key", "svc-key", agencyRepo, userRepo)

	_, err := svc.Register(context.Background(), services.RegisterRequest{
		FullName:   "Jane Doe",
		Email:      "jane@example.com",
		Password:   "password123",
		AgencyName: "Acme Co",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !deleteCalled {
		t.Error("expected auth user DELETE to be called for cleanup")
	}

	agencies, listErr := agencyRepo.List(context.Background())
	if listErr != nil {
		t.Fatalf("list agencies: %v", listErr)
	}
	if len(agencies) != 0 {
		t.Errorf("agency cleanup: got %d agencies remaining, want 0", len(agencies))
	}
}
