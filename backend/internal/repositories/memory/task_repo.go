package memory

import (
	"context"
	"fmt"
	"sync"

	"backend/internal/apperrors"
	domain "backend/internal/domain/task"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type taskRepo struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*domain.Task
}

func NewTaskRepo() repositories.TaskRepository {
	return &taskRepo{tasks: make(map[uuid.UUID]*domain.Task)}
}

func (r *taskRepo) Create(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *taskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %s: %w", id, apperrors.ErrNotFound)
	}
	return t, nil
}

func (r *taskRepo) List(ctx context.Context) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		result = append(result, t)
	}
	return result, nil
}

func (r *taskRepo) Update(ctx context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[task.ID]; !ok {
		return fmt.Errorf("task %s: %w", task.ID, apperrors.ErrNotFound)
	}
	r.tasks[task.ID] = task
	return nil
}

func (r *taskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.tasks[id]; !ok {
		return fmt.Errorf("task %s: %w", id, apperrors.ErrNotFound)
	}
	delete(r.tasks, id)
	return nil
}
