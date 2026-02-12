package store

import (
	"context"
	"errors"

	"ap_final/internal/domain"
)

var (
	// конфликт данных: overlap booking, duplicate email/phone, duplicate review for booking, etc
	ErrConflict = errors.New("conflict")
	// не нашли запись
	ErrNotFound = errors.New("not found")
	// запрет по доступу (ownership / роль)
	ErrForbidden = errors.New("forbidden")
)

/*
Контракт:
- store слой не знает про HTTP.
- store возвращает доменные модели + общие ошибки (ErrNotFound/ErrConflict/ErrForbidden).
- маппинг ошибок в HTTP коды делает handler/service слой.
*/

// ---------- Users ----------

type UserStore interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, bool, error)
	GetByEmail(ctx context.Context, email string) (domain.User, bool, error)
	GetByPhone(ctx context.Context, phone string) (domain.User, bool, error)
}

// ---------- Salons ----------

type SalonStore interface {
	// public marketplace list (marketplace_enabled = true)
	ListPublic(ctx context.Context, city string) ([]domain.Salon, error)
	GetByID(ctx context.Context, id string) (domain.Salon, bool, error)

	// admin CRUD
	Create(ctx context.Context, s domain.Salon) (domain.Salon, error)
	Update(ctx context.Context, s domain.Salon) (domain.Salon, error)
	Delete(ctx context.Context, id string) error
}

// Ownership связка salon_admins
type SalonAdminStore interface {
	IsAdminOfSalon(ctx context.Context, userID, salonID string) (bool, error)

	// удобно для админских страниц/выборок
	ListSalonIDsForAdmin(ctx context.Context, userID string) ([]string, error)
}

// ---------- Masters ----------

type MasterStore interface {
	ListBySalon(ctx context.Context, salonID string) ([]domain.Master, error)
	GetByID(ctx context.Context, id string) (domain.Master, bool, error)

	// NEW: нужна привязка user -> master для RBAC master
	GetIDByUserID(ctx context.Context, userID string) (string, bool, error)

	CreateForSalon(ctx context.Context, m domain.Master) (domain.Master, error)
	Update(ctx context.Context, m domain.Master) (domain.Master, error)

	// можно soft-disable вместо delete
	Disable(ctx context.Context, id string) error
}

// ---------- Services ----------

type ServiceStore interface {
	ListBySalon(ctx context.Context, salonID string) ([]domain.Service, error)
	GetByID(ctx context.Context, id string) (domain.Service, bool, error)

	CreateForSalon(ctx context.Context, s domain.Service) (domain.Service, error)
	Update(ctx context.Context, s domain.Service) (domain.Service, error)
	Disable(ctx context.Context, id string) error
}

// ---------- Bookings ----------

type BookingListFilter struct {
	ClientID string
	MasterID string
	SalonID  string
}

type BookingStore interface {
	Create(ctx context.Context, b domain.Booking) (domain.Booking, error)

	List(ctx context.Context, f BookingListFilter) ([]domain.Booking, error)

	GetByID(ctx context.Context, id string) (domain.Booking, bool, error)

	// клиент отменяет только свою бронь
	CancelByClient(ctx context.Context, bookingID, clientID string, cancelReason string, isLateCancel bool) (domain.Booking, error)

	// админ/платформа меняет статус (confirmed/completed/cancelled_by_salon)
	UpdateStatus(ctx context.Context, bookingID string, newStatus domain.BookingStatus, cancelReason string) (domain.Booking, error)
}

// ---------- Reviews ----------

type ReviewStore interface {
	Create(ctx context.Context, r domain.Review) (domain.Review, error)

	ListBySalon(ctx context.Context, salonID string) ([]domain.Review, error)

	// чтобы проверить "1 booking = 1 review"
	GetByBookingID(ctx context.Context, bookingID string) (domain.Review, bool, error)
}

// ---------- Audit ----------

type AuditStore interface {
	Insert(ctx context.Context, ev domain.AuditLog) error
}
