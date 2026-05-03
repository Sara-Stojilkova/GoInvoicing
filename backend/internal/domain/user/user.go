package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // "admin", "member"
	AgencyID  uuid.UUID `json:"agency_id"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"created_at"`
}

func (u User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u User) CanAssignTasks() bool {
	return u.IsAdmin()
}
