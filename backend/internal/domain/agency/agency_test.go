package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestAgency(t *testing.T) {
	now := time.Date(2026, 4, 3, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		agency    Agency
		wantName  string
	}{
		{"basic agency", Agency{ID: uuid.New(), Name: "Acme Corp", CreatedAt: now}, "Acme Corp"},
		{"agency with empty name", Agency{ID: uuid.New(), Name: "", CreatedAt: now}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.agency.Name != tt.wantName {
				t.Errorf("Agency.Name = %q, want %q", tt.agency.Name, tt.wantName)
			}
			if tt.agency.ID == (uuid.UUID{}) {
				t.Error("Agency.ID must not be zero")
			}
		})
	}
}
