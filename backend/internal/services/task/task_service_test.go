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

func TestCreateTask(t *testing.T) {
	agencyID := uuid.New()

	tests := []struct {
		name       string
		title      string
		priority   string
		agencyID   uuid.UUID
		wantStatus string
	}{
		{"todo status on creation", "Fix bug", "high", agencyID, "todo"},
		{"low priority task", "Write docs", "low", agencyID, "todo"},
		{"task belongs to agency", "Deploy service", "medium", agencyID, "todo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			task, err := svc.Create(ctx, tt.title, tt.priority, tt.agencyID)
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
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA)
				return task.ID
			},
			assigneeID:       assigneeID,
			assigneeAgencyID: agencyA,
			wantErr:          nil,
		},
		{
			name: "forbidden — different agency",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA)
				return task.ID
			},
			assigneeID:       assigneeID,
			assigneeAgencyID: agencyB,
			wantErr:          apperrors.ErrForbidden,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA)
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

func TestCompleteTask(t *testing.T) {
	agencyID := uuid.New()

	tests := []struct {
		name    string
		setup   func(svc *services.TaskService) uuid.UUID
		wantErr error
	}{
		{
			name: "success",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID)
				return task.ID
			},
			wantErr: nil,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID)
				id := task.ID
				id[0] ^= 0xFF
				return id
			},
			wantErr: apperrors.ErrNotFound,
		},
		{
			name: "already done",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID)
				svc.CompleteTask(ctx, task.ID, now)
				return task.ID
			},
			wantErr: apperrors.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTaskService()
			taskID := tt.setup(svc)
			err := svc.CompleteTask(ctx, taskID, now)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CompleteTask() error = %v, want %v", err, tt.wantErr)
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
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA)
				return task.ID
			},
			requesterAgency: agencyA,
			wantErr:         nil,
		},
		{
			name: "forbidden — different agency",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA)
				return task.ID
			},
			requesterAgency: agencyB,
			wantErr:         apperrors.ErrForbidden,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyA)
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
				svc.Create(ctx, "Task 1", "high", agencyA)
				svc.Create(ctx, "Task 2", "medium", agencyA)
				svc.Create(ctx, "Task 3", "low", agencyB)
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
				task, _ := svc.Create(ctx, "Overdue task", "high", agencyA)
				task.DueDate = &past
				// update via assign to persist due date
				svc.SetDueDate(ctx, task.ID, past)
			},
			agencyID:  agencyA,
			wantCount: 1,
		},
		{
			name: "future task does not appear",
			setup: func(svc *services.TaskService) {
				task, _ := svc.Create(ctx, "Future task", "low", agencyA)
				svc.SetDueDate(ctx, task.ID, future)
			},
			agencyID:  agencyA,
			wantCount: 0,
		},
		{
			name: "completed task is not overdue",
			setup: func(svc *services.TaskService) {
				task, _ := svc.Create(ctx, "Done task", "medium", agencyA)
				svc.SetDueDate(ctx, task.ID, past)
				svc.CompleteTask(ctx, task.ID, now)
			},
			agencyID:  agencyA,
			wantCount: 0,
		},
		{
			name: "only returns overdue tasks from requested agency",
			setup: func(svc *services.TaskService) {
				taskA, _ := svc.Create(ctx, "Agency A task", "high", agencyA)
				svc.SetDueDate(ctx, taskA.ID, past)
				taskB, _ := svc.Create(ctx, "Agency B task", "high", agencyB)
				svc.SetDueDate(ctx, taskB.ID, past)
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

	tests := []struct {
		name       string
		setup      func(svc *services.TaskService) uuid.UUID
		wantStatus string
		wantErr    error
	}{
		{
			name: "todo task transitions to in_progress",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID)
				return task.ID
			},
			wantStatus: "in_progress",
			wantErr:    nil,
		},
		{
			name: "not found",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID)
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
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID)
				svc.SetInProgress(ctx, task.ID)
				return task.ID
			},
			wantStatus: "in_progress",
			wantErr:    apperrors.ErrConflict,
		},
		{
			name: "done task transitions to in_progress with nil CompletedAt",
			setup: func(svc *services.TaskService) uuid.UUID {
				task, _ := svc.Create(ctx, "Fix bug", "high", agencyID)
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
