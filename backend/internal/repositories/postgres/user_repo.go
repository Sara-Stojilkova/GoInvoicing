package postgres

import (
	"context"
	"fmt"

	"backend/internal/apperrors"
	domain "backend/internal/domain/user"
	"backend/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) repositories.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx, `
		insert into users (id, agency_id, full_name, email)
		values ($1, $2, $3, $4)`,
		user.ID,
		user.AgencyID,
		user.FullName,
		user.Email,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", mapErr(err))
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row := r.db.QueryRow(ctx, `
		select id, agency_id, coalesce(full_name, ''), coalesce(email, ''), activated, created_at
		from users
		where id = $1 and deleted_at is null`, id)

	var u domain.User
	if err := row.Scan(&u.ID, &u.AgencyID, &u.FullName, &u.Email, &u.Activated, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("user %s: %w", id, mapErr(err))
	}
	return &u, nil
}

func (r *userRepo) List(ctx context.Context) ([]*domain.User, error) {
	rows, err := r.db.Query(ctx, `
		select id, agency_id, coalesce(full_name, ''), coalesce(email, ''), activated, created_at
		from users
		where deleted_at is null and agency_id is not null
		order by created_at desc`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.AgencyID, &u.FullName, &u.Email, &u.Activated, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("list users scan: %w", err)
		}
		users = append(users, &u)
	}
	return users, rows.Err()
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	tag, err := r.db.Exec(ctx, `
		update users
		set full_name = $1
		where id = $2 and deleted_at is null`,
		user.FullName,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user %s: %w", user.ID, mapErr(err))
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user %s: %w", user.ID, apperrors.ErrNotFound)
	}
	return nil
}

func (r *userRepo) UpdateSignupFields(ctx context.Context, id uuid.UUID, agencyID uuid.UUID, email string, activated bool) error {
	tag, err := r.db.Exec(ctx, `
		update users
		set agency_id = $1, email = $2, activated = $3
		where id = $4 and deleted_at is null`,
		agencyID, email, activated, id,
	)
	if err != nil {
		return fmt.Errorf("update signup fields for user %s: %w", id, mapErr(err))
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user %s: %w", id, apperrors.ErrNotFound)
	}
	return nil
}
