package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/apperrors"
	"backend/internal/repositories/memory"
	services "backend/internal/services/task"

	"github.com/google/uuid"
)

var (
	ctx = context.Background()
	now = time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)
)

func newTaskService() *services.TaskService {
	return services.NewTaskService(memory.NewTaskRepo())
}

func strPtr(s string) *string {
	return &s
}

func TestCreateTask(t *testing.T) {
	agencyID := uuid.New()

	tests := []struct {
		name        string
		title       string
		priority    string
		agencyID    uuid.UUID
		description *string
		asigneeId   *uuid.UUID
		dueDate     *time.Time
		wantStatus  string
	}{
		{"todo status on creation", "Fix bug", "high", agencyID, strPtr("error on Ln 88, Col 58 in task.go"), nil, &now, "todo"},
		{"low priority task", "Write docs", "low", agencyID, strPtr("write a README.md"), nil, &now, "todo"},
		{"task belongs to agency", "Deploy service", "medium", agencyID, nil, nil, nil, "todo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			task, err := svc.Create(ctx, tt.title, tt.priority, tt.agencyID, tt.description, tt.asigneeId, tt.dueDate)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if task.Title != tt.title {
				t.Errorf("Title = %q, want %q", task.Title, tt.title)
			}
			if task.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", task.Status, tt.wantStatus)
			}
			if task.Priority != tt.priority {
				t.Errorf("Priority = %q, want %q", task.Priority, tt.priority)
			}
			if task.AgencyID != tt.agencyID {
				t.Errorf("AgencyID = %v, want %v", task.AgencyID, tt.agencyID)
			}
			if tt.description == nil && task.Description != nil {
				t.Errorf("Description = %v, want nil", *task.Description)
			}
			if tt.description != nil && task.Description == nil {
				t.Errorf("Description = %v, want %v", task.Description, tt.description)
			}
			if tt.description != nil && *task.Description != *tt.description {
				t.Errorf("Description = %v, want %v", task.Description, tt.description)
			}
			if tt.asigneeId == nil && task.AssigneeID != nil {
				t.Errorf("AssigneeID = %v, want nil", *task.AssigneeID)
			}
			if tt.asigneeId != nil && task.AssigneeID == nil {
				t.Errorf("AssigneeID = %v, want %v", task.AssigneeID, tt.asigneeId)
			}
			if tt.asigneeId != nil && *task.AssigneeID != *tt.asigneeId {
				t.Errorf("AssigneeID = %v, want %v", task.AssigneeID, tt.asigneeId)
			}
			if tt.dueDate == nil && task.DueDate != nil {
				t.Errorf("DueDate = %v, want nil", task.DueDate)
			}
			if tt.dueDate != nil && task.DueDate == nil {
				t.Errorf("DueDate = %v, want %v", task.DueDate, tt.dueDate)
			}
			if tt.dueDate != nil && !task.DueDate.Equal(*tt.dueDate) {
				t.Errorf("DueDate = %v, want %v", task.DueDate, tt.dueDate)
			}
			if task.ID == (uuid.UUID{}) {
				t.Error("ID must not be zero")
			}
		})
	}
}

func TestAssignTask(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()
	assigneeID := uuid.New()

	tests := []struct {
		name             string
		setup            func(svc *services.TaskService) uuid.UUID
		assigneeID       uuid.UUID
		assigneeAgencyID uuid.UUID
		wantErr          error
	}{
		{
			name: "success — same agency",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA, nil, nil, nil)
				return task.ID
			},
			assigneeID:       assigneeID,
			assigneeAgencyID: agencyA,
			wantErr:          nil,
		},
		{
			name: "forbidden — different agency",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA, nil, nil, nil)
				return task.ID
			},
			assigneeID:       assigneeID,
			assigneeAgencyID: agencyB,
			wantErr:          apperrors.ErrForbidden,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA, nil, nil, nil)
				id := task.ID
				id[0] ^= 0xFF
				return id
			},
			assigneeID:       assigneeID,
			assigneeAgencyID: agencyA,
			wantErr:          apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			taskID := tt.setup(svc)
			err := svc.AssignTask(ctx, taskID, tt.assigneeID, tt.assigneeAgencyID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("AssignTask() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnassignTask(t *testing.T) {
	agencyID := uuid.New()
	assigneeID := uuid.New()

	tests := []struct {
		name    string
		setup   func(svc *services.TaskService) uuid.UUID
		wantErr error
	}{
		{
			name: "success — clears assignee",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, &assigneeID, nil)
				return task.ID
			},
			wantErr: nil,
		},
		{
			name: "success — unassigning already unassigned task",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				return task.ID
			},
			wantErr: nil,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				id := task.ID
				id[0] ^= 0xFF
				return id
			},
			wantErr: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			taskID := tt.setup(svc)
			err := svc.UnassignTask(ctx, taskID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UnassignTask() error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			task, err := svc.GetTask(ctx, taskID, agencyID)
			if err != nil {
				t.Fatalf("unexpected error fetching task: %v", err)
			}
			if task.AssigneeID != nil {
				t.Errorf("UnassignTask() AssigneeID = %v, want nil", task.AssigneeID)
			}
		})
	}
}

