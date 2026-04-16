package seed

import (
	"time"

	domainAgency "backend/internal/domain/agency"
	domainTask "backend/internal/domain/task"
	domainUser "backend/internal/domain/user"

	"github.com/google/uuid"
)

type Data struct {
	Agencies []domainAgency.Agency
	Users    []domainUser.User
	Tasks    []domainTask.Task
}

func Generate() Data {
	now := time.Now()

	// ---- Agencies ----
	agency1 := domainAgency.Agency{
		ID:        uuid.New(),
		Name:      "Bilans Solutions",
		CreatedAt: now,
	}

	// ---- Users ----
	sara := domainUser.User{
		ID:        uuid.New(),
		Name:      "Sara Stojilkova",
		Email:     "sara.stojilkova@bilans.dev",
		Role:      "admin",
		AgencyID:  agency1.ID,
		CreatedAt: now,
	}

	john := domainUser.User{
		ID:        uuid.New(),
		Name:      "John Doe",
		Email:     "john.doe@bilans.dev",
		Role:      "member",
		AgencyID:  agency1.ID,
		CreatedAt: now,
	}

	jane := domainUser.User{
		ID:        uuid.New(),
		Name:      "Jane Doe",
		Email:     "jane.doe@bilans.dev",
		Role:      "member",
		AgencyID:  agency1.ID,
		CreatedAt: now,
	}

	// ---- Tasks ----
	desc := "Important task"
	due := now.Add(48 * time.Hour)

	task1 := domainTask.Task{
		ID:          uuid.New(),
		Title:       "Finish report",
		Description: &desc,
		Status:      "todo",
		Priority:    "high",
		AgencyID:    agency1.ID,
		AssigneeID:  &sara.ID,
		CreatedAt:   now,
		DueDate:     &due,
	}

	task2 := domainTask.Task{
		ID:         uuid.New(),
		Title:      "Fix bug",
		Status:     "in_progress",
		Priority:   "medium",
		AgencyID:   agency1.ID,
		AssigneeID: &john.ID,
		CreatedAt:  now,
	}

	task3 := domainTask.Task{
		ID:        uuid.New(),
		Title:     "Prepare invoice",
		Status:    "done",
		Priority:  "low",
		AgencyID:  agency1.ID,
		CreatedAt: now,
	}

	task4 := domainTask.Task{
		ID:        uuid.New(),
		Title:     "Write email",
		Status:    "todo",
		Priority:  "low",
		AgencyID:  agency1.ID,
		CreatedAt: now,
	}

	task5 := domainTask.Task{
		ID:        uuid.New(),
		Title:     "Prepare presentation",
		Status:    "in_progress",
		Priority:  "high",
		AgencyID:  agency1.ID,
		CreatedAt: now,
	}

	_ = task3.Complete(now)

	return Data{
		Agencies: []domainAgency.Agency{agency1},
		Users:    []domainUser.User{sara, john, jane},
		Tasks:    []domainTask.Task{task1, task2, task3, task4, task5},
	}
}
