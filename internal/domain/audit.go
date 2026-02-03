package domain

import "time"

type AuditEvent struct {
	At        time.Time `json:"at"`
	Action    string    `json:"action"`
	BookingID string    `json:"bookingId"`
}
