package repositories

import (
	"context"

	domain "backend/internal/domain/user"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}
