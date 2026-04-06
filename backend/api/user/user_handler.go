package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/api"
	"backend/internal/apperrors"
	domain "backend/internal/domain/user"
	services "backend/internal/services/user"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

type createUserRequest struct {
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	AgencyID uuid.UUID `json:"agency_id"`
}

// POST /users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" || req.Role == "" {
		api.WriteError(w, http.StatusBadRequest, "name, email, and role are required")
		return
	}
	if req.AgencyID == uuid.Nil {
		api.WriteError(w, http.StatusBadRequest, "agency_id is required")
		return
	}
	user, err := h.svc.Create(r.Context(), req.Name, req.Email, req.Role, req.AgencyID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	api.WriteJSON(w, http.StatusCreated, user)
}

// GET /users/{id}
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	user, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "user not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	api.WriteJSON(w, http.StatusOK, user)
}

// GET /users?agency_id=<uuid>
func (h *UserHandler) ListByAgency(w http.ResponseWriter, r *http.Request) {
	agencyIDStr := r.URL.Query().Get("agency_id")
	if agencyIDStr == "" {
		api.WriteError(w, http.StatusBadRequest, "agency_id query param is required")
		return
	}
	agencyID, err := uuid.Parse(agencyIDStr)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid agency_id")
		return
	}
	users, err := h.svc.ListByAgency(r.Context(), agencyID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	if users == nil {
		users = make([]*domain.User, 0)
	}
	api.WriteJSON(w, http.StatusOK, users)
}
