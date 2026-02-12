package domain

import "time"

type Review struct {
	ID        string    `json:"id"`
	BookingID string    `json:"bookingId"`
	ClientID  string    `json:"clientId"`
	SalonID   string    `json:"salonId"`
	MasterID  string    `json:"masterId"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