func TestCompleteTask(t *testing.T) {
	agencyID := uuid.New()

	tests := []struct {
		name            string
		setup           func(svc *services.TaskService) (uuid.UUID, uuid.UUID)
		wantErr         error
		wantStatus      string
		wantCompletedAt *time.Time
	}{
		{
			name: "success",
			setup: func(svc *services.TaskService) (uuid.UUID, uuid.UUID) {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				return task.ID, task.AgencyID
			},
			wantErr:         nil,
			wantStatus:      "done",
			wantCompletedAt: &now,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) (uuid.UUID, uuid.UUID) {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				id := task.ID
				id[0] ^= 0xFF
				return id, task.AgencyID
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "already done",
			setup: func(svc *services.TaskService) (uuid.UUID, uuid.UUID) {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				svc.CompleteTask(ctx, task.ID, now)
				return task.ID, task.AgencyID
			},
			wantErr: apperrors.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			taskID, AgencyID := tt.setup(svc)
			err := svc.CompleteTask(ctx, taskID, now)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CompleteTask() error = %v, want %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			task, err := svc.GetTask(ctx, taskID, AgencyID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if task.Status != tt.wantStatus {
				t.Errorf("Status = %q, want %q", task.Status, tt.wantStatus)
			}
			if !task.CompletedAt.Equal(*tt.wantCompletedAt) {
				t.Errorf("CompletedAt = %q, want %q", task.CompletedAt, tt.wantCompletedAt)
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()

	tests := []struct {
		name            string
		setup           func(svc *services.TaskService) uuid.UUID
		requesterAgency uuid.UUID
		wantErr         error
	}{
		{
			name: "success — same agency",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA, nil, nil, nil)
				return task.ID
			},
			requesterAgency: agencyA,
			wantErr:         nil,
		},
		{
			name: "forbidden — different agency",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA, nil, nil, nil)
				return task.ID
			},
			requesterAgency: agencyB,
			wantErr:         apperrors.ErrForbidden,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA, nil, nil, nil)
				id := task.ID
				id[0] ^= 0xFF
				return id
			},
			requesterAgency: agencyA,
			wantErr:         apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			taskID := tt.setup(svc)
			task, err := svc.GetTask(ctx, taskID, tt.requesterAgency)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetTask() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && task == nil {
				t.Error("expected task, got nil")
			}
		})
	}
}

func TestListByAgency(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()

	tests := []struct {
		name      string
		setup     func(svc *services.TaskService)
		agencyID  uuid.UUID
		wantCount int
	}{
		{
			name:      "empty",
			setup:     func(svc *services.TaskService) {},
			agencyID:  agencyA,
			wantCount: 0,
		},
		{
			name: "only returns tasks from requested agency",
			setup: func(svc *services.TaskService) {
				svc.Create(ctx, "Task 1", "high", agencyA, nil, nil, nil)
				svc.Create(ctx, "Task 2", "medium", agencyA, nil, nil, nil)
				svc.Create(ctx, "Task 3", "low", agencyB, nil, nil, nil)
			},
			agencyID:  agencyA,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			tt.setup(svc)
			tasks, err := svc.ListByAgency(ctx, tt.agencyID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("len(tasks) = %d, want %d", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestListOverdue(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()
	past := now.Add(-48 * time.Hour)
	future := now.Add(48 * time.Hour)

	tests := []struct {
		name      string
		setup     func(svc *services.TaskService)
		agencyID  uuid.UUID
		wantCount int
	}{
		{
			name: "overdue task appears",
			setup: func(svc *services.TaskService) {
				svc.Create(ctx, "Overdue task", "high", agencyA, nil, nil, &past)
			},
			agencyID:  agencyA,
			wantCount: 1,
		},
		{
			name: "future task does not appear",
			setup: func(svc *services.TaskService) {
				svc.Create(ctx, "Future task", "low", agencyA, nil, nil, &future)
			},
			agencyID:  agencyA,
			wantCount: 0,
		},
		{
			name: "completed task is not overdue",
			setup: func(svc *services.TaskService) {
				task, _ := svc.Create(ctx, "Done task", "medium", agencyA, nil, nil, &past)
				svc.CompleteTask(ctx, task.ID, now)
			},
			agencyID:  agencyA,
			wantCount: 0,
		},
		{
			name: "only returns overdue tasks from requested agency",
			setup: func(svc *services.TaskService) {
				svc.Create(ctx, "Agency A task", "high", agencyA, nil, nil, &past)
				svc.Create(ctx, "Agency B task", "high", agencyB, nil, nil, &past)
			},
			agencyID:  agencyA,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			tt.setup(svc)
			tasks, err := svc.ListOverdue(ctx, tt.agencyID, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("len(tasks) = %d, want %d", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestSetInProgress(t *testing.T) {
	agencyID := uuid.New()
	past := now.Add(-48 * time.Hour)

	tests := []struct {
		name       string
		setup      func(svc *services.TaskService) uuid.UUID
		wantStatus string
		wantErr    error
	}{
		{
			name: "todo task transitions to in_progress",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				return task.ID
			},
			wantStatus: "in_progress",
			wantErr:    nil,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				id := task.ID
				id[0] ^= 0xFF
				return id
			},
			wantStatus: "",
			wantErr:    apperrors.ErrNotFound,
		},
		{
			name: "already in_progress stays unchanged",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, nil)
				svc.SetInProgress(ctx, task.ID)
				return task.ID
			},
			wantStatus: "in_progress",
			wantErr:    apperrors.ErrConflict,
		},
		{
			name: "done task transitions to in_progress with nil CompletedAt",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID, nil, nil, &past)
				svc.CompleteTask(ctx, task.ID, now)
				return task.ID
			},
			wantStatus: "in_progress",
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			taskID := tt.setup(svc)
			err := svc.SetInProgress(ctx, taskID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("SetInProgress() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == apperrors.ErrNotFound {
				return
			}
			task, _ := svc.GetTask(ctx, taskID, agencyID)
			if task.Status != tt.wantStatus {
				t.Errorf("SetInProgress() Status = %q, want %q", task.Status, tt.wantStatus)
			}
			if tt.wantErr == nil && task.CompletedAt != nil {
				t.Errorf("SetInProgress() CompletedAt = %v, want nil", task.CompletedAt)
			}
		})
	}
}
