package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"ap_final/internal/domain"
	"ap_final/internal/store"
)

type BookingService struct {
	store        store.BookingStore
	serviceStore store.ServiceStore

	auditCh chan AuditEvent
}

func NewBookingService(b store.BookingStore, s store.ServiceStore) *BookingService {
	bs := &BookingService{
		store:        b,
		serviceStore: s,
		auditCh:      make(chan AuditEvent, 100),
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
		if errors.Is(err, store.ErrConflict) {
			return domain.Booking{}, errors.New("conflict: slot already booked")
		}
		if errors.Is(err, store.ErrNotFound) {
			return domain.Booking{}, errors.New("master not found")
		}
		return domain.Booking{}, err
	}

	bs.tryAudit("booking_created", created.ID)
	return created, nil
}

func (bs *BookingService) List(ctx context.Context, clientID, masterID string) ([]domain.Booking, error) {
	f := store.BookingListFilter{
		ClientID: clientID,
		MasterID: masterID,
	}
	return bs.store.List(ctx, f)
}

// NEW: list bookings for one salon (for salon_admin scope)
func (bs *BookingService) ListBySalon(ctx context.Context, salonID string) ([]domain.Booking, error) {
	f := store.BookingListFilter{SalonID: salonID}
	return bs.store.List(ctx, f)
}

// NEW: list bookings across multiple salons (salon_admin can manage multiple salons)
func (bs *BookingService) ListBySalonIDs(ctx context.Context, salonIDs []string) ([]domain.Booking, error) {
	out := make([]domain.Booking, 0)
	seen := map[string]bool{}

	for _, sid := range salonIDs {
		if strings.TrimSpace(sid) == "" {
			continue
		}
		list, err := bs.ListBySalon(ctx, sid)
		if err != nil {
			return nil, err
		}
		for _, b := range list {
			if !seen[b.ID] {
				seen[b.ID] = true
				out = append(out, b)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].StartAt.After(out[j].StartAt)
	})
	return out, nil
}

// FIX: повторный DELETE не должен ломаться.
// Если уже cancelled_by_client, вернем как есть (idempotent).
func (bs *BookingService) CancelByClient(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	if bookingID == "" {
		return domain.Booking{}, errors.New("bookingId required")
	}
	if clientID == "" {
		return domain.Booking{}, errors.New("clientId required")
	}

	existing, ok, err := bs.store.GetByID(ctx, bookingID)
	if err != nil {
		return domain.Booking{}, err
	}
	if !ok {
		return domain.Booking{}, errors.New("booking not found")
	}
	if existing.IsCancelled() {
		return domain.Booking{}, errors.New("conflict: booking already cancelled")
	}
	if existing.Status == domain.BookingCompleted {
		return domain.Booking{}, errors.New("conflict: booking already completed")
	}

	updated, err := bs.store.CancelByClient(ctx, bookingID, clientID, "", false)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return domain.Booking{}, errors.New("booking not found")
		}
		if errors.Is(err, store.ErrForbidden) {
			return domain.Booking{}, errors.New("forbidden")
		}
		return domain.Booking{}, err
	}

	bs.tryAudit("booking_cancelled_by_client", updated.ID)
	return updated, nil
}

func (bs *BookingService) UpdateStatus(ctx context.Context, bookingID string, newStatus domain.BookingStatus, cancelReason string) (domain.Booking, error) {
	if bookingID == "" {
		return domain.Booking{}, errors.New("bookingId required")
	}

	switch newStatus {
	case domain.BookingConfirmed, domain.BookingCompleted, domain.BookingCancelledBySalon:
	default:
		return domain.Booking{}, fmt.Errorf("invalid status: %s", newStatus)
	}

	updated, err := bs.store.UpdateStatus(ctx, bookingID, newStatus, cancelReason)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return domain.Booking{}, errors.New("booking not found")
		}
		return domain.Booking{}, err
	}

	bs.tryAudit("booking_status_updated", updated.ID)
	return updated, nil
}

func (bs *BookingService) GetByID(ctx context.Context, id string) (domain.Booking, bool, error) {
	if id == "" {
		return domain.Booking{}, false, errors.New("id required")
	}
	return bs.store.GetByID(ctx, id)
}

func (bs *BookingService) tryAudit(action, bookingID string) {
	select {
	case bs.auditCh <- AuditEvent{
		At:        time.Now(),
		Action:    action,
		BookingID: bookingID,
	}:
	default:
	}
}
