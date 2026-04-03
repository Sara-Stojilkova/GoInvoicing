package services

import (
	"context"
	"time"

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

func (s *TaskService) Create(ctx context.Context, title, priority string, agencyID uuid.UUID) (*domain.Task, error) {
	panic("not implemented")
}

func (s *TaskService) AssignTask(ctx context.Context, taskID, assigneeID, assigneeAgencyID uuid.UUID) error {
	panic("not implemented")
}

func (s *TaskService) CompleteTask(ctx context.Context, taskID uuid.UUID, now time.Time) error {
	panic("not implemented")
}

func (s *TaskService) GetTask(ctx context.Context, taskID uuid.UUID, requesterAgencyID uuid.UUID) (*domain.Task, error) {
	panic("not implemented")
}

func (s *TaskService) ListByAgency(ctx context.Context, agencyID uuid.UUID) ([]*domain.Task, error) {
	panic("not implemented")
}

func (s *TaskService) ListOverdue(ctx context.Context, agencyID uuid.UUID, now time.Time) ([]*domain.Task, error) {
	panic("not implemented")
}

func (s *TaskService) SetDueDate(ctx context.Context, taskID uuid.UUID, dueDate time.Time) error {
	panic("not implemented")
}
