package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authAPI "backend/api/auth"
	agencyDomain "backend/internal/domain/agency"
	"backend/internal/repositories/memory"
	authServices "backend/internal/services/auth"

	"github.com/google/uuid"
)

func newHandler(t *testing.T, supabaseURL string) *authAPI.AuthHandler {
	t.Helper()
	svc := authServices.NewAuthService(supabaseURL, "anon-key", "svc-key", memory.NewAgencyRepo(), memory.NewUserRepo())
	return authAPI.NewAuthHandler(svc)
}

func fakeSupabase(t *testing.T, userID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/auth/v1/signup":
			fmt.Fprintf(w, `{"id":%q}`, userID)
		case r.Method == http.MethodPatch && r.URL.Path == "/auth/v1/admin/users/"+userID:
			fmt.Fprintf(w, `{"id":%q}`, userID)
		case r.Method == http.MethodPost && r.URL.Path == "/auth/v1/token":
			w.Write([]byte(`{"access_token":"tok","refresh_token":"ref","expires_in":3600,"token_type":"bearer"}`))
		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
		}
	}))
}

// --- Register ---

func TestAuthHandlerRegister_NewAgency(t *testing.T) {
	userID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	server := fakeSupabase(t, userID)
	defer server.Close()

	h := newHandler(t, server.URL)
	body := `{"full_name":"Jane Doe","email":"jane@example.com","password":"pass123","agency_name":"Acme"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["role"] != "admin" {
		t.Errorf("role: got %v, want admin", resp["role"])
	}
}

func TestAuthHandlerRegister_ExistingAgency(t *testing.T) {
	userID := "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	server := fakeSupabase(t, userID)
	defer server.Close()

	agencyRepo := memory.NewAgencyRepo()
	agencyUUID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	agencyRepo.Create(context.Background(), &agencyDomain.Agency{ID: agencyUUID, Name: "Existing Co"})

	svc := authServices.NewAuthService(server.URL, "anon-key", "svc-key", agencyRepo, memory.NewUserRepo())
	h := authAPI.NewAuthHandler(svc)

	body := fmt.Sprintf(`{"full_name":"Bob","email":"bob@example.com","password":"pass","agency_id":%q}`, agencyUUID.String())
	r := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["role"] != "member" {
		t.Errorf("role: got %v, want member", resp["role"])
	}
}

func TestAuthHandlerRegister_AgencyNotFound(t *testing.T) {
	h := newHandler(t, "http://unused")
	nonExistent := "dddddddd-dddd-dddd-dddd-dddddddddddd"
	body := fmt.Sprintf(`{"full_name":"Jane","email":"j@example.com","password":"pass","agency_id":%q}`, nonExistent)
	r := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestAuthHandlerRegister_BothAgencyFields(t *testing.T) {
	h := newHandler(t, "http://unused")
	body := `{"full_name":"Jane","email":"j@example.com","password":"p","agency_name":"A","agency_id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAuthHandlerRegister_NeitherAgencyField(t *testing.T) {
	h := newHandler(t, "http://unused")
	body := `{"full_name":"Jane","email":"j@example.com","password":"p"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAuthHandlerRegister_MissingRequiredFields(t *testing.T) {
	h := newHandler(t, "http://unused")
	// agency_name present but email missing
	body := `{"full_name":"Jane","password":"pass","agency_name":"Acme"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAuthHandlerRegister_MalformedJSON(t *testing.T) {
	h := newHandler(t, "http://unused")
	r := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(`{bad`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Register(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// --- Login ---

func TestAuthHandlerLogin_Success(t *testing.T) {
	server := fakeSupabase(t, "cccccccc-cccc-cccc-cccc-cccccccccccc")
	defer server.Close()

	h := newHandler(t, server.URL)
	body := `{"email":"user@example.com","password":"pass"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Login(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp["access_token"] != "tok" {
		t.Errorf("access_token: got %v, want tok", resp["access_token"])
	}
}

func TestAuthHandlerLogin_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	h := newHandler(t, server.URL)
	body := `{"email":"user@example.com","password":"wrong"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Login(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthHandlerLogin_MissingFields(t *testing.T) {
	h := newHandler(t, "http://unused")
	body := `{"email":"user@example.com"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Login(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAuthHandlerLogin_MalformedJSON(t *testing.T) {
	h := newHandler(t, "http://unused")
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{bad`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Login(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}
