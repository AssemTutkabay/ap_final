package domain

import "time"

type BookingStatus string

const (
	BookingCreated   BookingStatus = "created"
	BookingConfirmed BookingStatus = "confirmed"
	BookingCancelled BookingStatus = "cancelled"
	BookingCompleted BookingStatus = "completed"
)

type Booking struct {
	ID        string        `json:"id"`
	ClientID  string        `json:"clientId"`
	MasterID  string        `json:"masterId"`
	ServiceID string        `json:"serviceId"`
	StartAt   time.Time     `json:"startAt"`
	EndAt     time.Time     `json:"endAt"`
	Status    BookingStatus `json:"status"`
	CreatedAt time.Time     `json:"createdAt"`
}

func (b Booking) Overlaps(start, end time.Time) bool {
	return b.StartAt.Before(end) && start.Before(b.EndAt)
}

func (b Booking) IsActive() bool {
	return b.Status == BookingCreated || b.Status == BookingConfirmed
}
