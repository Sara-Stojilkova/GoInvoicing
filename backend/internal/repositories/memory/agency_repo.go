package memory

import (
	"context"
	"fmt"
	"sync"

	"backend/internal/apperrors"
	domain "backend/internal/domain/agency"
	"backend/internal/repositories"

	"github.com/google/uuid"
)

type agencyRepo struct {
	mu       sync.RWMutex
	agencies map[uuid.UUID]*domain.Agency
}

func NewAgencyRepo() repositories.AgencyRepository {
	return &agencyRepo{agencies: make(map[uuid.UUID]*domain.Agency)}
}

func (r *agencyRepo) Create(ctx context.Context, agency *domain.Agency) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agencies[agency.ID] = agency
	return nil
}

func (r *agencyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Agency, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agencies[id]
	if !ok {
		return nil, fmt.Errorf("agency %s: %w", id, apperrors.ErrNotFound)
	}
	return a, nil
}

func (r *agencyRepo) List(ctx context.Context) ([]*domain.Agency, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Agency, 0, len(r.agencies))
	for _, a := range r.agencies {
		result = append(result, a)
	}
	return result, nil
}

func (r *agencyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.agencies[id]; !ok {
		return fmt.Errorf("agency %s: %w", id, apperrors.ErrNotFound)
	}
	delete(r.agencies, id)
	return nil
}
