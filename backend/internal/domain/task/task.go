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
	Description *string    `json:"description"` // nil = not set
	Status      string     `json:"status"`      // "todo", "in_progress", "done"
	Priority    string     `json:"priority"`    // "low", "medium", "high"
	AgencyID    uuid.UUID  `json:"agency_id"`   // agency of the user who created this task
	AssigneeID  *uuid.UUID `json:"assignee_id"` // nil = unassigned
	CreatedAt   time.Time  `json:"created_at"`
	DueDate     *time.Time `json:"due_date"`     // nil = no due date
	CompletedAt *time.Time `json:"completed_at"` // nil = not complete
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

// CanBeAssignedTo returns true if the assignee belongs to the same agency as the task.
func (t Task) CanBeAssignedTo(assigneeAgencyID uuid.UUID) bool {
	return t.AgencyID == assigneeAgencyID
}

// IsAccessibleBy returns true if the requesting user belongs to the same agency as the task.
func (t Task) IsAccessibleBy(userAgencyID uuid.UUID) bool {
	return t.AgencyID == userAgencyID
}

func (t *Task) Assign(userID uuid.UUID) {
	t.AssigneeID = &userID
}

func (t *Task) Unassign() {
	t.AssigneeID = nil
}

func (t *Task) SetInProgress() error {
	if t.Status == "in_progress" {
		return fmt.Errorf("task %s: %w", t.ID, apperrors.ErrConflict)
	}
	t.Status = "in_progress"
	t.CompletedAt = nil
	return nil
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

func FilterByAgency(tasks []Task, agencyID uuid.UUID) []Task {
	var result []Task
	for _, t := range tasks {
		if t.AgencyID == agencyID {
			result = append(result, t)
		}
	}
	return result
}
