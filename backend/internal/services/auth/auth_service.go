package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"backend/internal/apperrors"
	"backend/internal/repositories"
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
