package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "backend/internal/domain/task"
	"backend/internal/repositories/memory"
	services "backend/internal/services/task"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func newTaskService() *services.TaskService {
	return services.NewTaskService(memory.NewTaskRepo())
}

func withChiParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func mustCreateTask(t *testing.T, svc *services.TaskService, agencyID uuid.UUID) *domain.Task {
	t.Helper()
	task, err := svc.Create(context.Background(), "Fix bug", "high", agencyID, uuid.New(), nil, nil, nil)
	if err != nil {
		t.Fatalf("setup Create: %v", err)
	}
	return task
}

// --- List ---

func TestTaskHandlerList(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()

	tests := []struct {
		name       string
		agencyID   string
		setup      func(*services.TaskService)
		wantStatus int
		wantLen    int
	}{
		{
			name:       "missing agency_id",
			agencyID:   "",
			setup:      nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid agency_id uuid",
			agencyID:   "not-a-uuid",
			setup:      nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty agency returns empty array",
			agencyID:   agencyA.String(),
			setup:      nil,
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name:     "returns only tasks from requested agency",
			agencyID: agencyA.String(),
			setup: func(svc *services.TaskService) {
				mustCreateTask(t, svc, agencyA)
				mustCreateTask(t, svc, agencyA)
				mustCreateTask(t, svc, agencyB)
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := NewTaskHandler(svc)

			url := "/api/tasks"
			if tt.agencyID != "" {
				url += "?agency_id=" + tt.agencyID
			}
			r := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			h.List(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var got []*domain.Task
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if len(got) != tt.wantLen {
					t.Errorf("len = %d, want %d", len(got), tt.wantLen)
				}
			}
		})
	}
}

// --- Create ---

