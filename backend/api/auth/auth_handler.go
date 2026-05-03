package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/api"
	"backend/internal/apperrors"
	services "backend/internal/services/auth"

	"github.com/google/uuid"
)

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	FullName   string     `json:"full_name"`
	Email      string     `json:"email"`
	Password   string     `json:"password"`
	AgencyID   *uuid.UUID `json:"agency_id"`
	AgencyName string     `json:"agency_name"`
}

// POST /api/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FullName == "" || req.Email == "" || req.Password == "" {
		api.WriteError(w, http.StatusBadRequest, "full_name, email, and password are required")
		return
	}

	hasAgencyID := req.AgencyID != nil && *req.AgencyID != uuid.Nil
	hasAgencyName := req.AgencyName != ""
	if hasAgencyID == hasAgencyName {
		api.WriteError(w, http.StatusBadRequest, "provide either agency_id or agency_name, not both or neither")
		return
	}

	result, err := h.svc.Register(r.Context(), services.RegisterRequest{
		FullName:   req.FullName,
		Email:      req.Email,
		Password:   req.Password,
		AgencyID:   req.AgencyID,
		AgencyName: req.AgencyName,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "agency not found")
			return
		}
		if errors.Is(err, apperrors.ErrConflict) {
			api.WriteError(w, http.StatusConflict, "email already registered")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	api.WriteJSON(w, http.StatusCreated, result)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		api.WriteError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	result, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			api.WriteError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "login failed")
		return
	}

	api.WriteJSON(w, http.StatusOK, result)
}
