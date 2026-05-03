package postgres

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/apperrors"
	domain "backend/internal/domain/task"
	"backend/internal/repositories"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type taskRepo struct {
	db *pgxpool.Pool
}

func NewTaskRepo(db *pgxpool.Pool) repositories.TaskRepository {
	return &taskRepo{db: db}
}

func (r *taskRepo) Create(ctx context.Context, task *domain.Task) error {
	_, err := r.db.Exec(ctx, `
		insert into tasks
			(id, agency_id, created_by, assigned_to, title, description,
			 status, priority, due_date, completed_at, tags)
		values
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		task.ID,
		task.AgencyID,
		task.CreatedBy,
		task.AssignedTo,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.CompletedAt,
		task.Tags,
	)
	if err != nil {
		return fmt.Errorf("create task: %w", mapErr(err))
	}
	return nil
}

func (r *taskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	row := r.db.QueryRow(ctx, `
		select id, agency_id, created_by, assigned_to, title, description,
		       status, priority, due_date, completed_at, created_at, tags
		from tasks
		where id = $1`, id)

	t, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("task %s: %w", id, mapErr(err))
	}
	return t, nil
}

func (r *taskRepo) List(ctx context.Context) ([]*domain.Task, error) {
	rows, err := r.db.Query(ctx, `
		select id, agency_id, created_by, assigned_to, title, description,
		       status, priority, due_date, completed_at, created_at, tags
		from tasks
		order by created_at desc`)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("list tasks scan: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *taskRepo) Update(ctx context.Context, task *domain.Task) error {
	tag, err := r.db.Exec(ctx, `
		update tasks
		set assigned_to  = $1,
		    title        = $2,
		    description  = $3,
		    status       = $4,
		    priority     = $5,
		    due_date     = $6,
		    completed_at = $7,
		    tags         = $8
		where id = $9`,
		task.AssignedTo,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		task.DueDate,
		task.CompletedAt,
		task.Tags,
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update task %s: %w", task.ID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task %s: %w", task.ID, apperrors.ErrNotFound)
	}
	return nil
}

func (r *taskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `delete from tasks where id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("task %s: %w", id, apperrors.ErrNotFound)
	}
	return nil
}

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperrors.ErrConflict
	}
	return err
}

// scanTask works with both pgx.Row and pgx.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (*domain.Task, error) {
	var t domain.Task
	err := s.Scan(
		&t.ID,
		&t.AgencyID,
		&t.CreatedBy,
		&t.AssignedTo,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.Priority,
		&t.DueDate,
		&t.CompletedAt,
		&t.CreatedAt,
		&t.Tags,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
