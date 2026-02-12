package handlers

import "time"

type CreateBookingRequest struct {
	MasterID  string    `json:"masterId"`
	ServiceID string    `json:"serviceId"`
	StartAt   time.Time `json:"startAt"`
}
