package api

import (
	"net/http"

	services "backend/internal/services/task"
)

type TaskHandler struct {
	svc *services.TaskService
}

func NewTaskHandler(svc *services.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *TaskHandler) Assign(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func (h *TaskHandler) Complete(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}
