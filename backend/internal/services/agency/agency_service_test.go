package services_test

import (
	"context"
	"errors"
	"testing"

	"backend/internal/apperrors"
	"backend/internal/repositories/memory"
	services "backend/internal/services/agency"

	"github.com/google/uuid"
)

var ctx = context.Background()

func newAgencyService() *services.AgencyService {
	return services.NewAgencyService(memory.NewAgencyRepo())
}

func TestCreateAgency(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
	}{
		{"creates with correct name", "Acme Corp", "Acme Corp"},
		{"creates with different name", "Globex", "Globex"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newAgencyService()
			agency, err := svc.Create(ctx, tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if agency.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", agency.Name, tt.wantName)
			}
			if agency.ID == (uuid.UUID{}) {
				t.Error("ID must not be zero")
			}
		})
	}
}

func TestGetAgency(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(svc *services.AgencyService) uuid.UUID
		wantErr error
	}{
		{
			name: "found",
			setup: func(svc *services.AgencyService) uuid.UUID {
				a, _ := svc.Create(ctx, "Acme Corp")
				return a.ID
			},
			wantErr: nil,
		},
		{
			name: "not found",
			setup: func(svc *services.AgencyService) uuid.UUID {
				a, _ := svc.Create(ctx, "Acme Corp")
				id := a.ID
				id[0] ^= 0xFF
				return id
			},
			wantErr: apperrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newAgencyService()
			id := tt.setup(svc)
			agency, err := svc.GetByID(ctx, id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetByID() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && agency == nil {
				t.Error("expected agency, got nil")
			}
		})
	}
}

func TestListAgencies(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(svc *services.AgencyService)
		wantCount int
	}{
		{
			name:      "empty",
			setup:     func(svc *services.AgencyService) {},
			wantCount: 0,
		},
		{
			name: "two agencies",
			setup: func(svc *services.AgencyService) {
				svc.Create(ctx, "Acme Corp")
				svc.Create(ctx, "Globex")
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newAgencyService()
			tt.setup(svc)
			agencies, err := svc.List(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(agencies) != tt.wantCount {
				t.Errorf("len(agencies) = %d, want %d", len(agencies), tt.wantCount)
			}
		})
	}
}
