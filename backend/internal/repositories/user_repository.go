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
	// UpdateSignupFields sets the email and activated flag on a user row.
	// Called after Supabase Auth creates the auth.users record (which triggers
	// public.users creation). email is always set; activated is true only for
	// the first user of a new agency.
	UpdateSignupFields(ctx context.Context, id uuid.UUID, email string, activated bool) error
}