func TestTaskHandlerCreate(t *testing.T) {
	agencyID := uuid.New()
	userID := uuid.New()
	validBody := fmt.Sprintf(`{"title":"Fix bug","priority":"high","agency_id":%q,"created_by":%q}`, agencyID, userID)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid request", validBody, http.StatusCreated},
		{"malformed json", `{bad json}`, http.StatusBadRequest},
		{"missing title", fmt.Sprintf(`{"priority":"high","agency_id":%q,"created_by":%q}`, agencyID, userID), http.StatusBadRequest},
		{"missing priority", fmt.Sprintf(`{"title":"Fix bug","agency_id":%q,"created_by":%q}`, agencyID, userID), http.StatusBadRequest},
		{"missing agency_id", fmt.Sprintf(`{"title":"Fix bug","priority":"high","created_by":%q}`, userID), http.StatusBadRequest},
		{"missing created_by", fmt.Sprintf(`{"title":"Fix bug","priority":"high","agency_id":%q}`, agencyID), http.StatusBadRequest},
		{"invalid agency_id uuid", `{"title":"Fix bug","priority":"high","agency_id":"bad"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewTaskHandler(newTaskService())

			r := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			h.Create(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusCreated {
				var task domain.Task
				if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if task.ID == uuid.Nil {
					t.Error("response task has zero ID")
				}
				if task.Status != "todo" {
					t.Errorf("status = %q, want %q", task.Status, "todo")
				}
				if task.AgencyID != agencyID {
					t.Errorf("agency_id = %v, want %v", task.AgencyID, agencyID)
				}
			}
		})
	}
}

// --- Get ---

func TestTaskHandlerGet(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()

	tests := []struct {
		name            string
		idStr           func(*services.TaskService) string
		requesterAgency string
		wantStatus      int
	}{
		{
			name:            "invalid task uuid",
			idStr:           func(*services.TaskService) string { return "not-a-uuid" },
			requesterAgency: agencyA.String(),
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "missing agency_id",
			idStr:           func(*services.TaskService) string { return uuid.New().String() },
			requesterAgency: "",
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "invalid agency_id uuid",
			idStr:           func(*services.TaskService) string { return uuid.New().String() },
			requesterAgency: "not-a-uuid",
			wantStatus:      http.StatusBadRequest,
		},
		{
			name:            "task not found",
			idStr:           func(*services.TaskService) string { return uuid.New().String() },
			requesterAgency: agencyA.String(),
			wantStatus:      http.StatusNotFound,
		},
		{
			name: "forbidden — different agency",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			requesterAgency: agencyB.String(),
			wantStatus:      http.StatusForbidden,
		},
		{
			name: "success",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			requesterAgency: agencyA.String(),
			wantStatus:      http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			idStr := tt.idStr(svc)
			h := NewTaskHandler(svc)

			url := "/api/tasks/" + idStr
			if tt.requesterAgency != "" {
				url += "?agency_id=" + tt.requesterAgency
			}
			r := httptest.NewRequest(http.MethodGet, url, nil)
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.Get(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var task domain.Task
				if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
					t.Fatalf("unmarshal response: %v", err)
				}
				if task.ID == uuid.Nil {
					t.Error("response task has zero ID")
				}
			}
		})
	}
}

// --- Assign ---

func TestTaskHandlerAssign(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()
	assigneeID := uuid.New()

	validBody := func(assigneeAgency uuid.UUID) string {
		return fmt.Sprintf(`{"assignee_id":%q,"assignee_agency_id":%q}`, assigneeID, assigneeAgency)
	}

	tests := []struct {
		name       string
		idStr      func(*services.TaskService) string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid task uuid",
			idStr:      func(*services.TaskService) string { return "not-a-uuid" },
			body:       validBody(agencyA),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed json",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       `{bad}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing assignee_id",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       fmt.Sprintf(`{"assignee_agency_id":%q}`, agencyA),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing assignee_agency_id",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       fmt.Sprintf(`{"assignee_id":%q}`, assigneeID),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "task not found",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       validBody(agencyA),
			wantStatus: http.StatusNotFound,
		},
		{
			name: "forbidden — different agency",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			body:       validBody(agencyB),
			wantStatus: http.StatusForbidden,
		},
		{
			name: "success",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			body:       validBody(agencyA),
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			idStr := tt.idStr(svc)
			h := NewTaskHandler(svc)

			r := httptest.NewRequest(http.MethodPost, "/api/tasks/"+idStr+"/assign", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.Assign(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- Unassign ---

func TestTaskHandlerUnassign(t *testing.T) {
	agencyA := uuid.New()
	assigneeID := uuid.New()

	tests := []struct {
		name       string
		idStr      func(*services.TaskService) string
		wantStatus int
	}{
		{
			name:       "invalid task uuid",
			idStr:      func(*services.TaskService) string { return "not-a-uuid" },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "task not found",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			wantStatus: http.StatusNotFound,
		},
		{
			name: "success — clears assignee",
			idStr: func(svc *services.TaskService) string {
				task := mustCreateTask(t, svc, agencyA)
				_ = svc.AssignTask(context.Background(), task.ID, assigneeID, agencyA)
				return task.ID.String()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "success — already unassigned",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			idStr := tt.idStr(svc)
			h := NewTaskHandler(svc)

			r := httptest.NewRequest(http.MethodPost, "/api/tasks/"+idStr+"/unassign", nil)
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.Unassign(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- Complete ---

func TestTaskHandlerComplete(t *testing.T) {
	agencyA := uuid.New()

	tests := []struct {
		name       string
		idStr      func(*services.TaskService) string
		wantStatus int
	}{
		{
			name:       "invalid task uuid",
			idStr:      func(*services.TaskService) string { return "not-a-uuid" },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "task not found",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			wantStatus: http.StatusNotFound,
		},
		{
			name: "already done",
			idStr: func(svc *services.TaskService) string {
				task := mustCreateTask(t, svc, agencyA)
				_ = svc.CompleteTask(context.Background(), task.ID, time.Now())
				return task.ID.String()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "success",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			idStr := tt.idStr(svc)
			h := NewTaskHandler(svc)

			r := httptest.NewRequest(http.MethodPost, "/api/tasks/"+idStr+"/complete", nil)
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.Complete(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- UpdateDescription ---

func TestTaskHandlerUpdateDescription(t *testing.T) {
	agencyA := uuid.New()

	tests := []struct {
		name       string
		idStr      func(*services.TaskService) string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid task uuid",
			idStr:      func(*services.TaskService) string { return "not-a-uuid" },
			body:       `{"description":"Fix the login flow"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed json",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       `{bad}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "task not found",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       `{"description":"Fix the login flow"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "success — sets description",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			body:       `{"description":"Fix the login flow"}`,
			wantStatus: http.StatusNoContent,
		},
		{
			name: "success — clears description",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			body:       `{"description":null}`,
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			idStr := tt.idStr(svc)
			h := NewTaskHandler(svc)

			r := httptest.NewRequest(http.MethodPatch, "/api/tasks/"+idStr+"/description", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.UpdateDescription(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- UpdateDueDate ---

func TestTaskHandlerUpdateDueDate(t *testing.T) {
	agencyA := uuid.New()

	tests := []struct {
		name       string
		idStr      func(*services.TaskService) string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid task uuid",
			idStr:      func(*services.TaskService) string { return "not-a-uuid" },
			body:       `{"due_date":"2026-06-01"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed json",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       `{bad}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid date format",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       `{"due_date":"06/01/2026"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "task not found",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			body:       `{"due_date":"2026-06-01"}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "success — sets a due date",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			body:       `{"due_date":"2026-06-01"}`,
			wantStatus: http.StatusNoContent,
		},
		{
			name: "success — clears the due date",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			body:       `{"due_date":null}`,
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			idStr := tt.idStr(svc)
			h := NewTaskHandler(svc)

			r := httptest.NewRequest(http.MethodPatch, "/api/tasks/"+idStr+"/due-date", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.UpdateDueDate(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// --- SetInProgress ---

func TestTaskHandlerSetInProgress(t *testing.T) {
	agencyA := uuid.New()

	tests := []struct {
		name       string
		idStr      func(*services.TaskService) string
		wantStatus int
	}{
		{
			name:       "invalid task uuid",
			idStr:      func(*services.TaskService) string { return "not-a-uuid" },
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "task not found",
			idStr:      func(*services.TaskService) string { return uuid.New().String() },
			wantStatus: http.StatusNotFound,
		},
		{
			name: "already in_progress",
			idStr: func(svc *services.TaskService) string {
				task := mustCreateTask(t, svc, agencyA)
				_ = svc.SetInProgress(context.Background(), task.ID)
				return task.ID.String()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "success from todo",
			idStr: func(svc *services.TaskService) string {
				return mustCreateTask(t, svc, agencyA).ID.String()
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "success from done",
			idStr: func(svc *services.TaskService) string {
				task := mustCreateTask(t, svc, agencyA)
				_ = svc.CompleteTask(context.Background(), task.ID, time.Now())
				return task.ID.String()
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			idStr := tt.idStr(svc)
			h := NewTaskHandler(svc)

			r := httptest.NewRequest(http.MethodPost, "/api/tasks/"+idStr+"/set-in-progress", nil)
			r = withChiParam(r, "id", idStr)
			w := httptest.NewRecorder()
			h.SetInProgress(w, r)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}
