package service

import (
	"context"
	"errors"
	"time"

	"ap_final/internal/domain"
	"ap_final/internal/store"
	"ap_final/internal/store/memory"
)

type BookingService struct {
	store        store.BookingStore
	serviceStore store.ServiceStore

	auditCh chan domain.AuditEvent
}

func NewBookingService(b store.BookingStore, s store.ServiceStore) *BookingService {
	bs := &BookingService{
		store:        b,
		serviceStore: s,
		auditCh:      make(chan domain.AuditEvent, 100),
	}
	go startAuditWorker(bs.auditCh)
	return bs
}

func (bs *BookingService) Create(ctx context.Context, clientID, masterID, serviceID string, startAt time.Time) (domain.Booking, error) {
	if clientID == "" || masterID == "" || serviceID == "" {
		return domain.Booking{}, errors.New("clientId/masterId/serviceId required")
	}
	if startAt.IsZero() {
		return domain.Booking{}, errors.New("startAt required")
	}

	svc, ok, err := bs.serviceStore.GetByID(ctx, serviceID)
	if err != nil {
		return domain.Booking{}, err
	}
	if !ok {
		return domain.Booking{}, errors.New("service not found")
	}
	if svc.DurationMin <= 0 {
		return domain.Booking{}, errors.New("invalid service duration")
	}

	endAt := startAt.Add(time.Duration(svc.DurationMin) * time.Minute)

	newB := domain.Booking{
		ClientID:  clientID,
		MasterID:  masterID,
		ServiceID: serviceID,
		StartAt:   startAt,
		EndAt:     endAt,
		Status:    domain.BookingCreated,
		CreatedAt: time.Now(),
	}

	created, err := bs.store.Create(ctx, newB)
	if err != nil {
		// Если Create делает conflict-check внутри store (как мы переделывали),
		// то на параллельных запросах будет возвращаться ErrConflict.
		if errors.Is(err, memory.ErrConflict) {
			return domain.Booking{}, errors.New("conflict: slot already booked")
		}
		return domain.Booking{}, err
	}

	// audit event
	select {
	case bs.auditCh <- domain.AuditEvent{
		At:        time.Now(),
		Action:    "booking_created",
		BookingID: created.ID,
	}:
	default:
		// если канал забит, не валим запрос (лог важен, но не важнее брони)
	}

	return created, nil
}

func (bs *BookingService) List(ctx context.Context, clientID, masterID string) ([]domain.Booking, error) {
	all, err := bs.store.List(ctx)
	if err != nil {
		return nil, err
	}

	if clientID == "" && masterID == "" {
		return all, nil
	}

	out := make([]domain.Booking, 0, len(all))
	for _, b := range all {
		if clientID != "" && b.ClientID != clientID {
			continue
		}
		if masterID != "" && b.MasterID != masterID {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

func (bs *BookingService) Cancel(ctx context.Context, id string) (domain.Booking, error) {
	if id == "" {
		return domain.Booking{}, errors.New("id required")
	}

	b, ok, err := bs.store.GetByID(ctx, id)
	if err != nil {
		return domain.Booking{}, err
	}
	if !ok {
		return domain.Booking{}, errors.New("booking not found")
	}
	if b.Status == domain.BookingCancelled {
		return b, nil
	}

	b.Status = domain.BookingCancelled
	updated, err := bs.store.Update(ctx, b)
	if err != nil {
		return domain.Booking{}, err
	}

	// audit event
	select {
	case bs.auditCh <- domain.AuditEvent{
		At:        time.Now(),
		Action:    "booking_cancelled",
		BookingID: updated.ID,
	}:
	default:
	}

	return updated, nil
}
