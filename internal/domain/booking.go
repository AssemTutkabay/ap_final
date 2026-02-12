package domain

import "time"

type BookingStatus string

const (
	BookingCreated           BookingStatus = "created"
	BookingConfirmed         BookingStatus = "confirmed"
	BookingCancelledByClient BookingStatus = "cancelled_by_client"
	BookingCancelledBySalon  BookingStatus = "cancelled_by_salon"
	BookingCompleted         BookingStatus = "completed"
)

type Booking struct {
	ID        string `json:"id"`
	SalonID   string `json:"salonId"`
	ClientID  string `json:"clientId"`
	MasterID  string `json:"masterId"`
	ServiceID string `json:"serviceId"`

	StartAt time.Time `json:"startAt"`
	EndAt   time.Time `json:"endAt"`

	Status       BookingStatus `json:"status"`
	IsLateCancel bool          `json:"isLateCancel"`
	CancelReason string        `json:"cancelReason,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
}

func (b Booking) IsActive() bool {
	return b.Status == BookingCreated || b.Status == BookingConfirmed
}

func (b Booking) IsCancelled() bool {
	return b.Status == BookingCancelledByClient || b.Status == BookingCancelledBySalon
}
