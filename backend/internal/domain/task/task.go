package domain

import (
	"fmt"
	"time"

	"backend/internal/apperrors"

	"github.com/google/uuid"
)

type Task struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`  // nil = not set
	Status      string     `json:"status"`        // "todo", "in_progress", "done"
	Priority    string     `json:"priority"`      // "low", "medium", "high"
	AssigneeID  *uuid.UUID `json:"assignee_id"`   // nil = unassigned
	CreatedAt   time.Time  `json:"created_at"`
	DueDate     *time.Time `json:"due_date"`      // nil = no due date
	CompletedAt *time.Time `json:"completed_at"`  // nil = not complete
}

func (t Task) IsAssigned() bool {
	return t.AssigneeID != nil
}

func (t Task) IsComplete() bool {
	return t.CompletedAt != nil
}

func (t Task) IsOverdue(now time.Time) bool {
	if t.IsComplete() || t.DueDate == nil {
		return false
	}
	return now.After(*t.DueDate)
}

func (t *Task) Assign(userID uuid.UUID) {
	t.AssigneeID = &userID
}

func (t *Task) Complete(now time.Time) error {
	if t.IsComplete() {
		return fmt.Errorf("task %s: %w", t.ID, apperrors.ErrConflict)
	}
	t.CompletedAt = &now
	t.Status = "done"
	return nil
}

func FilterByStatus(tasks []Task, status string) []Task {
	var result []Task
	for _, t := range tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

func FilterByAssignee(tasks []Task, assigneeID uuid.UUID) []Task {
	var result []Task
	for _, t := range tasks {
		if t.AssigneeID != nil && *t.AssigneeID == assigneeID {
			result = append(result, t)
		}
	}
	return result
}

func FilterByPriority(tasks []Task, priority string) []Task {
	var result []Task
	for _, t := range tasks {
		if t.Priority == priority {
			result = append(result, t)
		}
	}
	return result
}
