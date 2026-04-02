package domain

import (
	"errors"
	"testing"
	"time"

	"backend/internal/apperrors"

	"github.com/google/uuid"
)

func TestIsAssigned(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name string
		task Task
		want bool
	}{
		{"unassigned task", Task{ID: uuid.New(), AssigneeID: nil}, false},
		{"assigned task", Task{ID: uuid.New(), AssigneeID: &userID}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.IsAssigned()
			if got != tt.want {
				t.Errorf("IsAssigned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsComplete(t *testing.T) {
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		task Task
		want bool
	}{
		{"incomplete task", Task{ID: uuid.New(), CompletedAt: nil}, false},
		{"complete task", Task{ID: uuid.New(), CompletedAt: timePtr(now)}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.IsComplete()
			if got != tt.want {
				t.Errorf("IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskIsOverdue(t *testing.T) {
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)
	past := now.Add(-48 * time.Hour)
	future := now.Add(48 * time.Hour)

	tests := []struct {
		name string
		task Task
		want bool
	}{
		{"no due date", Task{ID: uuid.New(), DueDate: nil}, false},
		{"future due date", Task{ID: uuid.New(), DueDate: timePtr(future)}, false},
		{"past due date, incomplete", Task{ID: uuid.New(), DueDate: timePtr(past)}, true},
		{"past due date, complete", Task{ID: uuid.New(), DueDate: timePtr(past), CompletedAt: timePtr(now.Add(-24 * time.Hour))}, false},
		{"due exactly now", Task{ID: uuid.New(), DueDate: timePtr(now)}, false},
		{"due one second ago", Task{ID: uuid.New(), DueDate: timePtr(now.Add(-time.Second))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.IsOverdue(now)
			if got != tt.want {
				t.Errorf("IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssign(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()

	tests := []struct {
		name       string
		task       Task
		assigneeID uuid.UUID
		wantID     uuid.UUID
	}{
		{"assign to unassigned task", Task{ID: uuid.New()}, userA, userA},
		{"reassign to different user", Task{ID: uuid.New(), AssigneeID: &userA}, userB, userB},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := tt.task
			task.Assign(tt.assigneeID)
			if task.AssigneeID == nil {
				t.Fatal("Assign() did not set AssigneeID")
			}
			if *task.AssigneeID != tt.wantID {
				t.Errorf("Assign() AssigneeID = %v, want %v", *task.AssigneeID, tt.wantID)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		task    Task
		wantErr error
	}{
		{"todo task", Task{ID: uuid.New(), Status: "todo"}, nil},
		{"in_progress task", Task{ID: uuid.New(), Status: "in_progress"}, nil},
		{"already done", Task{ID: uuid.New(), Status: "done", CompletedAt: timePtr(now.Add(-time.Hour))}, apperrors.ErrConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := tt.task
			err := task.Complete(now)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Complete() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Complete() unexpected error: %v", err)
			}
			if task.CompletedAt == nil {
				t.Error("Complete() did not set CompletedAt")
			}
			if !task.CompletedAt.Equal(now) {
				t.Errorf("Complete() CompletedAt = %v, want %v", task.CompletedAt, now)
			}
			if task.Status != "done" {
				t.Errorf("Complete() Status = %q, want %q", task.Status, "done")
			}
		})
	}
}

func TestFilterByStatus(t *testing.T) {
	tasks := []Task{
		{ID: uuid.New(), Status: "todo"},
		{ID: uuid.New(), Status: "in_progress"},
		{ID: uuid.New(), Status: "done"},
		{ID: uuid.New(), Status: "todo"},
	}

	tests := []struct {
		name   string
		status string
		want   int
	}{
		{"filter todo", "todo", 2},
		{"filter in_progress", "in_progress", 1},
		{"filter done", "done", 1},
		{"filter nonexistent", "cancelled", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByStatus(tasks, tt.status)
			if len(got) != tt.want {
				t.Errorf("FilterByStatus(%q) returned %d tasks, want %d", tt.status, len(got), tt.want)
			}
		})
	}
}

func TestFilterByAssignee(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	tasks := []Task{
		{ID: uuid.New(), AssigneeID: &userA},
		{ID: uuid.New(), AssigneeID: &userB},
		{ID: uuid.New(), AssigneeID: &userA},
		{ID: uuid.New(), AssigneeID: nil},
	}

	tests := []struct {
		name       string
		assigneeID uuid.UUID
		want       int
	}{
		{"filter by userA", userA, 2},
		{"filter by userB", userB, 1},
		{"filter by unknown user", uuid.New(), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByAssignee(tasks, tt.assigneeID)
			if len(got) != tt.want {
				t.Errorf("FilterByAssignee() returned %d tasks, want %d", len(got), tt.want)
			}
		})
	}
}

func TestFilterByPriority(t *testing.T) {
	tasks := []Task{
		{ID: uuid.New(), Priority: "high"},
		{ID: uuid.New(), Priority: "medium"},
		{ID: uuid.New(), Priority: "low"},
		{ID: uuid.New(), Priority: "high"},
	}

	tests := []struct {
		name     string
		priority string
		want     int
	}{
		{"filter high", "high", 2},
		{"filter medium", "medium", 1},
		{"filter low", "low", 1},
		{"filter nonexistent", "critical", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterByPriority(tasks, tt.priority)
			if len(got) != tt.want {
				t.Errorf("FilterByPriority(%q) returned %d tasks, want %d", tt.priority, len(got), tt.want)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time { return &t }
