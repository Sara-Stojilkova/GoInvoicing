package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"backend/api"
	"backend/internal/apperrors"
	domain "backend/internal/domain/task"
	services "backend/internal/services/task"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TaskHandler struct {
	svc *services.TaskService
}

func NewTaskHandler(svc *services.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

// GET /tasks?agency_id=<uuid>
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	agencyID, err := parseUUIDParam(r, "agency_id")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "agency_id query param is required and must be a valid UUID")
		return
	}
	tasks, err := h.svc.ListByAgency(r.Context(), agencyID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	if tasks == nil {
		tasks = make([]*domain.Task, 0)
	}
	api.WriteJSON(w, http.StatusOK, tasks)
}

type createTaskRequest struct {
	Title    string    `json:"title"`
	Priority string    `json:"priority"`
	AgencyID uuid.UUID `json:"agency_id"`
}

// POST /tasks
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.Priority == "" {
		api.WriteError(w, http.StatusBadRequest, "title and priority are required")
		return
	}
	if req.AgencyID == uuid.Nil {
		api.WriteError(w, http.StatusBadRequest, "agency_id is required")
		return
	}
	task, err := h.svc.Create(r.Context(), req.Title, req.Priority, req.AgencyID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	api.WriteJSON(w, http.StatusCreated, task)
}

// GET /tasks/{id}?agency_id=<uuid>
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	agencyID, err := parseUUIDParam(r, "agency_id")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "agency_id query param is required and must be a valid UUID")
		return
	}
	task, err := h.svc.GetTask(r.Context(), taskID, agencyID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			api.WriteError(w, http.StatusForbidden, "access denied")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "failed to get task")
		return
	}
	api.WriteJSON(w, http.StatusOK, task)
}

type assignTaskRequest struct {
	AssigneeID       uuid.UUID `json:"assignee_id"`
	AssigneeAgencyID uuid.UUID `json:"assignee_agency_id"`
}

// POST /tasks/{id}/assign
func (h *TaskHandler) Assign(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var req assignTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AssigneeID == uuid.Nil {
		api.WriteError(w, http.StatusBadRequest, "assignee_id is required")
		return
	}
	if req.AssigneeAgencyID == uuid.Nil {
		api.WriteError(w, http.StatusBadRequest, "assignee_agency_id is required")
		return
	}
	if err := h.svc.AssignTask(r.Context(), taskID, req.AssigneeID, req.AssigneeAgencyID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, apperrors.ErrForbidden) {
			api.WriteError(w, http.StatusForbidden, "access denied")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "failed to assign task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /tasks/{id}/complete
func (h *TaskHandler) Complete(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	if err := h.svc.CompleteTask(r.Context(), taskID, time.Now()); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			api.WriteError(w, http.StatusNotFound, "task not found")
			return
		}
		if errors.Is(err, apperrors.ErrConflict) {
			api.WriteError(w, http.StatusConflict, "task already completed")
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "failed to complete task")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /tasks/{id}/set-in-progress
func (h *TaskHandler) SetInProgress(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return uuid.Nil, errors.New("missing")
	}
	return uuid.Parse(val)
}
