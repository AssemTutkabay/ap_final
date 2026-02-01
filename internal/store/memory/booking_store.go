package memory

import (
	"context"
	"errors"
	"sync"

	"ap_final/internal/domain"
)

var ErrConflict = errors.New("booking conflict")
var ErrNotFound = errors.New("booking not found")

type BookingStore struct {
	mu     sync.RWMutex
	byID   map[string]domain.Booking
	nextID int
}

func NewBookingStore() *BookingStore {
	return &BookingStore{
		byID:   make(map[string]domain.Booking),
		nextID: 1,
	}
}

// Create делает атомарно: проверка пересечения + запись под одним Lock().
func (s *BookingStore) Create(ctx context.Context, b domain.Booking) (domain.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Конфликт по мастеру и времени (active bookings)
	for _, existing := range s.byID {
		if existing.MasterID != b.MasterID {
			continue
		}
		if !existing.IsActive() {
			continue
		}
		if existing.Overlaps(b.StartAt, b.EndAt) {
			return domain.Booking{}, ErrConflict
		}
	}

	b.ID = itoa(s.nextID)
	s.nextID++

	s.byID[b.ID] = b
	return b, nil
}

func (s *BookingStore) List(ctx context.Context) ([]domain.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Booking, 0, len(s.byID))
	for _, b := range s.byID {
		out = append(out, b)
	}
	return out, nil
}

func (s *BookingStore) GetByID(ctx context.Context, id string) (domain.Booking, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.byID[id]
	return b, ok, nil
}

func (s *BookingStore) Update(ctx context.Context, b domain.Booking) (domain.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.byID[b.ID]; !ok {
		return domain.Booking{}, ErrNotFound
	}
	s.byID[b.ID] = b
	return b, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [32]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	return sign + string(buf[i:])
}
