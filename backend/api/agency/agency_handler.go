package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend/api"
	"backend/internal/apperrors"
	domain "backend/internal/domain/agency"
	services "backend/internal/services/agency"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AgencyHandler struct {
	svc *services.AgencyService
}

func NewAgencyHandler(svc *services.AgencyService) *AgencyHandler {
	return &AgencyHandler{svc: svc}
}

type createAgencyRequest struct {
	Name string `json:"name"`
}

// POST /agencies
func (h *AgencyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAgencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		api.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	agency, err := h.svc.Create(r.Context(), req.Name)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to create agency")
		return
	}
	api.WriteJSON(w, http.StatusCreated, agency)
}

// GET /agencies/{id}
func (h *AgencyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid agency id")
		return
	}
	agency, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "agency not found")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "failed to get agency")
		return
	}
	api.WriteJSON(w, http.StatusOK, agency)
}

// GET /agencies
func (h *AgencyHandler) List(w http.ResponseWriter, r *http.Request) {
	agencies, err := h.svc.List(r.Context())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to list agencies")
		return
	}
	if agencies == nil {
		agencies = make([]*domain.Agency, 0)
	}
	api.WriteJSON(w, http.StatusOK, agencies)
}
