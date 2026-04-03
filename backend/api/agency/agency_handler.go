package api

import (
	"net/http"

	services "backend/internal/services/agency"
)

type AgencyHandler struct {
	svc *services.AgencyService
}

func NewAgencyHandler(svc *services.AgencyService) *AgencyHandler {
	return &AgencyHandler{svc: svc}
}

func (h *AgencyHandler) Create(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *AgencyHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *AgencyHandler) List(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
