package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // "admin", "member"
	AgencyID  uuid.UUID `json:"agency_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (u User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u User) CanAssignTasks() bool {
	return u.IsAdmin()
}
