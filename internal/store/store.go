package store

import (
	"context"

	"ap_final/internal/domain"
)

type BookingStore interface {
	Create(ctx context.Context, b domain.Booking) (domain.Booking, error)
	List(ctx context.Context) ([]domain.Booking, error)
	GetByID(ctx context.Context, id string) (domain.Booking, bool, error)
	Update(ctx context.Context, b domain.Booking) (domain.Booking, error)
}

type ServiceStore interface {
	GetByID(ctx context.Context, id string) (domain.Service, bool, error)
	List(ctx context.Context, salonID string) ([]domain.Service, error)
}
