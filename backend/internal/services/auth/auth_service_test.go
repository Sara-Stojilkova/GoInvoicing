package services_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/apperrors"
	"backend/internal/repositories/memory"
	services "backend/internal/services/auth"
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
