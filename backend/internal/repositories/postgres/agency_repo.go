package postgres

import (
	"context"
	"fmt"

	domain "backend/internal/domain/agency"
	"backend/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type agencyRepo struct {
	db *pgxpool.Pool
}

func NewAgencyRepo(db *pgxpool.Pool) repositories.AgencyRepository {
	return &agencyRepo{db: db}
}

func (r *agencyRepo) Create(ctx context.Context, agency *domain.Agency) error {
	_, err := r.db.Exec(ctx, `
		insert into agencies (id, name)
		values ($1, $2)`,
		agency.ID,
		agency.Name,
	)
	if err != nil {
		return fmt.Errorf("create agency: %w", mapErr(err))
	}
	return nil
}

func (r *agencyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agency, error) {
	row := r.db.QueryRow(ctx, `
		select id, name, created_at
		from agencies
		where id = $1 and deleted_at is null`, id)

	var a domain.Agency
	if err := row.Scan(&a.ID, &a.Name, &a.CreatedAt); err != nil {
		return nil, fmt.Errorf("agency %s: %w", id, mapErr(err))
	}
	return &a, nil
}

func (r *agencyRepo) List(ctx context.Context) ([]*domain.Agency, error) {
	rows, err := r.db.Query(ctx, `
		select id, name, created_at
		from agencies
		where deleted_at is null
		order by created_at desc`)
	if err != nil {
		return nil, fmt.Errorf("list agencies: %w", err)
	}
	defer rows.Close()

	var agencies []*domain.Agency
	for rows.Next() {
		var a domain.Agency
		if err := rows.Scan(&a.ID, &a.Name, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("list agencies scan: %w", err)
		}
		agencies = append(agencies, &a)
	}
	return agencies, rows.Err()
}
