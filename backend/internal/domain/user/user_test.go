package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{"admin role",  User{ID: uuid.New(), Role: "admin"},  true},
		{"member role", User{ID: uuid.New(), Role: "member"}, false},
		{"empty role",  User{ID: uuid.New(), Role: ""},       false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.IsAdmin()
			if got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanAssignTasks(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{"admin can assign",    User{Role: "admin"},  true},
		{"member cannot assign", User{Role: "member"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.CanAssignTasks()
			if got != tt.want {
				t.Errorf("CanAssignTasks() = %v, want %v", got, tt.want)
			}
		})
	}
}
