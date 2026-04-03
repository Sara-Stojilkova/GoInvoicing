package api

import (
	"net/http"

	services "backend/internal/services/user"
)

type UserHandler struct {
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *UserHandler) ListByAgency(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
