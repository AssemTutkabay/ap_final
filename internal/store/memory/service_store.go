package memory

import (
	"context"
	"sync"

	"ap_final/internal/domain"
)

type ServiceStore struct {
	mu   sync.RWMutex
	byID map[string]domain.Service
}

func NewServiceStore(seed []domain.Service) *ServiceStore {
	m := make(map[string]domain.Service, len(seed))
	for _, s := range seed {
		m[s.ID] = s
	}
	return &ServiceStore{byID: m}
}

func (s *ServiceStore) GetByID(ctx context.Context, id string) (domain.Service, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.byID[id]
	return v, ok, nil
}

func (s *ServiceStore) List(ctx context.Context, salonID string) ([]domain.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Service, 0, len(s.byID))
	for _, v := range s.byID {
		if salonID == "" || v.SalonID == salonID {
			out = append(out, v)
		}
	}
	return out, nil
}
