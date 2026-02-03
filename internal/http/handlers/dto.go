package handlers

import "time"

type CreateBookingRequest struct {
	ClientID  string    `json:"clientId"`
	MasterID  string    `json:"masterId"`
	ServiceID string    `json:"serviceId"`
	StartAt   time.Time `json:"startAt"`
}
