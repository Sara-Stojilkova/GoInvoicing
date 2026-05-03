package services_test

import (
	"context"
	"errors"
	"testing"

	"backend/internal/apperrors"
	"backend/internal/repositories/memory"
	services "backend/internal/services/user"

	"github.com/google/uuid"
)

var ctx = context.Background()

func newUserService() *services.UserService {
	return services.NewUserService(memory.NewUserRepo())
}

func TestCreateUser(t *testing.T) {
	agencyID := uuid.New()

	tests := []struct {
		name     string
		userName string
		email    string
		role     string
		agencyID uuid.UUID
		wantRole string
	}{
		{"creates admin user", "Alice", "alice@example.com", "admin", agencyID, "admin"},
		{"creates member user", "Bob", "bob@example.com", "member", agencyID, "member"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserService()
			user, err := svc.Create(ctx, tt.userName, tt.email, tt.role, tt.agencyID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.FullName != tt.userName {
				t.Errorf("Name = %q, want %q", user.FullName, tt.userName)
			}
			if user.Email != tt.email {
				t.Errorf("Email = %q, want %q", user.Email, tt.email)
			}
			if user.Role != tt.wantRole {
				t.Errorf("Role = %q, want %q", user.Role, tt.wantRole)
			}
			if user.AgencyID != tt.agencyID {
				t.Errorf("AgencyID = %v, want %v", user.AgencyID, tt.agencyID)
			}
			if user.ID == (uuid.UUID{}) {
				t.Error("ID must not be zero")
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	agencyID := uuid.New()

	tests := []struct {
		name    string
		setup   func(svc *services.UserService) uuid.UUID
		wantErr error
	}{
		{
			name: "found",
			setup: func(svc *services.UserService) uuid.UUID {
				u, _ := svc.Create(ctx, "Alice", "alice@example.com", "admin", agencyID)
				return u.ID
			},
			wantErr: nil,
		},
		{
			name: "not found",
			setup: func(svc *services.UserService) uuid.UUID {
				u, _ := svc.Create(ctx, "Alice", "alice@example.com", "admin", agencyID)
				id := u.ID
				id[0] ^= 0xFF
				return id
			},
			wantErr: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserService()
			id := tt.setup(svc)
			user, err := svc.GetByID(ctx, id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetByID() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && user == nil {
				t.Error("expected user, got nil")
			}
		})
	}
}

func TestListUsersByAgency(t *testing.T) {
	agencyA := uuid.New()
	agencyB := uuid.New()

	tests := []struct {
		name      string
		setup     func(svc *services.UserService)
		agencyID  uuid.UUID
		wantCount int
	}{
		{
			name:      "empty agency",
			setup:     func(svc *services.UserService) {},
			agencyID:  agencyA,
			wantCount: 0,
		},
		{
			name: "only returns users from requested agency",
			setup: func(svc *services.UserService) {
				svc.Create(ctx, "Alice", "alice@example.com", "admin", agencyA)
				svc.Create(ctx, "Bob", "bob@example.com", "member", agencyA)
				svc.Create(ctx, "Carol", "carol@example.com", "member", agencyB)
			},
			agencyID:  agencyA,
			wantCount: 2,
		},
		{
			name: "filters to correct agency",
			setup: func(svc *services.UserService) {
				svc.Create(ctx, "Alice", "alice@example.com", "admin", agencyA)
				svc.Create(ctx, "Bob", "bob@example.com", "member", agencyB)
			},
			agencyID:  agencyB,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newUserService()
			tt.setup(svc)
			users, err := svc.ListByAgency(ctx, tt.agencyID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(users) != tt.wantCount {
				t.Errorf("len(users) = %d, want %d", len(users), tt.wantCount)
			}
		})
	}
}
