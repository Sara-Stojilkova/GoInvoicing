//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"backend/internal/apperrors"
	agencyDomain "backend/internal/domain/agency"
	taskDomain "backend/internal/domain/task"
	userDomain "backend/internal/domain/user"
	"backend/internal/repositories/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Default to local Supabase postgres started with `supabase start`.
const defaultDSN = "postgresql://postgres:postgres@localhost:54322/postgres"

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = defaultDSN
	}

	var err error
	pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic("open pool: " + err.Error())
	}
	if err := pool.Ping(context.Background()); err != nil {
		panic("ping db: " + err.Error())
	}

	os.Exit(m.Run())
}

// --- helpers -----------------------------------------------------------------

// seedAgency inserts a row into agencies and registers cleanup.
func seedAgency(t *testing.T, name string) *agencyDomain.Agency {
	t.Helper()
	a := &agencyDomain.Agency{ID: uuid.New(), Name: name}
	_, err := pool.Exec(context.Background(),
		`insert into agencies (id, name) values ($1, $2)`, a.ID, a.Name)
	if err != nil {
		t.Fatalf("seedAgency: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `delete from agencies where id = $1`, a.ID)
	})
	return a
}

// seedUser inserts a row into auth.users; the handle_new_user trigger
// automatically creates the public.users row from raw_user_meta_data.
// The row is then activated directly so tests can access agency data.
func seedUser(t *testing.T, agencyID uuid.UUID) *userDomain.User {
	t.Helper()
	id := uuid.New()
	email := id.String() + "@test.local"

	_, err := pool.Exec(context.Background(), `
		insert into auth.users
			(id, email, encrypted_password, created_at, updated_at,
			 instance_id, aud, role, raw_user_meta_data)
		values ($1, $2, '', now(), now(),
			'00000000-0000-0000-0000-000000000000', 'authenticated', 'authenticated',
			jsonb_build_object('agency_id', $3::text, 'full_name', 'Test User'))`,
		id, email, agencyID)
	if err != nil {
		t.Fatalf("seedUser auth.users: %v", err)
	}

	// Activate the trigger-created public.users row
	_, err = pool.Exec(context.Background(),
		`update users set activated = true where id = $1`, id)
	if err != nil {
		pool.Exec(context.Background(), `delete from auth.users where id = $1`, id)
		t.Fatalf("seedUser activate: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(context.Background(), `delete from tasks      where created_by = $1`, id)
		pool.Exec(context.Background(), `delete from users      where id = $1`, id)
		pool.Exec(context.Background(), `delete from auth.users where id = $1`, id)
	})

	return &userDomain.User{ID: id, AgencyID: agencyID, Name: "Test User"}
}

// seedTask inserts a task row directly and registers cleanup.
func seedTask(t *testing.T, task *taskDomain.Task) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		insert into tasks
			(id, agency_id, created_by, assigned_to, title, description,
			 status, priority, due_date, completed_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		task.ID, task.AgencyID, task.CreatedBy, task.AssigneeID,
		task.Title, task.Description, task.Status, task.Priority,
		task.DueDate, task.CompletedAt,
	)
	if err != nil {
		t.Fatalf("seedTask: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `delete from tasks where id = $1`, task.ID)
	})
}

func baseTask(agencyID, createdBy uuid.UUID) *taskDomain.Task {
	return &taskDomain.Task{
		ID:        uuid.New(),
		AgencyID:  agencyID,
		CreatedBy: createdBy,
		Title:     "Test task",
		Status:    "todo",
		Priority:  "medium",
	}
}

// --- AgencyRepo --------------------------------------------------------------

func TestAgencyRepo_Create(t *testing.T) {
	repo := postgres.NewAgencyRepo(pool)
	ctx := context.Background()

	a := &agencyDomain.Agency{ID: uuid.New(), Name: "Acme"}
	t.Cleanup(func() {
		pool.Exec(ctx, `delete from agencies where id = $1`, a.ID)
	})

	if err := repo.Create(ctx, a); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, a.ID)
	if err != nil {
		t.Fatalf("GetByID after Create: %v", err)
	}
	if got.Name != a.Name {
		t.Errorf("Name = %q, want %q", got.Name, a.Name)
	}
}

func TestAgencyRepo_Create_DuplicateID(t *testing.T) {
	repo := postgres.NewAgencyRepo(pool)
	ctx := context.Background()

	a := seedAgency(t, "Alpha")

	err := repo.Create(ctx, &agencyDomain.Agency{ID: a.ID, Name: "Duplicate"})
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Errorf("Create duplicate = %v, want ErrConflict", err)
	}
}

func TestAgencyRepo_GetByID_NotFound(t *testing.T) {
	repo := postgres.NewAgencyRepo(pool)
	_, err := repo.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("GetByID missing = %v, want ErrNotFound", err)
	}
}

func TestAgencyRepo_GetByID_SoftDeleted(t *testing.T) {
	repo := postgres.NewAgencyRepo(pool)
	ctx := context.Background()

	a := seedAgency(t, "ToDelete")
	pool.Exec(ctx, `update agencies set deleted_at = now() where id = $1`, a.ID)

	_, err := repo.GetByID(ctx, a.ID)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("GetByID soft-deleted = %v, want ErrNotFound", err)
	}
}

func TestAgencyRepo_List(t *testing.T) {
	repo := postgres.NewAgencyRepo(pool)
	ctx := context.Background()

	a1 := seedAgency(t, "ListAgency1")
	a2 := seedAgency(t, "ListAgency2")

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := map[uuid.UUID]bool{}
	for _, a := range all {
		found[a.ID] = true
	}
	if !found[a1.ID] || !found[a2.ID] {
		t.Error("List did not return both seeded agencies")
	}
}

