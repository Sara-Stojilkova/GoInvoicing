package memory

import (
	"context"
	"fmt"
	"sync"

	"backend/internal/apperrors"
	domain "backend/internal/domain/user"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type userRepo struct {
	mu    sync.RWMutex
	users map[uuid.UUID]*domain.User
}

func NewUserRepo() repositories.UserRepository {
	return &userRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("user %s: %w", id, apperrors.ErrNotFound)
	}
	return u, nil
}

func (r *userRepo) List(ctx context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.User, 0, len(r.users))
	for _, u := range r.users {
		result = append(result, u)
	}
	return result, nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[user.ID]; !ok {
		return fmt.Errorf("user %s: %w", user.ID, apperrors.ErrNotFound)
	}
	r.users[user.ID] = user
	return nil
}

func (r *userRepo) UpdateSignupFields(ctx context.Context, id uuid.UUID, email string, activated bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	u, ok := r.users[id]
	if !ok {
		// In production the DB trigger creates this row before this is called.
		// In unit tests the trigger doesn't fire, so we create a minimal record.
		r.users[id] = &domain.User{ID: id, Email: email}
		return nil
	}
	u.Email = email
	return nil
}
