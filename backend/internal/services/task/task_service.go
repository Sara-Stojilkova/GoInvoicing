package services

import (
	"context"
	"fmt"
	"time"

	"backend/internal/apperrors"
	domain "backend/internal/domain/task"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type TaskService struct {
	repo repositories.TaskRepository
}

func NewTaskService(repo repositories.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(
	ctx context.Context,
	title string,
	priority string,
	agencyID uuid.UUID,
	description *string,
	assigneeID *uuid.UUID,
	dueDate *time.Time,
) (*domain.Task, error) {
	task := &domain.Task{
		ID:          uuid.New(),
		Title:       title,
		Priority:    priority,
		AgencyID:    agencyID,
		Description: description,
		AssigneeID:  assigneeID,
		DueDate:     dueDate,
		Status:      "todo",
		CreatedAt:   time.Now(),
		CompletedAt: nil,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) AssignTask(ctx context.Context, taskID, assigneeID, assigneeAgencyID uuid.UUID) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if !task.CanBeAssignedTo(assigneeAgencyID) {
		return fmt.Errorf("task %s: %w", taskID, apperrors.ErrForbidden)
	}
	task.Assign(assigneeID)
	return s.repo.Update(ctx, task)
}

func (s *TaskService) CompleteTask(ctx context.Context, taskID uuid.UUID, now time.Time) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := task.Complete(now); err != nil {
		return err
	}
	return s.repo.Update(ctx, task)
}

func (s *TaskService) GetTask(ctx context.Context, taskID uuid.UUID, requesterAgencyID uuid.UUID) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if !task.IsAccessibleBy(requesterAgencyID) {
		return nil, fmt.Errorf("task %s: %w", taskID, apperrors.ErrForbidden)
	}
	return task, nil
}

func (s *TaskService) ListByAgency(ctx context.Context, agencyID uuid.UUID) ([]*domain.Task, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var result []*domain.Task
	for _, t := range all {
		if t.AgencyID == agencyID {
			result = append(result, t)
		}
	}
	return result, nil
}

func (s *TaskService) ListOverdue(ctx context.Context, agencyID uuid.UUID, now time.Time) ([]*domain.Task, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	var result []*domain.Task
	for _, t := range all {
		if t.AgencyID == agencyID && t.IsOverdue(now) {
			result = append(result, t)
		}
	}
	return result, nil
}

func (s *TaskService) SetInProgress(ctx context.Context, taskID uuid.UUID) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if err := task.SetInProgress(); err != nil {
		return err
	}
	return s.repo.Update(ctx, task)
}

func (s *TaskService) SetDueDate(ctx context.Context, taskID uuid.UUID, dueDate time.Time) error {
	task, err := s.repo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	task.DueDate = &dueDate
	return s.repo.Update(ctx, task)
}