func TestAgencyRepo_List_ExcludesSoftDeleted(t *testing.T) {
	repo := postgres.NewAgencyRepo(pool)
	ctx := context.Background()

	a := seedAgency(t, "SoftDeletedAgency")
	pool.Exec(ctx, `update agencies set deleted_at = now() where id = $1`, a.ID)

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, ag := range all {
		if ag.ID == a.ID {
			t.Error("List returned a soft-deleted agency")
		}
	}
}

// --- UserRepo ----------------------------------------------------------------

func TestUserRepo_Create(t *testing.T) {
	// The handle_new_user trigger creates the public.users row automatically
	// when auth.users is inserted. seedUser exercises that path end-to-end.
	repo := postgres.NewUserRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "UserCreateAgency")
	u := seedUser(t, agency.ID)

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.AgencyID != agency.ID {
		t.Errorf("AgencyID = %v, want %v", got.AgencyID, agency.ID)
	}
	if got.Name != "Test User" {
		t.Errorf("Name = %q, want %q", got.Name, "Test User")
	}
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	repo := postgres.NewUserRepo(pool)
	_, err := repo.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("GetByID missing = %v, want ErrNotFound", err)
	}
}

func TestUserRepo_Update(t *testing.T) {
	repo := postgres.NewUserRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "UserUpdateAgency")
	u := seedUser(t, agency.ID)

	u.Name = "Updated Name"
	if err := repo.Update(ctx, u); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if got.Name != "Updated Name" {
		t.Errorf("Name = %q, want %q", got.Name, "Updated Name")
	}
}

func TestUserRepo_Update_NotFound(t *testing.T) {
	repo := postgres.NewUserRepo(pool)
	u := &userDomain.User{ID: uuid.New(), Name: "Ghost"}
	err := repo.Update(context.Background(), u)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("Update missing = %v, want ErrNotFound", err)
	}
}

// --- TaskRepo ----------------------------------------------------------------

func TestTaskRepo_Create(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "TaskCreateAgency")
	user := seedUser(t, agency.ID)
	task := baseTask(agency.ID, user.ID)

	t.Cleanup(func() {
		pool.Exec(ctx, `delete from tasks where id = $1`, task.ID)
	})

	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID after Create: %v", err)
	}
	if got.Title != task.Title {
		t.Errorf("Title = %q, want %q", got.Title, task.Title)
	}
	if got.Status != "todo" {
		t.Errorf("Status = %q, want todo", got.Status)
	}
	if got.Priority != "medium" {
		t.Errorf("Priority = %q, want medium", got.Priority)
	}
}

func TestTaskRepo_Create_DuplicateID(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "TaskDupAgency")
	user := seedUser(t, agency.ID)
	task := baseTask(agency.ID, user.ID)
	seedTask(t, task)

	err := repo.Create(ctx, task)
	if !errors.Is(err, apperrors.ErrConflict) {
		t.Errorf("Create duplicate = %v, want ErrConflict", err)
	}
}

func TestTaskRepo_GetByID_NotFound(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	_, err := repo.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("GetByID missing = %v, want ErrNotFound", err)
	}
}

func TestTaskRepo_List(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "TaskListAgency")
	user := seedUser(t, agency.ID)
	t1 := baseTask(agency.ID, user.ID)
	t2 := baseTask(agency.ID, user.ID)
	t2.Title = "Second task"
	seedTask(t, t1)
	seedTask(t, t2)

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := map[uuid.UUID]bool{}
	for _, task := range all {
		found[task.ID] = true
	}
	if !found[t1.ID] || !found[t2.ID] {
		t.Error("List did not return both seeded tasks")
	}
}

func TestTaskRepo_Update(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "TaskUpdateAgency")
	user := seedUser(t, agency.ID)
	task := baseTask(agency.ID, user.ID)
	seedTask(t, task)

	desc := "updated description"
	task.Title = "Updated title"
	task.Description = &desc
	task.Status = "in_progress"

	if err := repo.Update(ctx, task); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if got.Title != "Updated title" {
		t.Errorf("Title = %q, want %q", got.Title, "Updated title")
	}
	if got.Description == nil || *got.Description != desc {
		t.Errorf("Description = %v, want %q", got.Description, desc)
	}
	if got.Status != "in_progress" {
		t.Errorf("Status = %q, want in_progress", got.Status)
	}
}

func TestTaskRepo_Update_NotFound(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	task := baseTask(uuid.New(), uuid.New())
	err := repo.Update(context.Background(), task)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("Update missing = %v, want ErrNotFound", err)
	}
}

func TestTaskRepo_Update_CompletedAt(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "TaskCompleteAgency")
	user := seedUser(t, agency.ID)
	task := baseTask(agency.ID, user.ID)
	seedTask(t, task)

	completedAt := time.Now().UTC().Truncate(time.Millisecond)
	task.Status = "done"
	task.CompletedAt = &completedAt

	if err := repo.Update(ctx, task); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.CompletedAt == nil {
		t.Fatal("CompletedAt is nil, want non-nil")
	}
	if !got.CompletedAt.Equal(completedAt) {
		t.Errorf("CompletedAt = %v, want %v", got.CompletedAt, completedAt)
	}
}

func TestTaskRepo_Delete(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	ctx := context.Background()

	agency := seedAgency(t, "TaskDeleteAgency")
	user := seedUser(t, agency.ID)
	task := baseTask(agency.ID, user.ID)
	seedTask(t, task)

	if err := repo.Delete(ctx, task.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, task.ID)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("GetByID after Delete = %v, want ErrNotFound", err)
	}
}

func TestTaskRepo_Delete_NotFound(t *testing.T) {
	repo := postgres.NewTaskRepo(pool)
	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("Delete missing = %v, want ErrNotFound", err)
	}
}
