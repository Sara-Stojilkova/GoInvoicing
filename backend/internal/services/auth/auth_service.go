package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"backend/internal/apperrors"
	agencyDomain "backend/internal/domain/agency"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

// AuthService proxies registration and login operations to Supabase Auth.
type AuthService struct {
	supabaseURL    string
	anonKey        string
	serviceRoleKey string
	agencyRepo     repositories.AgencyRepository
	userRepo       repositories.UserRepository
}

// NewAuthService creates an AuthService that proxies auth operations to Supabase.
// supabaseURL is the base URL (no trailing slash). anonKey is used for user-facing
// calls; serviceRoleKey is used for admin calls that require elevated privileges.
func NewAuthService(supabaseURL, anonKey, serviceRoleKey string, agencyRepo repositories.AgencyRepository, userRepo repositories.UserRepository) *AuthService {
	return &AuthService{
		supabaseURL:    supabaseURL,
		anonKey:        anonKey,
		serviceRoleKey: serviceRoleKey,
		agencyRepo:     agencyRepo,
		userRepo:       userRepo,
	}
}

// LoginResult holds the token data returned by Supabase on successful login.
type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Login authenticates a user via Supabase Auth and returns a token pair.
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	url := s.supabaseURL + "/auth/v1/token?grant_type=password"
	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, fmt.Errorf("login: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.anonKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		return nil, apperrors.ErrInvalidCredentials
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login: unexpected status %d", resp.StatusCode)
	}

	var result LoginResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("login: decode response: %w", err)
	}
	return &result, nil
}

// RegisterRequest holds the inputs for creating a new account.
// Exactly one of AgencyID or AgencyName must be set (validated by the handler).
type RegisterRequest struct {
	FullName   string
	Email      string
	Password   string
	AgencyID   *uuid.UUID // nil = create a new agency
	AgencyName string     // used when AgencyID is nil
}

// RegisterResult holds the account details created during registration.
type RegisterResult struct {
	UserID    uuid.UUID
	AgencyID  uuid.UUID
	Role      string
	Activated bool
}

// supabaseUserResponse is the relevant subset of Supabase's signup/admin response.
type supabaseUserResponse struct {
	ID string `json:"id"`
}

// Register creates a Supabase Auth account and wires it to a public.users row.
// If AgencyName is set a new agency is created and the user becomes its admin.
// If AgencyID is set the user joins an existing agency as a member.
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	// Resolve agency
	var agencyID uuid.UUID
	var role string
	var newAgency bool

	if req.AgencyName != "" {
		agency := &agencyDomain.Agency{ID: uuid.New(), Name: req.AgencyName}
		if err := s.agencyRepo.Create(ctx, agency); err != nil {
			return nil, fmt.Errorf("register: create agency: %w", err)
		}
		agencyID = agency.ID
		role = "admin"
		newAgency = true
	} else {
		agencyID = *req.AgencyID
		if _, err := s.agencyRepo.GetByID(ctx, agencyID); err != nil {
			return nil, fmt.Errorf("register: %w", err)
		}
		role = "member"
	}

	// Supabase signup — the handle_new_user trigger creates public.users from
	// raw_user_meta_data.agency_id and raw_user_meta_data.full_name.
	signupResp, err := s.supabaseSignup(ctx, req.Email, req.Password, req.FullName, agencyID)
	if err != nil {
		if newAgency {
			if delErr := s.agencyRepo.Delete(ctx, agencyID); delErr != nil {
				log.Printf("register cleanup: delete agency %s: %v", agencyID, delErr)
			}
		}
		return nil, fmt.Errorf("register: signup: %w", err)
	}

	userID, err := uuid.Parse(signupResp.ID)
	if err != nil {
		return nil, fmt.Errorf("register: parse user id %q: %w", signupResp.ID, err)
	}

	// Set app_metadata so that jwt_agency_id() and auth_user_is_admin() work.
	if err := s.supabaseSetAppMetadata(ctx, signupResp.ID, agencyID, role); err != nil {
		log.Printf("register cleanup: set_app_metadata failed for user %s: %v", signupResp.ID, err)
		if delErr := s.supabaseDeleteUser(ctx, signupResp.ID); delErr != nil {
			log.Printf("register cleanup: delete auth user %s: %v", signupResp.ID, delErr)
		}
		if newAgency {
			if delErr := s.agencyRepo.Delete(ctx, agencyID); delErr != nil {
				log.Printf("register cleanup: delete agency %s: %v", agencyID, delErr)
			}
		}
		return nil, fmt.Errorf("register: set app_metadata: %w", err)
	}

	// Update public.users with email and, for new-agency admins, activated=true.
	if err := s.userRepo.UpdateSignupFields(ctx, userID, req.Email, newAgency); err != nil {
		log.Printf("register cleanup: UpdateSignupFields failed for user %s: %v", userID, err)
		if delErr := s.supabaseDeleteUser(ctx, signupResp.ID); delErr != nil {
			log.Printf("register cleanup: delete auth user %s: %v", signupResp.ID, delErr)
		}
		if newAgency {
			if delErr := s.agencyRepo.Delete(ctx, agencyID); delErr != nil {
				log.Printf("register cleanup: delete agency %s: %v", agencyID, delErr)
			}
		}
		return nil, fmt.Errorf("register: update user: %w", err)
	}

	return &RegisterResult{
		UserID:    userID,
		AgencyID:  agencyID,
		Role:      role,
		Activated: newAgency,
	}, nil
}

// supabaseSignup calls POST /auth/v1/signup with the anon key.
func (s *AuthService) supabaseSignup(ctx context.Context, email, password, fullName string, agencyID uuid.UUID) (*supabaseUserResponse, error) {
	url := s.supabaseURL + "/auth/v1/signup"
	body, err := json.Marshal(map[string]interface{}{
		"email":    email,
		"password": password,
		"data": map[string]interface{}{
			"full_name": fullName,
			"agency_id": agencyID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal signup request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("signup: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.anonKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("signup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnprocessableEntity {
		return nil, apperrors.ErrConflict
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var user supabaseUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("signup: decode response: %w", err)
	}
	return &user, nil
}

// supabaseSetAppMetadata calls PUT /auth/v1/admin/users/{id} with the service-role key
// to set app_metadata.agency_id and app_metadata.role in the user's JWT claims.
func (s *AuthService) supabaseSetAppMetadata(ctx context.Context, userID string, agencyID uuid.UUID, role string) error {
	url := s.supabaseURL + "/auth/v1/admin/users/" + userID
	body, err := json.Marshal(map[string]interface{}{
		"app_metadata": map[string]interface{}{
			"agency_id": agencyID.String(),
			"role":      role,
		},
	})
	if err != nil {
		return fmt.Errorf("marshal app_metadata request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("set app_metadata: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Admin endpoints authenticate via Authorization Bearer, not the apikey header.
	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("set app_metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// supabaseDeleteUser calls DELETE /auth/v1/admin/users/{id}. Used for cleanup only.
func (s *AuthService) supabaseDeleteUser(ctx context.Context, userID string) error {
	url := s.supabaseURL + "/auth/v1/admin/users/" + userID
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("delete auth user: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceRoleKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete auth user: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("delete auth user: unexpected status %d", resp.StatusCode)
	}
	return nil
}
