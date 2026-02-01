package domain

import "time"

type SlotStatus string

const (
	SlotFree   SlotStatus = "free"
	SlotBooked SlotStatus = "booked"
)

type Slot struct {
	ID       string     `json:"id"`
	MasterID string     `json:"masterId"`
	StartAt  time.Time  `json:"startAt"`
	EndAt    time.Time  `json:"endAt"`
	Status   SlotStatus `json:"status"`
}
